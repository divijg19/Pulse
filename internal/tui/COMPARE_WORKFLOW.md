# Compare Workflow

↩ [internal/tui/README.md](README.md) · [COMPARE_CONSTITUTION.md](COMPARE_CONSTITUTION.md)

## Purpose

This document defines the Compare workflow lifecycle, behavioural invariants, and canonical vocabulary. It is the operator-facing companion to `COMPARE_CONSTITUTION.md` (which covers architecture and rendering).

---

## Lifecycle

```
Idle
  │
  │  c on result
  ▼
BaselineMarked
  │
  │  c on different result
  ├──────────────────────────────► Comparing
  │                                    │
  │                           ┌────────┼────────┐
  │                           │        │        │
  │                    c on baseline  s       Esc
  │                           │        │        │
  │                           ▼        ▼        ▼
  │                        Idle     Comparing  Observe
  │                           (swap)    (session preserved)
  │
  │  c on same result
  └──────────────────────────────► Idle
                                          │
                                   x or c on baseline
                                          │
                                          ▼
                                       Idle
```

### States

| State | Meaning |
|-------|---------|
| Idle | No comparison workflow active. Reference may still exist. |
| BaselineMarked | A baseline result is selected, but no candidate has been chosen. |
| Comparing | Both baseline and candidate are set. Comparison analysis is computed and displayed. |

### Transitions

All transitions are owned by `CompareWorkspace` operations. Baseline and
candidate are resolved `*model.Result` pointers — there are no indices.

| From | Trigger | To | Effect |
|------|---------|----|--------|
| Idle | `c` on result (no Reference) | BaselineMarked | `MarkBaseline(r)`; resolved baseline pointer set. |
| Idle | `c` on result (Reference set) | Comparing | Baseline = Reference, then `SelectCandidate(r)`. |
| BaselineMarked | `c` on same result | Idle | `Unmark()`; Reference preserved. |
| BaselineMarked | `c` on different result | Comparing | `SelectCandidate(r)`; analysis recomputed once. |
| Comparing | `c` on baseline result | Idle | `Clear()`; Reference preserved. |
| Comparing | `c` on candidate result | Comparing | Resume workspace (`Enter Compare`); comparison unchanged. |
| Comparing | `c` on different result | Comparing | `ReplaceCandidate(r)`; analysis recomputed once. |
| Comparing | `s` | Comparing | `Swap()`; baseline/candidate exchanged, analysis recomputed. |
| Comparing | `x` | Idle | `Clear()`; Reference preserved, returns to Observe. |
| Any | `Esc` | Observe | Esc handled inline (no workspace mutation); session preserved for resume. |
| Any | `q` | Quit dialog | Application quit on confirm. |

Outside Compare mode (Observe / Inspect), `x` is context-sensitive: it clears
the active comparison (Baseline + Candidate + State + Analysis) while preserving
the reference; when **only** a reference remains, `x` renounces the
reference entirely (`RenounceReference()`), ending its persistence across runs.

---

## Views

Once Comparing, the workspace is presented through discrete views. Every view
consumes the identical immutable `CompareContext`; switching views only changes
presentation, never the data.

Each participant's identity header carries its **request number** (`#001`,
1-based arrival order) and **timestamp**, so the operator always knows *what*
and *when* without leaving the workspace. The Raw view additionally prints a
`Request:` line with the same number and time for both baseline and candidate.

| View | Content |
|------|---------|
| Overview | Verdict + Why (direction-only sentences). Default. |
| Evidence | Categorical deltas: Status, Latency, Headers, Body, Errors. |
| Diff | URL change and full body diff. |
| Headers | Added / removed / changed response headers. |
| Body | Full baseline and candidate response bodies — scrollable. |
| Raw | Full baseline and candidate result dumps — scrollable. |

Navigation:

* `[` / `]` — previous / next view (wraps around).
* `Tab` / `Shift+Tab` — same as bracket navigation.
* `up` / `k` / `down` / `j` — scroll the active view (Body and Raw scroll through
  their full content; scroll position persists across view switches).
* `Esc` — Exit (preserves the view for resume).

The current view name is shown in the status line (`Comparing · <View>`).

---

## Collapsed Preview (the persistent workspace)

