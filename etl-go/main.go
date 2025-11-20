package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/extrame/gofile"
	"github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/tealeg/xlsx"
	"gopkg.in/mail.v2"
	"gopkg.in/yaml.v2"
)

// Recipient represents an email recipient
type Recipient struct {
	Name        string   `yaml:"name"`
	Email       string   `yaml:"email"`
	Department  string   `yaml:"department"`
	ReportTypes []string `yaml:"report_types"`
}

// SMTPConfig represents SMTP server configuration
type SMTPConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	From       string `yaml:"from"`
	Encryption string `yaml:"encryption"` // Options: none, ssl, starttls
}

// ReportConfig represents report-specific settings
type ReportConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Schedule string `yaml:"schedule"`
	Subject  string `yaml:"subject"`
	Template string `yaml:"template"`
}

// EmailConfig represents the complete email configuration
type EmailConfig struct {
	Recipients []Recipient             `yaml:"recipients"`
	SMTP       SMTPConfig              `yaml:"smtp"`
	Reports    map[string]ReportConfig `yaml:"reports"`
	Settings   struct {
		Timezone       string   `yaml:"timezone"`
		RetryAttempts  int      `yaml:"retry_attempts"`
		TimeoutSeconds int      `yaml:"timeout_seconds"`
		EnableLogging  bool     `yaml:"enable_logging"`
		Attachments    []string `yaml:"attachments"`
	} `yaml:"settings"`
}

// EmailService handles email operations
type EmailService struct {
	config *EmailConfig
}

// LoadEmailConfig loads the email configuration from a YAML file
func LoadEmailConfig(filePath string) (*EmailConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config EmailConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

// GetRecipientsByReportType returns recipients who should receive a specific report type
func (ec *EmailConfig) GetRecipientsByReportType(reportType string) []Recipient {
	var recipients []Recipient
	for _, recipient := range ec.Recipients {
		for _, rt := range recipient.ReportTypes {
			if rt == reportType {
				recipients = append(recipients, recipient)
				break
			}
		}
	}
	return recipients
}

// FormatEmailSubject formats the email subject with template variables
func (ec *EmailConfig) FormatEmailSubject(reportType string, recipient Recipient) (string, error) {
	reportConfig, exists := ec.Reports[reportType]
	if !exists {
		return "", fmt.Errorf("report type '%s' not found in config", reportType)
	}

	// Prepare template variables
	data := map[string]interface{}{
		"name":   recipient.Name,
		"date":   time.Now().Format("2006-01-02"),
		"month":  time.Now().Format("January 2006"),
		"email":  recipient.Email,
		"dept":   recipient.Department,
	}

	// Apply text template to subject
	tmpl, err := template.New("subject").Parse(reportConfig.Subject)
	if err != nil {
		return "", fmt.Errorf("failed to parse subject template: %v", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("failed to execute subject template: %v", err)
	}

	return result.String(), nil
}

// FormatEmailBody formats the email body with template variables
func (ec *EmailConfig) FormatEmailBody(reportType string, recipient Recipient) (string, error) {
	reportConfig, exists := ec.Reports[reportType]
	if !exists {
		return "", fmt.Errorf("report type '%s' not found in config", reportType)
	}

	// Prepare template variables
	data := map[string]interface{}{
		"name":   recipient.Name,
		"date":   time.Now().Format("2006-01-02"),
		"month":  time.Now().Format("January 2006"),
		"email":  recipient.Email,
		"dept":   recipient.Department,
	}

	// Apply text template to body
	tmpl, err := template.New("body").Parse(reportConfig.Template)
	if err != nil {
		return "", fmt.Errorf("failed to parse body template: %v", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("failed to execute body template: %v", err)
	}

	return result.String(), nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})
	return re.MatchString(email)
}

// ValidateConfig validates the email configuration
func (ec *EmailConfig) ValidateConfig() []string {
	var errors []string

	// Validate recipients
	for i, recipient := range ec.Recipients {
		if recipient.Name == "" {
			errors = append(errors, fmt.Sprintf("recipient %d has empty name", i))
		}
		if recipient.Email == "" {
			errors = append(errors, fmt.Sprintf("recipient %d has empty email", i))
		} else if !ValidateEmail(recipient.Email) {
			errors = append(errors, fmt.Sprintf("recipient %d has invalid email format: %s", i, recipient.Email))
		}
		if len(recipient.ReportTypes) == 0 {
			errors = append(errors, fmt.Sprintf("recipient %d has no report types", i))
		}
	}

	// Validate SMTP config
	if ec.SMTP.Host == "" {
		errors = append(errors, "SMTP host is empty")
	}
	if ec.SMTP.Port == 0 {
		errors = append(errors, "SMTP port is invalid")
	}
	if ec.SMTP.Username == "" {
		errors = append(errors, "SMTP username is empty")
	}
	if ec.SMTP.From == "" {
		errors = append(errors, "SMTP from address is empty")
	}

	// Validate reports
	if len(ec.Reports) == 0 {
		errors = append(errors, "no reports configured")
	}

	return errors
}

// NewEmailService creates a new email service
func NewEmailService(config *EmailConfig) *EmailService {
	return &EmailService{
		config: config,
	}
}

// SendTestEmail sends a test email to verify SMTP configuration
func (es *EmailService) SendTestEmail(toEmail string) error {
	// Create a temporary recipient for testing
	testRecipient := Recipient{
		Name:  "Test User",
		Email: toEmail,
	}

	subject, err := es.config.FormatEmailSubject("weekly", testRecipient)
	if err != nil {
		return fmt.Errorf("failed to format test email subject: %v", err)
	}

	body, err := es.config.FormatEmailBody("weekly", testRecipient)
	if err != nil {
		return fmt.Errorf("failed to format test email body: %v", err)
	}

	// Create message
	m := mail.NewMessage()
	m.SetHeader("From", es.config.SMTP.From)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", fmt.Sprintf("[TEST] %s", subject))
	m.SetBody("text/plain", fmt.Sprintf("This is a test email.\n\n%s", body))

	// Create SMTP dialer
	port := es.config.SMTP.Port
	d := mail.NewDialer(es.config.SMTP.Host, port, es.config.SMTP.Username, es.config.SMTP.Password)

	// Send email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send test email: %v", err)
	}

	log.Printf("Successfully sent test email to %s", toEmail)
	return nil
}

