package tui

import (
	"sort"
	"strings"
	"time"

	"github.com/divijg19/Pulse/internal/model"
)

type FlagSeverity int

const (
	FlagRegression FlagSeverity = iota
	FlagImprovement
	FlagAnomaly
	FlagInfo
)

type ValueDelta[T comparable] struct {
	Old     T
	New     T
	Changed bool
}

type MetadataDelta struct {
	Status  ValueDelta[int]
	Latency ValueDelta[time.Duration]
	URL     ValueDelta[string]
	Error   ValueDelta[string]
}

type HeaderEntry struct {
	Name  string
	Value string
}

type HeaderChange struct {
	Name     string
	OldValue string
	NewValue string
}

type HeaderDelta struct {
	Added   []HeaderEntry
	Removed []HeaderEntry
	Changed []HeaderChange
}

type BodySegmentKind int

const (
	SegmentEqual BodySegmentKind = iota
	SegmentDelete
	SegmentInsert
)

type BodySegment struct {
	Kind BodySegmentKind
	Old  []string
	New  []string
}

type BodyAnalysis struct {
	BaselineSize  int
	CandidateSize int
	ChangedLines  int
	Segments      []BodySegment
}

type Flag struct {
	Severity FlagSeverity
	Field    string
	Message  string
}

type Verdict int

const (
	VerdictEquivalent Verdict = iota
	VerdictImproved
	VerdictRegressed
	VerdictMixed
)

type StatusClass int

const (
	StatusUnknown StatusClass = iota
	StatusInfo
	StatusSuccess
	StatusRedirect
	StatusClientError
	StatusServerError
)

func ClassifyStatus(status int) StatusClass {
	switch {
	case status >= 500:
		return StatusServerError
	case status >= 400:
		return StatusClientError
	case status >= 300:
		return StatusRedirect
	case status >= 200:
		return StatusSuccess
	case status >= 100:
		return StatusInfo
	default:
		return StatusUnknown
	}
}

type ComparisonAnalysis struct {
	Baseline  model.Result
	Candidate model.Result
	Metadata  MetadataDelta
	Headers   HeaderDelta
	Body      BodyAnalysis
	Flags     []Flag
	Verdict   Verdict
}

func AnalyzeComparison(baseline, candidate model.Result) ComparisonAnalysis {
	meta := analyzeMetadata(baseline, candidate)
	hdrs := analyzeHeaders(baseline, candidate)
	body := analyzeBody(baseline, candidate)
	flags := generateFlags(meta, hdrs, body)
	verdict := determineVerdict(flags)

	return ComparisonAnalysis{
		Baseline:  baseline,
		Candidate: candidate,
		Metadata:  meta,
		Headers:   hdrs,
		Body:      body,
		Flags:     flags,
		Verdict:   verdict,
	}
}

func isErrorClass(c StatusClass) bool {
	return c == StatusClientError || c == StatusServerError
}

func determineVerdict(flags []Flag) Verdict {
	var hasRegression, hasImprovement bool
	for _, f := range flags {
		switch f.Severity {
		case FlagRegression:
			hasRegression = true
		case FlagImprovement:
			hasImprovement = true
		}
	}
	switch {
	case hasRegression && hasImprovement:
		return VerdictMixed
	case hasRegression:
		return VerdictRegressed
	case hasImprovement:
		return VerdictImproved
	default:
		return VerdictEquivalent
	}
}

func analyzeMetadata(baseline, candidate model.Result) MetadataDelta {
	return MetadataDelta{
		Status: ValueDelta[int]{
			Old: baseline.Status, New: candidate.Status,
			Changed: baseline.Status != candidate.Status,
		},
		Latency: ValueDelta[time.Duration]{
			Old: baseline.Latency, New: candidate.Latency,
			Changed: baseline.Latency != candidate.Latency,
		},
		URL: ValueDelta[string]{
			Old: baseline.RequestURL, New: candidate.RequestURL,
			Changed: baseline.RequestURL != candidate.RequestURL,
		},
		Error: ValueDelta[string]{
			Old: baseline.Error, New: candidate.Error,
			Changed: baseline.Error != candidate.Error,
		},
	}
}

func analyzeHeaders(baseline, candidate model.Result) HeaderDelta {
	allKeys := make(map[string]bool)
	for k := range baseline.ResponseHeaders {
		allKeys[k] = true
	}
	for k := range candidate.ResponseHeaders {
		allKeys[k] = true
	}

	var added []HeaderEntry
	var removed []HeaderEntry
	var changed []HeaderChange

	for k := range allKeys {
		oldV, inBaseline := baseline.ResponseHeaders[k]
		newV, inCandidate := candidate.ResponseHeaders[k]
		switch {
		case inBaseline && !inCandidate:
			removed = append(removed, HeaderEntry{Name: k, Value: oldV})
		case !inBaseline && inCandidate:
			added = append(added, HeaderEntry{Name: k, Value: newV})
		case oldV != newV:
			changed = append(changed, HeaderChange{Name: k, OldValue: oldV, NewValue: newV})
		}
	}

	sort.Slice(added, func(i, j int) bool { return added[i].Name < added[j].Name })
	sort.Slice(removed, func(i, j int) bool { return removed[i].Name < removed[j].Name })
	sort.Slice(changed, func(i, j int) bool { return changed[i].Name < changed[j].Name })

	return HeaderDelta{Added: added, Removed: removed, Changed: changed}
}

