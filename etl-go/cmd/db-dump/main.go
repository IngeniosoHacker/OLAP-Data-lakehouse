package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

// DBDumpService handles database dump operations
type DBDumpService struct {
	db *sql.DB
}

// NewDBDumpService creates a new database dump service
func NewDBDumpService(dbHost, dbName, dbUser, dbPassword string) (*DBDumpService, error) {
	psqlInfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %v", err)
	}

	return &DBDumpService{
		db: db,
	}, nil
}

// Close closes the database connection
func (dbs *DBDumpService) Close() {
	if dbs.db != nil {
		dbs.db.Close()
	}
}

// GetTableNames retrieves all table names from the database
func (dbs *DBDumpService) GetTableNames() ([]string, error) {
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
		ORDER BY table_name;
	`
	
	rows, err := dbs.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get table names: %v", err)
	}
	defer rows.Close()

	var tableNames []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tableNames = append(tableNames, tableName)
	}

	return tableNames, nil
}

// GetTableSchema retrieves the schema for a specific table
func (dbs *DBDumpService) GetTableSchema(tableName string) (string, error) {
	// Get columns info
	columnsQuery := `
		SELECT 
			column_name, 
			data_type,
			is_nullable,
			column_default
		FROM information_schema.columns
		WHERE table_name = $1
		ORDER BY ordinal_position;
	`
	
	rows, err := dbs.db.Query(columnsQuery, tableName)
	if err != nil {
		return "", fmt.Errorf("failed to get columns for table %s: %v", tableName, err)
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var colName, dataType, isNullable string
		var defaultValue *string // nullable string
		
		if err := rows.Scan(&colName, &dataType, &isNullable, &defaultValue); err != nil {
			return "", err
		}
		
		colDef := fmt.Sprintf("    \"%s\" %s", colName, strings.ToUpper(dataType))
		if isNullable == "NO" {
			colDef += " NOT NULL"
		}
		if defaultValue != nil {
			colDef += fmt.Sprintf(" DEFAULT %s", *defaultValue)
		}
		
		columns = append(columns, colDef)
	}

	// Get primary key info
	pkQuery := `
		SELECT ku.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage ku
			ON tc.constraint_name = ku.constraint_name
			AND tc.table_name = ku.table_name
		WHERE tc.constraint_type = 'PRIMARY KEY'
			AND tc.table_name = $1;
	`
	
	pkRows, err := dbs.db.Query(pkQuery, tableName)
	if err != nil {
		return "", fmt.Errorf("failed to get primary key for table %s: %v", tableName, err)
	}
	defer pkRows.Close()

	var pkColumns []string
	for pkRows.Next() {
		var colName string
		if err := pkRows.Scan(&colName); err != nil {
			return "", err
		}
		pkColumns = append(pkColumns, fmt.Sprintf("\"%s\"", colName))
	}

	schema := fmt.Sprintf("CREATE TABLE IF NOT EXISTS \"%s\" (\n", tableName)
	schema += strings.Join(columns, ",\n")
	
	if len(pkColumns) > 0 {
		schema += fmt.Sprintf(",\n    PRIMARY KEY (%s)", strings.Join(pkColumns, ", "))
	}
	
	schema += "\n);\n\n"
	
	return schema, nil
}

// GetTableData retrieves all data from a specific table
func (dbs *DBDumpService) GetTableData(tableName string) (string, error) {
	query := fmt.Sprintf("SELECT * FROM \"%s\"", tableName)
	
	rows, err := dbs.db.Query(query)
	if err != nil {
		return "", fmt.Errorf("failed to query table %s: %v", tableName, err)
	}
	defer rows.Close()

	// Get column information
	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("failed to get column info for table %s: %v", tableName, err)
	}

	// Build INSERT statements
	var inserts []string
	
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		
		if err := rows.Scan(valuePtrs...); err != nil {
			return "", fmt.Errorf("failed to scan row in table %s: %v", tableName, err)
		}
		
		// Build VALUES part of INSERT
		var valuesStr []string
		for _, val := range values {
			strVal := formatValue(val)
			valuesStr = append(valuesStr, strVal)
		}
		
		insert := fmt.Sprintf("INSERT INTO \"%s\" (%s) VALUES (%s);", 
			tableName, 
			"\"" + strings.Join(columns, "\", \"") + "\"",
			strings.Join(valuesStr, ", "))
		
		inserts = append(inserts, insert)
	}
	
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating rows in table %s: %v", tableName, err)
	}
	
	return strings.Join(inserts, "\n") + "\n\n", nil
}

// formatValue formats a value for SQL insertion
func formatValue(val interface{}) string {
	if val == nil {
		return "NULL"
	}
	
	switch v := val.(type) {
	case string:
		// Escape single quotes in string values
		escaped := strings.ReplaceAll(v, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%g", v)
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	default:
		// For other types, convert to string and escape
		strVal := fmt.Sprintf("%v", v)
		escaped := strings.ReplaceAll(strVal, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	}
}

// DumpDatabase creates a complete SQL dump of the database
func (dbs *DBDumpService) DumpDatabase(outputFile string) error {
	tableNames, err := dbs.GetTableNames()
	if err != nil {
		return fmt.Errorf("failed to get table names: %v", err)
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	// Write header
	header := `-- Database Dump Generated by Data Lakehouse ETL System
-- Exported on: ` + fmt.Sprintf("%s\n\n", "2025-11-19 10:00:00") + `
-- This dump contains schema and data for sharing with visualization tools
-- To import: psql -d database_name -f dump_file.sql

-- Disable triggers and constraints during import
SET session_replication_role = replica;

`
	
	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}

	// Dump each table's schema and data
	for _, tableName := range tableNames {
		// Write table schema
		schema, err := dbs.GetTableSchema(tableName)
		if err != nil {
			log.Printf("Warning: failed to get schema for table %s: %v", tableName, err)
			continue
		}
		
		if _, err := file.WriteString("-- Schema for table: " + tableName + "\n"); err != nil {
			return fmt.Errorf("failed to write schema comment: %v", err)
		}
		
		if _, err := file.WriteString(schema); err != nil {
			return fmt.Errorf("failed to write schema: %v", err)
		}

		// Write table data
		data, err := dbs.GetTableData(tableName)
		if err != nil {
			log.Printf("Warning: failed to get data for table %s: %v", tableName, err)
			continue
		}
		
		if _, err := file.WriteString("-- Data for table: " + tableName + "\n"); err != nil {
			return fmt.Errorf("failed to write data comment: %v", err)
		}
		
		if _, err := file.WriteString(data); err != nil {
			return fmt.Errorf("failed to write data: %v", err)
		}
	}

	// Write footer
	footer := `
-- Re-enable triggers and constraints
SET session_replication_role = DEFAULT;

-- End of dump
`
	
	if _, err := file.WriteString(footer); err != nil {
		return fmt.Errorf("failed to write footer: %v", err)
	}

	log.Printf("Database dump completed successfully: %s", outputFile)
	return nil
}

func main() {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Fatal("DB_NAME environment variable must be set")
	}
	
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		log.Fatal("DB_USER environment variable must be set")
	}
	
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		log.Fatal("DB_PASSWORD environment variable must be set")
	}
	
	outputFile := os.Getenv("OUTPUT_FILE")
	if outputFile == "" {
		outputFile = "database_dump.sql"
	}

	dumpService, err := NewDBDumpService(dbHost, dbName, dbUser, dbPassword)
	if err != nil {
		log.Fatalf("Failed to create dump service: %v", err)
	}
	defer dumpService.Close()

	if err := dumpService.DumpDatabase(outputFile); err != nil {
		log.Fatalf("Failed to dump database: %v", err)
	}
}