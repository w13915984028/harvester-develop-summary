
## How to update a subchart on Harvester

On issue like https://github.com/harvester/harvester/issues/9452, we need to bump the harvester-load-balancer rbac, it is worked via the `harvester/charts` repo.


### Local update and test

To be updated

### How to release the chart

#### Understand the branches on `harvester/charts`

`master`: for latest-code.

`release`: for the latest developing branch, for daily developing, `master` and `release` are the main branches.

`release-v1.7`: for a near-release/already-released branch.

#### Update chart to master branch

The chart change first targets to the master branch.

Example: https://github.com/harvester/charts/pull/469

:::note

A chart tgz file is not released at this moment.

:::

#### Update chart to release branch

Need to manually cherry-pick the master branch PR to release branch, and also remember to update the chart.yaml to have a new minor version

Example: https://github.com/harvester/charts/pull/477
Example: https://github.com/harvester/charts/pull/478

On the happy path, a new chart tgz file like [harvester-cloud-provider 1.8.0-dev.1](https://github.com/harvester/charts/releases/tag/harvester-load-balancer-1.8.0-dev.1) is released.

#### Update chart to Harvester

The sub-chart is bundled onto Harvester `managedchart`, some additional work is needed.

1. Manually update the Chart.yaml to new version

```
/go/src/github.com/harvester/harvester/deploy/charts/harvester$ git diff
diff --git a/deploy/charts/harvester/Chart.yaml b/deploy/charts/harvester/Chart.yaml
index c370f69e7..52e290f56 100644
--- a/deploy/charts/harvester/Chart.yaml
+++ b/deploy/charts/harvester/Chart.yaml
@@ -57,7 +57,7 @@ dependencies:
     version: 0.8.0
     repository: https://kube-vip.github.io/helm-charts
   - name: harvester-load-balancer
-    version: 1.8.0-dev.0
+    version: 1.8.0-dev.1
     repository: https://charts.harvesterhci.io
   - name: whereabouts
     version: 0.1.1
```

2. Run `helm dependency update .` to update.

```     
/go/src/github.com/harvester/harvester/deploy/charts/harvester$ helm dependency update .
Getting updates for unmanaged Helm repositories...
...Successfully got an update from the "https://kube-vip.github.io/helm-charts" chart repository
...Successfully got an update from the "https://charts.longhorn.io" chart repository
...Successfully got an update from the "https://charts.harvesterhci.io" chart repository
...Successfully got an update from the "https://charts.harvesterhci.io" chart repository
...Successfully got an update from the "https://charts.harvesterhci.io" chart repository
...Successfully got an update from the "https://charts.harvesterhci.io" chart repository
...Successfully got an update from the "https://charts.harvesterhci.io" chart repository
Hang tight while we grab the latest from your chart repositories...
...Unable to get an update from the "rancher-logging-local" chart repository (http://192.168.2.59:8000/patched-charts):
	Get "http://192.168.2.59:8000/patched-charts/index.yaml": dial tcp 192.168.2.59:8000: connect: connection refused
Update Complete. ⎈Happy Helming!⎈
Saving 12 charts
Downloading harvester-network-controller from repo https://charts.harvesterhci.io
Downloading harvester-networkfs-manager from repo https://charts.harvesterhci.io
Downloading harvester-node-disk-manager from repo https://charts.harvesterhci.io
Downloading longhorn from repo https://charts.longhorn.io
Downloading kube-vip from repo https://kube-vip.github.io/helm-charts
Downloading harvester-load-balancer from repo https://charts.harvesterhci.io
Downloading harvester-node-manager from repo https://charts.harvesterhci.io
Deleting outdated charts


git diff
diff --git a/deploy/charts/harvester/Chart.lock b/deploy/charts/harvester/Chart.lock
index 0f19ea377..60a792d8d 100644
--- a/deploy/charts/harvester/Chart.lock
+++ b/deploy/charts/harvester/Chart.lock
@@ -28,12 +28,12 @@ dependencies:
   version: 0.8.0
 - name: harvester-load-balancer
   repository: https://charts.harvesterhci.io
-  version: 1.8.0-dev.0
+  version: 1.8.0-dev.1
 - name: whereabouts
   repository: file://dependency_charts/whereabouts
   version: 0.1.1
 - name: harvester-node-manager
   repository: https://charts.harvesterhci.io
   version: 1.8.0-dev.0
-digest: sha256:4eab95b00d4bcb4cfbeead858004cff372b35bddfd7c1eb09e661a3438000a0d
-generated: "2026-02-13T10:20:40.768011+08:00"
+digest: sha256:eef00f10915871d732ef0b46fa30f81c3a373e1a4895429a3d6bd418eaf759bc
+generated: "2026-02-16T11:01:19.757477267+01:00"
```

3. File a PR to merge the change

https://github.com/harvester/harvester/pull/10063

#### Test via new ISO

Build a new ISO to test the changes.

When cluster is initialized, the managedcharts `fleet-local/harvester-crd` and `fleet-local/harvester` are installed, the above changes are deployed/released via them.

