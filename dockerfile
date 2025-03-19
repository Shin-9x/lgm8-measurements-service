# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files from the microservice directory
COPY microservices/lgm8-measurements-service/go.mod microservices/lgm8-measurements-service/go.sum ./

# Download all dependencies
RUN go mod download

# Copy the rest of the source code
COPY microservices/lgm8-measurements-service/ ./

# Compile the application
RUN CGO_ENABLED=0 GOOS=linux go build -o lgm8-measurements-service ./cmd/main.go

# Stage 2: Create a minimal runtime image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the compiled binary
COPY --from=builder /app/lgm8-measurements-service /app/lgm8-measurements-service

# Copy configuration files
COPY microservices/lgm8-measurements-service/config/ /app/config/

# Expose the port your application listens on
EXPOSE 8082

# Run the application
CMD ["./lgm8-measurements-service"]