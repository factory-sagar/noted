# Skill: Docker Deployment

## Overview
This skill covers containerizing and deploying SE Notes with Docker.

## Container Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Docker Network                        │
│                                                          │
│  ┌──────────────────┐      ┌──────────────────────┐    │
│  │    Frontend      │      │      Backend         │    │
│  │  (Node.js/Svelte)│      │    (Go/Gin)          │    │
│  │                  │      │                      │    │
│  │  Port: 3000      │─────►│  Port: 8080          │    │
│  │  (external)      │      │  (external)          │    │
│  └──────────────────┘      └──────────┬───────────┘    │
│                                        │                │
│                            ┌───────────▼───────────┐   │
│                            │   Volume: ./data      │   │
│                            │   (SQLite database)   │   │
│                            └───────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

## Files

### docker-compose.yml
```yaml
version: '3.8'

services:
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data    # Persist database
    environment:
      - PORT=8080
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    depends_on:
      backend:
        condition: service_healthy
    environment:
      - NODE_ENV=production
    restart: unless-stopped
```

### Backend Dockerfile
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/server/

# Runtime stage
FROM alpine:latest
RUN apk add --no-cache sqlite-libs ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
RUN mkdir -p /app/data
EXPOSE 8080
CMD ["./server"]
```

### Frontend Dockerfile
```dockerfile
# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Runtime stage
FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/build ./build
COPY --from=builder /app/package*.json ./
COPY --from=builder /app/node_modules ./node_modules
EXPOSE 3000
CMD ["node", "build"]
```

## Commands

### Build and Start
```bash
# Build images
docker-compose build

# Start containers (detached)
docker-compose up -d

# Build and start in one command
docker-compose up -d --build
```

### Stop and Remove
```bash
# Stop containers
docker-compose down

# Stop and remove volumes
docker-compose down -v

# Stop, remove, and delete images
docker-compose down --rmi all
```

### Logs and Debugging
```bash
# View all logs
docker-compose logs

# Follow logs
docker-compose logs -f

# Logs for specific service
docker-compose logs backend
docker-compose logs frontend

# Last N lines
docker-compose logs --tail=50
```

### Container Management
```bash
# List running containers
docker-compose ps

# Restart a service
docker-compose restart backend

# Execute command in container
docker-compose exec backend sh
docker-compose exec frontend sh

# View resource usage
docker stats
```

## Health Checks

Backend health endpoint: `GET /health`

```bash
# Test from host
curl http://localhost:8080/health
# Response: {"status":"ok"}

# Test from within container
docker-compose exec backend wget -qO- http://localhost:8080/health
```

## Data Persistence

Database is stored in `./data/notes.db` and mounted as a volume.

```bash
# Backup database
cp ./data/notes.db ./backups/notes-$(date +%Y%m%d).db

# Reset database (will recreate on next start)
rm ./data/notes.db
docker-compose restart backend
```

## Environment Variables

### Backend
| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | Server port |

### Frontend
| Variable | Default | Description |
|----------|---------|-------------|
| NODE_ENV | production | Environment mode |
| PORT | 3000 | Server port |
| HOST | 0.0.0.0 | Bind address |

## Troubleshooting

### Container won't start
```bash
# Check logs
docker-compose logs backend

# Common issues:
# - Port already in use: lsof -ti:8080 | xargs kill
# - Missing dependencies: docker-compose build --no-cache
```

### Database errors
```bash
# Check if data directory exists
ls -la ./data

# Reset database
rm ./data/notes.db
docker-compose restart backend
```

### Network issues between containers
```bash
# Frontend can't reach backend
# Check if backend is healthy
docker-compose ps

# Test connectivity
docker-compose exec frontend wget -qO- http://backend:8080/health
```

### Rebuilding after code changes
```bash
# Rebuild specific service
docker-compose build backend
docker-compose up -d backend

# Rebuild all
docker-compose up -d --build
```

## Production Considerations

1. **Reverse proxy**: Add nginx/traefik for SSL termination
2. **Database backup**: Schedule regular backups of `./data`
3. **Logging**: Consider adding log aggregation
4. **Monitoring**: Add health check monitoring
5. **Resources**: Set memory/CPU limits in docker-compose
