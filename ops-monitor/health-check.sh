#!/bin/sh
set -eu

echo "=== Dhaba Orders Cluster Health ==="
kubectl get nodes
kubectl get pods -A
echo "=== Staging Services ==="
kubectl get services -n staging
