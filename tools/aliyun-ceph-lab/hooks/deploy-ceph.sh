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
  sudo bash deploy-ceph.sh --cluster-name NAME --bootstrap-node-name NAME \
    --node-names CSV --public-ips CSV --private-ips CSV \
    --data-disk-counts CSV --wait-timeout-seconds N --dashboard-password PASSWORD

Run this script only on the first/bootstrap node after init-node.sh has
completed successfully on every node.

Options:
  --cluster-name NAME             Cluster name, for example ceph-dev.
  --bootstrap-node-name NAME      First/bootstrap node hostname.
  --node-names CSV                Comma-separated hostnames in cluster order.
  --public-ips CSV                Comma-separated public IPs in the same order.
  --private-ips CSV               Comma-separated private IPs in the same order.
  --data-disk-counts CSV          Comma-separated data disk counts in the same order.
  --wait-timeout-seconds N        Timeout for each readiness wait, for example 900.
  --dashboard-password PASSWORD   Dashboard admin password to configure and record.
  -h, --help                      Show this help.

Example:
  sudo bash deploy-ceph.sh \
    --cluster-name ceph-dev \
    --bootstrap-node-name ceph-node-1 \
    --node-names ceph-node-1,ceph-node-2,ceph-node-3 \
    --public-ips 8.8.8.1,8.8.8.2,8.8.8.3 \
    --private-ips 172.31.0.10,172.31.0.11,172.31.0.12 \
    --data-disk-counts 2,2,2 \
    --wait-timeout-seconds 900 \
    --dashboard-password 'CephTower#example'

Manual run order:
  1. Copy init-node.sh and deploy-ceph.sh to the nodes.
  2. Run init-node.sh on every node with that node's --node-name and
     --data-disk-count.
  3. Run deploy-ceph.sh on --bootstrap-node-name only.
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
			--bootstrap-node-name)
				require_option_value "$1" "$#"
				CEPH_LAB_BOOTSTRAP_NODE_NAME="${2:-}"
				shift 2
				;;
			--bootstrap-node-name=*)
				CEPH_LAB_BOOTSTRAP_NODE_NAME="${1#--bootstrap-node-name=}"
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
			--data-disk-counts)
				require_option_value "$1" "$#"
				CEPH_LAB_DATA_DISK_COUNTS="${2:-}"
				shift 2
				;;
			--data-disk-counts=*)
				CEPH_LAB_DATA_DISK_COUNTS="${1#--data-disk-counts=}"
				shift
				;;
			--wait-timeout-seconds)
				require_option_value "$1" "$#"
				CEPH_LAB_WAIT_TIMEOUT_SECONDS="${2:-}"
				shift 2
				;;
			--wait-timeout-seconds=*)
				CEPH_LAB_WAIT_TIMEOUT_SECONDS="${1#--wait-timeout-seconds=}"
				shift
				;;
			--dashboard-password)
				require_option_value "$1" "$#"
				CEPH_LAB_DASHBOARD_PASSWORD="${2:-}"
				shift 2
				;;
			--dashboard-password=*)
				CEPH_LAB_DASHBOARD_PASSWORD="${1#--dashboard-password=}"
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
	errorf 'deploy-ceph.sh failed: host=%s line=%s command=%s exit_status=%d' \
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
	"--bootstrap-node-name:CEPH_LAB_BOOTSTRAP_NODE_NAME"
	"--node-names:CEPH_LAB_NODE_NAMES"
	"--public-ips:CEPH_LAB_PUBLIC_IPS"
	"--private-ips:CEPH_LAB_PRIVATE_IPS"
	"--data-disk-counts:CEPH_LAB_DATA_DISK_COUNTS"
	"--wait-timeout-seconds:CEPH_LAB_WAIT_TIMEOUT_SECONDS"
	"--dashboard-password:CEPH_LAB_DASHBOARD_PASSWORD"
)
for item in "${required[@]}"; do
	option="${item%%:*}"
	name="${item#*:}"
	if [[ -z "${!name:-}" ]]; then
		errorf 'missing required option: option=%s' "${option}"
		exit 1
	fi
done

IFS=',' read -r -a node_names <<<"${CEPH_LAB_NODE_NAMES}"
IFS=',' read -r -a public_ips <<<"${CEPH_LAB_PUBLIC_IPS}"
IFS=',' read -r -a private_ips <<<"${CEPH_LAB_PRIVATE_IPS}"
IFS=',' read -r -a data_disk_counts <<<"${CEPH_LAB_DATA_DISK_COUNTS}"
node_count="${#node_names[@]}"
if (( node_count < 3 )); then
	errorf 'Ceph deployment requires at least 3 nodes: node_count=%d' "${node_count}"
	exit 1
