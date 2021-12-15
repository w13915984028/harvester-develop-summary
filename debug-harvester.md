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

Harvester-webhook follows general rule of K8S webhook, it works tightly with K8s components, within K8s cluster network. It is not easy to let it run locally.

We will build docker image and apply it in Harvester.


1. build docker image
```
Dockerfile.webhook.local is copid from Dockerfile.webhook under same folder, and changed a bit, entrypoint-webhook-local is copied from entrypoint-webhook, add "--debug" 

```
```
rancher@rancherserver1:~/source_code/harvester/package$ cat Dockerfile.webhook.local
FROM alpine
RUN apk update && apk add -u --no-cache tini bash

COPY ./entrypoint-webhook-local.sh  /usr/bin/entrypoint.sh
COPY ./harvester-webhook /usr/bin/harvester-webhook

RUN chmod +x /usr/bin/entrypoint.sh
RUN chmod +x /usr/bin/harvester-webhook


RUN find /usr/bin/entrypoint.sh
RUN find /usr/bin/harvester-webhook


ENTRYPOINT ["/usr/bin/entrypoint.sh"]
```

```
rancher@rancherserver1:~/source_code/harvester/package$ cat entrypoint-webhook-local.sh
#!/bin/bash
set -e

exec tini -- harvester-webhook "${@}" --debug
```

```
rancher@rancherserver1:~/source_code/harvester/package$ docker build -f Dockerfile.webhook.local -t rancher/harvester-webhook:wj03 .
Sending build context to Docker daemon  240.1MB
Step 1/9 : FROM alpine
 ---> 14119a10abf4
Step 2/9 : RUN apk update && apk add -u --no-cache tini bash
 ---> Using cache
 ---> 34fb9ebd3528
Step 3/9 : COPY ./entrypoint-webhook-local.sh  /usr/bin/entrypoint.sh
 ---> be516422a514
Step 4/9 : COPY ./harvester-webhook /usr/bin/harvester-webhook
 ---> 0887ba8cc870
Step 5/9 : RUN chmod +x /usr/bin/entrypoint.sh
 ---> Running in e32aa6e4e4df
Removing intermediate container e32aa6e4e4df
 ---> aada5c85aa69
Step 6/9 : RUN chmod +x /usr/bin/harvester-webhook
 ---> Running in a46f58a7cc5f
Removing intermediate container a46f58a7cc5f
 ---> 20fbbcaf64aa
Step 7/9 : RUN find /usr/bin/entrypoint.sh
 ---> Running in 0764a3e04c0d
/usr/bin/entrypoint.sh
Removing intermediate container 0764a3e04c0d
 ---> aa4b4d767b8b
Step 8/9 : RUN find /usr/bin/harvester-webhook
 ---> Running in 3c920a1fa980
/usr/bin/harvester-webhook
Removing intermediate container 3c920a1fa980
 ---> 6760704a4419
Step 9/9 : ENTRYPOINT ["/usr/bin/entrypoint.sh"]
 ---> Running in d5e87f57d888
Removing intermediate container d5e87f57d888
 ---> 1151e0b1fd91
Successfully built 1151e0b1fd91
Successfully tagged rancher/harvester-webhook:wj03
```

2.check docker image
```
rancher@rancherserver1:~/source_code/harvester/package$ docker image ls | grep wj03
rancher/harvester-webhook                                         wj03                                                 1151e0b1fd91   8 seconds ago    124MB
```

3. run the docker image locally, check if it works basically

```
rancher@rancherserver1:~/source_code/harvester/package$ docker run -it rancher/harvester-webhook:wj03 /bin/bash
2021/12/15 15:03:21 WARNING: proto: message VersionResponse is already registered
A future release will panic on registration conflicts. See:
https://developers.google.com/protocol-buffers/docs/reference/go/faq#namespace-conflict

2021/12/15 15:03:21 WARNING: proto: file "rpc.proto" is already registered
A future release will panic on registration conflicts. See:
https://developers.google.com/protocol-buffers/docs/reference/go/faq#namespace-conflict

NAME:
   Harvester Admission Webhook Server

USAGE:
   harvester-webhook [global options] command [command options] [arguments...]

VERSION:
   67c9ce40 (67c9ce40)

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --kubeconfig value              Kube config for accessing k8s cluster [$KUBECONFIG]
   --profile-listen-address value  Address to listen on for profiling (default: "0.0.0.0:6060")
   --debug                         Enable debug logs [$HARVESTER_DEBUG]
   --trace                         Enable trace logs [$HARVESTER_TRACE]
   --log-format value              Log format (default: "text") [$HARVESTER_LOG_FORMAT]
   --threadiness value             Specify controller threads (default: 5) [$THREADINESS]
   --https-port value              HTTPS listen port (default: 9443) [$HARVESTER_WEBHOOK_SERVER_HTTPS_PORT]
   --namespace value               The harvester namespace [$NAMESPACE]
   --controller-user value         The harvester controller username [$HARVESTER_CONTROLLER_USER_NAME]
   --gc-user value                 The system username that performs garbage collection (default: "system:serviceaccount:kube-system:generic-garbage-collector") [$GARBAGE_COLLECTION_USER_NAME]
   --help, -h                      show help
   --version, -v                   print the version
FATA[0000] Required flag "namespace" not set
rancher@rancherserver1:~/source_code/harvester/package$
```

4.
At this moment, at least, the harvester-webhook docker image can run, and the bin harvester-webhook can also run.
This can save you a big amount of time if you try to verify them in the cluster, in case any errors.


5. save and zip (optional) the image:
```
docker save -o harv-webhook-wj03.docker.img **rancher/harvester-webhook:wj03**

 tar -czf harv_webhook_docker.tar harv-webhook-wj03.docker.img
```

6. scp image to remote Harvester main node.  (when you have docker repository, it is also working)
```
scp ./harv-webhook-wj03.docker.img rancher@192.168.3.76://home/rancher/
```

7. In harvester main node: load docker image
Extract and load docker image:
```
sudo -i docker image rm rancher/harvester-webhook:wj03
sudo -i docker image load -i /home/rancher/harv-webhook-wj03.docker.img
sudo -i docker image ls | grep wj03
```

8.stop the POD and change deploy file
```
delete this line
"management.cattle.io/scale-available": "3"
via
kubectl edit deploy harvester-webhook -n harvester-system

kubectl scale --replicas=0 deployment/harvester-webhook -n harvester-system

kubectl edit deploy harvester -n harvester-system
change "image":
image: rancher/harvester-webhook:v1.0.0-rc1
---->
image: rancher/harvester-webhook:wj03
```

9.start the POD
```
kubectl scale --replicas=1 deployment/harvester-webhook -n harvester-system
```

10. check logs
```
kubectl logs deployment/harvester-webhook -n harvester-system
```

with the "--debug" enabled in locally build docker image, you will get a lot of useful information for debug.


Notice, when "kube scale"
Stop Harvester first, then Harvester-webhook
Start Harvester-webhook first, then Harvester (may be local running program)
