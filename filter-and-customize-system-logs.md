# Filter and customize system logs

Related issue: https://github.com/harvester/harvester/issues/3470


## The fetched system logs in Harvester

In harvester/harvester-installer/pkg/config/templates/patch/rancher-logging/100.1.3+up3.17.7/configmap.yaml, following logs are fetched.

```
    [INPUT]
        Name              systemd
        Tag               rke2
        Path              {{ .Values.systemdLogPath }}
        Systemd_Filter    _SYSTEMD_UNIT=rke2-server.service
        Systemd_Filter    _SYSTEMD_UNIT=rke2-agent.service
        Systemd_Filter    _SYSTEMD_UNIT=rancherd.service
        Systemd_Filter    _SYSTEMD_UNIT=rancher-system-agent.service
        Systemd_Filter    _SYSTEMD_UNIT=wicked.service
        Systemd_Filter    _SYSTEMD_UNIT=iscsid.service
        Systemd_Filter    _TRANSPORT=kernel
        {{- if .Values.additionalLoggingSources.rke2.stripUnderscores }}
        Strip_Underscores On
        {{- end }}

    [INPUT]
        Name              tail
        Tag               rke2
        Path              /var/lib/rancher/rke2/agent/logs/kubelet.log

    [INPUT]
        Name              tail
        Tag               rke2
        Path              /var/log/audit/audit.log
```

## What is missing

Unlike Kubernetes logs (they have format `${namespace_name}.${pod_name}.${container_name}` ), there are no unified format in system logs.

When we use the simple webhook server to receive log, the related log formats are:

### systemd log

#### rke2-server

> {"level":"warning","msg":"Proxy error: write failed: write tcp 127.0.0.1:9345->127.0.0.1:56296: write: connection reset by peer","_BOOT_ID":"5f9904016fd94d3b8da2671c3ae11888","PRIORITY":"6","SYSLOG_FACILITY":"3","_UID":"0","_GID":"0","_SELINUX_CONTEXT":"unconfined\n","_SYSTEMD_SLICE":"system.slice","_TRANSPORT":"stdout","_CAP_EFFECTIVE":"1ffffffffff","_HOSTNAME":"harv2","_MACHINE_ID":"8df1c5739989f60cc125adce6437b4a7","_SYSTEMD_CGROUP":"/system.slice/rke2-server.service","_SYSTEMD_UNIT":"rke2-server.service","SYSLOG_IDENTIFIER":"rke2","_COMM":"rke2","_EXE":"/opt/rke2/bin/rke2","_CMDLINE":"\"/opt/rke2/bin/rke2 server\"","_SYSTEMD_INVOCATION_ID":"af5a62e0d421461e8479210d91d95c7c","_STREAM_ID":"007f8179b2c641f88e440acaec75063f","_PID":"8169"}


#### wicked
> {"_BOOT_ID":"5f9904016fd94d3b8da2671c3ae11888","PRIORITY":"6","SYSLOG_FACILITY":"3","_UID":"0","_GID":"0","_SELINUX_CONTEXT":"unconfined\n","_SYSTEMD_SLICE":"system.slice","_TRANSPORT":"stdout","_CAP_EFFECTIVE":"1ffffffffff","_HOSTNAME":"harv2","_MACHINE_ID":"8df1c5739989f60cc125adce6437b4a7","SYSLOG_IDENTIFIER":"wicked","_COMM":"wicked","_SYSTEMD_CGROUP":"/system.slice/wicked.service","_SYSTEMD_UNIT":"wicked.service","MESSAGE":"ens3            device-ready","_SYSTEMD_INVOCATION_ID":"965e65e0061d40f08feea74390054f0d","_STREAM_ID":"710e316e070a43be9fb8b68bb6b8fceb","_PID":"20220"}


