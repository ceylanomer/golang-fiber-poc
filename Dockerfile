# Use the official Golang image as the base image
FROM golang:1.24-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -o main .

# Start a new stage from scratch
FROM alpine:3.19

# Set the Current Working Directory inside the container
WORKDIR /app

RUN adduser -D appuser
USER appuser

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .
COPY config/config.yaml .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]