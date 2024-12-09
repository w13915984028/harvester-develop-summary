# Test VM Migration with ResourceQuota

## Background

When VM is under migration, Harvester will scale up & down the related resourcequota to avoid migration is blocked due to quota limitation.

https://github.com/harvester/harvester/blob/992cd4c9ea20b95ed51d6d3a7654cec47255dc83/pkg/controller/master/migration/vmim_controller.go#L54

issue: https://github.com/harvester/harvester/issues/7161

Following log is observed from one environment:

```
./harvester-system.virt-controller-6bc767746-qr5rm.virt-controller.20241210.log.gz:{"component":"virt-controller","level":"info","msg":"reenqueuing Migration ns1/kubevirt-evacuation-dfp6l","pos":"migration.go:233","reason":"failed to create vmi migration target pod: pods \"virt-launcher-ns1-4ae21d5d-n2mtw-qgtdp\" is forbidden:

 exceeded quota: default-4tmmk,

requested: limits.cpu=4,015m,  limits.memory= 8,940,039,936,
used:      limits.cpu=28,505m, limits.memory=62,900,279,552,
limited:   limits.cpu=26,      limits.memory=54,242,646Ki",

"timestamp":"2024-12-10T14:36:47.533782Z"}
```

The error is printed from:

https://github.com/kubernetes/kubernetes/blob/b1f2af04328936c2fa79db4af14f5c6ad9160748/staging/src/k8s.io/apiserver/pkg/admission/plugin/resourcequota/controller.go#L571

```
...
		if allowed, exceeded := quota.LessThanOrEqual(maskedNewUsage, resourceQuota.Status.Hard); !allowed {
			failedRequestedUsage := quota.Mask(requestedUsage, exceeded)
			failedUsed := quota.Mask(resourceQuota.Status.Used, exceeded)
			failedHard := quota.Mask(resourceQuota.Status.Hard, exceeded)
			return nil, admission.NewForbidden(a,
				fmt.Errorf("exceeded quota: %s, requested: %s, used: %s, limited: %s",
					resourceQuota.Name,
					prettyPrint(failedRequestedUsage),
					prettyPrint(failedUsed),
					prettyPrint(failedHard)))
		}
...
```

## Local Test Environment

Harvester v1.2.1

```
runtimeversion: v1.25.9+rke2r1
rancherversion: v2.7.5
harvesterchartversion: 1.2.1
monitoringchartversion: 102.0.0+up40.1.2
```

Nodes:

```
$kubectl get nodes -A
NAME     STATUS   ROLES                       AGE   VERSION
harv2    Ready    <none>                      24m   v1.25.9+rke2r1
harv41   Ready    control-plane,etcd,master   40m   v1.25.9+rke2r1
```

## Manual VM Migration Test

### Create a namespace test-migration and related ResourceQuota

```
 cat > rq-test.yaml << 'EOF'
apiVersion: v1
kind: ResourceQuota
metadata:
  generateName: default-
  labels:
    resourcequota.management.cattle.io/default-resource-quota: "true"
  namespace: test-migration
spec:
  hard:
    configmaps: "15"
    limits.cpu: "30"
    limits.memory: 60Gi
    persistentvolumeclaims: "75"
    pods: "50"
    replicationcontrollers: "50"
    requests.cpu: "30"
    requests.memory: 60Gi
    requests.storage: 320Gi
    secrets: "65"
    services: "15"
    services.loadbalancers: "1"
    services.nodeports: "15"
EOF

kubectl create -f rq-test.yaml

```

After resourcequota is initially created:

```
$kubectl get resourcequota -A
NAMESPACE        NAME            AGE   REQUEST                                                                                                                                                                                                                                                  LIMIT
test-migration   default-gn4d2   15s   configmaps: 1/15, persistentvolumeclaims: 0/75, pods: 0/50, replicationcontrollers: 0/50, requests.cpu: 0/30, requests.memory: 0/60Gi, requests.storage: 0/320Gi, secrets: 0/65, services: 0/15, services.loadbalancers: 0/1, services.nodeports: 0/15   limits.cpu: 0/30, limits.memory: 0/60Gi

$kubectl get resourcequota -A -oyaml
apiVersion: v1
items:
- apiVersion: v1
  kind: ResourceQuota
  metadata:
    creationTimestamp: "2024-12-13T12:35:59Z"
    generateName: default-
    labels:
      resourcequota.management.cattle.io/default-resource-quota: "true"
    name: default-gn4d2
    namespace: test-migration
    resourceVersion: "30985"
    uid: f560ca71-7ecb-4c0e-8614-b63111ef92c4
  spec:
    hard:
      configmaps: "15"
      limits.cpu: "30"
      limits.memory: 60Gi
      persistentvolumeclaims: "75"
      pods: "50"
      replicationcontrollers: "50"
      requests.cpu: "30"
      requests.memory: 60Gi
      requests.storage: 320Gi
      secrets: "65"
      services: "15"
      services.loadbalancers: "1"
      services.nodeports: "15"
  status:
    hard:
      configmaps: "15"
      limits.cpu: "30"
      limits.memory: 60Gi
      persistentvolumeclaims: "75"
      pods: "50"
      replicationcontrollers: "50"
      requests.cpu: "30"
      requests.memory: 60Gi
      requests.storage: 320Gi
      secrets: "65"
      services: "15"
      services.loadbalancers: "1"
      services.nodeports: "15"
    used:
      configmaps: "1"
      limits.cpu: "0"
      limits.memory: "0"
      persistentvolumeclaims: "0"
      pods: "0"
      replicationcontrollers: "0"
      requests.cpu: "0"
      requests.memory: "0"
      requests.storage: "0"
      secrets: "0"
      services: "0"
      services.loadbalancers: "0"
      services.nodeports: "0"
kind: List
metadata:
  resourceVersion: ""


```

### Create a VM with 2 core, 4GI memory

note: yamls are taken after VM is migrated

