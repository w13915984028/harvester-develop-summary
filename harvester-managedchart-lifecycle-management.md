# Harvester Managedchart Lifecycle Management

## Background

Harvester itself is managed via a `managedchart` CRD object, many of the related resources are not exposed to Harvester WebUI, to change it, the `kubectl` is the main tool.

And, for most data, the `managedchart` is the source of truth.

When a downstream object like `harvester deployment` is changed directly, the `managedchart` will detect it and show the error/warning messages.

Harvester upgrade script will check if the `managedchart` is ready and no error, if not, the upgrade is blocked. In-between the upgrade process, the `managedchart` is also checked to wait it is ready.

## Initialization

1. ManagedChart Initialization

Via harvester-installer, with the predefined values and those values from user input/configuration file.

https://github.com/harvester/harvester-installer/blob/6fa1df88616727e74b789a22440120787dc86732/pkg/config/templates/rancherd-10-harvester.yaml#L70

```
  kind: ManagedChart
  metadata:
    name: harvester
    namespace: fleet-local
  spec:
    chart: harvester
    releaseName: harvester
    version: {{ .HarvesterChartVersion }}
    defaultNamespace: harvester-system
    timeoutSeconds: 600
    diff:
...
    values:

...
      harvester-network-controller:
        enabled: true                          // predefined
        vipEnabled: true
        image:
          pullPolicy: "IfNotPresent"

...
      promote:
        clusterPodCIDR: {{ or .ClusterPodCIDR "10.52.0.0/16" }}  // from user input or default value
        clusterServiceCIDR: {{ or .ClusterServiceCIDR "10.53.0.0/16" }}
        clusterDNS: {{ or .ClusterDNS "10.53.0.10" }}
```

2. Real Chart

From Harvester chart and those dependency charts/sub charts.

https://github.com/harvester/harvester/blob/41d3e1b5b1b5d3e9d3a91568638ab8c5b4071f8b/deploy/charts/harvester/values.yaml#L352

```
...

storageClass:
  ## Specify the default Storage Class of harvester-longhorn.
  ## Will be set "false" when upgrading for existing default Storage Class.
  ## defaults to "true".
  defaultStorageClass: true
  reclaimPolicy: Delete
  replicaCount: 3

harvester-network-controller:
  ## Specify to install harvester network controller,
  ## defaults to "false".
  enabled: true
  image:
    repository: rancher/harvester-network-controller
    pullPolicy: IfNotPresent
    # Overrides the image tag whose default is the chart appVersion.
    tag: master-head
  helper:
    image:
      repository: rancher/harvester-network-helper
      tag: master-head
  webhook:
    image:
      repository: rancher/harvester-network-webhook
      pullPolicy: IfNotPresent
      tag: master-head
...      
```

https://github.com/harvester/harvester/blob/41d3e1b5b1b5d3e9d3a91568638ab8c5b4071f8b/deploy/charts/harvester/templates/harvester-storageclass.yaml#L1C1-L10C11


```
{{- if .Values.longhorn.enabled -}}
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: harvester-longhorn
  annotations:
    harvesterhci.io/is-reserved-storageclass: "true"
{{- if .Values.storageClass.defaultStorageClass }}
    storageclass.kubernetes.io/is-default-class: "true"
{{- end }}

```

3. Dependency charts/sub charts

https://github.com/harvester/charts/blob/9d509044efcb8085a3dec5851628c106b3832338/charts/harvester-network-controller/values.yaml#L4

```
...
image:
  repository: rancher/harvester-network-controller
  pullPolicy: Always
  # Overrides the image tag whose default is the chart appVersion.
  tag: "master-head"

nameOverride: ""

# Specify whether to enable VIP, defaults to false
vipEnabled: false

resources:
  limits:
    cpu: 100m
    memory: 128Mi
  requests:
    cpu: 10m
    memory: 64Mi
```

The data from managedchart is propgated to real chart and it's dependency/sub charts.

### Data Category

#### Continuous Updated Data

e.g. deployment spec, images

#### One-time Initialized Data

e.g. The `VIP` configmap

https://github.com/harvester/harvester/blob/41d3e1b5b1b5d3e9d3a91568638ab8c5b4071f8b/deploy/charts/harvester/templates/configmap.yaml#L2

#### Read-only Data

The ClusterPodCIDR and related.

```
      promote:
        clusterPodCIDR: {{ or .ClusterPodCIDR "10.52.0.0/16" }}  // from user input or default value
        clusterServiceCIDR: {{ or .ClusterServiceCIDR "10.53.0.0/16" }}
        clusterDNS: {{ or .ClusterDNS "10.53.0.10" }}
```

