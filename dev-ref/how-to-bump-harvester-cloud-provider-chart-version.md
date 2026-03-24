How to release a new harvester-cloud-provider chart version and bump it to the upstream RKE2 and rancher-charts.

## Release a new chart on Harvester chart repo

e.g. The harvester-cloud-provider chart.

### Change an existing Chart

e.g. https://github.com/harvester/charts/pull/229

merge to `master` branch

Note:
  The chart version on `master` is `stable`, e.g. https://github.com/harvester/charts/blob/9cd25bd54370e38b3d2a1731c317c418d079b8ff/charts/harvester-cloud-provider/Chart.yaml#L21, it keeps as `version: 0.0.0-dev`
  Don't change it

### Relese a new version

Cherry-pick all changes upon the chart on `master` branch, and add another commit to amend the `version` field on `Chart.yaml`

After the PR to `release` branch is merged, a new chart version is released automatically.

e.g. https://github.com/harvester/charts/pull/230

e.g. https://github.com/harvester/charts/releases/tag/harvester-cloud-provider-0.2.4

The chart like `harvester-cloud-provider` and `harvester-csi-driver` also need to be bumped to upstream RKE2 and Rancher. The following secitons describe the details.

## Bump the chart to `RKE2`

### Update `rancher/rke2-charts`

Point to the new release on Harvester chart repo above.

https://github.com/rancher/rke2-charts/blob/main-source/packages/harvester-cloud-provider/package.yaml

e.g. https://github.com/rancher/rke2-charts/pull/454

### Update 'rancher/rke2` to use new charts as the default value

Point to the new chart version.

https://github.com/rancher/rke2/blob/master/charts/chart_versions.yaml

e.g. https://github.com/rancher/rke2/pull/5980

note: Add PR to each active rke2 branch.

## Bump the chart to `Rancher`

### Bump the chart to `rancher/charts`

First, read the document on this chart.

Then, follow the example: https://github.com/rancher/charts/pull/3234

we need to use those 2 paths to step-by-step generate the target file

```
  // used to download chart file in step (0)
	packages/harvester/harvester-cloud-provider/charts/

	// used to put the generated patch files
	packages/harvester/harvester-cloud-provider/generated-changes/
```

(0) update the package.yaml file first;  update the `url` and `version`

```
diff --git a/packages/harvester/harvester-cloud-provider/package.yaml b/packages/harvester/harvester-cloud-provider/package.yaml
index 7e86c3d2..5485f225 100644
--- a/packages/harvester/harvester-cloud-provider/package.yaml
+++ b/packages/harvester/harvester-cloud-provider/package.yaml
@@ -1,3 +1,3 @@
-url: https://github.com/harvester/charts/releases/download/harvester-cloud-provider-0.1.14/harvester-cloud-provider-0.1.14.tgz
+url: https://github.com/harvester/charts/releases/download/harvester-cloud-provider-0.2.2/harvester-cloud-provider-0.2.2.tgz
 version: 103.0.0
 doNotRelease: false
```

:::note

The Rancher version `103.0.0` needs to increase like `103.0.1` if `103.0.0` has been used, otherwise, the `index.yaml` will not be updated.

:::

(1) prepare dependency file if it comes for the first time

```
+++ b/packages/harvester/harvester-cloud-provider/generated-changes/dependencies/kube-vip/dependency.yaml
@@ -0,0 +1,3 @@
+workingDir: ""
+url: https://github.com/harvester/charts/releases/download/harvester-cloud-provider-0.2.2/harvester-cloud-provider-0.2.2.tgz
+subdirectory: dependency_charts/kube-vip
```

(2) make prepare, it will pull the source chart into path like `packages/harvester/harvester-cloud-provider`

// export PACKAGE=harvester/harvester-cloud-provider

// make prepare

