#!/usr/bin/env bash
set -Eeuo pipefail

report_error() {
	local status=$?
	echo "ERROR: deploy-ceph.sh failed on ${HOSTNAME:-unknown} at line ${BASH_LINENO[0]}: ${BASH_COMMAND} (exit ${status})" >&2
	exit "${status}"
}
trap report_error ERR

required=(
	CEPH_LAB_CLUSTER_NAME
	CEPH_LAB_BOOTSTRAP_NODE_NAME
	CEPH_LAB_NODE_NAMES
	CEPH_LAB_PUBLIC_IPS
	CEPH_LAB_PRIVATE_IPS
	CEPH_LAB_DATA_DISK_COUNTS
	CEPH_LAB_WAIT_TIMEOUT_SECONDS
)
for name in "${required[@]}"; do
	if [[ -z "${!name:-}" ]]; then
		echo "missing required environment variable: ${name}" >&2
		exit 1
	fi
done

IFS=',' read -r -a node_names <<<"${CEPH_LAB_NODE_NAMES}"
IFS=',' read -r -a public_ips <<<"${CEPH_LAB_PUBLIC_IPS}"
IFS=',' read -r -a private_ips <<<"${CEPH_LAB_PRIVATE_IPS}"
IFS=',' read -r -a data_disk_counts <<<"${CEPH_LAB_DATA_DISK_COUNTS}"
node_count="${#node_names[@]}"
if (( node_count < 3 )); then
	echo "Ceph deployment requires at least 3 nodes; got ${node_count}" >&2
	exit 1
fi
if (( ${#public_ips[@]} != node_count || ${#private_ips[@]} != node_count || ${#data_disk_counts[@]} != node_count )); then
	echo "node metadata counts differ" >&2
	exit 1
fi
if [[ "${CEPH_LAB_BOOTSTRAP_NODE_NAME}" != "${node_names[0]}" ]]; then
	echo "bootstrap node ${CEPH_LAB_BOOTSTRAP_NODE_NAME} is not the first configured node ${node_names[0]}" >&2
	exit 1
fi
current_hostname="$(hostname -s)"
if [[ "${current_hostname}" != "${CEPH_LAB_BOOTSTRAP_NODE_NAME}" ]]; then
	echo "deployment hook must run on first node ${CEPH_LAB_BOOTSTRAP_NODE_NAME}; current host is ${current_hostname}" >&2
	exit 1
fi

wait_until() {
	local description="$1"
	shift
	local deadline=$((SECONDS + CEPH_LAB_WAIT_TIMEOUT_SECONDS))
	while ! "$@"; do
		if (( SECONDS >= deadline )); then
			echo "timed out waiting for ${description}" >&2
			return 1
		fi
		echo "waiting for ${description}; retrying in 10s"
		sleep 10
	done
	echo "ready: ${description}"
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
		echo "${node_names[$index]} reported ${#node_devices[@]} OSD device(s), expected ${data_disk_counts[$index]}" >&2
		exit 1
	fi
	expected_osds=$((expected_osds + ${#node_devices[@]}))
done

echo "Ceph deployment metadata (Ceph traffic uses private IPs; public IPs are for external access):"
printf '%-24s %-16s %-16s %s\n' HOSTNAME PRIVATE_IP PUBLIC_IP OSD_DEVICES
for index in "${!node_names[@]}"; do
	printf '%-24s %-16s %-16s %s\n' "${node_names[$index]}" "${private_ips[$index]}" "${public_ips[$index]}" "${osd_devices_by_node[$index]:-(none)}"
done
echo "bootstrap host: ${node_names[0]} (private ${private_ips[0]}, public ${public_ips[0]})"
echo "external Dashboard URL after bootstrap: https://${public_ips[0]}:8443/"
echo "expected OSD count: ${expected_osds}"

if ceph status >/dev/null 2>&1 && ceph orch status >/dev/null 2>&1; then
	echo "ready: existing Ceph cluster and orchestrator; skipping bootstrap"
else
	cephadm bootstrap --mon-ip "${private_ips[0]}"
fi

ceph_cli_ready() {
	ceph status >/dev/null && ceph orch status >/dev/null
}
wait_until "Ceph CLI and orchestrator after bootstrap" ceph_cli_ready
ceph config set global log_to_file true
ceph config set global mon_cluster_log_to_file true

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
		echo "ready: ${node_names[$index]} is already registered with the orchestrator"
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
echo "Ceph cluster ${CEPH_LAB_CLUSTER_NAME} deployment completed"
