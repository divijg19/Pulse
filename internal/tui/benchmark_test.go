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

func BenchmarkRenderer_CompareAnalysis(b *testing.B) {
	results := testResults(20)
	baseline := results[0]
	candidate := results[1]
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		AnalyzeComparison(baseline, candidate)
	}
}

func BenchmarkRenderer_CompareRender(b *testing.B) {
	m := NewModel()
	m.results = testResults(20)
	m.workspace.compare.Baseline = &m.results[0]
	m.workspace.compare.Candidate = &m.results[1]
	m.workspace.compare.State = CompareComparing
	m.workspace.compare.refreshAnalysis()
	region := Region{Width: 100, Height: 30}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.renderCompare(region)
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
