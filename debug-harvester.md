It is possible to coding&debugging Harvester without GoLand.

1. Follow guide to install the Harvester from ISO or PXE
https://github.com/harvester/harvester

2. Good help documents for starters
https://github.com/harvester/harvester/wiki/Harvester-Development-Overview
https://github.com/harvester/harvester/wiki/Setting-Up-Harvester-Development-Environment

When you have no GoLand, or feel it complex, no worry.


3. Be familiar with main PODs of Harvester
Harvester is on-top of K8S and others, the K8S is the fundemental.

4. Some information of a running Harvester






**Develop Harvester**

It is exciting, the GO compiled bin can run directly on your local Linux, given you can do kubectl on it

Build and run:
Get the source code and build:

```
rancher@rancherserver1:~/source_code/harvester$ make build
...
INFO[0095] docker cp /go/src/github.com/harvester/harvester/bin .
INFO[0097] docker cp /go/src/github.com/harvester/harvester/dist .
INFO[0097] docker cp /go/src/github.com/harvester/harvester/package .


rancher@rancherserver1:~/source_code/harvester$ ./bin/harvester
harvester          harvester-webhook
```
Those two are the executable file of harvester, harvester-webhook, when they are deployed with K8s, they are the core parts of the container image.

Try to run ```harvester``` directly

```
rancher@rancherserver1:~/source_code/harvester$ ./bin/harvester
2021/12/14 10:03:24 WARNING: proto: message VersionResponse is already registered
A future release will panic on registration conflicts. See:
https://developers.google.com/protocol-buffers/docs/reference/go/faq#namespace-conflict

2021/12/14 10:03:24 WARNING: proto: file "rpc.proto" is already registered
A future release will panic on registration conflicts. See:
https://developers.google.com/protocol-buffers/docs/reference/go/faq#namespace-conflict

NAME:
   Harvester API Server

USAGE:
   harvester [global options] command [command options] [arguments...]

VERSION:
   04c1f199 (04c1f199)

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --kubeconfig value              Kube config for accessing k8s cluster [$KUBECONFIG]
   --profile-listen-address value  Address to listen on for profiling (default: "0.0.0.0:6060")
   --debug                         Enable debug logs [$HARVESTER_DEBUG]
   --trace                         Enable trace logs [$HARVESTER_TRACE]
   --log-format value              Log format (default: "text") [$HARVESTER_LOG_FORMAT]
   --threadiness value             Specify controller threads (default: 10) [$THREADINESS]
   --http-port value               HTTP listen port (default: 8080) [$HARVESTER_SERVER_HTTP_PORT]
   --https-port value              HTTPS listen port (default: 8443) [$HARVESTER_SERVER_HTTPS_PORT]
   --namespace value               The default namespace to store management resources [$NAMESPACE]
   --hci-mode                      Enable HCI mode. Additional controllers are registered in HCI mode [$HCI_MODE]
   --rancher-embedded              Specify whether the Harvester is running with embedded Rancher mode, default to false [$RANCHER_EMBEDDED]
   --help, -h                      show help
   --version, -v                   print the version
FATA[0000] Required flag "namespace" not set
rancher@rancherserver1:~/source_code/harvester$

```

It is failed, but it gives a big bunch of useful information, fill all those parameters, you will get it run.
(When you follow the guide to setup GoLand, most of the params are related to here)


check rancher ip, notice following IP and port, especially the "port"
 via kubectl get service -A
```
NAMESPACE                  NAME                                          TYPE           CLUSTER-IP      EXTERNAL-IP    PORT(S)                        AGE
kube-system                ingress-expose                                LoadBalancer   10.53.200.67    192.168.3.80   443:32661/TCP,80:31834/TCP     3h6m
...
```
notice the 443->32661 mapping


if need to see debug message, add "--debug" to the program
```
./bin/harvester --kubeconfig /home/rancher/.kube/config --debug --namespace harvester-system  --hci-mode true  --rancher-embedded true --rancher-server-url https://192.168.3.80:32661
```
Your havester could run in you native Linux, fell cool.

Co-operate with the cluster
```
rancher@rancherserver1:~$ kubectl get pods --namespace harvester-system
NAME                                                   READY   STATUS    RESTARTS   AGE
harvester-d544ddb6f-h64b5                              1/1     Running   3          69m    ------ harvester
harvester-load-balancer-59bf75f489-wngrc               1/1     Running   8          7d5h
harvester-network-controller-manager-55cd87b58-9s46z   1/1     Running   10         7d5h
harvester-network-controller-manager-55cd87b58-fn8dd   1/1     Running   10         7d5h
harvester-network-controller-nbb9q                     1/1     Running   11         7d5h
harvester-node-disk-manager-n569v                      1/1     Running   10         7d5h
harvester-webhook-67744f845f-96v8d                     0/1     Pending   0          69m   --------harvester-webhook
harvester-webhook-79ccf7f4d6-dvd9l                     1/1     Running   2          3h19m
kube-vip-cloud-provider-0                              1/1     Running   8          7d5h
kube-vip-hf42c                                         1/1     Running   8          7d5h
virt-api-86455cdb7d-8mkhz                              1/1     Running   9          7d5h
virt-api-86455cdb7d-xk5cj                              1/1     Running   10         7d5h
virt-controller-5f649999dd-7fd2j                       1/1     Running   9          7d5h
virt-controller-5f649999dd-dp9lz                       1/1     Running   8          7d5h
virt-handler-5sgmr                                     1/1     Running   8          7d5h
virt-operator-56c5bdc7b8-c2s9g                         1/1     Running   9          7d5h
```

Your local run is tended to replace the one in cluster, then your developed new feature/bug fix will be debugged/verified.
```
First, delete below line
  "management.cattle.io/scale-available": "3"
via
  kubectl edit deploy harvester-webhook -n harvester-system

Then,
kubectl scale --replicas=0 deployment/harvester -n harvester-system
```
check the running pods again, you will not find harvester POD anymore.

Open the web-ui to play with Harvester:
```
Suppose your local developing Linux, which runs harvester directly, has IP: 192.168.3.31,
then visit https://192.168.3.31:8443 
```

Further speed up:
The original "make build" will run in container, takes few minutes.
When you have a GO installed, build it directly via
```
go build  -o ./bin/harv_main ./main.go
```
run it like  (the harv_main could be any name you like)
```
./bin/harv_main --kubeconfig /home/rancher/.kube/config --debug --namespace harvester-system  --hci-mode true  --rancher-embedded true --rancher-server-url https://192.168.3.80:32661
```



**Develop Harvester-webhook**
A bit complex, the GO compiled bin can not run directly on you local Linux.
TBD.
