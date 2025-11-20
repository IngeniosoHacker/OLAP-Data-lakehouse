package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

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
	Enabled  bool              `yaml:"enabled"`
	Schedule string            `yaml:"schedule"`
	Subject  string            `yaml:"subject"`
	Template string            `yaml:"template"`
}

// EmailConfig represents the complete email configuration
type EmailConfig struct {
	Recipients []Recipient            `yaml:"recipients"`
	SMTP       SMTPConfig             `yaml:"smtp"`
	Reports    map[string]ReportConfig `yaml:"reports"`
	Settings   struct {
		Timezone       string   `yaml:"timezone"`
		RetryAttempts  int      `yaml:"retry_attempts"`
		TimeoutSeconds int      `yaml:"timeout_seconds"`
		EnableLogging  bool     `yaml:"enable_logging"`
		Attachments    []string `yaml:"attachments"`
	} `yaml:"settings"`
}

// LoadEmailConfig loads the email configuration from a YAML file
func LoadEmailConfig(filePath string) (*EmailConfig, error) {
	data, err := ioutil.ReadFile(filePath)
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
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
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

func main() {
	// Example usage
	configPath := os.Getenv("EMAIL_CONFIG_PATH")
	if configPath == "" {
		configPath = "email-config.yaml"
	}

	config, err := LoadEmailConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading email config: %v", err)
	}

	// Validate the configuration
	errors := config.ValidateConfig()
	if len(errors) > 0 {
		log.Printf("Configuration validation errors:")
		for _, err := range errors {
			log.Printf("  - %s", err)
		}
		return
	}

	log.Printf("Successfully loaded email configuration with %d recipients and %d reports", 
		len(config.Recipients), len(config.Reports))

	// Example: Get weekly report recipients
	weeklyRecipients := config.GetRecipientsByReportType("weekly")
	log.Printf("Weekly report recipients: %d", len(weeklyRecipients))
	for _, recipient := range weeklyRecipients {
		subject, err := config.FormatEmailSubject("weekly", recipient)
		if err != nil {
			log.Printf("Error formatting subject for %s: %v", recipient.Name, err)
			continue
		}
		log.Printf("  - %s (%s): Subject will be '%s'", recipient.Name, recipient.Email, subject)
	}
}