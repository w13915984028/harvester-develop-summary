A few quetions arised on this upgrade scenario.

## Why upgrade is stucking on `Waiting for plan`

1. There are none-matching node and machine objects.


```
machine (namespace: fleet-local):  has 5 members: providerID: rke2://*-01, 02, 03, 05, 06
- apiVersion: cluster.x-k8s.io/v1beta1
  kind: Machine
  
node: has 6 members, rke2://*-01, 02, 03, 04, 05, 06
- apiVersion: v1
  kind: Node
```

2. It happens to trigger following code bug.

The direct cause of upgrade-manifest is blocking on `2025-04-29T10:25:58.349353382Z Waiting for plan hvst-upgrade-ljgw9-skip-restart-rancher-system-agent to complete...`, is due to following code bug:

https://github.com/harvester/harvester/blob/939857e93c3d97de47f497e719ff219fe4df81ca/package/upgrade/upgrade_manifests.sh#L1223

```
...
apiVersion: upgrade.cattle.io/v1
kind: Plan
metadata:
  name: $plan_name
  namespace: cattle-system
spec:
  concurrency: 10
  nodeSelector:
    matchLabels:
      harvesterhci.io/managed: "true"
...
  # Wait for all nodes complete
  while [ true ]; do
    plan_label="plan.upgrade.cattle.io/$plan_name"
    plan_latest_version=$(kubectl get plans.upgrade.cattle.io "$plan_name" -n cattle-system -ojsonpath="{.status.latestVersion}")

    if [ "$plan_latest_version" = "$plan_version" ]; then
      plan_latest_hash=$(kubectl get plans.upgrade.cattle.io "$plan_name" -n cattle-system -ojsonpath="{.status.latestHash}")
      total_nodes_count=$(kubectl get nodes -o json | jq '.items | length')
      complete_nodes_count=$(kubectl get nodes --selector="plan.upgrade.cattle.io/$plan_name=$plan_latest_hash" -o json | jq '.items | length')

      if [ "$total_nodes_count" = "$complete_nodes_count" ]; then
        echo "Plan $plan_name completes."
        break
      fi
    fi

    echo "Waiting for plan $plan_name to complete..."
    sleep 10
  done
```

It get all nodes first; and check node by the `plan` selecotr.


e.g.: node `*-01` has related plan `plan.upgrade.cattle.io/hvst-upgrade-ljgw9-prepare`.

```
      kubernetes.io/hostname: *f-01
      kubernetes.io/os: linux
...
      node.kubernetes.io/instance-type: rke2
      plan.upgrade.cattle.io/hvst-upgrade-ljgw9-prepare: a03218a5dd880ee3473b6d4861a878d2d6addf1768a893f0c6be9f67
      plan.upgrade.cattle.io/hvst-upgrade-ljgw9-skip-restart-rancher-system-agent: 05c5779f9b99fa8288f7f2854b5efbae373b09b6490ab85543eb7781
```


But node `*-04` does not have the this `plan.upgrade.cattle.io/hvst-upgrade-ljgw9-prepare`. It stucks waiting.


The reason is, quite ticky and rarely, the node `*-01` has no label `harvesterhci.io/managed: "true"`. The value `total_nodes_count` is computing with a wrong assumption.

The node `*-04` was added to the cluster on `"2025-02-06T08:38:27Z"`, but failed to generate a corresponding `machine` object, hence the node object has no chance to be labeled with `harvesterhci.io/managed: "true"`.

### Workaround

#### Remove the latest upgrade object

Follow https://docs.harvesterhci.io/v1.3/upgrade/troubleshooting/#start-over-an-upgrade

List and delete the latest upgrade CR object, wait until it is gone

Example

```
# list the on-going upgrade
$ kubectl get upgrade.harvesterhci.io -n harvester-system -l harvesterhci.io/latestUpgrade=true
NAME                 AGE
hvst-upgrade-9gmg2   10m

# delte the upgrade
$ kubectl delete upgrade.harvesterhci.io/hvst-upgrade-9gmg2 -n harvester-system
```
#### Resume all managedcharts

Run below shell script to resume all managedcharts, which were paused by the upgrade.

https://github.com/harvester/harvester/blob/939857e93c3d97de47f497e719ff219fe4df81ca/package/upgrade/upgrade_manifests.sh#L1358

