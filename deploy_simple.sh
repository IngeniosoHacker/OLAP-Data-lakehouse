#!/bin/bash

# Simplified Data Lakehouse Deployment Script for k3s
# This script deploys a basic version considering the current disk pressure

set -e  # Exit on any error

echo "Starting simplified Data Lakehouse deployment..."

# Step 1: Apply namespaces
echo "1. Creating namespaces..."
k3s kubectl apply -f k8s-manifests/namespaces.yaml

# Step 2: Deploy services only (not StatefulSets yet)
echo "2. Deploying services only..."
k3s kubectl apply -f k8s-manifests/minio/service-nodeport.yaml
k3s kubectl apply -f k8s-manifests/minio/service-nodeport.yaml
k3s kubectl apply -f k8s-manifests/minio/statefulset.yaml  # This has tolerations

# Step 3: Build container images
echo "3. Building container images..."
cd etl-go
echo "Building ETL Go image..."
docker build -t datalakehouse-etl-go:latest .
cd ..

# For the R report container
cd dockerfiles/r-report
echo "Building R Report image..."
docker build -t datalakehouse-r-report:latest .
cd ../..

echo "Services deployed. The StatefulSets may take time to schedule due to disk pressure."
echo "Once the disk pressure is resolved, the pods should be scheduled."
echo ""
echo "To check status: k3s kubectl get pods --all-namespaces"
echo ""
echo "To resolve disk pressure, restart k3s service with: sudo systemctl restart k3s"
echo ""
echo "After disk pressure is resolved, run 'k3s kubectl get pods -n minio' to see MinIO pods."