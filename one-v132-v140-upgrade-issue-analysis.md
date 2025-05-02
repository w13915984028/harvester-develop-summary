A few quetions arised on this upgrade scenario.

## Why upgrade is stucking on `Waiting for plan`

1. There are none-matching node and machine objects.


```
machine:  has 5 members: providerID: rke2://*-01, 02, 03, 05, 06
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

### Enhancement

The pre-flight check should also detect this mis-matching of `node` and `machine`.

https://github.com/harvester/harvester/issues/8179

## Why node 04 is not provisioned? TBD

Following log is observed on pod `system-upgrade-controller`

```
2025-01-31T10:16:17.469113115Z time="2025-01-31T10:16:17Z" level=error msg="error syncing 'cattle-system/sync-additional-ca': handler system-upgrade-controller: secrets \"harvester-additional-ca\" not found, handler system-upgrade-controller: failed to create cattle-system/apply-sync-additional-ca-on-*-01-with- batch/v1, Kind=Job for system-upgrade-controller cattle-system/sync-additional-ca: Job.batch \"apply-sync-additional-ca-on-*-01-with-\" is invalid: [metadata.name: Invalid value: \"apply-sync-additional-ca-on-*-01-with-\": a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*'), spec.template.labels: Invalid value: \"apply-sync-additional-ca-on-*-01-with-\": a valid label must be an empty string or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyValue',  or 'my_value',  or '12345', regex used for validation is '(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?')], requeuing"
```

node 01: creationTimestamp: "2025-01-31T10:12:22Z"

node 02: creationTimestamp: "2025-01-31T10:44:58Z"

node 03: creationTimestamp: "2025-01-31T11:02:09Z"

node 04: creationTimestamp: "2025-02-06T08:38:27Z"

node 05: creationTimestamp: "2025-02-03T07:06:42Z"

node 06: creationTimestamp: "2025-01-31T12:44:11Z"


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

```
01  spec:
    podCIDR: 10.52.0.0/24
    podCIDRs:
    - 10.52.0.0/24
    providerID: rke2://*-01


02  spec:
    podCIDR: 10.52.1.0/24
    podCIDRs:
    - 10.52.1.0/24
    providerID: rke2://*-02


03  spec:
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


04  spec:
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


05  spec:
    podCIDR: 10.52.5.0/24
    podCIDRs:
    - 10.52.5.0/24
    providerID: rke2://*-05


06  spec:
    podCIDR: 10.52.4.0/24
    podCIDRs:
    - 10.52.4.0/24
    providerID: rke2://*-06
```
