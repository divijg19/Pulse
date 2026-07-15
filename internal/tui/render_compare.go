package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/divijg19/Pulse/internal/model"
)

var (
	styleRegression  = lipgloss.NewStyle().Foreground(lipgloss.Color(colorError)).Bold(true)
	styleImprovement = lipgloss.NewStyle().Foreground(lipgloss.Color(colorSuccess)).Bold(true)
	styleVerdict     = lipgloss.NewStyle().Bold(true)
	styleFieldName   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
	styleDiffWorse   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorError))
	styleDiffBetter  = lipgloss.NewStyle().Foreground(lipgloss.Color(colorSuccess))
	styleDiffNeutral = lipgloss.NewStyle().Foreground(lipgloss.Color(colorWarning))
	styleCompareMark = lipgloss.NewStyle().Foreground(lipgloss.Color(colorAccent)).Bold(true)
)

// renderCompare is the Compare workspace entry point. It builds the immutable
// CompareContext once and dispatches to the active view renderer. No view
// renderer performs analysis or mutates workflow state.
func (m Model) renderCompare(region Region) string {
	w := m.workspace.compare
	if !w.IsComparing() {
		if w.HasReference() {
			return regionStyle(region).Render(
				styleCompareMark.Render("● Pinned Baseline") + "\n" +
					styleMuted.Render("Select a result and press c to compare against the pinned baseline."))
		}
		return regionStyle(region).Render(styleMuted.Render("No comparison active."))
	}

	ctx := w.Context()
	if ctx.Analysis == nil {
		return regionStyle(region).Render(styleMuted.Render("Computing comparison..."))
	}
	if region.Width < 80 {
		return regionStyle(region).Render(styleMuted.Render("Comparison requires at least 80 columns."))
	}

	identity := m.renderComparisonIdentityBlock(&ctx.Baseline, &ctx.Candidate)

	var body string
	switch w.View {
	case CompareViewOverview:
		body = renderCompareOverview(ctx)
	case CompareViewEvidence:
		body = renderCompareEvidence(ctx)
	case CompareViewDiff:
		body = renderCompareDiff(ctx)
	case CompareViewHeaders:
		body = renderCompareHeaders(ctx)
	case CompareViewBody:
		body = m.renderCompareBody(ctx, bodyRegion(region, identity))
	case CompareViewRaw:
		body = m.renderCompareRaw(ctx, bodyRegion(region, identity))
	}

	var parts []string
	parts = appendNonEmpty(parts, identity, body)
	return regionStyle(region).Render(strings.Join(parts, "\n\n"))
}

// requestNumber returns the 1-based request number for r within the current run,
// or the stored Sequence if r is from a prior run (e.g., a reference). The number is
// a property of the request itself, not derived from UI position. When r has no
// resolvable number it returns a neutral placeholder so callers keep the #NNN form.
func (m Model) requestNumber(r model.Result) string {
	if r.Sequence != 0 {
		return fmt.Sprintf("#%03d", r.Sequence)
	}
	for i, res := range m.results {
		if resultsEqual(res, r) {
			return fmt.Sprintf("#%03d", i+1)
		}
	}
	return "#??"
}

func requestTime(r model.Result) string {
	if r.Timestamp.IsZero() {
		return "(no time)"
	}
	return r.Timestamp.Format("15:04:05")
}

// renderComparisonIdentityBlock renders the orientation header. It shows the
// baseline and, when present, the candidate, each annotated with its request
// number and timestamp. The same block is used by the full Compare workspace
// and the collapsed preview, so the two never drift.
func (m Model) renderComparisonIdentityBlock(baseline, candidate *model.Result) string {
	var parts []string
	if baseline != nil {
		parts = append(parts, m.renderIdentityLine("◆", "Baseline", *baseline))
	}
	if candidate != nil {
		parts = append(parts, m.renderIdentityLine("▶", "Candidate", *candidate))
	}
	return strings.Join(parts, "\n\n")
}

