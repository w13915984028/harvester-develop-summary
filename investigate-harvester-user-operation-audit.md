# Investigate Harvester User Operation Audit

In Harvester cluster, when user operates from Harvester UI, the operation like `create a VM` is processed on following path:

```
 nginx POD---------embedded Rancher POD---------Harvester POD---------kube-apiserver

 http proxy         user AAA & aud                http server                
```

Many of them have related logs.

## Nginx log


The nginx log records the URL, but no detailed user information.

```
kubectl logs -n kube-system rke2-ingress-nginx-controller-6frqq


192.168.122.145 - - [07/Mar/2025:09:20:23 +0000] "POST /v1/harvester/kubevirt.io.virtualmachines/default HTTP/2.0" 201 1309 "https://192.168.122.122/dashboard/harvester/c/local/kubevirt.io.virtualmachine/create" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36" 1614 0.031 [cattle-system-rancher-80] [] 10.52.0.249:80 1309 0.032 201 0ca207d1a71ffa94cab23d5458284c8e


192.168.122.145 - - [07/Mar/2025:09:20:23 +0000] "POST /v1/harvester/secrets/default HTTP/2.0" 201 649 "https://192.168.122.122/dashboard/harvester/c/local/kubevirt.io.virtualmachine/create" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36" 421 0.008 [cattle-system-rancher-80] [] 10.52.0.249:80 649 0.008 201 cf677c47a6936cfab7d4a0ee75f2f21e

```

## Kube-apiserver Audit

The kubeaudit log is saved to below path, it has internal RBAC information.

```
 /var/lib/rancher/rke2/server/logs/audit.log

```


```
{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"7740b93b-f3ca-461e-96d4-84fad256b17c","stage":"RequestReceived","requestURI":"/apis/kubevirt.io/v1/namespaces/default/virtualmachineinstances/vm33","verb":"patch","user":{"username":"system:serviceaccount:harvester-system:kubevirt-controller","uid":"983c51c5-a5e5-4bfe-8ecb-fc26f68d731f","groups":["system:serviceaccounts","system:serviceaccounts:harvester-system","system:authenticated"],"extra":{"authentication.kubernetes.io/credential-id":["JTI=01acf7c8-e02b-4f19-9b38-b097ec48c5ed"],"authentication.kubernetes.io/node-name":["harv2"],"authentication.kubernetes.io/node-uid":["e4052175-985f-469e-91c1-7bc36c072c82"],"authentication.kubernetes.io/pod-name":["virt-controller-5dd599df-jh8l7"],"authentication.kubernetes.io/pod-uid":["d083bcd8-654e-409b-b792-8145c0897a63"]}},"sourceIPs":["10.52.0.214"],"userAgent":"virt-controller/v0.0.0 (linux/amd64) kubernetes/$Format","objectRef":{"resource":"virtualmachineinstances","namespace":"default","name":"vm33","apiGroup":"kubevirt.io","apiVersion":"v1"},"requestReceivedTimestamp":"2025-03-07T09:20:41.916302Z","stageTimestamp":"2025-03-07T09:20:41.916302Z"}

```

## Embedded Rancher audit log

The default embedded Rancher pod config on Harvester:

```
pods -n cattle-system rancher-95d7bd4d8-vz7cw

  name: rancher-95d7bd4d8-vz7cw
  namespace: cattle-system

  containers:
  - args:
    - --no-cacerts
    - --http-listen-port=80
    - --https-listen-port=443
    - --add-local=true
    env:
    - name: CATTLE_NAMESPACE
      value: cattle-system
    - name: CATTLE_PEER_SERVICE
      value: rancher
    - name: CATTLE_FEATURES
      value: multi-cluster-management=false,multi-cluster-management-agent=false,managed-system-upgrade-controller=false
    - name: CATTLE_SYSTEM_CATALOG
      value: bundled
    - name: CATTLE_BOOTSTRAP_PASSWORD
      valueFrom:
        secretKeyRef:
          key: bootstrapPassword
          name: bootstrap-secret
    - name: CATTLE_AGENT_TLS_MODE
      value: system-store
    image: rancher/rancher:v2.10.1
```

Add `- --audit-level=2` to `args` manually, this will enable `audit-log`


```
      - args:
        - --no-cacerts
        - --http-listen-port=80
        - --https-listen-port=443
        - --add-local=true
        - --audit-level=2
```

Refer https://github.com/rancher/rancher/blob/1d8cb65e9395dfd00d2af26058c60350ea20e677/main.go#L136C1-L142C5

