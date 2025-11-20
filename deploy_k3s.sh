#!/bin/bash

# Data Lakehouse Deployment Script for k3s
# This script deploys the complete Data Lakehouse solution to k3s Kubernetes

set -e  # Exit on any error

echo "Starting Data Lakehouse deployment with k3s..."

# Step 1: Apply namespaces
echo "1. Creating namespaces..."
k3s kubectl apply -f k8s-manifests/namespaces.yaml

# Step 2: Apply persistent volumes and claims
echo "2. Creating persistent volumes and claims..."
k3s kubectl apply -f k8s-manifests/minio/pv-pvc.yaml
k3s kubectl apply -f k8s-manifests/postgres/pv-pvc.yaml

# Step 3: Deploy MinIO
echo "3. Deploying MinIO..."
k3s kubectl apply -f k8s-manifests/minio/

# Wait for MinIO to be ready
echo "Waiting for MinIO to be ready..."
k3s kubectl wait --for=condition=ready pod -l app=minio -n minio --timeout=120s

# Step 4: Deploy PostgreSQL
echo "4. Deploying PostgreSQL..."
k3s kubectl apply -f k8s-manifests/postgres/

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
k3s kubectl wait --for=condition=ready pod -l app=postgres -n postgres --timeout=180s

# Step 5: Build and push container images (you'll need to do this manually)
echo "5. Building container images..."
cd etl-go
echo "Building ETL Go image..."
docker build -t datalakehouse-etl-go:latest .
cd ..

# For the R report container
cd dockerfiles/r-report
echo "Building R Report image..."
docker build -t datalakehouse-r-report:latest .
cd ../..

# Step 6: Deploy ETL CronJob
echo "6. Deploying ETL CronJob..."
k3s kubectl apply -f k8s-manifests/etl/

# Step 7: Deploy R Report CronJob
echo "7. Deploying R Report CronJob..."
k3s kubectl apply -f k8s-manifests/r-report/

# Step 8: Verify all deployments
echo "8. Verifying deployments..."

echo "MinIO pods:"
k3s kubectl get pods -n minio

echo "PostgreSQL pods:"
k3s kubectl get pods -n postgres

echo "ETL pods:"
k3s kubectl get jobs -n etl

echo "R Report pods:"
k3s kubectl get jobs -n r-report

echo "Services:"
k3s kubectl get services --all-namespaces

echo ""
echo "Data Lakehouse deployment completed!"
echo ""
echo "Access information:"
echo "MinIO API: http://NODE_IP:30000"
echo "MinIO Console: http://NODE_IP:30001"
echo "PostgreSQL: NODE_IP:30432 (for Power BI/Tableau access)"
echo ""
echo "To access the web interface for file uploads, start the Python server:"
echo "python3 start_web_interface.py"
echo ""
echo "Then configure the MinIO settings in the web interface to point to your cluster."