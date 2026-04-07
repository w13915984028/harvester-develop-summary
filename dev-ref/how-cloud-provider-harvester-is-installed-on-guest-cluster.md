# Brief

A brief introduction about how guest-cluster is built up step by step, and how those internal charts are installed onto the guest cluster.

## A sample log


Process:
```
root        5522  0.5  0.8 1293424 32716 ?       Ssl  20:52   0:01 harvester-cloud-provider --cloud-config=/etc/kubernetes/cloud-config --cluster-name=gc6


root        1914 24.0  1.9 1374936 78908 ?       Sl   20:49   2:01 containerd -c /var/lib/rancher/rke2/agent/etc/containerd/config.toml

root        1938  4.1  2.1 1314748 85404 ?       Sl   20:49   0:20 kubelet --volume-plugin-dir=/var/lib/kubelet/volumeplugins --file-check-frequency=5s --sync-frequency=30s --cloud-provider=external --cloud-config=/var/lib/rancher/rke2/etc/config-files/cloud-provider-config --config-dir=/var/lib/rancher/rke2/agent/etc/kubelet.conf.d --containerd=/run/k3s/containerd/containerd.sock --hostname-override=gc6-pool1-z4w5s-v44nc --kubeconfig=/var/lib/rancher/rke2/agent/kubelet.kubeconfig --node-ip=192.168.122.133 --node-labels=rke.cattle.io/machine=e2209f8b-71c1-4131-9fa6-513daa88500c --read-only-port=0
```


logs:


rancher-system-agent

```
sudo -i journalctl | cat | grep rancher-system-agent
Mar 03 21:00:27 gc8-pool1-q99l7-6ndw9 cloud-init[1322]: [INFO]  Downloading rancher-system-agent binary from https://192.168.122.118/assets/rancher-system-agent-amd64
Mar 03 21:00:27 gc8-pool1-q99l7-6ndw9 cloud-init[1322]: [INFO]  Successfully downloaded the rancher-system-agent binary.
Mar 03 21:00:27 gc8-pool1-q99l7-6ndw9 cloud-init[1322]: [INFO]  Downloading rancher-system-agent-uninstall.sh script from https://192.168.122.118/assets/system-agent-uninstall.sh
Mar 03 21:00:28 gc8-pool1-q99l7-6ndw9 cloud-init[1322]: [INFO]  Successfully downloaded the rancher-system-agent-uninstall.sh script.
Mar 03 21:00:28 gc8-pool1-q99l7-6ndw9 cloud-init[1322]: [INFO]  Creating environment file /etc/systemd/system/rancher-system-agent.env
Mar 03 21:00:28 gc8-pool1-q99l7-6ndw9 cloud-init[1322]: [INFO]  Enabling rancher-system-agent.service
Mar 03 21:00:28 gc8-pool1-q99l7-6ndw9 cloud-init[1322]: Created symlink /etc/systemd/system/multi-user.target.wants/rancher-system-agent.service → /etc/systemd/system/rancher-system-agent.service.
Mar 03 21:00:28 gc8-pool1-q99l7-6ndw9 cloud-init[1322]: [INFO]  Starting/restarting rancher-system-agent.service
Mar 03 21:00:28 gc8-pool1-q99l7-6ndw9 rancher-system-agent[2115]: time="2026-03-03T21:00:28Z" level=info msg="Rancher System Agent version v0.3.15-rc.3 (37f4a947f277ef8b4c275e74891d313bb9d32a17) is starting"
Mar 03 21:00:28 gc8-pool1-q99l7-6ndw9 rancher-system-agent[2115]: time="2026-03-03T21:00:28Z" level=info msg="Using directory /var/lib/rancher/agent/work for work"
Mar 03 21:00:28 gc8-pool1-q99l7-6ndw9 rancher-system-agent[2115]: time="2026-03-03T21:00:28Z" level=info msg="Starting remote watch of plans"
Mar 03 21:00:28 gc8-pool1-q99l7-6ndw9 rancher-system-agent[2115]: time="2026-03-03T21:00:28Z" level=info msg="Starting /v1, Kind=Secret controller"
Mar 03 21:00:30 gc8-pool1-q99l7-6ndw9 rancher-system-agent[2115]: time="2026-03-03T21:00:30Z" level=info msg="Detected first start, force-applying one-time instruction set"
Mar 03 21:00:30 gc8-pool1-q99l7-6ndw9 rancher-system-agent[2115]: time="2026-03-03T21:00:30Z" level=info msg="[Applyinator] Applying one-time instructions for plan with checksum b1d19c7bc15ef5b5d551e082e0746cabc41c218cae3952e0ab17c41edba515da"

...
```

```
rancher@gc8-pool1-q99l7-6ndw9:~$ sudo -i cat /etc/systemd/system/rancher-system-agent.service 
[Unit]
Description=Rancher System Agent
Documentation=https://www.rancher.com
Wants=network-online.target
After=network-online.target
[Install]
WantedBy=multi-user.target
[Service]
EnvironmentFile=-/etc/default/rancher-system-agent
EnvironmentFile=-/etc/sysconfig/rancher-system-agent
EnvironmentFile=-/etc/systemd/system/rancher-system-agent.env
Type=simple
Restart=always
RestartSec=5s
Environment=CATTLE_LOGLEVEL=info
Environment=CATTLE_AGENT_CONFIG=/etc/rancher/agent/config.yaml
Environment=CATTLE_AGENT_STRICT_VERIFY=true
ExecStart=/usr/local/bin/rancher-system-agent sentinel



root@gc8-pool1-q99l7-6ndw9:~# cat /etc/rancher/agent/config.yaml
workDirectory: /var/lib/rancher/agent/work
appliedPlanDirectory: /var/lib/rancher/agent/applied
remoteEnabled: true
localEnabled: false
localPlanDirectory: /var/lib/rancher/agent/plans
interlockDirectory: /var/lib/rancher/agent/interlock
preserveWorkDirectory: false
connectionInfoFile: /var/lib/rancher/agent/rancher2_connection_info.json
root@gc8-pool1-q99l7-6ndw9:~# 




root@gc8-pool1-q99l7-6ndw9:~# cat /var/lib/rancher/agent/applied/20260303-210
20260303-210030-applied.plan  20260303-210432-applied.plan  
root@gc8-pool1-q99l7-6ndw9:~# cat /var/lib/rancher/agent/applied/20260303-210030-applied.plan 
{"Plan":{"files":[{"content":"eyJjb25maWdzIjp7fSwibWlycm9ycyI6bnVsbH0=","path":"/etc/rancher/agent/registries.yaml"},{"content":"YXBpVmVyc2lvbjogdj...GtTZwo=","path":"/var/lib/rancher/rke2/etc/config-files/cloud-provider-config"},{"content":"eyJjb25maWdzIjp7fSwibWlycm9ycyI6bnVsbH0=","path":"/etc/rancher/rke2/registries.yaml"},{"content":"ewo...NtOW5iemIiCn0=","path":"/etc/rancher/rke2/config.yaml.d/50-rancher.yaml"},{"content":"Ci0tLQ...XN0ZXItYWdlbnQK","path":"/var/lib/rancher/rke2/server/manifests/rancher/cluster-agent.yaml"},{"content":"CmFw...RkNBQUFBPT0K","path":"/var/lib/rancher/rke2/server/manifests/rancher/rke2-etcd-snapshot-extra-metadata.yaml"},{"path":"/var/lib/rancher/rke2/server/manifests/rancher/addons.yaml"},{"content":"YXBpV...tcHQiCmZpCg==","path":"/var/lib/rancher/capr/idempotence/idempotent.sh"}],

"instructions":[{"name":"install","image":"rancher/system-agent-installer-rke2:v1.35.1-rke2r1","env":["RESTART_STAMP=ed8f3538c0c1f85792f0c12f6ba9f0b05c555fc4e718a14be3ba2ebbb99159c1","DRAIN_HASH=06b7f7ec3864c89261ce193c339dcc1526be5ccc25556650e0975c6fec074783","RKE2_DATA_DIR=/var/lib/rancher/rke2"],

"args":["-c","run.sh"],"command":"sh"}],"probes":{"calico":{"initialDelaySeconds":1,"timeoutSeconds":5,"successThreshold":1,"failureThreshold":2,"httpGet":{"url":"http://localhost:9099/liveness"}},"etcd":{"initialDelaySeconds":1,"timeoutSeconds":5,"successThreshold":1,"failureThreshold":2,"httpGet":{"url":"http://localhost:2381/health"}},"kube-apiserver":{"initialDelaySeconds":1,"timeoutSeconds":5,"successThreshold":1,"failureThreshold":2,"httpGet":{"url":"https://localhost:6443/readyz","clientCert":"/var/lib/rancher/rke2/server/tls/client-kube-apiserver.crt","clientKey":"/var/lib/rancher/rke2/server/tls/client-kube-apiserver.key","caCert":"/var/lib/rancher/rke2/server/tls/server-ca.crt"}},"kube-controller-manager":{"initialDelaySeconds":1,"timeoutSeconds":5,"successThreshold":1,"failureThreshold":2,"httpGet":{"url":"https://localhost:10257/healthz","caCert":"/var/lib/rancher/rke2/server/tls/kube-controller-manager/kube-controller-manager.crt"}},"kube-scheduler":{"initialDelaySeconds":1,"timeoutSeconds":5,"successThreshold":1,"failureThreshold":2,"httpGet":{"url":"https://localhost:10259/healthz","caCert":"/var/lib/rancher/rke2/server/tls/kube-scheduler/kube-scheduler.crt"}},"kubelet":{"initialDelaySeconds":1,"timeoutSeconds":5,"successThreshold":1,"failureThreshold":2,"httpGet":{"url":"http://localhost:10248/healthz"}}}},"Checksum":"b1d19c7bc15ef5b5d551e082e0746cabc41c218cae3952e0ab17c41edba515da"}root@gc8-pool1-q99l7-6ndw9:~# 




root@gc8-pool1-q99l7-6ndw9:~# cat /etc/rancher/rke2/config.yaml.d/50-rancher.yaml
{
  "agent-token": "cgqkh578l7x7fqv5h59ztfnsqvg8rtr45xg4b8flxlj5s4w9qc5v7s",
  "cloud-provider-config": "/var/lib/rancher/rke2/etc/config-files/cloud-provider-config",
  "cloud-provider-name": "harvester",
  "cni": "calico",
  "disable-kube-proxy": false,
  "etcd-expose-metrics": false,
  "etcd-snapshot-retention": 5,
  "etcd-snapshot-schedule-cron": "0 */5 * * *",
  "ingress-controller": "traefik",
  "kube-controller-manager-arg": [
    "cert-dir=/var/lib/rancher/rke2/server/tls/kube-controller-manager",
    "secure-port=10257"
  ],
  "kube-controller-manager-extra-mount": [
    "/var/lib/rancher/rke2/server/tls/kube-controller-manager:/var/lib/rancher/rke2/server/tls/kube-controller-manager"
  ],
  "kube-scheduler-arg": [
    "cert-dir=/var/lib/rancher/rke2/server/tls/kube-scheduler",
    "secure-port=10259"
  ],
  "kube-scheduler-extra-mount": [
    "/var/lib/rancher/rke2/server/tls/kube-scheduler:/var/lib/rancher/rke2/server/tls/kube-scheduler"
  ],
  "node-label": [
    "rke.cattle.io/machine=cce6c1d9-9c6e-4cd0-b010-c728388bc764"
  ],
  "private-registry": "/etc/rancher/rke2/registries.yaml",
  "protect-kernel-defaults": false,
  "token": "t97hdqnbqg6274c7dz9nznvxvbcjxlc454fzs5mks5crkjmcm9nbzb"
}root@gc8-pool1-q99l7-6ndw9:~# 
root@gc8-pool1-q99l7-6ndw9:~# 



root@gc8-pool1-q99l7-6ndw9:~# cat /var/lib/rancher/rke2/etc/config-files/cloud-provider-config
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJlVENDQVIrZ0F3SUJBZ0lCQURBS0JnZ3Foa2pPUFFRREFqQWtNU0l3SUFZRFZRUUREQmx5YTJVeUxYTmwKY25abGNpMWpZVUF4TnpjeU5USTJPRFEwTUI0WERUSTJNRE13TXpBNE16UXdORm9YRFRNMk1ESXlPVEE0TXpRdwpORm93SkRFaU1DQUdBMVVFQXd3WmNtdGxNaTF6WlhKMlpYSXRZMkZBTVRjM01qVXlOamcwTkRCWk1CTUdCeXFHClNNNDlBZ0VHQ0NxR1NNNDlBd0VIQTBJQUJDRExUNFNPZkJpMjFXS1Y5QmdDL09DVHpyeUNzbytMMk9tc3pCVDEKR3RzcitLQk4wNWQ4WGw5aWcwdjd0UGNLYXpCUHNXbmhZc0h2M1JFblJBVUVOd21qUWpCQU1BNEdBMVVkRHdFQgovd1FFQXdJQ3BEQVBCZ05WSFJNQkFmOEVCVEFEQVFIL01CMEdBMVVkRGdRV0JCUjU2aDNDN1N4T0s0Ti93QVdtCndIQyt5WkVjVURBS0JnZ3Foa2pPUFFRREFnTklBREJGQWlCcHpyZnhSMDFiYWl2emdiQSs5Rm9FU3IveG5lZUYKR2hkRGx5T3ZrbU51d1FJaEFLbG1SV1l2cEZWUDhBRHdhSXA5Q3lJYkFwZm9aQ25PM0JteUlWcjRNZlg3Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    server: https://192.168.122.141:6443
  name: default
contexts:
- context:
    cluster: default
    namespace: default
    user: default
  name: default
current-context: default
kind: Config
preferences: {}
users:
- name: default
  user:
    token: eyJhbGciOiJSUzI1NiIsImtpZCI6InZvVjZnZWpfQ0ZldFRDNEZqS3hVN29jX185Q0V2SXZyY2xRYlVZVDJGWWMifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImdjOC10b2tlbiIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJnYzgiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiIyZDFmNDY2Ny0yYWY3LTRlYTAtYmQxZC1jMjZhZWE3MTMyNDciLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6ZGVmYXVsdDpnYzgifQ.VWUyDPUx8VogXNbszZiVQPLopoBOsEpeRhY2-xdM210K2t5Uzjoi_7E0XAUtmnjVTF8Hhc9HGzoBuvl5wx_YHcAM9p9xArYRoUI_bVmWKzt4l2Q7_ZjY65U3kU9lGux-ilDC0hR18BzWviE7rSaurLmNtUGyn2wZ6lu7KIO1eQn81ieT76MF9ij1fpMOtDk8EUvte55D1lXZdiguxrrJBLR8kM3HUlMj1Kc-AsyhRNU4wDF8IUdr2UqLZCiGCir6yowT6PvBVWXCBXhwRhN1s84ovp4iYUz3TpLdCHAAqDPRJsr43V3t61vIhqV0OQDnU7vyu0eekfpPv2C6P_0kSg
root@gc8-pool1-q99l7-6ndw9:~# 



root@gc8-pool1-q99l7-6ndw9:~# cat /var/lib/rancher/agent/applied/20260303-210432-applied.plan 
{"Plan":{"files":[{"content":"eyJjb25maWdzIjp7fSwibWlycm9ycyI6bnVsbH0=","path":"/etc/rancher/agent/registries.yaml"},{"content":"YXBpVmVyc2lvbjogdjEKY...TZwo=","path":"/var/lib/rancher/rke2/etc/config-files/cloud-provider-config"},{"content":"eyJjb25maWdzIjp7fSwibWlycm9ycyI6bnVsbH0=","path":"/etc/rancher/rke2/registries.yaml"},{"content":"ewogICJh...W5iemIiCn0=","path":"/etc/rancher/rke2/config.yaml.d/50-rancher.yaml"},{"content":"Ci0t...ZW50Cg==","path":"/var/lib/rancher/rke2/server/manifests/rancher/cluster-agent.yaml"},{"content":"CmFwa.hczhQcGN..T0K","path":"/var/lib/rancher/rke2/server/manifests/rancher/rke2-etcd-snapshot-extra-metadata.yaml"},{"path":"/var/lib/rancher/rke2/server/manifests/rancher/addons.yaml"},{"content":"YXBp...0nCg==","path":"/var/lib/rancher/rke2/server/manifests/rancher/managed-chart-config.yaml"},

{"content":"CiMh...pCg==","path":"/var/lib/rancher/capr/idempotence/idempotent.sh"}],


"instructions":[{"name":"install","image":"rancher/system-agent-installer-rke2:v1.35.1-rke2r1","env":["RESTART_STAMP=ed8f3538c0c1f85792f0c12f6ba9f0b05c555fc4e718a14be3ba2ebbb99159c1","DRAIN_HASH=06b7f7ec3864c89261ce193c339dcc1526be5ccc25556650e0975c6fec074783","RKE2_DATA_DIR=/var/lib/rancher/rke2"],"args":["-c","run.sh"],"command":"sh"}],"probes":{"calico":{"initialDelaySeconds":1,"timeoutSeconds":5,"successThreshold":1,"failureThreshold":2,"httpGet":{"url":"http://localhost:9099/liveness"}},"etcd":{"initialDelaySeconds":1,"timeoutSeconds":5,"successThreshold":1,"failureThreshold":2,"httpGet":{"url":"http://localhost:2381/health"}},"kube-apiserver":{"initialDelaySeconds":1,"timeoutSeconds":5,"successThreshold":1,"failureThreshold":2,"httpGet":{"url":"https://localhost:6443/readyz","clientCert":"/var/lib/rancher/rke2/server/tls/client-kube-apiserver.crt","clientKey":"/var/lib/rancher/rke2/server/tls/client-kube-apiserver.key","caCert":"/var/lib/rancher/rke2/server/tls/server-ca.crt"}},"kube-controller-manager":{"initialDelaySeconds":1,"timeoutSeconds":5,"successThreshold":1,"failureThreshold":2,"httpGet":{"url":"https://localhost:10257/healthz","caCert":"/var/lib/rancher/rke2/server/tls/kube-controller-manager/kube-controller-manager.crt"}},"kube-scheduler":{"initialDelaySeconds":1,"timeoutSeconds":5,"successThreshold":1,"failureThreshold":2,"httpGet":{"url":"https://localhost:10259/healthz","caCert":"/var/lib/rancher/rke2/server/tls/kube-scheduler/kube-scheduler.crt"}},"kubelet":{"initialDelaySeconds":1,"timeoutSeconds":5,"successThreshold":1,"failureThreshold":2,"httpGet":{"url":"http://localhost:10248/healthz"}}}},"Checksum":"6109c798ee69394aedce32dd30237247366068ed65e41600bf30f1a5aec7c8c6"}root@gc8-pool1-q99l7-6ndw9:~# 




root@gc8-pool1-q99l7-6ndw9:~# cat /var/lib/rancher/rke2/server/manifests/rancher/managed-chart-config.yaml
apiVersion: helm.cattle.io/v1
kind: HelmChartConfig
metadata:
  name: rke2-ingress-nginx
  namespace: kube-system
spec:
  valuesContent: '{"global":{"cattle":{"clusterId":"c-m-6z6dj7r7"}}}'

---
apiVersion: helm.cattle.io/v1
kind: HelmChartConfig
metadata:
  name: rke2-traefik
  namespace: kube-system
spec:
  valuesContent: '{"global":{"cattle":{"clusterId":"c-m-6z6dj7r7"}}}'
root@gc8-pool1-q99l7-6ndw9:~# 


```