func analyzeBody(baseline, candidate model.Result) BodyAnalysis {
	baseSize := len(baseline.ResponseBody)
	candSize := len(candidate.ResponseBody)

	segments := buildBodySegments(baseline.ResponseBody, candidate.ResponseBody)

	changedLines := 0
	for _, seg := range segments {
		switch seg.Kind {
		case SegmentDelete:
			changedLines += len(seg.Old)
		case SegmentInsert:
			changedLines += len(seg.New)
		}
	}

	return BodyAnalysis{
		BaselineSize:  baseSize,
		CandidateSize: candSize,
		ChangedLines:  changedLines,
		Segments:      segments,
	}
}

func buildBodySegments(baselineBody, candidateBody string) []BodySegment {
	baseLines := strings.Split(baselineBody, "\n")
	candLines := strings.Split(candidateBody, "\n")

	var segments []BodySegment

	i, j := 0, 0
	for i < len(baseLines) && j < len(candLines) {
		if baseLines[i] == candLines[j] {
			var eq []string
			for i < len(baseLines) && j < len(candLines) && baseLines[i] == candLines[j] {
				eq = append(eq, baseLines[i])
				i++
				j++
			}
			segments = append(segments, BodySegment{Kind: SegmentEqual, Old: eq, New: eq})
		} else {
			if j+1 < len(candLines) && baseLines[i] == candLines[j+1] {
				segments = append(segments, BodySegment{Kind: SegmentInsert, Old: nil, New: []string{candLines[j]}})
				j++
			} else if i+1 < len(baseLines) && baseLines[i+1] == candLines[j] {
				segments = append(segments, BodySegment{Kind: SegmentDelete, Old: []string{baseLines[i]}, New: nil})
				i++
			} else {
				segments = append(segments, BodySegment{Kind: SegmentDelete, Old: []string{baseLines[i]}, New: nil})
				segments = append(segments, BodySegment{Kind: SegmentInsert, Old: nil, New: []string{candLines[j]}})
				i++
				j++
			}
		}
	}

	for i < len(baseLines) {
		segments = append(segments, BodySegment{Kind: SegmentDelete, Old: []string{baseLines[i]}, New: nil})
		i++
	}

	for j < len(candLines) {
		segments = append(segments, BodySegment{Kind: SegmentInsert, Old: nil, New: []string{candLines[j]}})
		j++
	}

	return segments
}

func generateFlags(meta MetadataDelta, hdrs HeaderDelta, body BodyAnalysis) []Flag {
	var flags []Flag

	if meta.Status.Changed {
		oldClass := ClassifyStatus(meta.Status.Old)
		newClass := ClassifyStatus(meta.Status.New)
		switch {
		case !isErrorClass(oldClass) && isErrorClass(newClass):
			flags = append(flags, Flag{
				Severity: FlagRegression,
				Field:    "status",
				Message:  "Status regressed from " + itoa(meta.Status.Old) + " to " + itoa(meta.Status.New),
			})
		case isErrorClass(oldClass) && !isErrorClass(newClass):
			flags = append(flags, Flag{
				Severity: FlagImprovement,
				Field:    "status",
				Message:  "Status improved from " + itoa(meta.Status.Old) + " to " + itoa(meta.Status.New),
			})
		default:
			flags = append(flags, Flag{
				Severity: FlagInfo,
				Field:    "status",
				Message:  "Status changed from " + itoa(meta.Status.Old) + " to " + itoa(meta.Status.New),
			})
		}
	}

	if meta.Latency.Changed {
		delta := meta.Latency.New - meta.Latency.Old
		if delta > 0 {
			flags = append(flags, Flag{
				Severity: FlagRegression,
				Field:    "latency",
				Message:  "Latency increased by " + delta.String(),
			})
		} else {
			flags = append(flags, Flag{
				Severity: FlagImprovement,
				Field:    "latency",
				Message:  "Latency decreased by " + (-delta).String(),
			})
		}
	}

	if meta.Error.Changed {
		if meta.Error.New != "" && meta.Error.Old == "" {
			flags = append(flags, Flag{
				Severity: FlagRegression,
				Field:    "error",
				Message:  "New error introduced: " + meta.Error.New,
			})
		} else if meta.Error.Old != "" && meta.Error.New == "" {
			flags = append(flags, Flag{
				Severity: FlagImprovement,
				Field:    "error",
				Message:  "Error resolved",
			})
		} else {
			flags = append(flags, Flag{
				Severity: FlagAnomaly,
				Field:    "error",
				Message:  "Error changed from " + meta.Error.Old + " to " + meta.Error.New,
			})
		}
	}

	if len(hdrs.Added)+len(hdrs.Removed)+len(hdrs.Changed) > 0 {
		count := len(hdrs.Added) + len(hdrs.Removed) + len(hdrs.Changed)
		flags = append(flags, Flag{
			Severity: FlagInfo,
			Field:    "headers",
			Message:  itoa(count) + " header(s) differ",
		})
	}

	if body.BaselineSize != body.CandidateSize || body.ChangedLines > 0 {
		flags = append(flags, Flag{
			Severity: FlagInfo,
			Field:    "body",
			Message:  "Body changed (" + itoa(body.BaselineSize) + " bytes -> " + itoa(body.CandidateSize) + " bytes, " + itoa(body.ChangedLines) + " line(s) differ)",
		})
	}

	return flags
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
