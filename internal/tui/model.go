package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/divijg19/Pulse/internal/engine"
	"github.com/divijg19/Pulse/internal/metrics"
	"github.com/divijg19/Pulse/internal/model"
	"github.com/divijg19/Pulse/internal/runconfig"
	"github.com/divijg19/Pulse/internal/stream"
)

const (
	subfocusKey      = 0
	subfocusValue    = 1
	bodyFocus        = -1
	latencyRingSize  = 200
	defaultBodyWidth = 48
	maxTUIBodyBytes  = 1 << 20
)

type mode int

const (
	modeObserve mode = iota
	modeInspect
)

type dialog int

const (
	dialogNone dialog = iota
	dialogEndpoint
	dialogConcurrency
	dialogPayload
	dialogConfirmQuit
)

type view int

const (
	viewTimeline view = iota
	viewLogs
)

type headerRow struct {
	Key   textinput.Model
	Value textinput.Model
}

type Model struct {
	width  int
	height int

	mode   mode
	dialog dialog
	view   view

	methodIndex int
	urlInput    textinput.Model
	ccInput     textinput.Model
	bodyInput   textarea.Model

	headers        []headerRow
	selectedHead   int
	headerSubfocus int

	results  []model.Result
	selected int

	running   bool
	startedAt time.Time
	elapsed   time.Duration
	cancel    context.CancelFunc
	eventCh   <-chan model.Event
	status    string
	errMsg    string
	capped    bool
	summary   metrics.Summary

	latencyRing [latencyRingSize]time.Duration
	latencyHead int
	latencyLen  int
}

type resultMsg struct {
	Result model.Result
}

type runFinishedMsg struct{}

type eventErrorMsg struct {
	Err string
}

type tickMsg time.Time

type startupMsg struct{}

func NewModel() Model {
	url := textinput.New()
	url.Placeholder = "https://httpbin.org/delay/1"
	url.SetValue("https://httpbin.org/delay/1")
	url.Prompt = ""
	url.CharLimit = 2048

	cc := textinput.New()
	cc.SetValue(strconv.Itoa(runconfig.DefaultConcurrency))
	cc.Prompt = ""
	cc.CharLimit = 3
	cc.Width = 4

	body := textarea.New()
	body.Placeholder = `{"name":"pulse"}`
	body.Prompt = ""
	body.CharLimit = 1 << 20
	body.SetHeight(5)
	body.SetWidth(defaultBodyWidth)

	return Model{
		mode:        modeObserve,
		dialog:      dialogNone,
		view:        viewTimeline,
		methodIndex: 0,
		urlInput:    url,
		ccInput:     cc,
		bodyInput:   body,
		status:      "SYSTEM READY",
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, startupTimeout())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		innerWidth := msg.Width - 4
		leftWidth := max(28, innerWidth/2-2)
		rightWidth := max(28, innerWidth-leftWidth-3)
		m.bodyInput.SetWidth(max(10, rightWidth-6))
		return m, nil
	case startupMsg:
		if m.width == 0 {
			m.width = 80
			m.height = 24
			innerWidth := 80 - 4
			leftWidth := max(28, innerWidth/2-2)
			rightWidth := max(28, innerWidth-leftWidth-3)
			m.bodyInput.SetWidth(max(10, rightWidth-6))
		}
		return m, nil
	case tickMsg:
		if !m.running {
			return m, nil
		}
		m.elapsed = time.Since(m.startedAt)
		return m, tickCmd()
	case resultMsg:
		m.latencyRing[m.latencyHead] = msg.Result.Latency
		m.latencyHead = (m.latencyHead + 1) % latencyRingSize
		if m.latencyLen < latencyRingSize {
			m.latencyLen++
		}
		following := m.isFollowingTail()
		if len(m.results) < 10000 {
			m.results = append(m.results, msg.Result)
		} else {
			m.capped = true
		}
		m.summary = metrics.Compute(m.results, m.elapsed)
		if following {
			m.selected = len(m.results) - 1
		}
		if m.running && m.eventCh != nil {
			return m, waitForEvent(m.eventCh)
		}
		return m, nil
	case eventErrorMsg:
		m.elapsed = time.Since(m.startedAt)
		m.running = false
		m.status = "ERROR: " + msg.Err
		return m, nil
	case runFinishedMsg:
		m.elapsed = time.Since(m.startedAt)
		if m.running {
			m.status = "COMPLETE"
		}
		m.running = false
		if m.cancel != nil {
			m.cancel()
		}
		m.cancel = nil
		m.eventCh = nil
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.MouseMsg:
		return m, nil
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.errMsg = ""

	switch m.mode {
	case modeObserve:
		switch m.dialog {
		case dialogEndpoint:
			return m.handleEndpointKey(msg)
		case dialogConcurrency:
			return m.handleCCKey(msg)
		case dialogPayload:
			return m.handlePayloadKey(msg)
		case dialogConfirmQuit:
			return m.handleConfirmQuitKey(msg)
		case dialogNone:
			return m.handleObserveKey(msg)
		}
	case modeInspect:
		switch m.dialog {
		case dialogConfirmQuit:
			return m.handleConfirmQuitKey(msg)
		case dialogNone:
			return m.handleInspectKey(msg)
		}
	}
	return m, nil
}