```
$kubectl get vm -A -oyaml
apiVersion: v1
items:
- apiVersion: kubevirt.io/v1
  kind: VirtualMachine
  metadata:
    annotations:
      harvesterhci.io/timestamp: "2024-12-13T12:39:03Z"
      harvesterhci.io/vmRunStrategy: RerunOnFailure
      harvesterhci.io/volumeClaimTemplates: '[{"metadata":{"name":"vm1-disk-0-egtdn","annotations":{"harvesterhci.io/imageId":"test-migration/image-hgdh2"}},"spec":{"accessModes":["ReadWriteMany"],"resources":{"requests":{"storage":"10Gi"}},"volumeMode":"Block","storageClassName":"longhorn-image-hgdh2"}}]'
      kubevirt.io/latest-observed-api-version: v1
      kubevirt.io/storage-observed-api-version: v1alpha3
      network.harvesterhci.io/ips: '[]'
    creationTimestamp: "2024-12-13T12:37:22Z"
    finalizers:
    - harvesterhci.io/VMController.UnsetOwnerOfPVCs
    generation: 2
    labels:
      harvesterhci.io/creator: harvester
    name: vm1
    namespace: test-migration
    resourceVersion: "33858"
    uid: 0fea47ca-d9c6-4b0f-94cc-8bc03350d09a
  spec:
    runStrategy: RerunOnFailure
    template:
      metadata:
        annotations:
          harvesterhci.io/sshNames: '[]'
        creationTimestamp: null
        labels:
          harvesterhci.io/vmName: vm1
      spec:
        affinity: {}
        domain:
          cpu:
            cores: 2
            sockets: 1
            threads: 1
          devices:
            disks:
            - bootOrder: 1
              disk:
                bus: virtio
              name: disk-0
            - disk:
                bus: virtio
              name: cloudinitdisk
            inputs:
            - bus: usb
              name: tablet
              type: tablet
            interfaces:
            - macAddress: 52:54:00:b6:df:da
              masquerade: {}
              model: virtio
              name: default
          features:
            acpi:
              enabled: true
          machine:
            type: q35
          memory:
            guest: 3996Mi
          resources:
            limits:
              cpu: "2"
              memory: 4Gi
            requests:
              cpu: 125m
              memory: 2730Mi
        evictionStrategy: LiveMigrate
        hostname: vm1
        networks:
        - name: default
          pod: {}
        terminationGracePeriodSeconds: 120
        volumes:
        - name: disk-0
          persistentVolumeClaim:
            claimName: vm1-disk-0-egtdn
        - cloudInitNoCloud:
            networkDataSecretRef:
              name: vm1-mmz3d
            secretRef:
              name: vm1-mmz3d
          name: cloudinitdisk
  status:
    conditions:
    - lastProbeTime: null
      lastTransitionTime: "2024-12-13T12:39:03Z"
      status: "True"
      type: Ready
    - lastProbeTime: null
      lastTransitionTime: null
      status: "True"
      type: LiveMigratable
    created: true
    printableStatus: Running
    ready: true
    volumeSnapshotStatuses:
    - enabled: false
      name: disk-0
      reason: 2 matching VolumeSnapshotClasses for longhorn-image-hgdh2
    - enabled: false
      name: cloudinitdisk
      reason: Snapshot is not supported for this volumeSource type [cloudinitdisk]
kind: List
metadata:
  resourceVersion: ""

$kubectl get vmi -A -oyaml
apiVersion: v1
items:
- apiVersion: kubevirt.io/v1
  kind: VirtualMachineInstance
  metadata:
    annotations:
      harvesterhci.io/sshNames: '[]'
      kubevirt.io/latest-observed-api-version: v1
      kubevirt.io/storage-observed-api-version: v1alpha3
    creationTimestamp: "2024-12-13T12:37:23Z"
    finalizers:
    - kubevirt.io/virtualMachineControllerFinalize
    - foregroundDeleteVirtualMachine
    - wrangler.cattle.io/VMIController.UnsetOwnerOfPVCs
    - wrangler.cattle.io/harvester-lb-vmi-controller
    generation: 22
    labels:
      harvesterhci.io/vmName: vm1
      kubevirt.io/migrationTargetNodeName: harv41
      kubevirt.io/nodeName: harv41
    name: vm1
    namespace: test-migration
    ownerReferences:
    - apiVersion: kubevirt.io/v1
      blockOwnerDeletion: true
      controller: true
      kind: VirtualMachine
      name: vm1
      uid: 0fea47ca-d9c6-4b0f-94cc-8bc03350d09a
    resourceVersion: "33857"
    uid: d62c2a6f-d8f3-4cc5-a22e-4cffb7d5eac1
  spec:
    affinity: {}
    domain:
      cpu:
        cores: 2
        model: host-model
        sockets: 1
        threads: 1
      devices:
        disks:
        - bootOrder: 1
          disk:
            bus: virtio
          name: disk-0
        - disk:
            bus: virtio
          name: cloudinitdisk
        inputs:
        - bus: usb
          name: tablet
          type: tablet
        interfaces:
        - masquerade: {}
          model: virtio
          name: default
      features:
        acpi:
          enabled: true
      firmware:
        uuid: 133bf63e-9459-5126-9b21-b56e9b3d17b3
      machine:
        type: q35
      memory:
        guest: 3996Mi
      resources:
        limits:
          cpu: "2"
          memory: 4Gi
        requests:
          cpu: 125m
          memory: 2730Mi
    evictionStrategy: LiveMigrate
    hostname: vm1
    networks:
    - name: default
      pod: {}
    terminationGracePeriodSeconds: 120
    volumes:
    - name: disk-0
      persistentVolumeClaim:
        claimName: vm1-disk-0-egtdn
    - cloudInitNoCloud:
        networkDataSecretRef:
          name: vm1-mmz3d
        secretRef:
          name: vm1-mmz3d
      name: cloudinitdisk
  status:
    activePods:
      63b16150-8df1-4ea2-aada-a096c1f2615c: harv2
      a2205f79-1b2d-466d-87b3-59724d90e622: harv41
    conditions:
    - lastProbeTime: null
      lastTransitionTime: "2024-12-13T12:39:03Z"
      status: "True"
      type: Ready
    - lastProbeTime: null
      lastTransitionTime: null
      status: "True"
      type: LiveMigratable
    guestOSInfo: {}
    interfaces:
    - infoSource: domain
      ipAddress: 10.52.0.77
      ipAddresses:
      - 10.52.0.77
      mac: 52:54:00:b6:df:da
      name: default
    launcherContainerImageVersion: registry.suse.com/suse/sles/15.4/virt-launcher:0.54.0-150400.3.19.1
    migrationMethod: BlockMigration
    migrationState:
      completed: true
      endTimestamp: "2024-12-13T12:39:03Z"
      migrationConfiguration:
        allowAutoConverge: false
        allowPostCopy: false
        bandwidthPerMigration: "0"
        completionTimeoutPerGiB: 800
        nodeDrainTaintKey: kubevirt.io/drain
        parallelMigrationsPerCluster: 5
        parallelOutboundMigrationsPerNode: 2
        progressTimeout: 150
        unsafeMigrationOverride: false
      migrationUid: 9f18f964-5233-4f06-a7db-8f0fad5d77be
      mode: PreCopy
      sourceNode: harv2
      startTimestamp: "2024-12-13T12:38:58Z"
      targetDirectMigrationNodePorts:
        "33403": 49153
        "33561": 49152
        "37435": 0
      targetNode: harv41
      targetNodeAddress: 10.52.0.69
      targetNodeDomainDetected: true
      targetPod: virt-launcher-vm1-rmbgb
    migrationTransport: Unix
    nodeName: harv41
    phase: Running
    phaseTransitionTimestamps:
    - phase: Pending
      phaseTransitionTimestamp: "2024-12-13T12:37:23Z"
    - phase: Scheduling
      phaseTransitionTimestamp: "2024-12-13T12:37:23Z"
    - phase: Scheduled
      phaseTransitionTimestamp: "2024-12-13T12:37:43Z"
    - phase: Running
      phaseTransitionTimestamp: "2024-12-13T12:37:45Z"
    qosClass: Burstable
    runtimeUser: 0
    virtualMachineRevisionName: revision-start-vm-0fea47ca-d9c6-4b0f-94cc-8bc03350d09a-1
    volumeStatus:
    - name: cloudinitdisk
      size: 1048576
      target: vdb
    - name: disk-0
      persistentVolumeClaimInfo:
        accessModes:
        - ReadWriteMany
        capacity:
          storage: 10Gi
        filesystemOverhead: "0.055"
        requests:
          storage: 10Gi
        volumeMode: Block
      target: vda
kind: List
metadata:
  resourceVersion: ""

```

