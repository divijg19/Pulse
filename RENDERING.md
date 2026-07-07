# Rendering Constitution

## Lifecycle

The rendering lifecycle is strictly unidirectional and proceeds in exactly three
phases:

```
Layout Event (WindowSizeMsg, region change, dialog transition)
        │
        ▼
  ┌─────────────────┐
  │  Update()       │  Layout geometry, synchronize widget dimensions,
  │                 │  handle input, update model state.
  │  syncGeometry() │  Never render here.
  └────────┬────────┘
           │ state
           ▼
  ┌─────────────────┐
  │  View()         │  Functional rendering: consume model, produce string.
  │                 │  Never mutate model or widgets here.
  │  render_*.go    │
  └────────┬────────┘
           │ string
           ▼
  ┌─────────────────┐
  │  tea.Program    │  Bubble Tea renders the string to terminal.
  └─────────────────┘
```

Rendering is a **pure transformation** from immutable model state to string.

---

## Contract

### Render functions MAY:

- Read `m.(Model)` fields.
- Read widget state (`.Value()`, `.Focused()`, `.View()`).
- Format, style, compose strings.
- Build and return strings via `strings.Builder` or `lipgloss`.

### Render functions MAY NEVER:

- **Mutate widget geometry**: `.SetWidth()`, `.SetHeight()`, `.Width =`, `.Height =`.
- **Mutate widget focus**: `.Focus()`, `.Blur()`.
- **Mutate widget values**: `.SetValue()`, `.SetSuggestions()`.
- **Mutate model state**: `m.field = ...`, `m.someState = ...`.
- **Allocate persistent state**: channels, goroutines, file handles.
- **Perform business logic**: validation, computation, I/O.
- **Synchronize layout**: call `syncLayout()`, `syncPayloadGeometry()`, or any function
  that mutates widget dimensions.
- **Read from the environment**: files, network, time (except `time.Now()` for
  elapsed display — this is a read, not a mutation).

---

## Widget State Ownership

Widget state ownership is documented in
[internal/tui/STATE_OWNERSHIP.md](internal/tui/STATE_OWNERSHIP.md).
That file is the single canonical source for Model field classification,
lifetime, and mutation rules.

| Owner | Responsibilities |
|-------|-----------------|
| `Model.Update()` | Process input, synchronize layout, mutate widget state. |
| `syncPayloadGeometry()` | Single authority for payload widget widths. |
| `blurAll()` / `focusPayload*()` | Single authority for widget focus. |
| `handlePayloadHeaderKey()` | Single authority for header field values. |
| `renderPayloadDomain()` | Pure consumer of widget state — reads only. |
| `renderRibbon()` | Pure consumer of layout state — reads only. |

---

## Invalidation Events

The following events invalidate payload geometry and trigger synchronization:

| Event | Handler | Sync Function |
|-------|---------|---------------|
| `tea.WindowSizeMsg` | `Model.Update()` | `syncPayloadGeometry(payloadContentWidth(msg.Width, msg.Height))` |
| `startupMsg` | `Model.Update()` | `syncPayloadGeometry(payloadContentWidth(80, 24))` |
| Dialog open (`e`) | `Model.handleObserveKey()` | `syncPayloadGeometry(payloadContentWidth(w, 24))` |
| Context panel toggle | `Model.Update()` | `syncPayloadGeometry(payloadContentWidth(w, h))` |

When new invalidation events are added, the synchronization function must be
called **before** rendering, never during.

---

## Layer Model

The terminal UI follows a strict four-layer ownership model:

| Layer | Role | Renderers |
|---|---|---|
| Context | Persistent state frame | Top bar (method, URL, CC) |
| Identity | Workspace identification | Timeline, Logs, Inspector - Result #N, Endpoint, Concurrency, Payload |
| Content | Primary data display | Metrics, result rows, dialog forms, response details |
| Interaction | Immediate action signals | Status bar mode + hints, ConfirmQuit |

ConfirmQuit is an interaction-layer dialog only — it preserves the current
workspace identity and content, changing only the status bar.

---

## Identity Badges

Each workspace surface owns exactly one identity line:

