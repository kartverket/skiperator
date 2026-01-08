#!/bin/bash

CONTEXT="$1"

# For Rancher Desktop / Docker Desktop on Mac, use host.docker.internal
HOST_ADDR="host.docker.internal"

echo "Using host address: $HOST_ADDR"

# Get the IP that kind can reach
HOST_IP=$(docker run --rm --network kind busybox nslookup host.docker.internal | grep "Address:" | tail -n1 | awk '{print $2}')
if [ -z "$HOST_IP" ]; then
  echo "ERROR: Could not resolve host.docker.internal"
  echo "Falling back to Docker gateway IP..."
  HOST_IP=$(docker inspect skiperator-control-plane | jq -r '.[0].NetworkSettings.Networks.kind.Gateway')
fi

echo "Resolved IP: $HOST_IP"

echo "Deleting service..."
kubectl delete svc --context="$CONTEXT" skipjob-conversion-webhook -n skiperator-system


echo "Re-Creating service..."
kubectl apply --context="$CONTEXT" -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: skipjob-conversion-webhook
  namespace: skiperator-system
spec:
  ports:
  - name: https
    port: 443
    protocol: TCP
EOF


# Always update/create endpoints to point to current host IP
kubectl apply --context="$CONTEXT" -f - <<EOF
apiVersion: discovery.k8s.io/v1
kind: EndpointSlice
metadata:
  name: skipjob-conversion-webhook
  namespace: skiperator-system
  labels:
    kubernetes.io/service-name: skipjob-conversion-webhook
addressType: IPv4
endpoints:
- addresses:
  - ${HOST_IP}
ports:
- name: https
  port: 9443
  protocol: TCP
EOF

echo ""
echo "✅ Webhook routing configured:"
kubectl get endpointslice --context="$CONTEXT" skipjob-conversion-webhook -n skiperator-system
echo ""
echo "Service ClusterIP:"
kubectl get svc --context="$CONTEXT" skipjob-conversion-webhook -n skiperator-system -o jsonpath='{.spec.clusterIP}'
echo " -> routes to -> ${HOST_IP}:9443"
echo ""
echo "Ready to test conversion webhook!"
