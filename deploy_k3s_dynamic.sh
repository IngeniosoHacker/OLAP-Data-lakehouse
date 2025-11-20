#!/bin/bash

# Data Lakehouse Deployment Script for k3s with Dynamic Provisioning
# This script deploys the complete Data Lakehouse solution to k3s Kubernetes

set -e  # Exit on any error

echo "Starting Data Lakehouse deployment with k3s (dynamic provisioning)..."

# Step 1: Apply namespaces
echo "1. Creating namespaces..."
k3s kubectl apply -f k8s-manifests/namespaces.yaml

# Step 2: Deploy MinIO
echo "2. Deploying MinIO..."
k3s kubectl apply -f k8s-manifests/minio/

# Wait for MinIO to be ready (with longer timeout)
echo "Waiting for MinIO to be ready..."
k3s kubectl wait --for=condition=ready pod -l app=minio -n minio --timeout=300s

# Step 3: Deploy PostgreSQL
echo "3. Deploying PostgreSQL..."
k3s kubectl apply -f k8s-manifests/postgres/

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
k3s kubectl wait --for=condition=ready pod -l app=postgres -n postgres --timeout=300s

# Step 4: Build and push container images (you'll need to do this manually)
echo "4. Building container images..."
cd etl-go
echo "Building ETL Go image..."
docker build -t datalakehouse-etl-go:latest .
cd ..

# For the R report container
cd dockerfiles/r-report
echo "Building R Report image..."
docker build -t datalakehouse-r-report:latest .
cd ../..

# Step 5: Deploy ETL CronJob
echo "5. Deploying ETL CronJob..."
k3s kubectl apply -f k8s-manifests/etl/

# Step 6: Deploy R Report CronJob
echo "6. Deploying R Report CronJob..."
k3s kubectl apply -f k8s-manifests/r-report/

# Step 7: Verify all deployments
echo "7. Verifying deployments..."

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