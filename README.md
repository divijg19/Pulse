# `Pulse`

> A real-time observability surface for HTTP behavior.

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![Frontend](https://img.shields.io/badge/Frontend-SolidJS%20%2B%20Vite-blue)
![Streaming](https://img.shields.io/badge/Streaming-SSE-brightgreen)
![License](https://img.shields.io/badge/License-MIT-blue)

Install latest release directly:

```bash
curl -fsSL https://raw.githubusercontent.com/divijg19/Pulse/main/install.sh | bash
```

Pulse is a live, streaming API explorer for backend and infrastructure workflows.
Instead of waiting for a batch to finish, you see request outcomes as they happen.

## Why Pulse? (The Philosophy)

Most API tools answer one question: What did the server return?
Pulse focuses on the operational question: What is the system doing right now?

Design principles:

- Streaming over batching.
- Clarity over feature sprawl.
- Fast feedback over post-run forensics.
- Simple, inspectable architecture over framework complexity.

## Features

- Concurrent HTTP execution with immediate SSE result streaming.
- Payload editor for request headers and raw body.
- Interactive request drawer for status, latency, response headers, response body, and errors.
- Live timeline and log views for per-request behavior.
- Single deployable Go binary with embedded frontend assets.

## Core Experience

1. Set URL, method, and concurrency.
2. Optionally configure headers and body.
3. Run and watch each request stream in real time.
4. Click any result to inspect full response details.

## Documentation Map

- Architecture and contracts: [ARCHITECTURE.md](ARCHITECTURE.md)
- Frontend-only development details: [web/README.md](web/README.md)

The root README is product-level. Deep implementation details are centralized in ARCHITECTURE.md.

## Quick Start

### Option A: Install latest binary

```bash
curl -fsSL https://raw.githubusercontent.com/divijg19/Pulse/main/install.sh | bash
pulse
```

### Option B: Build from source

Prerequisites:

- Go 1.25+
- Bun

```bash
git clone https://github.com/divijg19/Pulse.git
cd Pulse

cd web
bun install --frozen-lockfile
bun run build
cd ..

go build -o pulse ./cmd/server
./pulse
```

Default server address: http://localhost:8080

## Endpoints

- GET /health
- POST /run
- GET /stream

For full request/response schema, validation rules, and limits, see [ARCHITECTURE.md](ARCHITECTURE.md).

## CLI Smoke Test

Stream channel:

```bash
curl -N http://localhost:8080/stream
```

Trigger run:

```bash
curl -X POST http://localhost:8080/run \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://httpbin.org/anything",
    "method": "POST",
    "headers": {
      "Content-Type": "application/json"
    },
    "body": "{\"ping\":\"pong\"}",
    "concurrency": 5
  }'
```

## CI/CD

- CI builds frontend, runs backend tests, and verifies server build.
- Release runs on `v*` tags and manual dispatch.
- Artifacts are versioned dynamically, for example:
  - `pulse-v0.6.3-linux-amd64`
  - `pulse-v0.6.3-windows-amd64.exe`

## License

MIT.