### Migrate VM and watch resourcequota

The RQ was scaled up and down during the migration.

### Migration is on-going

```
$kubectl get resourcequota -A -oyaml
apiVersion: v1
items:
- apiVersion: v1
  kind: ResourceQuota
  metadata:
    annotations:
      harvesterhci.io/migrating-vm1: '{"limits.cpu":"2","limits.memory":"4531245057"}'  // on-going, VM's limits was added to the original limits
    creationTimestamp: "2024-12-13T12:35:59Z"
    generateName: default-
    labels:
      resourcequota.management.cattle.io/default-resource-quota: "true"
    name: default-gn4d2
    namespace: test-migration
    resourceVersion: "33524"
    uid: f560ca71-7ecb-4c0e-8614-b63111ef92c4
  spec:
    hard:
      configmaps: "15"
      limits.cpu: "32"             // cores added
      limits.memory: "68955754497" // memory added
      persistentvolumeclaims: "75"
      pods: "50"
      replicationcontrollers: "50"
      requests.cpu: "30"
      requests.memory: 60Gi
      requests.storage: 320Gi
      secrets: "65"
      services: "15"
      services.loadbalancers: "1"
      services.nodeports: "15"
  status:
    hard:
      configmaps: "15"
      limits.cpu: "32"
      limits.memory: "68955754497"
      persistentvolumeclaims: "75"
      pods: "50"
      replicationcontrollers: "50"
      requests.cpu: "30"
      requests.memory: 60Gi
      requests.storage: 320Gi
      secrets: "65"
      services: "15"
      services.loadbalancers: "1"
      services.nodeports: "15"
    used:
      configmaps: "1"
      limits.cpu: "4"
      limits.memory: "9062490114"
      persistentvolumeclaims: "1"
      pods: "2"
      replicationcontrollers: "0"
      requests.cpu: 250m
      requests.memory: "6197780482"
      requests.storage: 10Gi
      secrets: "1"
      services: "0"
      services.loadbalancers: "0"
      services.nodeports: "0"
kind: List
metadata:
  resourceVersion: ""

```

### Migration is done spec is scaledown but `status.used` is not down yet

vm-migration is finished, but the temporary/old VMs/PODs are not cleaned yet.

```
$kubectl get resourcequota -A -oyaml
apiVersion: v1
items:
- apiVersion: v1
  kind: ResourceQuota
  metadata:
    creationTimestamp: "2024-12-13T12:35:59Z"
    generateName: default-
    labels:
      resourcequota.management.cattle.io/default-resource-quota: "true"
    name: default-gn4d2
    namespace: test-migration
    resourceVersion: "33850"
    uid: f560ca71-7ecb-4c0e-8614-b63111ef92c4
  spec:
    hard:
      configmaps: "15"
      limits.cpu: "30"
      limits.memory: "64424509440"
      persistentvolumeclaims: "75"
      pods: "50"
      replicationcontrollers: "50"
      requests.cpu: "30"
      requests.memory: 60Gi
      requests.storage: 320Gi
      secrets: "65"
      services: "15"
      services.loadbalancers: "1"
      services.nodeports: "15"
  status:
    hard:
      configmaps: "15"
      limits.cpu: "30"
      limits.memory: "64424509440"
      persistentvolumeclaims: "75"
      pods: "50"
      replicationcontrollers: "50"
      requests.cpu: "30"
      requests.memory: 60Gi
      requests.storage: 320Gi
      secrets: "65"
      services: "15"
      services.loadbalancers: "1"
      services.nodeports: "15"
    used:
      configmaps: "1"
      limits.cpu: "4"                    // two VMs, (2C 4Gi)*2
      limits.memory: "9062490114"
      persistentvolumeclaims: "1"
      pods: "2"
      replicationcontrollers: "0"
      requests.cpu: 250m
      requests.memory: "6197780482"
      requests.storage: 10Gi
      secrets: "1"
      services: "0"
      services.loadbalancers: "0"
      services.nodeports: "0"
kind: List
metadata:
  resourceVersion: ""
```

### Finally `status.used` is down

```
$kubectl get resourcequota -A -oyaml
apiVersion: v1
items:
- apiVersion: v1
  kind: ResourceQuota
  metadata:
    creationTimestamp: "2024-12-13T12:35:59Z"
    generateName: default-
    labels:
      resourcequota.management.cattle.io/default-resource-quota: "true"
    name: default-gn4d2
    namespace: test-migration
    resourceVersion: "33946"
    uid: f560ca71-7ecb-4c0e-8614-b63111ef92c4
  spec:
    hard:
      configmaps: "15"
      limits.cpu: "30"
      limits.memory: "64424509440"
      persistentvolumeclaims: "75"
      pods: "50"
      replicationcontrollers: "50"
      requests.cpu: "30"
      requests.memory: 60Gi
      requests.storage: 320Gi
      secrets: "65"
      services: "15"
      services.loadbalancers: "1"
      services.nodeports: "15"
  status:
    hard:
      configmaps: "15"
      limits.cpu: "30"
      limits.memory: "64424509440"
      persistentvolumeclaims: "75"
      pods: "50"
      replicationcontrollers: "50"
      requests.cpu: "30"
      requests.memory: 60Gi
      requests.storage: 320Gi
      secrets: "65"
      services: "15"
      services.loadbalancers: "1"
      services.nodeports: "15"
    used:
      configmaps: "1"
      limits.cpu: "2"                    // 1 VM, (2C 4Gi)
      limits.memory: "4531245057"
      persistentvolumeclaims: "1"
      pods: "1"
      replicationcontrollers: "0"
      requests.cpu: 125m
      requests.memory: "3098890241"
      requests.storage: 10Gi
      secrets: "1"
      services: "0"
      services.loadbalancers: "0"
      services.nodeports: "0"
kind: List
metadata:
  resourceVersion: ""

```

