# Base image for building the Go app
FROM golang:1.20-alpine as builder

# Enable Go modules
ENV GO111MODULE=on

# Set the working directory inside the container
WORKDIR /results-service

# Use Go proxy to avoid network issues (optional)
ENV GOPROXY=https://proxy.golang.org,direct

# Copy go.mod and go.sum files to the working directory
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the entire project to the working directory
COPY . .

# Build the Go application
RUN go build -o voting-results main.go

# Final image stage for running the Go app
FROM alpine:3.18

# Install necessary CA certificates and dependencies
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /results-service

# Copy the compiled binary from the builder image to the final stage
COPY --from=builder /results-service/voting-results /results-service/voting-results

# Copy static files and templates
COPY static /results-service/static
COPY templates /results-service/templates

# Copy .env file (ensure it's configured with the correct Redis host in production)
# COPY .env /results-service/.env

# Expose the port the app runs on
EXPOSE 8085

# Command to run the voting-service
CMD ["./voting-results"]