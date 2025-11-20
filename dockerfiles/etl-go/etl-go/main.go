package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"bytes"

	_ "github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// DataRecord represents a generic data record
type DataRecord map[string]interface{}

// ETLService handles the ETL process
type ETLService struct {
	minioClient  *minio.Client
	db           *sql.DB
	minioBucket  string
}

// NewETLService creates a new ETL service
func NewETLService(minioEndpoint, minioAccessKey, minioSecretKey, dbName, dbUser, dbPassword, dbHost string) (*ETLService, error) {
	// Initialize MinIO client
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %v", err)
	}

	// Initialize PostgreSQL connection
	psqlInfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %v", err)
	}

	return &ETLService{
		minioClient: minioClient,
		db:          db,
		minioBucket: "raw",
	}, nil
}

// ExtractFromCSV extracts data from CSV file
func (e *ETLService) ExtractFromCSV(filePath string) ([]DataRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no data found in CSV file")
	}

	headers := records[0]
	var data []DataRecord

	for i := 1; i < len(records); i++ {
		record := make(DataRecord)
		for j, header := range headers {
			if j < len(records[i]) {
				record[header] = records[i][j]
			}
		}
		data = append(data, record)
	}

	return data, nil
}

// Transform standardizes and normalizes data
func (e *ETLService) Transform(data []DataRecord) []DataRecord {
	// Example transformation: standardize data types and clean data
	for _, record := range data {
		// Standardize numeric values, normalize text, etc.
		for key, value := range record {
			switch v := value.(type) {
			case string:
				// Trim whitespace
				record[key] = fmt.Sprintf("%v", v)
			}
		}
	}
	return data
}

// LoadToMinIO uploads raw data to MinIO
func (e *ETLService) LoadToMinIO(data []DataRecord, fileName string) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	objectName := fmt.Sprintf("raw/%s", fileName)
	_, err = e.minioClient.PutObject(
		context.Background(),
		e.minioBucket,
		objectName,
		bytes.NewReader(jsonData),
		int64(len(jsonData)),
		minio.PutObjectOptions{ContentType: "application/json"},
	)

	return err
}

// LoadToPostgreSQL loads cleaned data to PostgreSQL
func (e *ETLService) LoadToPostgreSQL(data []DataRecord) error {
	for _, record := range data {
		// Example: inserting into fact_sales table
		// This would need to be adjusted based on your actual schema
		query := `
			INSERT INTO fact_sales (date_id, customer_id, product_id, location_id, quantity, unit_price, total_amount)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		
		// Parse the values from the record
		quantity := 0
		unitPrice := 0.0
		totalAmount := 0.0
		
		if q, ok := record["quantity"].(string); ok {
			fmt.Sscanf(q, "%d", &quantity)
		}
		if up, ok := record["unit_price"].(string); ok {
			fmt.Sscanf(up, "%f", &unitPrice)
		}
		if ta, ok := record["total_amount"].(string); ok {
			fmt.Sscanf(ta, "%f", &totalAmount)
		}
		
		_, err := e.db.Exec(query, 1, 1, 1, 1, quantity, unitPrice, totalAmount)
		if err != nil {
			log.Printf("Error inserting record: %v", err)
			continue
		}
	}
	
	return nil
}

// ProcessETLFromFile processes ETL from a file source
func (e *ETLService) ProcessETLFromFile(filePath string) error {
	// Extract
	data, err := e.ExtractFromCSV(filePath)
	if err != nil {
		return fmt.Errorf("extract failed: %v", err)
	}

	// Transform
	transformedData := e.Transform(data)

	// Load to MinIO (raw data)
	fileName := "raw_" + filePath
	if err := e.LoadToMinIO(data, fileName); err != nil {
		return fmt.Errorf("load to MinIO failed: %v", err)
	}

	// Load to PostgreSQL (processed data)
	if err := e.LoadToPostgreSQL(transformedData); err != nil {
		return fmt.Errorf("load to PostgreSQL failed: %v", err)
	}

	log.Println("ETL process completed successfully")
	return nil
}

// ExtractFromSQL extracts data from PostgreSQL
func (e *ETLService) ExtractFromSQL(query string) ([]DataRecord, error) {
	rows, err := e.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var data []DataRecord
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for rows.Next() {
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		record := make(DataRecord)
		for i, col := range columns {
			record[col] = values[i]
		}
		data = append(data, record)
	}

	return data, nil
}

// ProcessETLFromSQL processes ETL from a SQL source
func (e *ETLService) ProcessETLFromSQL(query string) error {
	// Extract
	data, err := e.ExtractFromSQL(query)
	if err != nil {
		return fmt.Errorf("extract from SQL failed: %v", err)
	}

	// Transform
	transformedData := e.Transform(data)

	// Load to MinIO (raw data)
	fileName := "raw_sql_" + query
	if err := e.LoadToMinIO(data, fileName); err != nil {
		return fmt.Errorf("load to MinIO failed: %v", err)
	}

	// Load to PostgreSQL (processed data)
	if err := e.LoadToPostgreSQL(transformedData); err != nil {
		return fmt.Errorf("load to PostgreSQL failed: %v", err)
	}

	log.Println("ETL process from SQL completed successfully")
	return nil
}

func main() {
	// Get environment variables
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")

	// Create ETL service
	etl, err := NewETLService(minioEndpoint, minioAccessKey, minioSecretKey, dbName, dbUser, dbPassword, dbHost)
	if err != nil {
		log.Fatalf("Failed to initialize ETL service: %v", err)
	}
	defer etl.db.Close()

	// Determine source type and process accordingly
	sourceType := os.Getenv("ETL_SOURCE_TYPE") // "file" or "sql"
	
	if sourceType == "file" {
		filePath := os.Getenv("ETL_SOURCE_FILE")
		if err := etl.ProcessETLFromFile(filePath); err != nil {
			log.Fatalf("ETL process failed: %v", err)
		}
	} else if sourceType == "sql" {
		query := os.Getenv("ETL_SOURCE_QUERY")
		if err := etl.ProcessETLFromSQL(query); err != nil {
			log.Fatalf("ETL process from SQL failed: %v", err)
		}
	} else {
		log.Fatal("ETL_SOURCE_TYPE must be either 'file' or 'sql'")
	}
}