```
jianwang@jianwang-pc:/go/src/github.com/w13915984028/rancher-charts$ make prepare
INFO[0000] Pulling https://github.com/harvester/charts/releases/download/harvester-cloud-provider-0.2.2/harvester-cloud-provider-0.2.2.tgz from upstream into charts 
INFO[0000] Loading dependencies for chart               
INFO[0000] found chart options for kube-vip in generated-changes/dependencies/kube-vip/dependency.yaml 
INFO[0000] Found chart options for kube-vip in generated-changes/dependencies/kube-vip/dependency.yaml 
INFO[0000] Pulling https://github.com/harvester/charts/releases/download/harvester-cloud-provider-0.2.2/harvester-cloud-provider-0.2.2.tgz[path=dependency_charts/kube-vip] from upstream into charts 
INFO[0000] Updating chart metadata with dependencies    
WARN[0000] Detected 'apiVersion:v2' within Chart.yaml; these types of charts require additional testing 
INFO[0000] Applying changes from generated-changes      
INFO[0000] Applying: generated-changes/patch/Chart.yaml.patch 
ERRO[0000] 
patching file Chart.yaml
Hunk #1 FAILED at 1.
1 out of 1 hunk FAILED -- saving rejects to file Chart.yaml.rej 
FATA[0000] encountered error while preparing main chart: encountered error while trying to apply changes to charts: unable to generate patch with error: exit status 1 
make: *** [Makefile:17: prepare] Error 1
```

::: note

if above error is shown, simply remove the file `generated-changes/patch/Chart.yaml.patch`, it means: the previous patch file has conflict, we need to generate a new one

// full path: ./packages/harvester/harvester-cloud-provider/generated-changes/patch/Chart.yaml.patch
  
:::

(3) prepare the normal & important patch file: `generated-changes/patch/Chart.yaml.patch`

// IMPORTANT, NOTE
// remove the origin `Chart.yaml.patch` first, and then re-generate it
// otherwise, it will mislead us

// in above step, when this file is removed, then after step 2, there is no such file
// the so-called `generated-changes`, is to make the diff file manually.
// e.g. Chart.yaml needs to be updated, the `catalog.cattle.io/rancher-version` field: `catalog.cattle.io/rancher-version: '>= 2.9.0-0 < 2.10.0-0'`
// cd packages/harvester/harvester-cloud-provider/charts/
// make a copy of Chart.yaml, and edit it, then diff a patch file 

jianwang@jianwang-pc:/go/src/github.com/w13915984028/rancher-charts/packages/harvester/harvester-cloud-provider/charts$ diff Chart.yaml Chart1.yaml 

```

7c7
<   catalog.cattle.io/rancher-version: '>= 2.7.0-0 < 2.8.0-0'
---
>   catalog.cattle.io/rancher-version: '>= 2.9.0-0 < 2.10.0-0'
jianwang@jianwang-pc:/go/src/github.com/w13915984028/rancher-charts/packages/harvester/harvester-cloud-provider/charts$ diff Chart.yaml Chart1.yaml  -u
--- Chart.yaml	2024-01-08 20:33:08.893433595 +0100
+++ Chart1.yaml	2024-01-08 21:07:53.378778026 +0100
@@ -4,7 +4,7 @@
   catalog.cattle.io/kube-version: '>= 1.23.0-0 < 1.27.0-0'
   catalog.cattle.io/namespace: kube-system
   catalog.cattle.io/os: linux
-  catalog.cattle.io/rancher-version: '>= 2.7.0-0 < 2.8.0-0'
+  catalog.cattle.io/rancher-version: '>= 2.9.0-0 < 2.10.0-0'
   catalog.cattle.io/release-name: harvester-cloud-provider
   catalog.cattle.io/ui-component: harvester-cloud-provider
   catalog.cattle.io/upstream-version: 0.2.0
```

touch and save to:
touch ./packages/harvester/harvester-cloud-provider/generated-changes/patch/Chart.yaml.patch

manually change it to:

