

**harvester-cloud-provider** chart is shipped to RKE2 and binded to the KDM file, each RKE2 version has a default HCP chart version.

Before the chart is released to upstream Rancher, RKE2, we need to do some local tests.

Follow steps show how to.

## Prepare local chart

1. Download a chart to local.

e.g. https://github.com/harvester/charts/releases/download/harvester-cloud-provider-0.2.12-rc2/harvester-cloud-provider-0.2.12-rc2.tgz

2. Edit chart to have your changes.

Change Chart.yaml to have a chart which meets RKE2 HCP chart nameing.

```
version: 0.2.12-rc2
```

to

```
version: 0.2.1300
```

Then package the chart.

```
helm package harvester-cloud-provider/
Successfully packaged chart and saved it to: /home/jianwang/develop/rke2-charts/harvester-chart/harvester-cloud-provider-0.2.1300.tgz
```

Run `helm repo index harvester-chart` to generate chart index.

And serve chart.

```
~/develop/rke2-charts/harvester-chart$ python3 -m http.server 8092
Serving HTTP on 0.0.0.0 port 8092 (http://0.0.0.0:8092/) ...
...
192.168.122.118 - - [19/Jun/2026 13:19:34] "GET /index.yaml HTTP/1.1" 200 -  // when following step `add repo` is done
```

## Add local repository to Rancher Manager

Prepare a local chart repository and then add it to Rancher UI

1. Go to **Rancher UI** -> **Cluster Management** -> **Repositories**

```
Active harvester	git	https://github.com/harvester/harvester-ui-extension	gh-pages	10 days	
Active harvester-charts-git	git	https://github.com/harvester/charts.git	gh-pages	9 days	
Active Partners	git	https://git.rancher.io/partner-charts	main	10 days	
Active Rancher	git	https://git.rancher.io/charts	release-v2.14	10 days	
Active RKE2	git	https://git.rancher.io/rke2-charts	main	10 days	

```

1. Add new one:

name: `my-repo`, index URL `http://192.168.2.59:8092`

## Customize KDM file and serve

```

{
 "K8sVersionServiceOptions": null,
 ...
 "rke2": {
  "appDefaults": []
  "channels": []

  "releases": [
    "charts": {
     "harvester-cloud-provider": {
      "repo": "my-repo",
      "version": "0.2.1300"
     },
     ...

    },
    "version": "v1.35.4+rke2r1"

  ]

 }
}

```

e.g. It is served on path `http://192.168.2.59:8000/rancher-v2.14-kdm-data-updated-with-hcp.json`.

```
/home/jianwang/develop/rke2-charts
~/develop/rke2-charts$ python3 -m http.server
Serving HTTP on 0.0.0.0 port 8000 (http://0.0.0.0:8000/) ...

```

## Set KDM file

1. Go to **Rancher UI** -> **Global Settings** -> **Settings** -> **rke-metadata-config**

The default value:

```
{
  "refresh-interval-minutes": "1440",
  "url": "https://releases.rancher.com/kontainer-driver-metadata/release-v2.14/data.json"
}
```

Set it with customized value:

```
{
  "refresh-interval-minutes": "1440",
  "url": "http://192.168.2.59:8000/rancher-v2.14-kdm-data-updated-with-hcp.json"
}
```

Following logs show that Rancher Manager has fetched the KDM file from local.

```
Serving HTTP on 0.0.0.0 port 8000 (http://0.0.0.0:8000/) ...
192.168.122.118 - - [19/Jun/2026 12:57:24] "GET /rancher-v2.14-kdm-data-updated-with-hcp.json HTTP/1.1" 200 -
192.168.122.118 - - [19/Jun/2026 12:57:24] "GET /rancher-v2.14-kdm-data-updated-with-hcp.json HTTP/1.1" 200 -
192.168.122.118 - - [19/Jun/2026 12:57:24] "GET /rancher-v2.14-kdm-data-updated-with-hcp.json HTTP/1.1" 200 -
192.168.122.118 - - [19/Jun/2026 12:57:24] "GET /rancher-v2.14-kdm-data-updated-with-hcp.json HTTP/1.1" 200 -
192.168.122.118 - - [19/Jun/2026 12:57:24] "GET /rancher-v2.14-kdm-data-updated-with-hcp.json HTTP/1.1" 200 -
```

:::note

This step ensures the new chart could be customized when you click the **Add-on: Harvester Cloud Provider** page when create guest cluster. New options mentioned on `question.yaml` will be mapped to UI directly.

:::

## Validation

When create Harvester based guest cluster from Rancher UI, select the cusotimized RKE2 version with local HCP version.

### Add following content to  `Additional Manifest`

