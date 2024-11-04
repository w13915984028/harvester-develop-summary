# Config indepedent nic with L2 network on Harvester

## Scenario

When running a third party storage network solution or some other workloads on Harvester, they require the Harvester node to have an independent nic with L2 VLAN config.

The related config can be applied manually, but it may be lost after the node/cluster is rebooted.

This guide shows how to persistant the config.

For example, add a nic named `ens8` to `vlan 100`.

## Solution

### Create a Harvester CloudInit Object

```
cat > inv100.yaml << EOF
apiVersion: node.harvesterhci.io/v1beta1
kind: CloudInit
metadata:
  name: independent-nic-vlan-100
spec:
  matchSelector: {}
  filename: 95_independent_nic_vlan_100.yaml
  contents: |
    stages:
      initramfs:
        - name: tpsn
          files:
            - path: /etc/sysconfig/network/ifcfg-vlan100
              permissions: 384
              owner: 0
              group: 0
              content: |
                STARTMODE='hotplug'
                BOOTPROTO='none'
                ETHERDEVICE="ens8"
                VLAN_ID="100"
                VLAN='yes'
              encoding: ""
              ownerstring: ""
EOF

$ kubectl create -f inv100.yaml
```

:::note

Adapt the `VLAN_ID`, `ETHERDEVICE` to your cluster.

If different nodes have different NIC names, then use [matchSelector](https://docs.harvesterhci.io/v1.3/host/#cloudinit-resource-spec) and create several objects accordingly.

:::

### Check the Harvester CloudInit Object

Check this object, there should be no errors.

```
$ kubectl get cloudinit -A
NAME                       AGE
independent-nic-vlan-100   9s

$ kubectl get cloudinit -A -oyaml
apiVersion: v1
items:
- apiVersion: node.harvesterhci.io/v1beta1
  kind: CloudInit
  metadata:
    annotations:
      node.harvesterhci.io/cloudinit-hash: d98138c58975f4aa5f7c386a14434c0b476567faf76a0946ca06e571dea2409a
    creationTimestamp: "2024-11-04T10:49:28Z"
    finalizers:
    - wrangler.cattle.io/harvester-node-cloud-init-controller
    generation: 1
    name: independent-nic-vlan-100
    resourceVersion: "1724514"
    uid: fed688b8-3cfe-414c-8840-a26a2cc6e732
  spec:
    contents: |
      stages:
        initramfs:
          - name: tpsn
            files:
              - path: /etc/sysconfig/network/ifcfg-vlan100
                permissions: 384
                owner: 0
                group: 0
                content: |
                  STARTMODE='hotplug'
                  BOOTPROTO='none'
                  ETHERDEVICE="ens8"
                  VLAN_ID="100"
                  VLAN='yes'
                encoding: ""
                ownerstring: ""
    filename: 95_independent_nic_vlan_100.yaml
    matchSelector: {}
  status:
    rollouts:
      harv41:
        conditions:
        - lastTransitionTime: "2024-11-04T10:49:28Z"
          message: ""
          reason: CloudInitApplicable
          status: "True"
          type: Applicable
        - lastTransitionTime: "2024-11-04T10:49:28Z"
          message: Local file checksum is the same as the CloudInit checksum
          reason: CloudInitChecksumMatch
          status: "False"
          type: OutOfSync
        - lastTransitionTime: "2024-11-04T10:49:28Z"
          message: 95_independent_nic_vlan_100.yaml is present under /oem
          reason: CloudInitPresentOnDisk
          status: "True"
          type: Present
kind: List
metadata:
  resourceVersion: ""
```

### Check the Backend Files

After a while, on the target nodes, a file `/oem/95_independent_nic_vlan_100.yaml ` will be created.

```
$  ls /oem -alth
...
-rw-------   1 root root  380 Nov  4 10:49 95_independent_nic_vlan_100.yaml
...

$ cat /oem/95_independent_nic_vlan_100.yaml 
stages:
  initramfs:
    - name: tpsn
      files:
        - path: /etc/sysconfig/network/ifcfg-vlan100
          permissions: 384
          owner: 0
          group: 0
          content: |
            STARTMODE='hotplug'
            BOOTPROTO='none'
            ETHERDEVICE="ens8"
            VLAN_ID="100"
            VLAN='yes'
          encoding: ""
          ownerstring: ""
```

The related Harvester network config file.

```
$  ls /etc/sysconfig/network/ -alth
total 88K
drwxr-xr-x. 1 root root  140 Nov  4 10:52 .
-rw-------  1 root root   81 Nov  4 10:52 ifcfg-vlan100
...

$ cat /etc/sysconfig/network/ifcfg-vlan100 
STARTMODE='hotplug'
BOOTPROTO='none'
ETHERDEVICE="ens8"
VLAN_ID="100"
VLAN='yes'
```

:::note

The file is not loaded automatically before the node is restarted or service is reloaded.

:::

### Manually apply the config

Use `wicked ifreload vlan100` command to apply the config manually before the node is rebooted.

A successful operation will have below output.

```
$ wicked ifreload vlan100
wicked: ifcfg-vlan100: generated missing vlan ETHERDEVICE 'ens8' config for 'vlan100'
ens8            up
vlan100         up
```

### Reboot the node or cluster

After a node is rebooted, the above config is reloaded automatically.

It sures the nic is up and a related vlan interface is created.

```
$ ip link | grep ens8
3: ens8: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP mode DEFAULT group default qlen 1000
6: vlan100@ens8: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP mode DEFAULT group default qlen 1000
```
