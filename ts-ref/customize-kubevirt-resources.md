

## Issue and workaround

Upstream issue: https://github.com/kubevirt/kubevirt/issues/13295 virt-handler memory consumption is very high

Sometimes, the `virt-handler` pod might shows taking big amount of memory.

Upstream assumed this is not an issue.

https://github.com/kubevirt/kubevirt/issues/13295#issuecomment-2587225726

```
I believe this was additionally discussed on slack and the conclusion was that the observed memory was virtual (reclaimable, more specifically it was cache) and not rss.
```

To avoid the POD is OOM killed, increase the `virt-handler` memory limitation is a solution to further monitor if the memory usage still goes up.

## Update Harvester managedchart to avoid warnings

Use below command to update the managedchart.

`$ kubectl  edit managedchart -n fleet-local harvester`

Add following fields after `comparePatches` to the managedchart.

```
spec:
  chart: harvester
  defaultNamespace: harvester-system
  diff:
    comparePatches:


    // add following 5 lines

       
    - apiVersion: kubevirt.io/v1
      kind: KubeVirt
      name: kubevirt
      jsonPointers:
      - /spec/customizeComponents/patches
...
```

:::note

Without this step, following changes might cause the `managedchart/bundle` is marked as changed, and affect the continuous upgrade.

:::


## Update kubevirt object

Use below command to update the kubevirt.

```
$ kubectl  edit kubevirt kubevirt -n harvester-system
```


```
...
  customizeComponents:
    patches:
...
    - patch: '{"spec":{"template":{"spec":{"containers":[{"name":"virt-handler", "resources":{"limits":{"cpu":"700m","memory":"1800Mi"}}}]}}}}'                   // update the memory from e.g. 1600Mi to 1800MI
      resourceName: virt-handler
      resourceType: DaemonSet
      type: strategic
....      
```

Wait some while, check the new pod.


```
$ kubectl get pods -n harvester-system
NAME                                                   READY   STATUS    RESTARTS      AGE
...
virt-api-6897cf4644-7lx7s                              1/1     Running   2 (42m ago)   2d6h
virt-controller-64d7894974-kfvml                       1/1     Running   2 (42m ago)   2d6h
virt-controller-64d7894974-pf8kv                       1/1     Running   2 (42m ago)   2d6h
virt-handler-4kp7s                                     1/1     Running   0             9m59s    // the replaced new pod
virt-operator-67754596d9-2rf9t                         1/1     Running   3 (42m ago)   2d6h
```

Check the pod resources limitation, it is as expectation.

```
$ kubectl  get pods -n harvester-system virt-handler-4kp7s -oyaml | grep limit -2
      timeoutSeconds: 10
    resources:
      limits:
        cpu: 700m
        memory: 1800Mi      // it is the new value
--
    ready: true
    resources:
      limits:
        cpu: 700m
        memory: 1800Mi

```


Check Harvester does not complain about the managedchart & bundle.

```
$ kubectl get bundle -A

NAMESPACE     NAME                         BUNDLEDEPLOYMENTS-READY   STATUS
fleet-local   fleet-agent-local            1/1                       
fleet-local   mcc-harvester                1/1                                        // no complain, otherwise, it will affect next round upgrade
fleet-local   mcc-harvester-crd            1/1                       
fleet-local   mcc-kubeovn-operator-crd     1/1                       
fleet-local   mcc-rancher-logging-crd      1/1                       
fleet-local   mcc-rancher-monitoring-crd   1/1                       
```