```
Nov 13 20:49:27 gc6-pool1-z4w5s-v44nc cloud-init[954]: [INFO]  Creating environment file /etc/systemd/system/rancher-system-agent.env
Nov 13 20:49:27 gc6-pool1-z4w5s-v44nc systemd[1]: Reloading.
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc systemd[1]: /lib/systemd/system/cloud-init-local.service:15: Unknown key name 'ConditionEnvironment' in section 'Unit', ignoring.
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc systemd[1]: /lib/systemd/system/cloud-init.service:19: Unknown key name 'ConditionEnvironment' in section 'Unit', ignoring.
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc systemd[1]: /lib/systemd/system/cloud-init.target:15: Unknown key name 'ConditionEnvironment' in section 'Unit', ignoring.
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc systemd[1]: /lib/systemd/system/cloud-final.service:9: Unknown key name 'ConditionEnvironment' in section 'Unit', ignoring.
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc systemd[1]: /lib/systemd/system/cloud-config.service:8: Unknown key name 'ConditionEnvironment' in section 'Unit', ignoring.

Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc cloud-init[954]: [INFO]  Enabling rancher-system-agent.service
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc cloud-init[954]: Created symlink /etc/systemd/system/multi-user.target.wants/rancher-system-agent.service → /etc/systemd/system/rancher-system-agent.service.

Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc systemd[1]: Reloading.
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc systemd[1]: /lib/systemd/system/cloud-init-local.service:15: Unknown key name 'ConditionEnvironment' in section 'Unit', ignoring.
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc systemd[1]: /lib/systemd/system/cloud-init.service:19: Unknown key name 'ConditionEnvironment' in section 'Unit', ignoring.
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc systemd[1]: /lib/systemd/system/cloud-init.target:15: Unknown key name 'ConditionEnvironment' in section 'Unit', ignoring.
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc systemd[1]: /lib/systemd/system/cloud-final.service:9: Unknown key name 'ConditionEnvironment' in section 'Unit', ignoring.
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc systemd[1]: /lib/systemd/system/cloud-config.service:8: Unknown key name 'ConditionEnvironment' in section 'Unit', ignoring.
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc cloud-init[954]: [INFO]  Starting/restarting rancher-system-agent.service

Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc systemd[1]: Started Rancher System Agent.
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc cloud-init[1704]: #############################################################
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc cloud-init[1709]: -----BEGIN SSH HOST KEY FINGERPRINTS-----
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc cloud-init[1727]: 256 SHA256:/OC+7ejprKVFCKs5H2dsrGsleRKXJKN97L5i3ELVVCQ root@gc6-pool1-z4w5s-v44nc (ECDSA)
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc cloud-init[1729]: 256 SHA256:CBsAxulNuTmCTqEMKa7M24xe9fRjsBXAhF19UwP08g0 root@gc6-pool1-z4w5s-v44nc (ED25519)
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc rancher-system-agent[1698]: time="2025-11-13T20:49:28Z" level=info msg="Rancher System Agent version v0.3.13 (5a64be2) is starting"
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc rancher-system-agent[1698]: time="2025-11-13T20:49:28Z" level=info msg="Using directory /var/lib/rancher/agent/work for work"
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc rancher-system-agent[1698]: time="2025-11-13T20:49:28Z" level=info msg="Starting remote watch of plans"
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc cloud-init[1733]: 3072 SHA256:KQlysmVPtQ4ldeIGlr4b8pkSYRcxLcqtkbx8i6SOLqc root@gc6-pool1-z4w5s-v44nc (RSA)
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc cloud-init[1734]: -----END SSH HOST KEY FINGERPRINTS-----
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc cloud-init[1736]: #############################################################
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc cloud-init[954]: Cloud-init v. 24.1.3-0ubuntu1~20.04.1 running 'modules:final' at Thu, 13 Nov 2025 20:49:11 +0000. Up 20.85 seconds.
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc cloud-init[954]: Cloud-init v. 24.1.3-0ubuntu1~20.04.1 finished at Thu, 13 Nov 2025 20:49:28 +0000. Datasource DataSourceNoCloud [seed=/dev/vda][dsmode=net].  Up 37.89 seconds
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc systemd[1]: Finished Execute cloud user/final scripts.

Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc systemd[1]: Reached target Cloud-init target.
Nov 13 20:49:28 gc6-pool1-z4w5s-v44nc systemd[1]: Startup finished in 7.530s (kernel) + 30.482s (userspace) = 38.013s.
No


Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rancher-system-agent[1698]: time="2025-11-13T20:49:35Z" level=info msg="[88aaada2e4e2c7c6803b3ffe0c9976fe8aca16e41f74cade5d9f5bd90ce3a8e1_0:stderr]: Created symlink /etc/systemd/system/multi-user.target.wants/rke2-server.service → /usr/local/lib/systemd/system/rke2-server.service."

Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc systemd[1]: Reloading.
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc systemd[1]: /lib/systemd/system/cloud-init-local.service:15: Unknown key name 'ConditionEnvironment' in section 'Unit', ignoring.
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc systemd[1]: /lib/systemd/system/cloud-init.service:19: Unknown key name 'ConditionEnvironment' in section 'Unit', ignoring.
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc systemd[1]: /lib/systemd/system/cloud-init.target:15: Unknown key name 'ConditionEnvironment' in section 'Unit', ignoring.
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc systemd[1]: /lib/systemd/system/cloud-final.service:9: Unknown key name 'ConditionEnvironment' in section 'Unit', ignoring.
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc systemd[1]: /lib/systemd/system/cloud-config.service:8: Unknown key name 'ConditionEnvironment' in section 'Unit', ignoring.
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rancher-system-agent[1698]: time="2025-11-13T20:49:35Z" level=info msg="[88aaada2e4e2c7c6803b3ffe0c9976fe8aca16e41f74cade5d9f5bd90ce3a8e1_0:stderr]: + [  = true ]"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rancher-system-agent[1698]: time="2025-11-13T20:49:35Z" level=info msg="[88aaada2e4e2c7c6803b3ffe0c9976fe8aca16e41f74cade5d9f5bd90ce3a8e1_0:stderr]: + [ true = true ]"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rancher-system-agent[1698]: time="2025-11-13T20:49:35Z" level=info msg="[88aaada2e4e2c7c6803b3ffe0c9976fe8aca16e41f74cade5d9f5bd90ce3a8e1_0:stderr]: + systemctl --no-block restart rke2-server"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc systemd[1]: Starting Rancher Kubernetes Engine v2 (server)...
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc sh[1868]: + /usr/bin/systemctl is-enabled --quiet nm-cloud-setup.service
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc sh[1869]: Failed to get unit file state for nm-cloud-setup.service: No such file or directory
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rancher-system-agent[1698]: time="2025-11-13T20:49:35Z" level=info msg="[Applyinator] Command sh [-c run.sh] finished with err: <nil> and exit code: 0"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc kernel: bridge: filtering via arp/ip/ip6tables is no longer available by default. Update your scripts to load br_netfilter if you need this.
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc kernel: Bridge firewalling registered
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=warning msg="not running in CIS mode"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="Applying Pod Security Admission Configuration"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="Starting rke2 v1.33.5+rke2r1 (d1092839cf08cb901b1d40461b0fa6e7ae6f8fc4)"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="Managed etcd cluster initializing"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="generated self-signed CA certificate CN=rke2-client-ca@1763066975: notBefore=2025-11-13 20:49:35.732592387 +0000 UTC notAfter=2035-11-11 20:49:35.732592387 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="certificate CN=system:admin,O=system:masters signed by CN=rke2-client-ca@1763066975: notBefore=2025-11-13 20:49:35 +0000 UTC notAfter=2026-11-13 20:49:35 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="certificate CN=system:rke2-supervisor,O=system:masters signed by CN=rke2-client-ca@1763066975: notBefore=2025-11-13 20:49:35 +0000 UTC notAfter=2026-11-13 20:49:35 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="certificate CN=system:kube-controller-manager signed by CN=rke2-client-ca@1763066975: notBefore=2025-11-13 20:49:35 +0000 UTC notAfter=2026-11-13 20:49:35 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="certificate CN=system:kube-scheduler signed by CN=rke2-client-ca@1763066975: notBefore=2025-11-13 20:49:35 +0000 UTC notAfter=2026-11-13 20:49:35 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="certificate CN=system:apiserver,O=system:masters signed by CN=rke2-client-ca@1763066975: notBefore=2025-11-13 20:49:35 +0000 UTC notAfter=2026-11-13 20:49:35 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="certificate CN=rke2-cloud-controller-manager signed by CN=rke2-client-ca@1763066975: notBefore=2025-11-13 20:49:35 +0000 UTC notAfter=2026-11-13 20:49:35 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="generated self-signed CA certificate CN=rke2-server-ca@1763066975: notBefore=2025-11-13 20:49:35.736229738 +0000 UTC notAfter=2035-11-11 20:49:35.736229738 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="certificate CN=kube-apiserver signed by CN=rke2-server-ca@1763066975: notBefore=2025-11-13 20:49:35 +0000 UTC notAfter=2026-11-13 20:49:35 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="certificate CN=kube-scheduler signed by CN=rke2-server-ca@1763066975: notBefore=2025-11-13 20:49:35 +0000 UTC notAfter=2026-11-13 20:49:35 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="certificate CN=kube-controller-manager signed by CN=rke2-server-ca@1763066975: notBefore=2025-11-13 20:49:35 +0000 UTC notAfter=2026-11-13 20:49:35 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="generated self-signed CA certificate CN=rke2-request-header-ca@1763066975: notBefore=2025-11-13 20:49:35.738179935 +0000 UTC notAfter=2035-11-11 20:49:35.738179935 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="certificate CN=system:auth-proxy signed by CN=rke2-request-header-ca@1763066975: notBefore=2025-11-13 20:49:35 +0000 UTC notAfter=2026-11-13 20:49:35 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="generated self-signed CA certificate CN=etcd-server-ca@1763066975: notBefore=2025-11-13 20:49:35.73896141 +0000 UTC notAfter=2035-11-11 20:49:35.73896141 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="certificate CN=etcd-client signed by CN=etcd-server-ca@1763066975: notBefore=2025-11-13 20:49:35 +0000 UTC notAfter=2026-11-13 20:49:35 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="generated self-signed CA certificate CN=etcd-peer-ca@1763066975: notBefore=2025-11-13 20:49:35.739760879 +0000 UTC notAfter=2035-11-11 20:49:35.739760879 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="certificate CN=etcd-peer signed by CN=etcd-peer-ca@1763066975: notBefore=2025-11-13 20:49:35 +0000 UTC notAfter=2026-11-13 20:49:35 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="certificate CN=etcd-server signed by CN=etcd-server-ca@1763066975: notBefore=2025-11-13 20:49:35 +0000 UTC notAfter=2026-11-13 20:49:35 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="certificate CN=rke2,O=rke2 signed by CN=rke2-server-ca@1763066975: notBefore=2025-11-13 20:49:35 +0000 UTC notAfter=2026-11-13 20:49:35 +0000 UTC"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=warning msg="dynamiclistener [::]:9345: no cached certificate available for preload - deferring certificate load until storage initialization or first client request"
Nov 13 20:49:35 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:35Z" level=info msg="Active TLS secret / (ver=) (count 10): map[listener.cattle.io/cn-10.43.0.1:10.43.0.1 listener.cattle.io/cn-127.0.0.1:127.0.0.1 listener.cattle.io/cn-192.168.122.133:192.168.122.133 listener.cattle.io/cn-__1-f16284:::1 listener.cattle.io/cn-gc6-pool1-z4w5s-v44nc:gc6-pool1-z4w5s-v44nc listener.cattle.io/cn-kubernetes:kubernetes listener.cattle.io/cn-kubernetes.default:kubernetes.default listener.cattle.io/cn-kubernetes.default.svc:kubernetes.default.svc listener.cattle.io/cn-kubernetes.default.svc.cluster.local:kubernetes.default.svc.cluster.local listener.cattle.io/cn-localhost:localhost listener.cattle.io/fingerprint:SHA1=77BA6A1CE1B88985E157FA709F2F540A9D7EA865]"
Nov 13 20:49:37 gc6-pool1-z4w5s-v44nc systemd-timesyncd[553]: Initial synchronization to time server 91.189.91.157:123 (ntp.ubuntu.com).
Nov 13 20:49:37 gc6-pool1-z4w5s-v44nc rancher-system-agent[1698]: time="2025-11-13T20:49:37Z" level=info msg="[K8s] updated plan secret fleet-default/gc6-pool1-z4w5s-v44nc-machine-plan with feedback"
Nov 13 20:49:37 gc6-pool1-z4w5s-v44nc rancher-system-agent[1698]: time="2025-11-13T20:49:37Z" level=info msg="[K8s] updated plan secret fleet-default/gc6-pool1-z4w5s-v44nc-machine-plan with feedback"
Nov 13 20:49:37 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:37Z" level=info msg="Password verified locally for node gc6-pool1-z4w5s-v44nc"
Nov 13 20:49:37 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:37Z" level=info msg="certificate CN=gc6-pool1-z4w5s-v44nc signed by CN=rke2-server-ca@1763066975: notBefore=2025-11-13 20:49:35 +0000 UTC notAfter=2026-11-13 20:49:37 +0000 UTC"
Nov 13 20:49:38 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:38Z" level=info msg="certificate CN=system:node:gc6-pool1-z4w5s-v44nc,O=system:nodes signed by CN=rke2-client-ca@1763066975: notBefore=2025-11-13 20:49:35 +0000 UTC notAfter=2026-11-13 20:49:38 +0000 UTC"
Nov 13 20:49:38 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:38Z" level=info msg="certificate CN=system:kube-proxy signed by CN=rke2-client-ca@1763066975: notBefore=2025-11-13 20:49:35 +0000 UTC notAfter=2026-11-13 20:49:38 +0000 UTC"
Nov 13 20:49:38 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:38Z" level=info msg="certificate CN=system:rke2-controller signed by CN=rke2-client-ca@1763066975: notBefore=2025-11-13 20:49:35 +0000 UTC notAfter=2026-11-13 20:49:38 +0000 UTC"
Nov 13 20:49:39 gc6-pool1-z4w5s-v44nc systemd[1]: systemd-hostnamed.service: Succeeded.
Nov 13 20:49:40 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:40Z" level=info msg="Using private registry config file at /etc/rancher/rke2/registries.yaml"
Nov 13 20:49:40 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:40Z" level=info msg="Module overlay was already loaded"
Nov 13 20:49:40 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:40Z" level=info msg="Module br_netfilter was already loaded"
Nov 13 20:49:40 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:40Z" level=info msg="Set sysctl 'net/ipv4/conf/all/forwarding' to 1"
Nov 13 20:49:40 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:40Z" level=info msg="Set sysctl 'net/netfilter/nf_conntrack_max' to 131072"
Nov 13 20:49:40 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:40Z" level=info msg="Set sysctl 'net/netfilter/nf_conntrack_tcp_timeout_established' to 86400"
Nov 13 20:49:40 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:40Z" level=info msg="Set sysctl 'net/netfilter/nf_conntrack_tcp_timeout_close_wait' to 3600"
Nov 13 20:49:40 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:40Z" level=warning msg="Failed to load runtime image index.docker.io/rancher/rke2-runtime:v1.33.5-rke2r1 from tarball: no local image available for index.docker.io/rancher/rke2-runtime:v1.33.5-rke2r1: directory /var/lib/rancher/rke2/agent/images does not exist: image not found"
Nov 13 20:49:40 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:40Z" level=warning msg="Failed to load runtime image index.docker.io/rancher/rke2-runtime:v1.33.5-rke2r1 from tarball: no local image available for index.docker.io/rancher/rke2-runtime:v1.33.5-rke2r1: directory /var/lib/rancher/rke2/agent/images does not exist: image not found"
Nov 13 20:49:40 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:40Z" level=info msg="Using private registry config file at /etc/rancher/rke2/registries.yaml"
Nov 13 20:49:40 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:40Z" level=info msg="Pulling runtime image index.docker.io/rancher/rke2-runtime:v1.33.5-rke2r1"
Nov 13 20:49:40 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:40Z" level=info msg="Waiting for cri connection: rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing: dial unix /run/k3s/containerd/containerd.sock: connect: no such file or directory\""
Nov 13 20:49:41 gc6-pool1-z4w5s-v44nc systemd[1]: systemd-timedated.service: Succeeded.


Nov 13 20:49:41 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:41Z" level=info msg="Creating directory /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/bin"
Nov 13 20:49:41 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:41Z" level=info msg="Extracting file bin/containerd to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/bin/containerd"
Nov 13 20:49:43 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:43Z" level=info msg="Extracting file bin/containerd-shim-runc-v2 to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/bin/containerd-shim-runc-v2"
Nov 13 20:49:44 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:44Z" level=info msg="Extracting file bin/crictl to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/bin/crictl"
Nov 13 20:49:44 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:44Z" level=info msg="Extracting file bin/ctr to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/bin/ctr"
Nov 13 20:49:45 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:45Z" level=info msg="Extracting file bin/kubectl to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/bin/kubectl"
Nov 13 20:49:46 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:46Z" level=info msg="Extracting file bin/kubelet to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/bin/kubelet"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file bin/runc to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/bin/runc"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Creating directory /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts"

Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/harvester-cloud-provider.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/harvester-cloud-provider.yaml"


Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/harvester-csi-driver.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/harvester-csi-driver.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/rancher-vsphere-cpi.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/rancher-vsphere-cpi.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/rancher-vsphere-csi.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/rancher-vsphere-csi.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/rke2-calico-crd.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/rke2-calico-crd.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/rke2-calico.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/rke2-calico.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/rke2-canal.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/rke2-canal.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/rke2-cilium.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/rke2-cilium.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/rke2-coredns.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/rke2-coredns.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/rke2-flannel.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/rke2-flannel.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/rke2-ingress-nginx.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/rke2-ingress-nginx.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/rke2-metrics-server.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/rke2-metrics-server.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/rke2-multus.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/rke2-multus.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/rke2-runtimeclasses.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/rke2-runtimeclasses.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/rke2-snapshot-controller-crd.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/rke2-snapshot-controller-crd.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/rke2-snapshot-controller.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/rke2-snapshot-controller.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/rke2-snapshot-validation-webhook.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/rke2-snapshot-validation-webhook.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/rke2-traefik-crd.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/rke2-traefik-crd.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Extracting file charts/rke2-traefik.yaml to /var/lib/rancher/rke2/data/v1.33.5-rke2r1-522501607126/charts/rke2-traefik.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="No cluster configuration value changes necessary for manifest /var/lib/rancher/rke2/server/manifests/rancher/rke2-etcd-snapshot-extra-metadata.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Updated manifest /var/lib/rancher/rke2/server/manifests/rancher-vsphere-csi.yaml to set cluster configuration values"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Updated manifest /var/lib/rancher/rke2/server/manifests/rke2-calico.yaml to set cluster configuration values"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Updated manifest /var/lib/rancher/rke2/server/manifests/rke2-flannel.yaml to set cluster configuration values"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Updated manifest /var/lib/rancher/rke2/server/manifests/rke2-runtimeclasses.yaml to set cluster configuration values"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Updated manifest /var/lib/rancher/rke2/server/manifests/rke2-snapshot-controller-crd.yaml to set cluster configuration values"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Updated manifest /var/lib/rancher/rke2/server/manifests/rke2-traefik-crd.yaml to set cluster configuration values"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Updated manifest /var/lib/rancher/rke2/server/manifests/rke2-traefik.yaml to set cluster configuration values"


Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Updated manifest /var/lib/rancher/rke2/server/manifests/harvester-cloud-provider.yaml to set cluster configuration values"


Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="No cluster configuration value changes necessary for manifest /var/lib/rancher/rke2/server/manifests/rancher/cluster-agent.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Updated manifest /var/lib/rancher/rke2/server/manifests/rke2-calico-crd.yaml to set cluster configuration values"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="No cluster configuration value changes necessary for manifest /var/lib/rancher/rke2/server/manifests/rke2-snapshot-validation-webhook.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="No cluster configuration value changes necessary for manifest /var/lib/rancher/rke2/server/manifests/rancher/addons.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Updated manifest /var/lib/rancher/rke2/server/manifests/rke2-canal.yaml to set cluster configuration values"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Updated manifest /var/lib/rancher/rke2/server/manifests/rke2-snapshot-controller.yaml to set cluster configuration values"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Updated manifest /var/lib/rancher/rke2/server/manifests/harvester-csi-driver.yaml to set cluster configuration values"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="No cluster configuration value changes necessary for manifest /var/lib/rancher/rke2/server/manifests/rancher/managed-chart-config.yaml"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Updated manifest /var/lib/rancher/rke2/server/manifests/rancher-vsphere-cpi.yaml to set cluster configuration values"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Updated manifest /var/lib/rancher/rke2/server/manifests/rke2-cilium.yaml to set cluster configuration values"
Nov 13 20:49:48 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:48Z" level=info msg="Updated manifest /var/lib/rancher/rke2/server/manifests/rke2-coredns.yaml to set cluster configuration values"
Nov 13 20:49:49 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:49Z" level=info msg="Updated manifest /var/lib/rancher/rke2/server/manifests/rke2-ingress-nginx.yaml to set cluster configuration values"
Nov 13 20:49:49 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:49Z" level=info msg="Updated manifest /var/lib/rancher/rke2/server/manifests/rke2-metrics-server.yaml to set cluster configuration values"
Nov 13 20:49:49 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:49Z" level=info msg="Updated manifest /var/lib/rancher/rke2/server/manifests/rke2-multus.yaml to set cluster configuration values"
Nov 13 20:49:49 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:49Z" level=info msg="Logging containerd to /var/lib/rancher/rke2/agent/containerd/containerd.log"
Nov 13 20:49:49 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:49Z" level=info msg="Running containerd -c /var/lib/rancher/rke2/agent/etc/containerd/config.toml"
Nov 13 20:49:50 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:50Z" level=info msg="containerd is now running"
Nov 13 20:49:50 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:50Z" level=info msg="Pulling images from /var/lib/rancher/rke2/agent/images/runtime-image.txt"
Nov 13 20:49:50 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:50Z" level=info msg="Pulling image index.docker.io/rancher/rke2-runtime:v1.33.5-rke2r1"
Nov 13 20:49:52 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:52Z" level=info msg="Connecting to proxy" url="wss://127.0.0.1:9345/v1-rke2/connect"
Nov 13 20:49:52 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:52Z" level=info msg="Creating rke2-cert-monitor event broadcaster"
Nov 13 20:49:52 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:52Z" level=info msg="Handling backend connection request [gc6-pool1-z4w5s-v44nc]"
Nov 13 20:49:52 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:52Z" level=info msg="Connected to proxy" url="wss://127.0.0.1:9345/v1-rke2/connect"
Nov 13 20:49:52 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:52Z" level=info msg="Remotedialer connected to proxy" url="wss://127.0.0.1:9345/v1-rke2/connect"
Nov 13 20:49:52 gc6-pool1-z4w5s-v44nc rke2[1892]: time="2025-11-13T20:49:52Z" level=info msg="Running kubelet --alsologtostderr=false --config-dir=/var/lib/rancher/rke2/agent/etc/kubelet.conf.d --containerd=/run/k3s/containerd/containerd.sock --hostname-override=gc6-pool1-z4w5s-v44nc --kubeconfig=/var/lib/rancher/rke2/agent/kubelet.kubeconfig --log-file=/var/lib/rancher/rke2/agent/logs/kubelet.log --log-file-max-size=50 --logtostderr=false --node-ip=192.168.122.133 --node-labels=rke.cattle.io/machine=e2209f8b-71c1-4131-9fa6-513daa88500c --read-only-port=0 --stderrthreshold=FATAL"
```


