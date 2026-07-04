package tui

import (
	"sort"
	"strings"

	"github.com/divijg19/Pulse/internal/model"
)

var promotedKeys = map[string]bool{
	"Content-Type":     true,
	"Content-Encoding": true,
	"Content-Length":   true,
}

func (m Model) renderInspectWhy(result model.Result, maxLines int) string {
	var b strings.Builder
	if result.Error != "" {
		b.WriteString(styleError.Render("Error: " + result.Error))
		return b.String()
	}
	if len(result.ResponseHeaders) == 0 {
		b.WriteString(styleMuted.Render("No headers captured."))
	} else {
		keys := make([]string, 0, len(result.ResponseHeaders))
		for key := range result.ResponseHeaders {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		lines := 0
		for _, key := range keys {
			if promotedKeys[key] {
				continue
			}
			if maxLines > 0 && lines >= maxLines {
				b.WriteString(styleMuted.Render("..."))
				break
			}
			b.WriteString(renderMetadata(key, result.ResponseHeaders[key]))
			b.WriteString("\n")
			lines++
		}
	}
	return strings.TrimSuffix(b.String(), "\n")
}
