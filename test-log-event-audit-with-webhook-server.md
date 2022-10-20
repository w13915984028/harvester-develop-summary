# Test logging, audit, event with webhook server

## Reference

### HEP DOC:
https://github.com/joshmeranda/harvester/blob/logging/enhancements/20220525-system-logging.md
https://github.com/w13915984028/harvester/blob/hep578/enhancements/20220822-event-and-audit.md


### Graylog investigation
https://github.com/w13915984028/harvester-develop-summary/blob/main/integrate-harvester-logging-with-log-servers.md


## Test

### Setup webhook server

The following code demos how to setup a simple webhook server, and which simply prints all received data.

[note] The webhook received data is a bunch of `logging`/`event`/`audit`.
```

cat > simple-webhook-server.py  << 'EOF'
import os
import web
from datetime import datetime

urls = ('/.*', 'hooks')

app = web.application(urls, globals())

class hooks:
    def POST(self):
        data = web.data().decode("utf-8") 
        ds = data.split("\n")

        # last element maybe empty
        if len(ds) > 1 and ds[-1] == '':
            ds.pop()

        print('\nDATA RECEIVED @ {}, length {length:,}, element-count {ec}'.format(datetime.now(), length=len(data), ec=len(ds))) 
       
        idx = 0
        for x in ds:
             idx += 1
             print("-------- element {} --------".format(idx))
             print(x+'\n')

        print("total {} printed\n".format(idx))

        return 'OK'

if __name__ == '__main__':
    print("usage: export PORT=1234 to set http server port number as 1234")
    pt=8080 #default
    if 'PORT' in os.environ:
        pt = os.environ['PORT']
    print("start a simple webhook server, PORT {} @ {} \n".format(pt, datetime.now()))
    app.run()
EOF


```

export PORT=8094

python3 simple-webhook-server.py


#### outputsample: logging

> DATA RECEIVED @ 2022-09-20 19:38:26.100861, length 16,869, element-count 13

> -------- element 1 --------

> {"stream":"stdout","logtag":"F","message":"10.52.0.186 - - [20/Sep/2022:19:34:35 +0000] \"GET /charts/index.yaml HTTP/1.1\" 200 10314 \"-\" \"Go-http-client/1.1\"","kubernetes":{"pod_name":"harvester-cluster-repo-56b5f7b585-qn2rr","namespace_name":"cattle-system","pod_id":"ce0302bf-928a-47f4-83d3-4957221b63f7","labels":{"app":"harvester-cluster-repo","pod-template-hash":"56b5f7b585"},"annotations":{"cni.projectcalico.org/containerID":"767441a6c1d13ceaeac5411274f150564ea97da9c7a2ea954208cb79ade2e190","cni.projectcalico.org/podIP":"10.52.0.174/32","cni.projectcalico.org/podIPs":"10.52.0.174/32","k8s.v1.cni.cncf.io/network-status":"[{\n    \"name\": \"k8s-pod-network\",\n    \"ips\": [\n        \"10.52.0.174\"\n    ],\n    \"default\": true,\n    \"dns\": {}\n}]","k8s.v1.cni.cncf.io/networks-status":"[{\n    \"name\": \"k8s-pod-network\",\n    \"ips\": [\n        \"10.52.0.174\"\n    ],\n    \"default\": true,\n    \"dns\": {}\n}]","kubernetes.io/psp":"global-unrestricted-psp"},"host":"harv41","container_name":"httpd","docker_id":"3fbb5108c8966eefd5d34991514ec6d32aabfe983d25ffcc5d14e72dd90ae579","container_hash":"sha256:5cb6ec87a3ccb57583c285c88dab15fd23de36c41972917ee7960ae1dab567d2","container_image":"docker.io/rancher/harvester-cluster-repo:master"}}


#### outputsample: audit

> DATA RECEIVED @ 2022-10-20 13:43:02.266439, length 2,667, element-count 3

> -------- element 1 --------

