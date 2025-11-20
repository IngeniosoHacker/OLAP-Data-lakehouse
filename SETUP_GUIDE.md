# Data Lakehouse Project - Setup and Configuration Guide

This document provides detailed information about the setup and configuration of the Data Lakehouse project that you have successfully deployed.

## Project Components

The Data Lakehouse solution consists of the following components:

1. **PostgreSQL (Data Warehouse)** - On port 5433
2. **MinIO (Data Lake)** - On ports 9002 (API) and 9003 (Console)
3. **ETL Service (Go-based)** - In progress (has dependency issues)
4. **Web Interface** - On port 8080

## Current Working Services

### PostgreSQL Database
- Port: `5433` (instead of default 5432 to avoid conflicts)
- Database: `datalakehouse`
- User: `admin`
- Password: `password123`
- Connection string: `localhost:5433`

### MinIO Object Storage
- API Port: `9002`
- Console Port: `9003`
- Access Key: `minioadmin`
- Secret Key: `minioadmin`
- Console URL: `http://localhost:9003`

### Web Interface
- Port: `8080`
- URL: `http://localhost:8080/minimal_upload.html`
- Default MinIO endpoint in form is set to `http://localhost:9002`

## Service Status

✅ **PostgreSQL**: Running and accessible  
✅ **MinIO**: Running and accessible  
✅ **Web Interface**: Running and file uploads functional  
❌ **ETL Service**: Not running due to dependency issues  

## How to Test the System

### 1. File Upload
1. Open the web interface at `http://localhost:8080/minimal_upload.html`
2. The MinIO configuration should have default values for your setup
3. Upload a test file (CSV, JSON, XLSX, etc.)
4. Files will be saved locally in the `uploads/` directory

### 2. Check PostgreSQL Database
```bash
PGPASSWORD=password123 psql -h localhost -p 5433 -U admin -d datalakehouse -c "\dt"
```

### 3. Check MinIO Console
1. Go to `http://localhost:9003`
2. Login with `minioadmin` / `minioadmin`
3. Check the `raw` bucket for uploaded files

## Known Issues & Limitations

### ETL Service Build Issue
The ETL-Go service has a dependency issue with `github.com/extrame/gofile` that prevents it from building. This affects the complete pipeline that would move files from the upload location to MinIO and PostgreSQL automatically.

### Port Conflicts
The services were configured to use non-standard ports (5433, 9002, 9003) to avoid conflicts with existing containers on your system.

## Troubleshooting

### If services are not accessible
1. Check if they're running:
   ```bash
   cd /home/juampa/UVG/DB1/PP3 && docker compose ps
   ```

2. Check logs for any errors:
   ```bash
   cd /home/juampa/UVG/DB1/PP3 && docker compose logs
   ```

### To restart services
```bash
cd /home/juampa/UVG/DB1/PP3 && docker compose down
cd /home/juampa/UVG/DB1/PP3 && docker compose up -d
```

### To restart the web interface
```bash
pkill -f start_web_interface.py
cd /home/juampa/UVG/DB1/PP3 && python3 start_web_interface.py
```

## Project Files Created/Modified

The following files were created or modified during setup:

- `etl-go/Dockerfile` - Fixed to include git dependency
- `docker-compose.yml` - Updated port mappings and removed problematic volume mount
- `web-upload/minimal_upload.html` - Updated MinIO endpoint default value
- `start_web_interface.py` - Enhanced to handle file uploads
- `uploads/` directory - Created to store uploaded files
- `SETUP_GUIDE.md` - This documentation

## Next Steps

1. **Fix ETL Service**: Resolve the Go dependency issue to enable full pipeline processing
2. **Test Complete Pipeline**: Once ETL service is running, file uploads will automatically process to PostgreSQL and MinIO
3. **Connect BI Tools**: Configure Power BI or Tableau to connect to PostgreSQL at localhost:5433
4. **Configure Email Service**: Set up email automation using `email-config.yaml`

## Architecture Overview

```
[Web Interface] -> [Uploaded Files] -> [MinIO (Data Lake)] -> [PostgreSQL (Data Warehouse)]
     (8080)           (uploads/)           (9002/9003)              (5433)
                        |
                        v
                [ETL Service (pending)]
```

## Access Credentials Summary

| Service | Host | Port | Username | Password |
|---------|------|------|----------|----------|
| PostgreSQL | localhost | 5433 | admin | password123 |
| MinIO API | localhost | 9002 | minioadmin | minioadmin |
| MinIO Console | localhost | 9003 | minioadmin | minioadmin |

## Notes

- The system is designed to work with CSV, JSON, XLSX, XLS, SQL dumps, and other data formats
- The web interface allows drag-and-drop file uploads
- The system automatically detects column types and creates appropriate table schemas
- Star schema views are generated for analytical queries
- The project is built to work with Data Lakehouse architecture combining data lake and data warehouse benefits