fi
if (( ${#public_ips[@]} != node_count || ${#private_ips[@]} != node_count || ${#data_disk_counts[@]} != node_count )); then
	errorf 'node metadata counts differ: nodes=%d public_ips=%d private_ips=%d data_disk_counts=%d' \
		"${node_count}" "${#public_ips[@]}" "${#private_ips[@]}" "${#data_disk_counts[@]}"
	exit 1
fi
if [[ "${CEPH_LAB_BOOTSTRAP_NODE_NAME}" != "${node_names[0]}" ]]; then
	errorf 'bootstrap node is not the first configured node: bootstrap=%s first_node=%s' \
		"${CEPH_LAB_BOOTSTRAP_NODE_NAME}" "${node_names[0]}"
	exit 1
fi
current_hostname="$(hostname -s)"
if [[ "${current_hostname}" != "${CEPH_LAB_BOOTSTRAP_NODE_NAME}" ]]; then
	errorf 'deployment hook must run on bootstrap node: bootstrap=%s current_host=%s' \
		"${CEPH_LAB_BOOTSTRAP_NODE_NAME}" "${current_hostname}"
	exit 1
fi

wait_until() {
	local description="$1"
	shift
	local deadline=$((SECONDS + CEPH_LAB_WAIT_TIMEOUT_SECONDS))
	while ! "$@"; do
		if (( SECONDS >= deadline )); then
			errorf 'timed out waiting for %s' "${description}"
			return 1
		fi
		infof 'waiting for %s; retrying in 10s' "${description}"
		sleep 10
	done
	infof 'ready: %s' "${description}"
}

ssh_options=(-o BatchMode=yes -o ConnectTimeout=10 -o StrictHostKeyChecking=yes -o IdentityFile=/root/.ssh/id_ed25519)
install -d -m 0700 /root/.ssh
touch /root/.ssh/known_hosts
chmod 0600 /root/.ssh/known_hosts
record_host_key() {
	local ip="$1" keys
	keys="$(ssh-keyscan -T 10 -H "${ip}" 2>/dev/null)" || return 1
	[[ -n "${keys}" ]] || return 1
	printf '%s\n' "${keys}" >> /root/.ssh/known_hosts
}
for index in "${!private_ips[@]}"; do
	if (( index == 0 )); then
		continue
	fi
	ip="${private_ips[$index]}"
	ssh-keygen -R "${ip}" -f /root/.ssh/known_hosts >/dev/null 2>&1 || true
	wait_until "SSH host key from ${node_names[$index]} (${ip})" record_host_key "${ip}"
done

ssh_ready() {
	ssh "${ssh_options[@]}" "root@$1" true >/dev/null
}
for index in "${!node_names[@]}"; do
	if (( index == 0 )); then
		continue
	fi
	wait_until "passwordless SSH to ${node_names[$index]} (${private_ips[$index]})" ssh_ready "${private_ips[$index]}"
done

declare -a osd_devices_by_node
expected_osds=0
for index in "${!node_names[@]}"; do
	if (( index == 0 )); then
		devices="$(paste -sd, /etc/ceph-lab/osd-devices)"
	else
		devices="$(ssh "${ssh_options[@]}" "root@${private_ips[$index]}" 'paste -sd, /etc/ceph-lab/osd-devices')"
	fi
	osd_devices_by_node[index]="${devices}"
	IFS=',' read -r -a node_devices <<<"${devices}"
	if (( ${#node_devices[@]} != data_disk_counts[index] )); then
		errorf 'OSD device count differs: node=%s reported=%d expected=%d' \
			"${node_names[$index]}" "${#node_devices[@]}" "${data_disk_counts[$index]}"
		exit 1
	fi
	expected_osds=$((expected_osds + ${#node_devices[@]}))
done

infof '%s' 'Ceph deployment metadata (Ceph traffic uses private IPs; public IPs are for external access):'
infof '%-24s %-16s %-16s %s' HOSTNAME PRIVATE_IP PUBLIC_IP OSD_DEVICES
for index in "${!node_names[@]}"; do
	infof '%-24s %-16s %-16s %s' \
		"${node_names[$index]}" "${private_ips[$index]}" "${public_ips[$index]}" "${osd_devices_by_node[$index]:-(none)}"
done
infof 'bootstrap host: name=%s private_ip=%s public_ip=%s' "${node_names[0]}" "${private_ips[0]}" "${public_ips[0]}"
infof 'external Dashboard URL after bootstrap: https://%s:8443/' "${public_ips[0]}"
infof 'expected OSD count: %d' "${expected_osds}"

if ceph status >/dev/null 2>&1 && ceph orch status >/dev/null 2>&1; then
	infof '%s' 'ready: existing Ceph cluster and orchestrator; skipping bootstrap'
else
	cephadm bootstrap --mon-ip "${private_ips[0]}"
fi

ceph_cli_ready() {
	ceph status >/dev/null && ceph orch status >/dev/null
}
wait_until "Ceph CLI and orchestrator after bootstrap" ceph_cli_ready
ceph config set global log_to_file true
ceph config set global mon_cluster_log_to_file true
dashboard_password_file="$(mktemp)"
chmod 0600 "${dashboard_password_file}"
printf '%s' "${CEPH_LAB_DASHBOARD_PASSWORD}" >"${dashboard_password_file}"
if ! ceph dashboard ac-user-set-password admin -i "${dashboard_password_file}"; then
	ceph dashboard ac-user-set-password admin "${CEPH_LAB_DASHBOARD_PASSWORD}"
fi
rm -f "${dashboard_password_file}"
ceph mgr services --format json

for index in "${!node_names[@]}"; do
	if (( index == 0 )); then
		continue
	fi
	ssh-copy-id -f -i /etc/ceph/ceph.pub "${ssh_options[@]}" "root@${private_ips[$index]}"
done

cephadm_host_ready() {
	ceph cephadm check-host "$1" >/dev/null
}
orchestrator_has_host() {
	ceph orch host ls --format json |
		jq -e --arg hostname "$1" 'any(.[]; .hostname == $hostname)' >/dev/null
}
for index in "${!node_names[@]}"; do
	if (( index == 0 )); then
		continue
	fi
	if orchestrator_has_host "${node_names[$index]}"; then
		infof 'ready: node %s is already registered with the orchestrator' "${node_names[$index]}"
	else
		ceph orch host add "${node_names[$index]}" "${private_ips[$index]}"
	fi
	wait_until "cephadm prerequisites on ${node_names[$index]}" cephadm_host_ready "${node_names[$index]}"
done

all_hosts_registered() {
	local output
	output="$(ceph orch host ls --format json)" || return 1
	for hostname in "${node_names[@]}"; do
		jq -e --arg hostname "${hostname}" \
			'any(.[]; .hostname == $hostname and ((.status // "") == ""))' <<<"${output}" >/dev/null || return 1
	done
}
wait_until "all ${node_count} hosts to be registered by the orchestrator" all_hosts_registered

device_inventory_ready() {
	local output hostname device devices
	output="$(ceph orch device ls --refresh --format json)" || return 1
	for index in "${!node_names[@]}"; do
		hostname="${node_names[$index]}"
		devices="${osd_devices_by_node[$index]}"
		[[ -n "${devices}" ]] || continue
		IFS=',' read -r -a node_devices <<<"${devices}"
		for device in "${node_devices[@]}"; do
			jq -e --arg hostname "${hostname}" --arg device "${device}" \
				'[
					.[] |
					if has("devices") then
						. as $host |
						.devices[] |
						. + {inventory_hostname: ($host.name // $host.hostname // $host.addr)}
					else
						. + {inventory_hostname: (.hostname // .name // .addr)}
					end
				] |
				any(.[]; .inventory_hostname == $hostname and .path == $device and .available == true)' \
				<<<"${output}" >/dev/null || return 1
		done
	done
}
if (( expected_osds > 0 )); then
	wait_until "all configured data disks to become available in the orchestrator inventory" device_inventory_ready
fi
ceph orch device ls

if (( expected_osds > 0 )); then
	for index in "${!node_names[@]}"; do
		if [[ -n "${osd_devices_by_node[$index]}" ]]; then
			ceph orch daemon add osd "${node_names[$index]}:${osd_devices_by_node[$index]}"
		fi
	done

	osds_ready() {
		local osd_count running_count
		osd_count="$(ceph osd stat --format json | jq -r '.num_osds // 0')" || return 1
		running_count="$(ceph orch ps --daemon-type osd --format json | jq '[.[] | select(.status_desc == "running")] | length')" || return 1
		(( osd_count >= expected_osds && running_count >= expected_osds ))
	}
	wait_until "${expected_osds} OSD daemon(s) to be registered and running" osds_ready
fi

for index in "${!node_names[@]}"; do
	if (( index == 0 )); then
		continue
	fi
	ssh "${ssh_options[@]}" "root@${private_ips[$index]}" 'install -d -m 0755 /etc/ceph'
	scp "${ssh_options[@]}" /etc/ceph/ceph.conf "root@${private_ips[$index]}:/etc/ceph/ceph.conf"
	scp "${ssh_options[@]}" /etc/ceph/ceph.client.admin.keyring "root@${private_ips[$index]}:/etc/ceph/ceph.client.admin.keyring"
done

ceph fs volume create cephfs
cephfs_ready() {
	ceph fs get cephfs --format json |
		jq -e '(.mdsmap.up | length) > 0 and (.mdsmap.in | length) > 0' >/dev/null
}
wait_until "CephFS volume cephfs" cephfs_ready

cluster_healthy() {
	ceph health --format json | jq -e '.status == "HEALTH_OK"' >/dev/null
}
wait_until "Ceph cluster health to become HEALTH_OK" cluster_healthy
ceph status
infof 'Ceph cluster deployment completed: cluster=%s' "${CEPH_LAB_CLUSTER_NAME}"