```
cat > resumeallcharts.sh << 'FOE'
resume_all_charts() {

  local patchfile="/tmp/charttmp.yaml"

  cat >"$patchfile" <<EOF
spec:
  paused: false
EOF
  echo "the to-be-patched file"
  cat "$patchfile"

  local charts="harvester harvester-crd rancher-monitoring-crd rancher-logging-crd"

  for chart in $charts; do
    echo "unapuse managedchart $chart"
    kubectl patch managedcharts.management.cattle.io $chart -n fleet-local --patch-file "$patchfile" --type merge || echo "failed, check reason"
  done

  rm "$patchfile"
}

resume_all_charts

FOE

chmod +x ./resumeallcharts.sh

./resumeallcharts.sh

```

### Enhancement

The pre-flight check should also detect this mis-matching of `node` and `machine`.

https://github.com/harvester/harvester/issues/8179

## Why node 04 is not provisioned? TBD

### Related node and machine object

```
01:
creationTimestamp: "2025-01-31T10:12:22Z"
machine: custom-aaced2378f1a
dataSecretName: custom-aaced2378f1a-machine-bootstrap

node spec:
    podCIDR: 10.52.0.0/24
    podCIDRs:
    - 10.52.0.0/24
    providerID: rke2://*-01


02:
creationTimestamp: "2025-01-31T10:44:58Z"
machine: custom-9b5b97b047c2
dataSecretName: custom-9b5b97b047c2-machine-bootstrap

node spec:
    podCIDR: 10.52.1.0/24
    podCIDRs:
    - 10.52.1.0/24
    providerID: rke2://*-02


03:
creationTimestamp: "2025-01-31T11:02:09Z"
machine: custom-1a473e8e3deb
dataSecretName: custom-1a473e8e3deb-machine-bootstrap

node spec:
    podCIDR: 10.52.3.0/24
    podCIDRs:
    - 10.52.3.0/24
    providerID: rke2://*-03

    conditions:
    - lastHeartbeatTime: "2025-05-02T07:18:21Z"
      lastTransitionTime: "2025-04-29T12:18:06Z"
      message: Node is not a member of the etcd cluster
      reason: NotAMember
      status: "False"
      type: EtcdIsVoter
    - lastHeartbeatTime: "2025-01-31T14:09:26Z"
      lastTransitionTime: "2025-01-31T14:09:26Z"
      message: Flannel is running on this node
      reason: FlannelIsUp
      status: "False"
      type: NetworkUnavailable



04:
creationTimestamp: "2025-02-06T08:38:27Z"
machine: custom-1a473e8e3deb

node spec:
    podCIDR: 10.52.7.0/24
    podCIDRs:
    - 10.52.7.0/24
    providerID: rke2://*-04
    creationTimestamp: "2025-02-06T08:38:27Z"
    conditions:
    - lastHeartbeatTime: "2025-05-02T07:15:51Z"
      lastTransitionTime: "2025-04-29T13:00:51Z"
      message: Node is a voting member of the etcd cluster
      reason: MemberNotLearner
      status: "True"
      type: EtcdIsVoter
    - lastHeartbeatTime: "2025-04-29T13:00:56Z"
      lastTransitionTime: "2025-04-29T13:00:56Z"
      message: Flannel is running on this node
      reason: FlannelIsUp
      status: "False"
      type: NetworkUnavailable



05:
creationTimestamp: "2025-02-03T07:06:42Z"
machine: custom-e4045a947698
custom-e4045a947698-machine-bootstrap

node spec:
    podCIDR: 10.52.5.0/24
    podCIDRs:
    - 10.52.5.0/24
    providerID: rke2://*-05


06:
creationTimestamp: "2025-01-31T12:44:11Z"
machine: custom-e41ee5469acc
custom-e41ee5469acc-machine-bootstrap

node spec:
    podCIDR: 10.52.4.0/24
    podCIDRs:
    - 10.52.4.0/24
    providerID: rke2://*-06
```

### Short summary about node-04

The whole cluster was (re)installed from `Jan 31 10:11:59`, this can be proved by the `node` logs and CRD objects like `node-01` creationTimestamp: `"2025-01-31T10:12:22Z"`.

But, the `node-04`, has log from `Jan 03 06:33:49` and related provision log, it `run until Jan 31 06:49:25 and rebooted`.

Then it `Detected first start, force-applying one-time instruction set` on `Jan 31 11:00:22 *-04 rancher-system-agent[7624]: time="2025-01-31T11:00:22Z" level=info msg="Detected first start, force-applying one-time instruction set"`