Refer PR: https://github.com/harvester/harvester/pull/7445

#### Feature Gate Data

Those `enabled`/`disabled` data.

#### Special Data

The annotation `storageclass.kubernetes.io/is-default-class: "true"` on downstream `harvester-longhorn` SC object is further changed when user sets another SC as default; but the managedchart is not ammended.

https://github.com/harvester/harvester/blob/41d3e1b5b1b5d3e9d3a91568638ab8c5b4071f8b/deploy/charts/harvester/templates/harvester-storageclass.yaml#L9


```
  annotations:
{{- if .Values.storageClass.defaultStorageClass }}
    storageclass.kubernetes.io/is-default-class: "true"
{{- end }}
```

## Customization on the ManagedChart

1. Change the Image tag

E.g. for debug purpose, or the current image has an known blocker bug and a new image is released to fix the bug, user needs to change below path of data on the managedchart.

```
.spec.harvester-network-controller.image

```

2. Change the resources

E.g. the default resources are low and OOM happens, set a higher resources

```
.spec.harvester-network-controller.resources

```

## Upgrade

The main process it to keep the current spec and upgrade the managedchart to new version.

Any direct change on downstream objects will be reverted by the `managedchart spec` plus `chart default values`, besides the special processing on upgrade script.

https://github.com/harvester/harvester/blob/41d3e1b5b1b5d3e9d3a91568638ab8c5b4071f8b/package/upgrade/upgrade_manifests.sh#L862

```
...
  cat >harvester.yaml <<EOF
apiVersion: management.cattle.io/v3
kind: ManagedChart
metadata:
  name: harvester
  namespace: fleet-local
EOF
  kubectl get managedcharts.management.cattle.io -n fleet-local harvester -o yaml | yq e '{"spec": .spec}' - >>harvester.yaml

  upgrade_managed_chart_from_version $UPGRADE_PREVIOUS_VERSION harvester harvester.yaml
..
  kubectl apply -f ./harvester.yaml  
```

### Potential Challenges

#### 1. One-time Initialized Data is Reinitialized

https://github.com/harvester/harvester/blob/41d3e1b5b1b5d3e9d3a91568638ab8c5b4071f8b/deploy/charts/harvester/dependency_charts/snapshot-validation-webhook/templates/webhook.yaml#L4

```
{{- $certificate := "" }}
{{- $key := "" }}
{{- if .Values.webhook.tls.autoGenerated }}
  {{- $serviceName := (printf "%s.%s.svc" (include "snapshot-validation-webhook.fullname" .) (include "snapshot-validation-webhook.namespace" .))  }}
  {{- $cert := genSelfSignedCert $serviceName nil (list $serviceName) 3650 }}
  {{- $certificate = b64enc $cert.Cert }}
  {{- $key = b64enc $cert.Key }}

```

#### 2. Customized Values vs Chart Default Values

e.g. The image was customized before, but when upgrading, a new image is more preferred. Now user needs to change manually; or upgrade script updates it.

```
.spec.harvester-network-controller.image

```

e.g. The resources setting, ideally, the bigger one should be kept. Now the `.spec...resources` is kept.

```
.spec.harvester-network-controller.resources

```

#### 3. Special Data

e.g. the `"storageclass.kubernetes.io/is-default-class"` annotation on `harvester-longhorn`.

It is dedicated processed on the upgrade script.

https://github.com/harvester/harvester/blob/41d3e1b5b1b5d3e9d3a91568638ab8c5b4071f8b/package/upgrade/upgrade_manifests.sh#L893C1-L896C5

```
  local sc=$(kubectl get sc -o json | jq '.items[] | select(.metadata.annotations."storageclass.kubernetes.io/is-default-class" == "true" and .metadata.name != "harvester-longhorn")')
  if [ -n "$sc" ] && [ "$UPGRADE_PREVIOUS_VERSION" != "v1.0.3" ]; then
      yq e '.spec.values.storageClass.defaultStorageClass = false' -i harvester.yaml
  fi
```

Why?

Harvester allows user to set any `storageclass` as default on WebUI, it removes the annotation from old sc and adds it to new sc. But the `managedchart harvester` is not updated. The processing is added to upgrade script.

### Idea Solution

Each data can have a policy to `keep current` / `prefer managedchart value` / `prefer chart value` / `read-only`, but apparently it is way more complex.
