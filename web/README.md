# Web Frontend (SolidJS)

This directory contains the Pulse frontend application built with SolidJS + Vite + Tailwind v4.

The backend serves compiled assets from the embedded static output, so this app is not deployed separately in production.

For backend endpoint contracts and execution semantics, see [../ARCHITECTURE.md](../ARCHITECTURE.md).

## Tooling

- Runtime framework: SolidJS
- Bundler/dev server: Vite
- Styling: Tailwind CSS v4
- Formatter/lint: Biome

## Scripts

Run from `web/`:

```bash
bun install --frozen-lockfile
```

Available commands:

- `bun run dev`: start Vite dev server.
- `bun run build`: type-check and build production bundle.
- `bun run preview`: preview built bundle locally.
- `bun run lint`: run Biome checks with write mode.
- `bun run format`: run Biome formatting.

## Frontend Behavior Contract

### Control flow

- Sends run configuration to `POST /run`.
- Subscribes to `GET /stream` with EventSource.
- Appends each `result` event as it arrives.

### Payload editor

- Allows optional request headers (key/value list).
- Allows optional request body (raw string).
- Serializes headers into an object before POST.

### Request drawer

- Opens from timeline/log row selection.
- Displays status, latency, request method/url, response headers, response body, and error.
- Drawer visibility is decoupled from selected payload to preserve close animation quality.

The expected result event schema is defined in [../ARCHITECTURE.md](../ARCHITECTURE.md).

## Build and Embed Workflow

Production build flow:

1. Build frontend (`bun run build`) in `web/`.
2. Copy or output build assets into server embedded static location.
3. Build Go server binary.

CI and release workflows execute this order automatically to keep embed assets valid.

## Styling Guidelines

- Keep the existing Vantablack + neon visual language.
- Use Tailwind utility classes only (no external UI kits).
- Avoid heavy dependencies for editors/charts.
- Prefer composable primitives (`createSignal`, `createEffect`, `For`, `Show`).