`node-04` was not re-installed, but it re-joined the cluster from `"2025-02-06T08:38:27Z"`; however due to it had been provisioned, it has only the related `node` object, but no `machine` object.

Open point: what happend to `node-04` ?

### node 01: Log starts from Jan 31 10:11:59;  the initial node of the cluster

```
Jan 31 10:11:59 *-01 systemd[1]: Starting Rancher Bootstrap...
Jan 31 10:12:00 *-01 rancherd[5052]: time="2025-01-31T10:12:00Z" level=info msg="Loading config file [/usr/share/rancher/rancherd/config.yaml.d/50-defaults.yaml]"
Jan 31 10:12:00 *-01 rancherd[5052]: time="2025-01-31T10:12:00Z" level=info msg="Loading config file [/usr/share/rancher/rancherd/config.yaml.d/91-harvester-bootstrap-repo.yaml]"
...
Jan 31 10:19:04 *-01 rancherd[5052]: time="2025-01-31T10:19:04Z" level=info msg="[stdout]: cluster.provisioning.cattle.io/local condition met"
Jan 31 10:19:04 *-01 rancherd[5052]: time="2025-01-31T10:19:04Z" level=info msg="Successfully Bootstrapped Rancher (v2.8.5/v1.28.12+rke2r1)"
Jan 31 10:19:04 *-01 systemd[1]: rancherd.service: Deactivated successfully.
Jan 31 10:19:04 *-01 systemd[1]: Finished Rancher Bootstrap.
```

### node 02: Log starts from Jan 31 10:44:55; and a node-role promotion (agent->server) is seen on Jan 31 11:00:38 

```
Jan 31 10:44:55 *-02 systemd[1]: Starting Rancher Kubernetes Engine v2 (agent)...
Jan 31 10:44:55 *-02 sh[5538]: + /usr/bin/systemctl is-enabled --quiet nm-cloud-setup.service

...
Jan 31 11:00:38 *-02 systemd[1]: rke2-agent.service: Unit process 57469 (containerd-shim) remains running after unit stopped.
Jan 31 11:00:38 *-02 systemd[1]: Stopped Rancher Kubernetes Engine v2 (agent).

...
Jan 31 11:00:39 *-02 systemd[1]: Starting Rancher Kubernetes Engine v2 (server)...
Jan 31 11:00:39 *-02 sh[59097]: + /usr/bin/systemctl is-enabled --quiet nm-cloud-setup.service
Jan 31 11:00:39 *-02 sh[59100]: Failed to get unit file state for nm-cloud-setup.service: No such file or directory
Jan 31 11:00:39 *-02 harv-update-rke2-server-url[59113]: + HARVESTER_CONFIG_FILE=/oem/harvester.config
```


### node 04: Log starts from `Jan 03 06:33:49` with an ever machine id `custom-b5da48538d24`

The `rancher-system-agent` log.

