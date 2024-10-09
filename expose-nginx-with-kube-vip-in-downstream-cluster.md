# Expose Nginx With Kube Vip in Downstream Cluster

## Precondition

1. Import a Harvester cluster as node-driver to Rancher manager.

2. Create a downstream RKE2 cluster from node-driver Harvester.

The `rke2-ingress-nginx` is enabled by default.

3. Deploy the `cloud-provider-harvester`.

The `kube-vip` is enabled by default.

## The Available Objects on this Guest Cluster

### DaemonSet rke2-ingress-nginx-controller

```
apiVersion: apps/v1
kind: DaemonSet
metadata:
  annotations:
    meta.helm.sh/release-name: rke2-ingress-nginx
    meta.helm.sh/release-namespace: kube-system
  generation: 1
  labels:
    app.kubernetes.io/component: controller
    app.kubernetes.io/instance: rke2-ingress-nginx
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: rke2-ingress-nginx
    app.kubernetes.io/part-of: rke2-ingress-nginx
    app.kubernetes.io/version: 1.10.4
    helm.sh/chart: rke2-ingress-nginx-4.10.401
  name: rke2-ingress-nginx-controller
  namespace: kube-system
```

### DaemonSet kube-vip

```
apiVersion: apps/v1
kind: DaemonSet
metadata:
  annotations:
    deprecated.daemonset.template.generation: "1"
    meta.helm.sh/release-name: harvester-cloud-provider
    meta.helm.sh/release-namespace: kube-system
  labels:
    app.kubernetes.io/managed-by: Helm
  name: kube-vip
  namespace: kube-system
spec:
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app.kubernetes.io/instance: harvester-cloud-provider
      app.kubernetes.io/name: kube-vip
  template:
    metadata:
      creationTimestamp: null
      labels:
        app.kubernetes.io/instance: harvester-cloud-provider
        app.kubernetes.io/name: kube-vip
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: node-role.kubernetes.io/controlplane
                operator: Exists
            - matchExpressions:
              - key: node-role.kubernetes.io/control-plane
                operator: Exists
```

## How to Expose an IP to External

### DaemonSet rke2-ingress-nginx-controller

```
- apiVersion: v1
  kind: Pod
  metadata:
    annotations:
      cni.projectcalico.org/containerID: 80d1aed46353a704d843318b27544551c9275e59567d88327bd99d57f0cf4330
      cni.projectcalico.org/podIP: 10.42.35.65/32
      cni.projectcalico.org/podIPs: 10.42.35.65/32
    creationTimestamp: "2024-10-02T20:49:47Z"
    generateName: rke2-ingress-nginx-controller-
...
      name: rke2-ingress-nginx-controller
      ports:
      - containerPort: 80
        hostPort: 80
        name: http
        protocol: TCP
      - containerPort: 443
        hostPort: 443
        name: https
        protocol: TCP
      - containerPort: 8443
        name: webhook
        protocol: TCP
...
    hostIP: 10.55.12.180
    hostIPs:
    - ip: 10.55.12.180
    phase: Running
    podIP: 10.42.35.65
    podIPs:
    - ip: 10.42.35.65            
```

A couple of `rke2-ingress-nginx-controller-` PODs are created, they can be accessed via the POD IP (cluster internal) or the PODs' running host IP. But when you need a fixed IP to access the nginx, it is a challenge.

### Expose the Nginx via a VIP

As `kube-vip` has been deployed, it works as a loadbalancer provider in this cluster, you can create a Kubernetes `Service` of type `LoadBalancer`, and select the `rke2-nginx` related deployment as the service backend.

The `kube-vip` is responsible for (optional) DHCP requesting and advertising the `vip` to external, the `service` is responsible for forwarding the traffic to a group of backend pods.

You can also check them on the Harvester cluster, Harvester has a service `ingress-expose` to expose its web server.

#### Static VIP

Change the `kube-vip.io/loadbalancerIPs: 192.168.122.148` and selector `app.kubernetes.io/name: rke2-ingress-nginx-example` to match your cluster deployment.

