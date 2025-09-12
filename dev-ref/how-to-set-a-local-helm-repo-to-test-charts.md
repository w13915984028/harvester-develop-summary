

## Update chart

Get the chart, and per https://github.com/rancher/charts/blob/dev-v2.12/charts/harvester-cloud-provider/107.0.1%2Bup0.2.10/Chart.yaml to change local to let it meet Rancher's format.

```
rancher@email:/go/src/github.com/w13915984028/harvester-charts/charts$ git diff HEAD~1 
diff --git a/charts/harvester-cloud-provider/Chart.yaml b/charts/harvester-cloud-provider/Chart.yaml
index c286fef..a2f3e2b 100644
--- a/charts/harvester-cloud-provider/Chart.yaml
+++ b/charts/harvester-cloud-provider/Chart.yaml
@@ -1,48 +1,28 @@
-apiVersion: v2
-name: harvester-cloud-provider
-description: A Helm chart for Harvester Cloud Provider
-
-# A chart can be either an 'application' or a 'library' chart.
-#
-# Application charts are a collection of templates that can be packaged into versioned archives
-# to be deployed.
-#
-# Library charts provide useful utilities or functions for the chart developer. They're included as
-# a dependency of application charts to inject those utilities and functions into the rendering
-# pipeline. Library charts do not define any templates and therefore cannot be deployed.
-type: application
-keywords:
-  - infrastructure
-  - harvester
-
-# This is the chart version. This version number should be incremented each time you make changes
-# to the chart and its templates, including the app version.
-# Versions are expected to follow Semantic Versioning (https://semver.org/)
-version: 0.0.0-dev
-
-# This is the version number of the application being deployed. This version number should be
-# incremented each time you make changes to the application. Versions are not expected to
-# follow Semantic Versioning. They should reflect the version the application is using.
-appVersion: v0.2.3
-
 annotations:
   catalog.cattle.io/certified: rancher
-  catalog.cattle.io/namespace: kube-system
+  catalog.cattle.io/display-name: Harvester Cloud Provider
   catalog.cattle.io/kube-version: '>= 1.23.0-0 < 1.34.0-0'
-  catalog.cattle.io/release-name: harvester-cloud-provider
+  catalog.cattle.io/namespace: kube-system
   catalog.cattle.io/os: linux
-  catalog.cattle.io/rancher-version: '>= 2.7.0-0 < 2.13.0-0'
+  catalog.cattle.io/permits-os: linux
+  catalog.cattle.io/rancher-version: '>= 2.12.0-0 < 2.13.0-0'
+  catalog.cattle.io/release-name: harvester-cloud-provider
   catalog.cattle.io/ui-component: harvester-cloud-provider
-  catalog.cattle.io/display-name: Harvester Cloud Provider
-  # The version of the upstream chart or app. It prevents the unexpected "downgrade"
-  # when upgrading an installed chart that uses our 100.x.x+upVersion version schema.```
-  catalog.cattle.io/upstream-version: "0.2.0"
-
-maintainers:
-  - name: harvester
-
+  catalog.cattle.io/upstream-version: 0.2.11
+apiVersion: v2
+appVersion: v0.2.4
 dependencies:
-  - name: kube-vip
-    condition: kube-vip.enabled
-    version: 0.8.0
-    repository: file://dependency_charts/kube-vip
+- condition: kube-vip.enabled
+  name: kube-vip
+  repository: file://dependency_charts/kube-vip
+  version: 0.8.0
+description: A Helm chart for Harvester Cloud Provider
+icon: file://assets/logos/harvester.svg
+keywords:
+- infrastructure
+- harvester
+maintainers:
+- name: harvester
+name: harvester-cloud-provider
+type: application
+version: 107.0.2+up0.2.11

```


## Package local chart


rancher@email:/go/src/github.com/w13915984028/harvester-charts/charts$ helm package harvester-cloud-provider/

Successfully packaged chart and saved it to: /go/src/github.com/w13915984028/harvester-charts/charts/harvester-cloud-provider-107.0.2+up0.2.11.tgz

rancher@email:/go/src/github.com/w13915984028/harvester-charts/charts$ ls

harvester-cloud-provider-107.0.2+up0.2.11.tgz 


## Create repo index.yaml

per https://helm.sh/docs/topics/chart_repository/

e.g., put all related charts under /home/rancher/chart-repo

run command `helm repo index charts/` to generate `index.yaml`


```
rancher@email:~/chart-repo$ helm repo index charts/

rancher@email:~/chart-repo$ ls charts/
harvester-cloud-provider-107.0.2+up0.2.11.tgz  index.yaml


