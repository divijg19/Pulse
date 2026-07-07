package tui

import (
	"fmt"
	"strings"

	"github.com/divijg19/Pulse/internal/model"
)

func (m Model) renderCompareDiff(marked, active model.Result) string {
	var b strings.Builder

	b.WriteString(sectionLine("DIFF SUMMARY", false))
	b.WriteString("\n")

	// Status diff
	if marked.Status != active.Status {
		b.WriteString(fmt.Sprintf("  Status: %s %s\n",
			styleDiffRemoved.Render(fmt.Sprintf("%d", marked.Status)),
			styleDiffAdded.Render(fmt.Sprintf("%d", active.Status)),
		))
	} else {
		b.WriteString(fmt.Sprintf("  Status: %s\n",
			styleMuted.Render(fmt.Sprintf("%d (no change)", marked.Status)),
		))
	}

	// Latency diff
	delta := active.Latency - marked.Latency
	if delta > 0 {
		b.WriteString(fmt.Sprintf("  Latency: %s %s\n",
			styleDiffRemoved.Render(fmt.Sprintf("%v", marked.Latency)),
			styleDiffAdded.Render(fmt.Sprintf("%v (+%v)", active.Latency, delta)),
		))
	} else if delta < 0 {
		b.WriteString(fmt.Sprintf("  Latency: %s %s\n",
			styleDiffRemoved.Render(fmt.Sprintf("%v", marked.Latency)),
			styleDiffAdded.Render(fmt.Sprintf("%v (%v)", active.Latency, delta)),
		))
	} else {
		b.WriteString(fmt.Sprintf("  Latency: %s\n",
			styleMuted.Render(fmt.Sprintf("%v (no change)", marked.Latency)),
		))
	}

	// URL diff
	if marked.RequestURL != active.RequestURL {
		b.WriteString(fmt.Sprintf("  URL:     %s → %s\n",
			styleDiffRemoved.Render(marked.RequestURL),
			styleDiffAdded.Render(active.RequestURL),
		))
	}

	// Error diff
	markedErr := marked.Error
	activeErr := active.Error
	if markedErr != activeErr {
		switch {
		case markedErr == "" && activeErr != "":
			b.WriteString(fmt.Sprintf("  Error:   %s\n", styleDiffAdded.Render(activeErr)))
		case markedErr != "" && activeErr == "":
			b.WriteString(fmt.Sprintf("  Error:   %s\n", styleDiffRemoved.Render(markedErr)))
		default:
			b.WriteString(fmt.Sprintf("  Error:   %s → %s\n",
				styleDiffRemoved.Render(markedErr),
				styleDiffAdded.Render(activeErr),
			))
		}
	}

	return b.String()
}