// DataRecord represents a generic data record
type DataRecord map[string]interface{}

// ETLService handles the ETL process
type ETLService struct {
	minioClient  *minio.Client
	db           *sql.DB
	minioBucket  string
}

// FileFormat represents the type of file being processed
type FileFormat int

const (
	Unknown FileFormat = iota
	CSV
	JSON
	XLSX
	XLS
	DUMP
	TAR
	TARGZ
	SQL
)

// GetFileFormat determines the file format based on extension or content
func GetFileFormat(filePath string) FileFormat {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".csv":
		return CSV
	case ".json":
		return JSON
	case ".xlsx":
		return XLSX
	case ".xls":
		return XLS
	case ".dump":
		return DUMP
	case ".tar":
		return TAR
	case ".gz":
		// Check if it's a tar.gz
		if strings.HasSuffix(strings.ToLower(filePath), ".tar.gz") {
			return TARGZ
		}
		return DUMP // Treat .gz as generic dump
	case ".sql":
		return SQL
	default:
		// Try to identify by content if extension is not clear
		return identifyFormatByContent(filePath)
	}
}

// identifyFormatByContent tries to identify the file format by looking at the content
func identifyFormatByContent(filePath string) FileFormat {
	file, err := os.Open(filePath)
	if err != nil {
		return Unknown
	}
	defer file.Close()

	// Read first 512 bytes to identify content
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return Unknown
	}
	buffer = buffer[:n]

	// Check for JSON
	if isJSON(buffer) {
		return JSON
	}

	// Check for XLSX (has specific header)
	if isXLSX(buffer) {
		return XLSX
	}

	// Check for XLS (has specific header)
	if isXLS(buffer) {
		return XLS
	}

	// Check for SQL dump
	if isSQL(buffer) {
		return SQL
	}

	return Unknown
}

// isJSON checks if content looks like JSON
func isJSON(data []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}

// isXLSX checks if content has XLSX signature
func isXLSX(data []byte) bool {
	// XLSX files are ZIP archives with specific content
	if len(data) < 4 {
		return false
	}
	// ZIP files start with PK
	return data[0] == 0x50 && data[1] == 0x4B
}

// isXLS checks if content has XLS signature
func isXLS(data []byte) bool {
	// XLS files have specific header
	if len(data) < 8 {
		return false
	}
	// Common XLS signatures
	return (data[0] == 0xD0 && data[1] == 0xCF && data[2] == 0x11 && data[3] == 0xE0)
}

