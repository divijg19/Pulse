package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	styleRegression  = lipgloss.NewStyle().Foreground(lipgloss.Color(colorError)).Bold(true)
	styleImprovement = lipgloss.NewStyle().Foreground(lipgloss.Color(colorSuccess)).Bold(true)
	styleVerdict     = lipgloss.NewStyle().Bold(true)
	styleFieldName   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
	styleDiffWorse   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorError))
	styleDiffBetter  = lipgloss.NewStyle().Foreground(lipgloss.Color(colorSuccess))
	styleDiffNeutral = lipgloss.NewStyle().Foreground(lipgloss.Color(colorWarning))
)

type compareSections struct {
	verdict  string
	why      string
	evidence string
	details  string
}

func (m Model) renderCompare(region Region) string {
	if m.workspace.compare.Session.State != SessionComparing {
		if m.workspace.compare.PinnedBaseline != nil {
			return regionStyle(region).Render(
				styleCompareMarked.Render("📌 Baseline pinned") + "\n" +
					styleMuted.Render("Select a result and press c to compare against the pinned baseline."))
		}
		return regionStyle(region).Render(styleMuted.Render("No comparison active."))
	}

	analysis := m.workspace.compare.Session.Analysis
	if analysis == nil {
		analysis = m.computeComparisonAnalysis()
	}
	if analysis == nil {
		return regionStyle(region).Render(styleMuted.Render("Computing comparison..."))
	}

	if region.Width < 80 {
		return regionStyle(region).Render(styleMuted.Render("Comparison requires at least 80 columns."))
	}

	sec := compareSections{
		verdict:  renderVerdict(analysis),
		why:      renderWhy(analysis.Flags),
		evidence: renderEvidenceSection(analysis),
		details:  renderDetailsSection(analysis),
	}

	var parts []string
	parts = appendNonEmpty(parts, sec.verdict, sec.why, sec.evidence, sec.details)
	return regionStyle(region).Render(strings.Join(parts, "\n\n"))
}

func renderVerdict(analysis *ComparisonAnalysis) string {
	switch analysis.Verdict {
	case VerdictRegressed:
		return styleRegression.Render("Regression detected")
	case VerdictImproved:
		return styleImprovement.Render("Improvement detected")
	case VerdictMixed:
		return styleVerdict.Render("Mixed results")
	default:
		return styleVerdict.Render("No significant changes")
	}
}

func renderWhy(flags []Flag) string {
	if len(flags) == 0 {
		return ""
	}

	type fieldDir struct {
		regression  bool
		improvement bool
		change      bool
	}

	dirs := make(map[string]*fieldDir)
	order := make([]string, 0, len(flags))

	for _, f := range flags {
		if _, ok := dirs[f.Field]; !ok {
			dirs[f.Field] = &fieldDir{}
			order = append(order, f.Field)
		}
		switch f.Severity {
		case FlagRegression:
			dirs[f.Field].regression = true
		case FlagImprovement:
			dirs[f.Field].improvement = true
		default:
			dirs[f.Field].change = true
		}
	}

	preferred := []string{"status", "latency", "headers", "body", "error"}
	sorted := make([]string, 0, len(order))
	for _, p := range preferred {
		for _, f := range order {
			if f == p {
				sorted = append(sorted, f)
			}
		}
	}
	for _, f := range order {
		found := false
		for _, s := range sorted {
			if f == s {
				found = true
				break
			}
		}
		if !found {
			sorted = append(sorted, f)
		}
	}

	type whyLine struct {
		prefix string
		text   string
	}

	sentences := map[string]func(*fieldDir) whyLine{
		"status": func(d *fieldDir) whyLine {
			switch {
			case d.regression:
				return whyLine{styleRegression.Render("▼"), "Status regressed"}
			case d.improvement:
				return whyLine{styleImprovement.Render("▲"), "Status improved"}
			default:
				return whyLine{styleMuted.Render("·"), "Status changed"}
			}
		},
		"latency": func(d *fieldDir) whyLine {
			switch {
			case d.regression:
				return whyLine{styleRegression.Render("▼"), "Latency increased"}
			case d.improvement:
				return whyLine{styleImprovement.Render("▲"), "Latency decreased"}
			default:
				return whyLine{styleMuted.Render("·"), "Latency changed"}
			}
		},
		"error": func(d *fieldDir) whyLine {
			switch {
			case d.regression:
				return whyLine{styleRegression.Render("▼"), "Error introduced"}
			case d.improvement:
				return whyLine{styleImprovement.Render("▲"), "Error resolved"}
			default:
				return whyLine{styleMuted.Render("·"), "Error changed"}
			}
		},
		"headers": func(d *fieldDir) whyLine {
			return whyLine{styleMuted.Render("·"), "Headers changed"}
		},
		"body": func(d *fieldDir) whyLine {
			return whyLine{styleMuted.Render("·"), "Body changed"}
		},
	}

	var lines []string
	for _, field := range sorted {
		fn, ok := sentences[field]
		if !ok {
			continue
		}
		wl := fn(dirs[field])
		lines = append(lines, "  "+wl.prefix+" "+wl.text+".")
	}

	return sectionLine("WHY", false) + "\n" + strings.Join(lines, "\n")
}

