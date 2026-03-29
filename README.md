# `Pulse`

> **A real-time observability surface for HTTP behavior.**

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![Zero Dependencies](https://img.shields.io/badge/Dependencies-Zero-brightgreen)
![License](https://img.shields.io/badge/License-MIT-blue)

**`Pulse`** is a live, streaming API explorer built for backend engineers and infrastructure enthusiasts. It bridges the gap between traditional API clients (like Postman) and heavy load-testing CLI tools (like `k6`). 

Instead of firing requests and waiting for a batched response, `Pulse` utilizes a concurrent worker pool and **Server-Sent Events (SSE)** to stream metrics, logs, and latency visualizations to a reactive UI the exact millisecond they happen.

*(Insert a 5-second high-framerate GIF of the UI streaming latency bars here)*

---

## 🧠 Why `Pulse`? (The Philosophy)
Most API tools answer the question: *"What did the server return?"*  
`Pulse` answers: ***"What is the system doing right now?"***

It is designed with strict principles:
* **Streaming > Batching:** You shouldn't wait for 100 requests to finish to see the first result.
* **Clarity > Completeness:** No accounts, no saved workspaces, no clutter. Just pure, real-time HTTP I/O.
* **Fast > Feature-Rich:** The UI relies on native DOM updates and CSS Flexbox for visualization—no heavy charting libraries.

### 🧹 Design Principles
Speed > features
Clarity > completeness
Streaming > polling
Simple > clever


---

## 🚀 Features
* **Concurrent Execution:** Fire `N` requests simultaneously via Go worker pools.
* **Live Latency Visualization:** Relative latency bars render in real-time, making performance outliers instantly obvious.
* **Terminal-Style Streaming Logs:** Watch success/failures auto-scroll as they resolve.
* **Breathing Metrics:** RPS (Requests per second), Success Rate, and Average Latency update dynamically as the batch progresses.
* **Built-in Safety:** Strict timeouts, context cancellation, and response-body truncation prevent memory leaks and OOM crashes.

## ⚡ Core Experience

1. Enter API endpoint
2. Set concurrency (e.g., 10)
3. Click Send

→ Requests fire in parallel
→ Responses stream in live
→ Logs update instantly
→ Metrics evolve in real time

---

## 🏗️ Architecture & Engineering Choices

[ Browser UI ]
       │
       │ HTTP (POST) + SSE (stream)
       ▼
[ Go API Server ]
       │
       ├── Request Engine      (concurrent execution)
       ├── Stream Hub          (event broadcasting)
       └── Metrics Aggregator  (live stats)

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

---

## ⚙️ Getting Started

### Prerequisites
* [Go 1.21+](https://go.dev/dl/) installed.

### 🚀 Deployment
Backend: single Go binary
Platforms: Render / Fly.io / Railway
Frontend: static (served or CDN)

### Test via CLI (cURL)
Because `Pulse` uses standard SSE, you can interact with it directly from your terminal:

*Start a stream session:*
```bash
curl -N http://localhost:8080/stream
```

*Trigger a batch (in another terminal):*
```bash
curl -X POST http://localhost:8080/run \
  -H "Content-Type: application/json" \
  -d '{"url": "https://jsonplaceholder.typicode.com/posts", "method": "GET", "concurrency": 20}'
```

---

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

---

## 📜 License
MIT License. Do whatever you want with it.