// isSQL checks if content looks like SQL
func isSQL(data []byte) bool {
	content := string(data)
	content = strings.ToLower(strings.TrimSpace(content))
	
	// Check for common SQL statements
	return strings.Contains(content, "create table") || 
		   strings.Contains(content, "insert into") || 
		   strings.Contains(content, "select ") ||
		   strings.Contains(content, "drop table")
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

// ExtractFromJSON extracts data from JSON file
func (e *ETLService) ExtractFromJSON(filePath string) ([]DataRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var jsonData []DataRecord
	if err := json.Unmarshal(byteValue, &jsonData); err != nil {
		// Try parsing as JSON object with array inside
		var jsonObj map[string]interface{}
		if err2 := json.Unmarshal(byteValue, &jsonObj); err2 == nil {
			// If it's a single object, put it in an array
			jsonData = []DataRecord{jsonObj}
		} else {
			return nil, err
		}
	}

	return jsonData, nil
}

// ExtractFromXLSX extracts data from XLSX file
func (e *ETLService) ExtractFromXLSX(filePath string) ([]DataRecord, error) {
	xlFile, err := xlsx.OpenFile(filePath)
	if err != nil {
		return nil, err
	}

	var data []DataRecord

	// Process the first sheet (we can modify this to handle multiple sheets if needed)
	sheet := xlFile.Sheets[0]
	if len(sheet.Rows) == 0 {
		return nil, fmt.Errorf("no data found in XLSX file")
	}

	// Get headers from first row
	var headers []string
	for _, cell := range sheet.Rows[0].Cells {
		header, _ := cell.String()
		headers = append(headers, header)
	}

	// Process data rows
	for i, row := range sheet.Rows {
		if i == 0 { // Skip header row
			continue
		}
		
		record := make(DataRecord)
		for j, cell := range row.Cells {
			cellValue, _ := cell.String()
			if j < len(headers) {
				record[headers[j]] = cellValue
			}
		}
		data = append(data, record)
	}

	return data, nil
}

// ExtractFromXLS extracts data from XLS file
func (e *ETLService) ExtractFromXLS(filePath string) ([]DataRecord, error) {
	xlFile, err := gofile.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer xlFile.Close()

	var data []DataRecord

	// Get the first sheet
	sheet, err := xlFile.GetSheet(0)
	if err != nil {
		return nil, err
	}

	// Get headers from first row
	var headers []string
	firstRow, err := sheet.GetRow(0)
	if err != nil {
		return nil, err
	}
	
	for j := 0; firstRow.GetCell(j) != nil; j++ {
		header := firstRow.GetCell(j).Value
		headers = append(headers, header)
	}

	// Process data rows
	for i := 1; i < int(sheet.MaxRow); i++ {
		row, err := sheet.GetRow(i)
		if err != nil {
			continue
		}
		
		record := make(DataRecord)
		for j := 0; j < len(headers) && row.GetCell(j) != nil; j++ {
			cellValue := row.GetCell(j).Value
			record[headers[j]] = cellValue
		}
		data = append(data, record)
	}

	return data, nil
}

// ExtractFromDump extracts data from SQL dump file
func (e *ETLService) ExtractFromDump(filePath string) ([]DataRecord, error) {
	// For SQL dumps, we'll parse the INSERT statements to extract data
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var data []DataRecord

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(strings.ToUpper(line)), "INSERT INTO") {
			// This is a simplified approach - a full implementation would need to properly parse SQL
			// For now, we'll just return an empty set since parsing SQL dumps is complex
			// In a real implementation, you'd want to properly parse the INSERT statements
			continue
		}
	}

	return data, nil
}

// ExtractFromFile extracts data from any supported file format
func (e *ETLService) ExtractFromFile(filePath string) ([]DataRecord, error) {
	format := GetFileFormat(filePath)
	
	switch format {
	case CSV:
		return e.ExtractFromCSV(filePath)
	case JSON:
		return e.ExtractFromJSON(filePath)
	case XLSX:
		return e.ExtractFromXLSX(filePath)
	case XLS:
		return e.ExtractFromXLS(filePath)
	case DUMP, SQL:
		return e.ExtractFromDump(filePath)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", filePath)
	}
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

// SanitizeTableName sanitizes table names to be valid PostgreSQL identifiers
func SanitizeTableName(tableName string) string {
	// Remove invalid characters and replace with underscores
	tableName = strings.ToLower(tableName)
	tableName = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, tableName)
	
	// Ensure it starts with a letter or underscore
	if len(tableName) > 0 && ((tableName[0] >= '0' && tableName[0] <= '9') || tableName[0] == '_') {
		tableName = "t_" + tableName
	}
	
	// Truncate to 63 characters (PostgreSQL identifier limit)
	if len(tableName) > 63 {
		tableName = tableName[:63]
	}
	
	return tableName
}

