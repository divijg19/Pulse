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
	styleAnomaly     = lipgloss.NewStyle().Foreground(lipgloss.Color(colorWarning)).Bold(true)
	styleVerdict     = lipgloss.NewStyle().Bold(true)
	styleFieldName   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
	styleDiffWorse   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorError))
	styleDiffBetter  = lipgloss.NewStyle().Foreground(lipgloss.Color(colorSuccess))
	styleDiffNeutral = lipgloss.NewStyle().Foreground(lipgloss.Color(colorWarning))
)

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

	verdict := renderVerdict(analysis)
	flags := renderFlags(analysis.Flags)
	meta := renderMetadataDeltas(analysis.Metadata)
	hdrs := renderHeaderDeltas(analysis.Headers)
	bodyS := renderBodyAnalysis(analysis.Body)

	switch {
	case region.Width >= 120:
		return renderCompareWide(region, verdict, flags, meta, hdrs, bodyS)
	case region.Width >= 80:
		return renderCompareMedium(region, verdict, flags, meta, hdrs, bodyS)
	default:
		return renderCompareNarrow(region, verdict, flags, meta, hdrs, bodyS)
	}
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

func renderFlags(flags []Flag) string {
	if len(flags) == 0 {
		return ""
	}
	var lines []string
	for _, f := range flags {
		var s string
		switch f.Severity {
		case FlagRegression:
			s = styleRegression.Render("▼ " + f.Message)
		case FlagImprovement:
			s = styleImprovement.Render("▲ " + f.Message)
		case FlagAnomaly:
			s = styleAnomaly.Render("! " + f.Message)
		default:
			s = styleMuted.Render("· " + f.Message)
		}
		lines = append(lines, "  "+s)
	}
	return strings.Join(lines, "\n")
}

func renderMetadataDeltas(meta MetadataDelta) string {
	var b strings.Builder
	b.WriteString(sectionLine("METADATA", false) + "\n")

	if !meta.Status.Changed && !meta.Latency.Changed && !meta.URL.Changed && !meta.Error.Changed {
		b.WriteString("  " + styleMuted.Render("No changes in metadata.") + "\n")
		return b.String()
	}

	if meta.Status.Changed {
		sty := statusDeltaColor(meta.Status.Old, meta.Status.New)
		b.WriteString(fmt.Sprintf("  %s %s → %s\n",
			styleFieldName.Render("Status:"),
			styleDiffWorse.Render(fmt.Sprintf("%d", meta.Status.Old)),
			sty.Render(fmt.Sprintf("%d", meta.Status.New))))
	}
	if meta.Latency.Changed {
		sty := latencyDeltaColor(meta.Latency.Old, meta.Latency.New)
		delta := meta.Latency.New - meta.Latency.Old
		deltaStr := fmt.Sprintf("%v", delta)
		if delta > 0 {
			deltaStr = "+" + deltaStr
		}
		b.WriteString(fmt.Sprintf("  %s %s → %s (%s)\n",
			styleFieldName.Render("Latency:"),
			styleMuted.Render(fmt.Sprintf("%v", meta.Latency.Old)),
			sty.Render(fmt.Sprintf("%v", meta.Latency.New)),
			sty.Render(deltaStr)))
	}
	if meta.URL.Changed {
		b.WriteString(fmt.Sprintf("  %s\n    %s\n    %s\n",
			styleFieldName.Render("URL:"),
			styleDiffWorse.Render(meta.URL.Old),
			styleDiffBetter.Render(meta.URL.New)))
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
		b.WriteString(fmt.Sprintf("  %s %s → %s\n", styleFieldName.Render("Error:"), oldS, newS))
	}
	return b.String()
}

func renderHeaderDeltas(hdrs HeaderDelta) string {
	var b strings.Builder
	total := len(hdrs.Added) + len(hdrs.Removed) + len(hdrs.Changed)
	if total == 0 {
		return ""
	}
	b.WriteString(sectionLine("HEADERS", false) + "\n")
	for _, h := range hdrs.Added {
		b.WriteString(fmt.Sprintf("  %s %s: %s\n", styleDiffBetter.Render("+"), h.Name, h.Value))
	}
	for _, h := range hdrs.Removed {
		b.WriteString(fmt.Sprintf("  %s %s: %s\n", styleDiffWorse.Render("-"), h.Name, h.Value))
	}
	for _, h := range hdrs.Changed {
		b.WriteString(fmt.Sprintf("  %s %s:\n    %s %s\n    %s %s\n",
			styleDiffNeutral.Render("~"), h.Name,
			styleDiffWorse.Render("-"), h.OldValue,
			styleDiffBetter.Render("+"), h.NewValue))
	}
	return b.String()
}

func renderBodyAnalysis(body BodyAnalysis) string {
	var b strings.Builder
	if body.BaselineSize == 0 && body.CandidateSize == 0 && body.ChangedLines == 0 {
		return ""
	}
	b.WriteString(sectionLine("BODY", false) + "\n")

	sizeDelta := body.CandidateSize - body.BaselineSize
	var sizeLine string
	if sizeDelta > 0 {
		sizeLine = fmt.Sprintf("  %s %d bytes → %d bytes (+%d bytes)",
			styleDiffWorse.Render("▼"), body.BaselineSize, body.CandidateSize, sizeDelta)
	} else if sizeDelta < 0 {
		sizeLine = fmt.Sprintf("  %s %d bytes → %d bytes (%d bytes)",
			styleDiffBetter.Render("▲"), body.BaselineSize, body.CandidateSize, sizeDelta)
	} else {
		sizeLine = fmt.Sprintf("  %s %d bytes (no change)", styleMuted.Render("·"), body.BaselineSize)
	}
	b.WriteString(sizeLine + "\n")

	if body.ChangedLines > 0 {
		b.WriteString(fmt.Sprintf("  %s %d line(s) differ\n", styleDiffNeutral.Render("~"), body.ChangedLines))
	}

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

func renderCompareWide(region Region, verdict, flags, meta, hdrs, bodyS string) string {
	leftW := region.Width*48/100 - 1
	if leftW < 10 {
		leftW = 10
	}
	rightW := region.Width - leftW - 2
	if rightW < 10 {
		rightW = 10
	}

	var leftParts, rightParts []string
	leftParts = appendNonEmpty(leftParts, verdict, flags, meta)
	rightParts = appendNonEmpty(rightParts, hdrs, bodyS)

	left := lipgloss.NewStyle().Width(leftW).Render(strings.Join(leftParts, "\n\n"))
	right := lipgloss.NewStyle().Width(rightW).Render(strings.Join(rightParts, "\n\n"))

	content := lipgloss.JoinHorizontal(lipgloss.Top, left, " │ ", right)
	return regionStyle(region).Render(content)
}

func renderCompareMedium(region Region, verdict, flags, meta, hdrs, bodyS string) string {
	var parts []string
	parts = appendNonEmpty(parts, verdict, flags, meta, hdrs, bodyS)
	return regionStyle(region).Render(strings.Join(parts, "\n\n"))
}

func renderCompareNarrow(region Region, verdict, flags, meta, hdrs, bodyS string) string {
	if region.Width < 80 {
		return regionStyle(region).Render(styleMuted.Render("Comparison requires at least 80 columns."))
	}
	var parts []string
	parts = appendNonEmpty(parts, verdict, flags, meta, hdrs, bodyS)
	return regionStyle(region).Render(strings.Join(parts, "\n\n"))
}

func appendNonEmpty(parts []string, sections ...string) []string {
	for _, s := range sections {
		if s != "" {
			parts = append(parts, s)
		}
	}
	return parts
}
