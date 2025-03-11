# # Use official Go image for building the application
# FROM golang:1.23.7 AS builder

# # Set the working directory inside the container
# WORKDIR /app

# # Copy go.mod and go.sum for dependency resolution
# COPY go.mod go.sum ./

# # Download dependencies
# RUN go mod download

# # Copy the rest of the application source code
# COPY . .

# # Build the Go application
# RUN make build

# # Use a minimal base image for the final container
# FROM alpine:latest

# # Set working directory
# WORKDIR /root/

# # Copy the compiled binary from the builder stage
# COPY --from=builder /app/goFirstProject .

# # Expose the port the app runs on
# EXPOSE 8080

# # Command to run the application
# CMD ["./app/goFirstProject"]


# Use official Go image for building the application
FROM golang:1.23.7 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum for dependency resolution
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install SQLC for generating Go code from SQL queries
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Copy the rest of the application source code
COPY . .

# Generate SQLC code
RUN make generateSQLC

# # Build the Go application
 RUN make build
# Build the Go application
# RUN go build -o goFirstProject main.go  # Ensure "main.go" exists

# Use a minimal base image for the final container
# FROM alpine:latest

# Set working directory
# WORKDIR /root/

# Copy the compiled binary from the builder stage
# COPY --from=builder /app/goFirstProject ./goFirstProject

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./goFirstProject"]

