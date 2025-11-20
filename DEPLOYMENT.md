# Data Lakehouse Solution

This project implements a Data Lakehouse architecture on Gentoo Linux using Kubernetes (k3s), MinIO, PostgreSQL, Go ETL, and R reporting.

## Architecture Overview

The system consists of:

1. **MinIO** - Data lake for storing raw and processed files
2. **PostgreSQL** - Data warehouse with star schema
3. **Go ETL** - Extract, Transform, Load process running as CronJob
4. **R Reports** - Weekly report generation and email delivery
5. **Power BI/Tableau** - Client access to data warehouse

## Project Structure

```
├── k8s-manifests/           # Kubernetes manifests
│   ├── namespaces.yaml      # Namespaces definition
│   ├── minio/              # MinIO manifests
│   ├── postgres/           # PostgreSQL manifests
│   ├── etl/                # ETL CronJob manifests
│   └── r-report/           # R reporting manifests
├── etl-go/                 # Go ETL application
├── r-scripts/              # R reporting scripts
├── dockerfiles/            # Dockerfiles for containers
└── connection-guide.md     # Client connection documentation
```

## Deployment Instructions

1. Set up k3s cluster on Gentoo (ensure prerequisites from original README are met)
2. Create namespaces:
   ```bash
   kubectl apply -f k8s-manifests/namespaces.yaml
   ```

3. Apply PV/PVC:
   ```bash
   kubectl apply -f k8s-manifests/minio/pv-pvc.yaml
   kubectl apply -f k8s-manifests/postgres/pv-pvc.yaml
   ```

4. Deploy MinIO:
   ```bash
   kubectl apply -f k8s-manifests/minio/
   ```

5. Deploy PostgreSQL:
   ```bash
   kubectl apply -f k8s-manifests/postgres/
   ```

6. Deploy ETL CronJob:
   ```bash
   kubectl apply -f k8s-manifests/etl/
   ```

7. Deploy R Report CronJob:
   ```bash
   kubectl apply -f k8s-manifests/r-report/
   ```

## Client Access

Power BI and Tableau clients can connect to the PostgreSQL data warehouse using the connection details in `connection-guide.md`.

## Data Flow

```
CSV/JSON Sources → ETL Pod → MinIO (raw) → Transformation → PostgreSQL (DW) → Views → Clients
Report Generation → R Pod → PDF Reports → Email
```

## Configuration

- MinIO: Access via node IP on port 30000 (API) and 30001 (Console)
- PostgreSQL: Access via node IP on port 30432
- Default credentials are in the manifest files (should be changed for production)

## Features

- Supports both file and SQL data sources (Data Lakehouse)
- Automated ETL processing via CronJob
- Weekly report generation with PDF output
- Read-only user for client access
- Star schema in data warehouse