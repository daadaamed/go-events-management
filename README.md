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
 -d '{"source":"app.web","type":"user_login","payload":{"a":1,"b":3}}' | jq
```

### Get events
```bash
curl -s localhost:8080/events | jq 
```

### Get event By ID
```bash
curl -s localhost:8080/events/:id | jq 
```