#### kernel itself
> {"_TRANSPORT":"kernel","SYSLOG_FACILITY":"0","SYSLOG_IDENTIFIER":"kernel","_BOOT_ID":"5f9904016fd94d3b8da2671c3ae11888","PRIORITY":"6","_HOSTNAME":"harv2","_MACHINE_ID":"8df1c5739989f60cc125adce6437b4a7","MESSAGE":"mgmt-bo: (slave ens3): Enslaving as a backup interface with a down link","_SOURCE_MONOTONIC_TIMESTAMP":"11815829637"}


### kubelet log

> {"level":"I","thread_id":"8257","filename":"reconciler.go","linenumber":"352","message":"\"operationExecutor.VerifyControllerAttachedVolume started for volume \\\"config\\\" (UniqueName: \\\"kubernetes.io/secret/5b92e75e-1ada-45ec-8710-5b09fe1ad6d3-config\\\") pod \\\"rancher-logging-root-fluentd-0\\\" (UID: \\\"5b92e75e-1ada-45ec-8710-5b09fe1ad6d3\\\") \" pod=\"cattle-logging-system/rancher-logging-root-fluentd-0\""}

## Filter and customize in ClusterFlow

We need at least two steps:

(1) Filter system related logs

The `ClusterFlow` filters can be like:

```
  filters:
  - grep:
     regexp:
     - key: SYSLOG_IDENTIFIER
       pattern: /(^.*rke2.*$|^.*wicked.*$|^.*iscsid.*$|^kernel$)/
```

Via the log key (field) `SYSLOG_IDENTIFIER`, the `pattern` `/(^.*rke2.*$|^.*wicked.*$|^.*iscsid.*$|^kernel$)/` filters those service units.

(2) Customize them with new fields

The `ClusterFlow` filters can further have `record_transformer` to add/modify records:

```
  - record_transformer:
      records:
      - appname: ${record["SYSLOG_IDENTIFIER"]}
```

The example add a new filed `appname`, which is from the log record field `SYSLOG_IDENTIFIER`.

### systemd log

#### ClusterFlow

```
cat > cf-logging1.yaml << 'EOF'
apiVersion: logging.banzaicloud.io/v1beta1
kind: ClusterFlow
metadata:
  name: harvester-logging-webhook
  namespace: cattle-logging-system
spec:
  filters:
  - grep:
     regexp:
     - key: SYSLOG_IDENTIFIER
       pattern: /(^.*rke2.*$|^.*wicked.*$|^.*iscsid.*$|^kernel$)/
  - record_transformer:
      records:
      - appname: ${record["SYSLOG_IDENTIFIER"]}
  - tag_normaliser: {}
  match: 
  globalOutputRefs:
    - harvester-logging-webhook
EOF
```

#### Output

The desired logs are filtered, and a new field "appname" is added.

> {"level":"warning","msg":"Proxy error: write failed: write tcp 127.0.0.1:9345->127.0.0.1:56296: write: connection reset by peer","_BOOT_ID":"5f9904016fd94d3b8da2671c3ae11888","PRIORITY":"6","SYSLOG_FACILITY":"3","_UID":"0","_GID":"0","_SELINUX_CONTEXT":"unconfined\n","_SYSTEMD_SLICE":"system.slice","_TRANSPORT":"stdout","_CAP_EFFECTIVE":"1ffffffffff","_HOSTNAME":"harv2","_MACHINE_ID":"8df1c5739989f60cc125adce6437b4a7","_SYSTEMD_CGROUP":"/system.slice/rke2-server.service","_SYSTEMD_UNIT":"rke2-server.service","SYSLOG_IDENTIFIER":"rke2","_COMM":"rke2","_EXE":"/opt/rke2/bin/rke2","_CMDLINE":"\"/opt/rke2/bin/rke2 server\"","_SYSTEMD_INVOCATION_ID":"af5a62e0d421461e8479210d91d95c7c","_STREAM_ID":"007f8179b2c641f88e440acaec75063f","_PID":"8169","appname":"rke2"}