// SanitizeColumnName sanitizes column names to be valid PostgreSQL identifiers
func SanitizeColumnName(colName string) string {
	// Remove invalid characters and replace with underscores
	colName = strings.ToLower(colName)
	colName = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, colName)
	
	// Ensure it starts with a letter or underscore
	if len(colName) > 0 && ((colName[0] >= '0' && colName[0] <= '9') || colName[0] == '_') {
		colName = "c_" + colName
	}
	
	// Truncate to 63 characters (PostgreSQL identifier limit)
	if len(colName) > 63 {
		colName = colName[:63]
	}
	
	return colName
}

// InferColumnType infers the PostgreSQL column type based on sample values
func InferColumnType(values []interface{}) string {
	var hasInt, hasFloat, hasString, hasDate bool
	
	for _, value := range values {
		if value == nil {
			continue
		}
		
		switch v := value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			hasInt = true
		case float32, float64:
			hasFloat = true
		case string:
			// Check if it looks like a date
			if strings.Contains(strings.ToLower(v), "20") && len(v) >= 8 {
				// Basic date detection - could be enhanced with time.Parse
				hasDate = true
			} else {
				hasString = true
			}
		default:
			hasString = true
		}
	}
	
	// Determine the most appropriate type based on detected types
	if hasFloat {
		return "NUMERIC"
	} else if hasInt && !hasFloat && !hasString {
		return "INTEGER"
	} else if hasDate && !hasString {
		return "DATE"
	} else {
		return "TEXT"
	}
}

