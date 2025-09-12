

## Brief introduction about `rancher-monitoring-crd` managedchart

The `rancher-monitoring-crd` managedchart is installed by default and never deleted in the Harvester cluster lifecycle, it deploys the required CRDs. Later, when enable/disable the `rancher-monitoring` add-on, the won't be blocked due to missing CRDs.

Harvester upgrades `rancher-monitoring-crd` managedchart in the upgrade path automatically.


### The pre-installed `managedcharts`, don't delete them

```
fleet-local/rancher-monitoring-crd
fleet-local/rancher-logging-crd
fleet-local/harvester-crd
fleet-local/harvester
```

## The versions

```
Harvester rancher-monitoring-crd

v1.3.0 103.0.3+up45.31.1
v1.3.1 103.0.3+up45.31.1
v1.3.2 103.0.3+up45.31.1

v1.4.0 103.1.1+up45.31.1
v1.4.1 103.1.1+up45.31.1
v1.4.2 103.1.1+up45.31.1
```

## Manual re-create the managedcharts

### Check exsting managedcharts

```
kubectl get managedcharts -A

```

If the `rancher-monitoring-crd` managedchart was deleted accidentally and hence missing, run below command to re-create.

### Re-create on Harvester v1.3.* cluster

```
cat > rmc-v13.yaml << 'EOF'

apiVersion: management.cattle.io/v3
kind: ManagedChart
metadata:
  name: rancher-monitoring-crd
  namespace: fleet-local
spec:
  chart: rancher-monitoring-crd
  releaseName: rancher-monitoring-crd
  version: 103.0.3+up45.31.1
  defaultNamespace: cattle-monitoring-system
  repoName: harvester-charts
  timeoutSeconds: 600
  targets:
  - clusterName: local
    clusterSelector:
      matchExpressions:
      - key: provisioning.cattle.io/unmanaged-system-agent
        operator: DoesNotExist
EOF

kubectl create -f rmc-v13.yaml
```

### Re-create on Harvester v1.4.* cluster

```
cat > rmc-v14.yaml << 'EOF'
apiVersion: management.cattle.io/v3
kind: ManagedChart
metadata:
  name: rancher-monitoring-crd
  namespace: fleet-local
spec:
  chart: rancher-monitoring-crd
  releaseName: rancher-monitoring-crd
  version: 103.1.1+up45.31.1
  defaultNamespace: cattle-monitoring-system
  repoName: harvester-charts
  timeoutSeconds: 600
  targets:
  - clusterName: local
    clusterSelector:
      matchExpressions:
      - key: provisioning.cattle.io/unmanaged-system-agent
        operator: DoesNotExist

EOF

kubectl create -f rmc-v14.yaml
```
