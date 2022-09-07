
# Test k8s event in Harvester

Precondition: Harvester v1.1.0 is installed.

Config `ClusterFlow` and `ClusterOutput` to guide `fluentd` send collected `k8s event` to http web server.

Following 2 yamls will create output and flow to filter k8s events.

Note: Change `endpoint` to your webserver address.


```

cat > cop1.yaml << 'EOF'
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

```

```

cat > cflow1.yaml << 'EOF'
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

```


Apply yaml.

```
kubectl apply -f cop1.yaml

wait few seconds

kubectl apply -f cflow1.yaml

```