## Analysis

### Harvester side objects

clusterrole:

```yaml
harv31:/var/lib/rancher/rke2 # kk get clusterrole -A | grep cloud
harvesterhci.io:cloudprovider                                          2026-03-31T10:46:29Z


harv31:/var/lib/rancher/rke2 # kk get clusterrole harvesterhci.io:cloudprovider
NAME                            CREATED AT
harvesterhci.io:cloudprovider   2026-03-31T10:46:29Z
harv31:/var/lib/rancher/rke2 # kk get clusterrole harvesterhci.io:cloudprovider -oyaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  annotations:
    meta.helm.sh/release-name: harvester
    meta.helm.sh/release-namespace: harvester-system
    objectset.rio.cattle.io/id: default-mcc-harvester-cattle-fleet-local-system
  creationTimestamp: "2026-03-31T10:46:29Z"
  labels:
    app.kubernetes.io/component: apiserver
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: harvester
    app.kubernetes.io/part-of: harvester
    app.kubernetes.io/version: master-fdca3882
    helm.sh/chart: harvester-0.0.0-master-fdca3882
    helm.sh/release: harvester
    objectset.rio.cattle.io/hash: e852fa897f5eae59a44b4bfe186aad80b10b94b3
  name: harvesterhci.io:cloudprovider
  resourceVersion: "3904"
  uid: db20e384-bb09-4c86-9343-f01a3e4e44b7
rules:
- apiGroups:
  - loadbalancer.harvesterhci.io
  resources:
  - loadbalancers
  verbs:
  - get
  - list
  - watch
  - update
  - create
  - delete
- apiGroups:
  - subresources.kubevirt.io
  resources:
  - virtualmachines/addvolume
  - virtualmachines/removevolume
  - virtualmachineinstances/addvolume
  - virtualmachineinstances/removevolume
  - virtualmachineinstances/guestosinfo
  verbs:
  - get
  - update
- apiGroups:
  - kubevirt.io
  resources:
  - virtualmachines
  - virtualmachineinstances
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - persistentvolumeclaims
  - persistentvolumeclaims/status
  verbs:
  - '*'
```

