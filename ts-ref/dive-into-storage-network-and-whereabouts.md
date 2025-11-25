# Dive into Harvester Storage Network and Whereabouts

## StorageNetwork NAD

From Harvester UI, set the storage-network, with:

```
clusternetwork: mgmt or others
vid: 50 or any
subnet: 192.168.50.0/24
exclude list: 192.168.50.1/32, 192.168.50.2/32

```

The generated NAD is like blow:

```yaml
- apiVersion: k8s.cni.cncf.io/v1
  kind: NetworkAttachmentDefinition
  metadata:
    annotations:
      storage-network.settings.harvesterhci.io: "true"
    creationTimestamp: "2025-11-24T18:37:28Z"
    finalizers:
    - wrangler.cattle.io/harvester-network-manager-nad-controller
    generateName: storagenetwork-
    generation: 1
    labels:
      network.harvesterhci.io/clusternetwork: mgmt
      network.harvesterhci.io/ready: "true"
      network.harvesterhci.io/type: L2VlanNetwork
      network.harvesterhci.io/vlan-id: "50"
      storage-network.settings.harvesterhci.io/hash: e07026b7f0f804be5c0745815c55d902f8e867a5
    name: storagenetwork-sg8d8
    namespace: harvester-system
    resourceVersion: "1196655"
    uid: 5ee16934-4292-44fe-a824-fa343611e61f
  spec:
    config: '{"cniVersion":"0.3.1","type":"bridge","bridge":"mgmt-br","promiscMode":true,"vlan":50,"ipam":{"type":"whereabouts","range":"192.168.50.0/24","exclude":["192.168.50.1/32","192.168.50.2/32"]}}'
```

The LH pod will have following `network-status` if it is successfully replaced:

```yaml
$ kubectl get pods -n longhorn-system instance-manager-6b77fd6b240ff0e6d0fe88c705b51f7f -oyaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    cni.projectcalico.org/containerID: 517151fa292bdcf9522e24fb05f1a99bb96db0a37fb647edb5a146c72b1a5017
    cni.projectcalico.org/podIP: 10.52.0.67/32
    cni.projectcalico.org/podIPs: 10.52.0.67/32
    k8s.v1.cni.cncf.io/network-status: |-
      [{
          "name": "k8s-pod-network",
          "ips": [
              "10.52.0.67"
          ],
          "default": true,
          "dns": {}
      },{
          "name": "harvester-system/storagenetwork-58mfl",
          "interface": "lhnet1",
          "ips": [
              "192.168.50.4"
          ],
          "mac": "da:3f:fb:7a:e2:eb",
          "dns": {}
      }]
    k8s.v1.cni.cncf.io/networks: '[{"namespace": "harvester-system", "name": "storagenetwork-58mfl",
      "interface": "lhnet1"}]'
    longhorn.io/last-applied-tolerations: '[{"key":"kubevirt.io/drain","operator":"Exists","effect":"NoSchedule"}]'
  creationTimestamp: "2025-11-24T12:49:24Z"
...  
  name: instance-manager-6b77fd6b240ff0e6d0fe88c705b51f7f
  namespace: longhorn-system
```


## Whereabouts

### CRDs and lifecycle managment

#### IPPool

It is mapping to NAD config's `ipam`.

```yaml
$ kubectl get ippools.whereabouts.cni.cncf.io -A
NAMESPACE     NAME              AGE
kube-system   192.168.50.0-24   82s


$ kubectl get ippools.whereabouts.cni.cncf.io -A -oyaml
apiVersion: v1
items:
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: IPPool
  metadata:
    creationTimestamp: "2025-11-24T12:49:09Z"
    generation: 3
    name: 192.168.50.0-24
    namespace: kube-system
    resourceVersion: "1120303"
    uid: 5597ce22-4846-4c33-8608-d13fe37f850e
  spec:
    allocations:
      "3":
        id: 12ff38588c143881bd5e6b5b13dad5e89b869c65f311c4a3ed288cd2abfae97e
        ifname: lhnet1
        podref: longhorn-system/bim-d3feb4b38b3614c5004a123b869edfa1bad8ecfef07ef323701c4433cea8c58b
      "4":
        id: 517151fa292bdcf9522e24fb05f1a99bb96db0a37fb647edb5a146c72b1a5017
        ifname: lhnet1
        podref: longhorn-system/instance-manager-6b77fd6b240ff0e6d0fe88c705b51f7f
    range: 192.168.50.0/24
kind: List
metadata:
  resourceVersion: ""
```

:::note

Whereabouts has no webhook to check the `IPPool` object creation, it can be created by third-party directly.

:::

##### IPPool lifecycle managment

Unlike many k8s CRDs and controllers:

 The `IPPool` object is not created when NAD is created, but when a POD allocates IP, then controller checks if `IPPool` is there, if not then creates it.

 The `IPPool` object is not exlicitly removed by whereabouts controller.

 After the `storage-network` is disabled, the `IPPool` is still left. It introduces some challenges on Harvester.


#### overlappingrangeipreservations

The `overlappingrangeipreservations` records the IP allocation with the `podref`.

It comes from `IPPool.spec.allocations[]`.

As the name implies, the object is mainly used to check the `overlapping` between different `IPPools`.

On Harvester, the `storage-network` and `migration-network` can run in parallel (each can have at most one instance), and the webhook ensures those two `IPPools` won't overlap.

```yaml
$ kubectl get overlappingrangeipreservations.whereabouts.cni.cncf.io -A
NAMESPACE     NAME           AGE
kube-system   192.168.50.3   4m9s
kube-system   192.168.50.4   3m54s

$ kubectl get overlappingrangeipreservations.whereabouts.cni.cncf.io -A -oyaml
apiVersion: v1
items:
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: OverlappingRangeIPReservation
  metadata:
    creationTimestamp: "2025-11-24T12:49:09Z"
    generation: 1
    name: 192.168.50.3
    namespace: kube-system
    resourceVersion: "1120106"
    uid: 0883f7ca-c046-467b-82b4-97e3d1362a33
  spec:
    ifname: lhnet1
    podref: longhorn-system/bim-d3feb4b38b3614c5004a123b869edfa1bad8ecfef07ef323701c4433cea8c58b
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: OverlappingRangeIPReservation
  metadata:
    creationTimestamp: "2025-11-24T12:49:24Z"
    generation: 1
    name: 192.168.50.4
    namespace: kube-system
    resourceVersion: "1120304"
    uid: d5753207-dc40-4777-ae8d-ed43e6462329
  spec:
    ifname: lhnet1
    podref: longhorn-system/instance-manager-6b77fd6b240ff0e6d0fe88c705b51f7f
kind: List
metadata:
  resourceVersion: ""
```

