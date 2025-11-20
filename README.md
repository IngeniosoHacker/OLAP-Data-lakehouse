# Data Lakehouse Solution - Complete Documentation

## Overview

This project implements a **Data Lakehouse architecture** that combines the flexibility of a data lake with the performance of a data warehouse. It automatically processes various data formats, creates appropriate schemas, identifies dimensions and facts using intelligent algorithms, and provides visualization-ready star schema views.

## Key Features

### 1. **Multi-Format Data Ingestion**
- **Supports:** CSV, JSON, XLSX, XLS, SQL dumps, .dump files, and more
- **Automatic format detection** based on file extension and content
- **Intelligent column extraction** from different data sources

### 2. **Automatic Schema Creation**
- **Dynamic table generation** based on detected column structures
- **Type inference** (INTEGER, NUMERIC, TEXT, DATE) from sample data
- **Automatic column sanitization** for valid PostgreSQL identifiers
- **Adaptive schema evolution** as new data structures are encountered

### 3. **Intelligent Data Analysis** (Simulated LLM)
- **Built-in pattern recognition algorithm** (no external LLM API required)
- **Column classification:** Automatically identifies dimensions vs facts based on:
  - Column name patterns ("name", "category", "date" → dimensions)
  - Data value patterns (quantitative vs categorical)
  - Semantic analysis of content and context
- **Relationship detection:** Identifies potential foreign key relationships
- **Star schema optimization:** Creates efficient analytical views

### 4. **Web-Based File Upload Interface**
- **Drag-and-drop file uploads** for easy data ingestion
- **Real-time progress tracking** during upload
- **Configuration management** for MinIO connection settings
- **Multi-file support** with batch processing capabilities

### 5. **ETL Pipeline**
- **Extract:** Supports multiple data sources (files, SQL queries)
- **Transform:** Standardizes, normalizes, and cleanses data
- **Load:** Automatically creates tables and loads processed data to PostgreSQL
- **Raw data storage:** Preserves original files in MinIO for auditability

### 6. **Visualization Integration**
- **Power BI/Tableau ready:** Direct PostgreSQL connection support
- **Pre-built star schema views** optimized for BI tools
- **Read-only user setup** for secure client access
- **Data mart views** for specific business domains

### 7. **Email Automation System**
- **Personalized email templates** with Hi <name> functionality
- **Configurable email lists** via YAML configuration
- **Scheduled reporting** with cron-based automation
- **Attachment support** for generated reports

### 8. **Time Simulation for Testing**
- **Time travel commands** for testing scheduled processes
- **Advance/go-back functionality** for email timing tests
- **Reset capabilities** for testing scenarios

## How It Works

### Data Flow Process
1. **File Ingestion:** User uploads files via web interface or system processes them via scheduled jobs
2. **Format Detection:** System analyzes file extension and content to determine processing method
3. **Schema Analysis:** Pattern recognition algorithm classifies columns as dimensions or facts
4. **Table Creation:** Automatic PostgreSQL table generation based on detected schema
5. **Data Loading:** Clean data loaded to PostgreSQL, raw data stored in MinIO
6. **Star Schema Generation:** Intelligent view creation for analytical queries
7. **Visualization Ready:** Ready for Power BI/Tableau connection

### Intelligent Analysis Algorithm
The system uses a **built-in algorithm** (not external LLM API) that:
- Analyzes column names for semantic patterns
- Examines data values to determine types and characteristics
- Applies business logic rules for categorization
- Creates optimized star schema relationships based on the analysis

## Installation Instructions

### Prerequisites

#### For Linux (Gentoo - Primary Platform)
- Gentoo GNU/Linux with updated kernel
- Kernel features: Namespaces, Cgroups v1/v2, OverlayFS, container support
- Docker or containerd
- k3s or Kubernetes cluster
- Git for version control
- At least 8GB RAM and 20GB free disk space (after Docker cleanup)

#### For macOS
- macOS 10.15 or later
- Docker Desktop for Mac
- Homebrew for package management
- At least 8GB RAM and 20GB free disk space
- Command Line Tools: `xcode-select --install`

#### For Windows
- Windows 10 version 2004 or Windows 11
- Docker Desktop for Windows (with WSL 2 backend recommended)
- At least 8GB RAM and 20GB free disk space
- Git for Windows

### Installation Steps

#### Linux (Gentoo)
```bash
# 1. Install k3s
curl -sfL https://get.k3s.io | sh -

# 2. Clone the repository
git clone <repository-url>
cd <repository-name>

# 3. Deploy the Data Lakehouse
./deploy_simple.sh

# 4. Start the web interface
python3 start_web_interface.py
```