### vmim is successfuly

```
$kubectl get vmim -A -oyaml
apiVersion: v1
items:
- apiVersion: kubevirt.io/v1
  kind: VirtualMachineInstanceMigration
  metadata:
    annotations:
      kubevirt.io/latest-observed-api-version: v1
      kubevirt.io/storage-observed-api-version: v1alpha3
    creationTimestamp: "2024-12-13T12:38:45Z"
    generateName: vm1-
    generation: 1
    labels:
      kubevirt.io/vmi-name: vm1
    name: vm1-9rdqx
    namespace: test-migration
    resourceVersion: "33849"
    uid: 9f18f964-5233-4f06-a7db-8f0fad5d77be
  spec:
    vmiName: vm1
  status:
    phase: Succeeded
kind: List
metadata:
  resourceVersion: ""
```

## Observation

At the view of KubeVirt, the VM migration is done, but from K8s' view, the backing POD is still occupying resources. This affects the following migration.

```
$ kubectl describe quota -n test-migration
Name:                   default-gn4d2
Namespace:              test-migration
Resource                Used        Hard
--------                ----        ----
configmaps              1           15
limits.cpu              4           30
limits.memory           9490309124  64424509440
persistentvolumeclaims  4           75
pods                    4           50
replicationcontrollers  0           50
requests.cpu            248m        30
requests.memory         6625599492  60Gi
requests.storage        22Gi        320Gi
secrets                 4           65
services                0           15
services.loadbalancers  0           1
services.nodeports      0           15
```


### Another test: trigger 4 VMs migration at same time

All 4 VMs' spec are topped to resourcequota.

```
$kubectl get resourcequota -n test-migration default-gn4d2 -ojsonpath="{.metadata.annotations}" && kubectl get vmim -A && kubectl get pods -n test-migration

{"harvesterhci.io/migrating-vm1":"{\"limits.cpu\":\"1\",\"limits.memory\":\"2372577281\"}","harvesterhci.io/migrating-vm2":"{\"limits.cpu\":\"1\",\"limits.memory\":\"2372577281\"}","harvesterhci.io/migrating-vm3":"{\"limits.cpu\":\"1\",\"limits.memory\":\"2372577281\"}","harvesterhci.io/migrating-vm4":"{\"limits.cpu\":\"1\",\"limits.memory\":\"2372577281\"}"}NAMESPACE        NAME                        PHASE             VMI
test-migration   kubevirt-evacuation-25wfl   Pending           vm3
test-migration   kubevirt-evacuation-qvdwf   PreparingTarget   vm4
test-migration   kubevirt-evacuation-skzvg   Running           vm1
test-migration   kubevirt-evacuation-vjs9r   Pending           vm2
NAME                      READY   STATUS    RESTARTS   AGE
virt-launcher-vm1-bsxrp   1/1     Running   0          25s
virt-launcher-vm1-lq5bt   1/1     Running   0          4m35s
virt-launcher-vm2-5ld86   1/1     Running   0          4m21s
virt-launcher-vm3-fs7g6   1/1     Running   0          4m4s
virt-launcher-vm4-bplt9   1/1     Running   0          24s
virt-launcher-vm4-jdcqk   1/1     Running   0          3m42s
```

VM1 is removed from resourcesquota annotation, but it's backping pod `virt-launcher-vm1-lq5bt` is still running.

```
$kubectl get resourcequota -n test-migration default-gn4d2 -ojsonpath="{.metadata.annotations}" && kubectl get vmim -A && kubectl get pods -n test-migration
{"harvesterhci.io/migrating-vm2":"{\"limits.cpu\":\"1\",\"limits.memory\":\"2372577281\"}","harvesterhci.io/migrating-vm3":"{\"limits.cpu\":\"1\",\"limits.memory\":\"2372577281\"}","harvesterhci.io/migrating-vm4":"{\"limits.cpu\":\"1\",\"limits.memory\":\"2372577281\"}"}
NAMESPACE        NAME                        PHASE        VMI
test-migration   kubevirt-evacuation-25wfl   Scheduling   vm3
test-migration   kubevirt-evacuation-qvdwf   Running      vm4
test-migration   kubevirt-evacuation-skzvg   Succeeded    vm1
test-migration   kubevirt-evacuation-vjs9r   Pending      vm2
NAME                      READY   STATUS              RESTARTS   AGE
virt-launcher-vm1-bsxrp   1/1     Running             0          27s
virt-launcher-vm1-lq5bt   1/1     Running             0          4m37s // source POD
virt-launcher-vm2-5ld86   1/1     Running             0          4m23s
virt-launcher-vm3-bj8hd   0/1     ContainerCreating   0          1s
virt-launcher-vm3-fs7g6   1/1     Running             0          4m6s
virt-launcher-vm4-bplt9   1/1     Running             0          26s
virt-launcher-vm4-jdcqk   1/1     Running             0          3m44s


$kubectl get resourcequota -n test-migration default-gn4d2 -ojsonpath="{.metadata.annotations}" && kubectl get vmim -A && kubectl get pods -n test-migration
{"harvesterhci.io/migrating-vm2":"{\"limits.cpu\":\"1\",\"limits.memory\":\"2372577281\"}","harvesterhci.io/migrating-vm3":"{\"limits.cpu\":\"1\",\"limits.memory\":\"2372577281\"}","harvesterhci.io/migrating-vm4":"{\"limits.cpu\":\"1\",\"limits.memory\":\"2372577281\"}"}
NAMESPACE        NAME                        PHASE        VMI
test-migration   kubevirt-evacuation-25wfl   Scheduling   vm3
test-migration   kubevirt-evacuation-qvdwf   Running      vm4
test-migration   kubevirt-evacuation-skzvg   Succeeded    vm1
test-migration   kubevirt-evacuation-vjs9r   Pending      vm2
NAME                      READY   STATUS              RESTARTS   AGE
virt-launcher-vm1-bsxrp   1/1     Running             0          28s
virt-launcher-vm1-lq5bt   1/1     Running             0          4m38s // source POD
virt-launcher-vm2-5ld86   1/1     Running             0          4m24s
virt-launcher-vm3-bj8hd   0/1     ContainerCreating   0          2s
virt-launcher-vm3-fs7g6   1/1     Running             0          4m7s
virt-launcher-vm4-bplt9   1/1     Running             0          27s
virt-launcher-vm4-jdcqk   1/1     Running             0          3m45s


$kubectl get resourcequota -n test-migration default-gn4d2 -ojsonpath="{.metadata.annotations}" && kubectl get vmim -A && kubectl get pods -n test-migration
{"harvesterhci.io/migrating-vm2":"{\"limits.cpu\":\"1\",\"limits.memory\":\"2372577281\"}","harvesterhci.io/migrating-vm3":"{\"limits.cpu\":\"1\",\"limits.memory\":\"2372577281\"}"}NAMESPACE        NAME                        PHASE        VMI
test-migration   kubevirt-evacuation-25wfl   Scheduling   vm3
test-migration   kubevirt-evacuation-qvdwf   Succeeded    vm4
test-migration   kubevirt-evacuation-skzvg   Succeeded    vm1
test-migration   kubevirt-evacuation-vjs9r   Pending      vm2
NAME                      READY   STATUS              RESTARTS   AGE
virt-launcher-vm1-bsxrp   1/1     Running             0          29s
virt-launcher-vm1-lq5bt   1/1     Running             0          4m39s // source POD
virt-launcher-vm2-5ld86   1/1     Running             0          4m25s
virt-launcher-vm3-bj8hd   0/1     ContainerCreating   0          3s
virt-launcher-vm3-fs7g6   1/1     Running             0          4m8s
virt-launcher-vm4-bplt9   1/1     Running             0          28s
virt-launcher-vm4-jdcqk   1/1     Running             0          3m46s
```

