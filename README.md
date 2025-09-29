# Entain Test Task

A Go-based transaction processing service that handles user balance management with concurrent transaction processing capabilities.

## API Endpoints

- `POST /user/{userId}/transaction` - Process a transaction for a user
- `GET /user/{userId}/balance` - Get current user balance

## Prerequisites

- Docker and Docker Compose
- Go 1.25

## How to Run

### Quick Start with Docker Compose

```bash
# Build the go application image
docker compose build

# Start the application and database
docker-compose up -d

# The service will be available at http://localhost:3000
```

This will:

1. Start PostgreSQL database
2. Run database migrations
3. Insert predefined users (IDs 1, 2, 3, 4)
4. Start the HTTP server on port 3000

## How to Test

### Unit Tests

Run all unit tests:

```bash
go test ./internal/... -v
```

Run specific package tests:

```bash
# Service layer tests
go test ./internal/service -v

# Model tests
go test ./internal/model -v

# Handler tests
go test ./internal/handler -v
```

### E2E API Tests

The API tests require a running PostgreSQL instance.

```bash
# Start dependencies
docker-compose up -d

# Run API tests
go test ./tests/api -v
```

### Performance Tests

Run performance tests specifically:

```bash
RUN_PERFORMANCE_TESTS=1 go test ./tests/api -run TestPerformance -v
```

## Example Usage

### Process a Transaction

```bash
curl -X POST http://localhost:3000/user/1/transaction \
  -H "Source-Type: game" \
  -H "Content-Type: application/json" \
  -d '{
    "state": "win",
    "amount": "10.15",
    "transactionId": "txn_123456789"
  }'
```

### Get User Balance

```bash
curl http://localhost:3000/user/1/balance
```

Response:

```json
{
  "userId": 1,
  "balance": "110.15"
}
```

## Project Structure

```text
├── cmd/main.go                    # Application entry point
├── internal/
│   ├── config/config.go           # Configuration management
│   ├── handler/                   # HTTP handlers
│   ├── http/http.go               # HTTP server setup
│   ├── model/                     # Data models and validation
│   ├── repository/                # Database operations
│   └── service/                   # Business logic
├── migrations/                    # Database migrations
├── tests/api/                     # End-to-end API tests
├── compose.yaml                   # Docker Compose configuration
└── Dockerfile                     # Container build configuration
```

## Configuration

The application uses environment variables:

| Variable    | Default   | Description       |
| ----------- | --------- | ----------------- |
| DB_HOST     | localhost | PostgreSQL host   |
| DB_PORT     | 5432      | PostgreSQL port   |
| DB_USER     | postgres  | Database username |
| DB_PASSWORD | password  | Database password |
| DB_NAME     | database  | Database name     |
| SERVER_PORT | 3000      | HTTP server port  |

## Database Schema

- **users**: Stores user balances with non-negative constraint
- **transactions**: Stores all processed transactions with deduplication

Initial users are created with IDs 1-4 and starting balances.

## Development

### Code Quality

Run linting:

```bash
golangci-lint run
```

## Notes

- Transaction IDs must be unique to prevent duplicate processing
- User balances cannot go negative
- The service is designed to handle at least 50 requests per second
- All monetary amounts are handled as strings with up to 2 decimal places
