# Testing the Data Lakehouse System

This document provides comprehensive instructions for testing the Data Lakehouse system with your files.

## Table of Contents
1. [System Overview](#system-overview)
2. [Prerequisites](#prerequisites)
3. [Testing Methods](#testing-methods)
4. [Web Interface Testing](#web-interface-testing)
5. [Command Line Testing](#command-line-testing)
6. [Verifying Results](#verifying-results)
7. [Troubleshooting](#troubleshooting)

## System Overview

The Data Lakehouse system automatically:
- Accepts various file formats (CSV, JSON, XLSX, XLS, SQL dumps, etc.)
- Creates appropriate database tables based on detected schemas
- Identifies dimensions and facts to create star schema views
- Processes data through the ELT pipeline

## Prerequisites

Before testing, ensure you have:

### For Web Interface Testing:
- A modern web browser (Chrome, Firefox, Safari, Edge)
- Python 3 installed for the local web server
- Access to the MinIO server (endpoint, credentials)

### For Command Line Testing:
- Go programming language installed (1.21+)
- Docker and Docker Compose (for local testing)
- Access to PostgreSQL and MinIO services

### For Kubernetes System Testing:
- Gentoo Linux with kernel supporting containers (namespaces, cgroups, overlayFS)
- Root privileges (sudo access) for k3s installation
- k3s installed and running
- kubectl configured for cluster access

## Testing Methods

### Method 1: Web Interface (Recommended for Business Users)

1. **Access the Web Interface**
   - Open a web browser and navigate to the upload page
   - If running locally: `http://localhost:8080` (or your server URL)
   - If using the file directly, open `minimal_upload.html` in your browser

2. **Configure MinIO Connection**
   - MinIO Endpoint: Your MinIO service URL (e.g., `http://localhost:9000`)
   - Access Key: Your MinIO access key
   - Secret Key: Your MinIO secret key
   - Bucket Name: The bucket where files should be stored (typically `raw`)

3. **Upload Files**
   - Drag and drop files into the upload area, or click "Browse Files"
   - Supported formats: CSV, JSON, XLSX, XLS, SQL dumps, TXT, XML, Parquet
   - Select multiple files at once if needed

4. **Start Upload Process**
   - Click "Upload Files to Data Lakehouse"
   - Monitor progress in the file list
   - Wait for all files to show "Uploaded" status

### Method 2: Command Line Interface

1. **Prepare Environment Variables**
   ```bash
   export MINIO_ENDPOINT="your-minio-endpoint"
   export MINIO_ACCESS_KEY="your-access-key"
   export MINIO_SECRET_KEY="your-secret-key"
   export DB_NAME="datalakehouse"
   export DB_USER="admin"
   export DB_PASSWORD="password123"
   export DB_HOST="localhost"
   ```

2. **Process Files Using ETL**
   ```bash
   # Set source type to file and specify file path
   export ETL_SOURCE_TYPE="file"
   export ETL_SOURCE_FILE="/path/to/your/datafile.csv"
   
   # Run the ETL process
   cd etl-go
   go run main.go
   ```

### Method 3: Docker Compose for Local Testing

1. **Start the Services**
   ```bash
   cd /path/to/your/project
   docker-compose up -d
   ```

2. **Upload Files Using Web Interface**
   - Access MinIO Console at `http://localhost:9001`
   - Create a bucket named `raw`
   - Upload your data files to this bucket
   - Or use the web interface mentioned in Method 1

## Verifying Results

### Check MinIO for Raw Files
1. Access MinIO Console (typically at port 9001)
2. Navigate to the `raw` bucket
3. Verify your files are present

### Check PostgreSQL Database
1. Connect to the database:
   ```bash
   psql -h localhost -p 5432 -U admin -d datalakehouse
   ```

2. Check for generated tables:
   ```sql
   -- List all tables
   \dt
   
   -- Look for auto-generated tables
   SELECT table_name FROM information_schema.tables 
   WHERE table_name LIKE 'auto_table_%';
   
   -- View star schema views
   \dv
   ```

3. Verify data:
   ```sql
   -- Look at a generated table (replace with actual table name)
   SELECT * FROM auto_table_test LIMIT 10;
   
   -- Look at the star schema view
   SELECT * FROM sales_summary LIMIT 10;
   ```

### Check ETL Processing Logs
1. View ETL service logs:
   ```bash
   kubectl logs -f deployment/etl-go  # For Kubernetes
   # Or
   docker logs datalakehouse-postgres  # For Docker
   ```

## Testing with Sample Data

### Sample CSV File (sales_data.csv)
```csv
customer_id,customer_name,product_id,product_name,quantity,unit_price,total_amount
1,John Doe,101,Laptop,2,1000.00,2000.00
2,Jane Smith,102,Mouse,5,25.00,125.00
3,Bob Johnson,103,Keyboard,3,75.00,225.00
```

### Sample JSON File (sales_data.json)
```json
[
  {
    "customer_id": 1,
    "customer_name": "John Doe",
    "product_id": 101,
    "product_name": "Laptop",
    "quantity": 2,
    "unit_price": 1000.00,
    "total_amount": 2000.00
  },
  {
    "customer_id": 2,
    "customer_name": "Jane Smith",
    "product_id": 102,
    "product_name": "Mouse",
    "quantity": 5,
    "unit_price": 25.00,
    "total_amount": 125.00
  }
]
```

## Business Data Compatibility

The system is designed to work with any business data structure:

### Retail Industry
- Sales transactions, customer data, product catalogs
- Files: sales.csv, customers.json, products.xlsx

### Healthcare Industry  
- Patient records, treatment data, provider information
- Files: patients.csv, treatments.json, providers.xlsx

### Finance Industry
- Account transactions, security data, portfolio information
- Files: transactions.csv, securities.json, portfolios.xlsx

### Manufacturing Industry
- Product specifications, supplier data, inventory records
- Files: products.csv, suppliers.json, inventory.xlsx

## Troubleshooting

### File Upload Issues
- **Problem**: Files not uploading to MinIO
- **Solution**: Verify MinIO endpoint, credentials, and CORS settings

- **Problem**: Unsupported file format
- **Solution**: Ensure file format is one of: CSV, JSON, XLSX, XLS, SQL dump

### Database Issues
- **Problem**: Tables not created automatically
- **Solution**: Check ETL service logs for errors

- **Problem**: Incorrect data types detected
- **Solution**: The system does its best to infer types, but complex data may need manual verification

### Connection Issues
- **Problem**: Cannot connect to MinIO/PostgreSQL
- **Solution**: Verify network connectivity and credentials

### Performance Issues
- Large files may take time to process
- System automatically samples data for schema detection
- Processing time depends on file size and complexity

## Next Steps After Successful Testing

1. **Connect Visualization Tools**
   - Use the connection details provided in `SHARING_DATABASE_ACCESS.md`
   - Connect Power BI, Tableau, or other tools to the PostgreSQL database

2. **Schedule Regular Processing**
   - Set up cron jobs or Kubernetes CronJobs for automated processing
   - Configure email reports if needed

3. **Monitor and Optimize**
   - Monitor system performance
   - Adjust configurations based on your data volume and processing needs