```
cat > service_lb_with_static_ip.yaml << 'EOF'
apiVersion: v1
kind: Service
metadata:
  annotations:
    kube-vip.io/ignore-service-security: "true"
    kube-vip.io/loadbalancerIPs: 192.168.122.148
  name: ingress-expose-example
  namespace: kube-system
spec:
  allocateLoadBalancerNodePorts: true
  externalTrafficPolicy: Cluster
  internalTrafficPolicy: Cluster
  selector:
    app.kubernetes.io/name: rke2-ingress-nginx-example
  sessionAffinity: None
  type: LoadBalancer
  ipFamilies:
  - IPv4
  ipFamilyPolicy: SingleStack
  ports:
  - name: https
    port: 443
    protocol: TCP
    targetPort: 443
  - name: http
    port: 80
    protocol: TCP
    targetPort: 80  
EOF
```

run: `kubectl create -f service_lb_with_static_ip.yaml`.

The created service:

```
$kubectl get service -n kube-system

NAME                                      TYPE           CLUSTER-IP      EXTERNAL-IP       PORT(S)                      AGE
ingress-expose-example                    LoadBalancer   10.53.158.76    192.168.122.148   443:30621/TCP,80:31829/TCP   10s

apiVersion: v1
kind: Service
metadata:
  annotations:
    kube-vip.io/ignore-service-security: "true"
    kube-vip.io/loadbalancerIPs: 192.168.122.148
    kube-vip.io/vipHost: harv41                        // the VIP current is floating on this host now
  creationTimestamp: "2024-10-09T06:55:14Z"
  name: ingress-expose-example
  namespace: kube-system
  resourceVersion: "85801"
  uid: dd76cf3a-5443-46bc-b8a6-64346fd8e8f4
spec:
  allocateLoadBalancerNodePorts: true
  clusterIP: 10.53.158.76
  clusterIPs:
  - 10.53.158.76
  externalTrafficPolicy: Cluster
  internalTrafficPolicy: Cluster
  ipFamilies:
  - IPv4
  ipFamilyPolicy: SingleStack
  ports:
  - name: https
    nodePort: 30621
    port: 443
    protocol: TCP
    targetPort: 443
  - name: http
    nodePort: 31829
    port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app.kubernetes.io/name: rke2-ingress-nginx-example
  sessionAffinity: None
  type: LoadBalancer
status:
  loadBalancer:
    ingress:
    - ip: 192.168.122.148
      ports:
      - port: 443
        protocol: TCP
      - port: 80
        protocol: TCP
```

#### DCHP based VIP with a given MAC

General DHCP server supports IP/MAC binding, you can prepare a MAC for the service, thus it will always get the same IP.

Note: Change the `kube-vip.io/hwaddr:` and `app.kubernetes.io/name: rke2-ingress-nginx-example-dhcp`.

```
cat > service_lb_with_dhcp_based_ip_with_mac.yaml << 'EOF'
apiVersion: v1
kind: Service
metadata:
  annotations:
    kube-vip.io/ignore-service-security: "true"
    kube-vip.io/loadbalancerIPs: "0.0.0.0"
    kube-vip.io/hwaddr: "52:54:00:03:3b:27"
  name: ingress-expose-example-dhcp
  namespace: kube-system
spec:
  allocateLoadBalancerNodePorts: true
  externalTrafficPolicy: Cluster
  internalTrafficPolicy: Cluster
  selector:
    app.kubernetes.io/name: rke2-ingress-nginx-example-dhcp
  sessionAffinity: None
  type: LoadBalancer
  ports:
  - name: https
    port: 443
    protocol: TCP
    targetPort: 443
  - name: http
    port: 80
    protocol: TCP
    targetPort: 80   
EOF
```