```
--- charts-original/Chart.yaml
+++ charts/Chart.yaml
@@ -4,7 +4,7 @@
   catalog.cattle.io/kube-version: '>= 1.23.0-0 < 1.27.0-0'
   catalog.cattle.io/namespace: kube-system
   catalog.cattle.io/os: linux
-  catalog.cattle.io/rancher-version: '>= 2.7.0-0 < 2.8.0-0'
+  catalog.cattle.io/rancher-version: '>= 2.9.0-0 < 2.10.0-0'
   catalog.cattle.io/release-name: harvester-cloud-provider
   catalog.cattle.io/ui-component: harvester-cloud-provider
   catalog.cattle.io/upstream-version: 0.2.0
```

and save as Chart.yaml.patch

(4) prepare special patch file manually:

such repository is not allowed by Rancher chart script, it expects only one `/`, we need to patch it

charts/dependency_charts/kube-vip/values.yaml:  repository: ghcr.io/kube-vip/kube-vip
charts/charts/kube-vip/values.yaml:  repository: ghcr.io/kube-vip/kube-vip

// below 2 files are very special in this case, it patches the dependency_charts file

jianwang@jianwang-pc:/go/src/github.com/w13915984028/rancher-charts/charts/harvester-cloud-provider/103.0.0+up0.2.2/dependency_charts/kube-vip$ diff -Naur values.yaml values1.yaml > values.yaml.patch
jianwang@jianwang-pc:/go/src/github.com/w13915984028/rancher-charts/charts/harvester-cloud-provider/103.0.0+up0.2.2/dependency_charts/kube-vip$ cat values.yaml.patch 

```
--- values.yaml	2023-11-03 22:53:06.468672017 +0100
+++ values1.yaml	2023-11-17 21:14:04.573425541 +0100
@@ -3,10 +3,10 @@
 # Declare variables to be passed into your templates.
 
 image:
-  repository: ghcr.io/kube-vip/kube-vip
+  repository: rancher/mirrored-kube-vip-kube-vip-iptables
   pullPolicy: IfNotPresent
   # Overrides the image tag whose default is the chart appVersion.
-  tag: "v0.4.1"
+  tag: "v0.6.0"
 
 config:
   address: ""
```

edit the file as:

packages/harvester/harvester-cloud-provider/generated-changes/patch/dependency_charts/kube-vip/values.yaml.patch

+++ b/packages/harvester/harvester-cloud-provider/generated-changes/patch/dependency_charts/kube-vip/values.yaml.patch
```
@@ -0,0 +1,15 @@
+--- charts-original/dependency_charts/kube-vip/values.yaml
++++ charts/dependency_charts/kube-vip/values.yaml
+@@ -3,10 +3,10 @@
+ # Declare variables to be passed into your templates.
+ 
+ image:
+-  repository: ghcr.io/kube-vip/kube-vip
++  repository: rancher/mirrored-kube-vip-kube-vip-iptables
+   pullPolicy: IfNotPresent
+   # Overrides the image tag whose default is the chart appVersion.
+-  tag: "v0.4.1"
++  tag: "v0.6.0"
+ 
+ config:
+   address: ""
```

// remember to change the file path

dependency_charts/kube-vip/values.yaml
and
charts/kube-vip/values.yaml

```
--- charts-original/dependency_charts/kube-vip/values.yaml
+++ charts/dependency_charts/kube-vip/values.yaml
@@ -3,10 +3,10 @@
 # Declare variables to be passed into your templates.
 
 image:
-  repository: ghcr.io/kube-vip/kube-vip
+  repository: rancher/mirrored-kube-vip-kube-vip-iptables
   pullPolicy: IfNotPresent
   # Overrides the image tag whose default is the chart appVersion.
-  tag: "v0.4.1"
+  tag: "v0.6.0"
 
 config:
   address: ""
```


```
--- charts-original/charts/kube-vip/values.yaml
+++ charts/charts/kube-vip/values.yaml
@@ -3,10 +3,10 @@
 # Declare variables to be passed into your templates.
 
 image:
-  repository: ghcr.io/kube-vip/kube-vip
+  repository: rancher/mirrored-kube-vip-kube-vip-iptables
   pullPolicy: IfNotPresent
   # Overrides the image tag whose default is the chart appVersion.
-  tag: "v0.4.1"
+  tag: "v0.6.0"
 
 config:
   address: ""
```

