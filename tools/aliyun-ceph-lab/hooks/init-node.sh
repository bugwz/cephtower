#!/usr/bin/env bash
set -Eeuo pipefail

logf() {
	local level="$1" format="$2"
	shift 2
	printf '[%s] %s ' "$(timestamp)" "${level}"
	printf "${format}" "$@"
	printf '\n'
}

timestamp() {
	date --iso-8601=seconds 2>/dev/null || date +%Y-%m-%dT%H:%M:%S%z
}

usage() {
	cat <<'EOF'
Usage:
  sudo bash init-node.sh --cluster-name NAME --node-name NAME \
    --node-names CSV --public-ips CSV --private-ips CSV \
    --data-disk-count N --ssh-private-key-base64 VALUE \
    --ssh-public-key-base64 VALUE

Options:
  --cluster-name NAME              Cluster name, for example ceph-dev.
  --node-name NAME                 Current node hostname, for example ceph-node-1.
  --node-names CSV                 Comma-separated hostnames in cluster order.
  --public-ips CSV                 Comma-separated public IPs in the same order.
  --private-ips CSV                Comma-separated private IPs in the same order.
  --data-disk-count N              Number of data disks on this node.
  --ssh-private-key-base64 VALUE   Shared cluster SSH private key, base64 encoded.
  --ssh-public-key-base64 VALUE    Shared cluster SSH public key, base64 encoded.
  -h, --help                       Show this help.

Example:
  sudo bash init-node.sh \
    --cluster-name ceph-dev \
    --node-name ceph-node-1 \
    --node-names ceph-node-1,ceph-node-2,ceph-node-3 \
    --public-ips 8.8.8.1,8.8.8.2,8.8.8.3 \
    --private-ips 172.31.0.10,172.31.0.11,172.31.0.12 \
    --data-disk-count 2 \
    --ssh-private-key-base64 '<base64 of shared private key>' \
    --ssh-public-key-base64 '<base64 of shared public key>'

Generate the shared key pair before running init-node.sh on any node:
  ssh-keygen -t ed25519 -f ceph_lab_ed25519 -N '' -C ceph-dev
  base64 -w0 ceph_lab_ed25519
  base64 -w0 ceph_lab_ed25519.pub

On macOS, use: base64 -i ceph_lab_ed25519 | tr -d '\n'
EOF
}

require_option_value() {
	local option="$1" remaining="$2"
	if (( remaining < 2 )); then
		errorf '%s requires a value' "${option}"
		exit 1
	fi
}

parse_args() {
	while (($# > 0)); do
		case "$1" in
			-h | --help)
				usage
				exit 0
				;;
			--cluster-name)
				require_option_value "$1" "$#"
				CEPH_LAB_CLUSTER_NAME="${2:-}"
				shift 2
				;;
			--cluster-name=*)
				CEPH_LAB_CLUSTER_NAME="${1#--cluster-name=}"
				shift
				;;
			--node-name)
				require_option_value "$1" "$#"
				CEPH_LAB_NODE_NAME="${2:-}"
				shift 2
				;;
			--node-name=*)
				CEPH_LAB_NODE_NAME="${1#--node-name=}"
				shift
				;;
			--node-names)
				require_option_value "$1" "$#"
				CEPH_LAB_NODE_NAMES="${2:-}"
				shift 2
				;;
			--node-names=*)
				CEPH_LAB_NODE_NAMES="${1#--node-names=}"
				shift
				;;
			--public-ips)
				require_option_value "$1" "$#"
				CEPH_LAB_PUBLIC_IPS="${2:-}"
				shift 2
				;;
			--public-ips=*)
				CEPH_LAB_PUBLIC_IPS="${1#--public-ips=}"
				shift
				;;
			--private-ips)
				require_option_value "$1" "$#"
				CEPH_LAB_PRIVATE_IPS="${2:-}"
				shift 2
				;;
			--private-ips=*)
				CEPH_LAB_PRIVATE_IPS="${1#--private-ips=}"
				shift
				;;
			--data-disk-count)
				require_option_value "$1" "$#"
				CEPH_LAB_DATA_DISK_COUNT="${2:-}"
				shift 2
				;;
			--data-disk-count=*)
				CEPH_LAB_DATA_DISK_COUNT="${1#--data-disk-count=}"
				shift
				;;
			--ssh-private-key-base64)
				require_option_value "$1" "$#"
				CEPH_LAB_SSH_PRIVATE_KEY_BASE64="${2:-}"
				shift 2
				;;
			--ssh-private-key-base64=*)
				CEPH_LAB_SSH_PRIVATE_KEY_BASE64="${1#--ssh-private-key-base64=}"
				shift
				;;
			--ssh-public-key-base64)
				require_option_value "$1" "$#"
				CEPH_LAB_SSH_PUBLIC_KEY_BASE64="${2:-}"
				shift 2
				;;
			--ssh-public-key-base64=*)
				CEPH_LAB_SSH_PUBLIC_KEY_BASE64="${1#--ssh-public-key-base64=}"
				shift
				;;
			*)
				errorf 'unknown argument: %s' "$1"
				usage >&2
				exit 1
				;;
		esac
	done
}

