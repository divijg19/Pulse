package tui

import (
	"strings"
	"testing"
)

func TestPayload_Hierarchy_Regression(t *testing.T) {
	m := NewModel()
	m.shell.Resize(100, 30)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.bodyInput.SetValue("line 1\nline 2\nline 3")

	// Ensure we have at least one header for the test
	if len(m.headers) == 0 {
		m.headers = append(m.headers, newHeaderRow())
		m.headers[0].Key.SetValue("Content-Type")
		m.headers[0].Value.SetValue("application/json")
	}

	out := m.renderPayloadDomain(100)
	lines := strings.Split(out, "\n")

	var headersHeading string
	var bodyHeading string
	var headerRow string
	var bodyLine string

	for _, line := range lines {
		if strings.Contains(line, "HEADERS") {
			headersHeading = line
		}
		if strings.Contains(line, "BODY") {
			bodyHeading = line
		}
		if strings.Contains(line, "Content-Type") {
			headerRow = line
		}
		if strings.Contains(line, "line 1") {
			bodyLine = line
		}
	}

	// 1. Verify section headings align
	if len(headersHeading) == 0 || len(bodyHeading) == 0 {
		t.Fatal("Headings not found")
	}
	if strings.Index(headersHeading, "HEADERS") != strings.Index(bodyHeading, "BODY") {
		t.Errorf("Headings misalignment: %q vs %q", headersHeading, bodyHeading)
	}

	// 2. Verify content alignment
	if len(headerRow) == 0 || len(bodyLine) == 0 {
		t.Fatal("Content lines not found")
	}
	if strings.Index(headerRow, "Content-Type") != strings.Index(bodyLine, "line 1") {
		t.Errorf("Content misalignment: %q vs %q", headerRow, bodyLine)
	}

	// 3. Verify hierarchy (Section < Content)
	if strings.Index(headerRow, "Content-Type") <= strings.Index(headersHeading, "HEADERS") {
		t.Errorf("Header row not indented beneath section")
	}
}

func TestPayload_Hierarchy_ContextPanelWide(t *testing.T) {
	m := NewModel()
	m.shell.Resize(160, 30)
	m.workspace.dialog = dialogRequest
	m.activeDomain = DomainPayload
	m.bodyInput.SetValue("line 1\nline 2\nline 3")

	if len(m.headers) == 0 {
		m.headers = append(m.headers, newHeaderRow())
		m.headers[0].Key.SetValue("Content-Type")
		m.headers[0].Value.SetValue("application/json")
	}

	out := m.renderPayloadDomain(100)
	lines := strings.Split(out, "\n")

	var headersHeading string
	var bodyHeading string
	var headerRow string
	var bodyLine string

	for _, line := range lines {
		if strings.Contains(line, "HEADERS") {
			headersHeading = line
		}
		if strings.Contains(line, "BODY") {
			bodyHeading = line
		}
		if strings.Contains(line, "Content-Type") {
			headerRow = line
		}
		if strings.Contains(line, "line 1") {
			bodyLine = line
		}
	}

	// 1. Verify section headings align at context-panel width
	if len(headersHeading) == 0 || len(bodyHeading) == 0 {
		t.Fatal("Headings not found at wide width")
	}
	if strings.Index(headersHeading, "HEADERS") != strings.Index(bodyHeading, "BODY") {
		t.Errorf("Wide: Headings misalignment: %q vs %q", headersHeading, bodyHeading)
	}

	// 2. Verify content alignment at context-panel width
	if len(headerRow) == 0 || len(bodyLine) == 0 {
		t.Fatal("Content lines not found at wide width")
	}
	if strings.Index(headerRow, "Content-Type") != strings.Index(bodyLine, "line 1") {
		t.Errorf("Wide: Content misalignment: %q vs %q", headerRow, bodyLine)
	}
}
