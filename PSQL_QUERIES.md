# PostgreSQL Queries for Data Lakehouse Exploration

Use these psql commands to explore your dimension and fact tables in the PostgreSQL database.

## Connection Details
```bash
PGPASSWORD=password123 psql -h localhost -p 5433 -U admin -d datalakehouse
```

## Quick Exploration Commands

### 1. List All Tables
```sql
\dt
```

### 2. List All Tables with Additional Information
```sql
SELECT table_name, column_name, data_type 
FROM information_schema.columns 
WHERE table_schema = 'public' 
ORDER BY table_name, ordinal_position;
```

### 3. View Star Schema Views
```sql
SELECT table_name 
FROM information_schema.views 
WHERE table_schema = 'public' 
AND table_name LIKE '%_star_view';
```

### 4. Check a Specific Table Structure
```sql
-- Replace 'your_table_name' with actual table name
\d your_table_name
```

### 5. Count Records in Each Table
```sql
SELECT 
    table_name, 
    (SELECT COUNT(*) FROM your_table_name) as record_count
FROM information_schema.tables 
WHERE table_schema = 'public';
```
*Note: You'll need to manually run count queries for each table as the above is illustrative*

## Detailed Exploration Commands

### 6. List All Star Schema Views
```sql
SELECT table_name 
FROM information_schema.tables 
WHERE table_type = 'VIEW' 
AND table_name LIKE '%star%'
ORDER BY table_name;
```

### 7. See All Columns in a Specific Table
```sql
-- Replace 'your_table_name' with the actual table name
SELECT column_name, data_type, is_nullable 
FROM information_schema.columns 
WHERE table_name = 'your_table_name'
ORDER BY ordinal_position;
```

### 8. Check Sample Data from Tables
```sql
-- Replace 'your_table_name' with actual table name
SELECT * FROM your_table_name LIMIT 10;
```

### 9. Find Potential Dimension and Fact Tables
```sql
-- Look for tables that might be dimensions (often containing name, category, date fields)
SELECT table_name, column_name, data_type 
FROM information_schema.columns 
WHERE table_schema = 'public'
AND (
    column_name LIKE '%name%' 
    OR column_name LIKE '%category%' 
    OR column_name LIKE '%type%' 
    OR column_name LIKE '%date%'
)
ORDER BY table_name, column_name;
```

### 10. Find Potential Fact Tables (often containing numeric measures)
```sql
SELECT table_name, column_name, data_type 
FROM information_schema.columns 
WHERE table_schema = 'public'
AND (
    column_name LIKE '%amount%' 
    OR column_name LIKE '%price%' 
    OR column_name LIKE '%quantity%' 
    OR column_name LIKE '%count%' 
    OR column_name LIKE '%total%'
)
ORDER BY table_name, column_name;
```

### 11. Check All Views (likely star schema views)
```sql
SELECT table_name, view_definition 
FROM information_schema.views 
WHERE table_schema = 'public';
```

### 12. Get Information About Star Schema Views Specifically
```sql
SELECT table_name 
FROM information_schema.views 
WHERE table_name LIKE '%_star_view%';
```

### 13. Show All Data Types Used in Your Tables
```sql
SELECT DISTINCT data_type 
FROM information_schema.columns 
WHERE table_schema = 'public'
ORDER BY data_type;
```

### 14. Find Tables with Specific Column Patterns (e.g., likely dimensions)
```sql
SELECT DISTINCT t.table_name
FROM information_schema.tables t
JOIN information_schema.columns c ON t.table_name = c.table_name
WHERE t.table_schema = 'public'
AND (
    c.column_name ILIKE '%name%'
    OR c.column_name ILIKE '%desc%'
    OR c.column_name ILIKE '%category%'
    OR c.column_name ILIKE '%type%'
    OR c.column_name ILIKE '%date%'
);
```

## Example Session Sequence

To start exploring:
1. Connect: `PGPASSWORD=password123 psql -h localhost -p 5433 -U admin -d datalakehouse`
2. List tables: `\dt`
3. Look for your auto-generated table: `SELECT * FROM your_auto_table_name LIMIT 5;`
4. Check for star views: `\dv`
5. Look at star view definitions: `SELECT table_name FROM information_schema.views;`
6. Exit: `\q`

## Common Queries for Analysis

### For Dimension Tables:
```sql
-- Get distinct values for categorical data
SELECT DISTINCT column_name FROM table_name ORDER BY column_name LIMIT 20;

-- Get count of distinct values
SELECT COUNT(DISTINCT column_name) FROM table_name;
```

### For Fact Tables:
```sql
-- Get statistical summaries
SELECT 
    MIN(column_name), 
    MAX(column_name), 
    AVG(column_name), 
    SUM(column_name), 
    COUNT(*) 
FROM table_name;
```

## Sample Database Schema

Based on your current database, here's what was automatically created by the ETL process:

### Fact Table: fact_sales
- Contains sales transactions with foreign keys linking to dimension tables
- Fields: sale_id, date_id, customer_id, product_id, location_id, quantity, unit_price, total_amount, discount_amount, tax_amount, created_timestamp

### Dimension Tables:
- **dim_customer**: customer_id, customer_name, customer_email, customer_phone, customer_address, customer_city, customer_state, customer_country
- **dim_date**: date_id, date_value, day_of_week, day_of_month, month, quarter, year
- **dim_location**: location_id, location_name, location_address, location_city, location_state, location_country
- **dim_product**: product_id, product_name, product_category, product_description, product_price, product_cost, product_manufacturer

### Star Schema View: sales_summary
- A pre-built view joining fact and dimension tables for easier analysis
- Contains denormalized data for common analytical queries

## Example Queries

### Check data in fact table:
```sql
SELECT * FROM fact_sales LIMIT 5;
```

### Check data in dimension tables:
```sql
SELECT * FROM dim_customer LIMIT 5;
SELECT * FROM dim_product LIMIT 5;
SELECT * FROM dim_location LIMIT 5;
SELECT * FROM dim_date LIMIT 5;
```

### Check the star schema view:
```sql
SELECT * FROM sales_summary LIMIT 5;
```

### Analytical Queries:
```sql
-- Total sales by customer
SELECT customer_name, SUM(total_amount) as total_sales
FROM sales_summary
GROUP BY customer_name
ORDER BY total_sales DESC;

-- Sales by product category
SELECT product_category, COUNT(*) as number_of_sales, SUM(total_amount) as total_revenue
FROM sales_summary
GROUP BY product_category;

-- Sales over time
SELECT date_value, SUM(total_amount) as daily_sales
FROM sales_summary
GROUP BY date_value
ORDER BY date_value;
```

These commands will help you explore your data warehouse and understand the dimension and fact tables created by the ETL process.