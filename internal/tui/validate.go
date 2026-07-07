package tui

import (
	"encoding/json"
	"net/url"
	"strings"
)

func (m Model) fieldErrors() map[string]string {
	errs := make(map[string]string)

	urlVal := strings.TrimSpace(m.urlInput.Value())
	if urlVal == "" {
		errs["url"] = "URL is required"
	} else {
		parsed, err := url.ParseRequestURI(urlVal)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			errs["url"] = "Must be a valid absolute URL (http:// or https://)"
		} else if parsed.Scheme != "http" && parsed.Scheme != "https" {
			errs["url"] = "URL scheme must be http or https"
		}
	}

	bodyVal := strings.TrimSpace(m.bodyInput.Value())
	if bodyVal != "" {
		if !json.Valid([]byte(bodyVal)) {
			errs["body"] = "Body must be valid JSON"
		}
	}

	for i := range m.headers {
		key := strings.TrimSpace(m.headers[i].Key.Value())
		if key == "" {
			errs["header"] = "Header key is required"
			break
		}
	}

	return errs
}
