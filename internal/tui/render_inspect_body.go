package tui

import (
	"strings"
	"unicode/utf8"

	"github.com/divijg19/Pulse/internal/model"
)

func (m Model) renderInspectBody(result model.Result, maxLines, width int) string {
	body := result.ResponseBody
	if body == "" {
		return styleMuted.Render("No body captured.")
	}

	if isBinaryContent(result) {
		ct := result.ResponseHeaders["Content-Type"]
		if ct == "" {
			ct = "unknown"
		}
		ce := result.ResponseHeaders["Content-Encoding"]
		info := "Binary content"
		if ce != "" {
			info += " (" + ce + ")"
		}
		info += " — " + ct
		cl := len(body)
		if cl > 0 {
			info += " — " + formatContentLength(cl)
		}
		return styleMuted.Render(info)
	}

	display := renderBodyPreview(body, maxLines+m.inspectBodyOffset)
	lines := splitLines(display, width)
	if m.inspectBodyOffset >= len(lines) {
		m.inspectBodyOffset = max(0, len(lines)-1)
	}

	start := m.inspectBodyOffset
	end := start + maxLines
	if end > len(lines) {
		end = len(lines)
	}
	if start >= end {
		return styleMuted.Render("(empty)")
	}
	return strings.Join(lines[start:end], "\n")
}
func splitLines(display string, width int) []string {
	raw := strings.Split(display, "\n")
	if width <= 0 {
		return raw
	}
	var result []string
	for _, line := range raw {
		if utf8.RuneCountInString(line) <= width {
			result = append(result, line)
			continue
		}
		for len(line) > 0 {
			if utf8.RuneCountInString(line) <= width {
				result = append(result, line)
				break
			}
			runes := []rune(line)
			result = append(result, string(runes[:width]))
			line = string(runes[width:])
		}
	}
	return result
}

func isBinaryContent(result model.Result) bool {
	ct := result.ResponseHeaders["Content-Type"]
	if ct == "" {
		return false
	}
	lower := strings.ToLower(ct)
	binaryPrefixes := []string{
		"image/", "audio/", "video/",
		"application/octet-stream",
		"application/pdf",
		"application/zip",
		"application/x-tar",
		"application/gzip",
		"application/xml",
	}
	for _, prefix := range binaryPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}
