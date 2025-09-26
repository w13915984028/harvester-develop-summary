# Config additionalGuestMemoryOverheadRatio on Harvester v132

## Scenario

If the VM suffers [OOM issue](https://github.com/w13915984028/harvester-develop-summary/blob/main/oom-related-issue-investigation.md) on Harvester v1.3.* cluster, the param `additionalGuestMemoryOverheadRatio` can be adjusted to set a higher overhead room for VM to eliminate the OOM effectively.

## Solution

### Edit kubevirt to set an additionalGuestMemoryOverheadRatio value

When additionalGuestMemoryOverheadRatio is not set, it uses the default value "1".

```
kubectl edit kubevirt -n harvester-system kubevirt


spec:
  configuration:
    additionalGuestMemoryOverheadRatio: "1.5"  // set this filed to a string value
```

:::note

- This is applied to the kubevirt CRD object, it persists after system rebooting/upgrading.

- This is a global setting, it affects all VM.

- It takes effects when a VM starts/reboots, which leads to the VM's backend POD is newly created, this param is used to compute the memory limits and then apply to the POD.

- When the cluster is upgraded from v1.3.2 to v1.4.0, Harvester will inheriate the existing valid value (empty, "0", "1.0" ~ "10.0") and convert it to a Harvester setting.

:::

Following tests are done on Harvester v1.3.2 cluster.

### Test 1: VM with 16Gi Requested Memory

VM requests memory = 16Gi; guest OS available memory = 16284Mi

```
...
$ kubectl get vm vmname -oyaml

          memory:
            guest: 16284Mi     // available to guest OS, Harvester reserves 100Mi by default if you did not set "Reserved Memory" field
          resources:
            limits:
              cpu: "1"
              memory: 16Gi     // vm requests 
            requests:
              cpu: 62m
              memory: 3276Mi
...
```


1. additionalGuestMemoryOverheadRatio="1" (default), POD limits = 16632 Mi; total overhead = 348 Mi


```
...
$ kubectl get pod virt-launcher-vmname-** -oyaml

      name: compute
      resources:
        limits:
          cpu: "1"
          devices.kubevirt.io/kvm: "1"
          devices.kubevirt.io/tun: "1"
          devices.kubevirt.io/vhost-net: "1"
          memory: 17031576Ki                     // ~ 16632 Mi
        requests:
          cpu: 62m
          devices.kubevirt.io/kvm: "1"
          devices.kubevirt.io/tun: "1"
          devices.kubevirt.io/vhost-net: "1"
          ephemeral-storage: 50M
          memory: 3608984Ki
...
```

2. additionalGuestMemoryOverheadRatio="1.5", POD limits = 16756 Mi; total overhead = 472 Mi

```

      name: compute
      resources:
        limits:
          cpu: "1"
          devices.kubevirt.io/kvm: "1"
          devices.kubevirt.io/tun: "1"
          devices.kubevirt.io/vhost-net: "1"
          memory: "17570565824"                  // ~ 16756 Mi
        requests:
          cpu: 62m
          devices.kubevirt.io/kvm: "1"
          devices.kubevirt.io/tun: "1"
          devices.kubevirt.io/vhost-net: "1"
          ephemeral-storage: 50M
          memory: "3825831616"      
```

3. additionalGuestMemoryOverheadRatio="2.0", POD limits = 16880 Mi; total overhead = 596 Mi

```
...
      resources:
        limits:
          cpu: "1"
          devices.kubevirt.io/kvm: "1"
          devices.kubevirt.io/tun: "1"
          devices.kubevirt.io/vhost-net: "1"
          memory: "17700798824"                  // ~ 16880 Mi
        requests:
          cpu: 62m
          devices.kubevirt.io/kvm: "1"
          devices.kubevirt.io/tun: "1"
          devices.kubevirt.io/vhost-net: "1"
          ephemeral-storage: 50M
          memory: "3956064616"
...
```

### Test 2: VM with 8Gi Requested Memory

VM requests memory = 8Gi, guest OS available = 8092Mi

```
          memory:
            guest: 8092Mi
          resources:
            limits:
              cpu: "1"
              memory: 8Gi
            requests:
              cpu: 62m
              memory: 1638Mi              
```


1. additionalGuestMemoryOverheadRatio="1" (default), POD limits = 8437 Mi; total overhead = 345 Mi

```
      name: compute
      resources:
        limits:
          cpu: "1"
          devices.kubevirt.io/kvm: "1"
          devices.kubevirt.io/tun: "1"
          devices.kubevirt.io/vhost-net: "1"
          memory: "8847044609"                  // ~ 8437 Mi
        requests:
          cpu: 62m
          devices.kubevirt.io/kvm: "1"
          devices.kubevirt.io/tun: "1"
          devices.kubevirt.io/vhost-net: "1"
          ephemeral-storage: 50M
          memory: "1974677505"
```


2. additionalGuestMemoryOverheadRatio="1.5", POD limits = 8559 Mi; total overhead = 467 Mi

```
      name: compute
      resources:
        limits:
          cpu: "1"
          devices.kubevirt.io/kvm: "1"
          devices.kubevirt.io/tun: "1"
          devices.kubevirt.io/vhost-net: "1"
          memory: "8975599609"                  // ~ 8559 Mi
        requests:
          cpu: 62m
          devices.kubevirt.io/kvm: "1"
          devices.kubevirt.io/tun: "1"
          devices.kubevirt.io/vhost-net: "1"
          ephemeral-storage: 50M
          memory: "2103232505"        
```

3. additionalGuestMemoryOverheadRatio="2.0", POD limits = 8682 Mi; total overhead = 590 Mi

```
      name: compute
      resources:
        limits:
          cpu: "1"
          devices.kubevirt.io/kvm: "1"
          devices.kubevirt.io/tun: "1"
          devices.kubevirt.io/vhost-net: "1"
          memory: "9104155609"                  // ~ 8682 Mi
        requests:
          cpu: 62m
          devices.kubevirt.io/kvm: "1"
          devices.kubevirt.io/tun: "1"
          devices.kubevirt.io/vhost-net: "1"
          ephemeral-storage: 50M
          memory: "2231788505"          
```

### Summary

The overhead is related to VM's specifications like devices types and numbers. QEMU, driver and others need memory to simulate everything for the VM. It is very hard to have an one-fit-all solution.

With a higher `additionalGuestMemoryOverheadRatio`, Harvester will sets more overhead memory for the VM to reduce the chance of hitting OOM. At the meantime, it also means the available memory to other workloads becomes less. A good balanced value is achieved via testing and tuning.

The current known best practice is: set the `additionalGuestMemoryOverheadRatio` to around `"2"`, and for some specific/critical VMs, update the [VM Reserved Memory](https://docs.harvesterhci.io/v1.4/vm/index#reserved-memory) to give them an even higher than average overhead memory.

## Known Limitations

### Resource Quota

Together with Rancher, Harvester supports to use the `Resource Quota` to control the resources on a namespace/project, and it has a [auto-scaling](https://docs.harvesterhci.io/v1.3/rancher/resource-quota#automatic-adjustment-of-resourcequota-during-migration) feature when VM is migrating.

The `auto-scaling` will scale up the quota to make rooms for the migration target VM, and the calculation uses the default `additionalGuestMemoryOverheadRatio` value `"1"`, but the real VM's POD will be based on the above set value of `additionalGuestMemoryOverheadRatio` e.g. `"2"`. A gap may exist.

In exteme case, the migration may fail due to insufficient memory quota. Ensure the `Resource Quota` has enough free spaces before take actions like `system ugrade` and `host maintenance`.

https://github.com/harvester/harvester/blob/939857e93c3d97de47f497e719ff219fe4df81ca/pkg/controller/master/migration/vmim_controller.go#L161C20-L161C37

https://github.com/harvester/harvester/blob/939857e93c3d97de47f497e719ff219fe4df81ca/pkg/util/resourcequota/calculator.go#L200

The limitation does not exist on Harvester v1.4.2 and later versions.

## Referrences

[VM OOM](https://github.com/w13915984028/harvester-develop-summary/blob/main/oom-related-issue-investigation.md)

[VM Reserved Memory](https://docs.harvesterhci.io/v1.4/vm/index#reserved-memory)

[Harvester setting additional-guest-memory-overhead-ratio](https://docs.harvesterhci.io/v1.4/advanced/index#additional-guest-memory-overhead-ratio)


upstream source code

https://github.com/kubevirt/kubevirt/blob/9d0345041704fe0d12a7cbb8eeecb2823d1cb703/pkg/virt-controller/services/template.go#L114-L123

https://github.com/kubevirt/kubevirt/blob/9d0345041704fe0d12a7cbb8eeecb2823d1cb703/pkg/virt-controller/services/renderresources.go#L401-L419