```
Jan 03 06:33:49 *-04 systemd[1]: Started Rancher System Agent.
Jan 03 06:33:49 *-04 rancher-system-agent[5619]: time="2025-01-03T06:33:49Z" level=info msg="Rancher System Agent version v0.3.6 (41c07d0) is starting"
Jan 03 06:33:49 *-04 rancher-system-agent[5619]: time="2025-01-03T06:33:49Z" level=info msg="Using directory /var/lib/rancher/agent/work for work"
Jan 03 06:33:49 *-04 rancher-system-agent[5619]: time="2025-01-03T06:33:49Z" level=info msg="Starting remote watch of plans"
Jan 03 06:33:49 *-04 rancher-system-agent[5619]: time="2025-01-03T06:33:49Z" level=info msg="Starting /v1, Kind=Secret controller"
Jan 03 06:33:49 *-04 rancher-system-agent[5619]: time="2025-01-03T06:33:49Z" level=info msg="Detected first start, force-applying one-time instruction set"
Jan 03 06:33:49 *-04 rancher-system-agent[5619]: time="2025-01-03T06:33:49Z" level=info msg="[Applyinator] Applying one-time instructions for plan with checksum d3c6a2bc2a72b46122dc3e2706a4186a00e2db1087a40ca1d0c8e96fa9b0b1b0"
Jan 03 06:33:49 *-04 rancher-system-agent[5619]: time="2025-01-03T06:33:49Z" level=info msg="[Applyinator] Extracting image rancher/system-agent-installer-rke2:v1.28.12-rke2r1 to directory /var/lib/rancher/agent/work/20250103-063349/d3c6a2bc2a72b46122dc3e2706a4186a00e2db1087a40ca1d0c8e96fa9b0b1b0_0"
Jan 03 06:33:49 *-04 rancher-system-agent[5619]: time="2025-01-03T06:33:49Z" level=info msg="Checking local image archives in /var/lib/rancher/agent/images for index.docker.io/rancher/system-agent-installer-rke2:v1.28.12-rke2r1"
Jan 03 06:33:50 *-04 rancher-system-agent[5619]: time="2025-01-03T06:33:50Z" level=info msg="Extracting file installer.sh to /var/lib/rancher/agent/work/20250103-063349/d3c6a2bc2a72b46122dc3e2706a4186a00e2db1087a40ca1d0c8e96fa9b0b1b0_0/installer.sh"
Jan 03 06:33:50 *-04 rancher-system-agent[5619]: time="2025-01-03T06:33:50Z" level=info msg="Extracting file rke2.linux-amd64.tar.gz to /var/lib/rancher/agent/work/20250103-063349/d3c6a2bc2a72b46122dc3e2706a4186a00e2db1087a40ca1d0c8e96fa9b0b1b0_0/rke2.linux-amd64.tar.gz"
Jan 03 06:33:50 *-04 rancher-system-agent[5619]: time="2025-01-03T06:33:50Z" level=info msg="Extracting file sha256sum-amd64.txt to /var/lib/rancher/agent/work/20250103-063349/d3c6a2bc2a72b46122dc3e2706a4186a00e2db1087a40ca1d0c8e96fa9b0b1b0_0/sha256sum-amd64.txt"
Jan 03 06:33:50 *-04 rancher-system-agent[5619]: time="2025-01-03T06:33:50Z" level=info msg="Extracting file run.sh to /var/lib/rancher/agent/work/20250103-063349/d3c6a2bc2a72b46122dc3e2706a4186a00e2db1087a40ca1d0c8e96fa9b0b1b0_0/run.sh"
Jan 03 06:33:50 *-04 rancher-system-agent[5619]: time="2025-01-03T06:33:50Z" level=info msg="[Applyinator] Running command: sh [-c run.sh]"

Jan 03 06:33:52 *-04 rancher-system-agent[5619]: time="2025-01-03T06:33:52Z" level=info msg="[Applyinator] Command sh [-c run.sh] finished with err: <nil> and exit code: 0"
Jan 03 06:33:53 *-04 rancher-system-agent[5619]: time="2025-01-03T06:33:53Z" level=info msg="[K8s] updated plan secret fleet-local/custom-b5da48538d24-machine-plan with feedback"
Jan 03 06:33:53 *-04 rancher-system-agent[5619]: time="2025-01-03T06:33:53Z" level=info msg="[K8s] updated plan secret fleet-local/custom-b5da48538d24-machine-plan with feedback"
Jan 03 06:33:53 *-04 rancher-system-agent[5619]: time="2025-01-03T06:33:53Z" level=error msg="[K8s] received secret to process that was older than the last secret operated on. (29874063 vs 29874065)"
Jan 03 06:33:53 *-04 rancher-system-agent[5619]: time="2025-01-03T06:33:53Z" level=error msg="error syncing 'fleet-local/custom-b5da48538d24-machine-plan': handler secret-watch: secret received was too old, requeuing"
...

// run until Jan 31 06:49:25 and rebooted

Jan 31 06:49:25 *-04 systemd[1]: Stopping Rancher System Agent...
Jan 31 06:49:25 *-04 systemd[1]: rancher-system-agent.service: Deactivated successfully.
Jan 31 06:49:25 *-04 systemd[1]: Stopped Rancher System Agent.
-- Boot dc6fc90a2be644e79872c61ef612300c --
Jan 31 10:59:26 *-04 systemd[1]: Started Rancher System Agent.


Jan 31 11:00:22 *-04 rancher-system-agent[7624]: time="2025-01-31T11:00:22Z" level=info msg="Rancher System Agent version v0.3.6 (41c07d0) is starting"
Jan 31 11:00:22 *-04 rancher-system-agent[7624]: time="2025-01-31T11:00:22Z" level=info msg="Using directory /var/lib/rancher/agent/work for work"
Jan 31 11:00:22 *-04 rancher-system-agent[7624]: time="2025-01-31T11:00:22Z" level=info msg="Starting remote watch of plans"
Jan 31 11:00:22 *-04 rancher-system-agent[7624]: time="2025-01-31T11:00:22Z" level=info msg="Starting /v1, Kind=Secret controller"
Jan 31 11:00:22 *-04 rancher-system-agent[7624]: time="2025-01-31T11:00:22Z" level=info msg="Detected first start, force-applying one-time instruction set"
Jan 31 11:00:22 *-04 rancher-system-agent[7624]: time="2025-01-31T11:00:22Z" level=info msg="[Applyinator] Applying one-time instructions for plan with checksum 9f2e2e9045183d0a0c951efbfa065804838047eff56ec5fcaac32c4c44810098"
,,,


...
Feb 06 08:36:53 *-04 rancher-system-agent[110899]: time="2025-02-06T08:36:53Z" level=info msg="Rancher System Agent version v0.3.6 (41c07d0) is starting"
Feb 06 08:36:53 *-04 rancher-system-agent[110899]: time="2025-02-06T08:36:53Z" level=info msg="Using directory /var/lib/rancher/agent/work for work"
Feb 06 08:36:53 *-04 rancher-system-agent[110899]: time="2025-02-06T08:36:53Z" level=info msg="Starting remote watch of plans"
Feb 06 08:39:03 *-04 rancher-system-agent[110899]: time="2025-02-06T08:39:03Z" level=fatal msg="error while connecting to Kubernetes cluster: Get \"https://10.53.98.171/version\": dial tcp 10.53.98.171:443: connect: connection timed out"

...

Feb 06 08:39:03 *-04 systemd[1]: rancher-system-agent.service: Main process exited, code=exited, status=1/FAILURE
Feb 06 08:39:03 *-04 systemd[1]: rancher-system-agent.service: Failed with result 'exit-code'.
Feb 06 08:39:08 *-04 systemd[1]: rancher-system-agent.service: Scheduled restart job, restart counter is at 46.
Feb 06 08:39:08 *-04 systemd[1]: Stopped Rancher System Agent.
Feb 06 08:39:08 *-04 systemd[1]: Started Rancher System Agent.
Feb 06 08:39:08 *-04 rancher-system-agent[115962]: time="2025-02-06T08:39:08Z" level=info msg="Rancher System Agent version v0.3.6 (41c07d0) is starting"
Feb 06 08:39:08 *-04 rancher-system-agent[115962]: time="2025-02-06T08:39:08Z" level=info msg="Using directory /var/lib/rancher/agent/work for work"
Feb 06 08:39:08 *-04 rancher-system-agent[115962]: time="2025-02-06T08:39:08Z" level=info msg="Starting remote watch of plans"
Feb 06 08:39:35 *-04 systemd[1]: Stopping Rancher System Agent...
Feb 06 08:39:35 *-04 systemd[1]: rancher-system-agent.service: Deactivated successfully.
Feb 06 08:39:35 *-04 systemd[1]: Stopped Rancher System Agent.
Feb 06 08:39:36 *-04 systemd[1]: Started Rancher System Agent.
Feb 06 08:39:36 *-04 rancher-system-agent[128619]: time="2025-02-06T08:39:36Z" level=info msg="Rancher System Agent version v0.3.6 (41c07d0) is starting"
Feb 06 08:39:36 *-04 rancher-system-agent[128619]: time="2025-02-06T08:39:36Z" level=info msg="Using directory /var/lib/rancher/agent/work for work"
Feb 06 08:39:36 *-04 rancher-system-agent[128619]: time="2025-02-06T08:39:36Z" level=info msg="Starting remote watch of plans"
Feb 06 08:39:36 *-04 rancher-system-agent[128619]: time="2025-02-06T08:39:36Z" level=info msg="Starting /v1, Kind=Secret controller"


Apr 29 10:05:09 *-04 rancher-system-agent[128619]: time="2025-04-29T10:05:09Z" level=fatal msg="[K8s] received nil secret that was nil, stopping"
Apr 29 10:05:09 *-04 systemd[1]: rancher-system-agent.service: Main process exited, code=exited, status=1/FAILURE
Apr 29 10:05:09 *-04 systemd[1]: rancher-system-agent.service: Failed with result 'exit-code'.
Apr 29 10:05:14 *-04 systemd[1]: rancher-system-agent.service: Scheduled restart job, restart counter is at 1.
Apr 29 10:05:14 *-04 systemd[1]: Stopped Rancher System Agent.
Apr 29 10:05:14 *-04 systemd[1]: Started Rancher System Agent.
Apr 29 10:05:14 *-04 rancher-system-agent[59140]: time="2025-04-29T10:05:14Z" level=info msg="Rancher System Agent version v0.3.6 (41c07d0) is starting"
Apr 29 10:05:14 *-04 rancher-system-agent[59140]: time="2025-04-29T10:05:14Z" level=info msg="Using directory /var/lib/rancher/agent/work for work"
Apr 29 10:05:14 *-04 rancher-system-agent[59140]: time="2025-04-29T10:05:14Z" level=info msg="Starting remote watch of plans"
Apr 29 10:05:14 *-04 rancher-system-agent[59140]: time="2025-04-29T10:05:14Z" level=fatal msg="error while connecting to Kubernetes cluster: the server has asked for the client to provide credentials"



May 02 11:28:49 *-04 systemd[1]: Stopped Rancher System Agent.
May 02 11:28:49 *-04 systemd[1]: Started Rancher System Agent.
May 02 11:28:49 *-04 rancher-system-agent[126108]: time="2025-05-02T11:28:49Z" level=info msg="Rancher System Agent version v0.3.6 (41c07d0) is starting"
May 02 11:28:49 *-04 rancher-system-agent[126108]: time="2025-05-02T11:28:49Z" level=info msg="Using directory /var/lib/rancher/agent/work for work"
May 02 11:28:49 *-04 rancher-system-agent[126108]: time="2025-05-02T11:28:49Z" level=info msg="Starting remote watch of plans"
May 02 11:28:49 *-04 rancher-system-agent[126108]: time="2025-05-02T11:28:49Z" level=fatal msg="error while connecting to Kubernetes cluster: the server has asked for the client to provide credentials"
May 02 11:28:49 *-04 systemd[1]: rancher-system-agent.service: Main process exited, code=exited, status=1/FAILURE
May 02 11:28:49 *-04 systemd[1]: rancher-system-agent.service: Failed with result 'exit-code'.
May 02 11:28:54 *-04 systemd[1]: rancher-system-agent.service: Scheduled restart job, restart counter is at 48549.
May 02 11:28:54 *-04 systemd[1]: Stopped Rancher System Agent.
May 02 11:28:54 *-04 systemd[1]: Started Rancher System Agent.
May 02 11:28:54 *-04 rancher-system-agent[126317]: time="2025-05-02T11:28:54Z" level=info msg="Rancher System Agent version v0.3.6 (41c07d0) is starting"
May 02 11:28:54 *-04 rancher-system-agent[126317]: time="2025-05-02T11:28:54Z" level=info msg="Using directory /var/lib/rancher/agent/work for work"
May 02 11:28:54 *-04 rancher-system-agent[126317]: time="2025-05-02T11:28:54Z" level=info msg="Starting remote watch of plans"
May 02 11:28:54 *-04 rancher-system-agent[126317]: time="2025-05-02T11:28:54Z" level=fatal msg="error while connecting to Kubernetes cluster: the server has asked for the client to provide credentials"
May 02 11:28:54 *-04 systemd[1]: rancher-system-agent.service: Main process exited, code=exited, status=1/FAILURE
May 02 11:28:54 *-04 systemd[1]: rancher-system-agent.service: Failed with result 'exit-code'.
May 02 11:28:59 *-04 systemd[1]: rancher-system-agent.service: Scheduled restart job, restart counter is at 48550.
May 02 11:28:59 *-04 systemd[1]: Stopped Rancher System Agent.
```

