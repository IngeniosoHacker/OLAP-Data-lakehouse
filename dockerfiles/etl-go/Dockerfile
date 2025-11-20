# Use the official Golang image to create the binary
FROM golang:1.21-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go mod files
COPY etl-go/go.mod etl-go/go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY etl-go/ .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o etl-go .

# Use a minimal alpine image for the final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN adduser -D -s /bin/sh appuser

# Create the application directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/etl-go .

# Change ownership to the non-root user
RUN chown -R appuser:appuser /app

# Switch to the non-root user
USER appuser

# Expose the port if needed (though ETL is command-line)
EXPOSE 8080

# Command to run the binary
CMD ["./etl-go"]