func (m Model) handleConfirmQuitKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.dialog = dialogNone
	switch msg.String() {
	case "ctrl+c", "q", "enter":
		if m.running && m.cancel != nil {
			m.cancel()
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleEndpointKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.dialog = dialogNone
		m.urlInput.Blur()
		return m, nil
	case "ctrl+r":
		return m.startRun()
	case "ctrl+x":
		return m.cancelRun(), nil
	}
	var cmd tea.Cmd
	m.urlInput, cmd = m.urlInput.Update(msg)
	return m, cmd
}

func (m Model) handleCCKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.dialog = dialogNone
		m.ccInput.Blur()
		return m, nil
	case "up", "k":
		m.setConcurrency(m.concurrency() + 1)
		return m, nil
	case "down", "j":
		m.setConcurrency(m.concurrency() - 1)
		return m, nil
	case "ctrl+r":
		return m.startRun()
	case "ctrl+x":
		return m.cancelRun(), nil
	}
	var cmd tea.Cmd
	m.ccInput, cmd = m.ccInput.Update(msg)
	return m, cmd
}

func (m Model) handlePayloadKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.dialog = dialogNone
		m.urlInput.Blur()
		m.ccInput.Blur()
		m.bodyInput.Blur()
		for i := range m.headers {
			m.headers[i].Key.Blur()
			m.headers[i].Value.Blur()
		}
		return m, nil
	case "ctrl+r":
		return m.startRun()
	case "ctrl+x":
		return m.cancelRun(), nil
	}

	if m.selectedHead == bodyFocus {
		switch msg.String() {
		case "tab":
			m.selectedHead = 0
			if len(m.headers) == 0 {
				m.headers = append(m.headers, newHeaderRow())
			}
			m.urlInput.Blur()
			m.ccInput.Blur()
			m.bodyInput.Blur()
			for i := range m.headers {
				m.headers[i].Key.Blur()
				m.headers[i].Value.Blur()
			}
			if m.headerSubfocus == subfocusKey {
				m.headers[m.selectedHead].Key.Focus()
			} else {
				m.headers[m.selectedHead].Value.Focus()
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.bodyInput, cmd = m.bodyInput.Update(msg)
		return m, cmd
	}

	if len(m.headers) == 0 {
		m.headers = append(m.headers, newHeaderRow())
	}

	switch msg.String() {
	case "ctrl+n":
		m.headers = append(m.headers, newHeaderRow())
		m.selectedHead = len(m.headers) - 1
		m.headerSubfocus = subfocusKey
		m.urlInput.Blur()
		m.ccInput.Blur()
		m.bodyInput.Blur()
		for i := range m.headers {
			m.headers[i].Key.Blur()
			m.headers[i].Value.Blur()
		}
		m.headers[m.selectedHead].Key.Focus()
		return m, nil
	case "ctrl+d":
		if len(m.headers) > 0 {
			m.headers = append(m.headers[:m.selectedHead], m.headers[m.selectedHead+1:]...)
			if m.selectedHead >= len(m.headers) {
				m.selectedHead = len(m.headers) - 1
			}
			if m.selectedHead < 0 {
				m.selectedHead = 0
			}
		}
		return m, nil
	case "up", "k":
		if m.selectedHead > 0 {
			m.selectedHead--
		}
		return m, nil
	case "down", "j":
		if m.selectedHead < len(m.headers)-1 {
			m.selectedHead++
		}
		return m, nil
	case "left", "h":
		m.headerSubfocus = subfocusKey
		m.urlInput.Blur()
		m.ccInput.Blur()
		m.bodyInput.Blur()
		for i := range m.headers {
			m.headers[i].Key.Blur()
			m.headers[i].Value.Blur()
		}
		m.headers[m.selectedHead].Key.Focus()
		return m, nil
	case "right", "l":
		m.headerSubfocus = subfocusValue
		m.urlInput.Blur()
		m.ccInput.Blur()
		m.bodyInput.Blur()
		for i := range m.headers {
			m.headers[i].Key.Blur()
			m.headers[i].Value.Blur()
		}
		m.headers[m.selectedHead].Value.Focus()
		return m, nil
	case "tab":
		m.selectedHead = bodyFocus
		m.urlInput.Blur()
		m.ccInput.Blur()
		m.bodyInput.Blur()
		for i := range m.headers {
			m.headers[i].Key.Blur()
			m.headers[i].Value.Blur()
		}
		m.bodyInput.Focus()
		return m, nil
	default:
		if m.selectedHead >= 0 && m.selectedHead < len(m.headers) {
			var cmd tea.Cmd
			if m.headerSubfocus == subfocusKey {
				m.headers[m.selectedHead].Key, cmd = m.headers[m.selectedHead].Key.Update(msg)
			} else {
				m.headers[m.selectedHead].Value, cmd = m.headers[m.selectedHead].Value.Update(msg)
			}
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) handleObserveKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
		return m, nil
	case "down", "j":
		if m.selected < len(m.results)-1 {
			m.selected++
		}
		return m, nil
	case "pgup":
		if len(m.results) > 0 {
			pageSize := max(5, m.height/2)
			m.selected = max(0, m.selected-pageSize)
		}
		return m, nil
	case "pgdown":
		if len(m.results) > 0 {
			pageSize := max(5, m.height/2)
			m.selected = min(len(m.results)-1, m.selected+pageSize)
		}
		return m, nil
	case "enter":
		if len(m.results) > 0 {
			m.mode = modeInspect
		}
		return m, nil
	case "[":
		m.view = viewTimeline
		return m, nil
	case "]":
		m.view = viewLogs
		return m, nil
	case "e":
		m.dialog = dialogEndpoint
		m.urlInput.Focus()
		return m, nil
	case "c":
		m.dialog = dialogConcurrency
		m.ccInput.Focus()
		return m, nil
	case "p":
		m.dialog = dialogPayload
		m.selectedHead = 0
		m.headerSubfocus = subfocusKey
		if len(m.headers) == 0 {
			m.headers = append(m.headers, newHeaderRow())
		}
		m.urlInput.Blur()
		m.ccInput.Blur()
		m.bodyInput.Blur()
		for i := range m.headers {
			m.headers[i].Key.Blur()
			m.headers[i].Value.Blur()
		}
		m.headers[m.selectedHead].Key.Focus()
		return m, nil
	case "ctrl+r":
		return m.startRun()
	case "ctrl+x":
		return m.cancelRun(), nil
	case "q", "ctrl+c":
		if m.running {
			m.dialog = dialogConfirmQuit
			return m, nil
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleInspectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeObserve
		return m, nil
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
		return m, nil
	case "down", "j":
		if m.selected < len(m.results)-1 {
			m.selected++
		}
		return m, nil
	case "q", "ctrl+c":
		m.dialog = dialogConfirmQuit
		return m, nil
	}
	return m, nil
}

func (m Model) startRun() (Model, tea.Cmd) {
	if m.running {
		return m, nil
	}

	req := model.RunRequest{
		URL:         m.urlInput.Value(),
		Method:      runconfig.AllowedMethods()[m.methodIndex],
		Headers:     m.headerMap(),
		Body:        m.bodyInput.Value(),
		Concurrency: m.concurrency(),
	}

	validated, err := runconfig.Validate(req)
	if err != nil {
		m.errMsg = strings.ToUpper(err.Error())
		m.status = strings.ToUpper(err.Error())
		return m, nil
	}

	if len(validated.Body) > maxTUIBodyBytes {
		m.errMsg = "BODY TOO LARGE (MAX 1MB)"
		m.status = "BODY TOO LARGE (MAX 1MB)"
		return m, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	hub := stream.NewHub()
	eventCh := make(chan model.Event, runconfig.MaxConcurrency)
	hub.Add(eventCh)

	m.mode = modeObserve
	m.dialog = dialogNone
	m.running = true
	m.cancel = cancel
	m.eventCh = eventCh
	m.startedAt = time.Now()
	m.elapsed = 0
	m.results = nil
	m.selected = 0
	m.capped = false
	m.errMsg = ""
	m.status = "RUNNING"
	m.summary = metrics.Summary{}
	m.latencyHead = 0
	m.latencyLen = 0

	go func() {
		defer hub.Remove(eventCh)
		engine.ExecuteConcurrent(ctx, validated, hub)
	}()

	return m, tea.Batch(waitForEvent(eventCh), tickCmd())
}

func (m Model) isFollowingTail() bool {
	return len(m.results) == 0 || m.selected >= len(m.results)-1
}

func (m Model) cancelRun() Model {
	if m.running && m.cancel != nil {
		m.cancel()
		m.status = "CANCELLED"
		m.running = false
	}
	return m
}

func waitForEvent(ch <-chan model.Event) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-ch
		if !ok {
			return runFinishedMsg{}
		}
		result, ok := event.Data.(model.Result)
		if !ok {
			return eventErrorMsg{Err: fmt.Sprintf("unexpected event: %s", event.Type)}
		}
		return resultMsg{Result: result}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func startupTimeout() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return startupMsg{}
	})
}

