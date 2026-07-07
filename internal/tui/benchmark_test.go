package tui

import (
	"testing"
)

func BenchmarkView_Ready(b *testing.B) {
	m := NewModel()
	m.shell.Resize(80, 24)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.View()
	}
}

func BenchmarkView_TimelineRunning(b *testing.B) {
	m := newTimelineRunningModel()
	m.shell.Resize(100, 30)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.View()
	}
}

func BenchmarkView_RequestDialog(b *testing.B) {
	m := newRequestPayloadModel()
	m.shell.Resize(100, 30)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.View()
	}
}

func BenchmarkView_Inspect(b *testing.B) {
	m := newInspectModel()
	m.shell.Resize(100, 30)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.View()
	}
}

func BenchmarkRenderer_RequestDomain(b *testing.B) {
	m := newRequestModel()
	m.shell.Resize(100, 30)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.renderRequestDomain(100)
	}
}

func BenchmarkRenderer_PayloadDomain(b *testing.B) {
	m := newRequestPayloadModel()
	m.shell.Resize(100, 30)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.renderPayloadDomain(100)
	}
}

func BenchmarkRenderer_CompareDiff(b *testing.B) {
	m := NewModel()
	results := testResults(20)
	marked := results[0]
	active := results[1]
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.renderCompareDiff(marked, active)
	}
}

func BenchmarkRenderer_Inspect(b *testing.B) {
	m := newInspectModel()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.renderInspect(Region{Width: 100, Height: 30})
	}
}

func BenchmarkRenderer_Statusline(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := NewModel()
		m.shell.Resize(100, 24)
		m.renderStatusline(m.ShellState(), 100)
	}
}
