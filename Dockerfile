# Stage 1: Builder
# Use a specific and small Go image for the builder
FROM --platform=linux/amd64 golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go module files
# COPY go.mod ./
RUN go mod init dps-scanner-gateout

# Download and cache go modules. This layer is only rebuilt when go.mod or go.sum change.
RUN go mod download
RUN go mod verify

# Copy the rest of the application source code
COPY . .

# Ensure all dependencies are clean and tidy
RUN go mod tidy

# Build the Go application, creating a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/main .


# Stage 2: Final Image
# Use a minimal base image for a small and secure final image
FROM --platform=linux/amd64 alpine:latest

# Set the timezone
ENV TZ=Asia/Jakarta
RUN apk add --no-cache tzdata

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Copy static assets and configuration files
COPY --from=builder /app/static ./static/
COPY --from=builder /app/.env .
COPY --from=builder /app/layout.json .

# Expose the port the application runs on (adjust if different)
EXPOSE 8080

# The command to run the application
ENTRYPOINT ["/app/main"]