rancher@email:~/chart-repo$ cat charts/index.yaml 
apiVersion: v1
entries:
  harvester-cloud-provider:
  - annotations:
      catalog.cattle.io/certified: rancher
      catalog.cattle.io/display-name: Harvester Cloud Provider
      catalog.cattle.io/kube-version: '>= 1.23.0-0 < 1.34.0-0'
      catalog.cattle.io/namespace: kube-system
      catalog.cattle.io/os: linux
      catalog.cattle.io/permits-os: linux
      catalog.cattle.io/rancher-version: '>= 2.12.0-0 < 2.13.0-0'
      catalog.cattle.io/release-name: harvester-cloud-provider
      catalog.cattle.io/ui-component: harvester-cloud-provider
      catalog.cattle.io/upstream-version: 0.2.11
    apiVersion: v2
    appVersion: v0.2.4
    created: "2025-08-29T16:14:48.171568561Z"
    dependencies:
    - condition: kube-vip.enabled
      name: kube-vip
      repository: file://dependency_charts/kube-vip
      version: 0.8.0
    description: A Helm chart for Harvester Cloud Provider
    digest: 1d09f7b2985c1e3f62196bb5ac16a87aa3a827ceae3db1305f7780b90dfc205d
    icon: file://assets/logos/harvester.svg
    keywords:
    - infrastructure
    - harvester
    maintainers:
    - name: harvester
    name: harvester-cloud-provider
    type: application
    urls:
    - harvester-cloud-provider-107.0.2+up0.2.11.tgz
    version: 107.0.2+up0.2.11