| Surface | Identity | Style |
|---|---|---|
| Ready | `READY` | Badge |
| Timeline | `OBSERVE` (Timeline) | Bold + Accent |
| Logs | `OBSERVE` (Logs) | Bold + Accent |
| Inspect | `INSPECT` | Badge |
| Compare | `COMPARE` | Badge |
| Request | `REQUEST` | Badge |
| Quit | `QUIT` | Badge |

---

## Visual Invariants

| Concept | Representation |
|---|---|
| Cursor (navigation position) | `▶` glyph |
| Highlight (active target) | Accent foreground + dark background |

All selection-capable surfaces follow this invariant:

| Surface | Cursor | Highlight |
|---|---|---|
| Timeline | ✓ | ✓ |
| Logs | ✓ | ✓ |
| Payload header rows | ✓ | ✓ |
| Payload body | N/A | ✓ |
| Endpoint method selector | ✓ | ✓ |
| Concurrency value | ✓ | ✓ |
| Inspect | N/A | N/A (read-only) |
| Compare | N/A | N/A (read-only) |

---

## Layout and Regions

The terminal display is divided into a vertical stack:

```
Top bar (method, URL, concurrency)
Separator
Body (surface-specific content)
Separator
Status bar (mode hints, messages)
```

The shell layout (`m.shell.Layout()`) computes the workspace content region
from shell dimensions. Each surface renderer receives a `Region` with its
allocated width and height and produces a string constrained to those bounds.

The body area switches between surfaces based on dialog state and mode:

| State | Surface |
|---|---|
| No dialog, mode = ready | Ready |
| No dialog, mode = observe | Timeline or Logs |
| Dialog = request | Request (payload editor) |
| Dialog = inspect | Inspect |
| Dialog = compare | Compare |
| Dialog = quit | Current surface + quit prompt |

The context panel (visible at `shellWidth >= 140`) sits alongside the primary
content region, displaying the active run configuration. It reduces the primary
region width by approximately one third.

### Renderer Dispatch

`View()` is a pure dispatch function:

1. Extract layout from `m.shell.Layout()`, compute workspace content region.
2. Dispatch to the correct surface renderer based on dialog, mode, and state.
3. Compose the layout: Top bar → Separator → Body → Separator → Status bar.
4. Wrap in base style with explicit width and height.

No renderer performs another renderer's work. No rendering logic lives in `View()`.

---

## CI Enforcement

Render purity is enforced by the certification gate:

```bash
# Architectural audit — render_*.go must contain zero mutations
! grep -qn '\.SetWidth\|\.SetHeight\|\.Width =\|\.Height =\|\.Focus()\|\.Blur()\|\.SetValue' internal/tui/render_*.go
```

This check runs as part of every architectural certification milestone.

## Manual Audit

To verify render purity locally:

```bash
grep -n '\.SetWidth\|\.SetHeight\|\.Width =\|\.Height =\|\.Focus()\|\.Blur()\|\.SetValue' internal/tui/render_*.go
```

Should return **zero matches**. If a match appears, either:

1. Move the mutation to `Model.Update()` or an explicit sync function.
2. Add a justified exception with a comment explaining why it cannot move.

There are currently **zero** exceptions.

---

## Rationale

Render purity eliminates an entire class of bugs:

- **Geometry drift**: Widget widths set during render cause `View()` output
  to change unpredictably across render cycles.
- **Focus flicker**: `Focus()`/`Blur()` during render causes cursor state to
  change outside the input lifecycle.
- **Value corruption**: `SetValue()` during render overwrites user input.
- **Layout inconsistency**: Mutating layout state during render means the
  layout used for rendering differs from the layout the model thinks is active.

By enforcing render purity, Pulse's layout becomes deterministic, testable,
and auditable.

---

## See also

| Document | What it answers |
|---|---|
| [README.md](README.md) | Product overview, installation, quick start |
| [ARCHITECTURE.md](ARCHITECTURE.md) | System architecture, components, APIs, engine, concurrency |
| [internal/tui/README.md](internal/tui/README.md) | TUI package guide, file layout, navigation |
| [internal/tui/STATE_OWNERSHIP.md](internal/tui/STATE_OWNERSHIP.md) | Model field ownership, lifetime, mutation rules |