### Others

Following log is observed on pod `system-upgrade-controller`

```
2025-01-31T10:16:17.469113115Z time="2025-01-31T10:16:17Z" level=error msg="error syncing 'cattle-system/sync-additional-ca': handler system-upgrade-controller: secrets \"harvester-additional-ca\" not found, handler system-upgrade-controller: failed to create cattle-system/apply-sync-additional-ca-on-*-01-with- batch/v1, Kind=Job for system-upgrade-controller cattle-system/sync-additional-ca: Job.batch \"apply-sync-additional-ca-on-*-01-with-\" is invalid: [metadata.name: Invalid value: \"apply-sync-additional-ca-on-*-01-with-\": a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*'), spec.template.labels: Invalid value: \"apply-sync-additional-ca-on-*-01-with-\": a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')], requeuing"
```

```
node 01: creationTimestamp: "2025-01-31T10:12:22Z"

node 02: creationTimestamp: "2025-01-31T10:44:58Z"

node 03: creationTimestamp: "2025-01-31T11:02:09Z"

node 04: creationTimestamp: "2025-02-06T08:38:27Z"

node 05: creationTimestamp: "2025-02-03T07:06:42Z"

node 06: creationTimestamp: "2025-01-31T12:44:11Z"
```

## Why node 03 is softly `Down` and 04 was triggered promotion? TBD

