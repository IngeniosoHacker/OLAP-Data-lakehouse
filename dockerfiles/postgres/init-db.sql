-- Database initialization script for Data Lakehouse
-- Creates schema and loads sample data

-- Create dimension tables
CREATE TABLE IF NOT EXISTS dim_customer (
    customer_id SERIAL PRIMARY KEY,
    customer_name TEXT NOT NULL,
    customer_email TEXT,
    customer_phone TEXT,
    customer_address TEXT,
    customer_city TEXT,
    customer_state TEXT,
    customer_country TEXT,
    created_date DATE
);

CREATE TABLE IF NOT EXISTS dim_product (
    product_id SERIAL PRIMARY KEY,
    product_name TEXT NOT NULL,
    product_category TEXT,
    product_description TEXT,
    product_price NUMERIC(10,2),
    product_cost NUMERIC(10,2),
    product_manufacturer TEXT,
    created_date DATE
);

CREATE TABLE IF NOT EXISTS dim_location (
    location_id SERIAL PRIMARY KEY,
    location_name TEXT NOT NULL,
    location_address TEXT,
    location_city TEXT,
    location_state TEXT,
    location_country TEXT,
    location_zipcode TEXT,
    created_date DATE
);

CREATE TABLE IF NOT EXISTS dim_date (
    date_id SERIAL PRIMARY KEY,
    date_value DATE NOT NULL,
    day_of_week INTEGER,
    day_name TEXT,
    month_number INTEGER,
    month_name TEXT,
    year_number INTEGER,
    quarter_number INTEGER,
    is_weekend BOOLEAN,
    is_holiday BOOLEAN DEFAULT FALSE
);

-- Create fact table
CREATE TABLE IF NOT EXISTS fact_sales (
    sale_id SERIAL PRIMARY KEY,
    date_id INTEGER,
    customer_id INTEGER,
    product_id INTEGER,
    location_id INTEGER,
    quantity INTEGER,
    unit_price NUMERIC(10,2),
    total_amount NUMERIC(12,2),
    discount_amount NUMERIC(10,2) DEFAULT 0,
    tax_amount NUMERIC(10,2) DEFAULT 0,
    created_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create additional table for testing the automatic table creation
CREATE TABLE IF NOT EXISTS auto_table_test (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER,
    customer_name TEXT,
    product_id INTEGER,
    product_name TEXT,
    quantity INTEGER,
    unit_price NUMERIC(10,2),
    total_amount NUMERIC(12,2),
    created_date DATE
);

-- Create star schema view
CREATE OR REPLACE VIEW sales_summary AS
SELECT 
    s.sale_id,
    d.date_value,
    c.customer_name,
    p.product_name,
    l.location_name,
    s.quantity,
    s.unit_price,
    s.total_amount,
    (s.total_amount / s.quantity) AS avg_price_per_unit
FROM fact_sales s
JOIN dim_date d ON s.date_id = d.date_id
JOIN dim_customer c ON s.customer_id = c.customer_id
JOIN dim_product p ON s.product_id = p.product_id
JOIN dim_location l ON s.location_id = l.location_id;

-- Create read-only user for visualization tools
DO
$do$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_catalog.pg_user
      WHERE usename = 'client_reader') THEN

      CREATE USER client_reader WITH PASSWORD 'reader_pass123';
      GRANT CONNECT ON DATABASE datalakehouse TO client_reader;
      GRANT USAGE ON SCHEMA public TO client_reader;
      
      -- Grant SELECT permissions on all existing tables and views
      GRANT SELECT ON ALL TABLES IN SCHEMA public TO client_reader;
      GRANT SELECT ON ALL SEQUENCES IN SCHEMA public TO client_reader;
      
      -- Grant SELECT permissions on future tables and views
      ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO client_reader;
      ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON SEQUENCES TO client_reader;
   END IF;
END
$do$;

-- Insert sample data
INSERT INTO dim_customer (customer_name, customer_email, customer_phone, customer_address, customer_city, customer_state, customer_country, created_date) VALUES
('John Doe', 'john.doe@example.com', '555-0101', '123 Main St', 'New York', 'NY', 'USA', '2023-01-15'),
('Jane Smith', 'jane.smith@example.com', '555-0102', '456 Oak Ave', 'Los Angeles', 'CA', 'USA', '2023-02-20'),
('Bob Johnson', 'bob.johnson@example.com', '555-0103', '789 Pine Rd', 'Chicago', 'IL', 'USA', '2023-03-10'),
('Alice Williams', 'alice.williams@example.com', '555-0104', '321 Elm St', 'Houston', 'TX', 'USA', '2023-04-05');

