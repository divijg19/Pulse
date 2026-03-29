## Architecture & Engineering Choices

`Pulse` is built to handle highly concurrent networking tasks without melting the server or the browser.

### 1. Concurrency & Memory Safety
* **Goroutine Fan-Out:** Requests are executed concurrently but bound by a strict context lifecycle. If a user drops the connection or hits "Stop", `context.Cancel` is propagated, instantly killing hanging TCP dials.
* **OOM Prevention:** Response bodies are read using `io.LimitReader` (capped at 10KB). If a user accidentally points `Pulse` at a 50MB file with a concurrency of 100, the server stays safe.

### 2. SSE over WebSockets
`Pulse` relies on a strictly uni-directional data flow (Server → Client). By choosing **Server-Sent Events (SSE)** over WebSockets, the architecture remains stateless, cache-proxy friendly, and automatically handles reconnects. It also means you can test the stream via a simple `curl` command.

### 3. Non-Blocking Event Hub
The Stream Hub routes telemetry from the worker pool to connected clients. To prevent slow clients (e.g., a browser tab on a bad connection) from backing up the entire system, the broadcast channel uses a non-blocking `select`:
```go
select {
case client <- event:
default:
    // Drop event to protect system stability
}
```

### 4. DOM-Friendly Frontend
To prevent browser thrashing when receiving 100+ events per second, the frontend leverages a rolling buffer for logs and uses lightweight CSS (Flexbox percentages) instead of SVG/Canvas to render the kinetic latency bars at 60fps.

                ┌────────────────────┐
                │     Browser UI     │
                └────────┬───────────┘
                         │
         HTTP (control)  │   SSE (data stream)
                         │
                         ▼
                ┌────────────────────┐
                │     API Layer      │
                └────────┬───────────┘
                         │
         ┌───────────────┼────────────────┐
         ▼               ▼                ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  Run Manager │  │  Stream Hub  │  │   Metrics     │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                 │
       ▼                 │                 │
┌──────────────┐         │                 │
│ Request      │─────────┴─────────────────┘
│ Engine       │   (events)
└──────────────┘


## 🗂️ Project Structure

```text
Pulse/
├── cmd/
│   └── server/
│       └── main.go       # Entry point
├── internal/
│   ├── api/              # HTTP handlers & validation
│   ├── engine/           # Worker pool & HTTP execution
│   ├── stream/           # SSE Hub & pub/sub logic
│   ├── metrics/          # Live RPS & stats aggregation
│   └── model/            # Shared DTOs and types
├── static/               # Zero-dependency HTML/JS/CSS
├── pkg/                  # Reusable packages (if needed)
└── go.mod
```