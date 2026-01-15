# Build stage
FROM golang:1.23-alpine AS builder

# Install git and other dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o mgsearch .

# Run stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/mgsearch .

# Copy any other necessary files (e.g., templates, static files if any)
# COPY --from=builder /app/config ./config

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./mgsearch"]
