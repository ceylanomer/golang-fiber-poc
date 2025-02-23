# Golang Fiber POC

This project is a proof of concept (POC) for a web server built using the [Fiber](https://gofiber.io/) web framework in Go. It includes basic CRUD operations for a product entity, health checks, metrics, and tracing.

## Features

- Health check endpoint
- Metrics endpoint
- Basic authentication for product endpoints
- CRUD operations for products
- OpenTelemetry tracing
- Graceful shutdown

## Requirements

- Go 1.18+
- Couchbase server
- OpenTelemetry Collector

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/golang-fiber-poc.git
    cd golang-fiber-poc
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```

3. Set up Couchbase and OpenTelemetry Collector as per your environment.

## Configuration

Configure the application by setting the following environment variables:

- `APP_PORT`: Port on which the server will run (default: `8080`)

## Running the Application

Start the server:
```sh
go run main.go