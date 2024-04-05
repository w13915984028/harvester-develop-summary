# Test the Harvester AlertManager quickly

## Configure a simple http webhook server to print the received alerts

```
cat << 'EOF' > alert-webhookserver.py
import web
import json
from datetime import datetime

urls = ('/.*', 'hooks')

app = web.application(urls, globals())

def printAlert(jo):
	for x in jo["alerts"]:
		print(x)
		print()

class hooks:
    def POST(self):
        data = web.data().decode("utf-8")
        jo = json.loads(data)
        print('DATA RECEIVED: len:'+str(len(data)) +" :contains:"+str(len(jo["alerts"])) +" alerts: " +str(datetime.now()))
        printAlert(jo)
        print(jo)
        print()
        return 'OK'

if __name__ == '__main__':
    # usage
    # export PORT=8090  or any, default is 8080
    print("start a simple webhook server to receive alert " + str(datetime.now()))
    app.run()

EOF

export PORT=8090

$ python3 alert-webhookserver.py 
start a simple webhook server to receive alert 2022-08-29 19:39:25.847186
http://0.0.0.0:8090/
```

## Configure the Harvester cluster to send the alerts to the above server

Make sure, from inside the Harvester cluster, you can ping the above server's IP address.

```
cat << 'EOF' > a-single-receiver.yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: AlertmanagerConfig
metadata:
  name: amc-example
  # namespace: your value
  labels:
    alertmanagerConfig: example
spec:
  route:
    continue: true
    groupBy:
    - cluster
    - alertname
    receiver: "amc-webhook-receiver"
  receivers:
  - name: "amc-webhook-receiver"
    webhookConfigs:
    - sendResolved: true
      url: "http://192.168.122.159:8090/"
EOF

# kubectl apply -f a-single-receiver.yaml
alertmanagerconfig.monitoring.coreos.com/amc-example created

# kubectl get alertmanagerconfig -A
NAMESPACE   NAME          AGE
default     amc-example   27s

```

## The received alerts

Wait some minutes, you will receive alerts like:

```
DATA RECEIVED: len:3821 :contains:3 alerts: 2022-08-29 19:39:36.188128
{'status': 'firing', 'labels': {'alertname': 'LonghornVolumeStatusWarning', 'container': 'longhorn-manager', 'endpoint': 'manager', 'instance': '10.52.0.83:9500', 'issue': 'Longhorn volume is Degraded.', 'job': 'longhorn-backend', 'namespace': 'longhorn-system', 'node': 'harv2', 'pod': 'longhorn-manager-r5bgm', 'prometheus': 'cattle-monitoring-system/rancher-monitoring-prometheus', 'service': 'longhorn-backend', 'severity': 'warning', 'volume': 'pvc-1b835a32-fe0f-4b66-b1b3-4163f6c9f7b7'}, 'annotations': {'description': 'Longhorn volume is Degraded for more than 5 minutes.', 'runbook_url': 'https://longhorn.io/docs/1.3.0/monitoring/metrics/', 'summary': 'Longhorn volume is Degraded'}, 'startsAt': '2022-08-29T19:16:16.16Z', 'endsAt': '0001-01-01T00:00:00Z', 'generatorURL': 'https://192.168.122.200/api/v1/namespaces/cattle-monitoring-system/services/http:rancher-monitoring-prometheus:9090/proxy/graph?g0.expr=longhorn_volume_robustness+%3D%3D+2&g0.tab=1', 'fingerprint': '4312a97f9b9c46d1'}

{'status': 'firing', 'labels': {'alertname': 'LonghornVolumeStatusWarning', 'container': 'longhorn-manager', 'endpoint': 'manager', 'instance': '10.52.0.83:9500', 'issue': 'Longhorn volume is Degraded.', 'job': 'longhorn-backend', 'namespace': 'longhorn-system', 'node': 'harv2', 'pod': 'longhorn-manager-r5bgm', 'prometheus': 'cattle-monitoring-system/rancher-monitoring-prometheus', 'service': 'longhorn-backend', 'severity': 'warning', 'volume': 'pvc-5f602d48-c1e3-494d-9533-aa85f3970135'}, 'annotations': {'description': 'Longhorn volume is Degraded for more than 5 minutes.', 'runbook_url': 'https://longhorn.io/docs/1.3.0/monitoring/metrics/', 'summary': 'Longhorn volume is Degraded'}, 'startsAt': '2022-08-29T19:16:16.16Z', 'endsAt': '0001-01-01T00:00:00Z', 'generatorURL': 'https://192.168.122.200/api/v1/namespaces/cattle-monitoring-system/services/http:rancher-monitoring-prometheus:9090/proxy/graph?g0.expr=longhorn_volume_robustness+%3D%3D+2&g0.tab=1', 'fingerprint': '5c83b35d1e363d1e'}
```