Around 31s later, the VM1's source POD becomes `Completed`.

```
$kubectl get resourcequota -n test-migration default-gn4d2 -ojsonpath="{.metadata.annotations}" && kubectl get vmim -A && kubectl get pods -n test-migration
{"harvesterhci.io/migrating-vm2":"{\"limits.cpu\":\"1\",\"limits.memory\":\"2372577281\"}","harvesterhci.io/migrating-vm3":"{\"limits.cpu\":\"1\",\"limits.memory\":\"2372577281\"}"}
NAMESPACE        NAME                        PHASE        VMI
test-migration   kubevirt-evacuation-25wfl   Scheduling   vm3
test-migration   kubevirt-evacuation-qvdwf   Succeeded    vm4
test-migration   kubevirt-evacuation-skzvg   Succeeded    vm1
test-migration   kubevirt-evacuation-vjs9r   Pending      vm2
NAME                      READY   STATUS              RESTARTS   AGE
virt-launcher-vm1-bsxrp   1/1     Running             0          31s
virt-launcher-vm1-lq5bt   0/1     Completed           0          4m41s // source POD
virt-launcher-vm2-5ld86   1/1     Running             0          4m27s
virt-launcher-vm3-bj8hd   0/1     ContainerCreating   0          5s
virt-launcher-vm3-fs7g6   1/1     Running             0          4m10s
virt-launcher-vm4-bplt9   1/1     Running             0          30s
virt-launcher-vm4-jdcqk   1/1     Running             0          3m48s
```

### Hint

The migration and resourcequota are from kubevirt and k8s, they have different lifecycle.

The Harvester resourcequota auto-scaling may not fully work as expected.

## Rreproduce the issue

### Resource quota and VMs

3 VMs each with (1C 1Gi), resourcequota is just above the limits.

```
$ kubectl describe resourcequota -n test-migration default-gn4d2 
Name:                   default-gn4d2
Namespace:              test-migration
Resource                Used        Hard
--------                ----        ----
configmaps              1           15
limits.cpu              3           3
limits.memory           3892310016  4404019200
persistentvolumeclaims  6           75
pods                    3           50
replicationcontrollers  0           50
requests.cpu            186m        3
requests.memory         2750460Ki   4200Mi
requests.storage        24Gi        320Gi
secrets                 3           65
services                0           15
services.loadbalancers  0           1
services.nodeports      0           15
```

### Trigger node-maitenance, all 3 VMs are migrated

Finally, all are successful.

```
$kubectl get vmim -A
NAMESPACE        NAME                        PHASE       VMI
test-migration   kubevirt-evacuation-68k62   Succeeded   vm2  // triggered by node-maitenance
test-migration   kubevirt-evacuation-rmnzb   Succeeded   vm1
test-migration   kubevirt-evacuation-wgk69   Succeeded   vm3

test-migration   vm1-dpqbl                   Succeeded   vm1 // manually triggered
test-migration   vm2-2kc6n                   Succeeded   vm2
test-migration   vm3-gm55w                   Succeeded   vm3


$kubectl get vmim -n test-migration kubevirt-evacuation-wgk69 -oyaml
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstanceMigration
metadata:
  annotations:
    kubevirt.io/evacuationMigration: harv2
    kubevirt.io/latest-observed-api-version: v1
    kubevirt.io/storage-observed-api-version: v1alpha3
  creationTimestamp: "2024-12-13T15:25:31Z"
  generateName: kubevirt-evacuation-
  generation: 1
  labels:
    kubevirt.io/vmi-name: vm3
  name: kubevirt-evacuation-wgk69
  namespace: test-migration
  resourceVersion: "180094"
  uid: d8012bc4-76c2-4246-b7ac-97b1bf90fbb7
spec:
  vmiName: vm3
status:
  phase: Succeeded

```

But `virt-controller` prints following error log, it waits until the `used` falls down the hard limit and finally successfully starts the last vm's migration (vm3).

```
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-wgk69","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm3-bj6vb\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:31.314933Z"}

{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-wgk69","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm3-b67k9\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:31.328961Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-wgk69","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm3-zr6fx\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:31.360898Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-wgk69","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm3-jjtz4\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:31.389557Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-wgk69","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm3-szq8l\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:31.436730Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-wgk69","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm3-mpbv5\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:31.523385Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-68k62","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm2-tdzbq\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:31.537147Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-68k62","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm2-tb6k9\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:31.631872Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-68k62","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm2-bzbjt\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:31.731918Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-wgk69","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm3-7v8d8\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:31.831955Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-68k62","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm2-f87jz\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:31.935585Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-68k62","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm2-lfvfm\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:32.031490Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-68k62","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm2-l5tjn\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:32.131426Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-wgk69","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm3-9mlqs\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:32.232578Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-68k62","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm2-sttsc\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:32.332488Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-68k62","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm2-qfjps\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:32.660497Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-wgk69","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm3-ltght\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:32.877693Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-68k62","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm2-kz5mb\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:33.306371Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-wgk69","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm3-jbrth\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:34.164973Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-68k62","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm2-h2f98\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:34.593945Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-wgk69","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm3-q7snc\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:36.731698Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-68k62","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm2-pbx88\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:37.159556Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-wgk69","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm3-n4n94\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:41.857964Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-68k62","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm2-rmkwq\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:42.285731Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-wgk69","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm3-sxb48\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:52.103033Z"}
{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-68k62","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm2-mv4dc\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:25:52.530110Z"}


{"component":"virt-controller","level":"info","msg":"reenqueuing Migration test-migration/kubevirt-evacuation-68k62","pos":"migration.go:199","reason":"failed to create vmi migration target pod: pods \"virt-launcher-vm2-kbn4v\" is forbidden: exceeded quota: default-gn4d2, requested: requests.memory=916820Ki, used: requests.memory=3667280Ki, limited: requests.memory=4200Mi","timestamp":"2024-12-13T15:26:13.015992Z"}

```

