# Architecture & Engineering Choices

This document describes the production architecture, data contracts, lifecycle behavior, and safety constraints for Pulse.

## System Overview

Pulse has two execution planes:

- Control plane: `POST /run` starts a concurrent run.
- Data plane: `GET /stream` pushes each result to subscribed clients over SSE.

High-level flow:

```text
Browser UI
    ├─ POST /run  ---> API validation ---> Engine fan-out (N goroutines)
    └─ GET /stream <--- Hub broadcast <--- Per-request Result production
```

## Components

### cmd/server

- Boots an `http.Server` with explicit timeouts.
- Serves embedded static assets.
- Registers `/health`, `/run`, `/stream`.

### internal/api

- `run.go`:
    - Accepts only `POST`.
    - Caps request body at 1 MiB.
    - Rejects unknown JSON fields.
    - Validates URL and concurrency bounds.
    - Dispatches run to engine with request-scoped context.
- `stream.go`:
    - Creates per-client buffered channel.
    - Streams `event: result` lines.
    - Exits cleanly on write failure or client disconnect.

### internal/engine

- `engine.go`:
    - Spawns `concurrency` goroutines for a run.
    - Aborts early when context is canceled.
    - Broadcasts each completed result to hub immediately.
- `http.go`:
    - Builds request via `http.NewRequestWithContext`.
    - Applies optional request headers and body.
    - Executes with bounded client timeout.
    - Reads response body with strict 10 KiB cap.
    - Flattens response headers to `map[string]string`.

### internal/stream

- Thread-safe hub using mutex-protected client map.
- Non-blocking broadcast so slow clients do not stall producers.
- Idempotent remove path prevents close-on-closed channel panics.

### web

- SolidJS UI consuming SSE stream.
- Payload editor for outbound headers/body.
- Result drawer for response headers/body/error inspection.
- EventSource is cleaned up on unmount.

## API Contract

### POST /run

Request body:

```json
{
    "url": "https://httpbin.org/anything",
    "method": "POST",
    "headers": {
        "Content-Type": "application/json"
    },
    "body": "{\"x\":1}",
    "concurrency": 10
}
```

Validation:

- `url` must be present and parse as request URI.
- `concurrency` must be between 1 and 100.
- body payload is capped to 1 MiB at decode layer.

Behavior:

- The handler responds with the accepted request payload for immediate client confirmation.
- Execution then proceeds asynchronously through concurrent workers and SSE result events.

### GET /stream

SSE event type: `result`

Event payload fields:

- `status` HTTP status code.
- `latency` Go duration encoded in nanoseconds.
- `timestamp` request completion timestamp.
- `error` error string when execution/read fails.
- `responseHeaders` flattened response header map.
- `responseBody` truncated response body capture (10 KiB max).

Streaming semantics:

- Event type is `result`.
- SSE writer flushes after each event.
- Stream loop exits on client disconnect, channel close, or write failure.

## Concurrency and Cancellation Model

Cancellation propagation path:

1. Client disconnects or request context is canceled.
2. API handler context closes.
3. Engine fan-out loop and workers observe `ctx.Done()`.
4. In-flight HTTP request is canceled by context-aware request.
5. Worker exits without extra broadcasts.

This avoids orphan worker goroutines and stale downstream writes.

## Resource-Safety Guarantees

### Memory

- `/run` JSON decode input bounded to 1 MiB.
- Per-response body read bounded to 10 KiB.
- SSE client channels are removed and closed on disconnect.

Result payload implications:

- Bounded response bodies cap memory growth at source.
- Flattened header maps avoid retaining large internal header structures.

### Goroutines

- Worker goroutines tied to request context lifecycle.
- SSE loop returns on write error, channel close, or context cancellation.

### Backpressure and Deadlock Avoidance

- Hub broadcast sends are non-blocking (`select` + `default`).
- Slow consumers drop events rather than stall producers.

## Operational Notes

- Health endpoint: `GET /health`.
- Binary serves embedded frontend assets; no separate runtime static server required.
- Release artifacts are versioned by tag and target tuple.

Release naming contract:

- `pulse-${VERSION}-linux-amd64`
- `pulse-${VERSION}-windows-amd64.exe`
- `pulse-${VERSION}-mac-amd64`
- `pulse-${VERSION}-mac-arm64`

## Repository Structure

```text
Pulse/
├── cmd/server/            # server bootstrap + embedded static hosting
├── internal/api/          # request validation + SSE HTTP handlers
├── internal/engine/       # concurrent HTTP execution
├── internal/stream/       # pub/sub hub for SSE fan-out
├── internal/model/        # DTO contracts
├── web/                   # SolidJS frontend source
└── .github/workflows/     # CI and release pipelines
```