On guest cluster creation page, go to **Cluster Configuration** and click **Additional Manifest**, then post following content, it guides RKE2 to use local chart to bootstrap the cluster. Without this step, even though Rancher UI shows configuration options of new HCP chart, RKE2 still uses the old chart to bootstrap guest cluster.

```
apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  name: harvester-cloud-provider
  namespace: kube-system
spec:
  repo: http://192.168.2.59:8092
  chart: harvester-cloud-provider
  version: 0.2.1300
  bootstrap: true
  valuesContent: '{"cloudConfigPath":"/var/lib/rancher/rke2/etc/config-files/cloud-provider-config","global":{"cattle":{"clusterName":"gc5"}}}'
```

:::important

`bootstrap: true` is critical for HCP chart, without it, the helm-install job can't be scheduled.

:::


The chart repo will show such log:

```
Serving HTTP on 0.0.0.0 port 8092 (http://0.0.0.0:8092/) ...
192.168.122.118 - - [19/Jun/2026 13:19:34] "GET /index.yaml HTTP/1.1" 200 -
192.168.122.118 - - [19/Jun/2026 13:21:30] "GET /harvester-cloud-provider-0.2.1300.tgz HTTP/1.1" 200 -
```


check guest cluster log:

```

oot@gc5-pool1-kp748-jxf7q:~# kubectl logs -n kube-system helm-install-harvester-cloud-provider-pgkxn
if [[ ${KUBERNETES_SERVICE_HOST} =~ .*:.* ]]; then
	echo "KUBERNETES_SERVICE_HOST is using IPv6"
	CHART="${CHART//%\{KUBERNETES_API\}%/[${KUBERNETES_SERVICE_HOST}]:${KUBERNETES_SERVICE_PORT}}"
else
	CHART="${CHART//%\{KUBERNETES_API\}%/${KUBERNETES_SERVICE_HOST}:${KUBERNETES_SERVICE_PORT}}"
fi

set +v -x
+ [[ '' == v2 ]]
+ shopt -s nullglob
+ [[ -f /config/ca-file.pem ]]
+ [[ -f /tmp/ca-file.pem ]]
+ [[ false == true ]]
+ [[ false == true ]]
+ [[ -n '' ]]
+ helm_content_decode
+ set -e
+ ENC_CHART_PATH=/chart/harvester-cloud-provider.tgz.base64
+ CHART_PATH=/tmp/harvester-cloud-provider.tgz
+ [[ ! -f /chart/harvester-cloud-provider.tgz.base64 ]]
+ return
+ [[ install != delete ]]
+ helm_repo_init
+ grep -q -e 'https\?://'
+ [[ harvester-cloud-provider/harvester-cloud-provider == stable/* ]]
+ [[ -n http://192.168.2.59:8092 ]]
+ [[ -f /auth/username ]]
+ [[ -f /auth/tls.crt ]]
+ helm repo add harvester-cloud-provider http://192.168.2.59:8092
"harvester-cloud-provider" has been added to your repositories
+ helm repo update
Hang tight while we grab the latest from your chart repositories...
...Successfully got an update from the "harvester-cloud-provider" chart repository
Update Complete. ⎈Happy Helming!⎈
+ helm_update install --version 0.2.1300
++ helm ls --all -f '^harvester-cloud-provider$' --namespace kube-system --output ++ jq -r '"\(.[0].chart),\(.[0].status),\(.[0].revision)"'
json
++ tr '[:upper:]' '[:lower:]'
+ LINE=null,null,null
+ IFS=,
+ read -r INSTALLED_VERSION STATUS REVISION _
+ VALUES=
+ for VALUES_FILE in /config/*.yaml
+ VALUES=' --values /config/values-0-000-HelmChart-ValuesContent.yaml'
+ for VALUES_FILE in /config/*.yaml
+ VALUES=' --values /config/values-0-000-HelmChart-ValuesContent.yaml --values /config/values-1-000-HelmChartConfig-ValuesContent.yaml'
+ [[ install = delete ]]
+ [[ null =~ ^(|null)$ ]]
+ [[ null =~ ^(|null)$ ]]
+ echo 'Installing helm chart'
+ helm install --version 0.2.1300 harvester-cloud-provider harvester-cloud-provider/harvester-cloud-provider --values /config/values-0-000-HelmChart-ValuesContent.yaml --values /config/values-1-000-HelmChartConfig-ValuesContent.yaml
NAME: harvester-cloud-provider
LAST DEPLOYED: Fri Jun 19 12:47:58 2026
NAMESPACE: kube-system
STATUS: deployed
REVISION: 1
TEST SUITE: None
+ exit
```