The collapsed preview is not a secondary surface — it **is** the Compare
workspace rendered with partial context. It appears in the Observe/Inspect
context panel the moment a baseline is marked, and persists until cleared.

It always shows:

* baseline identity (◆) and, when present, candidate identity (▶), each with its
  request number and timestamp; a pinned-only state shows `● Pinned Baseline`,
* the verdict when a comparison is active,
* a context-specific keybinding line so the next action is obvious without
  leaving the drawer: `c on ▶ open · x clear · s swap · [ ] view` (comparing),
  `c compare · x clear` (baseline marked), or `c compare · x renounce`
  (reference only).

The preview is the right-side drawer: it is deliberately richer than a preview
so the operator can act on the comparison (open, clear, swap, renounce) directly
from Observe/Inspect without entering the full workspace.

Even when the context panel is hidden (terminal narrower than 140 columns), the
status line reports the active comparison (`Comparing · c on ▶ to open` /
`Baseline marked · c to compare` / `Reference · x renounces`), so the
operator never loses orientation.

Because the preview shares `renderComparisonIdentityBlock` with the full
workspace, the two never drift.

---

## Behavioural Invariants

1. **Compare never destroys a baseline implicitly.** Clear (`x`) resets the session but leaves Reference intact. Only a new `startRun` may update Reference.

2. **Swap never changes comparison content.** Swap exchanges the resolved baseline and candidate pointers. The same two results are compared — only the perspective changes, so the directional verdict reflects the new orientation.

3. **Exit never clears comparison.** `Esc` from Compare mode returns to Observe while preserving Session state. The operator can resume viewing the comparison at any time.

4. **Clear only clears the comparison session.** `x` resets Baseline, Candidate, State, and Analysis to their zero values via `Clear()`. Reference is never modified by Clear.

5. **Rendering never mutates workflow state.** All render functions accept the model as a value receiver and produce output without side effects. Workflow state changes occur only in `handle*Key` methods.

6. **Mark is idempotent.** Pressing `c` on an already-marked baseline clears the session. Pressing `c` again on the same or different result starts a new workflow naturally.

7. **The operator never needs to remember workflow state.** Timeline markers (▶, ◆, ●) indicate the current comparison state at a glance. The Compare identity header establishes orientation. The ribbon shows available actions.

---

## Vocabulary

| Concept | Canonical Term | Used In |
|---------|---------------|---------|
| Select a reference result | Mark Baseline | Ribbon, status, code comments |
| Persistent reference that survives runs | Pinned Baseline | Code, docs, state model |
| Result being compared against baseline | Candidate | Ribbon, renderer, status |
| Workflow name | Compare | Orientation, mode, docs |
| Exchange baseline and candidate | Swap | Ribbon, action |
| End the active comparison | Clear Comparison | Ribbon, status |
| Leave Compare without clearing | Exit Compare | Status, docs |
| Start comparison with Reference | Compare (auto) | Status, implicit |

### Prohibited synonyms

| Avoid | Use Instead |
|-------|-------------|
| Mark / Reference interchangeably | Mark Baseline / Reference |
| Active result | Candidate |
| Clear All / Reset | Clear Comparison |
| Flip / Reverse | Swap |
| Go back / Leave | Exit Compare |

---

## Behaviour-first Tests

Workflow tests in `architecture_test.go` verify operator-facing behaviour:

| Test | Story |
|------|-------|
| `TestV0102Workflow_MarkReplace` | Mark A, c on B → B is candidate |
| `TestV0102Workflow_ClearPreservesReference` | Compare, x → session cleared, reference intact |
| `TestV0102Workflow_SwapTwiceReturns` | s then s → original orientation restored |
| `TestV0102Workflow_ExitPreservesSession` | Esc → Observe → session preserved |
| `TestV0102Workflow_ClearComparison` | Compare, x → Idle, Observe |
| `TestV0102Workflow_CrossRunReference` | Mark → startRun → new results → compare with reference |
| `TestV0102Workflow_MarkSameUnmarks` | c on already-marked → clears |
| `TestV0102Workflow_NoBaselineCleared` | Reference only → UI shows ● Pinned Baseline |

---

## Related Documents

- `COMPARE_CONSTITUTION.md` — Architecture, engine, rendering hierarchy
- `RENDERING.md` — Visual presentation rules
- `STATE_OWNERSHIP.md` — Data ownership and lifecycle