INSERT INTO dim_product (product_name, product_category, product_description, product_price, product_cost, product_manufacturer, created_date) VALUES
('Laptop', 'Electronics', 'High-performance laptop', 1200.00, 900.00, 'TechCorp', '2023-01-10'),
('Mouse', 'Electronics', 'Wireless optical mouse', 25.00, 10.00, 'TechCorp', '2023-01-10'),
('Keyboard', 'Electronics', 'Mechanical keyboard', 80.00, 40.00, 'TechCorp', '2023-01-10'),
('Monitor', 'Electronics', '24-inch display', 200.00, 120.00, 'DisplayCorp', '2023-02-01'),
('Headphones', 'Electronics', 'Noise-cancelling headphones', 150.00, 70.00, 'AudioCorp', '2023-02-15');

INSERT INTO dim_location (location_name, location_address, location_city, location_state, location_country, location_zipcode, created_date) VALUES
('NY Flagship Store', '1001 Broadway', 'New York', 'NY', 'USA', '10001', '2023-01-01'),
('LA Outlet', '2002 Hollywood Blvd', 'Los Angeles', 'CA', 'USA', '90028', '2023-01-01'),
('Chicago Branch', '3003 Michigan Ave', 'Chicago', 'IL', 'USA', '60616', '2023-01-01'),
('Houston Location', '4004 Dallas St', 'Houston', 'TX', 'USA', '77002', '2023-01-01');

INSERT INTO dim_date (date_value, day_of_week, day_name, month_number, month_name, year_number, quarter_number, is_weekend) VALUES
('2023-01-15', 7, 'Sunday', 1, 'January', 2023, 1, TRUE),
('2023-02-20', 1, 'Monday', 2, 'February', 2023, 1, FALSE),
('2023-03-10', 5, 'Friday', 3, 'March', 2023, 1, FALSE),
('2023-04-05', 3, 'Wednesday', 4, 'April', 2023, 2, FALSE),
('2023-05-15', 1, 'Monday', 5, 'May', 2023, 2, FALSE),
('2023-06-20', 2, 'Tuesday', 6, 'June', 2023, 2, FALSE),
('2023-07-10', 1, 'Monday', 7, 'July', 2023, 3, FALSE),
('2023-08-05', 6, 'Saturday', 8, 'August', 2023, 3, TRUE),
('2023-09-15', 5, 'Friday', 9, 'September', 2023, 3, FALSE),
('2023-10-20', 5, 'Friday', 10, 'October', 2023, 4, FALSE),
('2023-11-10', 5, 'Friday', 11, 'November', 2023, 4, FALSE),
('2023-12-25', 1, 'Monday', 12, 'December', 2023, 4, FALSE);

INSERT INTO fact_sales (date_id, customer_id, product_id, location_id, quantity, unit_price, total_amount) VALUES
(1, 1, 1, 1, 2, 1200.00, 2400.00),
(2, 2, 2, 2, 5, 25.00, 125.00),
(3, 3, 3, 3, 3, 80.00, 240.00),
(4, 4, 4, 4, 1, 200.00, 200.00),
(5, 1, 5, 1, 2, 150.00, 300.00),
(6, 2, 1, 2, 1, 1200.00, 1200.00),
(7, 3, 2, 3, 10, 25.00, 250.00),
(8, 4, 3, 4, 1, 80.00, 80.00),
(9, 1, 4, 1, 3, 200.00, 600.00),
(10, 2, 5, 2, 1, 150.00, 150.00);

INSERT INTO auto_table_test (customer_id, customer_name, product_id, product_name, quantity, unit_price, total_amount, created_date) VALUES
(1, 'John Doe', 101, 'Laptop', 2, 1000.00, 2000.00, '2023-06-15'),
(2, 'Jane Smith', 102, 'Mouse', 5, 25.00, 125.00, '2023-06-20'),
(3, 'Bob Johnson', 103, 'Keyboard', 3, 75.00, 225.00, '2023-06-25');