> {"_BOOT_ID":"5f9904016fd94d3b8da2671c3ae11888","PRIORITY":"6","SYSLOG_FACILITY":"3","_UID":"0","_GID":"0","_SELINUX_CONTEXT":"unconfined\n","_SYSTEMD_SLICE":"system.slice","_TRANSPORT":"stdout","_CAP_EFFECTIVE":"1ffffffffff","_HOSTNAME":"harv2","_MACHINE_ID":"8df1c5739989f60cc125adce6437b4a7","SYSLOG_IDENTIFIER":"wicked","_COMM":"wicked","_SYSTEMD_CGROUP":"/system.slice/wicked.service","_SYSTEMD_UNIT":"wicked.service","MESSAGE":"ens3            device-ready","_SYSTEMD_INVOCATION_ID":"965e65e0061d40f08feea74390054f0d","_STREAM_ID":"710e316e070a43be9fb8b68bb6b8fceb","_PID":"20220","appname":"wicked"}

> {"_TRANSPORT":"kernel","SYSLOG_FACILITY":"0","SYSLOG_IDENTIFIER":"kernel","_BOOT_ID":"5f9904016fd94d3b8da2671c3ae11888","PRIORITY":"6","_HOSTNAME":"harv2","_MACHINE_ID":"8df1c5739989f60cc125adce6437b4a7","MESSAGE":"mgmt-bo: (slave ens3): Enslaving as a backup interface with a down link","_SOURCE_MONOTONIC_TIMESTAMP":"11815829637","appname":"kernel"}


### kubelet log

#### ClusterFlow

```
cat > cf-logging-kubelet.yaml << 'EOF'
apiVersion: logging.banzaicloud.io/v1beta1
kind: ClusterFlow
metadata:
  name: harvester-logging-kubelet
  namespace: cattle-logging-system
spec:
  filters:
  - grep:
     regexp:
     - key: thread_id
       pattern: /^\d+$/
  - record_transformer:
      records:
      - appname: "kubelet"
  - tag_normaliser: {}
  match: 
  globalOutputRefs:
    - harvester-logging-webhook
EOF
```

#### Output

> {"level":"I","thread_id":"8257","filename":"reconciler.go","linenumber":"352","message":"\"operationExecutor.VerifyControllerAttachedVolume started for volume \\\"output-secret\\\" (UniqueName: \\\"kubernetes.io/secret/5b92e75e-1ada-45ec-8710-5b09fe1ad6d3-output-secret\\\") pod \\\"rancher-logging-root-fluentd-0\\\" (UID: \\\"5b92e75e-1ada-45ec-8710-5b09fe1ad6d3\\\") \" pod=\"cattle-logging-system/rancher-logging-root-fluentd-0\"","appname":"kubelet"}


## Utilize the new fields in ClusterOutput

Issue https://github.com/harvester/harvester/issues/3470 defines a `ClusterOutput`, which takes `kubernetes.pod_name` as `app_name_field`.

```
clusteroutput:
apiVersion: logging.banzaicloud.io/v1beta1
kind: ClusterOutput
metadata:
  name: logging
  namespace: cattle-logging-system
spec:
  syslog:
    buffer:
      ...
    format:
      app_name_field: kubernetes.pod_name
```

A new similar `ClusterOutput` can be defined to take `appname` as the `app_name_field`, and let the related `ClusterFlow` to match this new `ClusterOutput`.

```
clusteroutput:
apiVersion: logging.banzaicloud.io/v1beta1
kind: ClusterOutput
metadata:
  name: logging
  namespace: cattle-logging-system
spec:
  syslog:
    buffer:
      ...
    format:
      app_name_field: appname
```


## Syslog output config

The `syslog` output plugin takes specific record field as `log_field`, each of the kind of log has an example.