created example:
```
harv41:/home/rancher # kk get service -n kube-system ingress-expose-example-dhcp -oyaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    kube-vip.io/hwaddr: 52:54:00:03:3b:27
    kube-vip.io/ignore-service-security: "true"
    kube-vip.io/loadbalancerIPs: 0.0.0.0
    kube-vip.io/requestedIP: 192.168.122.142     // DHCP server returned IP
    kube-vip.io/vipHost: harv41                  // the VIP current is floating on this host now
  creationTimestamp: "2024-10-09T07:06:49Z"
  name: ingress-expose-example-dhcp
  namespace: kube-system
  resourceVersion: "93520"
  uid: 4cfa4812-43d6-4980-b0e3-65b2468b5f4f
spec:
  allocateLoadBalancerNodePorts: true
  clusterIP: 10.53.195.145
  clusterIPs:
  - 10.53.195.145
  externalTrafficPolicy: Cluster
  internalTrafficPolicy: Cluster
  ipFamilies:
  - IPv4
  ipFamilyPolicy: SingleStack
  ports:
  - name: https
    nodePort: 32709
    port: 443
    protocol: TCP
    targetPort: 443
  - name: http
    nodePort: 31747
    port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app.kubernetes.io/name: rke2-ingress-nginx-example-dhcp
  sessionAffinity: None
  type: LoadBalancer
status:
  loadBalancer:
    ingress:
    - ip: 192.168.122.142
      ports:
      - port: 443
        protocol: TCP
      - port: 80
        protocol: TCP
```

#### DCHP based VIP without a given MAC

If you don't have a MAC, the annotation `kube-vip.io/hwaddr:` can be deleted, the `kube-vip` will generate a dynamic MAC, but the DHCP server may offer different IPs if this dynamic MAC is not binded to an IP.

```
cat > service_lb_with_dhcp_based_ip_without_mac.yaml << 'EOF'
apiVersion: v1
kind: Service
metadata:
  annotations:
    kube-vip.io/ignore-service-security: "true"
    kube-vip.io/loadbalancerIPs: "0.0.0.0"
  name: ingress-expose-example-dhcp
  namespace: kube-system
spec:
  allocateLoadBalancerNodePorts: true
  externalTrafficPolicy: Cluster
  internalTrafficPolicy: Cluster
  selector:
    app.kubernetes.io/name: rke2-ingress-nginx-example-dhcp
  sessionAffinity: None
  type: LoadBalancer
  ports:
  - name: https
    port: 443
    protocol: TCP
    targetPort: 443
  - name: http
    port: 80
    protocol: TCP
    targetPort: 80   
EOF

```

The created service:

```
apiVersion: v1
kind: Service
metadata:
  annotations:
    kube-vip.io/hwaddr: 00:00:6c:eb:4d:14      // dynamically generated MAC
    kube-vip.io/ignore-service-security: "true"
    kube-vip.io/loadbalancerIPs: 0.0.0.0
    kube-vip.io/requestedIP: 192.168.122.111   // DHCP server returned IP
    kube-vip.io/vipHost: harv41                // the VIP current is floating on this host now
  creationTimestamp: "2024-10-09T07:08:09Z"
  name: ingress-expose-example-dhcp
  namespace: kube-system
  resourceVersion: "94407"
  uid: fc1d8b5f-d8cf-4b1c-8e05-63fe74b1ef24
spec:
  allocateLoadBalancerNodePorts: true
  clusterIP: 10.53.31.223
  clusterIPs:
  - 10.53.31.223
  externalTrafficPolicy: Cluster
  internalTrafficPolicy: Cluster
  ipFamilies:
  - IPv4
  ipFamilyPolicy: SingleStack
  ports:
  - name: https
    nodePort: 31479
    port: 443
    protocol: TCP
    targetPort: 443
  - name: http
    nodePort: 31577
    port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app.kubernetes.io/name: rke2-ingress-nginx-example-dhcp
  sessionAffinity: None
  type: LoadBalancer
status:
  loadBalancer:
    ingress:
    - ip: 192.168.122.111
      ports:
      - port: 443
        protocol: TCP
      - port: 80
        protocol: TCP
```


### Troubleshooting

1. The logs of `kube-vip`

