package main

import (
	"bytes"
	"fmt"
	"log"
	"text/template"
	"time"

	"gopkg.in/mail.v2"
)

// EmailService handles email operations
type EmailService struct {
	config *EmailConfig
}

// NewEmailService creates a new email service
func NewEmailService(config *EmailConfig) *EmailService {
	return &EmailService{
		config: config,
	}
}

// SendEmail sends a personalized email to a recipient
func (es *EmailService) SendEmail(recipient Recipient, reportType string, attachments []string) error {
	// Format email subject and body
	subject, err := es.config.FormatEmailSubject(reportType, recipient)
	if err != nil {
		return fmt.Errorf("failed to format email subject: %v", err)
	}

	body, err := es.config.FormatEmailBody(reportType, recipient)
	if err != nil {
		return fmt.Errorf("failed to format email body: %v", err)
	}

	// Create message
	m := mail.NewMessage()
	m.SetHeader("From", es.config.SMTP.From)
	m.SetHeader("To", recipient.Email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	// Add attachments
	for _, attachmentPath := range attachments {
		if attachmentPath != "" {
			m.Attach(attachmentPath)
		}
	}

	// Create SMTP dialer
	port := es.config.SMTP.Port
	d := mail.NewDialer(es.config.SMTP.Host, port, es.config.SMTP.Username, es.config.SMTP.Password)

	// Send email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email to %s: %v", recipient.Email, err)
	}

	log.Printf("Successfully sent %s report to %s (%s)", reportType, recipient.Name, recipient.Email)
	return nil
}

// SendReportEmails sends reports to all configured recipients for a given report type
func (es *EmailService) SendReportEmails(reportType string, attachments []string) error {
	// Get report config
	reportConfig, exists := es.config.Reports[reportType]
	if !exists {
		return fmt.Errorf("report type '%s' not found in config", reportType)
	}

	if !reportConfig.Enabled {
		log.Printf("Report type '%s' is disabled, skipping", reportType)
		return nil
	}

	// Get recipients for this report type
	recipients := es.config.GetRecipientsByReportType(reportType)
	if len(recipients) == 0 {
		log.Printf("No recipients configured for report type '%s'", reportType)
		return nil
	}

	log.Printf("Sending %s reports to %d recipients", reportType, len(recipients))

	// Send email to each recipient
	var successCount, errorCount int
	for _, recipient := range recipients {
		if err := es.SendEmail(recipient, reportType, attachments); err != nil {
			log.Printf("Error sending %s report to %s: %v", reportType, recipient.Name, err)
			errorCount++
		} else {
			successCount++
		}

		// Small delay between emails to avoid overwhelming the SMTP server
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("Completed sending %s reports: %d successful, %d failed", reportType, successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("failed to send %d of %d emails", errorCount, len(recipients))
	}

	return nil
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

// SendPersonalizedEmail sends an email with custom subject and body to a recipient
func (es *EmailService) SendPersonalizedEmail(recipient Recipient, subjectTemplate, bodyTemplate string, data map[string]interface{}) error {
	// Merge recipient data with custom data
	templateData := make(map[string]interface{})
	for k, v := range data {
		templateData[k] = v
	}
	
	// Add recipient-specific data
	templateData["name"] = recipient.Name
	templateData["email"] = recipient.Email
	templateData["date"] = time.Now().Format("2006-01-02")
	templateData["month"] = time.Now().Format("January 2006")

	// Format subject and body
	subject, err := executeTemplate(subjectTemplate, templateData)
	if err != nil {
		return fmt.Errorf("failed to format email subject: %v", err)
	}

	body, err := executeTemplate(bodyTemplate, templateData)
	if err != nil {
		return fmt.Errorf("failed to format email body: %v", err)
	}

	// Create message
	m := mail.NewMessage()
	m.SetHeader("From", es.config.SMTP.From)
	m.SetHeader("To", recipient.Email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	// Create SMTP dialer
	port := es.config.SMTP.Port
	d := mail.NewDialer(es.config.SMTP.Host, port, es.config.SMTP.Username, es.config.SMTP.Password)

	// Send email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send personalized email to %s: %v", recipient.Email, err)
	}

	log.Printf("Successfully sent personalized email to %s (%s)", recipient.Name, recipient.Email)
	return nil
}

// executeTemplate executes a text template with the given data
func executeTemplate(tmplStr string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("email").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	return buf.String(), nil
}

// ScheduleReportSending schedules report sending based on cron-like expressions
// This is a simplified version - in production, you'd use a proper scheduler
func (es *EmailService) ScheduleReportSending(reportType string, attachments []string) error {
	reportConfig, exists := es.config.Reports[reportType]
	if !exists {
		return fmt.Errorf("report type '%s' not found in config", reportType)
	}

	if !reportConfig.Enabled {
		log.Printf("Report type '%s' is disabled, not scheduling", reportType)
		return nil
	}

	log.Printf("Scheduling %s report with schedule: %s", reportType, reportConfig.Schedule)
	
	// In a real implementation, you would parse the cron expression and schedule
	// the task using a scheduler library like robfig/cron
	// For now, we'll just log that the report is scheduled
	
	return nil
}