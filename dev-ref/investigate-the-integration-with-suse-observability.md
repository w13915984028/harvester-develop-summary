
# Investigate the integration with SUSE-Observability

Harvester can be integrated with SUSE-Observability (`SO`). And we need a SUSE-Observability instance.

This doc talks about how to setup a SUSE-Observability instance for development and test.

## Install the SUSE-Observability onto a Harvester cluster

General doc: https://github.com/StackVista/stackstate-docs/blob/suse-observability/setup/install-stackstate/kubernetes_openshift/kubernetes_install.md

SUSE-Observability can be installed onto a K8s cluster in most cases, and Harvester cluster satifies this.

Following steps are tested on a Harvester (v1.7.0) cluster.

### Prepare helm repo

```sh
helm repo add suse-observability https://charts.rancher.com/server-charts/prime/suse-observability

helm repo update
```

### Prepare the SO installation template

In production, the SO needs strong resources. For test/PoC, a small footprint also works. This is done via param `sizing.profile='trial'`.

#### Prepare a second VIP for SO to use

The SO relies on `baseUrl`. In Harvester, the ingress is managed via `rancher-expose`. To make life easier, we can prepare a second vip to serve SO dedicately. The VIP can be either static or DHCP. If in DHCP mode, [create the service](#create-a-new-lb-type-service) first to get an IP for template usage. The static mode is recommended.

Assume a second VIP `192.168.168.233` is prepared for SO to use.

#### Generate the template

```sh

generate template:
  
export VALUES_DIR=.
helm template \
  --set license='V2...' \
  --set baseUrl='http:///192.168.122.233:8090' \
  --set sizing.profile='trial' \
  suse-observability-values \
  suse-observability/suse-observability-values --output-dir $VALUES_DIR  
```

:::note

1. The `profile='trial'` guides the template to use as few resources as possible.

2. Manually edit the generated files under path `./suse-observability-values`, to reduce the `request` resources like `cpu, memory`, and also adjust the PVC size. The `trial` still requests a big amount of resources, if your Harvester cluster is installed with the entry level resources.

3. If your `baseUrl` is changed, either re-generate the template, or edit the generated files directly. 

4. The `baseUrl` is critical, if the external accessing is not from this URL, SO will return error like `503`. The UI asks for `License` and `username/password`, and then stucks loading, if your access link is not same as `baseUrl`.

5. The `admin` password is also saved on the generated file.

6. For the Licese, please contact the author.

:::

### Install SO

```sh

install:

export VALUES_DIR=.

helm upgrade \
   --install \
   --namespace suse-observability \
   --values $VALUES_DIR/suse-observability-values/templates/baseConfig_values.yaml \
   --values $VALUES_DIR/suse-observability-values/templates/sizing_values.yaml \
 suse-observability \
 suse-observability/suse-observability

```


```sh
list instllation

helm list --namespace suse-observability


NAME              	NAMESPACE         	REVISION	UPDATED                                	STATUS  	CHART                   	APP VERSION
suse-observability	suse-observability	1       	2026-01-13 22:10:53.470427079 +0000 UTC	deployed	suse-observability-2.6.3	7.0.0-snapshot.20251120152824-master-8cfee78
```

#### Adjust privilege

Before the change, the default `suse-observability-elasticsearch-master` statefulset will fail, and the `...receiver` POD is not ready. The SO UI is ready, but we can't register a second Harvester cluster onto OS, as the `receiver` is abnormal.

```sh
harv41:/home/rancher # kso
NAME                                                              READY   STATUS     RESTARTS       AGE
suse-observability-clickhouse-shard0-0                            2/2     Running    0              23m
suse-observability-correlate-7bdb8b476-4klcn                      1/1     Running    3 (19m ago)    23m
suse-observability-e2es-9b96d66db-56bv6                           0/1     Init:0/1   3 (5m9s ago)   23m
suse-observability-hbase-stackgraph-0                             1/1     Running    2 (23m ago)    23m
suse-observability-hbase-tephra-mono-0                            1/1     Running    0              23m
suse-observability-kafka-0                                        2/2     Running    2 (23m ago)    23m
suse-observability-kafkaup-operator-kafkaup-cb4c68b68-wth7t       1/1     Running    0              23m
suse-observability-otel-collector-0                               1/1     Running    0              23m
suse-observability-prometheus-elasticsearch-exporter-b8fc8rws45   1/1     Running    0              23m
suse-observability-rbac-agent-6d7d974f5c-gtptn                    1/1     Running    0              23m
suse-observability-receiver-85556f44c9-qlxmp                      0/1     Init:0/1   3 (5m9s ago)   23m  // stucking
suse-observability-router-79f768cb44-z89lp                        1/1     Running    0              23m
suse-observability-server-6c564c47cc-4rgkv                        1/1     Running    1 (20m ago)    23m
suse-observability-ui-7d4558bc8c-6lp64                            2/2     Running    0              23m
suse-observability-victoria-metrics-0-0                           1/1     Running    0              23m
suse-observability-vmagent-0                                      1/1     Running    0              23m
suse-observability-zookeeper-0                                    1/1     Running    0              23m


$ cat /var/log/pods/suse-observability_suse-observability-receiver-85556f44c9-t7g8f_55af6722-2646-4706-96d0-cf805c325647/receiver-init/6.log 
2026-01-15T20:43:19.46161546Z stdout F Waiting for suse-observability-kafka-headless:9092...
2026-01-15T20:43:19.464126932Z stderr F Connection to suse-observability-kafka-headless (10.52.0.67) 9092 port [tcp/*] succeeded!
2026-01-15T20:43:19.464266512Z stdout F -> suse-observability-kafka-headless:9092 is up!
2026-01-15T20:43:19.468065763Z stdout F Waiting for suse-observability-elasticsearch-master-headless:9200...
2026-01-15T20:43:19.471504897Z stderr F nc: getaddrinfo for host "suse-observability-elasticsearch-master-headless" port 9200: Name does not resolve
...


$ kubectl get statefulset -n $SO suse-observability-elasticsearch-master -oyaml
apiVersion: apps/v1
kind: StatefulSet
...
Events:
  Type     Reason            Age                   From                    Message
  ----     ------            ----                  ----                    -------
  Normal   SuccessfulCreate  27m                   statefulset-controller  create Claim data-suse-observability-elasticsearch-master-0 Pod suse-observability-elasticsearch-master-0 in StatefulSet suse-observability-elasticsearch-master success
  Warning  FailedCreate      5m55s (x19 over 27m)  statefulset-controller  create Pod suse-observability-elasticsearch-master-0 in StatefulSet suse-observability-elasticsearch-master failed error: pods "suse-observability-elasticsearch-master-0" is forbidden: violates PodSecurity "baseline:latest": privileged (container "configure-sysctl" must not set securityContext.privileged=true)
```



```sh
// set priviledged,  to ensure the elastic can be installed

kubectl label --overwrite ns suse-observability pod-security.kubernetes.io/enforce=privileged

kubectl rollout restart -n suse-observability suse-observability-elasticsearch-master
```

Quickly, all PoDs are ready.

#### Create a new LB type service

This service, maps the SO accessing to a format like `http:///192.168.122.233:8090`.

The backend SO router is served on port `8080`, and the service port can be any value you like.

```sh
cat > observ-expose.yaml << 'EOF'
apiVersion: v1
kind: Service
metadata:
  name: suse-observability-router-expose
  namespace: suse-observability
spec:
  ipFamilies:
  - IPv4
  ipFamilyPolicy: SingleStack
  loadBalancerIP: 192.168.122.233
  ports:
  - name: router
    port: 8090
    protocol: TCP
    targetPort: 8080
  selector:
    app.kubernetes.io/component: router
    app.kubernetes.io/instance: suse-observability
    app.kubernetes.io/name: suse-observability
  sessionAffinity: None
  type: LoadBalancer
EOF

kubectl create -f observ-expose.yaml
```

Access the SO UI via above like `http:///192.168.122.233:8090`.

### Uninstall SO

```sh

helm uninstall suse-observability -n suse-observability

manually remove all remaining pvcs on suse-observability namespace
```

The SO can be fully removed from Harvester and re-install is possible.

## Register a second Harvester cluster to SUSE-Observability

### Generate service token from SO

Follow the SO UI guide, create a new cluster with name like `harvester1`, and generate a token.

```sh
generate a service token from SO UI:


e.g.: svctok-OxZrVBdB5g7UUESBNW1ozx5u7NrqaaBx


cluster name: harvester1
StackState Ingest URL:

http://192.168.122.233:8090/receiver/stsAgent

```

### Prepare helm repo on Harvester

```
helm repo:

helm repo add suse-observability https://charts.rancher.com/server-charts/prime/suse-observability


helm repo update
```

#### Hack the agent helm chart

Run `wget https://charts.rancher.com/server-charts/prime/suse-observability/index.yaml` to get an index file.

As of 2026.01.16, the agent has latest version 1.2.14

```yaml
    name: suse-observability-agent
    urls:
    - suse-observability-agent-1.2.14.tgz
    version: 1.2.14
```

### Install SO-agent pack on Harvester

Copy the helm related command on SO UI.


```sh
export SERVICE_TOKEN=svctok-OxZrVBdB5g7UUESBNW1ozx5u7NrqaaBx

$ helm upgrade --install \
--namespace suse-observability \
--create-namespace \
--set-string 'stackstate.apiKey'=$SERVICE_TOKEN \
--set-string 'stackstate.cluster.name'='harvester1' \
--set-string 'stackstate.url'='http://192.168.122.233:8090/receiver/stsAgent' \
suse-observability-agent suse-observability/suse-observability-agent



$ helm list --namespace suse-observability

NAME                    	NAMESPACE         	REVISION	UPDATED                                	STATUS  	CHART                          	APP VERSION
suse-observability-agent	suse-observability	1       	2026-01-16 16:12:27.067810638 +0000 UTC	deployed	suse-observability-agent-1.2.14	3.0.0


$ kubectl get pods -n suse-observability

NAME                                                      READY   STATUS    RESTARTS   AGE
suse-observability-agent-checks-agent-64c5ff6d84-mvj2p    1/1     Running   0          98s
suse-observability-agent-cluster-agent-68bd859548-sqz77   1/1     Running   0          98s
suse-observability-agent-logs-agent-cpxkf                 1/1     Running   0          98s
suse-observability-agent-node-agent-x6dhg                 2/2     Running   0          98s
suse-observability-agent-rbac-agent-7888cc47c9-zwsz4      1/1     Running   0          98s
```

:::note

1. The `stackstate.url` must strictly begin with the aforementioned `baseUrl`.

2. There are 5 pods installed, check log of them to toubleshooting.

:::

Wait a while, the `SO` will show Harvester instance is ready and it can be browsered.

Refer: https://github.com/harvester/harvester/issues/8282#issuecomment-3757034464


### Example of remote SO down

```sh
kubectl logs -n suse-observability suse-observability-agent-cluster-agent-68bd859548-sqz77
...
2026-01-16 16:22:18 UTC | CLUSTER | ERROR | (comp/forwarder/defaultforwarder/worker.go:187 in process) | Too many errors for endpoint 'http://192.168.122.233:8090/receiver/stsAgent/api/v1/series': retrying later
2026-01-16 16:22:20 UTC | CLUSTER | INFO | (pkg/serializer/serializer.go:488 in SendProcessesMetadata) | Sent processes metadata payload, size: 2689771 bytes.
2026-01-16 16:22:20 UTC | CLUSTER | INFO | (pkg/collector/corechecks/cluster/kubeapi/kubernetes_topology.go:284 in Run) | Topology Check for cluster: harvester1 completed successfully
2026-01-16 16:22:20 UTC | CLUSTER | ERROR | (comp/forwarder/defaultforwarder/worker.go:187 in process) | Too many errors for endpoint 'http://192.168.122.233:8090/receiver/stsAgent/intake/': retrying later
```

### Uninstall agent

```sh
list instllation

helm list --namespace suse-observability

uninstall agent

helm uninstall suse-observability-agent -n suse-observability
```

## Others

An example of the generated config file, and the tailered `sizing_values.yaml` to request fewer resources.

```sh

cat suse-observability-values/templates/baseConfig_values.yaml
---
# Source: suse-observability-values/templates/baseConfig_values.yaml
global:
  imageRegistry: "registry.rancher.com"

stackstate:
  baseUrl: "http://192.168.122.233:8090"
  authentication:
    adminPassword: "$2a$10$bXJrkiezRT13/d55BJHfGenUD0O2HEfhHH/2byIjocKwni6r7QVY2"
  apiKey:
    key: "uYYVb9wcmM3sqTHwH4uNJvDXaWDs4fDz"
  license:
    key: "V2ZN0-..."
...

$ cat suse-observability-values/templates/sizing_values.yaml
---
# Source: suse-observability-values/templates/sizing_values.yaml
# profile trial
clickhouse:
  replicaCount: 1
  persistence:
    size: 2Gi

elasticsearch:
  prometheus-elasticsearch-exporter:
    resources:
      limits:
        cpu: "50m"
        memory: "50Mi"
      requests:
        cpu: "50m"
        memory: "50Mi"
  minimumMasterNodes: 1
  replicas: 1
  # Only overriding memory settings
  esJavaOpts: "-Xmx1500m -Xms1500m -Des.allow_insecure_settings=true"
  resources:
    requests:
      cpu: 50m
      memory: 250Mi
    limits:
      cpu: 1000m
      memory: 2500Mi
  volumeClaimTemplate:
    resources:
      requests:
        storage: 2Gi
hbase:
  version: "2.5"
  deployment:
    mode: "Mono"
  stackgraph:
    persistence:
      size: 2Gi
    resources:
      requests:
        memory: "225Mi"
        cpu: "50m"
      limits:
        cpu: "1500m"
        memory: "2250Mi"
  tephra:
    resources:
      limits:
        cpu: "100m"
        memory: "512Mi"
      requests:
        memory: "128Mi"
        cpu: "50m"
    replicaCount: 1
kafka:
  defaultReplicationFactor: 1
  offsetsTopicReplicationFactor: 1
  replicaCount: 1
  transactionStateLogReplicationFactor: 1
  resources:
    requests:
      cpu: "80m"
      memory: "248Mi"
    limits:
      memory: "2048Mi"
      cpu: "1600m"
  persistence:
    size: 6Gi
stackstate:
  features:
    server:
      split: false

  components:
    all:
      extraEnv:
        open:
          CONFIG_FORCE_stackstate_topologyQueryService_maxStackElementsPerQuery: "100"
          CONFIG_FORCE_stackstate_topologyQueryService_maxLoadedElementsPerQuery: "100"
          CONFIG_FORCE_stackstate_agents_agentLimit: "10"
          # 14 day retention as NONHA is not completely meant for prod installations
          CONFIG_FORCE_stackgraph_retentionWindowMs: 259200000
          CONFIG_FORCE_stackstate_traces_retentionDays: "3"
    server:
      extraEnv:
        open:
          CONFIG_FORCE_stackstate_sync_initializationBatchParallelism: "1"
          CONFIG_FORCE_stackstate_healthSync_initialLoadParallelism: "1"
          CONFIG_FORCE_stackstate_stateService_initializationParallelism: "1"
          CONFIG_FORCE_stackstate_stateService_initialLoadTransactionSize: "2500"
      resources:
        limits:
          ephemeral-storage: 3Gi
          cpu: 3000m
          memory: 5Gi
        requests:
          cpu: 150m
          memory: 256Mi
    e2es:
      retention: 3
      resources:
        requests:
          memory: "128Mi"
          cpu: "50m"
        limits:
          memory: "512Mi"
    correlate:
      resources:
        requests:
          memory: "125Mi"
          cpu: "100m"
        limits:
          cpu: "1000m"
          memory: "1250Mi"
    receiver:
      retention: 3
      split:
        enabled: false
      extraEnv:
        open:
          CONFIG_FORCE_akka_http_host__connection__pool_max__open__requests: "256"
      resources:
        requests:
          memory: "100Mi"
          cpu: "100m"
        limits:
          memory: "1000Mi"
          cpu: "2000m"
    vmagent:
      resources:
        limits:
          memory: "640Mi"
        requests:
          memory: "128Mi"
    ui:
      replicaCount: 1
victoria-metrics-0:
  server:
    resources:
      requests:
        cpu: "100m"
        memory: 150Mi
      limits:
        cpu: "1000m"
        memory: 1750Mi
    persistentVolume:
      size: 2Gi
    retentionPeriod: 3d
  backup:
    vmbackup:
      resources:
        requests:
          memory: 50Mi
        limits:
          memory: 50Mi
victoria-metrics-1:
  enabled: false
  server:
    persistentVolume:
      size: 2Gi
    retentionPeriod: 3d
  backup:
    vmbackup:
      resources:
        requests:
          memory: 128Mi
        limits:
          memory: 256Mi
zookeeper:
  replicaCount: 1
  persistence:
    size: 2Gi
```