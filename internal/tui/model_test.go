package tui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/divijg19/Pulse/internal/model"
)

func TestFocusMovement(t *testing.T) {
	m := NewModel()
	if m.focus != focusURL {
		t.Fatalf("initial focus = %v", m.focus)
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.focus != focusConcurrency {
		t.Fatalf("focus after tab = %v", m.focus)
	}
}

func TestMethodSelection(t *testing.T) {
	m := NewModel()
	m.focus = focusMethod

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if got := m.methodIndex; got != 1 {
		t.Fatalf("method index = %d", got)
	}
}

func TestConcurrencyClamping(t *testing.T) {
	m := NewModel()
	m.setConcurrency(500)
	if got := m.concurrency(); got != 100 {
		t.Fatalf("concurrency high clamp = %d", got)
	}
	m.setConcurrency(-10)
	if got := m.concurrency(); got != 1 {
		t.Fatalf("concurrency low clamp = %d", got)
	}
}

func TestPayloadEditorState(t *testing.T) {
	m := NewModel()
	m.focus = focusPayload

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if !m.showPayload {
		t.Fatal("payload editor should be visible")
	}
	if len(m.headers) != 1 {
		t.Fatalf("headers len = %d", len(m.headers))
	}
}

func TestRunStartAndCancelTransitions(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	m := NewModel()
	m.urlInput.SetValue(upstream.URL)
	m.setConcurrency(1)

	started, cmd := m.startRun()
	if cmd == nil {
		t.Fatal("startRun should return a command")
	}
	if !started.running {
		t.Fatal("model should be running")
	}

	cancelled := started.cancelRun()
	if cancelled.status != "CANCELLED" {
		t.Fatalf("status = %q", cancelled.status)
	}
}

func TestTabSwitching(t *testing.T) {
	m := NewModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	m = updated.(Model)
	if m.activeTab != tabLogs {
		t.Fatalf("active tab = %v", m.activeTab)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	m = updated.(Model)
	if m.activeTab != tabTimeline {
		t.Fatalf("active tab = %v", m.activeTab)
	}
}

func TestResultSelectionAndInspector(t *testing.T) {
	m := NewModel()
	m.focus = focusResults
	m.results = []model.Result{
		{Status: http.StatusOK},
		{Status: http.StatusInternalServerError},
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.selected != 1 {
		t.Fatalf("selected = %d", m.selected)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if !m.inspector {
		t.Fatal("inspector should be open")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.inspector {
		t.Fatal("inspector should close on esc")
	}
}

func TestHeaderAddRemove(t *testing.T) {
	m := NewModel()
	m.showPayload = true
	m.focus = focusHeaders

	// handleHeaderKey auto-adds one header when the slice is empty,
	// so after the first ctrl+n there will be 2.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	m = updated.(Model)
	if len(m.headers) != 2 {
		t.Fatalf("after ctrl+n headers = %d (expected 2)", len(m.headers))
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = updated.(Model)
	if len(m.headers) != 1 {
		t.Fatalf("after ctrl+d headers = %d (expected 1)", len(m.headers))
	}
}

func TestRejectsOversizedBody(t *testing.T) {
	m := NewModel()
	m.urlInput.SetValue("https://example.com/api")
	m.setConcurrency(1)
	m.bodyInput.CharLimit = maxTUIBodyBytes + 100
	m.bodyInput.SetValue(strings.Repeat("x", maxTUIBodyBytes+1))

	started, cmd := m.startRun()
	if cmd != nil {
		t.Fatal("startRun should return nil cmd when body too large")
	}
	if started.running {
		t.Fatal("model should not be running with oversized body")
	}
	if started.status != "BODY TOO LARGE (MAX 1MB)" {
		t.Fatalf("status = %q", started.status)
	}
}

func TestAutoScrollToggle(t *testing.T) {
	m := NewModel()
	if !m.autoScroll {
		t.Fatal("autoScroll should default to true")
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	m = updated.(Model)
	if m.autoScroll {
		t.Fatal("autoScroll should toggle to false")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	m = updated.(Model)
	if !m.autoScroll {
		t.Fatal("autoScroll should toggle back to true")
	}

	// autoScroll should reset to true on startRun
	m.autoScroll = false
	m.urlInput.SetValue("https://example.com/api")
	m, _ = m.startRun()
	if !m.autoScroll {
		t.Fatal("autoScroll should reset to true on startRun")
	}
}
