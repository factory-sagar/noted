.PHONY: dev dev-backend dev-frontend build build-backend build-frontend docker docker-build docker-up docker-down clean help

# Default target
help:
	@echo "SE Notes - Development Commands"
	@echo ""
	@echo "Development:"
	@echo "  make dev           - Run both backend and frontend in development mode"
	@echo "  make dev-backend   - Run backend only"
	@echo "  make dev-frontend  - Run frontend only"
	@echo ""
	@echo "Build:"
	@echo "  make build         - Build both backend and frontend"
	@echo "  make build-backend - Build backend binary"
	@echo "  make build-frontend- Build frontend for production"
	@echo ""
	@echo "Docker:"
	@echo "  make docker        - Build and run with Docker Compose"
	@echo "  make docker-build  - Build Docker images"
	@echo "  make docker-up     - Start Docker containers"
	@echo "  make docker-down   - Stop Docker containers"
	@echo ""
	@echo "Other:"
	@echo "  make clean         - Remove build artifacts"

# Development
dev:
	@echo "Starting development servers..."
	@make -j2 dev-backend dev-frontend

dev-backend:
	@echo "Starting backend on :8080..."
	cd backend && go run ./cmd/server/

dev-frontend:
	@echo "Starting frontend on :5173..."
	cd frontend && npm run dev

# Build
build: build-backend build-frontend

build-backend:
	@echo "Building backend..."
	cd backend && go build -o server ./cmd/server/

build-frontend:
	@echo "Building frontend..."
	cd frontend && npm run build

# Docker
docker: docker-build docker-up

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d
	@echo ""
	@echo "Application is running!"
	@echo "Frontend: http://localhost:3000"
	@echo "Backend:  http://localhost:8080"

docker-down:
	docker-compose down

# Clean
clean:
	rm -f backend/server
	rm -rf frontend/build
	rm -rf frontend/.svelte-kit
