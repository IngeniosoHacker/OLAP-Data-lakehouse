# report_generation.R
# Script to generate weekly data reports and send via email

# Install required packages if not already installed
install_if_missing <- function(pkg) {
  if (!require(pkg, character.only = TRUE)) {
    install.packages(pkg, repos = "https://cran.rstudio.com/")
    library(pkg, character.only = TRUE)
  }
}

# Install and load required packages
pkgs <- c("DBI", "RPostgreSQL", "ggplot2", "dplyr", "knitr", "rmarkdown", "sendmailR")
lapply(pkgs, install_if_missing)

# Database connection parameters
db_host <- Sys.getenv("DB_HOST", unset = "postgres-service.postgres")
db_name <- Sys.getenv("DB_NAME", unset = "datalakehouse")
db_user <- Sys.getenv("DB_USER", unset = "admin")
db_password <- Sys.getenv("DB_PASSWORD", unset = "password123")

# Connect to PostgreSQL
con <- dbConnect(
  RPostgreSQL::PostgreSQL(),
  host = db_host,
  dbname = db_name,
  user = db_user,
  password = db_password
)

# Query data from the data warehouse
sales_data <- dbGetQuery(con, "SELECT * FROM sales_summary LIMIT 1000;")
dim_date <- dbGetQuery(con, "SELECT * FROM dim_date LIMIT 1000;")
fact_sales <- dbGetQuery(con, "SELECT * FROM fact_sales LIMIT 1000;")

# Close database connection
dbDisconnect(con)

# Create data visualizations
p1 <- ggplot(head(sales_data, 20), aes(x = customer_name, y = total_amount)) +
  geom_bar(stat = "identity", fill = "steelblue") +
  theme_minimal() +
  labs(title = "Top Customers by Sales Amount", x = "Customer", y = "Total Amount") +
  theme(axis.text.x = element_text(angle = 45, hjust = 1))

# Save plot to file
ggsave("customer_sales_chart.png", p1, width = 10, height = 6)

# Generate report summary statistics
summary_stats <- sales_data %>%
  summarise(
    total_sales = sum(total_amount, na.rm = TRUE),
    avg_sales = mean(total_amount, na.rm = TRUE),
    max_sales = max(total_amount, na.rm = TRUE),
    min_sales = min(total_amount, na.rm = TRUE),
    total_transactions = n()
  )

# Create report in RMarkdown format
report_content <- sprintf("
# Weekly Data Report

## Summary Statistics
- Total Sales: $%.2f
- Average Sale: $%.2f
- Max Sale: $%.2f
- Min Sale: $%.2f
- Total Transactions: %d

## Visualizations
![Customer Sales Chart](customer_sales_chart.png)

## Top 10 Sales Records
```{r, echo=FALSE}
head(sales_data, 10)
```

", 
summary_stats$total_sales, 
summary_stats$avg_sales,
summary_stats$max_sales,
summary_stats$min_sales,
summary_stats$total_transactions)

# Write the report content to an Rmd file
writeLines(report_content, "weekly_report.Rmd")

# Convert Rmd to PDF
rmarkdown::render("weekly_report.Rmd", output_format = "pdf_document", output_file = "weekly_data_report.pdf")

# Send email with the report using system mail command (requires proper mail configuration on the system)
recipient_email <- Sys.getenv("RECIPIENT_EMAIL", unset = "recipient@example.com")

# The actual email sending would happen through an external service or properly configured mail system
# For demonstration purposes, we'll just note that the email would be sent
cat("Report generated. Email would be sent to:", recipient_email, "\n")

# Alternative approach: use a system mail command if available
# This requires proper mail configuration on the container/system
# system(sprintf("echo 'Weekly report attached' | mail -s 'Weekly Data Report' -A weekly_data_report.pdf %s", recipient_email))

# Print completion message
cat("Weekly report generated successfully: weekly_data_report.pdf\n")