RBAC

```yaml
harv31:/var/lib/rancher/rke2 # kk get serviceaccount
NAME      AGE
default   7d1h
gc2       57m
gc3       16m


harv31:/var/lib/rancher/rke2 # kk get serviceaccount gc2 -oyaml
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: "2026-04-07T11:25:43Z"
  name: gc2
  namespace: default
  resourceVersion: "587443"
  uid: 436bff7b-e9de-4c04-a5f0-ec7e77e03696


harv31:/var/lib/rancher/rke2 # kk get rolebinding 
NAME            ROLE                                        AGE
default-gc2     ClusterRole/harvesterhci.io:cloudprovider   60m
default-gc3     ClusterRole/harvesterhci.io:cloudprovider   19m
rb-tebos2jzo5   ClusterRole/project-owner                   7d1h
rb-x5ycffqkie   ClusterRole/admin                           7d1h


kk get rolebinding  default-gc2 -oyaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  creationTimestamp: "2026-04-07T11:25:44Z"
  name: default-gc2
  namespace: default
  ownerReferences:
  - apiVersion: v1
    kind: ServiceAccount
    name: gc2
    uid: 436bff7b-e9de-4c04-a5f0-ec7e77e03696
  resourceVersion: "587449"
  uid: 860b756a-c7a7-411b-8b2e-a98e8be03c48
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: harvesterhci.io:cloudprovider
subjects:
- kind: ServiceAccount
  name: gc2
  namespace: default
```

secret:

```yaml

kk get secret
NAME                              TYPE                                  DATA   AGE
gc2-pool1-8kfhg-rff6b-cloudinit   Opaque                                1      65m
gc2-token                         kubernetes.io/service-account-token   3      65m
gc3-pool1-d6qfl-wskdm-cloudinit   Opaque                                1      24m
gc3-token                         kubernetes.io/service-account-token   3      24m


harv31:/var/lib/rancher/rke2 # kk get secret gc2-token -oyaml
apiVersion: v1
data:
  ca.crt: LS0tLS1C......LS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
  namespace: ZGVmYXVsdA==
  token: ZXlKaGJH...2h3
kind: Secret
metadata:
  annotations:
    kubernetes.io/service-account.name: gc2
    kubernetes.io/service-account.uid: 436bff7b-e9de-4c04-a5f0-ec7e77e03696
  creationTimestamp: "2026-04-07T11:25:43Z"
  labels:
    kubernetes.io/legacy-token-last-used: "2026-04-07"
  name: gc2-token
  namespace: default
  ownerReferences:
  - apiVersion: v1
    kind: ServiceAccount
    name: gc2
    uid: 436bff7b-e9de-4c04-a5f0-ec7e77e03696
  resourceVersion: "594771"
  uid: 483e83df-00b9-415b-aa76-b2d23abdbfc2
type: kubernetes.io/service-account-token
harv31:/var/lib/rancher/rke2 # 
harv31:/var/lib/rancher/rke2 # kk get secret gc2-pool1-8kfhg-rff6b-cloudinit
NAME                              TYPE     DATA   AGE
gc2-pool1-8kfhg-rff6b-cloudinit   Opaque   1      65m
harv31:/var/lib/rancher/rke2 # kk get secret gc2-pool1-8kfhg-rff6b-cloudinit -oyaml
apiVersion: v1
data:
  userdata: I2Nsb3VkLWNvbmZ...
kind: Secret
metadata:
  creationTimestamp: "2026-04-07T11:25:58Z"
  name: gc2-pool1-8kfhg-rff6b-cloudinit
  namespace: default
  ownerReferences:
  - apiVersion: kubevirt.io/v1
    kind: VirtualMachine
    name: gc2-pool1-8kfhg-rff6b
    uid: 4005e84d-3fcf-47de-9be4-5e9af194e96e
  resourceVersion: "587638"
  uid: afbff5e0-cd49-4301-8acd-c2965a3442e2
type: Opaque


```


### cluster-name injected from Rancher UI

https://github.com/rancher/dashboard/blob/eabde495ba43942ba5b571a3001d2670da3ad0fa/shell/edit/provisioning.cattle.io.cluster/rke2.vue#L1812-L1814

```yaml
Rancher Manager > kubectl get secret -n fleet-default
NAME                                                              TYPE                                          DATA   AGE
51-kubeconfig                                                     Opaque                                        2      6m8s
51-pool1-gbxrc-kc8cx-machine-bootstrap                            rke.cattle.io/bootstrap                       1      7m35s
51-pool1-gbxrc-kc8cx-machine-bootstrap-token-nccxm                kubernetes.io/service-account-token           3      7m35s
51-pool1-gbxrc-kc8cx-machine-driver-secret                        Opaque                                        3      7m25s
51-pool1-gbxrc-kc8cx-machine-plan                                 rke.cattle.io/machine-plan                    13     7m35s
51-pool1-gbxrc-kc8cx-machine-plan-token-l7z74                     kubernetes.io/service-account-token           3      6m11s
51-pool1-gbxrc-kc8cx-machine-state                                rke.cattle.io/machine-state                   1      7m25s
51-rke-state                                                      rke.cattle.io/cluster-state                   2      7m35s
c-hxmz2-kubeconfig                                                Opaque                                        2      18m
harvesterconfigndk26                                              secret                                        1      7m36s
import-token-c-hxmz2                                              fleet.cattle.io/cluster-registration-values   1      18m
import-token-c-hxmz2-eea71d39-05eb-44de-a31d-43abbd40e382-token   kubernetes.io/service-account-token           3      18m
machine-certs-56ae28089bc6c16b41a5dbc58549305b                    Opaque                                        2      7m25s

> kubectl get secret -n fleet-default harvesterconfigndk26 -oyaml
apiVersion: v1
data:
  credential: YXBpVmVyc2lvbjogdjEKY2...bk9nCg==
kind: Secret
metadata:
  annotations:
    v2prov-authorized-secret-deletes-on-cluster-removal: "true"
    v2prov-secret-authorized-for-cluster: "51"
  creationTimestamp: "2026-03-03T22:05:56Z"
  generateName: harvesterconfig
  name: harvesterconfigndk26
  namespace: fleet-default
  resourceVersion: "25260"
  uid: e431aa39-a78c-4e67-8dac-d864a5ca3776
type: secret
> 

> kubectl get harvesterconfig -n fleet-default nc-gc3-pool1-48qkw -oyaml
apiVersion: rke-machine-config.cattle.io/v1
cloudConfig: ""
clusterId: ""
clusterName: ""
clusterType: ""
cpuCount: "2"
cpuModel: ""
cpuPinning: false
diskBus: ""
diskInfo: '{"disks":[{"imageName":"default/image-pmtjz","bootOrder":1,"size":40}]}'
diskSize: "0"
enableEfi: false
enableSecureBoot: false
enableTpm: false
imageName: ""
isolateEmulatorThread: false
keyPairName: ""
kind: HarvesterConfig
kubeconfigContent: ""
memorySize: "4"
metadata:
  annotations:
    field.cattle.io/creatorId: user-m8pjn
    ownerBindingsCreated: "true"
  creationTimestamp: "2026-04-07T12:06:54Z"
  generateName: nc-gc3-pool1-
  generation: 1
  name: nc-gc3-pool1-48qkw
  namespace: fleet-default
  ownerReferences:
  - apiVersion: provisioning.cattle.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: Cluster
    name: gc3
    uid: 829d1e3c-2fa2-4231-b7e0-c03c5f916291
  resourceVersion: "260587"
  uid: 58a479d2-7f81-42e1-bd2f-ee75156e0cbf
networkData: ""
networkInfo: '{"interfaces":[{"networkName":"default/vm-untag","macAddress":""}]}'
networkModel: ""
networkName: ""
networkType: ""
reservedMemorySize: "-1"
sshPassword: ""
sshPort: "22"
sshPrivateKeyPath: ""
sshUser: rancher
userData: I2Nsb3VkLWNvbmZpZwpwYWNrYWdl...ZS5jb20=
vgpuInfo: ""
vmAffinity: ""
vmNamespace: default
> 

```



### Rancher provisioning object

