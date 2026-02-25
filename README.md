# ZeitPass Backend (Go)

Go backend API for ZeitPass platform, replacing Airtable + Make.com webhooks.

## Tech Stack

- **Framework**: Fiber (Express-like)
- **Database**: PostgreSQL
- **ORM**: GORM
- **Config**: godotenv

## Local Development

### Prerequisites

- Go 1.22+
- PostgreSQL 16+
- Make (optional)

### Setup

```bash
# Install dependencies
go mod download

# Copy environment file
cp .env.example .env

# Edit .env with your database credentials

# Run database migrations
go run cmd/migrate/main.go

# Start development server
go run cmd/api/main.go
```

Server runs on http://localhost:8080

### API Endpoints

- `GET /health` - Health check
- `GET /api/v1/curation?uid=<uid>` - Get curated events for user
- `GET /api/v1/events/:eventId` - Get event details
- `POST /api/v1/bookings` - Create booking
- `POST /api/v1/bookings/:bookingId/feedback` - Submit feedback
- `GET /api/v1/profile/:uid` - Get user profile
- `POST /api/v1/activity` - Log activity

## Production Deployment

### Build

```bash
# Build binary
go build -o zeitpass-api ./cmd/api

# Or with optimizations
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o zeitpass-api ./cmd/api
```

### Deploy

```bash
# Copy binary to server
scp zeitpass-api user@77.42.64.36:/opt/zeitpass/

# Copy .env
scp .env.production user@77.42.64.36:/opt/zeitpass/.env

# SSH and start service
ssh user@77.42.64.36
sudo systemctl restart zeitpass
```

## Project Structure

```
.
├── cmd/
│   ├── api/          # Main application entry point
│   └── migrate/      # Database migration tool
├── internal/
│   ├── handlers/     # HTTP request handlers
│   ├── services/     # Business logic layer
│   ├── repository/   # Database access layer
│   ├── models/       # Data models
│   └── middleware/   # HTTP middleware
├── migrations/       # SQL migrations
├── scripts/          # Utility scripts
├── go.mod
└── README.md
```

## License

Proprietary - 11data.ai