#### macOS
```bash
# 1. Install dependencies
brew install docker kubernetes-cli
# Install Docker Desktop from https://docker.com

# 2. Start Docker Desktop
# Enable Kubernetes in Docker Desktop settings

# 3. Clone and deploy
git clone <repository-url>
cd <repository-name>

# For Kubernetes deployment (requires more resources):
kubectl apply -f k8s-manifests/namespaces.yaml
# ... apply other manifests

# 4. Start the web interface
python3 start_web_interface.py
```

#### Windows
```powershell
# 1. Install Docker Desktop for Windows
# Download from https://docker.com and install

# 2. Enable Kubernetes in Docker Desktop settings

# 3. Clone the repository
git clone <repository-url>
cd <repository-name>

# 4. Start web interface (for file uploads)
python start_web_interface.py

# 5. For Kubernetes deployment:
kubectl apply -f k8s-manifests/namespaces.yaml
# ... deploy other components
```

### Quick Start for Testing
```bash
# For immediate testing without Kubernetes (using Docker Compose):
docker-compose up -d

# Access the web interface:
# Open http://localhost:8000/minimal_upload.html in your browser

# Connect visualization tools to:
# Server: localhost:5432
# Database: datalakehouse
# Username: client_reader
# Password: reader_pass123
```

## Usage Instructions

### 1. File Upload Process
1. Access the web interface at `http://localhost:8000/minimal_upload.html`
2. Configure MinIO settings (endpoint, credentials, bucket name)
3. Drag and drop your data files or click "Browse Files"
4. Click "Upload Files to Data Lakehouse"
5. Monitor progress in the interface

### 2. Data Processing
- Files are automatically processed through the ETL pipeline
- Schema detection happens automatically
- Tables are created in PostgreSQL
- Star schema views are generated

### 3. Visualization Connection
- **Power BI:** Get Data → Database → PostgreSQL database
- **Tableau:** Connect to PostgreSQL
- Server: Your cluster IP:30432 (or localhost:5432 for local)
- Database: datalakehouse
- Username: client_reader (read-only access)
- Password: reader_pass123

### 4. Email Configuration
- Edit `email-config.yaml` to configure recipients and SMTP settings
- Set up scheduled reports in the configuration

## Architecture Components

### 1. **MinIO (Data Lake)**
- Object storage for raw files
- Buckets: `raw/` for original files
- API and console access

### 2. **PostgreSQL (Data Warehouse)**
- Star schema implementation
- Automatically created tables based on data
- Pre-built analytical views
- Read-only user for visualization tools

### 3. **ETL Service (Go)**
- Multi-format file processing
- Automatic schema detection
- Data transformation and cleansing
- Automatic table creation

### 4. **Web Interface**
- Drag-and-drop file uploads
- Configuration management
- Real-time status updates

### 5. **R Reporting Service**
- Automated report generation
- PDF creation with visualizations
- Email delivery system
- Scheduled execution

## System Requirements

### Minimum Requirements
- **CPU:** 4 cores
- **RAM:** 8GB
- **Storage:** 20GB free space
- **Network:** Stable internet connection (for initial setup)

### Recommended Requirements
- **CPU:** 8 cores
- **RAM:** 16GB
- **Storage:** 50GB free space (for large datasets)
- **Network:** High-speed connection

## Supported Operating Systems

| OS | Support Level | Notes |
|----|---------------|-------|
| Gentoo Linux | Full | Primary development platform |
| Other Linux | Full | Requires Docker/kubernetes setup |
| macOS | Full | Docker Desktop required |
| Windows | Full | Docker Desktop with WSL2 recommended |

## Troubleshooting

### Common Issues
1. **Disk Pressure:** Run `docker system prune -a` to free space
2. **Pod Scheduling:** Check node status with `kubectl describe nodes`
3. **Web Interface Not Starting:** Ensure Python 3 is installed
4. **Connection Issues:** Verify port availability and firewall settings

### Performance Tips
- Regular Docker cleanup for optimal performance
- Monitor disk space in production environments
- Adjust resource limits based on data volume

## Security Considerations

- Default passwords should be changed in production
- Use TLS/SSL for production deployments
- Restrict direct database access
- Implement proper backup strategies

## Next Steps

1. Upload your business data files via the web interface
2. Connect Power BI or Tableau to the PostgreSQL database
3. Configure email reports for automated delivery
4. Monitor and optimize based on usage patterns

---

**Note:** This system is designed to work with any business data structure and scales automatically based on your data volume. The intelligent pattern recognition algorithm adapts to different data formats without requiring predefined schemas.