func renderEvidenceSection(analysis *ComparisonAnalysis) string {
	meta := analysis.Metadata
	hdrs := analysis.Headers
	body := analysis.Body

	categories := make([]string, 0, 5)

	if meta.Status.Changed {
		sty := statusDeltaColor(meta.Status.Old, meta.Status.New)
		categories = append(categories,
			fmt.Sprintf("  %s %s → %s",
				styleFieldName.Render("Status:"),
				sty.Render(fmt.Sprintf("%d", meta.Status.Old)),
				sty.Render(fmt.Sprintf("%d", meta.Status.New))))
	}

	if meta.Latency.Changed {
		sty := latencyDeltaColor(meta.Latency.Old, meta.Latency.New)
		delta := meta.Latency.New - meta.Latency.Old
		deltaStr := fmt.Sprintf("%v", delta)
		if delta > 0 {
			deltaStr = "+" + deltaStr
		}
		categories = append(categories,
			fmt.Sprintf("  %s %s → %s (%s)",
				styleFieldName.Render("Latency:"),
				sty.Render(fmt.Sprintf("%v", meta.Latency.Old)),
				sty.Render(fmt.Sprintf("%v", meta.Latency.New)),
				sty.Render(deltaStr)))
	}

	totalHdrs := len(hdrs.Added) + len(hdrs.Removed) + len(hdrs.Changed)
	if totalHdrs > 0 {
		categories = append(categories,
			fmt.Sprintf("  %s %d header(s) differ",
				styleFieldName.Render("Headers:"), totalHdrs))
	}

	if body.BaselineSize != body.CandidateSize || body.ChangedLines > 0 {
		sizeDelta := body.CandidateSize - body.BaselineSize
		var sizePart string
		switch {
		case sizeDelta > 0:
			sizePart = fmt.Sprintf("%d → %d bytes (+%d)", body.BaselineSize, body.CandidateSize, sizeDelta)
		case sizeDelta < 0:
			sizePart = fmt.Sprintf("%d → %d bytes (%d)", body.BaselineSize, body.CandidateSize, sizeDelta)
		default:
			sizePart = fmt.Sprintf("%d bytes (no change)", body.BaselineSize)
		}
		if body.ChangedLines > 0 {
			sizePart += fmt.Sprintf(", %d line(s) differ", body.ChangedLines)
		}
		categories = append(categories,
			fmt.Sprintf("  %s %s", styleFieldName.Render("Body:"), sizePart))
	}

	if meta.Error.Changed {
		var oldS, newS string
		if meta.Error.Old == "" {
			oldS = styleMuted.Render("(none)")
		} else {
			oldS = styleDiffWorse.Render(meta.Error.Old)
		}
		if meta.Error.New == "" {
			newS = styleImprovement.Render("(resolved)")
		} else {
			newS = styleDiffWorse.Render(meta.Error.New)
		}
		categories = append(categories,
			fmt.Sprintf("  %s %s → %s", styleFieldName.Render("Error:"), oldS, newS))
	}

	if len(categories) == 0 {
		return ""
	}

	return sectionLine("EVIDENCE", false) + "\n" +
		strings.Join(categories, "\n")
}

func renderDetailsSection(analysis *ComparisonAnalysis) string {
	var b strings.Builder
	meta := analysis.Metadata
	body := analysis.Body

	if meta.URL.Changed {
		b.WriteString(fmt.Sprintf("  %s\n    %s\n    %s\n",
			styleFieldName.Render("URL:"),
			styleDiffWorse.Render(meta.URL.Old),
			styleDiffBetter.Render(meta.URL.New)))
	}

	if body.BaselineSize != body.CandidateSize || body.ChangedLines > 0 {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString(renderBodyDiff(body))
	}

	if b.Len() == 0 {
		return ""
	}
	return sectionLine("DETAILS", false) + "\n" + strings.TrimRight(b.String(), "\n")
}

func renderBodyDiff(body BodyAnalysis) string {
	var b strings.Builder
	hasContent := false
	for _, seg := range body.Segments {
		switch seg.Kind {
		case SegmentEqual:
			if hasContent {
				for _, line := range seg.Old {
					b.WriteString("  ")
					b.WriteString(styleMuted.Render(line))
					b.WriteString("\n")
				}
			}
		case SegmentDelete:
			hasContent = true
			for _, line := range seg.Old {
				b.WriteString("  ")
				b.WriteString(styleDiffWorse.Render("- " + line))
				b.WriteString("\n")
			}
		case SegmentInsert:
			hasContent = true
			for _, line := range seg.New {
				b.WriteString("  ")
				b.WriteString(styleDiffBetter.Render("+ " + line))
				b.WriteString("\n")
			}
		}
	}
	return b.String()
}

func statusDeltaColor(old, new int) lipgloss.Style {
	oldClass := ClassifyStatus(old)
	newClass := ClassifyStatus(new)
	if (newClass == StatusClientError || newClass == StatusServerError) &&
		oldClass != StatusClientError && oldClass != StatusServerError {
		return styleDiffWorse
	}
	if (oldClass == StatusClientError || oldClass == StatusServerError) &&
		newClass != StatusClientError && newClass != StatusServerError {
		return styleDiffBetter
	}
	return styleDiffNeutral
}

func latencyDeltaColor(old, new time.Duration) lipgloss.Style {
	if new < old {
		return styleDiffBetter
	}
	if new > old {
		return styleDiffWorse
	}
	return styleMuted
}

func appendNonEmpty(parts []string, sections ...string) []string {
	for _, s := range sections {
		if s != "" {
			parts = append(parts, s)
		}
	}
	return parts
}