generated: "2025-08-29T16:14:48.169981119Z"
```

## Serve the repo

```
rancher@email:~/chart-repo/charts$ python3 -m http.server
Serving HTTP on 0.0.0.0 port 8000 (http://0.0.0.0:8000/) ...
```

log by remote rancher

```
rancher@email:~/chart-repo/charts$ python3 -m http.server
Serving HTTP on 0.0.0.0 port 8000 (http://0.0.0.0:8000/) ...
192.168.122.163 - - [29/Aug/2025 16:18:26] "GET /index.yaml HTTP/1.1" 200 -
192.168.122.163 - - [29/Aug/2025 16:19:44] "GET /index.yaml HTTP/1.1" 200 -
192.168.122.163 - - [29/Aug/2025 16:20:28] "GET /harvester-cloud-provider-107.0.2+up0.2.11.tgz HTTP/1.1" 200 -
192.168.122.163 - - [29/Aug/2025 16:21:05] "GET /harvester-cloud-provider-107.0.2+up0.2.11.tgz HTTP/1.1" 200 -
```

## Add repository on Rancher manager

From a path like below: `guest cluster -> Apps -> Repositories`

e.g. local repo `http://192.168.122.191:8000`


Add your own repo and select `Refresh`

https://192.168.122.191/dashboard/c/c-m-rtppw764/apps/catalog.cattle.io.clusterrepo


server log:
```
192.168.122.187 - - [02/Sep/2025 19:42:10] "GET /harvester-cloud-provider-107.0.2+up0.2.11.tgz HTTP/1.1" 200 -

```

## Find the chart to upgrade

Go to `Apps`, search `Harvester-cloud` it will show a new version from your local chart repo

## Upgrade log


```
helm upgrade --history-max=5 --install=true --labels=catalog.cattle.io/cluster-repo-name=wj-local --namespace=kube-system --timeout=10m0s --values=/home/shell/helm/values-harvester-cloud-provider-107.0.2-up0.2.11.yaml --version=107.0.2+up0.2.11 --wait=true harvester-cloud-provider /home/shell/helm/harvester-cloud-provider-107.0.2-up0.2.11.tgz
Release "harvester-cloud-provider" has been upgraded. Happy Helming!
NAME: harvester-cloud-provider
LAST DEPLOYED: Fri Aug 29 16:21:09 2025
NAMESPACE: kube-system
STATUS: deployed
REVISION: 2
TEST SUITE: None
---------------------------------------------------------------------
SUCCESS: helm upgrade --history-max=5 --install=true --labels=catalog.cattle.io/cluster-repo-name=wj-local --namespace=kube-system --timeout=10m0s --values=/home/shell/helm/values-harvester-cloud-provider-107.0.2-up0.2.11.yaml --version=107.0.2+up0.2.11 --wait=true harvester-cloud-provider /home/shell/helm/harvester-cloud-provider-107.0.2-up0.2.11.tgz
---------------------------------------------------------------------
```

## Kube-vip log

```
> kubectl logs -n kube-system kube-vip-xfrmn
2025/08/29 16:21:17 INFO kube-vip.io version=v0.9.2 build=b56b80cd30497207e340d581a0db34469aa41c57
2025/08/29 16:21:17 INFO starting namespace=kube-system Mode=ARP "Control Plane"=false Services=true
2025/08/29 16:21:17 INFO No interface is specified for VIP in config, auto-detecting default Interface
2025/08/29 16:21:17 INFO prometheus HTTP server started
2025/08/29 16:21:17 INFO kube-vip bind interface=enp1s0
2025/08/29 16:21:17 WARN Node name is missing from the config, fall back to hostname
2025/08/29 16:21:17 INFO using node name name=gc22-pool1-xqg4w-bgpwp
2025/08/29 16:21:17 INFO Starting Kube-vip Manager with the ARP engine
2025/08/29 16:21:17 INFO Start ARP/NDP advertisement
2025/08/29 16:21:17 INFO beginning services leadership namespace=kube-system "lock name"=plndr-svcs-lock id=gc22-pool1-xqg4w-bgpwp
I0829 16:21:17.429047       1 leaderelection.go:257] attempting to acquire leader lease kube-system/plndr-svcs-lock...
2025/08/29 16:21:17 INFO Starting UPNP Port Refresher
2025/08/29 16:21:17 INFO Starting ARP/NDP advertisement
I0829 16:21:17.443357       1 leaderelection.go:271] successfully acquired lease kube-system/plndr-svcs-lock
2025/08/29 16:21:17 INFO (svcs) starting services watcher for all namespaces
2025/08/29 16:21:17 INFO Using existing macvlan interface for DHCP interface=vip-111b3536
2025/08/29 16:21:17 INFO (svcs) adding VIP ip=192.168.122.71 interface=vip-111b3536 namespace=default name=gc22-lb1
2025/08/29 16:21:17 INFO layer 2 broadcaster starting
2025/08/29 16:21:17 INFO inserting ARP/NDP instance name=192.168.122.71/32-vip-111b3536
2025/08/29 16:21:17 INFO [service] service=gc22-lb1 namespace=default "synchronised in"=76ms
2025/08/29 16:26:17 INFO [UPNP] Refreshing Instances "number of instances"=1
2025/08/29 16:31:17 INFO [UPNP] Refreshing Instances "number of instances"=1
2025/08/29 16:36:17 INFO [UPNP] Refreshing Instances "number of instances"=1
2025/08/29 16:41:17 INFO [UPNP] Refreshing Instances "number of instances"=1
2025/08/29 16:46:17 INFO [UPNP] Refreshing Instances "number of instances"=1
2025/08/29 16:51:17 ERROR renew failed err="create renew client failed, error: unable to apply option: unable to start listening UDP port: listen udp4 192.168.122.71:68: bind: permission denied, server ip: 192.168.122.1"
2025/08/29 16:51:17 INFO [UPNP] Refreshing Instances "number of instances"=1
2025/08/29 16:56:17 INFO [UPNP] Refreshing Instances "number of instances"=1
2025/08/29 17:01:17 INFO [UPNP] Refreshing Instances "number of instances"=1

```

## Deploy nginx to test

`kubectl apply -f https://k8s.io/examples/application/deployment.yaml`

label key:
```
      labels:
        app: nginx
```



## push tag to ttl.sh

build pc:

re-tag and push  image

```
 docker image tag rancher/harvester-cloud-provider:59e0f919-amd64 ttl.sh/harvester-cloud-provider:1h
 docker push ttl.sh/harvester-cloud-provider:1h 
The push refers to repository [ttl.sh/harvester-cloud-provider]
5f762d41f1f2: Pushed 
0e924c40cc58: Pushed 
509ff0bdf839: Pushed 
1h: digest: sha256:98dcfcd30f264aeabba7d2dd286559c7716890933c2e46bc6684f1ed3f3d4a6a size: 947
```


## pull image tag from ttl.sh



```

harv41:/home/rancher # docker pull ttl.sh/harvester-cloud-provider:1h
ttl.sh/harvester-cloud-provider:1h:                                               resolved       |++++++++++++++++++++++++++++++++++++++| 
manifest-sha256:98dcfcd30f264aeabba7d2dd286559c7716890933c2e46bc6684f1ed3f3d4a6a: done           |++++++++++++++++++++++++++++++++++++++| 
config-sha256:b77b706830927670847b650e08dcbef31b6a8bd26f8a96d3ef69f2fb3e290f16:   done           |++++++++++++++++++++++++++++++++++++++| 
layer-sha256:f854a1142d6853ab315ae729614e2d68df761c16cef5bcb9881fe94da5c77c03:    done           |++++++++++++++++++++++++++++++++++++++| 
layer-sha256:fd9bec9c996d224b2bd3930ec0c5f34385f53dd3ba586488dcf5135f9dbee592:    done           |++++++++++++++++++++++++++++++++++++++| 
layer-sha256:886ca302f1914eed4552cad0a0c0229150037cde48536ad3042a262958624f01:    done           |++++++++++++++++++++++++++++++++++++++| 
elapsed: 8.3 s                                                                    total:  39.5 M (4.8 MiB/s)                                       


ttl.sh/harvester-cloud-provider                                    1h                                                     98dcfcd30f26    About a minute ago    linux/amd64    136.0 MiB    39.5 MiB
```


## test new image

kubectl edit deployment -n kube-system harvester-cloud-provider

        image: ttl.sh/harvester-cloud-provider:1h
        imagePullPolicy: IfNotPresent
