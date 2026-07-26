#!/usr/bin/env bash
set -Eeuo pipefail

report_error() {
	local status=$?
	echo "ERROR: init-node.sh failed on ${HOSTNAME:-unknown} at line ${BASH_LINENO[0]}: ${BASH_COMMAND} (exit ${status})" >&2
	exit "${status}"
}
trap report_error ERR

required=(
	CEPH_LAB_CLUSTER_NAME
	CEPH_LAB_NODE_NAME
	CEPH_LAB_NODE_NAMES
	CEPH_LAB_PUBLIC_IPS
	CEPH_LAB_PRIVATE_IPS
	CEPH_LAB_DATA_DISK_COUNT
	CEPH_LAB_SSH_PRIVATE_KEY_BASE64
	CEPH_LAB_SSH_PUBLIC_KEY_BASE64
)
for name in "${required[@]}"; do
	if [[ -z "${!name:-}" ]]; then
		echo "missing required environment variable: ${name}" >&2
		exit 1
	fi
done
ssh_private_key_base64="${CEPH_LAB_SSH_PRIVATE_KEY_BASE64}"
ssh_public_key_base64="${CEPH_LAB_SSH_PUBLIC_KEY_BASE64}"
unset CEPH_LAB_SSH_PRIVATE_KEY_BASE64 CEPH_LAB_SSH_PUBLIC_KEY_BASE64

if ! command -v dnf >/dev/null 2>&1; then
	echo "dnf is required by the Aliyun Ceph node initialization script" >&2
	exit 1
fi
if ! command -v ssh >/dev/null 2>&1; then
	echo "openssh-clients must be present before SSH initialization starts" >&2
	exit 1
fi

# Keep all package operations non-interactive. These are intentionally executed
# in dependency order because the Ceph repository package provides cephadm.
dnf -y makecache
dnf -y install epel-release
dnf -y groupinstall "Development Tools"
dnf -y install centos-release-ceph-tentacle
dnf -y install cephadm
dnf -y install git cmake wget curl vim
dnf -y install ceph-common
dnf -y install ca-certificates chrony jq lvm2 podman python3

if command -v systemctl >/dev/null 2>&1; then
	systemctl enable --now chronyd
fi

IFS=',' read -r -a node_names <<<"${CEPH_LAB_NODE_NAMES}"
IFS=',' read -r -a public_ips <<<"${CEPH_LAB_PUBLIC_IPS}"
IFS=',' read -r -a private_ips <<<"${CEPH_LAB_PRIVATE_IPS}"
if (( ${#node_names[@]} == ${#public_ips[@]} && ${#node_names[@]} == ${#private_ips[@]} )); then
	:
else
	echo "node name, public IP, and private IP counts differ" >&2
	exit 1
fi

hosts_tmp="$(mktemp)"
trap 'rm -f "${hosts_tmp}"' EXIT
awk '
  $0 == "# BEGIN CEPH LAB" {skip=1; next}
  $0 == "# END CEPH LAB" {skip=0; next}
  !skip {print}
' /etc/hosts >"${hosts_tmp}"
{
	echo "# BEGIN CEPH LAB"
	for index in "${!node_names[@]}"; do
		printf '%s %s\n' "${private_ips[$index]}" "${node_names[$index]}"
	done
	echo "# END CEPH LAB"
} >>"${hosts_tmp}"
install -m 0644 "${hosts_tmp}" /etc/hosts

# This key pair is generated once per create operation and streamed to every
# node. It is never stored in config.yaml or in the program source.
install -d -m 0700 /root/.ssh
printf '%s' "${ssh_private_key_base64}" | base64 -d > /root/.ssh/id_ed25519
printf '%s' "${ssh_public_key_base64}" | base64 -d > /root/.ssh/id_ed25519.pub
chmod 0600 /root/.ssh/id_ed25519
chmod 0644 /root/.ssh/id_ed25519.pub
unset ssh_private_key_base64 ssh_public_key_base64
touch /root/.ssh/authorized_keys
chmod 0600 /root/.ssh/authorized_keys
public_key="$(cat /root/.ssh/id_ed25519.pub)"
grep -qxF "${public_key}" /root/.ssh/authorized_keys || printf '%s\n' "${public_key}" >> /root/.ssh/authorized_keys
touch /root/.ssh/known_hosts
chmod 0600 /root/.ssh/known_hosts
for ip in "${private_ips[@]}"; do
	ssh-keygen -R "${ip}" -f /root/.ssh/known_hosts >/dev/null 2>&1 || true
	ssh-keyscan -T 10 -H "${ip}" >> /root/.ssh/known_hosts 2>/dev/null || true
done
cat > /root/.ssh/config <<'EOF'
Host *
    BatchMode yes
    ConnectTimeout 10
    IdentityFile /root/.ssh/id_ed25519
    StrictHostKeyChecking accept-new
    UserKnownHostsFile /root/.ssh/known_hosts
EOF
chmod 0600 /root/.ssh/config

# Record only unmounted whole disks other than the disk that owns '/'. The
# configured data-disk count prevents unrelated instance disks from becoming
# OSDs accidentally.
install -d -m 0755 /etc/ceph-lab
root_source="$(findmnt -n -o SOURCE /)"
root_parent="$(lsblk -ndo PKNAME "${root_source}" 2>/dev/null || true)"
if [[ -z "${root_parent}" ]]; then
	root_parent="$(basename "${root_source}")"
fi
mapfile -t candidates < <(
	lsblk -dnpo NAME,TYPE | awk '$2 == "disk" {print $1}' | while read -r device; do
		if [[ "$(basename "${device}")" == "${root_parent}" ]]; then
			continue
		fi
		if lsblk -nrpo MOUNTPOINT "${device}" | grep -qE '.+'; then
			continue
		fi
		printf '%s\n' "${device}"
	done
)
if (( ${#candidates[@]} < CEPH_LAB_DATA_DISK_COUNT )); then
	echo "found ${#candidates[@]} eligible data disk(s), expected ${CEPH_LAB_DATA_DISK_COUNT}" >&2
	exit 1
fi
printf '%s\n' "${candidates[@]:0:CEPH_LAB_DATA_DISK_COUNT}" > /etc/ceph-lab/osd-devices
chmod 0644 /etc/ceph-lab/osd-devices

echo "initialized ${CEPH_LAB_NODE_NAME} for ${CEPH_LAB_CLUSTER_NAME}; hostnames use private IPs; public IPs remain available for external access"