and normally, another file:

--- a/packages/harvester/harvester-cloud-provider/generated-changes/patch/Chart.yaml.patch


for dev-v2.8, expect catalog.cattle.io/rancher-version: '>= 2.8.0-0 < 2.9.0-0'


(2) prepare dependency file

```
+++ b/packages/harvester/harvester-cloud-provider/generated-changes/dependencies/kube-vip/dependency.yaml
@@ -0,0 +1,3 @@
+workingDir: ""
+url: https://github.com/harvester/charts/releases/download/harvester-cloud-provider-0.2.2/harvester-cloud-provider-0.2.2.tgz
+subdirectory: dependency_charts/kube-vip
```

///// make again

```
// export PACKAGE=harvester/harvester-cloud-provider


// make prepare

/// this time, those patches will be applied to the chart.
/// try grep "mirrored-kube-vip-kube-vip-iptables"
jianwang@jianwang-pc:/go/src/github.com/w13915984028/rancher-charts$ gy "mirrored-kube-vip-kube-vip-iptables"
packages/harvester/harvester-cloud-provider/charts/dependency_charts/kube-vip/values.yaml:  repository: rancher/mirrored-kube-vip-kube-vip-iptables
packages/harvester/harvester-cloud-provider/charts/values.yaml:    repository: rancher/mirrored-kube-vip-kube-vip-iptables
packages/harvester/harvester-cloud-provider/charts/charts/kube-vip/values.yaml:  repository: rancher/mirrored-kube-vip-kube-vip-iptables
```

### Make

The whole command list is:

```
$ export PACKAGE=harvester/harvester-cloud-provider
$ make prepare
$ make patch
$ make clean
$ make charts
```


### Validate and check images

```
make validate
make check-images
```



///// make validate

such error will happen:
```
INFO[0178] harvester-cloud-provider/103.0.0+up0.2.2 is untracked 
ERRO[0181] The following new assets have been introduced: map[harvester-cloud-provider:[103.0.0+up0.2.2]] 
ERRO[0181] The following released assets have been removed: map[] 
ERRO[0181] The following released assets have been modified: map[] 
ERRO[0181] If this was intentional, to allow validation to pass, these charts must be added to the release.yaml. 
INFO[0181] Dumping release.yaml tracking changes that have been introduced 
```

// update the release.yaml, commit ,then validate again

```
...
INFO[0188] index.yaml is up-to-date                     
INFO[0188] Doing a final check to ensure Git is clean   
INFO[0190] Successfully validated current repository!   




commit 412b507b830bf1ece1d51f8f081bbc3a2d46ff32 (HEAD -> bumphcp1)
Author: Jian Wang <w13915984028@gmail.com>
Date:   Fri Nov 17 23:00:15 2023 +0100

    Add harvester-cloud-provider 103.0.0+up0.2.2 to release
    
    Signed-off-by: Jian Wang <w13915984028@gmail.com>
```

jianwang@jianwang-pc:/go/src/github.com/w13915984028/rancher-charts$ gy "ghcr.io/kube-vip/kube-vip"
charts/harvester-cloud-provider/103.0.0+up0.2.2/charts/kube-vip/values.yaml:  repository: ghcr.io/kube-vip/kube-vip

### last step

```
make check-images
```

errors like:

```
time="2023-11-03T21:56:40Z" level=fatal msg="failed to generate namespace and repository for image: ghcr.io/kube-vip/kube-vip"
make: *** [Makefile:17: check-images] Error 1

due to the make charts does not allow image format like `ghcr.io/kube-vip/kube-vip`, only one `/` can exist
```

(5) add to rke2 charts as well

https://github.com/rancher/rke2-charts/blob/main-source/packages/harvester-cloud-provider/package.yaml

(6) set new chart as the default value

https://github.com/rancher/rke2/blob/master/charts/chart_versions.yaml

https://github.com/rancher/rke2/pull/5980