### More

The above test VM is very simple. For those production VMs, the migration may take long time and old VM POD also take long time to be evicted and release the resource occupation.


## Another issue: `ResourceQuota` was changed unexpected


first occurance of "exceeded quota" error, it reports :exceeded quota `limits.cpu=30,limits.memory=60Gi`


### The migration started on 2024-12-10T14:05:34

```
2024-12-10T14:05:34:

{"component":"virt-controller","level":"info","msg":"node: lpedge01003, migrations: 0, candidates: 9, selected: 2","pos":"evacuation.go:432","timestamp":"2024-12-10T14:05:34.661377Z"}
{"component":"virt-controller","level":"info","msg":"node: lpedge01003, migrations: 2, candidates: 7, selected: 2","pos":"evacuation.go:432","timestamp":"2024-12-10T14:05:34.694592Z"}

2024-12-10T14:05:34:
{"component":"virt-controller","level":"info","msg":"node: lpedge01003, migrations: 6, candidates: 3, selected: 2","pos":"evacuation.go:432","timestamp":"2024-12-10T14:05:34.727790Z"}
{"component":"virt-controller","level":"info","msg":"node: lpedge01003, migrations: 8, candidates: 1, selected: 1","pos":"evacuation.go:432","timestamp":"2024-12-10T14:05:34.748159Z"}



{"component":"virt-controller","level":"info","msg":"node: ...005, migrations: 5, candidates: 7, selected: 1","pos":"evacuation.go:432","timestamp":"2024-12-10T14:33:04.772536Z"}
{"component":"virt-controller","level":"info","msg":"node: ...005, migrations: 6, candidates: 6, selected: 1","pos":"evacuation.go:432","timestamp":"2024-12-10T14:33:04.797739Z"}
{"component":"virt-controller","level":"info","msg":"node: ...005, migrations: 7, candidates: 5, selected: 1","pos":"evacuation.go:432","timestamp":"2024-12-10T14:33:04.814873Z"}
{"component":"virt-controller","level":"info","msg":"node: ...005, migrations: 8, candidates: 4, selected: 1","pos":"evacuation.go:432","timestamp":"2024-12-10T14:33:04.831214Z"}
```

### vm poc...-f6f0abad-vdrzq stated to migrate on 2024-12-10T14:32:09, used: `cpu=28405`; 6*4Gi + above migration ~ 28Gi

```
{"component":"virt-controller","kind":"","level":"info","msg":"Created migration target pod ...app-dev/virt-launcher-...-f6f0abad-vdrzq-fk9lr with uuid 2d3dd475-c78d-4474-87e4-4e874576d139 for migration kubevirt-evacuation-x4bcq with uuid c315a4ee-0158-4af2-bd07-2a4000398046","name":"...-f6f0abad-vdrzq","namespace":"...-app-dev","pos":"migration.go:717","timestamp":"2024-12-10T14:32:09.596568Z","uid":"8ecd6869-3e1d-4de5-a7fa-2ce389a72b16"}
```

Harvester POD has no failure log, it meant the scaleup was done successfully.

### vm ...-1231906e-gzrdg encountered error `exceeded quota` on 2024-12-10T14:33:04, used: `cpu=28405`; 6*4Gi + above migration ~ 28Gi , why failed?

```
{"component":"virt-controller","level":"info","msg":"node: ...01005, migrations: 6, candidates: 6, selected: 1","pos":"evacuation.go:432","timestamp":"2024-12-10T14:33:04.797739Z"}

{"component":"virt-controller","level":"info","msg":"reenqueuing Migration ...-app-dev/kubevirt-evacuation-m226w","pos":"migration.go:233","reason":"failed to create vmi migration target pod: pods \"virt-launcher-...-1231906e-gzrdg-dprq2\" is forbidden: exceeded quota: default-xt5b2, requested: limits.cpu=4015m,limits.memory=8940039936, used: limits.cpu=28405m,limits.memory=62820279552, limited: limits.cpu=30,limits.memory=60Gi","timestamp":"2024-12-10T14:33:04.808638Z"}

...
```

But now, the limit is reverted to `limits.cpu=30,limits.memory=60Gi`

### the scaled result, looked to be reverted by Rancher

https://github.com/rancher/rancher/blob/b435a2d786c50b03bd1ba7279a6a621ebcd19c84/pkg/controllers/managementuser/resourcequota/resource_quota_sync.go#L114C26-L114C45


### evidence 1: ResourceQuota was updated by `rancher`, and the time-stamp was after the above migration

It means Rancher update it several times, and `2024-12-10T15:36:44Z` was the last updating time.

Last updating time from Harvester `2024-12-10T16:18:37Z`.

```
- apiVersion: v1
  kind: ResourceQuota
  metadata:
    creationTimestamp: "2024-07-24T09:40:09Z"
    generateName: default-
    labels:
      cattle.io/creator: norman
      resourcequota.management.cattle.io/default-resource-quota: "true"
    managedFields:
    - apiVersion: v1
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:generateName: {}
          f:labels:
            .: {}
            f:cattle.io/creator: {}
            f:resourcequota.management.cattle.io/default-resource-quota: {}
        f:spec:
          f:hard:
            .: {}
            f:configmaps: {}
            f:persistentvolumeclaims: {}
            f:pods: {}
            f:replicationcontrollers: {}
            f:requests.cpu: {}
            f:requests.memory: {}
            f:requests.storage: {}
            f:secrets: {}
            f:services: {}
            f:services.loadbalancers: {}
            f:services.nodeports: {}
      manager: rancher
      operation: Update
      time: "2024-12-10T15:36:44Z"
    - apiVersion: v1
      fieldsType: FieldsV1
      fieldsV1:
        f:spec:
          f:hard:
            f:limits.cpu: {}
            f:limits.memory: {}
      manager: harvester
      operation: Update
      time: "2024-12-10T16:18:37Z"
    - apiVersion: v1
      fieldsType: FieldsV1
      fieldsV1:
        f:status:
          f:hard:
            .: {}
            f:configmaps: {}
            f:limits.cpu: {}
            f:limits.memory: {}
            f:persistentvolumeclaims: {}
            f:pods: {}
            f:replicationcontrollers: {}
            f:requests.cpu: {}
            f:requests.memory: {}
            f:requests.storage: {}
            f:secrets: {}
            f:services: {}
            f:services.loadbalancers: {}
            f:services.nodeports: {}
          f:used:
            .: {}
            f:configmaps: {}
            f:limits.cpu: {}
            f:limits.memory: {}
            f:pods: {}
            f:replicationcontrollers: {}
            f:requests.cpu: {}
            f:requests.memory: {}
            f:services: {}
            f:services.loadbalancers: {}
            f:services.nodeports: {}
      manager: kube-controller-manager
      operation: Update
      subresource: status
      time: "2024-12-10T16:19:31Z"
```