### node 03 kubelet log:

node 03 failed to update self

```
E0429 12:17:59.828138  155021 controller.go:146] "Failed to ensure lease exists, will retry" err="Get \"https://127.0.0.1:6443/apis/coordination.k8s.io/v1/namespaces/kube-node-lease/leases/*-03?timeout=10s\": dial tcp 127.0.0.1:6443: connect: connection refused" interval="200ms"

E0429 12:18:00.029544  155021 controller.go:146] "Failed to ensure lease exists, will retry" err="Get \"https://127.0.0.1:6443/apis/coordination.k8s.io/v1/namespaces/kube-node-lease/leases/*-03?timeout=10s\": dial tcp 127.0.0.1:6443: connect: connection refused" interval="400ms"
E0429 12:18:00.430683  155021 controller.go:146] "Failed to ensure lease exists, will retry" err="Get \"https://127.0.0.1:6443/apis/coordination.k8s.io/v1/namespaces/kube-node-lease/leases/*-03?timeout=10s\": dial tcp 127.0.0.1:6443: connect: connection refused" interval="800ms"


E0429 12:22:17.483041  155021 controller.go:146] "Failed to ensure lease exists, will retry" err="Get \"https://127.0.0.1:6443/apis/coordination.k8s.io/v1/namespaces/kube-node-lease/leases/*-03?timeout=10s\": dial tcp 127.0.0.1:6443: connect: connection refused" interval="7s"
E0429 12:22:18.189924  155021 kubelet_node_status.go:540] "Error updating node status, will retry" err="error getting node \"*-03\": Get \"https://127.0.0.1:6443/api/v1/nodes/*-03?resourceVersion=0&timeout=10s\": dial tcp 127.0.0.1:6443: connect: connection refused"
E0429 12:22:18.190572  155021 kubelet_node_status.go:540] "Error updating node status, will retry" err="error getting node \"*-03\": Get \"https://127.0.0.1:6443/api/v1/nodes/*-03?timeout=10s\": dial tcp 127.0.0.1:6443: connect: connection refused"
E0429 12:22:18.191165  155021 kubelet_node_status.go:540] "Error updating node status, will retry" err="error getting node \"*-03\": Get \"https://127.0.0.1:6443/api/v1/nodes/*-03?timeout=10s\": dial tcp 127.0.0.1:6443: connect: connection refused"
E0429 12:22:18.191633  155021 kubelet_node_status.go:540] "Error updating node status, will retry" err="error getting node \"*-03\": Get \"https://127.0.0.1:6443/api/v1/nodes/*-03?timeout=10s\": dial tcp 127.0.0.1:6443: connect: connection refused"
E0429 12:22:18.192099  155021 kubelet_node_status.go:540] "Error updating node status, will retry" err="error getting node \"*-03\": Get \"https://127.0.0.1:6443/api/v1/nodes/*-03?timeout=10s\": dial tcp 127.0.0.1:6443: connect: connection refused"
E0429 12:22:18.192118  155021 kubelet_node_status.go:527] "Unable to update node status" err="update node status exceeds retry count"
```

### rancher pod log:
```
2025-04-29T12:18:28.119988851Z 2025/04/29 12:18:28 [ERROR] Failed to handle tunnel request from remote address 10.52.3.39:44436: response 401: failed authentication

2025-04-29T12:18:33.122987969Z 2025/04/29 12:18:33 [ERROR] Failed to handle tunnel request from remote address 10.52.3.39:42690: response 401: failed authentication
2025-04-29T12:24:23.345290388Z 2025/04/29 12:24:23 [ERROR] Failed to handle tunnel request from remote address 10.52.3.39:40860: response 401: failed authentication
```


### Node 03 lost etcd member role and node 04 take it