infof() {
	logf INFO "$@"
}

errorf() {
	logf ERROR "$@" >&2
}

report_error() {
	local status=$?
	errorf 'init-node.sh failed: host=%s line=%s command=%s exit_status=%d' \
		"${HOSTNAME:-unknown}" "${BASH_LINENO[0]}" "${BASH_COMMAND}" "${status}"
	exit "${status}"
}
trap report_error ERR

parse_args "$@"

trace_command() {
	local command="$1"
	[[ -n "${command}" ]] || return 0
	case "${command}" in
		trace_command* | logf* | timestamp* | infof* | errorf* | report_error* | printf\ * | shift\ * | local\ * | return\ * | exit\ *)
			return 0
			;;
	esac
	logf COMMAND '%s' "${command}" || true
}
trap 'trace_command "${BASH_COMMAND}"' DEBUG

required=(
	"--cluster-name:CEPH_LAB_CLUSTER_NAME"
	"--node-name:CEPH_LAB_NODE_NAME"
	"--node-names:CEPH_LAB_NODE_NAMES"
	"--public-ips:CEPH_LAB_PUBLIC_IPS"
	"--private-ips:CEPH_LAB_PRIVATE_IPS"
	"--data-disk-count:CEPH_LAB_DATA_DISK_COUNT"
	"--ssh-private-key-base64:CEPH_LAB_SSH_PRIVATE_KEY_BASE64"
	"--ssh-public-key-base64:CEPH_LAB_SSH_PUBLIC_KEY_BASE64"
)
for item in "${required[@]}"; do
	option="${item%%:*}"
	name="${item#*:}"
	if [[ -z "${!name:-}" ]]; then
		errorf 'missing required option: option=%s' "${option}"
		exit 1
	fi
done
ssh_private_key_base64="${CEPH_LAB_SSH_PRIVATE_KEY_BASE64}"
ssh_public_key_base64="${CEPH_LAB_SSH_PUBLIC_KEY_BASE64}"
unset CEPH_LAB_SSH_PRIVATE_KEY_BASE64 CEPH_LAB_SSH_PUBLIC_KEY_BASE64

if ! command -v dnf >/dev/null 2>&1; then
	errorf '%s' 'dnf is required by the Aliyun Ceph node initialization script'
	exit 1
fi
if ! command -v ssh >/dev/null 2>&1; then
	errorf '%s' 'openssh-clients must be present before SSH initialization starts'
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
	errorf 'node metadata counts differ: nodes=%d public_ips=%d private_ips=%d' \
		"${#node_names[@]}" "${#public_ips[@]}" "${#private_ips[@]}"
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
	errorf 'not enough eligible data disks: found=%d expected=%d' \
		"${#candidates[@]}" "${CEPH_LAB_DATA_DISK_COUNT}"
	exit 1
fi
printf '%s\n' "${candidates[@]:0:CEPH_LAB_DATA_DISK_COUNT}" > /etc/ceph-lab/osd-devices
chmod 0644 /etc/ceph-lab/osd-devices

infof 'node initialized: node=%s cluster=%s; hostnames use private IPs; public IPs remain available for external access' \
	"${CEPH_LAB_NODE_NAME}" "${CEPH_LAB_CLUSTER_NAME}"
