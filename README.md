# Event Management API

A simple Go + Gin + MongoDB event management service.

## Quick Start

Using Docker
```bash
docker-compose up --build
```

## API Endpoints

GET /events - List events (limit: 50)

GET /events/:id - Get event by ID

POST /events - Create/update event

## Development
```bash
go mod tidy
go run main.go
```

### Add an event
```bash
 curl -sX POST localhost:8080/events -H 'Content-Type: application/json' \
 -d '{"source":"app.web","type":"user_login","payload":{"level":"warn","user_id":"id_1","system_status":"healthy"}}' | jq
```

### Get events
```bash
curl -s localhost:8080/events | jq 

# Get recent events with pagination
curl "localhost:8080/events?limit=20&offset=0"

# Filter by source and type
curl "localhost:8080/events?source=webapp&type=user_signup"

# Filter by time range
curl "localhost:8080/events?from=2024-01-01T00:00:00Z&to=2024-01-31T23:59:59Z"

# Combine filters
curl "localhost:8080/events?source=webapp&type=click&from=2024-01-15T00:00:00Z&limit=50"

```

### Get event By ID
```bash
curl -s localhost:8080/events/:id | jq 
```

### Run Test
```bash
go test ./... -v
```

### Database choice

MongoDB was chosen for flexibility ( payload could contain different datas) and futur improvements for the app.
Atomic upsert with $setOnInsert + $inc enables counter increments easily.

### Handling concurrent request 

The unique index prevents duplicates.

The single upsert ensures count increments are not lost even with many concurrent writers.

Timestamps: first_added stays fixed; last_added/updated_at update on each ingestion.

### Improvements (features)

Rate Limiting

Strict Payload Validation

Authentication and authorization

Caching

API Versioning

#### Metrics and monitoring

Ingestion, Database, API metrics
Error rates and types
Alerting rules
Identify slow queries and bottlenecks