```
time="2024-10-09T06:49:50Z" level=info msg="Starting kube-vip.io [v0.8.1]"
time="2024-10-09T06:49:50Z" level=info msg="namespace [kube-system], Mode: [ARP], Features(s): Control Plane:[false], Services:[true]"
time="2024-10-09T06:49:50Z" level=info msg="No interface is specified for VIP in config, auto-detecting default Interface"
time="2024-10-09T06:49:50Z" level=info msg="kube-vip will bind to interface [mgmt-br]"
time="2024-10-09T06:49:50Z" level=info msg="prometheus HTTP server started"
time="2024-10-09T06:49:50Z" level=warning msg="Node name is missing from the config, fall back to hostname"
time="2024-10-09T06:49:50Z" level=info msg="Using node name [harv41]"
time="2024-10-09T06:49:50Z" level=info msg="Starting Kube-vip Manager with the ARP engine"
time="2024-10-09T06:49:50Z" level=info msg="beginning services leadership, namespace [harvester-system], lock name [plndr-svcs-lock], id [harv41]"
I1009 06:49:50.671799       1 leaderelection.go:250] attempting to acquire leader lease harvester-system/plndr-svcs-lock...
I1009 06:49:50.717275       1 leaderelection.go:260] successfully acquired lease harvester-system/plndr-svcs-lock
time="2024-10-09T06:49:50Z" level=info msg="(svcs) starting services watcher for all namespaces"
time="2024-10-09T06:49:50Z" level=info msg="(svcs) adding VIP [192.168.122.144] via mgmt-br for [kube-system/ingress-expose]"
time="2024-10-09T06:49:50Z" level=info msg="[service] synchronised in 2ms"
time="2024-10-09T06:55:14Z" level=info msg="(svcs) adding VIP [192.168.122.148] via mgmt-br for [kube-system/ingress-expose-example]"
time="2024-10-09T06:55:14Z" level=info msg="[service] synchronised in 14ms"
time="2024-10-09T06:55:14Z" level=warning msg="(svcs) already found existing address [192.168.122.148] on adapter [mgmt-br]"
time="2024-10-09T06:55:17Z" level=warning msg="Re-applying the VIP configuration [192.168.122.148] to the interface [mgmt-br]"
time="2024-10-09T06:57:15Z" level=info msg="[LOADBALANCER] Stopping load balancers"
time="2024-10-09T06:57:15Z" level=info msg="[VIP] Releasing the Virtual IP [192.168.122.148]"
time="2024-10-09T06:57:15Z" level=info msg="Removed [dd76cf3a-5443-46bc-b8a6-64346fd8e8f4] from manager, [1] advertised services remain"
time="2024-10-09T06:57:15Z" level=info msg="(svcs) [kube-system/ingress-expose-example] has been deleted"
time="2024-10-09T07:00:54Z" level=info msg="(svcs) [kube-system/ingress-expose-example] has been deleted"
time="2024-10-09T07:02:26Z" level=info msg="(svcs) [kube-system/ingress-expose-example] has been deleted"
```

2. The kubenertes object `endpointslice`.

Eech service has at least one `endpointslice` object, which records the backend PODs' IP as `endpoints`, if this list is empty, it means the service `selector` has issue, it does not match any backend PODs.


```
harv41:~ # kubectl get endpointslice -n kube-system ingress-expose-kcshj -oyaml
addressType: IPv4
apiVersion: discovery.k8s.io/v1
endpoints:
- addresses:
  - 10.52.0.97
  conditions:
    ready: true
    serving: true
    terminating: false
  nodeName: harv41
  targetRef:
    kind: Pod
    name: rke2-ingress-nginx-controller-fr44m
    namespace: kube-system
    uid: 8ff9bd68-9fce-4efd-b3cf-133e4dc85c1d
kind: EndpointSlice
metadata:
..
  name: ingress-expose-kcshj
  namespace: kube-system
  ownerReferences:
  - apiVersion: v1
    blockOwnerDeletion: true
    controller: true
    kind: Service
    name: ingress-expose
    uid: 683dd969-3797-4fae-999c-66c46b797a5d
..
ports:
- name: http
  port: 80
  protocol: TCP
- name: https-internal
  port: 443
  protocol: TCP
```