:::note

Whereabouts has no webhook to check the `overlappingrangeipreservations` object creation, it can be created by third-party directly.

:::

##### lifecycle managment

Triggered by POD creation and deletion.


### IP Leaking

The following cases will be tested.

Per official documents, the IP leaking seems not avoidable.

[Known limitations](https://github.com/k8snetworkplumbingwg/whereabouts/blob/master/README.md#known-limitations)

```
A hard system crash on a node might leave behind stranded IP allocations, so if you have a trashing system, this might exhaust IPs.
Potentially we need an operator to ensure data is clean, even if just at some kind of interval (e.g. with a cron job)

There's probably a lot of comparison of IP addresses that could be optimized, lots of string conversion.
The etcd method has a number of limitations, in that it uses an all ASCII methodology. If this was binary, it could probably store more and have more efficient IP address comparison.
Unlikely to work in Canada, apparently it would have to be "where aboots?" for Canadians to be able to operate it.
In case of wide IPv6 CIDRs (range≤/64) only the first /65 range is addressable by Whereabouts due to uint64 offset calculation.
```

IP Reconciliation[https://github.com/k8snetworkplumbingwg/whereabouts/blob/master/doc/extended-configuration.md#ip-reconciliation]

```

Whereabouts includes a tool which is intended be run as a k8s CronJob. This utility scans the currently allocated IP addresses, and reconciles them against the currently running pods, and deallocates IP addresses which have been left stranded. Stranded IP addresses can occur due to node failures (e.g. a sudden power off / reboot event) or potentially from pods that have been force deleted (e.g. kubectl delete pod foo --grace-period=0 --force)

A reference deployment of this tool is available in the /docs/ip-reconcilier-job.yaml file.

```

We will try following ways to reproduce them.

#### Forcely deleted PODs

Forcely deleting PODs can affect whereabouts, but if the pod is using fixed name and it is replaced, then it is fine. But if PODs are using dynamic name, then it still can cause `ip leaking`.

#### Node Down and Removal

1. Setup a two-node cluster, enable addons like `rancher-monigoring`, `rancher-logging`

```sh
$ kubectl get nodes -A
NAME     STATUS   ROLES                       AGE   VERSION
harv2    Ready    <none>                      21d   v1.33.3+rke2r1
harv21   Ready    control-plane,etcd,master   32d   v1.33.3+rke2r1
```

2. Set the SN

There are 4 IP allocation records by default.

```sh
$ kubectl get overlappingrangeipreservations.whereabouts.cni.cncf.io -A -oyaml
apiVersion: v1
items:
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: OverlappingRangeIPReservation
  metadata:
    creationTimestamp: "2025-11-25T11:07:11Z"
    generation: 1
    name: 192.168.50.3
    namespace: kube-system
    resourceVersion: "1650921"
    uid: 3a3ab54b-6e4e-4681-9a46-f3f061e00c89
  spec:
    ifname: lhnet1
    podref: longhorn-system/bim-4ba2e66a1eda4c99916dc8ffe6a770c5a54d7602539e13e9847c0fd33d907432
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: OverlappingRangeIPReservation
  metadata:
    creationTimestamp: "2025-11-25T11:07:12Z"
    generation: 1
    name: 192.168.50.4
    namespace: kube-system
    resourceVersion: "1651009"
    uid: 30e405e4-9aab-46fe-b97b-ba572113dac7
  spec:
    ifname: lhnet1
    podref: longhorn-system/bim-d29cffa6648c18cdf95fd9da32107caca582ea863f421b909ed8357193879358
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: OverlappingRangeIPReservation
  metadata:
    creationTimestamp: "2025-11-25T11:07:13Z"
    generation: 1
    name: 192.168.50.5
    namespace: kube-system
    resourceVersion: "1651019"
    uid: 335e0f49-b042-4a0d-9a06-d5a29b76e390
  spec:
    ifname: lhnet1
    podref: longhorn-system/instance-manager-f6e4865355e8386452e45fc74a445e86
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: OverlappingRangeIPReservation
  metadata:
    creationTimestamp: "2025-11-25T11:07:21Z"
    generation: 1
    name: 192.168.50.6
    namespace: kube-system
    resourceVersion: "1651232"
    uid: c6dde06b-591f-4de5-9215-ce8ee7fc0f11
  spec:
    ifname: lhnet1
    podref: longhorn-system/instance-manager-ac9a7c78000218e7ce4a91ad97a99787
kind: List
metadata:
  resourceVersion: ""


$ kubectl  get overlappingrangeipreservations.whereabouts.cni.cncf.io -A
NAMESPACE     NAME           AGE
kube-system   192.168.50.3   19s
kube-system   192.168.50.4   18s
kube-system   192.168.50.5   17s
kube-system   192.168.50.6   9s

$ kubectl get pods -n longhorn-system  -owide
NAME                                                                   READY   STATUS    RESTARTS       AGE    IP           NODE     NOMINATED NODE   READINESS GATES
bim-4ba2e66a1eda4c99916dc8ffe6a770c5a54d7602539e13e9847c0fd33d907432   1/1     Running   0              100s   10.52.0.73   harv21   <none>           <none>
bim-d29cffa6648c18cdf95fd9da32107caca582ea863f421b909ed8357193879358   1/1     Running   0              99s    10.52.1.25   harv2    <none>           <none>
csi-attacher-d6c499464-kr24s                                           1/1     Running   30 (15m ago)   32d    10.52.0.29   harv21   <none>           <none>
csi-attacher-d6c499464-l6bwp                                           1/1     Running   34 (15m ago)   32d    10.52.0.36   harv21   <none>           <none>
csi-attacher-d6c499464-pr686                                           1/1     Running   32 (15m ago)   32d    10.52.0.22   harv21   <none>           <none>
csi-provisioner-8664fc989f-cpdrj                                       1/1     Running   33 (15m ago)   32d    10.52.0.40   harv21   <none>           <none>
csi-provisioner-8664fc989f-l8h29                                       1/1     Running   34 (15m ago)   32d    10.52.0.4    harv21   <none>           <none>
csi-provisioner-8664fc989f-qfjxf                                       1/1     Running   33 (15m ago)   32d    10.52.0.38   harv21   <none>           <none>
csi-resizer-5c59b9c497-5vpk6                                           1/1     Running   33 (15m ago)   32d    10.52.0.15   harv21   <none>           <none>
csi-resizer-5c59b9c497-68p9g                                           1/1     Running   33 (15m ago)   32d    10.52.0.49   harv21   <none>           <none>
csi-resizer-5c59b9c497-tc7w8                                           1/1     Running   32 (15m ago)   32d    10.52.0.57   harv21   <none>           <none>
csi-snapshotter-9869b4cf9-66xhm                                        1/1     Running   33 (15m ago)   32d    10.52.0.55   harv21   <none>           <none>
csi-snapshotter-9869b4cf9-pjgws                                        1/1     Running   32 (15m ago)   32d    10.52.0.37   harv21   <none>           <none>
csi-snapshotter-9869b4cf9-qg26c                                        1/1     Running   32 (15m ago)   32d    10.52.0.59   harv21   <none>           <none>
engine-image-ei-26bab25d-9jsrb                                         1/1     Running   12 (14m ago)   21d    10.52.1.3    harv2    <none>           <none>
engine-image-ei-26bab25d-m5qfz                                         1/1     Running   14 (17m ago)   32d    10.52.0.53   harv21   <none>           <none>
instance-manager-ac9a7c78000218e7ce4a91ad97a99787                      1/1     Running   0              89s    10.52.0.74   harv21   <none>           <none>
instance-manager-f6e4865355e8386452e45fc74a445e86                      1/1     Running   0              99s    10.52.1.26   harv2    <none>           <none>
longhorn-csi-plugin-7srxn                                              3/3     Running   64 (15m ago)   32d    10.52.0.14   harv21   <none>           <none>
longhorn-csi-plugin-dxkxn                                              3/3     Running   36 (14m ago)   21d    10.52.1.2    harv2    <none>           <none>
longhorn-driver-deployer-6849f7674f-n4lmz                              1/1     Running   17 (17m ago)   32d    10.52.0.5    harv21   <none>           <none>
longhorn-manager-7ctxz                                                 2/2     Running   29 (17m ago)   32d    10.52.0.64   harv21   <none>           <none>
longhorn-manager-kfpq6                                                 2/2     Running   24 (14m ago)   21d    10.52.1.15   harv2    <none>           <none>
longhorn-ui-69bd699664-kzlg8                                           1/1     Running   37 (15m ago)   32d    10.52.0.17   harv21   <none>           <none>
longhorn-ui-69bd699664-z7662                                           1/1     Running   38 (15m ago)   32d    10.52.0.6    harv21   <none>           <none>
```

3. Poweroff node `harv2`

This simulates the outage/broken nodes on the running cluster. LH SN related pods are on `Terminating` status on the off node.

```sh
$ kubectl get pods -n longhorn-system -owide
NAME                                                                   READY   STATUS        RESTARTS       AGE   IP           NODE     NOMINATED NODE   READINESS GATES
bim-4ba2e66a1eda4c99916dc8ffe6a770c5a54d7602539e13e9847c0fd33d907432   1/1     Running       0              10m   10.52.0.70   harv21   <none>           <none>
bim-d29cffa6648c18cdf95fd9da32107caca582ea863f421b909ed8357193879358   1/1     Terminating   0              29m   10.52.1.25   harv2    <none>           <none>
...
instance-manager-ac9a7c78000218e7ce4a91ad97a99787                      1/1     Running       0              10m   10.52.0.69   harv21   <none>           <none>
instance-manager-f6e4865355e8386452e45fc74a445e86                      1/1     Terminating   0              29m   10.52.1.26   harv2    <none>           <none>


$ kubectl get overlappingrangeipreservations.whereabouts.cni.cncf.io -A
NAMESPACE     NAME           AGE
kube-system   192.168.50.3   10m
kube-system   192.168.50.4   29m
kube-system   192.168.50.5   29m
kube-system   192.168.50.6   10m
```

4. Delete the `node` from Harvester UI.

Monitor the node's paring `machine` object, if it complains like below, help to kill the related PoDs to let the task moves ahead.

```yaml
$ kubectl get machine -n fleet-local custom-07f7d4292bc8 -oyaml
apiVersion: cluster.x-k8s.io/v1beta1
kind: Machine
metadata:
  clusterName: local
  infrastructureRef:
    apiVersion: rke.cattle.io/v1
    kind: CustomMachine
    name: custom-07f7d4292bc8
    namespace: fleet-local
  nodeDeletionTimeout: 10s
  providerID: rke2://harv2
...
    - lastTransitionTime: "2025-11-25T11:38:11Z"
      message: |-
        Drain not completed yet (started at 2025-11-25T11:38:11Z):
        * Pod cattle-monitoring-system/prometheus-rancher-monitoring-prometheus-0: deletionTimestamp set, but still not removed from the Node
      observedGeneration: 4
      reason: DrainingNode
      status: "True"
      type: Deleting
```

5. Disable SN

It takes time to wait until some PVCs are fully detached.

```yaml
$ kubectl get settings.harvesterhci storage-network -oyaml
apiVersion: harvesterhci.io/v1beta1
kind: Setting
metadata:
  annotations:
    storage-network.settings.harvesterhci.io/hash: da39a3ee5e6b4b0d3255bfef95601890afd80709
    storage-network.settings.harvesterhci.io/net-attach-def: ""
    storage-network.settings.harvesterhci.io/old-net-attach-def: harvester-system/storagenetwork-q6pt5
  creationTimestamp: "2025-10-23T19:43:15Z"
  generation: 17
  name: storage-network
  resourceVersion: "1678241"
  uid: a1e42ea5-1570-4223-a78e-88321b83ff92
status:
  conditions:
  - lastUpdateTime: "2025-11-25T11:42:23Z"
    message: 'waiting for all volumes detached: pvc-2bdb794c-9652-4495-a633-ecbd1366fb21,pvc-3314ab6c-2d83-4b3e-adfd-ab92c71ce453'
    reason: In Progress
    status: "False"
    type: configured
```

6. SN is fully stopped, whereabouts has leftover records

```yaml
$ kubectl get settings.harvesterhci storage-network -oyaml
apiVersion: harvesterhci.io/v1beta1
kind: Setting
metadata:
  annotations:
    storage-network.settings.harvesterhci.io/hash: da39a3ee5e6b4b0d3255bfef95601890afd80709
    storage-network.settings.harvesterhci.io/net-attach-def: ""
    storage-network.settings.harvesterhci.io/old-net-attach-def: ""
  creationTimestamp: "2025-10-23T19:43:15Z"
  generation: 21
  name: storage-network
  resourceVersion: "1682123"
  uid: a1e42ea5-1570-4223-a78e-88321b83ff92
status:
  conditions:
  - lastUpdateTime: "2025-11-25T11:47:46Z"
    reason: Completed
    status: "True"
    type: configured


$ kubectl  get overlappingrangeipreservations.whereabouts.cni.cncf.io -A
NAMESPACE     NAME           AGE
kube-system   192.168.50.4   43m
kube-system   192.168.50.5   43m

$ kubectl get ippools.whereabouts.cni.cncf.io -A -oyaml
apiVersion: v1
items:
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: IPPool
  metadata:
    creationTimestamp: "2025-11-25T11:07:11Z"
    generation: 11
    name: 192.168.50.0-24
    namespace: kube-system
    resourceVersion: "1682205"
    uid: 86282296-9e71-48b6-a54d-3c6071bc8edb
  spec:
    allocations:
      "4":
        id: bbfd444f3a601e59e3780cd8ef30ed2c332e28893279b773518b8bc32d7a5f15
        ifname: lhnet1
        podref: longhorn-system/bim-d29cffa6648c18cdf95fd9da32107caca582ea863f421b909ed8357193879358
      "5":
        id: dca28ab564df675a568a004a8f35edf0e49748c806a4e1180f152bbd07b5f0b7
        ifname: lhnet1
        podref: longhorn-system/instance-manager-f6e4865355e8386452e45fc74a445e86
    range: 192.168.50.0/24
kind: List
metadata:
  resourceVersion: ""


$ kubectl  get pods -n  longhorn-system
NAME                                                                   READY   STATUS    RESTARTS       AGE
bim-4ba2e66a1eda4c99916dc8ffe6a770c5a54d7602539e13e9847c0fd33d907432   1/1     Running   0              3m24s
csi-attacher-d6c499464-kr24s                                           1/1     Running   32 (24m ago)   32d
...
instance-manager-ac9a7c78000218e7ce4a91ad97a99787                      1/1     Running   0              3m24s
...

```


7. Whereabouts PoD log

The removal of NAD impacts whereabouts, but how can Harvester know when is the best time to remove NAD?


```sh
$ kubectl logs -n kube-system harvester-whereabouts-mw4sc
Done configuring CNI.  Sleep=false
2025-11-25T11:25:11Z [debug] Filtering pods with filter key 'spec.nodeName' and filter value 'harv21'
2025-11-25T11:25:11Z [verbose] pod controller created
2025-11-25T11:25:11Z [verbose] Starting informer factories ...
2025-11-25T11:25:11Z [verbose] Informer factories started
2025-11-25T11:25:11Z [verbose] starting network controller
2025-11-25T11:25:12Z [verbose] using expression: 30 4 * * *
E1125 11:25:30.891411      30 reflector.go:158] "Unhandled Error" err="github.com/k8snetworkplumbingwg/whereabouts/pkg/generated/informers/externalversions/factory.go:140: Failed to watch *v1alpha1.IPPool: unknown (get ippools.whereabouts.cni.cncf.io)" logger="UnhandledError"
E1125 11:25:50.158193      30 reflector.go:158] "Unhandled Error" err="github.com/k8snetworkplumbingwg/whereabouts/pkg/generated/informers/externalversions/factory.go:140: Failed to watch *v1alpha1.IPPool: unknown (get ippools.whereabouts.cni.cncf.io) - error from a previous attempt: read tcp 192.168.122.121:40298->10.53.0.1:443: read: connection reset by peer" logger="UnhandledError"
E1125 11:25:50.158228      30 reflector.go:158] "Unhandled Error" err="k8s.io/client-go/informers/factory.go:160: Failed to watch *v1.Pod: unknown (get pods) - error from a previous attempt: read tcp 192.168.122.121:40280->10.53.0.1:443: read: connection reset by peer" logger="UnhandledError"
E1125 11:25:50.158256      30 reflector.go:158] "Unhandled Error" err="github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/client/informers/externalversions/factory.go:117: Failed to watch *v1.NetworkAttachmentDefinition: unknown (get network-attachment-definitions.k8s.cni.cncf.io) - error from a previous attempt: read tcp 192.168.122.121:40282->10.53.0.1:443: read: connection reset by peer" logger="UnhandledError"
2025-11-25T11:26:55Z [verbose] deleted pod [longhorn-system/bim-4ba2e66a1eda4c99916dc8ffe6a770c5a54d7602539e13e9847c0fd33d907432]
2025-11-25T11:26:55Z [verbose] deleted pod [longhorn-system/instance-manager-ac9a7c78000218e7ce4a91ad97a99787]
2025-11-25T11:26:55Z [verbose] skipped net-attach-def for default network
2025-11-25T11:26:55Z [debug] pod's network status: {Name:harvester-system/storagenetwork-q6pt5 Interface:lhnet1 IPs:[192.168.50.3] Mac:ce:06:35:17:63:f4 Mtu:0 Default:false DNS:{Nameservers:[] Domain: Search:[] Options:[]} DeviceInfo:<nil> Gateway:[]}
2025-11-25T11:26:55Z [verbose] the NAD's config: {{"cniVersion":"0.3.1","type":"bridge","bridge":"mgmt-br","promiscMode":true,"vlan":50,"ipam":{"type":"whereabouts","range":"192.168.50.0/24","exclude":["192.168.50.1/32","192.168.50.2/32"]}}}
2025-11-25T11:26:55Z [debug] Used defaults from parsed flat file config @ /host/etc/cni/net.d/whereabouts.d/whereabouts.conf
2025-11-25T11:26:55Z [verbose] pool range [192.168.50.0/24]
2025-11-25T11:26:55Z [verbose] result of garbage collecting pods: <nil>
2025-11-25T11:26:55Z [verbose] skipped net-attach-def for default network
2025-11-25T11:26:55Z [debug] pod's network status: {Name:harvester-system/storagenetwork-q6pt5 Interface:lhnet1 IPs:[192.168.50.6] Mac:1e:1d:f3:e1:b5:5d Mtu:0 Default:false DNS:{Nameservers:[] Domain: Search:[] Options:[]} DeviceInfo:<nil> Gateway:[]}
2025-11-25T11:26:55Z [verbose] the NAD's config: {{"cniVersion":"0.3.1","type":"bridge","bridge":"mgmt-br","promiscMode":true,"vlan":50,"ipam":{"type":"whereabouts","range":"192.168.50.0/24","exclude":["192.168.50.1/32","192.168.50.2/32"]}}}
2025-11-25T11:26:55Z [debug] Used defaults from parsed flat file config @ /host/etc/cni/net.d/whereabouts.d/whereabouts.conf
2025-11-25T11:26:55Z [verbose] pool range [192.168.50.0/24]
2025-11-25T11:26:55Z [verbose] result of garbage collecting pods: <nil>
2025-11-25T11:40:35Z [verbose] deleted pod [cattle-system/cattle-cluster-agent-665568cdcf-6glxd]
2025-11-25T11:40:35Z [verbose] skipped net-attach-def for default network
2025-11-25T11:40:35Z [verbose] result of garbage collecting pods: <nil>
2025-11-25T11:40:35Z [verbose] deleted pod [harvester-system/harvester-node-manager-webhook-64dd9f8c6f-wfp4c]
2025-11-25T11:40:35Z [verbose] skipped net-attach-def for default network
2025-11-25T11:40:35Z [verbose] result of garbage collecting pods: <nil>
2025-11-25T11:40:43Z [verbose] deleted pod [harvester-system/virt-api-6897cf4644-49k9j]
2025-11-25T11:40:43Z [verbose] skipped net-attach-def for default network
2025-11-25T11:40:43Z [verbose] result of garbage collecting pods: <nil>
2025-11-25T11:42:22Z [verbose] deleted pod [cattle-monitoring-system/rancher-monitoring-grafana-7bcb7f655-8mg2t]
2025-11-25T11:42:22Z [verbose] result of garbage collecting pods: <nil>
2025-11-25T11:42:22Z [verbose] deleted pod [cattle-monitoring-system/alertmanager-rancher-monitoring-alertmanager-0]
2025-11-25T11:42:22Z [verbose] result of garbage collecting pods: <nil>
2025-11-25T11:42:22Z [verbose] deleted pod [cattle-monitoring-system/prometheus-rancher-monitoring-prometheus-0]
2025-11-25T11:42:22Z [verbose] result of garbage collecting pods: <nil>
2025-11-25T11:47:47Z [verbose] deleted pod [longhorn-system/instance-manager-ac9a7c78000218e7ce4a91ad97a99787]
2025-11-25T11:47:47Z [verbose] skipped net-attach-def for default network
2025-11-25T11:47:47Z [debug] pod's network status: {Name:harvester-system/storagenetwork-q6pt5 Interface:lhnet1 IPs:[192.168.50.3] Mac:9a:b6:d1:2e:c3:11 Mtu:0 Default:false DNS:{Nameservers:[] Domain: Search:[] Options:[]} DeviceInfo:<nil> Gateway:[]}
2025-11-25T11:47:47Z [verbose] result of garbage collecting pods: failed to get network-attachment-definition for iface harvester-system/storagenetwork-q6pt5: networkattachmentdefinition.k8s.cni.cncf.io "storagenetwork-q6pt5" not found
2025-11-25T11:47:47Z [verbose] re-queuing IP address reconciliation request for pod longhorn-system/instance-manager-ac9a7c78000218e7ce4a91ad97a99787; retry #: 0
2025-11-25T11:47:47Z [verbose] skipped net-attach-def for default network
2025-11-25T11:47:47Z [debug] pod's network status: {Name:harvester-system/storagenetwork-q6pt5 Interface:lhnet1 IPs:[192.168.50.3] Mac:9a:b6:d1:2e:c3:11 Mtu:0 Default:false DNS:{Nameservers:[] Domain: Search:[] Options:[]} DeviceInfo:<nil> Gateway:[]}
2025-11-25T11:47:47Z [verbose] result of garbage collecting pods: failed to get network-attachment-definition for iface harvester-system/storagenetwork-q6pt5: networkattachmentdefinition.k8s.cni.cncf.io "storagenetwork-q6pt5" not found
2025-11-25T11:47:47Z [verbose] re-queuing IP address reconciliation request for pod longhorn-system/instance-manager-ac9a7c78000218e7ce4a91ad97a99787; retry #: 1
2025-11-25T11:47:47Z [verbose] skipped net-attach-def for default network
2025-11-25T11:47:47Z [debug] pod's network status: {Name:harvester-system/storagenetwork-q6pt5 Interface:lhnet1 IPs:[192.168.50.3] Mac:9a:b6:d1:2e:c3:11 Mtu:0 Default:false DNS:{Nameservers:[] Domain: Search:[] Options:[]} DeviceInfo:<nil> Gateway:[]}
2025-11-25T11:47:47Z [verbose] result of garbage collecting pods: failed to get network-attachment-definition for iface harvester-system/storagenetwork-q6pt5: networkattachmentdefinition.k8s.cni.cncf.io "storagenetwork-q6pt5" not found
2025-11-25T11:47:47Z [verbose] re-queuing IP address reconciliation request for pod longhorn-system/instance-manager-ac9a7c78000218e7ce4a91ad97a99787; retry #: 2
2025-11-25T11:47:47Z [verbose] deleted pod [longhorn-system/bim-4ba2e66a1eda4c99916dc8ffe6a770c5a54d7602539e13e9847c0fd33d907432]
2025-11-25T11:47:47Z [verbose] skipped net-attach-def for default network
2025-11-25T11:47:47Z [debug] pod's network status: {Name:harvester-system/storagenetwork-q6pt5 Interface:lhnet1 IPs:[192.168.50.6] Mac:66:61:63:26:55:3b Mtu:0 Default:false DNS:{Nameservers:[] Domain: Search:[] Options:[]} DeviceInfo:<nil> Gateway:[]}
2025-11-25T11:47:47Z [verbose] result of garbage collecting pods: failed to get network-attachment-definition for iface harvester-system/storagenetwork-q6pt5: networkattachmentdefinition.k8s.cni.cncf.io "storagenetwork-q6pt5" not found
2025-11-25T11:47:47Z [verbose] re-queuing IP address reconciliation request for pod longhorn-system/bim-4ba2e66a1eda4c99916dc8ffe6a770c5a54d7602539e13e9847c0fd33d907432; retry #: 0
2025-11-25T11:47:47Z [verbose] skipped net-attach-def for default network
2025-11-25T11:47:47Z [debug] pod's network status: {Name:harvester-system/storagenetwork-q6pt5 Interface:lhnet1 IPs:[192.168.50.6] Mac:66:61:63:26:55:3b Mtu:0 Default:false DNS:{Nameservers:[] Domain: Search:[] Options:[]} DeviceInfo:<nil> Gateway:[]}
2025-11-25T11:47:47Z [verbose] result of garbage collecting pods: failed to get network-attachment-definition for iface harvester-system/storagenetwork-q6pt5: networkattachmentdefinition.k8s.cni.cncf.io "storagenetwork-q6pt5" not found
2025-11-25T11:47:47Z [verbose] re-queuing IP address reconciliation request for pod longhorn-system/bim-4ba2e66a1eda4c99916dc8ffe6a770c5a54d7602539e13e9847c0fd33d907432; retry #: 1
2025-11-25T11:47:47Z [verbose] skipped net-attach-def for default network
2025-11-25T11:47:47Z [debug] pod's network status: {Name:harvester-system/storagenetwork-q6pt5 Interface:lhnet1 IPs:[192.168.50.3] Mac:9a:b6:d1:2e:c3:11 Mtu:0 Default:false DNS:{Nameservers:[] Domain: Search:[] Options:[]} DeviceInfo:<nil> Gateway:[]}
2025-11-25T11:47:47Z [verbose] result of garbage collecting pods: failed to get network-attachment-definition for iface harvester-system/storagenetwork-q6pt5: networkattachmentdefinition.k8s.cni.cncf.io "storagenetwork-q6pt5" not found
2025-11-25T11:47:47Z [error] dropping pod [longhorn-system/instance-manager-ac9a7c78000218e7ce4a91ad97a99787] deletion out of the queue - could not reconcile IP: failed to get network-attachment-definition for iface harvester-system/storagenetwork-q6pt5: networkattachmentdefinition.k8s.cni.cncf.io "storagenetwork-q6pt5" not found
2025-11-25T11:47:47Z [verbose] Event(v1.ObjectReference{Kind:"Pod", Namespace:"longhorn-system", Name:"instance-manager-ac9a7c78000218e7ce4a91ad97a99787", UID:"9334564e-6719-48a5-a384-58f5016b3c98", APIVersion:"v1", ResourceVersion:"1682260", FieldPath:""}): type: 'Warning' reason: 'IPAddressGarbageCollectionFailed' failed to garbage collect addresses for pod longhorn-system/instance-manager-ac9a7c78000218e7ce4a91ad97a99787
2025-11-25T11:47:47Z [verbose] skipped net-attach-def for default network
2025-11-25T11:47:47Z [debug] pod's network status: {Name:harvester-system/storagenetwork-q6pt5 Interface:lhnet1 IPs:[192.168.50.6] Mac:66:61:63:26:55:3b Mtu:0 Default:false DNS:{Nameservers:[] Domain: Search:[] Options:[]} DeviceInfo:<nil> Gateway:[]}
2025-11-25T11:47:47Z [verbose] result of garbage collecting pods: failed to get network-attachment-definition for iface harvester-system/storagenetwork-q6pt5: networkattachmentdefinition.k8s.cni.cncf.io "storagenetwork-q6pt5" not found
2025-11-25T11:47:47Z [verbose] re-queuing IP address reconciliation request for pod longhorn-system/bim-4ba2e66a1eda4c99916dc8ffe6a770c5a54d7602539e13e9847c0fd33d907432; retry #: 2
2025-11-25T11:47:47Z [verbose] skipped net-attach-def for default network
2025-11-25T11:47:47Z [debug] pod's network status: {Name:harvester-system/storagenetwork-q6pt5 Interface:lhnet1 IPs:[192.168.50.6] Mac:66:61:63:26:55:3b Mtu:0 Default:false DNS:{Nameservers:[] Domain: Search:[] Options:[]} DeviceInfo:<nil> Gateway:[]}
2025-11-25T11:47:47Z [verbose] result of garbage collecting pods: failed to get network-attachment-definition for iface harvester-system/storagenetwork-q6pt5: networkattachmentdefinition.k8s.cni.cncf.io "storagenetwork-q6pt5" not found
2025-11-25T11:47:47Z [error] dropping pod [longhorn-system/bim-4ba2e66a1eda4c99916dc8ffe6a770c5a54d7602539e13e9847c0fd33d907432] deletion out of the queue - could not reconcile IP: failed to get network-attachment-definition for iface harvester-system/storagenetwork-q6pt5: networkattachmentdefinition.k8s.cni.cncf.io "storagenetwork-q6pt5" not found
2025-11-25T11:47:47Z [verbose] Event(v1.ObjectReference{Kind:"Pod", Namespace:"longhorn-system", Name:"bim-4ba2e66a1eda4c99916dc8ffe6a770c5a54d7602539e13e9847c0fd33d907432", UID:"0fccb3ba-dc92-42f1-a50d-cda5fe7e39b4", APIVersion:"v1", ResourceVersion:"1682269", FieldPath:""}): type: 'Warning' reason: 'IPAddressGarbageCollectionFailed' failed to garbage collect addresses for pod longhorn-system/bim-4ba2e66a1eda4c99916dc8ffe6a770c5a54d7602539e13e9847c0fd33d907432
2025-11-25T11:53:42Z [verbose] deleted pod [cattle-system/helm-operation-5sxcj]
2025-11-25T11:53:42Z [verbose] skipped net-attach-def for default network
2025-11-25T11:53:42Z [verbose] result of garbage collecting pods: <nil>
2025-11-25T11:53:50Z [verbose] deleted pod [cattle-system/helm-operation-mmw5l]
2025-11-25T11:53:50Z [verbose] skipped net-attach-def for default network
2025-11-25T11:53:50Z [verbose] result of garbage collecting pods: <nil>
```

##### Node Removal Summary

The whereabouts log is very similar to the one reported on JIRA ticket.

Whereabouts has silent leftover objects in this scenario.

#### Manual tests upon the leftover objects

##### Leftover `overlappingrangeipreservations` objects

1. A normal enabled SN will have following records

```yaml
$ kubectl get overlappingrangeipreservations.whereabouts.cni.cncf.io -A -oyaml
apiVersion: v1
items:
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: OverlappingRangeIPReservation
  metadata:
    creationTimestamp: "2025-11-25T08:46:31Z"
    generation: 1
    name: 192.168.50.3
    namespace: kube-system
    resourceVersion: "1382291"
    uid: f1919fff-6c40-41a2-9ec9-9ec10dbf6ecd
  spec:
    ifname: lhnet1
    podref: longhorn-system/bim-d3feb4b38b3614c5004a123b869edfa1bad8ecfef07ef323701c4433cea8c58b
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: OverlappingRangeIPReservation
  metadata:
    creationTimestamp: "2025-11-25T08:46:34Z"
    generation: 1
    name: 192.168.50.4
    namespace: kube-system
    resourceVersion: "1382346"
    uid: 8c9a28b2-d2ac-432d-a72a-d1f2b9ffc9c2
  spec:
    ifname: lhnet1
    podref: longhorn-system/instance-manager-6b77fd6b240ff0e6d0fe88c705b51f7f
kind: List
metadata:
  resourceVersion: ""
```

2. Disable the SN

```yaml
$ kubectl get overlappingrangeipreservations.whereabouts.cni.cncf.io -A
No resources found
```

3. Manually create a OverlappingRangeIPReservation object to occupy `192.168.50.3`.

```yaml
cat > olr.yaml << 'EOF'
apiVersion: whereabouts.cni.cncf.io/v1alpha1
kind: OverlappingRangeIPReservation
metadata:
  name: 192.168.50.3
  namespace: kube-system
spec:
  ifname: lhnet1
  podref: longhorn-system/bim-f3feb4b38b3614c5004a123b869edfa1bad8ecfef07ef323701c4433cea8c58c
EOF

kubectl create -f olr.yaml
```

```yaml
$ kubectl get overlappingrangeipreservations.whereabouts.cni.cncf.io -A
NAMESPACE     NAME           AGE
kube-system   192.168.50.3   4s

$ kubectl get overlappingrangeipreservations.whereabouts.cni.cncf.io -A -oyaml
apiVersion: v1
items:
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: OverlappingRangeIPReservation
  metadata:
    creationTimestamp: "2025-11-25T08:57:22Z"
    generation: 1
    name: 192.168.50.3
    namespace: kube-system
    resourceVersion: "1389890"
    uid: 798d849e-055e-4de5-a089-efe0ec123499
  spec:
    ifname: lhnet1
    podref: longhorn-system/bim-f3feb4b38b3614c5004a123b869edfa1bad8ecfef07ef323701c4433cea8c58c
kind: List
metadata:
  resourceVersion: ""

```

4. The leftover `overlappingrangeipreservations` object affects `IPPool` allocating

```yaml
$ kubectl get overlappingrangeipreservations.whereabouts.cni.cncf.io -A
NAMESPACE     NAME           AGE
kube-system   192.168.50.3   2m25s  // leftover
kube-system   192.168.50.4   37s
kube-system   192.168.50.5   2s

$ kubectl get overlappingrangeipreservations.whereabouts.cni.cncf.io -A -oyaml
apiVersion: v1
items:
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: OverlappingRangeIPReservation
  metadata:
    creationTimestamp: "2025-11-25T08:57:22Z"
    generation: 1
    name: 192.168.50.3
    namespace: kube-system
    resourceVersion: "1389890"
    uid: 798d849e-055e-4de5-a089-efe0ec123499
  spec:
    ifname: lhnet1
    podref: longhorn-system/bim-f3feb4b38b3614c5004a123b869edfa1bad8ecfef07ef323701c4433cea8c58c
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: OverlappingRangeIPReservation
  metadata:
    creationTimestamp: "2025-11-25T08:59:10Z"
    generation: 1
    name: 192.168.50.4
    namespace: kube-system
    resourceVersion: "1391151"
    uid: e00c6639-eb87-4452-92b4-5e39393a49d0
  spec:
    ifname: lhnet1
    podref: longhorn-system/bim-d3feb4b38b3614c5004a123b869edfa1bad8ecfef07ef323701c4433cea8c58b
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: OverlappingRangeIPReservation
  metadata:
    creationTimestamp: "2025-11-25T08:59:45Z"
    generation: 1
    name: 192.168.50.5
    namespace: kube-system
    resourceVersion: "1391582"
    uid: 88c8741c-cf8f-41fd-b756-09ddfc738543
  spec:
    ifname: lhnet1
    podref: longhorn-system/instance-manager-6b77fd6b240ff0e6d0fe88c705b51f7f
kind: List
metadata:
  resourceVersion: ""

$ kubectl get ippools.whereabouts.cni.cncf.io -A -oyaml
apiVersion: v1
items:
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: IPPool
  metadata:
    creationTimestamp: "2025-11-24T18:37:29Z"
    generation: 19
    name: 192.168.50.0-24
    namespace: kube-system
    resourceVersion: "1391577"
    uid: e7a5d3b2-337e-4c16-9f09-49b92e0cc3ee
  spec:
    allocations:
      "4":
        id: 7300435f92947fd66392fa3ba3f369045df44b0ebbe5e46acc762ce761c83f77
        ifname: lhnet1
        podref: longhorn-system/bim-d3feb4b38b3614c5004a123b869edfa1bad8ecfef07ef323701c4433cea8c58b
      "5":
        id: bd577644b320300a5aa13863b7d971ec7957ce8f647821058cadd4844a438458
        ifname: lhnet1
        podref: longhorn-system/instance-manager-6b77fd6b240ff0e6d0fe88c705b51f7f
    range: 192.168.50.0/24
kind: List
metadata:
  resourceVersion: ""

```

5. Disable SN

The leftover `overlappingrangeipreservations` is still there, and could affect next time allocating.

```yaml

$ kubectl get ippools.whereabouts.cni.cncf.io -A -oyaml
apiVersion: v1
items:
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: IPPool
  metadata:
    creationTimestamp: "2025-11-24T18:37:29Z"
    generation: 21
    name: 192.168.50.0-24
    namespace: kube-system
    resourceVersion: "1394165"
    uid: e7a5d3b2-337e-4c16-9f09-49b92e0cc3ee
  spec:
    allocations: {}
    range: 192.168.50.0/24
kind: List
metadata:
  resourceVersion: ""


$ kubectl get overlappingrangeipreservations.whereabouts.cni.cncf.io -A -oyaml
apiVersion: v1
items:
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: OverlappingRangeIPReservation
  metadata:
    creationTimestamp: "2025-11-25T08:57:22Z"
    generation: 1
    name: 192.168.50.3
    namespace: kube-system
    resourceVersion: "1389890"
    uid: 798d849e-055e-4de5-a089-efe0ec123499
  spec:
    ifname: lhnet1
    podref: longhorn-system/bim-f3feb4b38b3614c5004a123b869edfa1bad8ecfef07ef323701c4433cea8c58c
kind: List
metadata:
  resourceVersion: ""
harv41:/home/rancher # 
```

6. Delete `IPPool` object explicitly

```yaml
$ kubectl delete ippools.whereabouts.cni.cncf.io 192.168.50.0-24 -n kube-system
ippool.whereabouts.cni.cncf.io "192.168.50.0-24" deleted from kube-system namespace

$ kubectl get ippools.whereabouts.cni.cncf.io -A
No resources found
```

7. Enable `SN` again

```yaml
$ kubectl get ippools.whereabouts.cni.cncf.io -A
NAMESPACE     NAME              AGE
kube-system   192.168.50.0-24   6s

$ kubectl get overlappingrangeipreservations.whereabouts.cni.cncf.io -A
NAMESPACE     NAME           AGE
kube-system   192.168.50.3   9m31s
kube-system   192.168.50.4   13s
kube-system   192.168.50.5   8s

$ kubectl get ippools.whereabouts.cni.cncf.io -A -oyaml
apiVersion: v1
items:
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: IPPool
  metadata:
    creationTimestamp: "2025-11-25T09:06:40Z"
    generation: 3
    name: 192.168.50.0-24
    namespace: kube-system
    resourceVersion: "1396422"
    uid: 94688453-1729-4a5c-ba60-f8615ac04bd0
  spec:
    allocations:
      "4":
        id: 87d76a4617d61a51972a26882d856cb25021166947ba90d5ddee180061780c27
        ifname: lhnet1
        podref: longhorn-system/bim-d3feb4b38b3614c5004a123b869edfa1bad8ecfef07ef323701c4433cea8c58b
      "5":
        id: 60022b47a60e54ae43c98716bf5cf210635979d34aef44fb26251c854a90fd7f
        ifname: lhnet1
        podref: longhorn-system/instance-manager-6b77fd6b240ff0e6d0fe88c705b51f7f
    range: 192.168.50.0/24
kind: List
metadata:
  resourceVersion: ""
```

Summary: The leftover `overlappingrangeipreservations` matters.

##### Leftover `IPPools` objects

1. Remove any potential IPPool objects

```yaml
kubectl get ippools.whereabouts.cni.cncf.io -A
No resources found
```

2. Manually construct and build it

```bash
cat > ipp.yaml << 'EOF'
apiVersion: whereabouts.cni.cncf.io/v1alpha1
kind: IPPool
metadata:
  name: 192.168.50.0-24
  namespace: kube-system
spec:
  allocations:
    "4":
      id: aad76a4617d61a51972a26882d856cb25021166947ba90d5ddee180061780c27
      ifname: lhnet1
      podref: longhorn-system/bim-f3feb4b38b3614c5004a123b869edfa1bad8ecfef07ef323701c4433cea8c58c
    "5":
      id: aa022b47a60e54ae43c98716bf5cf210635979d34aef44fb26251c854a90fd7f
      ifname: lhnet1
      podref: longhorn-system/instance-manager-fb77fd6b240ff0e6d0fe88c705b51f7c
  range: 192.168.50.0/24
EOF
```

```yaml
$ kubectl get ippools.whereabouts.cni.cncf.io -A
NAMESPACE     NAME              AGE
kube-system   192.168.50.0-24   10s

$ kubectl get ippools.whereabouts.cni.cncf.io -A -oyaml
apiVersion: v1
items:
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: IPPool
  metadata:
    creationTimestamp: "2025-11-25T09:13:40Z"
    generation: 1
    name: 192.168.50.0-24
    namespace: kube-system
    resourceVersion: "1401151"
    uid: 661ee191-146c-4a9b-b100-eb676ac4fe0a
  spec:
    allocations:
      "4":
        id: aad76a4617d61a51972a26882d856cb25021166947ba90d5ddee180061780c27
        ifname: lhnet1
        podref: longhorn-system/bim-f3feb4b38b3614c5004a123b869edfa1bad8ecfef07ef323701c4433cea8c58c
      "5":
        id: aa022b47a60e54ae43c98716bf5cf210635979d34aef44fb26251c854a90fd7f
        ifname: lhnet1
        podref: longhorn-system/instance-manager-fb77fd6b240ff0e6d0fe88c705b51f7c
    range: 192.168.50.0/24
kind: List
metadata:
  resourceVersion: ""


$ kubectl kubectl get overlappingrangeipreservations.whereabouts.cni.cncf.io -A
No resources found
```

3. Enable SN

```yaml
$ kubectl get overlappingrangeipreservations.whereabouts.cni.cncf.io -A
NAMESPACE     NAME           AGE
kube-system   192.168.50.3   14s
kube-system   192.168.50.6   12s   #  note the 192.168.50.4, 192.168.50.5 are jumped, as they are occupied by existing records

$ kubectl kubectl get ippools.whereabouts.cni.cncf.io -A -oyaml
apiVersion: v1
items:
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: IPPool
  metadata:
    creationTimestamp: "2025-11-25T09:13:40Z"
    generation: 3
    name: 192.168.50.0-24
    namespace: kube-system
    resourceVersion: "1403754"
    uid: 661ee191-146c-4a9b-b100-eb676ac4fe0a
  spec:
    allocations:
      "3":
        id: fa1343b2c4de268060cf0baca82ba573bda7af4c3a50995866b79747a7a1a742
        ifname: lhnet1
        podref: longhorn-system/bim-d3feb4b38b3614c5004a123b869edfa1bad8ecfef07ef323701c4433cea8c58b
      "4":
        id: aad76a4617d61a51972a26882d856cb25021166947ba90d5ddee180061780c27
        ifname: lhnet1
        podref: longhorn-system/bim-f3feb4b38b3614c5004a123b869edfa1bad8ecfef07ef323701c4433cea8c58c
      "5":
        id: aa022b47a60e54ae43c98716bf5cf210635979d34aef44fb26251c854a90fd7f
        ifname: lhnet1
        podref: longhorn-system/instance-manager-fb77fd6b240ff0e6d0fe88c705b51f7c
      "6":
        id: 9e9735e34ff542f4fc63096e205aabe4fd89279e60c35c875e37868b32b99e10
        ifname: lhnet1
        podref: longhorn-system/instance-manager-6b77fd6b240ff0e6d0fe88c705b51f7f
    range: 192.168.50.0/24
kind: List
metadata:
  resourceVersion: ""
```


4. Disable SN

```yaml
$ kubectl get ippools.whereabouts.cni.cncf.io -A -oyaml
apiVersion: v1
items:
- apiVersion: whereabouts.cni.cncf.io/v1alpha1
  kind: IPPool
  metadata:
    creationTimestamp: "2025-11-25T09:13:40Z"
    generation: 5
    name: 192.168.50.0-24
    namespace: kube-system
    resourceVersion: "1404375"
    uid: 661ee191-146c-4a9b-b100-eb676ac4fe0a
  spec:
    allocations:
      "4":
        id: aad76a4617d61a51972a26882d856cb25021166947ba90d5ddee180061780c27
        ifname: lhnet1
        podref: longhorn-system/bim-f3feb4b38b3614c5004a123b869edfa1bad8ecfef07ef323701c4433cea8c58c
      "5":
        id: aa022b47a60e54ae43c98716bf5cf210635979d34aef44fb26251c854a90fd7f
        ifname: lhnet1
        podref: longhorn-system/instance-manager-fb77fd6b240ff0e6d0fe88c705b51f7c
    range: 192.168.50.0/24
kind: List
metadata:
  resourceVersion: ""

$ kubectl get overlappingrangeipreservations.whereabouts.cni.cncf.io -A
No resources found

```

Summary: The leftover `ippools.spec.allocations` matters.

### Self Healing


## Additional

Related bugs and PRs

[PR: Ensure storage network configured from clean state](https://github.com/harvester/harvester/pull/9453)

[enh Check storage-network is fully stopped on LH pods before re-enable storage-network](https://github.com/harvester/harvester/issues/9141)

[bug Storage-network NAD can be deleted from kubectl directly](https://github.com/harvester/harvester/issues/9623)

[enh Ensure or Support to configure the exclusive VID for storage-network](https://github.com/harvester/harvester/issues/9622)

[bug NAD leaking is seen when you configure the storage-network quickly](https://github.com/harvester/harvester/issues/9621)