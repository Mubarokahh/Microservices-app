# Dockerfile for PostgreSQL
FROM postgres:13-alpine

# Set environment variables for PostgreSQL
ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=Barakat1243
ENV POSTGRES_DB=postgres

# Expose the PostgreSQL port
EXPOSE 5432

# Start PostgreSQL
CMD ["postgres"]