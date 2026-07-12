# `internal/tui` — Terminal UI

This package implements the canonical Pulse experience: a real-time terminal
UI for exploring HTTP behavior. It builds on [Bubble Tea][bubbletea] and runs
in any xterm-compatible terminal.

## Purpose

The TUI presents a multi-surface workspace:

| Surface | Mode / Dialog | What you see |
|---|---|---|
| Ready | default | Launch pad: set URL, method, concurrency |
| Timeline | `mode=observe`, `view=timeline` | Live streaming results as a timeline |
| Logs | `mode=observe`, `view=logs` | Live streaming results as a log |
| Request | `dialog=request` | Payload editor (headers + body) |
| Inspect | `dialog=inspect` | Response details (status, headers, body, error) |
| Compare | `dialog=compare` | Investigation comparison with Verdict → Why → Evidence → Details hierarchy |

The active surface is determined by two state fields on `Model`:

- `workspace.mode` — are we observing or on the launch pad?
- `workspace.dialog` — is a dialog overlay open?

## File Map

### Entry point

| File | Role |
|---|---|
| `view.go` | `View()` — pure dispatch to the correct surface renderer. Composes top bar, body, status bar. |

### Model and update loop

| File | Role |
|---|---|
| `model.go` | `Model` struct, `Update()`, key handling dispatch, window resize handling. |
| `domain.go` | Active UI domain tracking (request, payload, URL). |
| `workspace.go` | Workspace state: mode, dialog, view selection, compare slots. |
| `update_compare.go` | Compare dialog update logic (key handlers, navigation). |
| `update_inspect.go` | Inspect dialog update logic (zone switching, scrolling). |

### Rendering

| File | Role |
|---|---|
| `render_ready.go` | Ready surface (launch pad). |
| `render_observe.go` | Timeline and Logs surfaces. |
| `render_request.go` | Request dialog (method, URL, concurrency fields). |
| `render_payload.go` | Payload editor (header rows + body textarea). |
| `render_context.go` | Context panel (active run configuration). |
| `render_inspect.go` | Inspect dialog dispatch. |
| `render_inspect_summary.go` | Inspect result summary (status, latency, URL, error). |
| `render_inspect_body.go` | Inspect response body view. |
| `render_inspect_why.go` | Inspect failure analysis pane. |
| `render_compare.go` | Compare dialog dispatch and diff rendering. |
| `render_common.go` | Shared rendering utilities (separators, hints). |

### Geometry and layout

| File | Role |
|---|---|
| `shell.go` | Shell layout computation (dimensions, workspace region). |
| `layout_constants.go` | Named layout sizes and thresholds. |
| `geometry.go` | Payload geometry calculation, `syncPayloadGeometry()`. |
| `region.go` | `Region` type and content region computation. |
| `surface.go` | Surface definitions (workspace regions, surface identity). |

### Styles

| File | Role |
|---|---|
| `styles.go` | Lipgloss style definitions (base, accent, borders, badges). |

### Validation

| File | Role |
|---|---|
| `validate.go` | Input validation helpers (URL, concurrency). |

### Execution

| File | Role |
|---|---|
| `run.go` | Run lifecycle: start, cancel, result ingestion. |

### Audit

| File | Role |
|---|---|
| `audit.go` | Architecture audit checks (render purity, field classification). |

## Update Pipeline

Every user interaction flows through a single `Update()` method:

```
tea.Msg → Model.Update()
  ├─ tea.WindowSizeMsg → resize shell, sync geometry
  ├─ tea.KeyMsg → dispatch to workspace key handler
   │    ├─ Enter → open Inspect (from the Observe result list)
   │    ├─ [ ] → switch Timeline / Logs view (Observe)
   │    ├─ Ctrl+R → run · Ctrl+X → cancel
   │    ├─ dialog keys (e configure, c compare, w export, q quit)
   │    ├─ Tab / Shift+Tab → switch field or zone (dialog / inspect / compare)
   │    └─ navigation keys (up/k, down/j, pgup, pgdown) → move within surfaces
  ├─ resultMsg → append result to results list
  ├─ tickMsg → update elapsed timer
  └─ runFinishedMsg → finalize run, compute summary
```

Key handlers delegate to dialog-specific update functions for complex
interactions (e.g., `updateCompareKey()`, `handleInspectKey()`).

## Render Pipeline

See [RENDERING.md](../../RENDERING.md) for the full rendering constitution.

Render functions are pure — they consume `Model` state and produce strings.
They never mutate widgets, set geometry, or perform I/O.

## Where to Start

New to the TUI codebase? Read in this order:

1. `model.go` — `Model` struct and `Update()` to understand state transitions.
2. `view.go` — `View()` to see how rendering is dispatched.
3. `workspace.go` — workspace state: modes, dialogs, surfaces.
4. `shell.go` — layout computation and region allocation.
5. One surface renderer (`render_ready.go` or `render_observe.go`) to see
   how a surface produces its output.

## See also

| Document | What it answers |
|---|---|
| `STATE_OWNERSHIP.md` | Model field ownership, lifetime, mutation rules |
| `COMPARE_CONSTITUTION.md` | Comparison engine contract: state model, analysis, render projections |
| `COMPARE_WORKFLOW.md` | Comparison UX: keybindings, lifecycle, persistence, preview behaviour |
| [RENDERING.md](../../RENDERING.md) | TUI rendering architecture, layout, render lifecycle, constitution |
| [ARCHITECTURE.md](../../ARCHITECTURE.md) | System architecture, components, APIs, engine, concurrency |
| [README.md](../../README.md) | Product overview, installation, quick start |

[bubbletea]: https://github.com/charmbracelet/bubbletea