```
> kubectl get provisioning -A          
NAMESPACE       NAME      VERSION          CLUSTER NAME   AGE     KUBECONFIG           READY
fleet-default   51        v1.35.1+rke2r1   c-m-nl2hgfgg   6m12s                        
fleet-default   c-hxmz2                    c-hxmz2        17m     c-hxmz2-kubeconfig   true
fleet-local     local                      local          9h      local-kubeconfig     true
> 
> 
> kubectl get provisioning 51
Error from server (NotFound): clusters.provisioning.cattle.io "51" not found
> kubectl get provisioning -n fleet-default 51
NAME   VERSION          CLUSTER NAME   AGE     KUBECONFIG   READY
51     v1.35.1+rke2r1   c-m-nl2hgfgg   6m25s     

           
> kubectl get provisioning -n fleet-default 51 -oyaml
apiVersion: provisioning.cattle.io/v1
kind: Cluster
metadata:
  annotations:
    field.cattle.io/creatorId: user-mgz7x
    provisioning.cattle.io/management-cluster-display-name: "51"
  creationTimestamp: "2026-03-03T22:05:56Z"
  finalizers:
  - wrangler.cattle.io/provisioning-cluster-remove
  - wrangler.cattle.io/rke-cluster-remove
  - wrangler.cattle.io/cloud-config-secret-remover
  generation: 3
  name: "51"
  namespace: fleet-default
  resourceVersion: "26986"
  uid: f528e3d3-e723-4214-bc3d-70527584249a
spec:
  cloudCredentialSecretName: cattle-global-data:cc-4dv6m
  kubernetesVersion: v1.35.1+rke2r1
  localClusterAuthEndpoint: {}
  rkeConfig:
    chartValues:
      harvester-cloud-provider:
        cloudConfigPath: /var/lib/rancher/rke2/etc/config-files/cloud-provider-config
        global:
          cattle:
            clusterName: "51"
      rke2-calico: {}
    dataDirectories: {}
    etcd:
      snapshotRetention: 5
      snapshotScheduleCron: 0 */5 * * *
    machineGlobalConfig:
      cni: calico
      disable-kube-proxy: false
      etcd-expose-metrics: false
      ingress-controller: ingress-nginx
    machinePoolDefaults: {}
    machinePools:
    - controlPlaneRole: true
      drainBeforeDelete: true
      dynamicSchemaSpec: '{"resourceFields":{"cloudConfig":{"type":"string","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"just
        keep it empty, this value will be filled by rancher-machine"},"clusterId":{"type":"string","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"harvester
        cluster id"},"clusterName":{"type":"string","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"harvester
        cluster name"},"clusterType":{"type":"string","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"harvester
        cluster type"},"cpuCount":{"type":"string","default":{"stringValue":"2","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"number
        of CPUs for machine"},"cpuPinning":{"type":"boolean","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"enable
        vm cpu pinning, please ensure the harvester cluster has cpu manager enabled
        in at least one node"},"diskBus":{"type":"string","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"bus
        of disk for machine"},"diskInfo":{"type":"string","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"harvester
        disk info"},"diskSize":{"type":"string","default":{"stringValue":"0","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"size
        of disk for machine (in GiB)"},"enableEfi":{"type":"boolean","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"enable
        vm efi"},"enableSecureBoot":{"type":"boolean","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"enable
        vm secure boot, only works when enable efi"},"enableTpm":{"type":"boolean","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"enable
        vm TPM"},"imageName":{"type":"string","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"harvester
        image name"},"isolateEmulatorThread":{"type":"boolean","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"enable
        vm isolatated emulator thread"},"keyPairName":{"type":"string","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"harvester
        key pair name"},"kubeconfigContent":{"type":"password","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"contents
        of kubeconfig file for harvester cluster, base64 is supported"},"memorySize":{"type":"string","default":{"stringValue":"4","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"size
        of memory for machine (in GiB)"},"networkData":{"type":"string","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"networkData
        content of cloud-init for machine, base64 is supported"},"networkInfo":{"type":"string","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"harvester
        network info"},"networkModel":{"type":"string","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"harvester
        network model"},"networkName":{"type":"string","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"harvester
        network name"},"networkType":{"type":"string","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"harvester
        network type"},"reservedMemorySize":{"type":"string","default":{"stringValue":"-1","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"size
        of reserved memory for machine (in MiB, integer value)"},"sshPassword":{"type":"string","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"SSH
        password"},"sshPort":{"type":"string","default":{"stringValue":"22","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"SSH
        port"},"sshPrivateKeyPath":{"type":"string","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"SSH
        private key path "},"sshUser":{"type":"string","default":{"stringValue":"root","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"SSH
        username"},"userData":{"type":"string","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"userData
        content of cloud-init for machine, base64 is supported"},"vgpuInfo":{"type":"string","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"harvester-vgpu-info"},"vmAffinity":{"type":"string","default":{"stringValue":"","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"harvester
        vm affinity, base64 is supported"},"vmNamespace":{"type":"string","default":{"stringValue":"default","intValue":0,"boolValue":false,"stringSliceValue":null},"create":true,"update":true,"description":"harvester
        vm namespace"}}}'
      etcdRole: true
      machineConfigRef:
        kind: HarvesterConfig
        name: nc-51-pool1-wfw6m
      name: pool1
      quantity: 1
      unhealthyNodeTimeout: 0s
      workerRole: true
    machineSelectorConfig:
    - config:
        cloud-provider-config: secret://fleet-default:harvesterconfigndk26
        cloud-provider-name: harvester
        protect-kernel-defaults: false
    networking: {}
    registries: {}
    upgradeStrategy:
      controlPlaneConcurrency: "1"
      controlPlaneDrainOptions:
        deleteEmptyDirData: true
        disableEviction: false
        enabled: false
        force: false
        gracePeriod: -1
        ignoreDaemonSets: true
        skipWaitForDeleteTimeoutSeconds: 0
        timeout: 120
      workerConcurrency: "1"
      workerDrainOptions:
        deleteEmptyDirData: true
        disableEviction: false
        enabled: false
        force: false
        gracePeriod: -1
        ignoreDaemonSets: true
        skipWaitForDeleteTimeoutSeconds: 0
        timeout: 120
status:
  clusterName: c-m-nl2hgfgg
  conditions:
  - lastUpdateTime: "2026-03-03T22:09:35Z"
    status: "True"
    type: RKECluster
  - lastUpdateTime: "2026-03-03T22:05:56Z"
    reason: Reconciling
    status: "True"
    type: Reconciling
  - lastUpdateTime: "2026-03-03T22:05:56Z"
    status: "False"
    type: Stalled
  - lastUpdateTime: "2026-03-03T22:06:05Z"
    status: "True"
    type: Created
  - lastUpdateTime: "2026-03-03T22:05:56Z"
    status: "True"
    type: HarvesterCloudProviderConfigMigrated
  - lastUpdateTime: "2026-03-03T22:09:35Z"
    message: 'configuring bootstrap node(s) 51-pool1-gbxrc-kc8cx: waiting for cluster
      agent to connect'
    reason: Waiting
    status: Unknown
    type: Updated
  - lastUpdateTime: "2026-03-03T22:09:35Z"
    message: 'configuring bootstrap node(s) 51-pool1-gbxrc-kc8cx: waiting for cluster
      agent to connect'
    reason: Waiting
    status: Unknown
    type: Provisioned
  - lastUpdateTime: "2026-03-03T22:09:35Z"
    message: 'configuring bootstrap node(s) 51-pool1-gbxrc-kc8cx: waiting for cluster
      agent to connect'
    reason: Waiting
    status: Unknown
    type: Ready
  - lastUpdateTime: "2026-03-03T22:05:57Z"
    status: "True"
    type: BackingNamespaceCreated
  - lastUpdateTime: "2026-03-03T22:05:57Z"
    status: "True"
    type: DefaultProjectCreated
  - lastUpdateTime: "2026-03-03T22:05:57Z"
    status: "True"
    type: SystemProjectCreated
  - lastUpdateTime: "2026-03-03T22:05:57Z"
    status: "True"
    type: InitialRolesPopulated
  - lastUpdateTime: "2026-03-03T22:05:57Z"
    status: "True"
    type: CreatorMadeOwner
  - lastUpdateTime: "2026-03-03T22:05:57Z"
    status: "True"
    type: NoDiskPressure
  - lastUpdateTime: "2026-03-03T22:05:57Z"
    status: "True"
    type: NoMemoryPressure
  - lastUpdateTime: "2026-03-03T22:05:57Z"
    status: "True"
    type: ServiceAccountSecretsMigrated
  - lastUpdateTime: "2026-03-03T22:06:05Z"
    status: "False"
    type: Connected
  fleetWorkspaceName: fleet-default
  observedGeneration: 3
> 
```




### RKE2 bootstrap charts

```
rancher@gc8-pool1-q99l7-6ndw9:~$ cat /var/lib/rancher/rke2/server/manifests/harvester-cloud-provider.yaml
apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  annotations:
    helm.cattle.io/chart-url: https://rke2-charts.rancher.io/assets/harvester-cloud-provider/harvester-cloud-provider-0.2.1100.tgz
    rke2.cattle.io/inject-cluster-config: "true"
  name: harvester-cloud-provider
  namespace: kube-system
spec:
  bootstrap: true
  chartContent: H4sICEd/...
  set:
    global.clusterCIDR: 10.42.0.0/16
    global.clusterCIDRv4: 10.42.0.0/16
    global.clusterDNS: 10.43.0.10
    global.clusterDomain: cluster.local
    global.rke2DataDir: /var/lib/rancher/rke2
    global.serviceCIDR: 10.43.0.0/16
    global.systemDefaultIngressClass: traefik
  takeOwnership: false

```


### Guest cluster VM

```
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  annotations:
    harvesterhci.io/mac-address: '{"nic-0":"6e:22:0b:5d:e6:a7"}'
    harvesterhci.io/vmRunStrategy: RerunOnFailure
    harvesterhci.io/volumeClaimTemplates: >-
      [{"metadata":{"name":"gc8-pool1-q99l7-6ndw9-disk-0-hxzx7","creationTimestamp":null,"annotations":{"harvesterhci.io/imageId":"default/image-8854n"}},"spec":{"accessModes":["ReadWriteMany"],"resources":{"requests":{"storage":"40Gi"}},"storageClassName":"longhorn-image-8854n","volumeMode":"Block"},"status":{}}]
    kubevirt.io/latest-observed-api-version: v1
    kubevirt.io/storage-observed-api-version: v1
  creationTimestamp: '2026-03-03T20:59:28Z'
  finalizers:
    - wrangler.cattle.io/harvester-lb-vm-controller
    - kubevirt.io/virtualMachineControllerFinalize
    - wrangler.cattle.io/VMController.CleanupPVCAndSnapshot
  generation: 1
  labels:
    guestcluster.harvesterhci.io/name: gc8
    harvesterhci.io/creator: docker-machine-driver-harvester
    harvesterhci.io/machineSetName: default-gc8-pool1
    harvesterhci.io/vmName: gc8-pool1-q99l7-6ndw9
    nodepool.harvesterhci.io/name: pool1-q99l7
  managedFields:
    - apiVersion: kubevirt.io/v1
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:annotations:
            .: {}
            f:harvesterhci.io/volumeClaimTemplates: {}
          f:labels:
            .: {}
            f:guestcluster.harvesterhci.io/name: {}
            f:harvesterhci.io/creator: {}
            f:harvesterhci.io/machineSetName: {}
            f:harvesterhci.io/vmName: {}
            f:nodepool.harvesterhci.io/name: {}
        f:spec:
          .: {}
          f:runStrategy: {}
          f:template:
            .: {}
            f:metadata:
              .: {}
              f:annotations:
                .: {}
                f:harvesterhci.io/sshNames: {}
                f:harvesterhci.io/waitForLeaseInterfaceNames: {}
              f:creationTimestamp: {}
              f:labels:
                .: {}
                f:guestcluster.harvesterhci.io/name: {}
                f:harvesterhci.io/creator: {}
                f:harvesterhci.io/machineSetName: {}
                f:harvesterhci.io/vmName: {}
                f:nodepool.harvesterhci.io/name: {}
            f:spec:
              .: {}
              f:affinity:
                .: {}
                f:podAntiAffinity:
                  .: {}
                  f:preferredDuringSchedulingIgnoredDuringExecution: {}
              f:domain:
                .: {}
                f:cpu:
                  .: {}
                  f:cores: {}
                f:devices:
                  .: {}
                  f:disks: {}
                  f:interfaces: {}
                f:resources:
                  .: {}
                  f:limits:
                    .: {}
                    f:cpu: {}
                    f:memory: {}
              f:evictionStrategy: {}
              f:networks: {}
              f:volumes: {}
      manager: docker-machine-driver-harvester
      operation: Update
      time: '2026-03-03T20:59:28Z'
    - apiVersion: kubevirt.io/v1
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:finalizers:
            .: {}
            v:"wrangler.cattle.io/harvester-lb-vm-controller": {}
      manager: harvester-load-balancer
      operation: Update
      time: '2026-03-03T20:59:28Z'
    - apiVersion: kubevirt.io/v1
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:annotations:
            f:kubevirt.io/latest-observed-api-version: {}
            f:kubevirt.io/storage-observed-api-version: {}
          f:finalizers:
            v:"kubevirt.io/virtualMachineControllerFinalize": {}
      manager: virt-controller
      operation: Update
      time: '2026-03-03T20:59:28Z'
    - apiVersion: kubevirt.io/v1
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:annotations:
            f:harvesterhci.io/mac-address: {}
            f:harvesterhci.io/vmRunStrategy: {}
          f:finalizers:
            v:"wrangler.cattle.io/VMController.CleanupPVCAndSnapshot": {}
      manager: harvester
      operation: Update
      time: '2026-03-03T20:59:44Z'
    - apiVersion: kubevirt.io/v1
      fieldsType: FieldsV1
      fieldsV1:
        f:status:
          .: {}
          f:conditions: {}
          f:created: {}
          f:desiredGeneration: {}
          f:observedGeneration: {}
          f:printableStatus: {}
          f:ready: {}
          f:runStrategy: {}
          f:volumeSnapshotStatuses: {}
      manager: virt-controller
      operation: Update
      subresource: status
      time: '2026-03-03T21:00:28Z'
  name: gc8-pool1-q99l7-6ndw9
  namespace: default
  resourceVersion: '287780'
  uid: d348fcef-a65f-4ff0-88f8-7b1b28176156
spec:
  runStrategy: RerunOnFailure
  template:
    metadata:
      annotations:
        harvesterhci.io/sshNames: '[]'
        harvesterhci.io/waitForLeaseInterfaceNames: '[]'
      creationTimestamp: null
      labels:
        guestcluster.harvesterhci.io/name: gc8
        harvesterhci.io/creator: docker-machine-driver-harvester
        harvesterhci.io/machineSetName: default-gc8-pool1
        harvesterhci.io/vmName: gc8-pool1-q99l7-6ndw9
        nodepool.harvesterhci.io/name: pool1-q99l7
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - key: network.harvesterhci.io/mgmt
                    operator: In
                    values:
                      - 'true'
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - podAffinityTerm:
                labelSelector:
                  matchExpressions:
                    - key: harvesterhci.io/creator
                      operator: Exists
                topologyKey: kubernetes.io/hostname
              weight: 100
      architecture: amd64
      domain:
        cpu:
          cores: 2
          maxSockets: 1
          sockets: 1
          threads: 1
        devices:
          disks:
            - disk:
                bus: virtio
              name: cloudinitdisk
            - bootOrder: 1
              disk:
                bus: virtio
              name: disk-0
          interfaces:
            - bridge: {}
              macAddress: 6e:22:0b:5d:e6:a7
              model: virtio
              name: nic-0
        firmware:
          serial: d9f1baad-c68d-490c-834e-de799503e315
          uuid: 885ddb81-1491-4b92-889d-dc5098645b62
        machine:
          type: q35
        memory:
          guest: 4Gi
        resources:
          limits:
            cpu: '2'
            memory: 4Gi
          requests:
            cpu: 125m
            memory: 2730Mi
      evictionStrategy: LiveMigrateIfPossible
      networks:
        - multus:
            networkName: default/vm-untag
          name: nic-0
      terminationGracePeriodSeconds: 120
      volumes:
        - cloudInitNoCloud:
            secretRef:
              name: gc8-pool1-q99l7-6ndw9-cloudinit
          name: cloudinitdisk
        - name: disk-0
          persistentVolumeClaim:
            claimName: gc8-pool1-q99l7-6ndw9-disk-0-hxzx7
status:
  conditions:
    - lastProbeTime: null
      lastTransitionTime: '2026-03-03T20:59:43Z'
      status: 'True'
      type: Ready
    - lastProbeTime: null
      lastTransitionTime: null
      status: 'True'
      type: LiveMigratable
    - lastProbeTime: null
      lastTransitionTime: null
      status: 'True'
      type: StorageLiveMigratable
    - lastProbeTime: '2026-03-03T21:00:28Z'
      lastTransitionTime: null
      status: 'True'
      type: AgentConnected
  created: true
  desiredGeneration: 1
  observedGeneration: 1
  printableStatus: Running
  ready: true
  runStrategy: RerunOnFailure
  volumeSnapshotStatuses:
    - enabled: false
      name: cloudinitdisk
      reason: Snapshot is not supported for this volumeSource type [cloudinitdisk]
    - enabled: true
      name: disk-0
```


### Cloud-init


```
harv41:/home/rancher # kubectl get secret gc8-pool1-q99l7-6ndw9-cloudinit
NAME                              TYPE     DATA   AGE
gc8-pool1-q99l7-6ndw9-cloudinit   Opaque   1      10m

harv41:/home/rancher # kubectl get secret gc8-pool1-q99l7-6ndw9-cloudinit -oyaml
apiVersion: v1
data:
  userdata: I2Nsb3VkLWNvbmZpZwpob3N...hbGwuc2gKICBwZXJtaXNzaW9uczogIjA2NDQiCg==
kind: Secret
metadata:
  creationTimestamp: "2026-03-03T20:59:28Z"
  name: gc8-pool1-q99l7-6ndw9-cloudinit
  namespace: default
  ownerReferences:
  - apiVersion: kubevirt.io/v1
    kind: VirtualMachine
    name: gc8-pool1-q99l7-6ndw9
    uid: d348fcef-a65f-4ff0-88f8-7b1b28176156
  resourceVersion: "286901"
  uid: 34ff4502-9025-415f-a673-3fbf3ba99b27
type: Opaque
```


