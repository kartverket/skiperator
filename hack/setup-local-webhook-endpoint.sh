#!/bin/bash

# For Rancher Desktop / Docker Desktop on Mac, use host.docker.internal
HOST_ADDR="host.docker.internal"

echo "Using host address: $HOST_ADDR"

# Get the IP that kind can reach
#HOST_IP=$(docker run --rm --network kind busybox nslookup host.docker.internal | grep "Address:" | tail -n1 | awk '{print $2}')
HOST_IP=192.168.5.2
if [ -z "$HOST_IP" ]; then
  echo "ERROR: Could not resolve host.docker.internal"
  echo "Falling back to Docker gateway IP..."
  HOST_IP=$(docker inspect skiperator-control-plane | jq -r '.[0].NetworkSettings.Networks.kind.Gateway')
fi

echo "Resolved IP: $HOST_IP"

echo "Deleting service..."
kubectl delete svc skipjob-conversion-webhook -n skiperator-system


echo "Re-Creating service..."
kubectl apply -f - <<EOF
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
kubectl apply -f - <<EOF
apiVersion: v1
kind: Endpoints
metadata:
  name: skipjob-conversion-webhook
  namespace: skiperator-system
subsets:
- addresses:
  - ip: ${HOST_IP}
  ports:
  - name: https
    port: 9443
    protocol: TCP
EOF

echo ""
echo "✅ Webhook routing configured:"
kubectl get endpoints skipjob-conversion-webhook -n skiperator-system
echo ""
echo "Service ClusterIP:"
kubectl get svc skipjob-conversion-webhook -n skiperator-system -o jsonpath='{.spec.clusterIP}'
echo " -> routes to -> ${HOST_IP}:9443"
echo ""
echo "Ready to test conversion webhook!"
