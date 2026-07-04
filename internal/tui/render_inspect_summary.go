package tui

import (
	"fmt"
	"strings"

	"github.com/divijg19/Pulse/internal/model"
)

func (m Model) renderInspectSummary(result model.Result, method, reqURL string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s %s", renderMethod(method), reqURL))
	b.WriteString("\n")
	b.WriteString(renderStatusBadge(result))
	b.WriteString("\n")
	b.WriteString(renderMetadata("Latency", formatDuration(result.Latency)))
	b.WriteString("\n")
	if ct := result.ResponseHeaders["Content-Type"]; ct != "" {
		b.WriteString(renderMetadata("Content-Type", ct))
		b.WriteString("\n")
	}
	if cl := formatContentLength(len(result.ResponseBody)); cl != "" {
		b.WriteString(renderMetadata("Content-Length", cl))
		b.WriteString("\n")
	}
	if ce := result.ResponseHeaders["Content-Encoding"]; ce != "" {
		b.WriteString(renderMetadata("Encoding", ce))
		b.WriteString("\n")
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func formatContentLength(n int) string {
	if n <= 0 {
		return ""
	}
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
