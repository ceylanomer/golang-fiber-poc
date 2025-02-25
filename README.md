# Golang Fiber POC

This project is a proof of concept (POC) for a web server built using the [Fiber](https://gofiber.io/) web framework in Go. It includes basic CRUD operations for product entities, health checks, metrics, and tracing features.

## Features

- Health check endpoint
- Metrics endpoint
- Basic authentication for product endpoints
- CRUD operations for products
- OpenTelemetry tracing
- Graceful shutdown
- Circuit breaker pattern implementation
- Retryable HTTP client
- Prometheus metrics collection
- Grafana dashboards for visualization
- Kubernetes deployment support

## Requirements

- Go 1.18+
- Couchbase server
- OpenTelemetry Collector
- Docker and Docker Compose (for local development)
- Kubernetes (for deployment)

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

3. Set up infrastructure using Docker Compose:
    ```sh
    docker-compose up -d
    ```
   This will start Prometheus, Grafana, Jaeger, and Couchbase.

4. Configure Couchbase:
   - Access Couchbase dashboard at http://localhost:8091
   - Create a new bucket named "products"
   - Set up credentials (default: Administrator/123456789)

## Configuration

Configure the application by setting the following environment variables:

- `APP_PORT`: Port on which the server will run (default: `8080`)
- `COUCHBASE_HOST`: Couchbase server host (default: `localhost`)
- `COUCHBASE_USERNAME`: Couchbase username (default: `Administrator`)
- `COUCHBASE_PASSWORD`: Couchbase password (default: `123456789`)
- `OTEL_EXPORTER_OTLP_ENDPOINT`: OpenTelemetry collector endpoint (default: `localhost:4318`)

You can also configure the application through the `config/config.yaml` file.

## Running the Application

Start the server:
```sh
go run main.go
```

## API Endpoints

### General Endpoints

- `GET /healthcheck` - Health check endpoint
- `GET /metrics` - Prometheus metrics endpoint
- `GET /` - Simple hello world endpoint

### Product Endpoints (Requires Basic Auth)

All product endpoints are protected with basic authentication:
- Username: `admin`
- Password: `password`

Endpoints:
- `GET /api/v1/product/:id` - Get a product by ID
- `POST /api/v1/product` - Create a new product
- `PUT /api/v1/product/:id` - Update an existing product

### Example Requests

#### Create Product
```sh
curl -X POST http://localhost:8080/api/v1/product \
  -u admin:password \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Product"}'
```

#### Get Product
```sh
curl -X GET http://localhost:8080/api/v1/product/{id} \
  -u admin:password
```

#### Update Product
```sh
curl -X PUT http://localhost:8080/api/v1/product/{id} \
  -u admin:password \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Product"}'
```

## Observability

### Prometheus and Grafana

- Prometheus is available at http://localhost:9090
- Grafana is available at http://localhost:3000 (admin/admin)
- Pre-configured Go metrics dashboard is provided

### Distributed Tracing

- Jaeger UI is available at http://localhost:16686
- Application sends OpenTelemetry traces to collector at http://localhost:4318

## Docker

The project includes a Dockerfile for containerization. Build the Docker image:

```sh
docker build -t golang-fiber-poc .
```

Run the container:
```sh
docker run -p 8080:8080 golang-fiber-poc
```

## Kubernetes Deployment

Kubernetes manifest files are provided in the `.deploy/kubernetes` directory:

1. Apply deployment:
   ```sh
   kubectl apply -f .deploy/kubernetes/deployment.yaml
   ```

2. Apply service:
   ```sh
   kubectl apply -f .deploy/kubernetes/service.yaml
   ```

3. Configure autoscaling (HPA):
   ```sh
   kubectl apply -f .deploy/kubernetes/hpa.yaml
   ```

## Project Structure

```
├── app/                  # Application logic
│   ├── client/           # HTTP client implementations
│   ├── healthcheck/      # Health check handler
│   └── product/          # Product domain handlers
├── config/               # Configuration files
├── .deploy/              # Deployment configurations
│   ├── grafana/          # Grafana dashboards
│   ├── kubernetes/       # Kubernetes manifests
│   └── prometheus/       # Prometheus configuration
├── domain/               # Domain entities
├── infra/                # Infrastructure implementations
│   └── couchbase/        # Couchbase repository
├── pkg/                  # Shared packages
│   ├── circuitbreaker/   # Circuit breaker implementation
│   ├── config/           # Configuration loader
│   ├── customvalidator/  # Request validation
│   ├── handler/          # Generic handler
│   ├── log/              # Logging setup
│   ├── middlewares/      # Middleware implementations
│   └── tracer/           # OpenTelemetry tracer setup
├── docker-compose.yml    # Docker Compose configuration
├── Dockerfile            # Docker build configuration
├── go.mod                # Go module definition
├── go.sum                # Go dependencies checksum
└── main.go               # Application entry point
```

## Contributing

We welcome your contributions! Please feel free to submit Pull Requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.