```
harv41:/home/rancher # kubectl get secret gc8-pool1-q99l7-6ndw9-cloudinit -ojsonpath="{.data.userdata}" | base64 -d
#cloud-config
hostname: gc8-pool1-q99l7-6ndw9
package_update: true
packages:
- qemu-guest-agent
runcmd:
- - systemctl
  - enable
  - --now
  - qemu-guest-agent.service
- sh /usr/local/custom_script/install.sh
ssh_authorized_keys:
- ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDYYbLEVCIwQ4FqsYKXoee+SV4dz0JRN2Spf0lYwpxtTfqXdoqm6B6V2QEAyCtQ+RI4Mij8H3v+TKeOORKZuEQDIokf9mcEvdA6pmyHipauHAgeWdCQrBURiAWwtC/vvWQOFGwnF7AwNeBd6rxrjibKvpeAH2YlgKp4lGBDdLYt9EKG0iHmTM7j6nA9Ze98N13EA5xdsGqMcGiAqIbqqAogO3m4IH3ncSJ3kEtZQUEFJ8OMSj22wrDoo7I2QHBM4+UmAJWbtnv0rjwKlJzjMotQcP0MPYakU4NWQBEyYT1X/VR7Ds9e09rK+WzdxoBMzPwoWC/+QJSK7Un0SVD+lbaR
users:
- groups:
  - sudo
  lock_passwd: false
  name: rancher
  passwd: $6$EF9iJzkMwD/CUQ$e8ELGbeLjs7sFEkmB75Oeok0pjsqLSW7iu1bmQwLlcSjTuBBjphyXo4yBy/flb9RDD/osr5nRzuCK9A280xKQ/
  shell: /bin/bash
  ssh-authorized-keys:
  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDukDaTZWUwq57hDZhXWoL0qjCOGwxmbn5ogWp+JIv3HA/0ff3+LaBFXSbCmAqB7qeIG7AQSSggSpsIBmpiI89dPkQYtkizxqEj7677+ZjSLWgARodpl6K5ASgm+48Fq+t3f/JVNyRhdkqAVM8OdPHhQD3BSqHmDHQmkt2S5GnUUukowjHI/vpG1f5RKnfLduQjWXmcuRtNOD1cyqT3uxRnUGkd3PeBjJ5li6nMx8tom+WmTe0ETd9Z0TyB48aK10GtzOLme+D/8tNgGEOCsG6f7A5kCkU9AAKAqeEuSWOb/Whz5YYLXysgChDYjWlFbSnqbAx9Qg2eprqCfYBGlD4utihEMoEmX5wNkWJ7CY1r2blmmoBjWtbHiY42LQ5gMOGaWArO/JGOhGV6zHWDwnJuC5H/06bxuU5gGscV7X1sb5uJ/sIuAGKmIo4Aq2uRiFAlCRd9dTqxI386cVTDv8ROX4AUVv0qr4F7897AF9x3ZdL6U5zPJcOo8Ssw3yzG/QM=
    jian.wang@suse.com
write_files:
- content: H4sIAAAA...  // content is to below file /usr/local/custom_script/install.sh
  encoding: gzip+b64
  path: /usr/local/custom_script/install.sh
  permissions: "0644"
```



### Install.sh to install the guest k8s cluster


