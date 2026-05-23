# Feature Flags System

A personal learning project exploring Azion Edge Functions with a feature flags system. Consists of two main components:

- **Edge Function** (`function/`) — TypeScript/Hono service for serving flag snapshots at the edge
- **Go API** (`api/`) — REST API for managing flags, backed by Redis

## Quick Start

### Edge Function

```bash
cd function
npm install
npm run dev
```

See [`function/README.md`](./function/README.md) for full setup and deployment.

### Go API

```bash
cd api
go mod tidy
go run ./cmd/api
```

See [`api/README.md`](./api/README.md) for full setup and deployment.

## Architecture

Refer to:
- [`PRD.md`](./PRD.md) — Complete product specification
- [`function/ARCHITECTURE.md`](./function/ARCHITECTURE.md) — Edge Function design
- [`api/ARCHITECTURE.md`](./api/ARCHITECTURE.md) — Go API design

## Project Structure

```
flags/
├── PRD.md                    # Product specification
├── README.md                 # This file
├── .gitignore
├── function/                 # Edge Function (TypeScript/Hono)
│   ├── ARCHITECTURE.md
│   ├── README.md
│   ├── package.json
│   ├── src/
│   ├── azion/
│   └── scripts/
└── api/                      # Go API (REST)
    ├── ARCHITECTURE.md
    ├── README.md
    ├── go.mod
    ├── cmd/
    ├── internal/
    └── Dockerfile
```

## Status

This is an MVP learning project. No production usage intended.