func (m Model) concurrency() int {
	value, err := strconv.Atoi(strings.TrimSpace(m.ccInput.Value()))
	if err != nil {
		return runconfig.DefaultConcurrency
	}
	return runconfig.ClampConcurrency(value)
}

func (m *Model) setConcurrency(value int) {
	m.ccInput.SetValue(strconv.Itoa(runconfig.ClampConcurrency(value)))
}

func (m Model) headerMap() map[string]string {
	headers := map[string]string{}
	for _, row := range m.headers {
		key := strings.TrimSpace(row.Key.Value())
		if key == "" {
			continue
		}
		headers[key] = row.Value.Value()
	}
	return headers
}

func newHeaderRow() headerRow {
	key := textinput.New()
	key.Prompt = ""
	key.Placeholder = "Header"
	key.CharLimit = 256
	key.Width = 20

	value := textinput.New()
	value.Prompt = ""
	value.Placeholder = "Value"
	value.CharLimit = 2048
	value.Width = 28

	return headerRow{Key: key, Value: value}
}

func formatDuration(duration time.Duration) string {
	if duration <= 0 {
		return "0.00s"
	}
	secs := duration.Seconds()
	if secs >= 60 {
		mins := int(secs) / 60
		left := secs - float64(mins*60)
		return fmt.Sprintf("%dm %.0fs", mins, left)
	}
	return fmt.Sprintf("%.2fs", secs)
}