```
# kubelet log
# syslog log field is from record `message`
# has no field `_HOSTNAME`

cat > cf-kubelet.yaml << 'EOF'
apiVersion: logging.banzaicloud.io/v1beta1
kind: ClusterFlow
metadata:
  name: logging-kubelet
  namespace: cattle-logging-system
spec:
  filters:
  - grep:
     regexp:
     - key: thread_id
       pattern: /^\d+$/
  - record_transformer:
      records:
      - custom-cluster-based-hostname: harvester-staging
      - appname: "kubelet"
  - tag_normaliser: {}
  match: 
  globalOutputRefs:
    - logging-kubelet
EOF


cat > co-kubelet.yaml << 'EOF'
apiVersion: logging.banzaicloud.io/v1beta1
kind: ClusterOutput
metadata:
  name: logging-kubelet
  namespace: cattle-logging-system
spec:
  syslog:
    buffer:
      total_limit_size: 2GB
      flush_thread_count: 8
      timekey: 10m
      timekey_use_utc: true
      timekey_wait: 1m
    format:
      app_name_field: appname
      hostname_field: custom-cluster-based-hostname
      log_field: message
      rfc6587_message_size: false
    host: 192.168.122.159
    insecure: true
    port: 514
    transport: udp
EOF


kubectl apply -f co-kubelet.yaml
sleep 5
kubectl apply -f cf-kubelet.yaml


# rke2 log
# `_SYSTEMD_UNIT` is `rke2-server.service`, but `SYSLOG_IDENTIFIER` is `rke2`
# syslog log field is from record `msg`
# also has `_HOSTNAME` field
# the `${record["_HOSTNAME"]}` can also be used to have a combined name with HOSTNAME
# output example:  2023-04-25T09:19:04.000000+00:00 harvester-staging.harv2 rke2-server.service Starting batch/v1, Kind=Job controller

cat > cf-rke2.yaml << 'EOF'
apiVersion: logging.banzaicloud.io/v1beta1
kind: ClusterFlow
metadata:
  name: logging-rke2
  namespace: cattle-logging-system
spec:
  filters:
  - grep:
     regexp:
     - key: _SYSTEMD_UNIT
       pattern: /(^.*rke2.*$)/
  - record_transformer:
      records:
      - custom-cluster-based-hostname: harvester-staging.${record["_HOSTNAME"]}
      - appname: ${record["_SYSTEMD_UNIT"]}
  globalOutputRefs:
  - logging-rke2
EOF


cat > co-rke2.yaml << 'EOF'
apiVersion: logging.banzaicloud.io/v1beta1
kind: ClusterOutput
metadata:
  name: logging-rke2
  namespace: cattle-logging-system
spec:
  syslog:
    buffer:
      total_limit_size: 2GB
      flush_thread_count: 8
      timekey: 10m
      timekey_use_utc: true
      timekey_wait: 1m
    format:
      app_name_field: appname
      hostname_field: custom-cluster-based-hostname
      log_field: msg
      rfc6587_message_size: false
    host: 192.168.122.159
    insecure: true
    port: 514
    transport: udp
EOF


kubectl apply -f co-rke2.yaml
sleep 5
kubectl apply -f cf-rke2.yaml






# kernel, wicked, iscsid log
# kernel itself has no `_SYSTEMD_UNIT`, only `SYSLOG_IDENTIFIER`
# syslog log field is from record `MESSAGE`
# also has `_HOSTNAME` field

cat > cf-os.yaml << 'EOF'
apiVersion: logging.banzaicloud.io/v1beta1
kind: ClusterFlow
metadata:
  name: logging-os
  namespace: cattle-logging-system
spec:
  filters:
  - grep:
     regexp:
     - key: SYSLOG_IDENTIFIER
       pattern: /(^.*wicked.*$|^.*iscsid.*$|^kernel$)/
  - record_transformer:
      records:
      - custom-cluster-based-hostname: harvester-staging
      - appname: ${record["SYSLOG_IDENTIFIER"]}
  globalOutputRefs:
  - logging-os
EOF


cat > co-os.yaml << 'EOF'
apiVersion: logging.banzaicloud.io/v1beta1
kind: ClusterOutput
metadata:
  name: logging-os
  namespace: cattle-logging-system
spec:
  syslog:
    buffer:
      total_limit_size: 2GB
      flush_thread_count: 8
      timekey: 10m
      timekey_use_utc: true
      timekey_wait: 1m
    format:
      app_name_field: appname
      hostname_field: custom-cluster-based-hostname
      log_field: MESSAGE
      rfc6587_message_size: false
    host: 192.168.122.159
    insecure: true
    port: 514
    transport: udp
EOF

kubectl apply -f co-os.yaml
sleep 5
kubectl apply -f cf-os.yaml


# normal k8s log
# syslog log field is from record `message`

cat > cf-k8s.yaml << 'EOF'
apiVersion: logging.banzaicloud.io/v1beta1
kind: ClusterFlow
metadata:
  name: logging
  namespace: cattle-logging-system
spec:
  filters:
  - record_transformer:
      records:
      - custom-cluster-based-hostname: harvester-staging
  globalOutputRefs:
  - logging
EOF

cat > co-k8s.yaml << 'EOF'
apiVersion: logging.banzaicloud.io/v1beta1
kind: ClusterOutput
metadata:
  name: logging
  namespace: cattle-logging-system
spec:
  syslog:
    buffer:
      total_limit_size: 2GB
      flush_thread_count: 8
      timekey: 10m
      timekey_use_utc: true
      timekey_wait: 1m
    format:
      app_name_field: kubernetes.pod_name
      hostname_field: custom-cluster-based-hostname
      log_field: message
      rfc6587_message_size: false
    host: 192.168.122.159
    insecure: true
    port: 514
    transport: udp
EOF

kubectl apply -f co-k8s.yaml
sleep 5
kubectl apply -f cf-k8s.yaml
```

