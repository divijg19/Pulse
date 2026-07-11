# Compare Workflow

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
| Idle | No comparison workflow active. PinnedBaseline may still exist. |
| BaselineMarked | A baseline result is selected, but no candidate has been chosen. |
| Comparing | Both baseline and candidate are set. Comparison analysis is computed and displayed. |

### Transitions

All transitions are owned by `CompareWorkspace` operations. Baseline and
candidate are resolved `*model.Result` pointers — there are no indices.

| From | Trigger | To | Effect |
|------|---------|----|--------|
| Idle | `c` on result (no PinnedBaseline) | BaselineMarked | `MarkBaseline(r)`; resolved baseline pointer set. |
| Idle | `c` on result (PinnedBaseline set) | Comparing | Baseline = PinnedBaseline, then `SelectCandidate(r)`. |
| BaselineMarked | `c` on same result | Idle | `Unmark()`; PinnedBaseline preserved. |
| BaselineMarked | `c` on different result | Comparing | `SelectCandidate(r)`; analysis recomputed once. |
| Comparing | `c` on baseline result | Idle | `Clear()`; PinnedBaseline preserved. |
| Comparing | `c` on candidate result | Comparing | Resume workspace (`Enter Compare`); comparison unchanged. |
| Comparing | `c` on different result | Comparing | `ReplaceCandidate(r)`; analysis recomputed once. |
| Comparing | `s` | Comparing | `Swap()`; baseline/candidate exchanged, analysis recomputed. |
| Comparing | `x` | Idle | `Clear()`; PinnedBaseline preserved, returns to Observe. |
| Any | `Esc` | Observe | `Exit()` — no mutation; session preserved for resume. |
| Any | `q` | Quit dialog | Application quit on confirm. |

---

## Views

Once Comparing, the workspace is presented through discrete views. Every view
consumes the identical immutable `CompareContext`; switching views only changes
presentation, never the data.

| View | Content |
|------|---------|
| Overview | Verdict + Why (direction-only sentences). Default. |
| Evidence | Categorical deltas: Status, Latency, Headers, Body, Errors. |
| Diff | URL change and full body diff. |
| Headers | Added / removed / changed response headers. |
| Body | Full baseline and candidate response bodies. |
| Raw | Full baseline and candidate result dumps. |

Navigation:

* `[` / `]` — previous / next view (wraps around).
* `Tab` / `Shift+Tab` — same as bracket navigation.
* `up` / `k` / `down` / `j` — scroll the active view.
* `Esc` — Exit (preserves the view for resume).

The current view name is shown in the status line (`Comparing · <View>`).

---

## Collapsed Preview (the persistent workspace)

The collapsed preview is not a secondary surface — it **is** the Compare
workspace rendered with partial context. It appears in the Observe/Inspect
context panel the moment a baseline is marked, and persists until cleared.

It always shows:

* baseline identity (◆) and, when present, candidate identity (▶),
* the verdict when a comparison is active,
* an orientation hint: `c on ▶ to open · x clears` (comparing) or
  `c on a result to compare` (baseline only).

Even when the context panel is hidden (terminal narrower than 140 columns), the
status line reports the active comparison (`Comparing · c on ▶ to open` /
`Baseline marked · c to compare`), so the operator never loses orientation.

Because the preview shares `renderComparisonIdentityBlock` with the full
workspace, the two never drift.

---

## Behavioural Invariants

1. **Compare never destroys a baseline implicitly.** Clear (`x`) resets the session but leaves PinnedBaseline intact. Only a new `startRun` may update PinnedBaseline.

2. **Swap never changes comparison content.** Swap exchanges the resolved baseline and candidate pointers. The same two results are compared — only the perspective changes, so the directional verdict reflects the new orientation.

3. **Exit never clears comparison.** `Esc` from Compare mode returns to Observe while preserving Session state. The operator can resume viewing the comparison at any time.

4. **Clear only clears the comparison session.** `x` resets Baseline, Candidate, State, and Analysis to their zero values via `Clear()`. PinnedBaseline is never modified by Clear.

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
| Start comparison with PinnedBaseline | Compare (auto) | Status, implicit |

### Prohibited synonyms

| Avoid | Use Instead |
|-------|-------------|
| Mark / Pin / Reference interchangeably | Mark Baseline / Pinned Baseline |
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
| `TestV0102Workflow_ClearPreservesPin` | Compare, x → session cleared, pin intact |
| `TestV0102Workflow_SwapTwiceReturns` | s then s → original orientation restored |
| `TestV0102Workflow_ExitPreservesSession` | Esc → Observe → session preserved |
| `TestV0102Workflow_ClearComparison` | Compare, x → Idle, Observe |
| `TestV0102Workflow_CrossRunPin` | Mark → startRun → new results → compare with pin |
| `TestV0102Workflow_MarkSameUnmarks` | c on already-marked → clears |
| `TestV0102Workflow_NoBaselineCleared` | PinnedBaseline only → UI shows ● Pinned Baseline |

---

## Related Documents

- `COMPARE_CONSTITUTION.md` — Architecture, engine, rendering hierarchy
- `RENDERING.md` — Visual presentation rules
- `STATE_OWNERSHIP.md` — Data ownership and lifecycle