// renderIdentityLine renders one participant's marker, label, request number,
// timestamp and method+path on two lines. Shared by the comparison identity
// block and the preview drawer so the two presentations never diverge.
func (m Model) renderIdentityLine(marker, label string, r model.Result) string {
	var b strings.Builder
	b.WriteString(styleCompareMark.Render(fmt.Sprintf("%s %s %s", marker, label, m.requestNumber(r))))
	b.WriteString(styleMuted.Render(" · " + requestTime(r)))
	b.WriteString("\n")
	b.WriteString(styleMuted.Render(methodPath(r)))
	return b.String()
}

func methodPath(result model.Result) string {
	method := result.RequestMethod
	if method == "" {
		method = "GET"
	}
	return method + " " + result.RequestURL
}

// --- View renderers --------------------------------------------------------

// Every view renderer consumes an immutable CompareContext.

func renderCompareOverview(ctx CompareContext) string {
	var parts []string
	parts = appendNonEmpty(parts, renderVerdict(ctx.Analysis), renderWhy(ctx.Analysis.Flags))
	return strings.Join(parts, "\n\n")
}

func renderCompareEvidence(ctx CompareContext) string {
	return renderEvidenceSection(ctx.Analysis)
}

func renderCompareDiff(ctx CompareContext) string {
	return renderDetailsSection(ctx.Analysis)
}

