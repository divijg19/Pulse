# `Pulse`

> A real-time terminal-first observability surface for HTTP behavior.

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)
![Web UI](https://img.shields.io/badge/Web-SolidJS%20%2B%20Vite-blue)
![Terminal UI](https://img.shields.io/badge/TUI-Bubble%20Tea-7C3AED)
![Streaming](https://img.shields.io/badge/Streaming-SSE-brightgreen)
![License](https://img.shields.io/badge/License-MIT-blue)

Install latest release directly:

```bash
curl -fsSL https://raw.githubusercontent.com/divijg19/Pulse/main/install.sh | bash
```

Pulse is a live, streaming API explorer for backend and infrastructure workflows.
Instead of waiting for a batch to finish, you see request outcomes as they happen.
The canonical experience is the native terminal UI; `pulse web` starts the browser WebUI when you want it.

## Why Pulse? (The Philosophy)

Most API tools answer one question: What did the server return?
Pulse focuses on the operational question: What is the system doing right now?

Design principles:

- Streaming over batching.
- Clarity over feature sprawl.
- Fast feedback over post-run forensics.
- Simple, inspectable architecture over framework complexity.

## Features

- Canonical terminal UI with Ready launch pad, payload editor, live metrics, timeline/log views, response inspector, and investigation comparison.
- Optional browser WebUI started with `pulse web`.
- Concurrent HTTP execution with immediate result streaming.
- Payload editor for request headers and raw body.
- Single deployable Go binary with embedded frontend assets.

## Core Experience

1. Set URL, method, and concurrency.
2. Optionally configure headers and body.
3. Run and watch each request stream in real time.
4. Select any result and press Enter to inspect full response details.
5. Press `c` to mark a result, select another, and press `c` again to compare investigations (verdict, why, evidence, details).

## Documentation Map

| Document | What it answers |
|---|---|
| [ARCHITECTURE.md](ARCHITECTURE.md) | System architecture, components, APIs, engine, concurrency, resource safety |
| [RENDERING.md](RENDERING.md) | TUI rendering architecture, layout, render lifecycle, rendering constitution |
| [internal/tui/README.md](internal/tui/README.md) | TUI package guide, file layout, navigation |
| [internal/tui/STATE_OWNERSHIP.md](internal/tui/STATE_OWNERSHIP.md) | Model field ownership, lifetime, mutation rules |
| [web/README.md](web/README.md) | Frontend-only development details |

Every document answers exactly one question. No concept is explained
in depth in more than one place. Read the doc whose question matches yours.

The root README is product-level. Deep implementation details are centralized
in ARCHITECTURE.md.

## Quick Start

### Option A: Install latest binary

```bash
curl -fsSL https://raw.githubusercontent.com/divijg19/Pulse/main/install.sh | bash
pulse
```

`pulse` opens the terminal UI. To start the browser WebUI instead:

```bash
pulse web
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

go build -o pulse ./cmd/pulse
./pulse
```

Default WebUI address: http://localhost:8080

```bash
./pulse web
```

## Endpoints

- GET /health
- POST /run
- GET /stream

For full request/response schema, validation rules, and limits, see [ARCHITECTURE.md](ARCHITECTURE.md).

## WebUI Smoke Test

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