```
		cli.IntFlag{
			Name:        "audit-level",
			Value:       0,
			EnvVar:      "AUDIT_LEVEL",
			Usage:       "Audit log level: 0 - disable audit log, 1 - log event metadata, 2 - log event metadata and request body, 3 - log event metadata, request body and response body",
			Destination: &config.AuditLevel,
		},
```


All audit logs are saved to following path on Rancher POD.

```
/var/log/auditlog/rancher-api-audit.log
```

Log formats:

```
{"auditID":"887f5e05-bd98-4ed9-9717-8c179f417236","requestURI":"/v1/harvester/kubevirt.io.virtualmachines/default","user":{"name":"user-vxq5m","group":["system:authenticated","system:cattle:authenticated"],"extra":{"principalid":["local://user-vxq5m"],"username":["admin"]}},"method":"POST","remoteAddr":"10.52.0.237:42344","requestTimestamp":"2025-03-07T09:20:23Z","responseTimestamp":"2025-03-07T09:20:23Z","responseCode":201,"requestHeader":{"Accept":["application/json"],"Accept-Encoding":["gzip, deflate, br, zstd"],"Accept-Language":["en-US,en;q=0.9"],"Content-Length":["1513"],"Content-Type":["application/json"],"Dnt":["1"],"Origin":["https://192.168.122.122"],"Priority":["u=1, i"],"Referer":["https://192.168.122.122/dashboard/harvester/c/local/kubevirt.io.virtualmachine/create"],"Sec-Ch-Ua":["\"Not A(Brand\";v=\"8\", \"Chromium\";v=\"132\", \"Google Chrome\";v=\"132\""],"Sec-Ch-Ua-Mobile":["?0"],"Sec-Ch-Ua-Platform":["\"Linux\""],"Sec-Fetch-Dest":["empty"],"Sec-Fetch-Mode":["cors"],"Sec-Fetch-Site":["same-origin"],"User-Agent":["Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36"],"X-Api-Csrf":["1ebdb507946de829dc0b4feac754bb08"],"X-Forwarded-For":["192.168.122.145"],"X-Forwarded-Host":["192.168.122.122"],"X-Forwarded-Port":["443"],"X-Forwarded-Proto":["https"],"X-Forwarded-Scheme":["https"],"X-Real-Ip":["192.168.122.145"],"X-Request-Id":["0ca207d1a71ffa94cab23d5458284c8e"],"X-Scheme":["https"]},"responseHeader":{"Cache-Control":["no-cache, no-store, must-revalidate"],"Content-Encoding":["gzip"],"Content-Length":["1309"],"Content-Type":["application/json"],"Date":["Fri, 07 Mar 2025 09:20:23 GMT"],"Expires":["Wed 24 Feb 1982 18:42:00 GMT"],"Warning":["299 - unknown field \"type\""],"X-Api-Cattle-Auth":["true"],"X-Api-Schemas":["https://192.168.122.122/v1/harvester/schemas"],"X-Content-Type-Options":["nosniff"]},"requestBody":{"type":"kubevirt.io.virtualmachine","metadata":{"namespace":"default","annotations":{"harvesterhci.io/volumeClaimTemplates":"[{\"metadata\":{\"name\":\"vm33-disk-0-yeeuj\",\"annotations\":{\"harvesterhci.io/imageId\":\"default/image-mfhfz\"}},\"spec\":{\"accessModes\":[\"ReadWriteMany\"],\"resources\":{\"requests\":{\"storage\":\"10Gi\"}},\"volumeMode\":\"Block\",\"storageClassName\":\"longhorn-image-mfhfz\"}}]","network.harvesterhci.io/ips":"[]"},"labels":{"harvesterhci.io/creator":"harvester","harvesterhci.io/os":"linux"},"name":"vm33"},"spec":{"runStrategy":"RerunOnFailure","template":{"metadata":{"annotations":{"harvesterhci.io/sshNames":"[]"},"labels":{"harvesterhci.io/vmName":"vm33"}},"spec":{"domain":{"machine":{"type":""},"cpu":{"cores":1,"sockets":1,"threads":1},"devices":{"inputs":[{"bus":"usb","name":"tablet","type":"tablet"}],"interfaces":[{"masquerade":{},"model":"virtio","name":"default"}],"disks":[{"name":"disk-0","disk":{"bus":"virtio"},"bootOrder":1},{"name":"cloudinitdisk","disk":{"bus":"virtio"}}]},"resources":{"limits":{"memory":"1Gi","cpu":"1"}},"features":{"acpi":{"enabled":true}}},"evictionStrategy":"LiveMigrateIfPossible","hostname":"vm33","networks":[{"name":"default","pod":{}}],"volumes":[{"name":"disk-0","persistentVolumeClaim":{"claimName":"vm33-disk-0-yeeuj"}},{"name":"cloudinitdisk","cloudInitNoCloud":{"secretRef":{"name":"vm33-zn2qw"},"networkDataSecretRef":{"name":"vm33-zn2qw"}}}],"affinity":{},"terminationGracePeriodSeconds":120,"accessCredentials":[]}}}}}


{"auditID":"5dcb7eb2-8d9d-4cb3-bb3f-dc1d607cf2a2","requestURI":"/v1/harvester/secrets/default","user":{"name":"user-vxq5m","group":["system:authenticated","system:cattle:authenticated"],"extra":{"principalid":["local://user-vxq5m"],"username":["admin"]}},"method":"POST","remoteAddr":"10.52.0.237:42344","requestTimestamp":"2025-03-07T09:20:23Z","responseTimestamp":"2025-03-07T09:20:23Z","responseCode":201,"requestHeader":{"Accept":["application/json"],"Accept-Encoding":["gzip, deflate, br, zstd"],"Accept-Language":["en-US,en;q=0.9"],"Content-Length":["369"],"Content-Type":["application/json"],"Dnt":["1"],"Origin":["https://192.168.122.122"],"Priority":["u=1, i"],"Referer":["https://192.168.122.122/dashboard/harvester/c/local/kubevirt.io.virtualmachine/create"],"Sec-Ch-Ua":["\"Not A(Brand\";v=\"8\", \"Chromium\";v=\"132\", \"Google Chrome\";v=\"132\""],"Sec-Ch-Ua-Mobile":["?0"],"Sec-Ch-Ua-Platform":["\"Linux\""],"Sec-Fetch-Dest":["empty"],"Sec-Fetch-Mode":["cors"],"Sec-Fetch-Site":["same-origin"],"User-Agent":["Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36"],"X-Api-Csrf":["1ebdb507946de829dc0b4feac754bb08"],"X-Forwarded-For":["192.168.122.145"],"X-Forwarded-Host":["192.168.122.122"],"X-Forwarded-Port":["443"],"X-Forwarded-Proto":["https"],"X-Forwarded-Scheme":["https"],"X-Real-Ip":["192.168.122.145"],"X-Request-Id":["cf677c47a6936cfab7d4a0ee75f2f21e"],"X-Scheme":["https"]},"responseHeader":{"Cache-Control":["no-cache, no-store, must-revalidate"],"Content-Encoding":["gzip"],"Content-Length":["649"],"Content-Type":["application/json"],"Date":["Fri, 07 Mar 2025 09:20:23 GMT"],"Expires":["Wed 24 Feb 1982 18:42:00 GMT"],"X-Api-Cattle-Auth":["true"],"X-Api-Schemas":["https://192.168.122.122/v1/harvester/schemas"],"X-Content-Type-Options":["nosniff"]},"requestBody":{"data":"[redacted]","metadata":{"labels":{"harvesterhci.io/cloud-init-template":"harvester"},"name":"vm33-zn2qw","namespace":"default"},"type":"secret"}}



The log is decoded as:


{
"auditID":"887f5e05-bd98-4ed9-9717-8c179f417236",
"requestURI":"/v1/harvester/kubevirt.io.virtualmachines/default",
"user":
{
"name":"user-vxq5m",
"group":["system:authenticated","system:cattle:authenticated"],
"extra":
{
"principalid":["local://user-vxq5m"],
"username":["admin"]
}
},

"method":"POST",
"remoteAddr":"10.52.0.237:42344",
"requestTimestamp":"2025-03-07T09:20:23Z",
"responseTimestamp":"2025-03-07T09:20:23Z",
"responseCode":201,

"requestHeader":
{
"Accept":["application/json"],
"Accept-Encoding":["gzip, deflate, br, zstd"],
"Accept-Language":["en-US,en;q=0.9"],
"Content-Length":["1513"],
"Content-Type":["application/json"],
"Dnt":["1"],
"Origin":["https://192.168.122.122"],
"Priority":["u=1, i"],
"Referer":["https://192.168.122.122/dashboard/harvester/c/local/kubevirt.io.virtualmachine/create"],
"Sec-Ch-Ua":["\"Not A(Brand\";v=\"8\", \"Chromium\";v=\"132\", \"Google Chrome\";v=\"132\""],
"Sec-Ch-Ua-Mobile":["?0"],
"Sec-Ch-Ua-Platform":["\"Linux\""],
"Sec-Fetch-Dest":["empty"],
"Sec-Fetch-Mode":["cors"],
"Sec-Fetch-Site":["same-origin"],
"User-Agent":["Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36"],
"X-Api-Csrf":["1ebdb507946de829dc0b4feac754bb08"],
"X-Forwarded-For":["192.168.122.145"],
"X-Forwarded-Host":["192.168.122.122"],
"X-Forwarded-Port":["443"],
"X-Forwarded-Proto":["https"],
"X-Forwarded-Scheme":["https"],
"X-Real-Ip":["192.168.122.145"],
"X-Request-Id":["0ca207d1a71ffa94cab23d5458284c8e"],
"X-Scheme":["https"]
},

"responseHeader":
{"Cache-Control":["no-cache, no-store, must-revalidate"],"Content-Encoding":["gzip"],"Content-Length":["1309"],"Content-Type":["application/json"],"Date":["Fri, 07 Mar 2025 09:20:23 GMT"],"Expires":["Wed 24 Feb 1982 18:42:00 GMT"],"Warning":["299 - unknown field \"type\""],"X-Api-Cattle-Auth":["true"],"X-Api-Schemas":["https://192.168.122.122/v1/harvester/schemas"],"X-Content-Type-Options":["nosniff"]
},

"requestBody":
{
"type":"kubevirt.io.virtualmachine",
"metadata":
{
"namespace":"default","annotations":{"harvesterhci.io/volumeClaimTemplates":"[{\"metadata\":{\"name\":\"vm33-disk-0-yeeuj\",\"annotations\":{\"harvesterhci.io/imageId\":\"default/image-mfhfz\"}},\"spec\":{\"accessModes\":[\"ReadWriteMany\"],\"resources\":{\"requests\":{\"storage\":\"10Gi\"}},\"volumeMode\":\"Block\",\"storageClassName\":\"longhorn-image-mfhfz\"}}]","network.harvesterhci.io/ips":"[]"},"labels":{"harvesterhci.io/creator":"harvester","harvesterhci.io/os":"linux"},"name":"vm33"
},
"spec":
{
"runStrategy":"RerunOnFailure","template":{"metadata":{"annotations":{"harvesterhci.io/sshNames":"[]"},"labels":{"harvesterhci.io/vmName":"vm33"}},"spec":{"domain":{"machine":{"type":""},"cpu":{"cores":1,"sockets":1,"threads":1},"devices":{"inputs":[{"bus":"usb","name":"tablet","type":"tablet"}],"interfaces":[{"masquerade":{},"model":"virtio","name":"default"}],"disks":[{"name":"disk-0","disk":{"bus":"virtio"},"bootOrder":1},{"name":"cloudinitdisk","disk":{"bus":"virtio"}}]},"resources":{"limits":{"memory":"1Gi","cpu":"1"}},"features":{"acpi":{"enabled":true}}},"evictionStrategy":"LiveMigrateIfPossible","hostname":"vm33","networks":[{"name":"default","pod":{}}],"volumes":[{"name":"disk-0","persistentVolumeClaim":{"claimName":"vm33-disk-0-yeeuj"}},{"name":"cloudinitdisk","cloudInitNoCloud":{"secretRef":{"name":"vm33-zn2qw"},"networkDataSecretRef":{"name":"vm33-zn2qw"}}}],"affinity":{},"terminationGracePeriodSeconds":120,"accessCredentials":[]}}
}
}

}

```

### Useful information pieces

```
"requestURI":"/v1/harvester/kubevirt.io.virtualmachines/default",
"user":
{
"name":"user-vxq5m",
"group":["system:authenticated","system:cattle:authenticated"],
"extra":
{
"principalid":["local://user-vxq5m"],       // local admin user when login to Harvester UI directly
"username":["admin"]
}
},

...
"X-Forwarded-For":["192.168.122.145"],      // source IP
"X-Forwarded-Host":["192.168.122.122"],     // Harvester cluster VIP

```

### Further work on Harvester

1. Enable Rancher audit log by default

2. Integrate with `rancher-logging` addon, add related rancher-audit deployment or clusterflow to scrap log from this file.

### Limitations

If you operate Harvester from the `Rancher Manager`, then the audit log is not saved in the embedded rancher POD, instead, the `Rancher Manager` is the right place to look.

```
 Rancher Manager ----------Harvester cluster rancher-agent pod------------Harveser pod----------api server

 user AAA & audit
```

## Get similar audit log from Rancher Manager

The `Rancher Manager` is generally deployed on an independent k8s cluster, it should be similar to enable it's audit log. (To be confirmed)