### evidence 2: ...-app-dev has following error, the `limits.cpu=26`


```
...-app-dev/kubevirt-evacuation-dfp6

error:

exceeded quota: default-4tmmk, 

requested: limits.cpu=4015m,    limits.memory= 8,940,039,936,
used:      limits.cpu=28505m,   limits.memory=62,900,279,552,
limited:   limits.cpu=26,       limits.memory=54,242,646Ki",      "timestamp":"2024-12-10T14:36:47.533782Z
```

How? Rancher reverted the quota `limits.cpu=30`, and after migration, Harvester scaleddown it to `30-4=26 Core(each VM is 4 Core)`.


### Test 1: manually kill cattle-cluster-agent pod, Rancher Manager reverts the RQ change

1. create a project and namespace test1sub1 with RQ

```
$kubectl get namespaces test1sub1 -oyaml
apiVersion: v1
kind: Namespace
metadata:
  annotations:
    cattle.io/status: '{"Conditions":[{"Type":"ResourceQuotaValidated","Status":"True","Message":"","LastUpdateTime":"2024-12-16T12:37:18Z"},{"Type":"ResourceQuotaInit","Status":"True","Message":"","LastUpdateTime":"2024-12-16T12:37:18Z"},{"Type":"InitialRolesPopulated","Status":"True","Message":"","LastUpdateTime":"2024-12-16T12:37:18Z"}]}'
    field.cattle.io/containerDefaultResourceLimit: '{}'
    field.cattle.io/projectId: c-m-nwmsdwdc:p-snkkp
    field.cattle.io/resourceQuota: '{"limit":{"limitsCpu":"2000m"}}'
    lifecycle.cattle.io/create.namespace-auth: "true"
  creationTimestamp: "2024-12-16T12:37:16Z"
  finalizers:
  - controller.cattle.io/namespace-auth
  labels:
    field.cattle.io/projectId: p-snkkp
    kubernetes.io/metadata.name: test1sub1
  name: test1sub1
  resourceVersion: "357044"
  uid: f1838a3e-f1d7-42de-b322-b7f593efadba
spec:
  finalizers:
  - kubernetes
status:
  phase: Active
```

2. The RQ is as expected

```
$kubectl get resourcequota -n test1sub1 default-k4n2w -oyaml
apiVersion: v1
kind: ResourceQuota
metadata:
  creationTimestamp: "2024-12-16T12:37:17Z"
  generateName: default-
  labels:
    cattle.io/creator: norman
    resourcequota.management.cattle.io/default-resource-quota: "true"
  name: default-k4n2w
  namespace: test1sub1
  resourceVersion: "357030"
  uid: b73ef61c-a946-4cca-ae7a-71e26d5e9d42
spec:
  hard:
    limits.cpu: "2"
status:
  hard:
    limits.cpu: "2"
  used:
    limits.cpu: "0"
harv21:/home/rancher # kk edit resourcequota -n test1sub1 default-k4n2w
resourcequota/default-k4n2w edited
```

3. Update he RQ from kubectl

```
$kubectl get resourcequota -n test1sub1 default-k4n2w -oyaml
apiVersion: v1
kind: ResourceQuota
metadata:
  creationTimestamp: "2024-12-16T12:37:17Z"
  generateName: default-
  labels:
    cattle.io/creator: norman
    resourcequota.management.cattle.io/default-resource-quota: "true"
  name: default-k4n2w
  namespace: test1sub1
  resourceVersion: "361636"
  uid: b73ef61c-a946-4cca-ae7a-71e26d5e9d42
spec:
  hard:
    limits.cpu: "3"
status:
  hard:
    limits.cpu: "3"
  used:
    limits.cpu: "0"
    
$kubectl get resourcequota -n test1sub1 default-k4n2w -oyaml
apiVersion: v1
kind: ResourceQuota
metadata:
  creationTimestamp: "2024-12-16T12:37:17Z"
  generateName: default-
  labels:
    cattle.io/creator: norman
    resourcequota.management.cattle.io/default-resource-quota: "true"
  name: default-k4n2w
  namespace: test1sub1
  resourceVersion: "361636"
  uid: b73ef61c-a946-4cca-ae7a-71e26d5e9d42
spec:
  hard:
    limits.cpu: "3"
status:
  hard:
    limits.cpu: "3"
  used:
    limits.cpu: "0"
```


4. Kill `cattle-cluster-agent` POD, the RQ is reverted.

```
kk delete pod -n cattle-system cattle-cluster-agent-769854d75d-lf9d9 --force


$kubectl get resourcequota -n test1sub1 default-k4n2w -oyaml
apiVersion: v1
kind: ResourceQuota
metadata:
  creationTimestamp: "2024-12-16T12:37:17Z"
  generateName: default-
  labels:
    cattle.io/creator: norman
    resourcequota.management.cattle.io/default-resource-quota: "true"
  name: default-k4n2w
  namespace: test1sub1
  resourceVersion: "363918"
  uid: b73ef61c-a946-4cca-ae7a-71e26d5e9d42
spec:
  hard:
    limits.cpu: "2"
status:
  hard:
    limits.cpu: "2"
  used:
    limits.cpu: "0"
```


### Test 2: Logs from SB show `cattle-cluster-agent*` POD started several times in-between above VM migration time window


In following time `I1210 14:05:40, I1210 14:12:17, I1210 15:20:41`, the `cattle-cluster-agent*` POD restarted.

And per the test, it had synced the resourcequota and reverted the values even though Harvester assumed they shoud have been scaled.


