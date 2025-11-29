.PHONY: dev dev-backend dev-frontend build build-backend build-frontend docker docker-build docker-up docker-down clean help setup setup-hooks wails-dev wails-build

# Default target
help:
	@echo "Noted - Development Commands"
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
	@echo "Native App (Wails):"
	@echo "  make wails-build   - Build native macOS .app bundle"
	@echo "  make wails-install - Install Noted.app to /Applications"
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

# Setup
setup: setup-hooks
	@echo "Installing dependencies..."
	cd backend && go mod download
	cd frontend && npm install
	@echo ""
	@echo "Setup complete! Run 'make dev' to start development servers."

setup-hooks:
	@echo "Configuring git hooks..."
	git config core.hooksPath .githooks
	chmod +x .githooks/pre-commit .githooks/post-commit .githooks/commit-msg
	@echo "Git hooks configured!"

# Wails (Native App)
wails-dev:
	@echo "Starting Wails development mode..."
	@echo "Note: For development, use 'make dev' instead. Wails dev mode is for testing the native wrapper."
	cd backend/cmd/wails && ~/go/bin/wails dev

wails-build:
	@echo "Building native macOS app..."
	cd frontend && npm run build
	rm -rf backend/cmd/wails/frontend
	cp -r frontend/build backend/cmd/wails/frontend
	cd backend/cmd/wails && ~/go/bin/wails build -platform darwin/arm64
	@echo ""
	@echo "Build complete! App is at: backend/cmd/wails/build/bin/Noted.app"
	@echo "To install: cp -r backend/cmd/wails/build/bin/Noted.app /Applications/"

wails-install:
	@echo "Installing Noted.app to /Applications..."
	cp -r backend/cmd/wails/build/bin/Noted.app /Applications/
	@echo "Done! You can now launch Noted from your Applications folder."

# Clean
clean:
	rm -f backend/server
	rm -rf frontend/build
	rm -rf frontend/.svelte-kit
	rm -rf .dashcode_queue
	rm -rf backend/cmd/wails/build
	rm -rf backend/cmd/wails/frontend
