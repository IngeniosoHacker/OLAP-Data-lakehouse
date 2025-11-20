# Power BI/Tableau Connection Guide

## Connection Details

To connect Power BI or Tableau to the PostgreSQL data warehouse:

- **Server**: [Gentoo Node IP Address]:30432 (or whatever NodePort is configured)
- **Database**: datalakehouse
- **Username**: client_reader
- **Password**: reader_pass123

## Connection String Example

For direct PostgreSQL connection:
```
Host=[Gentoo_Node_IP];Port=30432;Database=datalakehouse;Username=client_reader;Password=reader_pass123;
```

## Available Views/Datamarts

The following views are available for reporting:

- `sales_summary` - Aggregated sales data with customer, product, and location information
- `dim_customer` - Customer dimension table
- `dim_product` - Product dimension table
- `dim_date` - Date dimension table
- `fact_sales` - Core fact table with sales transactions

## Connection Steps (Power BI)

1. Open Power BI Desktop
2. Click "Get Data" → "Database" → "PostgreSQL database"
3. Enter server IP and database name
4. Authenticate with the read-only credentials
5. Select the required views/tables

## Connection Steps (Tableau)

1. Open Tableau Desktop
2. Select "PostgreSQL" from the connections
3. Enter server IP and database name
4. Authenticate with the read-only credentials
5. Connect to the required views/tables