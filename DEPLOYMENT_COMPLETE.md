# Complete Deployment Guide for Data Lakehouse on Gentoo

This guide covers deploying the complete Data Lakehouse solution on your Gentoo system with k3s.

## Prerequisites

Before starting the deployment, ensure you have:

### System Requirements
- Gentoo Linux with kernel supporting containers (namespaces, cgroups, overlayFS)
- At least 4GB RAM and 20GB free disk space
- Docker or containerd installed (k3s includes containerd)
- Git for version control
- Root privileges (sudo access) for k3s installation

### Software Installation
You need to install k3s first:
```bash
# Install k3s as root user
curl -sfL https://get.k3s.io | sh -

# Or if you are a regular user with sudo privileges:
curl -sfL https://get.k3s.io | INSTALL_K3S_SKIP_START=true sh -
sudo systemctl enable k3s
sudo systemctl start k3s

# Verify k3s is running
sudo systemctl status k3s
kubectl cluster-info
```

## Alternative: Using the Automated Script
If you prefer to use the automated script provided with this project:
```bash
# This will install k3s if not present and start it
sudo /home/juampa/UVG/DB1/PP3/start_k3s.sh
```

## Deployment Steps

### Step 1: Prepare the Project
```bash
# Navigate to project directory
cd /home/juampa/UVG/DB1/PP3

# Verify all necessary files exist
ls -la
```

### Step 2: Start the Web Interface (Optional - for file uploads)
```bash
# In a separate terminal, start the web interface
python3 start_web_interface.py

# This will open the upload interface at: http://localhost:8000/minimal_upload.html
```

### Step 3: Deploy to Kubernetes
```bash
# Run the deployment script
./deploy.sh
```

## Manual Deployment Steps (Alternative to deploy.sh)

If you prefer to deploy manually or the script doesn't work, follow these steps:

### 1. Apply Namespaces
```bash
kubectl apply -f k8s-manifests/namespaces.yaml
```

### 2. Apply Persistent Volumes
```bash
kubectl apply -f k8s-manifests/minio/pv-pvc.yaml
kubectl apply -f k8s-manifests/postgres/pv-pvc.yaml
```

### 3. Deploy MinIO
```bash
kubectl apply -f k8s-manifests/minio/
```

Wait for MinIO to be ready:
```bash
kubectl wait --for=condition=ready pod -l app=minio -n minio --timeout=120s
```

### 4. Deploy PostgreSQL
```bash
kubectl apply -f k8s-manifests/postgres/
```

Wait for PostgreSQL to be ready:
```bash
kubectl wait --for=condition=ready pod -l app=postgres -n postgres --timeout=180s
```

### 5. Build Container Images
```bash
# Build ETL Go image
cd etl-go
docker build -t datalakehouse-etl-go:latest .
cd ..

# Build R Report image
cd dockerfiles/r-report
docker build -t datalakehouse-r-report:latest .
cd ../../
```

### 6. Deploy ETL and Report Services
```bash
kubectl apply -f k8s-manifests/etl/
kubectl apply -f k8s-manifests/r-report/
```

## Verification

### Check All Deployments
```bash
# Check all namespaces
kubectl get namespaces

# Check pods in each namespace
kubectl get pods -n minio
kubectl get pods -n postgres
kubectl get pods -n etl
kubectl get pods -n r-report

# Check services
kubectl get services --all-namespaces

# Check persistent volumes
kubectl get pv,pvc --all-namespaces
```

### Access Services
- **MinIO API**: `http://YOUR_NODE_IP:30000`
- **MinIO Console**: `http://YOUR_NODE_IP:30001`
- **PostgreSQL**: `YOUR_NODE_IP:30432` (for Power BI/Tableau access)

## Configuration

### MinIO Access
Default credentials (from k8s manifests):
- Access Key: `minioadmin`
- Secret Key: `minioadmin`

### PostgreSQL Access
Default credentials (from k8s manifests):
- Username: `admin`
- Password: `password123`
- Database: `datalakehouse`

For Power BI/Tableau access:
- Username: `client_reader`
- Password: `reader_pass123`

## Testing the System

### 1. Upload Data
- Use the web interface at `http://localhost:8000/minimal_upload.html`
- Or upload directly to MinIO via the console

### 2. Verify Processing
```bash
# Check ETL logs
kubectl logs -l app=etl-go -n etl

# Check PostgreSQL tables
kubectl exec -it -n postgres deployment/postgres-deployment -- psql -U admin -d datalakehouse -c "\dt"
```

### 3. Test Visualization Connection
Connect Power BI or Tableau to:
- Server: `YOUR_NODE_IP:30432`
- Database: `datalakehouse`
- Username: `client_reader`
- Password: `reader_pass123`

## Troubleshooting

### Common Issues

1. **k3s not running**: Check if k3s service is active
```bash
sudo systemctl status k3s
# If not running: sudo systemctl start k3s
```

2. **Pods not starting**: Check if persistent volumes are properly created
```bash
kubectl get pv,pvc --all-namespaces
```

3. **MinIO not accessible**: Verify the NodePort service is working
```bash
kubectl get service minio-service -n minio
```

4. **PostgreSQL not accessible**: Check firewall settings and NodePort
```bash
kubectl get service postgres-service -n postgres
```

5. **ETL jobs failing**: Check logs and ensure MinIO and PostgreSQL are accessible
```bash
kubectl logs -l app=etl-go -n etl
```

### Useful Commands

```bash
# Check all resources
kubectl get all --all-namespaces

# Describe a specific resource for detailed info
kubectl describe pod <pod-name> -n <namespace>

# Get logs from a specific pod
kubectl logs <pod-name> -n <namespace>

# Execute commands inside a pod
kubectl exec -it <pod-name> -n <namespace> -- sh
```

## Scaling and Maintenance

### Updating Configurations
```bash
# After making changes to manifests, apply again
kubectl apply -f k8s-manifests/minio/
kubectl apply -f k8s-manifests/postgres/
```

### Checking Resource Usage
```bash
# Check resource usage
kubectl top nodes
kubectl top pods --all-namespaces
```

## Cleanup (if needed)
```bash
# To remove all deployments
kubectl delete -f k8s-manifests/
```

## Next Steps

1. Upload your business data files
2. Configure ETL CronJob schedule as needed
3. Set up email reporting with your SMTP settings
4. Connect visualization tools to the PostgreSQL database