func renderCompareHeaders(ctx CompareContext) string {
	hdrs := ctx.Analysis.Headers

	// Each section shares the same shape (title + per-entry lines); only the
	// entry formatting differs between added/removed and changed.
	sections := []struct {
		title  string
		render func() string
	}{
		{"ADDED HEADERS", func() string {
			var sb strings.Builder
			for _, e := range hdrs.Added {
				sb.WriteString(fmt.Sprintf("  %s: %s\n", styleFieldName.Render(e.Name), e.Value))
			}
			return sb.String()
		}},
		{"REMOVED HEADERS", func() string {
			var sb strings.Builder
			for _, e := range hdrs.Removed {
				sb.WriteString(fmt.Sprintf("  %s: %s\n", styleFieldName.Render(e.Name), e.Value))
			}
			return sb.String()
		}},
		{"CHANGED HEADERS", func() string {
			var sb strings.Builder
			for _, e := range hdrs.Changed {
				sb.WriteString(fmt.Sprintf("  %s: %s → %s\n", styleFieldName.Render(e.Name), e.OldValue, e.NewValue))
			}
			return sb.String()
		}},
	}

	var b strings.Builder
	for _, s := range sections {
		body := s.render()
		if body == "" {
			continue
		}
		b.WriteString(sectionLine(s.title, false))
		b.WriteString(body)
		b.WriteString("\n")
	}
	if b.Len() == 0 {
		return styleMuted.Render("No header differences")
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m Model) renderCompareBody(ctx CompareContext, region Region) string {
	var b strings.Builder
	b.WriteString(sectionLine("BASELINE BODY", false))
	b.WriteString("\n")
	b.WriteString(indentBlock(ctx.Baseline.ResponseBody))
	b.WriteString("\n\n")
	b.WriteString(sectionLine("CANDIDATE BODY", false))
	b.WriteString("\n")
	b.WriteString(indentBlock(ctx.Candidate.ResponseBody))
	return withScrollHint(b.String(), m.inspectBodyOffset, region.Height)
}

func (m Model) renderCompareRaw(ctx CompareContext, region Region) string {
	var b strings.Builder
	b.WriteString(sectionLine("BASELINE", false))
	b.WriteString("\n")
	b.WriteString(m.renderRawResult(ctx.Baseline))
	b.WriteString("\n\n")
	b.WriteString(sectionLine("CANDIDATE", false))
	b.WriteString("\n")
	b.WriteString(m.renderRawResult(ctx.Candidate))
	return withScrollHint(b.String(), m.inspectBodyOffset, region.Height)
}

// withScrollHint clips content vertically (like scrollContent) and, when the
// content overflows, appends a muted hint showing the current scroll position
// on the final line so the operator knows the view is scrollable.
func withScrollHint(content string, offset, height int) string {
	if height < 1 {
		height = 1
	}
	lines := strings.Split(content, "\n")
	maxOffset := len(lines) - height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if maxOffset == 0 {
		return scrollContent(content, offset, height)
	}
	// Only show the hint when there is room for both content and the hint
	// line; otherwise fall back to plain scrolling to avoid overflow.
	if height < 2 {
		return scrollContent(content, offset, height)
	}
	// Reserve the last row for the hint.
	scrolled := scrollContent(content, offset, height-1)
	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	hint := fmt.Sprintf("↑/↓ scroll · %d/%d", offset+1, maxOffset+1)
	return scrolled + "\n" + styleMuted.Render(hint)
}

func (m Model) renderRawResult(r model.Result) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("  %s %s\n", styleFieldName.Render("Request:"), m.requestNumber(r)+" · "+requestTime(r)))
	b.WriteString(fmt.Sprintf("  %s %d\n", styleFieldName.Render("Status:"), r.Status))
	b.WriteString(fmt.Sprintf("  %s %s\n", styleFieldName.Render("Latency:"), formatDuration(r.Latency)))
	b.WriteString(fmt.Sprintf("  %s %s\n", styleFieldName.Render("Method:"), r.RequestMethod))
	b.WriteString(fmt.Sprintf("  %s %s\n", styleFieldName.Render("URL:"), r.RequestURL))
	if len(r.ResponseHeaders) > 0 {
		b.WriteString(fmt.Sprintf("  %s\n", styleFieldName.Render("Headers:")))
		for k, v := range r.ResponseHeaders {
			b.WriteString(fmt.Sprintf("    %s: %s\n", k, v))
		}
	}
	if r.ResponseBody != "" {
		b.WriteString(fmt.Sprintf("  %s\n", styleFieldName.Render("Body:")))
		b.WriteString(indentBlock(r.ResponseBody))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// bodyRegion returns the vertical space available to the scrollable Body/Raw
// views after the fixed identity header. It never returns a non-positive
// height so the scroll helper always has valid bounds.
func bodyRegion(region Region, identity string) Region {
	h := region.Height - (strings.Count(identity, "\n") + 1) - 2
	if h < 1 {
		h = 1
	}
	return Region{Width: region.Width, Height: h}
}

// scrollContent clips content to height rows starting at the given scroll
// offset (clamped to valid bounds). Horizontal width is applied by the
// caller's region style; only vertical scrolling is performed here.
func scrollContent(content string, offset, height int) string {
	if height < 1 {
		height = 1
	}
	lines := strings.Split(content, "\n")
	if offset < 0 {
		offset = 0
	}
	maxOffset := len(lines) - height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	end := offset + height
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[offset:end], "\n")
}

func indentBlock(s string) string {
	if s == "" {
		return "  " + styleMuted.Render("(empty)")
	}
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = "  " + l
	}
	return strings.Join(lines, "\n")
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
	prefSet := make(map[string]bool, len(preferred))
	for _, p := range preferred {
		prefSet[p] = true
	}
	sorted := make([]string, 0, len(order))
	for _, p := range preferred {
		for _, f := range order {
			if f == p {
				sorted = append(sorted, f)
			}
		}
	}
	for _, f := range order {
		if !prefSet[f] {
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
			switch {
			case d.regression:
				return whyLine{styleRegression.Render("▼"), "Headers regressed"}
			case d.improvement:
				return whyLine{styleImprovement.Render("▲"), "Headers improved"}
			default:
				return whyLine{styleMuted.Render("·"), "Headers changed"}
			}
		},
		"body": func(d *fieldDir) whyLine {
			switch {
			case d.regression:
				return whyLine{styleRegression.Render("▼"), "Body regressed"}
			case d.improvement:
				return whyLine{styleImprovement.Render("▲"), "Body improved"}
			default:
				return whyLine{styleMuted.Render("·"), "Body changed"}
			}
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