> {"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"965580b8-8715-4f88-828d-9af7f086a6df","stage":"RequestReceived","requestURI
":"/apis/authorization.k8s.io/v1/subjectaccessreviews?timeout=10s","verb":"create","user":{"username":"system:serviceaccount:kube-system:rke2-metrics-se
rver","uid":"19592b74-d331-462d-8d87-6ee7a3d9e7be","groups":["system:serviceaccounts","system:serviceaccounts:kube-system","system:authenticated"],"extr
a":{"authentication.kubernetes.io/pod-name":["rke2-metrics-server-67697454f8-92v88"],"authentication.kubernetes.io/pod-uid":["bc2b7356-a12a-49f1-b105-a5
6c0cac4ea1"]}},"sourceIPs":["10.52.0.108"],"userAgent":"Go-http-client/2.0","objectRef":{"resource":"subjectaccessreviews","apiGroup":"authorization.k8s
.io","apiVersion":"v1"},"requestReceivedTimestamp":"2022-10-20T11:41:52.766616Z","stageTimestamp":"2022-10-20T11:41:52.766616Z"}

> -------- element 2 --------

> {"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"9dae6ad1-ffe3-4979-b4ea-28f67bc46bd5","stage":"RequestReceived","requestURI
":"/apis/authorization.k8s.io/v1/subjectaccessreviews?timeout=10s","verb":"create","user":{"username":"system:serviceaccount:kube-system:rke2-metrics-se
rver","uid":"19592b74-d331-462d-8d87-6ee7a3d9e7be","groups":["system:serviceaccounts","system:serviceaccounts:kube-system","system:authenticated"],"extr
a":{"authentication.kubernetes.io/pod-name":["rke2-metrics-server-67697454f8-92v88"],"authentication.kubernetes.io/pod-uid":["bc2b7356-a12a-49f1-b105-a5
6c0cac4ea1"]}},"sourceIPs":["10.52.0.108"],"userAgent":"Go-http-client/2.0","objectRef":{"resource":"subjectaccessreviews","apiGroup":"authorization.k8s
.io","apiVersion":"v1"},"requestReceivedTimestamp":"2022-10-20T11:41:52.766624Z","stageTimestamp":"2022-10-20T11:41:52.766624Z"}


#### outputsample: event

> DATA RECEIVED @ 2022-09-20 19:41:04.317929, length 6,810, element-count 2

> -------- element 1 --------

> {"stream":"stdout","logtag":"F","message":"{\"verb\":\"UPDATED\",\"event\":{\"metadata\":{\"name\":\"harvester-harvester-vm-import-controller.1716a6e89f3d3b37\",\"namespace\":\"harvester-system\",\"uid\":\"bd709db8-e2f6-45c6-8c31-ab7530583bdf\",\"resourceVersion\":\"112306\",\"creationTimestamp\":\"2022-09-20T19:03:50Z\",\"managedFields\":[{\"manager\":\"kube-controller-manager\",\"operation\":\"Update\",\"apiVersion\":\"v1\",\"time\":\"2022-09-20T19:03:50Z\"}]},\"involvedObject\":{\"kind\":\"PersistentVolumeClaim\",\"namespace\":\"harvester-system\",\"name\":\"harvester-harvester-vm-import-controller\",\"uid\":\"a23037be-e517-4111-b880-ba2ed53266a8\",\"apiVersion\":\"v1\",\"resourceVersion\":\"3549\"},\"reason\":\"FailedBinding\",\"message\":\"no persistent volumes available for this claim and no storage class is set\",\"source\":{\"component\":\"persistentvolume-controller\"},\"firstTimestamp\":\"2022-09-20T19:03:50Z\",\"lastTimestamp\":\"2022-09-20T19:38:50Z\",\"count\":141,\"type\":\"Normal\",\"eventTime\":null,\"reportingComponent\":\"\",\"reportingInstance\":\"\"},\"old_event\":{\"metadata\":{\"name\":\"harvester-harvester-vm-import-controller.1716a6e89f3d3b37\",\"namespace\":\"harvester-system\",\"uid\":\"bd709db8-e2f6-45c6-8c31-ab7530583bdf\",\"resourceVersion\":\"108653\",\"creationTimestamp\":\"2022-09-20T19:03:50Z\",\"managedFields\":[{\"manager\":\"kube-controller-manager\",\"operation\":\"Update\",\"apiVersion\":\"v1\",\"time\":\"2022-09-20T19:03:50Z\"}]},\"involvedObject\":{\"kind\":\"PersistentVolumeClaim\",\"namespace\":\"harvester-system\",\"name\":\"harvester-harvester-vm-import-controller\",\"uid\":\"a23037be-e517-4111-b880-ba2ed53266a8\",\"apiVersion\":\"v1\",\"resourceVersion\":\"3549\"},\"reason\":\"FailedBinding\",\"message\":\"no persistent volumes available for this claim and no storage class is set\",\"source\":{\"component\":\"persistentvolume-controller\"},\"firstTimestamp\":\"2022-09-20T19:03:50Z\",\"lastTimestamp\":\"2022-09-20T19:33:50Z\",\"count\":121,\"type\":\"Normal\",\"eventTime\":null,\"reportingComponent\":\"\",\"reportingInstance\":\"\"}}","kubernetes":{"pod_name":"harvester-default-event-tailer-0","namespace_name":"cattle-logging-system","pod_id":"33d19f48-f123-4e96-978c-04a875b1a8c8","labels":{"app.kubernetes.io/instance":"harvester-default-event-tailer","app.kubernetes.io/name":"event-tailer","controller-revision-hash":"harvester-default-event-tailer-d59f6fddc","statefulset.kubernetes.io/pod-name":"harvester-default-event-tailer-0"},"annotations":{"cni.projectcalico.org/containerID":"2c16f399bce2204a3f6a53b047cbfdd40cd9f249bbb5df37ada4796600c97485","cni.projectcalico.org/podIP":"10.52.0.203/32","cni.projectcalico.org/podIPs":"10.52.0.203/32","k8s.v1.cni.cncf.io/network-status":"[{\n    \"name\": \"k8s-pod-network\",\n    \"ips\": [\n        \"10.52.0.203\"\n    ],\n    \"default\": true,\n    \"dns\": {}\n}]","k8s.v1.cni.cncf.io/networks-status":"[{\n    \"name\": \"k8s-pod-network\",\n    \"ips\": [\n        \"10.52.0.203\"\n    ],\n    \"default\": true,\n    \"dns\": {}\n}]","kubernetes.io/psp":"global-unrestricted-psp"},"host":"harv41","container_name":"event-tailer","docker_id":"4118de11721ed00164377ab57998f7e92698ad543efd49cc918798c8978a73c0","container_hash":"sha256:00f781432854f8fc0d0ae40db0aeda5f46ae2bde113a26346c3aa5a93af157fb","container_image":"docker.io/banzaicloud/eventrouter:v0.1.0"}}



### 1. test logging

```

cat > co-logging1.yaml << 'EOF'
apiVersion: logging.banzaicloud.io/v1beta1
kind: ClusterOutput
metadata:
  name: harvester-logging-webhook
  namespace: cattle-logging-system
spec:
  http:
    endpoint: "http://192.168.122.159:8098/"
    open_timeout: 3
    format:
      type: "json"
    buffer:
      chunk_limit_size: 3MB
      timekey: 2m
      timekey_wait: 1m
EOF

cat > cf-logging1.yaml << 'EOF'
apiVersion: logging.banzaicloud.io/v1beta1
kind: ClusterFlow
metadata:
  name: harvester-logging-webhook
  namespace: cattle-logging-system
spec:
  filters:
  - tag_normaliser: {}
  match: 
  globalOutputRefs:
    - harvester-logging-webhook
EOF



kubectl apply -f co-logging1.yaml

kubectl apply -f cf-logging1.yaml


```

#### check the input

Normally, after the `apply`, the related objects are created and the `ACTIVE` is `true`.

```
harv41:~ # kubectl get clusterflow -A
NAMESPACE               NAME                        ACTIVE   PROBLEMS
cattle-logging-system   harvester-logging-webhook   true     


harv41:~ # kubectl get clusteroutput -A
NAMESPACE               NAME                        ACTIVE   PROBLEMS
cattle-logging-system   harvester-logging-webhook   true     
harv41:~ # 
```

### 2. test event

```

cat > co-event1.yaml << 'EOF'
apiVersion: logging.banzaicloud.io/v1beta1
kind: ClusterOutput
metadata:
  name: harvester-event-webhook
  namespace: cattle-logging-system
spec:
  http:
    endpoint: "http://192.168.122.159:8094/"
    open_timeout: 3
    format:
      type: "json"
    buffer:
      chunk_limit_size: 3MB
      timekey: 2m
      timekey_wait: 1m
EOF

cat > cf-event1.yaml << 'EOF'
apiVersion: logging.banzaicloud.io/v1beta1
kind: ClusterFlow
metadata:
  name: harvester-event-webhook
  namespace: cattle-logging-system
spec:
  filters:
  - tag_normaliser: {}
  match:
  - select:
      labels:
        app.kubernetes.io/name: event-tailer
  globalOutputRefs:
    - harvester-event-webhook
EOF


kubectl apply -f co-event1.yaml

kubectl apply -f cf-event1.yaml


```

### 3. test audit

```

cat > co-audit1.yaml << 'EOF'
apiVersion: logging.banzaicloud.io/v1beta1
kind: ClusterOutput
metadata:
  name: "harvester-audit-webhook"
  namespace: "cattle-logging-system"
spec:
  http:
    endpoint: "http://192.168.122.159:8096/"
    open_timeout: 3
    format: 
      type: "json"
    buffer:
      chunk_limit_size: 3MB
      timekey: 2m
      timekey_wait: 1m
  loggingRef: harvester-kube-audit-log-ref   # this reference is fixed and must be here
EOF


cat > cf-audit1.yaml << 'EOF'
apiVersion: logging.banzaicloud.io/v1beta1
kind: ClusterFlow
metadata:
  name: "harvester-audit-webhook"
  namespace: "cattle-logging-system"
spec:
  globalOutputRefs:
    - "harvester-audit-webhook"
  loggingRef: harvester-kube-audit-log-ref  # this reference is fixed and must be here
EOF

kubectl apply -f co-audit1.yaml

kubectl apply -f cf-audit1.yaml


```