### Generated syslog log files

The example is from an `rsyslog` server in Ubuntu-20.04.

```
..$ sudo -i ls /var/log/remotelogs/192.168.122.206 -alth
total 20M
-rw-r--r-- 1 syslog syslog 800K Apr 25 11:19 rke2-server.service.log
-rw-r--r-- 1 syslog syslog 1.4M Apr 25 11:17 kubelet.log
-rw-r--r-- 1 syslog syslog 723K Apr 25 11:16 kernel.log


-rw-r--r-- 1 syslog syslog 5.2K Apr 25 11:11 rancher-monitoring-crd-create-9bls4.log
drwx------ 2 syslog syslog 4.0K Apr 25 11:11 .
-rw-r--r-- 1 syslog syslog 5.2K Apr 25 11:11 rancher-monitoring-crd-create-shc57.log
-rw-r--r-- 1 syslog syslog  27K Apr 25 11:11 longhorn-csi-plugin-qmpxt.log
-rw-r--r-- 1 syslog syslog  675 Apr 25 11:11 longhorn-post-upgrade-8gh2j.log
-rw-r--r-- 1 syslog syslog 5.2K Apr 25 11:11 rancher-monitoring-crd-create-9rtwd.log
-rw-r--r-- 1 syslog syslog 229K Apr 25 11:11 rancher-logging-kube-audit-fluentbit-zxcfd.log
-rw-r--r-- 1 syslog syslog 584K Apr 25 11:11 fleet-agent-6ffcdc7f94-rwqfb.log
-rw-r--r-- 1 syslog syslog  493 Apr 25 11:11 longhorn-ui-bf9cbc848-svlx9.log
-rw-r--r-- 1 syslog syslog 126K Apr 25 11:11 rancher-8bc7b78d5-f5rlb.log
-rw-r--r-- 1 syslog syslog  20K Apr 25 11:11 virt-handler-4789l.log
-rw-r--r-- 1 syslog syslog  65K Apr 25 11:11 harvester-79bb4ccf44-s7mld.log
-rw-r--r-- 1 syslog syslog  293 Apr 25 11:11 rancher-logging-kube-audit-fluentd-0.log
-rw-r--r-- 1 syslog syslog  14K Apr 25 11:11 longhorn-admission-webhook-7c7d9b9dd9-b5sn6.log
-rw-r--r-- 1 syslog syslog 3.1M Apr 25 11:11 virt-api-877b485b-nrbnd.log
-rw-r--r-- 1 syslog syslog 112K Apr 25 11:11 kube-controller-manager-harv2.log
-rw-r--r-- 1 syslog syslog  59K Apr 25 11:11 longhorn-conversion-webhook-5f86fd8fcc-tpnkb.log
-rw-r--r-- 1 syslog syslog  41K Apr 25 11:11 virt-operator-6775887765-zqvlr.log
-rw-r--r-- 1 syslog syslog  611 Apr 25 11:11 rke2-coredns-rke2-coredns-58fd75f64b-xpfh4.log
-rw-r--r-- 1 syslog syslog  39K Apr 25 11:11 longhorn-manager-swgdk.log
-rw-r--r-- 1 syslog syslog  48K Apr 25 11:11 rke2-ingress-nginx-controller-p5qsn.log
-rw-r--r-- 1 syslog syslog 3.8K Apr 25 11:11 csi-snapshotter-6c4d5c4fc9-vdtwm.log
-rw-r--r-- 1 syslog syslog 5.5M Apr 25 11:11 harvester-default-event-tailer-0.log
-rw-r--r-- 1 syslog syslog 4.0K Apr 25 11:11 csi-provisioner-77b757f445-4r8sc.log
-rw-r--r-- 1 syslog syslog 2.0K Apr 25 11:11 rke2-coredns-rke2-coredns-autoscaler-768bfc5985-.log
-rw-r--r-- 1 syslog syslog  25K Apr 25 11:11 longhorn-conversion-webhook-5f86fd8fcc-mhm72.log
-rw-r--r-- 1 syslog syslog 1.4M Apr 25 11:11 -.log
-rw-r--r-- 1 syslog syslog  13K Apr 25 11:11 longhorn-admission-webhook-7c7d9b9dd9-h4bqx.log
-rw-r--r-- 1 syslog syslog  13K Apr 25 11:11 rancher-webhook-7577bd9c47-5zpdd.log
-rw-r--r-- 1 syslog syslog  24K Apr 25 11:11 harvester-webhook-6766b5d5c8-q9rrb.log
-rw-r--r-- 1 syslog syslog  74K Apr 25 11:11 virt-controller-59f5d9cc47-vkv8g.log
-rw-r--r-- 1 syslog syslog  18K Apr 25 11:11 rancher-logging-6b5b478f78-4gvfb.log
-rw-r--r-- 1 syslog syslog  950 Apr 25 11:11 rancher-logging-root-fluentd-0.log
-rw-r--r-- 1 syslog syslog 226K Apr 25 11:11 snapshot-controller-6cc58d858f-cdd5s.log
-rw-r--r-- 1 syslog syslog 6.2K Apr 25 11:11 virt-controller-59f5d9cc47-lx5cs.log
-rw-r--r-- 1 syslog syslog  53K Apr 25 11:11 harvester-cluster-repo-b96bd7cd5-pjb5j.log
-rw-r--r-- 1 syslog syslog 5.6K Apr 25 11:11 kube-vip-cloud-provider-0.log
-rw-r--r-- 1 syslog syslog 190K Apr 25 11:11 snapshot-controller-6cc58d858f-88qkp.log
-rw-r--r-- 1 syslog syslog 7.6K Apr 25 11:11 harvester-node-disk-manager-4t2x6.log
-rw-r--r-- 1 syslog syslog  11K Apr 25 11:11 fleet-controller-5d546cc5b8-2t4lx.log
-rw-r--r-- 1 syslog syslog 4.2K Apr 25 11:11 longhorn-driver-deployer-78fcc6fcfb-q26d7.log
-rw-r--r-- 1 syslog syslog 2.9K Apr 25 11:11 rke2-metrics-server-74f878b999-jp59d.log
-rw-r--r-- 1 syslog syslog 5.0K Apr 25 11:11 helm-operation-cd2rn.log
-rw-r--r-- 1 syslog syslog 4.7K Apr 25 11:11 helm-operation-dqbcr.log
-rw-r--r-- 1 syslog syslog 3.8K Apr 25 11:11 harvester-node-manager-7b9jz.log
-rw-r--r-- 1 syslog syslog 5.4K Apr 25 11:11 cloud-controller-manager-harv2.log
-rw-r--r-- 1 syslog syslog 1.5K Apr 25 11:11 instance-manager-e-07ccdad3.log
-rw-r--r-- 1 syslog syslog 1.3K Apr 25 11:11 instance-manager-r-e2767389.log
-rw-r--r-- 1 syslog syslog 8.3K Apr 25 11:11 harvester-load-balancer-78f7f798bf-5zvng.log
-rw-r--r-- 1 syslog syslog 5.8K Apr 25 11:11 gitjob-5c5979d844-fh9cl.log
-rw-r--r-- 1 syslog syslog 3.6K Apr 25 11:11 csi-attacher-5d4cc8cfc8-t9k98.log
-rw-r--r-- 1 syslog syslog 6.8K Apr 25 11:11 csi-provisioner-77b757f445-cjppr.log
-rw-r--r-- 1 syslog syslog 4.2K Apr 25 11:11 csi-resizer-76fdffd797-74vp7.log
-rw-r--r-- 1 syslog syslog 3.8K Apr 25 11:11 csi-provisioner-77b757f445-m6dpr.log
-rw-r--r-- 1 syslog syslog 3.2K Apr 25 11:11 csi-resizer-76fdffd797-c5wrq.log
-rw-r--r-- 1 syslog syslog 3.0K Apr 25 11:11 csi-resizer-76fdffd797-s4v59.log
-rw-r--r-- 1 syslog syslog 4.0K Apr 25 11:11 csi-snapshotter-6c4d5c4fc9-2htvp.log
-rw-r--r-- 1 syslog syslog 4.6K Apr 25 11:11 csi-snapshotter-6c4d5c4fc9-cz7fn.log
-rw-r--r-- 1 syslog syslog 3.3K Apr 25 11:11 csi-attacher-5d4cc8cfc8-rpb54.log
-rw-r--r-- 1 syslog syslog 3.0M Apr 25 11:11 rke2-canal-xzwc8.log
-rw-r--r-- 1 syslog syslog 598K Apr 25 11:11 kube-apiserver-harv2.log
-rw-r--r-- 1 syslog syslog 346K Apr 25 11:11 rancher-logging-rke2-journald-aggregator-7q8zm.log
-rw-r--r-- 1 syslog syslog 127K Apr 25 11:11 etcd-harv2.log
-rw-r--r-- 1 syslog syslog  89K Apr 25 11:11 rancher-logging-root-fluentbit-2zfr9.log
-rw-r--r-- 1 syslog syslog 176K Apr 25 11:11 longhorn-loop-device-cleaner-6dtbd.log
-rw-r--r-- 1 syslog syslog 5.3K Apr 25 11:11 system-upgrade-controller-79fc9c84b7-rthg5.log
-rw-r--r-- 1 syslog syslog  29K Apr 25 11:11 harvester-network-webhook-b5b67774b-5v88b.log
-rw-r--r-- 1 syslog syslog  201 Apr 25 11:11 rke2-multus-ds-mbj55.log
-rw-r--r-- 1 syslog syslog  24K Apr 25 11:11 kube-scheduler-harv2.log
-rw-r--r-- 1 syslog syslog 3.0K Apr 25 11:11 harvester-network-controller-manager-696cc7dd85-.log
-rw-r--r-- 1 syslog syslog 6.2K Apr 25 11:11 harvester-whereabouts-zl94c.log
-rw-r--r-- 1 syslog syslog 3.6K Apr 25 11:11 harvester-network-controller-rtc5m.log
-rw-r--r-- 1 syslog syslog 4.6K Apr 25 11:11 csi-attacher-5d4cc8cfc8-rs2cx.log

-rw-r--r-- 1 syslog syslog 5.2K Apr 25 11:04 wicked.log

-rw-r--r-- 1 syslog syslog  39K Apr 24 19:41 rancher-logging-root-fluentd-configcheck-c634418.log
drwx------ 4 syslog syslog 4.0K Apr 24 18:51 ..
```


