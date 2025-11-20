# Use the official PostgreSQL image
FROM postgres:15

# Set environment variables
ENV POSTGRES_DB=datalakehouse
ENV POSTGRES_USER=admin
ENV POSTGRES_PASSWORD=password123

# Copy initialization script
COPY init-db.sql /docker-entrypoint-initdb.d/

# Expose PostgreSQL port
EXPOSE 5432

# The container will start PostgreSQL automatically