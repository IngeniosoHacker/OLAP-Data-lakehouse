# Sharing Database Access for Visualization Tools

This document outlines the different methods to share database access with other users who need to connect visualization tools (like Power BI or Tableau) to your Data Lakehouse system.

## Option 1: Docker Image (Recommended for Development/Testing)

### Quick Start
```bash
# Clone the repository
git clone <your-repo-url>
cd <repo-directory>

# Start the database container
docker-compose up -d

# Wait for the database to be ready (takes about 30 seconds)
```

### Connection Details
- **Server**: `localhost:5432`
- **Database**: `datalakehouse`
- **Username**: `client_reader` (read-only)
- **Password**: `reader_pass123`

### For visualization tools:
- Power BI: PostgreSQL database → Server: localhost, Port: 5432, Database: datalakehouse
- Tableau: PostgreSQL → Server: localhost, Port: 5432, Database: datalakehouse

## Option 2: Database Dump File

### Export Current Database
```bash
# Build the dump utility
cd etl-go/cmd/db-dump
go build -o db-dump .

# Run the dump utility (set environment variables first)
export DB_HOST=localhost
export DB_NAME=datalakehouse
export DB_USER=admin
export DB_PASSWORD=password123
export OUTPUT_FILE=database_dump.sql

./db-dump
```

### Import Database Dump
```bash
# Create a new database
createdb -U postgres -h localhost visualization_db

# Import the dump
psql -U postgres -h localhost -d visualization_db -f database_dump.sql
```

### Connection Details for Imported Database
- **Server**: Your PostgreSQL server
- **Database**: `visualization_db` (or whatever you named it)
- **Username**: Your PostgreSQL username
- **Password**: Your PostgreSQL password

## Option 3: Direct Access to Your Kubernetes Cluster

### Prerequisites
- Access to the Kubernetes cluster where your Data Lakehouse is deployed
- `kubectl` installed and configured

### Connection Details
- **Server**: `<your-cluster-node-ip>:30432` (or the NodePort configured for PostgreSQL)
- **Database**: `datalakehouse`
- **Username**: `client_reader` (read-only)
- **Password**: `reader_pass123`

## Available Views and Tables

### Star Schema Views
- `sales_summary` - Main view with sales data joined across dimensions

### Dimension Tables
- `dim_customer` - Customer information
- `dim_product` - Product information
- `dim_location` - Location information
- `dim_date` - Date dimension

### Fact Tables
- `fact_sales` - Core sales transactions
- `auto_table_test` - Example of automatically created table

## Power BI Connection Steps

1. Open Power BI Desktop
2. Click "Get Data" → "Database" → "PostgreSQL database"
3. Enter server details:
   - Server: `localhost` (for Docker) or your cluster IP
   - Database: `datalakehouse`
4. Click "OK" and then select "Database" authentication
5. Enter:
   - Username: `client_reader`
   - Password: `reader_pass123`
6. Select the views/tables you need

## Tableau Connection Steps

1. Open Tableau Desktop
2. Select "PostgreSQL" from the connections
3. Enter server details:
   - Server: `localhost` (for Docker) or your cluster IP
   - Port: `5432`
   - Database: `datalakehouse`
4. Authentication: Username and Password
5. Enter:
   - Username: `client_reader`
   - Password: `reader_pass123`
6. Connect to the required views/tables

## Docker Compose Management

### Start the Services
```bash
docker-compose up -d
```

### Check Service Status
```bash
docker-compose ps
```

### View Logs
```bash
docker-compose logs postgres
```

### Stop the Services
```bash
docker-compose down
```

### Stop and Remove Volumes (Data will be lost)
```bash
docker-compose down -v
```

## Troubleshooting

### Connection Issues
- Ensure the PostgreSQL service is running: `docker-compose ps`
- Check firewall settings if connecting from another machine
- Verify the connection details match your setup

### Docker Issues
- Make sure Docker and Docker Compose are installed and running
- Check available disk space for Docker volumes

### Sample Data
The database comes pre-populated with sample data, including:
- 4 customers
- 5 products
- 4 locations
- 12 dates
- 10 sales records
- Example auto-generated table

## Security Considerations

- The `client_reader` user has read-only access to all tables and views
- Never use admin credentials for visualization tools
- Change default passwords in production environments
- Consider using connection encryption (SSL) in production

## For Production Use

### Building the PostgreSQL Docker Image
```bash
# Navigate to the Dockerfile location
cd dockerfiles/postgres

# Build the image
docker build -t datalakehouse-postgres:latest .
```

### Sharing the Image
```bash
# Tag for sharing
docker tag datalakehouse-postgres:latest <your-registry>/datalakehouse-postgres:latest

# Push to registry
docker push <your-registry>/datalakehouse-postgres:latest
```

The recipient can then pull and run:
```bash
docker run -d -p 5432:5432 --name datalakehouse-postgres datalakehouse-postgres:latest
```