# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Pocket Coder is a remote control system that allows programmers to control AI editing tools (Claude Code, Cursor, Aider) on desktop computers from mobile devices. It consists of three components:

- **Backend Server (Go)**: `/server/` - Gin web framework, GORM ORM, MySQL, Redis
- **Frontend (React)**: `/pocket-coder-app/` - React 19, TypeScript, Vite, Tailwind CSS
- **Desktop CLI (Go)**: `/cli/` - Cobra CLI framework, WebSocket client

## Build & Development Commands

### Backend Server
```bash
cd server
go mod tidy
go run cmd/server/main.go                    # Run dev server (port 8080)
CGO_ENABLED=0 go build -o server ./cmd/server # Build binary
```

### Frontend
```bash
cd pocket-coder-app
npm install
npm run dev      # Start dev server (port 3000)
npm run build    # Production build
```

### Desktop CLI
```bash
cd cli
go mod tidy
go run main.go   # Run CLI
go build -o pocket-coder .  # Build binary
```

### Docker
```bash
docker-compose up -d    # Start all services
docker-compose down     # Stop all services
```

## Architecture

```
Mobile App (React PWA) → WebSocket/HTTPS → Go Backend → WebSocket → Desktop CLI
```

### Backend Structure (`/server/`)
- `cmd/server/main.go` - Entry point, initializes DB/Redis, sets up routes
- `internal/config/` - Viper configuration management
- `internal/model/` - GORM models: User, Desktop, Session, Message
- `internal/repository/` - Data access layer
- `internal/service/` - Business logic (Auth, User, Desktop, Session)
- `internal/handler/` - HTTP handlers
- `internal/websocket/` - Hub manages mobile↔desktop message routing
- `internal/middleware/` - CORS, JWT auth, logging
- `pkg/jwt/` - JWT token generation/validation

### Frontend Structure (`/pocket-coder-app/`)
- `hooks/useAuth.tsx` - Auth context, token management in localStorage
- `hooks/usePocketWS.ts` - WebSocket connection hook
- `api/client.ts` - API client with Bearer token auth
- `pages/` - LoginPage, RegisterPage, DesktopsPage, SessionPage
- `components/Terminal.tsx` - xterm.js terminal output

### CLI Structure (`/cli/`)
- `cmd/` - Cobra commands: login, logout, status, root (interactive)
- `internal/config/` - Config at `~/.pocket-coder/config.yaml`
- `internal/websocket/` - WebSocket client for receiving instructions
- `internal/terminal/` - PTY handling for command execution

## Key API Endpoints

```
POST /api/v1/auth/login              # User login → returns JWT tokens
POST /api/v1/auth/device/code        # Request 6-digit device code
POST /api/v1/auth/device/authorize   # Authorize device with code
GET  /api/v1/desktops                # List user's devices
POST /api/v1/sessions                # Create chat session
GET  /api/v1/sessions/:id/messages   # Get conversation history
WS   /ws                             # WebSocket connection
```

## Configuration

### Backend (`/server/configs/config.yaml`)
Database and Redis credentials are configured here. Environment variables (e.g., `MYSQL_HOST`, `REDIS_HOST`, `JWT_SECRET`) override config file values.

### Frontend
Set `VITE_API_BASE` environment variable for API URL (defaults to `http://localhost:8080`).

### CLI
Credentials stored at `~/.pocket-coder/config.yaml` after device login.

## Authentication Flow

1. User logs in via mobile app → receives JWT access/refresh tokens
2. Desktop CLI requests device code → displays 6-digit code
3. User enters code in mobile app → server authorizes device
4. CLI stores device token, connects via WebSocket
5. Instructions flow: Mobile → Server → Desktop CLI → execution → results streamed back

## Database Models

- **User**: username, password_hash, email, status
- **Desktop**: user_id, device_token, name, agent_type, status, last_heartbeat
- **Session**: desktop_id, agent_type, working_dir
- **Message**: session_id, role (user/assistant), content

GORM auto-migration creates/updates tables on server startup.