```
I1210 14:05:40

/...005/logs/kubelet.log:I1210 14:05:40.506853    3806 topology_manager.go:210] "Topology Admit Handler" podUID=37c586f4-2a43-47e1-85f5-c1faf000f034 podNamespace="cattle-system" podName="cattle-cluster-agent-5cd5778bd5-5fhsz"
./...005/logs/kubelet.log:I1210 14:05:40.612266    3806 reconciler_common.go:253] "operationExecutor.VerifyControllerAttachedVolume started for volume \"kube-api-access-c2qfn\" (UniqueName: \"kubernetes.io/projected/37c586f4-2a43-47e1-85f5-c1faf000f034-kube-api-access-c2qfn\") pod \"cattle-cluster-agent-5cd5778bd5-5fhsz\" (UID: \"37c586f4-2a43-47e1-85f5-c1faf000f034\") " pod="cattle-system/cattle-cluster-agent-5cd5778bd5-5fhsz"
./...005/logs/kubelet.log:I1210 14:05:40.612332    3806 reconciler_common.go:253] "operationExecutor.VerifyControllerAttachedVolume started for volume \"cattle-credentials\" (UniqueName: \"kubernetes.io/secret/37c586f4-2a43-47e1-85f5-c1faf000f034-cattle-credentials\") pod \"cattle-cluster-agent-5cd5778bd5-5fhsz\" (UID: \"37c586f4-2a43-47e1-85f5-c1faf000f034\") " pod="cattle-system/cattle-cluster-agent-5cd5778bd5-5fhsz"
./...005/logs/kubelet.log:I1210 14:05:42.131364    3806 pod_startup_latency_tracker.go:102] "Observed pod startup duration" pod="cattle-system/cattle-cluster-agent-5cd5778bd5-5fhsz" podStartSLOduration=2.131321639 pod.CreationTimestamp="2024-12-10 14:05:40 +0000 UTC" firstStartedPulling="0001-01-01 00:00:00 +0000 UTC" lastFinishedPulling="0001-01-01 00:00:00 +0000 UTC" observedRunningTime="2024-12-10 14:05:42.128499715 +0000 UTC m=+685453.304479175" watchObservedRunningTime="2024-12-10 14:05:42.131321639 +0000 UTC m=+685453.307301089"



I1210 14:12:17

./...005/logs/kubelet.log:I1210 14:12:17.656694    3806 topology_manager.go:210] "Topology Admit Handler" podUID=29614ee5-e61e-4719-ac99-0ccc5afd2bf4 podNamespace="cattle-system" podName="cattle-cluster-agent-5f974d864f-s5xw6"
./...005/logs/kubelet.log:I1210 14:12:17.761958    3806 reconciler_common.go:253] "operationExecutor.VerifyControllerAttachedVolume started for volume \"kube-api-access-f8c89\" (UniqueName: \"kubernetes.io/projected/29614ee5-e61e-4719-ac99-0ccc5afd2bf4-kube-api-access-f8c89\") pod \"cattle-cluster-agent-5f974d864f-s5xw6\" (UID: \"29614ee5-e61e-4719-ac99-0ccc5afd2bf4\") " pod="cattle-system/cattle-cluster-agent-5f974d864f-s5xw6"
./...005/logs/kubelet.log:I1210 14:12:17.762147    3806 reconciler_common.go:253] "operationExecutor.VerifyControllerAttachedVolume started for volume \"cattle-credentials\" (UniqueName: \"kubernetes.io/secret/29614ee5-e61e-4719-ac99-0ccc5afd2bf4-cattle-credentials\") pod \"cattle-cluster-agent-5f974d864f-s5xw6\" (UID: \"29614ee5-e61e-4719-ac99-0ccc5afd2bf4\") " pod="cattle-system/cattle-cluster-agent-5f974d864f-s5xw6"
./...005/logs/kubelet.log:I1210 14:12:18.657683    3806 pod_startup_latency_tracker.go:102] "Observed pod startup duration" pod="cattle-system/cattle-cluster-agent-5f974d864f-s5xw6" podStartSLOduration=1.657600784 pod.CreationTimestamp="2024-12-10 14:12:17 +0000 UTC" firstStartedPulling="0001-01-01 00:00:00 +0000 UTC" lastFinishedPulling="0001-01-01 00:00:00 +0000 UTC" observedRunningTime="2024-12-10 14:12:18.654627621 +0000 UTC m=+685849.830607088" watchObservedRunningTime="2024-12-10 14:12:18.657600784 +0000 UTC m=+685849.833580234"



I1210 15:20:41

./...005/logs/kubelet.log:I1210 15:20:41.541881    3758 topology_manager.go:212] "Topology Admit Handler" podUID=9443907a-f47d-489d-b620-7ba4e69b4d10 podNamespace="cattle-system" podName="cattle-cluster-agent-5cd5778bd5-xv4lc"
./...005/logs/kubelet.log:I1210 15:20:41.729973    3758 reconciler_common.go:258] "operationExecutor.VerifyControllerAttachedVolume started for volume \"cattle-credentials\" (UniqueName: \"kubernetes.io/secret/9443907a-f47d-489d-b620-7ba4e69b4d10-cattle-credentials\") pod \"cattle-cluster-agent-5cd5778bd5-xv4lc\" (UID: \"9443907a-f47d-489d-b620-7ba4e69b4d10\") " pod="cattle-system/cattle-cluster-agent-5cd5778bd5-xv4lc"
./...005/logs/kubelet.log:I1210 15:20:41.730006    3758 reconciler_common.go:258] "operationExecutor.VerifyControllerAttachedVolume started for volume \"kube-api-access-7ggk8\" (UniqueName: \"kubernetes.io/projected/9443907a-f47d-489d-b620-7ba4e69b4d10-kube-api-access-7ggk8\") pod \"cattle-cluster-agent-5cd5778bd5-xv4lc\" (UID: \"9443907a-f47d-489d-b620-7ba4e69b4d10\") " pod="cattle-system/cattle-cluster-agent-5cd5778bd5-xv4lc"
./...005/logs/kubelet.log:I1210 15:20:42.633403    3758 pod_startup_latency_tracker.go:102] "Observed pod startup duration" pod="cattle-system/cattle-cluster-agent-5cd5778bd5-xv4lc" podStartSLOduration=1.633361641 podCreationTimestamp="2024-12-10 15:20:41 +0000 UTC" firstStartedPulling="0001-01-01 00:00:00 +0000 UTC" lastFinishedPulling="0001-01-01 00:00:00 +0000 UTC" observedRunningTime="2024-12-10 15:20:42.632501117 +0000 UTC m=+616.495489753" watchObservedRunningTime="2024-12-10 15:20:42.633361641 +0000 UTC m=+616.496350272"
```

### Short summary

The VM-migration was affected by the relacement of POD `cattle-cluster-agent`, the latter reverted the RQ scaling.

When drain/maintain the node, the `cattle-cluster-agent` PODs may also be affected and replaced to another node, and they run in parallel with VM migration, when new `cattle-cluster-agent` POD/Rancher Manager reverts the RQ after Harvester has scalled it, then the new VM migration target POD may encounter error `exceeded quota`. Sure, it does not always happen.

When following conditions are met, the issue may happen:

- The ResourceQuota for a given namespace has been used 80+%.
- A couple of VMs from this namespace are running on a same node.
- The `cattle-cluster-agent` related PODs are also running on this node.

When encountering this issue, manual stop the VM migration and trigger the migration again can solve it.
