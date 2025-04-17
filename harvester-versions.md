# Harvester Versions

## v1.5.0


## v1.4.2

```
Component	Version
Longhorn	v1.7.2
KubeVirt	v1.3.1
Embedded Rancher	v2.10.1
RKE2	v1.31.4+rke2r1
SLE Micro for Rancher	5.5


runtimeversion: v1.31.4+rke2r1
rancherversion: v2.10.1
harvesterchartversion: 1.4.2
monitoringchartversion: 103.1.1+up45.31.1
loggingchartversion: 103.1.0+up4.4.0

longhornio/longhorn-manager:v1.7.2

  targetKubeVirtRegistry: registry.suse.com/suse/sles/15.6
  targetKubeVirtVersion: 1.3.1-150600.5.9.1
  registry.suse.com/suse/sles/15.6/virt-operator:1.3.1-150600.5.9.1


$ helm history -n cattle-system rancher
REVISION	UPDATED                 	STATUS  	CHART                              	APP VERSION  	DESCRIPTION     
1       	Wed Mar 19 09:09:01 2025	deployed	rancher-2.10.1-dirty.commit-47da90f	47da90f-dirty	Install complete


$ kubectl get deployment -n cattle-system rancher -oyaml | grep image
        image: rancher/rancher:v2.10.1
        imagePullPolicy: IfNotPresent
```



## v1.4.1

```
Component	Version
Longhorn	v1.7.2
KubeVirt	v1.2.2
Embedded Rancher	v2.9.3
RKE2	v1.30.7+rke2r1
SLE Micro for Rancher	5.5
```

## v1.4.0

```
Component	Version
Longhorn	v1.7.2
KubeVirt	v1.2.2
Embedded Rancher	v2.9.2
RKE2	v1.29.9+rke2r1
SLE Micro for Rancher	5.5
```

## v1.3.2

```
Longhorn	v1.6.2
KubeVirt	v1.1.1
Embedded Rancher	v2.8.5
RKE2	v1.28.12+rke2r1
SLE Micro for Rancher	5.4
```

## v1.3.1

```
Component	Version
Longhorn	v1.6.2
KubeVirt	v1.1.1
Embedded Rancher	v2.8.3
RKE2	v1.27.13+rke2r1
SLE Micro for Rancher	5.4
```

## v1.3.0

```
Component	Version
Longhorn	v1.6.0
KubeVirt	v1.1.0
Embedded Rancher	v2.8.2
RKE2	v1.27.10+rke2r1


RKE2_VERSION="v1.27.10+rke2r1"
RANCHER_VERSION="v2.8.2"
MONITORING_VERSION=103.0.3+up45.31.1
LOGGING_VERSION=103.0.0+up3.17.10

kubevirt: v1.1.0
https://github.com/harvester/harvester/blob/release-1.3/deploy/charts/harvester/values.yaml
        repository: registry.suse.com/suse/sles/15.5/virt-operator
        tag: &kubevirtVersion 1.1.0-150500.8.6.1
```

## v1.2.2

```
Component	Version
Longhorn	v1.5.5
KubeVirt	v1.1.1
Embedded Rancher	v2.8.2
RKE2	v1.26.15+rke2r1
SLE Micro for Rancher	5.4


https://github.com/harvester/harvester-installer/blob/v1.2/scripts/version-rke2

RKE2_VERSION="v1.26.13+rke2r1"

RANCHER_VERSION="v2.8.2"
MONITORING_VERSION=102.0.0+up40.1.2
LOGGING_VERSION=102.0.0+up3.17.10

kubevirt: 1.1.0
        repository: registry.suse.com/suse/sles/15.5/virt-operator
        tag: &kubevirtVersion 1.1.0-150500.8.6.1
https://github.com/harvester/harvester/blob/v1.2/deploy/charts/harvester/values.yaml


runtimeversion: v1.26.13+rke2r1
rancherversion: v2.8.2
```


## v1.2.1

```

Component	Version
Longhorn	v1.4.3
KubeVirt	v0.54.0
Embedded Rancher	v2.7.5
RKE2	v1.25.9+rke2r1


runtimeversion: v1.25.9+rke2r1
rancherversion: v2.7.5
harvesterchartversion: 1.2.1
monitoringchartversion: 102.0.0+up40.1.2
systemsettings:
  ntp-servers: '{"ntpServers":["0.suse.pool.ntp.org"]}'
loggingchartversion: 102.0.0+up3.17.10

kubevirt: 0.54.0

https://github.com/harvester/harvester/blob/release-1.2/deploy/charts/harvester/values.yaml
        repository: registry.suse.com/suse/sles/15.4/virt-operator
        tag: &kubevirtVersion 0.54.0-150400.3.19.1
```

## v1.2.0

```
Component	Version
Longhorn	v1.4.3
KubeVirt	v0.54.0
Embedded Rancher	v2.7.5
RKE2	v1.25.9+rke2r1
```