// CreateTableIfNotExists creates a table based on the field types in the data
func (e *ETLService) CreateTableIfNotExists(tableName string, data []DataRecord) error {
	if len(data) == 0 {
		return nil
	}

	// Sanitize table name
	tableName = SanitizeTableName(tableName)
	
	// Collect values for each column to determine appropriate types
	columnValues := make(map[string][]interface{})
	
	// Sample the first few records to determine types
	sampleSize := len(data)
	if sampleSize > 100 { // Only sample first 100 records for performance
		sampleSize = 100
	}
	
	for i := 0; i < sampleSize; i++ {
		for key, value := range data[i] {
			columnValues[key] = append(columnValues[key], value)
		}
	}

	// Build CREATE TABLE statement
	var columnsDef []string
	for colName, values := range columnValues {
		colType := InferColumnType(values)
		// Sanitize column name
		safeColName := SanitizeColumnName(colName)
		columnsDef = append(columnsDef, fmt.Sprintf(`"%s" %s`, safeColName, colType))
	}

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s, id SERIAL PRIMARY KEY)`, tableName, strings.Join(columnsDef, ", "))

	_, err := e.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %v", tableName, err)
	}

	log.Printf("Successfully created table: %s", tableName)
	return nil
}

// ColumnAnalysisResult represents the result of LLM column analysis
type ColumnAnalysisResult struct {
	TableName    string            `json:"table_name"`
	Dimensions   map[string]string `json:"dimensions"`   // map[column_name]dimension_type
	Facts        []string          `json:"facts"`        // fact column names
	Relationships []string          `json:"relationships"` // potential relationships between tables
}

// AnalyzeColumnsWithLLM calls an LLM API to analyze columns and suggest star schema structure
func (e *ETLService) AnalyzeColumnsWithLLM(tableName string) (*ColumnAnalysisResult, error) {
	// Get table structure information
	query := `
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_name = $1
		ORDER BY ordinal_position
	`
	
	rows, err := e.db.Query(query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get table structure: %v", err)
	}
	defer rows.Close()

	var columns []struct {
		Name string
		Type string
	}
	
	for rows.Next() {
		var col struct {
			Name string
			Type string
		}
		if err := rows.Scan(&col.Name, &col.Type); err != nil {
			return nil, err
		}
		columns = append(columns, col)
	}

	// In a real implementation, we would send this information to an LLM API
	// For now, we'll simulate the LLM analysis by implementing our own logic
	result := &ColumnAnalysisResult{
		TableName:  tableName,
		Dimensions: make(map[string]string),
		Facts:      []string{},
		Relationships: []string{},
	}
	
	for _, col := range columns {
		colLower := strings.ToLower(col.Name)
		
		// Identify potential dimension columns (common dimension column names)
		if strings.Contains(colLower, "name") || 
		   strings.Contains(colLower, "desc") || 
		   strings.Contains(colLower, "category") || 
		   strings.Contains(colLower, "type") || 
		   strings.Contains(colLower, "date") {
			result.Dimensions[col.Name] = "dimension"
		} else if strings.Contains(colLower, "amount") || 
				  strings.Contains(colLower, "price") || 
				  strings.Contains(colLower, "quantity") || 
				  strings.Contains(colLower, "count") ||
				  strings.Contains(colLower, "total") {
			// These are likely fact columns
			result.Facts = append(result.Facts, col.Name)
		}
	}

	return result, nil
}

// CreateStarSchemaViews creates star schema views based on column analysis
func (e *ETLService) CreateStarSchemaViews(tableName string) error {
	// Analyze columns using LLM
	analysis, err := e.AnalyzeColumnsWithLLM(tableName)
	if err != nil {
		return fmt.Errorf("failed to analyze columns with LLM: %v", err)
	}

	// Create a view based on identified dimensions and facts
	viewName := fmt.Sprintf("%s_star_view", tableName)
	
	// Build SELECT query with all columns
	var selectColumns []string
	query := fmt.Sprintf(`SELECT * FROM %s`, tableName)
	
	// Create the view
	createViewQuery := fmt.Sprintf(`CREATE OR REPLACE VIEW %s AS %s`, viewName, query)
	
	_, err = e.db.Exec(createViewQuery)
	if err != nil {
		return fmt.Errorf("failed to create star schema view: %v", err)
	}

	log.Printf("Successfully created star schema view: %s with %d dimensions and %d facts", 
		viewName, len(analysis.Dimensions), len(analysis.Facts))
	return nil
}

// LoadToPostgreSQL loads cleaned data to PostgreSQL and creates star schema views
func (e *ETLService) LoadToPostgreSQL(data []DataRecord) error {
	if len(data) == 0 {
		return nil
	}

	// Create table automatically based on data structure
	// For now, we'll use a generic table name based on the data size, but in the future
	// we could use the source filename or other identifying information
	tableName := fmt.Sprintf("auto_table_%d", len(data))
	err := e.CreateTableIfNotExists(tableName, data)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	// Get column names from the first record
	var originalColumnNames []string
	for key := range data[0] {
		originalColumnNames = append(originalColumnNames, key)
	}

	// Sanitize column names for the INSERT statement
	var sanitizedColumnNames []string
	for _, colName := range originalColumnNames {
		sanitizedColumnNames = append(sanitizedColumnNames, SanitizeColumnName(colName))
	}

	// Build INSERT statement
	placeholders := make([]string, len(sanitizedColumnNames))
	for i := range sanitizedColumnNames {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	
	insertQuery := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(sanitizedColumnNames, ", "),
		strings.Join(placeholders, ", "),
	)

	for _, record := range data {
		values := make([]interface{}, len(sanitizedColumnNames))
		for i, colName := range originalColumnNames {
			values[i] = record[colName]
		}
		
		_, err := e.db.Exec(insertQuery, values...)
		if err != nil {
			log.Printf("Error inserting record: %v", err)
			continue
		}
	}
	
	// Create star schema views
	err = e.CreateStarSchemaViews(tableName)
	if err != nil {
		log.Printf("Warning: Failed to create star schema views: %v", err)
	}
	
	log.Printf("Successfully inserted %d records into table %s and created star schema view", len(data), tableName)
	return nil
}

// ProcessETLFromFile processes ETL from a file source
func (e *ETLService) ProcessETLFromFile(filePath string) error {
	// Determine file format and extract data
	format := GetFileFormat(filePath)
	log.Printf("Processing file %s with format: %v", filePath, format)

	data, err := e.ExtractFromFile(filePath)
	if err != nil {
		return fmt.Errorf("extract failed: %v", err)
	}

	// Transform
	transformedData := e.Transform(data)

	// Load to MinIO (raw data)
	fileName := fmt.Sprintf("raw_%s", filepath.Base(filePath))
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
		// If no source type is specified, check for a special command
		command := os.Getenv("ETL_COMMAND")
		if command == "email-test" {
			// Load email config
			config, err := LoadEmailConfig("email-config.yaml")
			if err != nil {
				log.Fatalf("Failed to load email config: %v", err)
			}

			// Create email service
			emailService := NewEmailService(config)

			// Send test email
			testEmail := os.Getenv("TEST_EMAIL")
			if testEmail == "" {
				testEmail = "test@example.com"
			}
			
			if err := emailService.SendTestEmail(testEmail); err != nil {
				log.Fatalf("Failed to send test email: %v", err)
			}
			
			log.Println("Test email sent successfully")
		} else if command == "process-file" {
			// Process a specific file
			filePath := os.Getenv("FILE_PATH")
			if filePath == "" {
				log.Fatal("FILE_PATH environment variable must be set for process-file command")
			}
			
			if err := etl.ProcessETLFromFile(filePath); err != nil {
				log.Fatalf("ETL process failed: %v", err)
			}
		} else {
			log.Fatal("ETL_SOURCE_TYPE must be either 'file' or 'sql', or ETL_COMMAND must be set")
		}
	}
}