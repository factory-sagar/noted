# Setup Guide

## Prerequisites

### Required
- **Go 1.21+** - [Install Go](https://go.dev/doc/install)
- **Node.js 20+** - [Install Node.js](https://nodejs.org/)
- **SQLite** - Usually pre-installed on macOS/Linux

### Optional
- **Docker** - For containerized deployment
- **Make** - For simplified commands

## Development Setup

### 1. Clone the Repository
```bash
git clone https://github.com/factory-sagar/notes-droid.git
cd notes-droid
```

### 2. Backend Setup
```bash
cd backend

# Download dependencies
go mod download

# Run the server
go run ./cmd/server/

# Server starts on http://localhost:8080
```

The SQLite database is automatically created at `./data/notes.db`.

### 3. Frontend Setup
```bash
cd frontend

# Install dependencies
npm install

# Run development server
npm run dev

# App starts on http://localhost:5173
```

### 4. Verify Setup
- Open http://localhost:5173 in your browser
- The dashboard should load (may show "Failed to load" until backend is running)
- Create an account and note to test

## Production Deployment

### Using Docker Compose (Recommended)

```bash
# Build and start containers
docker-compose up -d

# View logs
docker-compose logs -f

# Stop containers
docker-compose down
```

Access the app at http://localhost:3000

### Manual Build

#### Backend
```bash
cd backend
go build -o server ./cmd/server/
./server
```

#### Frontend
```bash
cd frontend
npm run build
npm run preview  # Or deploy the 'build' folder
```

## Google Calendar Integration (Coming Soon)

### 1. Create Google Cloud Project
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project
3. Enable the Google Calendar API

### 2. Create OAuth Credentials
1. Go to APIs & Services > Credentials
2. Create OAuth 2.0 Client ID
3. Set authorized redirect URI to `http://localhost:8080/api/auth/google/callback`

### 3. Configure Environment
```bash
export GOOGLE_CLIENT_ID=your-client-id
export GOOGLE_CLIENT_SECRET=your-client-secret
```

## Troubleshooting

### Backend won't start
- Check if port 8080 is in use: `lsof -i :8080`
- Ensure Go is installed: `go version`
- Check for CGO issues (needed for SQLite): `CGO_ENABLED=1 go build`

### Frontend won't start
- Check if port 5173 is in use: `lsof -i :5173`
- Clear node_modules and reinstall: `rm -rf node_modules && npm install`
- Clear SvelteKit cache: `rm -rf .svelte-kit`

### Database issues
- Database file is at `./data/notes.db`
- To reset: delete the file and restart backend
- Check permissions on the data directory

### CORS errors
- Ensure backend is running on port 8080
- Check browser console for specific errors
- Verify CORS configuration in `backend/cmd/server/main.go`

## Development Tips

### Hot Reload
- Backend: Use `air` for auto-reload (`go install github.com/air-verse/air@latest`)
- Frontend: Built-in with Vite

### Database Inspection
```bash
# Open SQLite CLI
sqlite3 data/notes.db

# List tables
.tables

# View schema
.schema notes
```

### API Testing
```bash
# Health check
curl http://localhost:8080/health

# List accounts
curl http://localhost:8080/api/accounts

# Create account
curl -X POST http://localhost:8080/api/accounts \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Account"}'
```
