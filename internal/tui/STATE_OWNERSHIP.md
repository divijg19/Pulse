# State Ownership

Every field on `Model` is classified by lifecycle and ownership.
Render code may read any field; it may never write to any field.

## Classifications

| Classification  | Persists across frames? | Survives navigation? | Example                    |
|-----------------|------------------------|----------------------|----------------------------|
| Persistent      | Yes                    | Yes                  | `urlInput`                 |
| Derived         | Recalculated           | No (re-derives)      | `payloadGeometry`          |
| Ephemeral       | No (message-driven)    | N/A                  | `summary`, `results`       |
| Transient UI    | Frame-only             | No                   | `selectHead`, `inspectZone`|

## Field Registry

| Field              | Class        | Owner            | Updated By                                        | Read By                                       |
|--------------------|--------------|------------------|---------------------------------------------------|-----------------------------------------------|
| `shell`            | Persistent   | `model.go`       | `tea.WindowSizeMsg`                               | `View()`, `workspaceContentWidth()`           |
| `workspace`        | Persistent   | `workspace.go`   | Key handlers (`mode`, `dialog`, `view`, `compare`)| `View()`, `orientationLabel()`                |
| `payloadGeometry`  | Derived      | `geometry.go`    | `syncPayloadGeometry()` → `Update()`              | `renderPayloadDomain()`                       |
| `activeDomain`     | Transient UI | `model.go`       | Key handlers (Tab, Shift+Tab)                     | `renderRequestDomain()`, `renderPayloadDomain()` |
| `methodIndex`      | Persistent   | `model.go`       | Key handlers (up/down in method field)            | `renderRequestDomain()`                       |
| `requestField`     | Transient UI | `model.go`       | Key handlers (Tab, up/down)                       | `renderRequestDomain()`                       |
| `urlInput`         | Persistent   | `model.go`       | `textinput` update model (key events)              | `renderRequestDomain()`, `Update()`           |
| `concurrencyInput` | Persistent   | `model.go`       | `textinput` update model (key events)              | `renderRequestDomain()`                       |
| `bodyInput`        | Persistent   | `model.go`       | `textarea` update model (key events)               | `renderPayloadDomain()`, `syncPayloadGeometry()` |
| `headers`          | Persistent   | `model.go`       | Key handlers (add/remove/edit)                     | `renderPayloadDomain()`, `syncPayloadGeometry()` |
| `selectedHead`     | Transient UI | `model.go`       | Key handlers (up/down)                             | `renderPayloadDomain()`, `fieldErrors()`       |
| `headerSubfocus`   | Transient UI | `model.go`       | Key handlers (Tab in header)                       | `renderPayloadDomain()`                        |
| `results`          | Ephemeral    | `model.go`       | `resultMsg`, `runFinishedMsg`, key Clear            | `renderInspect*()`, `renderCompare*()`, `View()` |
| `selected`         | Transient UI | `model.go`       | Key handlers (up/down in result list)               | `renderInspectSummary()`, `renderCompareSummary()` |
| `running`          | Ephemeral    | `model.go`       | `runFinishedMsg`, Enter key                         | `View()`, `orientationLabel()`                 |
| `startedAt`        | Ephemeral    | `model.go`       | Enter key, `runFinishedMsg`                         | `View()`                                       |
| `elapsed`          | Derived      | `model.go`       | `tickMsg`                                           | `View()`                                       |
| `cancel`           | Ephemeral    | `model.go`       | Enter key, Esc key                                  | `Update()`                                     |
| `eventCh`          | Ephemeral    | `model.go`       | Enter key (run start)                               | `Update()`                                     |
| `status`           | Ephemeral    | `model.go`       | `tickMsg`, `eventErrorMsg`, key handlers             | `View()`                                       |
| `errMsg`           | Ephemeral    | `model.go`       | `eventErrorMsg`                                     | `View()`                                       |
| `summary`          | Derived      | `model.go`       | `metrics.Accumulate()` → `runFinishedMsg`            | `View()`, `renderInspectSummary()`             |
| `inspectZone`      | Transient UI | `model.go`       | Key handlers (left/right, tab)                       | `renderInspectBody()`, `renderInspectWhy()`    |
| `inspectBodyOffset`| Transient UI | `model.go`       | Key handlers (up/down in inspect body scroll)        | `renderInspectBody()`                          |

## Invariants

1. **Render never writes**: No `render_*.go` file may write to any Model field.
2. **Geometry single authority**: `payloadGeometry` is set only by `syncPayloadGeometry()` in `Update()`.
3. **Ephemeral fields are reset**: `results`, `running`, `cancel`, `eventCh`, `errMsg` are explicitly zeroed on transition boundaries (dialog open, mode change, clear).
4. **Transient UI fields are local**: Their values have no meaning outside the current View frame and must be re-initialized on dialog/mode transitions.

---

## See also

| Document | What it answers |
|---|---|
| [../../README.md](../../README.md) | Product overview, installation, quick start |
| [../../ARCHITECTURE.md](../../ARCHITECTURE.md) | System architecture, components, APIs, engine, concurrency |
| [../../RENDERING.md](../../RENDERING.md) | TUI rendering architecture, layout, render lifecycle, constitution |
| [../README.md](../README.md) | TUI package guide, file layout, navigation |