```
rancher@gc8-pool1-q99l7-6ndw9:~$ cat /usr/local/custom_script/install.sh
STRICT_VERIFY="true"

CATTLE_AGENT_BINARY_BASE_URL="https://192.168.122.118/assets"
CATTLE_SERVER=https://192.168.122.118
CATTLE_CA_CHECKSUM="fb692854ad82f5eafedcd2c7fe817b482ff28243825599f4099a575606dd0754"
CATTLE_ROLE_NONE=true
CATTLE_TOKEN="Q4DK99ynUBZB8nS_yGMqPALK-p-K9_Ex3pRD4JoPc3c="

#!/bin/sh

if [ "${DEBUG}" = 1 ]; then
    set -x
    CURL_LOG="-v"
else
    CURL_LOG="-sS"
fi

# Usage:
#   curl ... | ENV_VAR=... sh -
#       or
#   ENV_VAR=... ./install.sh
#

# Environment variables:
#   System Agent Variables
#   - CATTLE_AGENT_LOGLEVEL (default: info)
#   - CATTLE_AGENT_CONFIG_DIR (default: /etc/rancher/agent)
#   - CATTLE_AGENT_VAR_DIR (default: /var/lib/rancher/agent)
#   - CATTLE_AGENT_BIN_PREFIX (default: /usr/local)
#
#   Rancher 2.6+ Variables
#   - CATTLE_SERVER
#   - CATTLE_TOKEN
#   - CATTLE_CA_CHECKSUM
#   - CATTLE_ROLE_CONTROLPLANE=false
#   - CATTLE_ROLE_ETCD=false
#   - CATTLE_ROLE_WORKER=false
#   - CATTLE_ROLE_NONE=false
#   - CATTLE_LABELS
#   - CATTLE_TAINTS
#
#   Advanced Environment Variables
#   - CATTLE_AGENT_BINARY_BASE_URL (default: latest GitHub release)
#   - CATTLE_AGENT_BINARY_URL (default: latest GitHub release)
#   - CATTLE_AGENT_UNINSTALL_URL (default: latest GitHub release)
#   - CATTLE_PRESERVE_WORKDIR (default: false)
#   - CATTLE_REMOTE_ENABLED (default: true)
#   - CATTLE_LOCAL_ENABLED (default: false)
#   - CATTLE_ID (default: autogenerate)
#   - CATTLE_AGENT_BINARY_LOCAL (default: false)
#   - CATTLE_AGENT_BINARY_LOCAL_LOCATION (default: )
#   - CATTLE_AGENT_UNINSTALL_LOCAL (default: false)
#   - CATTLE_AGENT_UNINSTALL_LOCAL_LOCATION (default: )
#   - CATTLE_AGENT_STRICT_VERIFY | STRICT_VERIFY (default: false)
#   - CATTLE_AGENT_FALLBACK_PATH (default: )

FALLBACK=v0.3.13
CACERTS_PATH=cacerts
RETRYCOUNT=4500
APPLYINATOR_ACTIVE_WAIT_COUNT=60 # If the system-agent is unhealthy but had created an interlock file to indicate it was actively applying a plan, after 5 minutes, ignore the interlock.
DEFAULT_BIN_PREFIX=/usr/local

# info logs the given argument at info log level.
info() {
    echo "[INFO] " "$@"
}

# warn logs the given argument at warn log level.
warn() {
    echo "[WARN] " "$@" >&2
}

# error logs the given argument at error log level.
error() {
    echo "[ERROR] " "$@" >&2
}

# fatal logs the given argument at fatal log level.
fatal() {
    echo "[FATAL] " "$@" >&2
    exit 1
}

# check_target_mountpoint return success if the target directory is on a dedicated mount point
check_target_mountpoint() {
    mountpoint -q "${DEFAULT_BIN_PREFIX}"
}

# check_target_ro returns success if the target directory is read-only
check_target_ro() {
    touch "${DEFAULT_BIN_PREFIX}"/.r-sa-ro-test && rm -rf "${DEFAULT_BIN_PREFIX}"/.r-sa-ro-test
    test $? -ne 0
}

# check_rootfs_rw returns success if the root filesystem is read-write so we can check for transactional-update system
check_rootfs_rw() {
    touch /.rootfs-rw-test && rm -rf /.rootfs-rw-test
    test $? -eq 0
}

# parse_args will inspect the argv for --server, --token, --controlplane, --etcd, and --worker, --label x=y, and --taint dead=beef:NoSchedule
parse_args() {
    while [ $# -gt 0 ]; do
        case "$1" in
        "-a" | "--all-roles")
            info "All roles requested"
            CATTLE_ROLE_CONTROLPLANE=true
            CATTLE_ROLE_ETCD=true
            CATTLE_ROLE_WORKER=true
            shift 1
            ;;
        "-p" | "--controlplane")
            info "Role requested: controlplane"
            CATTLE_ROLE_CONTROLPLANE=true
            shift 1
            ;;
        "-e" | "--etcd")
            info "Role requested: etcd"
            CATTLE_ROLE_ETCD=true
            shift 1
                ;;
        "-w" | "--worker")
            info "Role requested: worker"
            CATTLE_ROLE_WORKER=true
		        shift 1
            ;;
        "--no-roles")
            info "Role requested: none"
            CATTLE_ROLE_NONE=true
            shift 1
            ;;
        "-n" | "--node-name")
            CATTLE_NODE_NAME="$2"
		        shift 2
            ;;
        "-a" | "--address")
            CATTLE_ADDRESS="$2"
		        shift 2
            ;;
        "-i" | "--internal-address")
            CATTLE_INTERNAL_ADDRESS="$2"
		        shift 2
            ;;
        "-l" | "--label")
            info "Label: $2"
            if [ -n "${CATTLE_LABELS}" ]; then
                CATTLE_LABELS="${CATTLE_LABELS},$2"
            else
                CATTLE_LABELS="$2"
            fi
		        shift 2
            ;;
        "--taint" | "--taints")
            info "Taint: $2"
            if [ -n "${CATTLE_TAINTS}" ]; then
                CATTLE_TAINTS="${CATTLE_TAINTS},$2"
            else
                CATTLE_TAINTS="$2"
            fi
		        shift 2
            ;;
        "-s" | "--server")
            CATTLE_SERVER="$2"
		        shift 2
            ;;
        "-t" | "--token")
            CATTLE_TOKEN="$2"
		        shift 2
            ;;
        "-c" | "--ca-checksum")
            CATTLE_CA_CHECKSUM="$2"
            shift 2
            ;;
        *)
            fatal "Unknown argument passed in ($1)"
            ;;
        esac
    done
}

in_no_proxy() {
    # Get just the host name/IP
    ip_addr="${1#http://}"
    ip_addr="${ip_addr#https://}"
    ip_addr="${ip_addr%%/*}"
    ip_addr="${ip_addr%%:*}"

    # If this isn't an IP address, then there is nothing to check
    if [ "$(valid_ip "$ip_addr")" = "1" ]; then
      echo 1
      return
    fi

    i=1
    proxy_ip=$(echo "$NO_PROXY" | cut -d',' -f$i)
    while [ -n "$proxy_ip" ]; do
      subnet_ip=$(echo "${proxy_ip}" | cut -d'/' -f1)
      cidr_mask=$(echo "${proxy_ip}" | cut -d'/' -f2)

      if [ "$(valid_ip "$subnet_ip")" = "0" ]; then
        # If these were the same, then proxy_ip is an IP address, not a CIDR. curl handles this correctly.
        if [ "$cidr_mask" != "$subnet_ip" ]; then
          cidr_mask=$(( 32 - cidr_mask ))
          shift_multiply=1
          while [ "$cidr_mask" -gt 0 ]; do
            shift_multiply=$(( shift_multiply * 2 ))
            cidr_mask=$(( cidr_mask - 1 ))
          done

          # Manual left-shift (<<) by original cidr_mask value
          netmask=$(( 0xFFFFFFFF * shift_multiply ))

          # Apply netmask to both the subnet IP and the given IP address
          ip_addr_subnet=$(and "$(ip_to_int "$subnet_ip")" $netmask)
          subnet=$(and "$(ip_to_int "$ip_addr")" $netmask)

          # Subnet IPs will match if given IP address is in CIDR subnet
          if [ "${ip_addr_subnet}" -eq "${subnet}" ]; then
            echo 0
            return
          fi
        fi
      fi

      i=$(( i + 1 ))
      proxy_ip=$(echo "$NO_PROXY" | cut -d',' -s -f$i)
    done

    echo 1
}

# bitwise 'and' in /bin/sh is not supported, so we do it manually.
and() {
    ret=0
    first=${1}
    second=${2}
    if [ "$first" -gt "$second" ]; then
        tmp=$first
        first=$second
        second=$tmp
    fi

    while [ "$first" -gt 0 ]; do
        ret=$(( ret * 2 ))
        d1=$(( first % 2 ))
        d2=$(( second % 2 ))
        ans=$(( d1 * d2 ))
        if [ "$ans" -eq 1 ]; then
            ret=$(( ret + 1 ))
        fi
        second=$(( second / 2 ))
        first=$(( first / 2 ))
    done

    echo $ret
}

ip_to_int() {
    ip_addr="${1}"

    ip_1=$(echo "${ip_addr}" | cut -d'.' -f1)
    ip_2=$(echo "${ip_addr}" | cut -d'.' -f2)
    ip_3=$(echo "${ip_addr}" | cut -d'.' -f3)
    ip_4=$(echo "${ip_addr}" | cut -d'.' -f4)

    echo $(( $ip_1 * 256*256*256 + $ip_2 * 256*256 + $ip_3 * 256 + $ip_4 ))
}

valid_ip() {
    local IP="$1" IFS="." PART
    set -- $IP
    [ "$#" != 4 ] && echo 1 && return
    for PART; do
        case "$PART" in
            *[!0-9]*) echo 1 && return
        esac
        [ "$PART" -gt 255 ] && echo 1 && return
    done
    echo 0
}

setup_env() {
    if [ -z "${CATTLE_ROLE_CONTROLPLANE}" ]; then
        CATTLE_ROLE_CONTROLPLANE=false
    fi

    if [ -z "${CATTLE_ROLE_ETCD}" ]; then
        CATTLE_ROLE_ETCD=false
    fi

    if [ -z "${CATTLE_ROLE_WORKER}" ]; then
        CATTLE_ROLE_WORKER=false
    fi

    if [ -z "${CATTLE_ROLE_NONE}" ]; then
        CATTLE_ROLE_NONE=false
    fi

    if [ "${CATTLE_ROLE_NONE}" = "true" ]; then
        info "--no-roles flag passed, unsetting all other requested roles"
        CATTLE_ROLE_CONTROLPLANE=false
        CATTLE_ROLE_ETCD=false
        CATTLE_ROLE_WORKER=false
    fi

    if [ -z "${CATTLE_LOCAL_ENABLED}" ]; then
        CATTLE_LOCAL_ENABLED=false
    else
        CATTLE_LOCAL_ENABLED=$(echo "${CATTLE_LOCAL_ENABLED}" | tr '[:upper:]' '[:lower:]')
    fi

    if [ -z "${CATTLE_REMOTE_ENABLED}" ]; then
        CATTLE_REMOTE_ENABLED=true
    else
        CATTLE_REMOTE_ENABLED=$(echo "${CATTLE_REMOTE_ENABLED}" | tr '[:upper:]' '[:lower:]')
    fi

    if [ "${CATTLE_LOCAL_ENABLED}" = "false" ] && [ "${CATTLE_REMOTE_ENABLED}" = "false" ]; then
        fatal "Neither local or remote plan support was enabled"
    fi

    if [ -z "${CATTLE_PRESERVE_WORKDIR}" ]; then
        CATTLE_PRESERVE_WORKDIR=false
    else
        CATTLE_PRESERVE_WORKDIR=$(echo "${CATTLE_PRESERVE_WORKDIR}" | tr '[:upper:]' '[:lower:]')
    fi

    if [ -z "${CATTLE_AGENT_LOGLEVEL}" ]; then
        CATTLE_AGENT_LOGLEVEL=info
    else
        CATTLE_AGENT_LOGLEVEL=$(echo "${CATTLE_AGENT_LOGLEVEL}" | tr '[:upper:]' '[:lower:]')
    fi

    if [ "${CATTLE_AGENT_BINARY_LOCAL}" = "true" ]; then
        if [ -z "${CATTLE_AGENT_BINARY_LOCAL_LOCATION}" ]; then
            fatal "No local binary location was specified"
        fi
        BINARY_SOURCE=local
    else
        BINARY_SOURCE=remote

        if [ -z "${CATTLE_AGENT_BINARY_URL}" ] && [ -n "${CATTLE_AGENT_BINARY_BASE_URL}" ]; then
            CATTLE_AGENT_BINARY_URL="${CATTLE_AGENT_BINARY_BASE_URL}/rancher-system-agent-${ARCH}"
        fi

        if [ -z "${CATTLE_AGENT_BINARY_URL}" ]; then
            if [ $(curl --connect-timeout 60 --max-time 60 -s https://api.github.com/rate_limit | grep '"rate":' -A 4 | grep '"remaining":' | sed -E 's/.*"[^"]+": (.*),/\1/') = 0 ]; then
                info "GitHub Rate Limit exceeded, falling back to known good version"
                VERSION=$FALLBACK
            else
                VERSION=$(curl --connect-timeout 60 --max-time 60 -s "https://api.github.com/repos/rancher/system-agent/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
                if [ -z "$VERSION" ]; then # Fall back to a known good fallback version because we had an error pulling the latest
                    info "Error contacting GitHub to retrieve the latest version"
                    VERSION=$FALLBACK
                fi
            fi
            CATTLE_AGENT_BINARY_URL="https://github.com/rancher/system-agent/releases/download/${VERSION}/rancher-system-agent-${ARCH}"
            BINARY_SOURCE=upstream
        fi
    fi

    if [ "${CATTLE_AGENT_UNINSTALL_LOCAL}" = "true" ]; then
        if [ -z "${CATTLE_AGENT_UNINSTALL_LOCAL_LOCATION}" ]; then
            fatal "No local uninstall location was specified"
        fi
        UNINSTALL_SOURCE=local
    else
        UNINSTALL_SOURCE=remote

        if [ -z "${CATTLE_AGENT_UNINSTALL_URL}" ] && [ -n "${CATTLE_AGENT_BINARY_BASE_URL}" ]; then
            CATTLE_AGENT_UNINSTALL_URL="${CATTLE_AGENT_BINARY_BASE_URL}/system-agent-uninstall.sh"
        fi

        if [ -z "${CATTLE_AGENT_UNINSTALL_URL}" ]; then
            if [ -n "${VERSION}" ]; then
                info "Version ${VERSION} used for downloading the rancher-system-agent binary, will reuse for uninstall script"
            elif [ $(curl --connect-timeout 60 --max-time 60 -s https://api.github.com/rate_limit | grep '"rate":' -A 4 | grep '"remaining":' | sed -E 's/.*"[^"]+": (.*),/\1/') = 0 ]; then
                info "GitHub Rate Limit exceeded, falling back to known good version"
                VERSION=$FALLBACK
            else
                VERSION=$(curl --connect-timeout 60 --max-time 60 -s "https://api.github.com/repos/rancher/system-agent/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
                if [ -z "$VERSION" ]; then # Fall back to a known good fallback version because we had an error pulling the latest
                    info "Error contacting GitHub to retrieve the latest version"
                    VERSION=$FALLBACK
                fi
            fi
            CATTLE_AGENT_UNINSTALL_URL="https://github.com/rancher/system-agent/releases/download/${VERSION}/system-agent-uninstall.sh"
            UNINSTALL_SOURCE=upstream
        fi
    fi

    if [ "${CATTLE_REMOTE_ENABLED}" = "true" ]; then
        if [ -z "${CATTLE_TOKEN}" ]; then
            fatal "\$CATTLE_TOKEN was not set."
        fi
        if [ -z "${CATTLE_SERVER}" ]; then
            fatal "\$CATTLE_SERVER was not set"
        fi
        if [ "${CATTLE_ROLE_CONTROLPLANE}" = "false" ] && [ "${CATTLE_ROLE_ETCD}" = "false" ] && [ "${CATTLE_ROLE_WORKER}" = "false" ] && [ "${CATTLE_ROLE_NONE}" = "false" ]; then
            fatal "You must select at least one role."
        fi
    fi

    if [ -z "${CATTLE_AGENT_STRICT_VERIFY}" ]; then
      CATTLE_AGENT_STRICT_VERIFY=false
      if [ -n "${STRICT_VERIFY}" ]; then
        CATTLE_AGENT_STRICT_VERIFY=${STRICT_VERIFY}
      fi
      info "CA strict verification is set to ${CATTLE_AGENT_STRICT_VERIFY}"
    fi

    if [ -z "${CATTLE_AGENT_CONFIG_DIR}" ]; then
        CATTLE_AGENT_CONFIG_DIR=/etc/rancher/agent
        info "Using default agent configuration directory ${CATTLE_AGENT_CONFIG_DIR}"
    fi

    # --- install to /var/lib/rancher/agent by default, except if we are running within transactional-update
    # --- in which case we install into /etc/rancher/agent/var as /var is not mounted to the snapshot.
    if [ -z "${CATTLE_AGENT_VAR_DIR}" ]; then
        if [ -x /usr/sbin/transactional-update ] && check_rootfs_rw; then
            CATTLE_AGENT_VAR_DIR=/etc/rancher/agent/var
            info "Detected a transactional-update server, using ${CATTLE_AGENT_VAR_DIR} for agent var directory"
        else
            CATTLE_AGENT_VAR_DIR=/var/lib/rancher/agent
            info "Using default agent var directory ${CATTLE_AGENT_VAR_DIR}"
        fi
    fi

    # --- install to /usr/local by default, except if /usr/local is on a separate partition or is read-only
    # --- in which case we go into /opt/rancher-system-agent. If we are running within transactional-update
    # --- we install to /usr as /usr/local and /opt are not mounted to the snapshot.
    if [ -z "${CATTLE_AGENT_BIN_PREFIX}" ]; then
        CATTLE_AGENT_BIN_PREFIX="/usr/local"
        if check_target_mountpoint || check_target_ro; then
            CATTLE_AGENT_BIN_PREFIX="/opt/rancher-system-agent"
            warn "/usr/local is read-only or a mount point; installing to ${CATTLE_AGENT_BIN_PREFIX}"
        fi
        if [ -x /usr/sbin/transactional-update ] && check_rootfs_rw; then
            CATTLE_AGENT_BIN_PREFIX=/usr
            warn "Detected transactional-update in progress; installing to ${CATTLE_AGENT_BIN_PREFIX}"
        fi
    fi

    CATTLE_ADDRESS=$(get_address "${CATTLE_ADDRESS}")
    CATTLE_INTERNAL_ADDRESS=$(get_address "${CATTLE_INTERNAL_ADDRESS}")
}

ensure_directories() {
    mkdir -p ${CATTLE_AGENT_VAR_DIR}/interlock
    mkdir -p ${CATTLE_AGENT_CONFIG_DIR}
    chmod 700 ${CATTLE_AGENT_VAR_DIR}
    chmod 700 ${CATTLE_AGENT_VAR_DIR}/interlock
    chmod 700 ${CATTLE_AGENT_CONFIG_DIR}
    chown root:root ${CATTLE_AGENT_VAR_DIR}
    chown root:root ${CATTLE_AGENT_VAR_DIR}/interlock
    chown root:root ${CATTLE_AGENT_CONFIG_DIR}
}

# setup_arch set arch and suffix,
# fatal if architecture not supported.
setup_arch() {
    case ${ARCH:=$(uname -m)} in
    amd64)
        ARCH=amd64
        SUFFIX=$(uname -s | tr '[:upper:]' '[:lower:]')-${ARCH}
        ;;
    x86_64)
        ARCH=amd64
        SUFFIX=$(uname -s | tr '[:upper:]' '[:lower:]')-${ARCH}
        ;;
    arm64)
        ARCH=arm64
        SUFFIX=-${ARCH}
        ;;
    s390x)
        ARCH=s390x
        SUFFIX=-${ARCH}
        ;;
    aarch64)
        ARCH=arm64
        SUFFIX=-${ARCH}
        ;;
    arm*)
        ARCH=arm
        SUFFIX=-${ARCH}hf
        ;;
    *)
        fatal "unsupported architecture ${ARCH}"
        ;;
    esac
}

get_address()
{
    local address=$1
    # If nothing is given, return empty (it will be automatically determined later if empty)
    if [ -z $address ]; then
        echo ""
    # If given address is a network interface on the system, retrieve configured IP on that interface (only the first configured IP is taken)
    elif [ -n "$(find /sys/devices -name $address)" ]; then
        echo $(ip addr show dev $address | grep -w inet | awk '{print $2}' | cut -f1 -d/ | head -1)
    # Loop through cloud provider options to get IP from metadata, if not found return given value
    else
        noproxy=""
        if [ "$(in_no_proxy "169.254.169.254")" -eq 0 ]; then
          noproxy="--noproxy '*'"
        fi
        case $address in
            awslocal)
                echo $(curl $noproxy --connect-timeout 60 --max-time 60 -s http://169.254.169.254/latest/meta-data/local-ipv4)
                ;;
            awspublic)
                echo $(curl $noproxy --connect-timeout 60 --max-time 60 -s http://169.254.169.254/latest/meta-data/public-ipv4)
                ;;
            doprivate)
                echo $(curl $noproxy --connect-timeout 60 --max-time 60 -s http://169.254.169.254/metadata/v1/interfaces/private/0/ipv4/address)
                ;;
            dopublic)
                echo $(curl $noproxy --connect-timeout 60 --max-time 60 -s http://169.254.169.254/metadata/v1/interfaces/public/0/ipv4/address)
                ;;
            azprivate)
                echo $(curl $noproxy --connect-timeout 60 --max-time 60 -s -H Metadata:true "http://169.254.169.254/metadata/instance/network/interface/0/ipv4/ipAddress/0/privateIpAddress?api-version=2017-08-01&format=text")
                ;;
            azpublic)
                echo $(curl $noproxy --connect-timeout 60 --max-time 60 -s -H Metadata:true "http://169.254.169.254/metadata/instance/network/interface/0/ipv4/ipAddress/0/publicIpAddress?api-version=2017-08-01&format=text")
                ;;
            gceinternal)
                echo $(curl $noproxy --connect-timeout 60 --max-time 60 -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/ip)
                ;;
            gceexternal)
                echo $(curl $noproxy --connect-timeout 60 --max-time 60 -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip)
                ;;
            packetlocal)
                echo $(curl $noproxy --connect-timeout 60 --max-time 60 -s https://metadata.packet.net/2009-04-04/meta-data/local-ipv4)
                ;;
            packetpublic)
                echo $(curl $noproxy --connect-timeout 60 --max-time 60 -s https://metadata.packet.net/2009-04-04/meta-data/public-ipv4)
                ;;
            ipify)
                echo $(curl $noproxy --connect-timeout 60 --max-time 60 -s https://api.ipify.org)
                ;;
            *)
                echo $address
                ;;
        esac
    fi
}

# verify_downloader verifies existence of
# network downloader executable.
verify_downloader() {
    cmd="$(command -v "${1}")"
    if [ -z "${cmd}" ]; then
        return 1
    fi
    if [ ! -x "${cmd}" ]; then
        return 1
    fi

    # Set verified executable as our downloader program and return success
    DOWNLOADER=${cmd}
    return 0
}

# --- write systemd service file ---
create_systemd_service_file() {
    info "systemd: Creating service file"

    UMASK=$(umask)
    umask 022

    cat <<-EOF >"/etc/systemd/system/rancher-system-agent.service"
[Unit]
Description=Rancher System Agent
Documentation=https://www.rancher.com
Wants=network-online.target
After=network-online.target
[Install]
WantedBy=multi-user.target
[Service]
EnvironmentFile=-/etc/default/rancher-system-agent
EnvironmentFile=-/etc/sysconfig/rancher-system-agent
EnvironmentFile=-/etc/systemd/system/rancher-system-agent.env
Type=simple
Restart=always
RestartSec=5s
Environment=CATTLE_LOGLEVEL=${CATTLE_AGENT_LOGLEVEL}
Environment=CATTLE_AGENT_CONFIG=${CATTLE_AGENT_CONFIG_DIR}/config.yaml
Environment=CATTLE_AGENT_STRICT_VERIFY=${CATTLE_AGENT_STRICT_VERIFY}
ExecStart=${CATTLE_AGENT_BIN_PREFIX}/bin/rancher-system-agent sentinel
EOF

    umask "${UMASK}"
}

download_rancher_files() {
  mkdir -p ${CATTLE_AGENT_BIN_PREFIX}/bin

  download_rancher_file "rancher-system-agent" "binary" "${CATTLE_AGENT_BINARY_URL}" "${CATTLE_AGENT_BINARY_LOCAL}" "${CATTLE_AGENT_BINARY_LOCAL_LOCATION}" "${BINARY_SOURCE}"
  download_rancher_file "rancher-system-agent-uninstall.sh" "script" "${CATTLE_AGENT_UNINSTALL_URL}" "${CATTLE_AGENT_UNINSTALL_LOCAL}" "${CATTLE_AGENT_UNINSTALL_LOCAL_LOCATION}" "${UNINSTALL_SOURCE}"
}

download_rancher_file() {
  name=$1
  category=$2
  url=$3
  local=$4
  local_location=$5
  source=$6

  if [ "${local}" = "true" ]; then
      info "Using local ${name} ${category} from ${local_location}"
      cp -f "${local_location}" "${CATTLE_AGENT_BIN_PREFIX}/bin/${name}"
  else
      info "Downloading ${name} ${category} from ${url}"
      if [ "${source}" != "upstream" ]; then
          CURL_BIN_CAFLAG="${CURL_CAFLAG}"
      else
          CURL_BIN_CAFLAG=""
      fi
      i=1
      while [ "${i}" -ne "${RETRYCOUNT}" ]; do
          noproxy=""
          if [ "$(in_no_proxy "${url}")" = "0" ]; then
              noproxy="--noproxy '*'"
          fi
          RESPONSE=$(curl $noproxy --connect-timeout 60 --max-time 300 --write-out "%{http_code}\n" ${CURL_BIN_CAFLAG} ${CURL_LOG} -fL "${url}" -o "${CATTLE_AGENT_BIN_PREFIX}/bin/${name}")
          case "${RESPONSE}" in
          200)
              info "Successfully downloaded the ${name} ${category}."
              break
              ;;
          *)
              i=$((i + 1))
              error "$RESPONSE received while downloading the ${name} ${category}. Sleeping for 5 seconds and trying again"
              sleep 5
              continue
              ;;
          esac
      done
      chmod +x "${CATTLE_AGENT_BIN_PREFIX}/bin/${name}"
  fi
}

check_x509_cert()
{
    cert=$1
    err=$(openssl x509 -in "${cert}" -noout 2>&1)
    if [ $? -eq 0 ]
    then
        echo ""
    else
        echo "${err}"
    fi
}

validate_ca_checksum() {
    if [ -n "${CATTLE_CA_CHECKSUM}" ]; then
        CACERT=$(mktemp)
        i=1
        while [ "${i}" -ne "${RETRYCOUNT}" ]; do
            noproxy=""
            if [ "$(in_no_proxy ${CATTLE_AGENT_BINARY_URL})" = "0" ]; then
                noproxy="--noproxy '*'"
            fi
            RESPONSE=$(curl $noproxy --connect-timeout 60 --max-time 60 --write-out "%{http_code}\n" --insecure ${CURL_LOG} -fL "${CATTLE_SERVER}/${CACERTS_PATH}" -o ${CACERT})
            case "${RESPONSE}" in
            200)
                info "Successfully downloaded CA certificate"
                break
                ;;
            *)
                i=$((i + 1))
                error "$RESPONSE received while downloading the CA certificate. Sleeping for 5 seconds and trying again"
                sleep 5
                continue
                ;;
            esac
        done
        if [ ! -s "${CACERT}" ]; then
          error "The environment variable CATTLE_CA_CHECKSUM is set but there is no CA certificate configured at ${CATTLE_SERVER}/${CACERTS_PATH}"
          exit 1
        fi
        err=$(check_x509_cert "${CACERT}")
        if [ -n "${err}" ]; then
            error "Value from ${CATTLE_SERVER}/${CACERTS_PATH} does not look like an x509 certificate (${err})"
            error "Retrieved cacerts:"
            cat "${CACERT}"
            rm -f "${CACERT}"
            exit 1
        else
            info "Value from ${CATTLE_SERVER}/${CACERTS_PATH} is an x509 certificate"
        fi
        CATTLE_SERVER_CHECKSUM=$(sha256sum "${CACERT}" | awk '{print $1}')
        if [ "${CATTLE_SERVER_CHECKSUM}" != "${CATTLE_CA_CHECKSUM}" ]; then
            rm -f "${CACERT}"
            error "Configured cacerts checksum ($CATTLE_SERVER_CHECKSUM) does not match given --ca-checksum ($CATTLE_CA_CHECKSUM)"
            error "Please check if the correct certificate is configured at${CATTLE_SERVER}/${CACERTS_PATH}"
            exit 1
        fi
        CURL_CAFLAG="--cacert ${CACERT}"
    fi
}

validate_rancher_connection() {
    RANCHER_SUCCESS=false
    if [ -n "${CATTLE_SERVER}" ] && [ "${CATTLE_REMOTE_ENABLED}" = "true" ]; then
        i=1
        while [ "${i}" -ne "${RETRYCOUNT}" ]; do
            noproxy=""
            if [ "$(in_no_proxy ${CATTLE_AGENT_BINARY_URL})" = "0" ]; then
                noproxy="--noproxy '*'"
            fi
            RESPONSE=$(curl $noproxy --connect-timeout 60 --max-time 60 --write-out "%{http_code}\n" ${CURL_CAFLAG} ${CURL_LOG} -fL "${CATTLE_SERVER}/healthz" -o /dev/null)
            case "${RESPONSE}" in
            200)
                info "Successfully tested Rancher connection"
                RANCHER_SUCCESS=true
                break
                ;;
            *)
                i=$((i + 1))
                error "$RESPONSE received while testing Rancher connection. Sleeping for 5 seconds and trying again"
                sleep 5
                continue
                ;;
            esac
        done
        if [ "${RANCHER_SUCCESS}" != "true" ]; then
          fatal "Error connecting to Rancher. Perhaps --ca-checksum needs to be set?"
        fi
    fi
}

validate_ca_required() {
    CA_REQUIRED=false
    if [ -n "${CATTLE_SERVER}" ] && [ "${CATTLE_REMOTE_ENABLED}" = "true" ]; then
        i=1
        while [ "${i}" -ne "${RETRYCOUNT}" ]; do
            noproxy=""
            if [ "$(in_no_proxy ${CATTLE_AGENT_BINARY_URL})" = "0" ]; then
                noproxy="--noproxy '*'"
            fi
            VERIFY_RESULT=$(curl $noproxy --connect-timeout 60 --max-time 60 --write-out "%{ssl_verify_result}\n" ${CURL_LOG} -fL "${CATTLE_SERVER}/healthz" -o /dev/null 2>/dev/null)
            CURL_EXIT="$?"
            case "${CURL_EXIT}" in
              0|60)
                case "${VERIFY_RESULT}" in
                  0)
                    info "Determined CA is not necessary to connect to Rancher"
                    CA_REQUIRED=false
                    CATTLE_CA_CHECKSUM=""
                    break
                    ;;
                  *)
                    i=$((i + 1))
                    if [ "${CURL_EXIT}" -eq "60" ]; then
                      info "Determined CA is necessary to connect to Rancher"
                      CA_REQUIRED=true
                      break
                    fi
                    error "Error received while testing necessity of CA. Sleeping for 5 seconds and trying again"
                    sleep 5
                    continue
                    ;;
                esac
                ;;
              *)
                error "Error while connecting to Rancher to verify CA necessity. Sleeping for 5 seconds and trying again."
                sleep 5
                continue
                ;;
            esac
        done
    fi
}

retrieve_connection_info() {
    if [ "${CATTLE_REMOTE_ENABLED}" = "true" ]; then
        UMASK=$(umask)
        umask 0177
        i=1
        while [ "${i}" -ne "${RETRYCOUNT}" ]; do
            noproxy=""
            if [ "$(in_no_proxy ${CATTLE_AGENT_BINARY_URL})" = "0" ]; then
                noproxy="--noproxy '*'"
            fi
            RESPONSE=$(curl $noproxy --connect-timeout 60 --max-time 60 --write-out "%{http_code}\n" ${CURL_CAFLAG} ${CURL_LOG} -H "Authorization: Bearer ${CATTLE_TOKEN}" -H "X-Cattle-Id: ${CATTLE_ID}" -H "X-Cattle-Role-Etcd: ${CATTLE_ROLE_ETCD}" -H "X-Cattle-Role-Control-Plane: ${CATTLE_ROLE_CONTROLPLANE}" -H "X-Cattle-Role-Worker: ${CATTLE_ROLE_WORKER}" -H "X-Cattle-Node-Name: ${CATTLE_NODE_NAME}" -H "X-Cattle-Address: ${CATTLE_ADDRESS}" -H "X-Cattle-Internal-Address: ${CATTLE_INTERNAL_ADDRESS}" -H "X-Cattle-Labels: ${CATTLE_LABELS}" -H "X-Cattle-Taints: ${CATTLE_TAINTS}" "${CATTLE_SERVER}"/v3/connect/agent -o ${CATTLE_AGENT_VAR_DIR}/rancher2_connection_info.json)
            case "${RESPONSE}" in
            200)
                info "Successfully downloaded Rancher connection information"
                umask "${UMASK}"
                return 0
                ;;
            *)
                i=$((i + 1))
                error "$RESPONSE received while downloading Rancher connection information. Sleeping for 5 seconds and trying again"
                sleep 5
                continue
                ;;
            esac
        done
        error "Failed to download Rancher connection information in ${i} attempts"
        umask "${UMASK}"
        # Clean up invalid rancher2_connection_info.json file
        rm -f ${CATTLE_AGENT_VAR_DIR}/rancher2_connection_info.json
        return 1
    fi
}

generate_config() {
    UMASK=$(umask)
    umask 0177
cat <<-EOF >"${CATTLE_AGENT_CONFIG_DIR}/config.yaml"
workDirectory: ${CATTLE_AGENT_VAR_DIR}/work
appliedPlanDirectory: ${CATTLE_AGENT_VAR_DIR}/applied
remoteEnabled: ${CATTLE_REMOTE_ENABLED}
localEnabled: ${CATTLE_LOCAL_ENABLED}
localPlanDirectory: ${CATTLE_AGENT_VAR_DIR}/plans
interlockDirectory: ${CATTLE_AGENT_VAR_DIR}/interlock
preserveWorkDirectory: ${CATTLE_PRESERVE_WORKDIR}
EOF
    umask "${UMASK}"
    if [ "${CATTLE_REMOTE_ENABLED}" = "true" ]; then
        echo connectionInfoFile: ${CATTLE_AGENT_VAR_DIR}/rancher2_connection_info.json >> "${CATTLE_AGENT_CONFIG_DIR}/config.yaml"
    fi
}

generate_cattle_identifier() {
    if [ -z "${CATTLE_ID}" ]; then
        info "Generating Cattle ID"
        if [ -f "${CATTLE_AGENT_CONFIG_DIR}/cattle-id" ]; then
            CATTLE_ID=$(cat ${CATTLE_AGENT_CONFIG_DIR}/cattle-id);
            if [ -z "${CATTLE_ID}" ]; then
              fatal "Cattle ID was empty, aborting installation"
            fi
            info "Cattle ID was already detected as ${CATTLE_ID}. Not generating a new one."
            return
        fi

        CATTLE_ID=$(dd if=/dev/urandom count=1 bs=512 2>/dev/null | sha256sum | awk '{print $1}' | head -c 63);
        UMASK=$(umask)
        umask 0177
        echo "${CATTLE_ID}" > ${CATTLE_AGENT_CONFIG_DIR}/cattle-id
        umask "${UMASK}"
        if [ ! -s ${CATTLE_AGENT_CONFIG_DIR}/cattle-id ]; then
          fatal "Cattle ID could not be persisted. Aborting installation"
        fi
        return
    fi
    info "Not generating Cattle ID"
}


ensure_systemd_service_stopped() {
    if systemctl is-active --quiet rancher-system-agent.service; then
        info "Rancher System Agent was detected on this host. Ensuring the rancher-system-agent is stopped."
        systemctl stop rancher-system-agent
    fi
}

create_env_file() {
    FILE_SA_ENV="/etc/systemd/system/rancher-system-agent.env"
    info "Creating environment file ${FILE_SA_ENV}"
    install -m 0600 /dev/null "${FILE_SA_ENV}"
    for i in "HTTP_PROXY" "HTTPS_PROXY" "NO_PROXY"; do
      eval v=\"\$$i\"
      if [ -z "${v}" ]; then
        env | grep -E -i "^${i}" | tee -a ${FILE_SA_ENV} >/dev/null
      else
        echo "$i=$v" | tee -a ${FILE_SA_ENV} >/dev/null
      fi
    done

    # if /usr/local/ is ready only or on a separate partition, we want to add the bin dirs of rke2/k3s to our path
    if check_target_mountpoint || check_target_ro; then
      info "${DEFAULT_BIN_PREFIX} is unsuitable for installation: adding fallback path to systemd unit env file."
      if [ -n "${CATTLE_AGENT_FALLBACK_PATH}" ]; then
        echo "PATH=${PATH}:${CATTLE_AGENT_FALLBACK_PATH}" | tee -a ${FILE_SA_ENV} >/dev/null
      else
        echo "PATH=${PATH}:/opt/rke2/bin:/opt/bin" | tee -a ${FILE_SA_ENV} >/dev/null
      fi
    fi
}

ensure_applyinator_not_active() {
    i=1
    while [ "${i}" -ne "${APPLYINATOR_ACTIVE_WAIT_COUNT}" ]; do
      if [ -f "${CATTLE_AGENT_VAR_DIR}/interlock/applyinator-active" ]; then
        i=$((i + 1))
        info "Active plan reconciliation detected. Sleeping for 5 seconds and retrying check"
        sleep 5
        continue
      fi
      break
    done
}

do_install() {
    if [ $(id -u) != 0 ]; then
      fatal "This script must be run as root."
    fi

    parse_args "$@"
    setup_arch
    setup_env
    ensure_directories
    verify_downloader curl || fatal "can not find curl for downloading files"

    touch ${CATTLE_AGENT_VAR_DIR}/interlock/restart-pending
    ensure_applyinator_not_active

    if [ -z "${CATTLE_CA_CHECKSUM}" ] && [ $(echo "${CATTLE_AGENT_STRICT_VERIFY}" | tr '[:upper:]' '[:lower:]') = "true" ]; then
      fatal "Aborting system-agent installation due to requested strict CA verification with no CA checksum provided"
    fi
    if [ -n "${CATTLE_CA_CHECKSUM}" ] && [ $(echo "${CATTLE_AGENT_STRICT_VERIFY}" | tr '[:upper:]' '[:lower:]') != "true" ]; then
        validate_ca_required
    fi
    validate_ca_checksum
    validate_rancher_connection

    ensure_systemd_service_stopped

    download_rancher_files
    generate_config

    if [ -n "${CATTLE_TOKEN}" ]; then
        generate_cattle_identifier
        retrieve_connection_info || fatal "Aborting system-agent installation due to failure to retrieve Rancher connection information"
    fi
    create_systemd_service_file
    create_env_file
    systemctl daemon-reload >/dev/null
    info "Enabling rancher-system-agent.service"
    systemctl enable rancher-system-agent
    info "Starting/restarting rancher-system-agent.service"
    systemctl restart rancher-system-agent
    rm -f ${CATTLE_AGENT_VAR_DIR}/interlock/restart-pending
}

do_install "$@"
exit 0

rancher@gc8-pool1-q99l7-6ndw9:~$ 
```
