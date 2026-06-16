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

type focusTarget int

const (
	focusMethod focusTarget = iota
	focusURL
	focusConcurrency
	focusPayload
	focusHeaders
	focusBody
	focusResults
)

const (
	subfocusKey      = 0
	subfocusValue    = 1
	latencyRingSize  = 40
	defaultBodyWidth = 48
	maxTUIBodyBytes  = 1 << 20
)

type resultTab int

const (
	tabTimeline resultTab = iota
	tabLogs
)

type headerRow struct {
	Key   textinput.Model
	Value textinput.Model
}

type Model struct {
	width  int
	height int

	focus       focusTarget
	methodIndex int
	urlInput    textinput.Model
	ccInput     textinput.Model
	bodyInput   textarea.Model

	showPayload    bool
	headers        []headerRow
	selectedHead   int
	headerSubfocus int

	activeTab  resultTab
	results    []model.Result
	selected   int
	inspector  bool
	dotGlow    bool
	autoScroll bool

	running     bool
	startedAt   time.Time
	elapsed     time.Duration
	cancel      context.CancelFunc
	eventCh     <-chan model.Event
	status      string
	summary     metrics.Summary
	latencyRing [latencyRingSize]time.Duration
	latencyHead int
	latencyLen  int
}

type resultMsg struct {
	Result model.Result
}

type runFinishedMsg struct{}

type tickMsg time.Time

func NewModel() Model {
	url := textinput.New()
	url.Placeholder = "https://httpbin.org/delay/1"
	url.SetValue("https://httpbin.org/delay/1")
	url.Prompt = ""
	url.CharLimit = 2048
	url.Focus()

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
		focus:       focusURL,
		methodIndex: 0,
		urlInput:    url,
		ccInput:     cc,
		bodyInput:   body,
		activeTab:   tabTimeline,
		autoScroll:  true,
		status:      "SYSTEM READY",
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tickMsg:
		if !m.running {
			return m, nil
		}
		m.elapsed = time.Since(m.startedAt)
		m.dotGlow = !m.dotGlow
		return m, tickCmd()
	case resultMsg:
		m.latencyRing[m.latencyHead] = msg.Result.Latency
		m.latencyHead = (m.latencyHead + 1) % latencyRingSize
		if m.latencyLen < latencyRingSize {
			m.latencyLen++
		}
		if len(m.results) < 10000 {
			m.results = append(m.results, msg.Result)
		}
		m.summary = metrics.Compute(m.results, m.elapsed)
		if m.autoScroll && m.selected >= len(m.results)-1 {
			m.selected = len(m.results) - 1
		}
		if m.running && m.eventCh != nil {
			return m, waitForEvent(m.eventCh)
		}
		return m, nil
	case runFinishedMsg:
		m.elapsed = time.Since(m.startedAt)
		m.running = false
		if m.cancel != nil {
			m.cancel()
		}
		m.cancel = nil
		m.eventCh = nil
		if m.status == "RUNNING" {
			m.status = "COMPLETE"
		}
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m.updateFocusedInput(msg)
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.running && m.cancel != nil {
			m.cancel()
		}
		return m, tea.Quit
	case "ctrl+r":
		return m.startRun()
	case "ctrl+x":
		return m.cancelRun(), nil
	case "tab":
		return m.moveFocus(1), nil
	case "shift+tab":
		return m.moveFocus(-1), nil
	case "[":
		m.activeTab = tabTimeline
		return m, nil
	case "]":
		m.activeTab = tabLogs
		return m, nil
	case "ctrl+a":
		m.autoScroll = !m.autoScroll
		return m, nil
	case "esc":
		if m.inspector {
			m.inspector = false
			return m, nil
		}
		if m.focus == focusBody {
			m.focus = focusHeaders
			return m.syncFocus(), nil
		}
		return m, nil
	}

	switch m.focus {
	case focusMethod:
		switch msg.String() {
		case "left", "h", "up", "k":
			m.methodIndex = (m.methodIndex + len(runconfig.AllowedMethods()) - 1) % len(runconfig.AllowedMethods())
			return m, nil
		case "right", "l", "down", "j", "enter":
			m.methodIndex = (m.methodIndex + 1) % len(runconfig.AllowedMethods())
			return m, nil
		}
	case focusConcurrency:
		switch msg.String() {
		case "left", "h", "down", "j":
			m.setConcurrency(m.concurrency() - 1)
			return m, nil
		case "right", "l", "up", "k":
			m.setConcurrency(m.concurrency() + 1)
			return m, nil
		}
	case focusPayload:
		if msg.String() == "enter" || msg.String() == " " {
			m.showPayload = !m.showPayload
			if m.showPayload && len(m.headers) == 0 {
				m.headers = append(m.headers, newHeaderRow())
			}
			return m, nil
		}
	case focusHeaders:
		return m.handleHeaderKey(msg)
	case focusResults:
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
		case "enter":
			if len(m.results) > 0 {
				m.inspector = true
			}
			return m, nil
		}
	}

	return m.updateFocusedInput(msg)
}

func (m Model) handleHeaderKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if !m.showPayload {
		return m, nil
	}
	if len(m.headers) == 0 {
		m.headers = append(m.headers, newHeaderRow())
	}

	switch msg.String() {
	case "ctrl+n":
		m.headers = append(m.headers, newHeaderRow())
		m.selectedHead = len(m.headers) - 1
		m.headerSubfocus = subfocusKey
		return m.syncFocus(), nil
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
		return m.syncFocus(), nil
	case "up", "k":
		if m.selectedHead > 0 {
			m.selectedHead--
		}
		return m.syncFocus(), nil
	case "down", "j":
		if m.selectedHead < len(m.headers)-1 {
			m.selectedHead++
		}
		return m.syncFocus(), nil
	case "left", "h":
		m.headerSubfocus = subfocusKey
		return m.syncFocus(), nil
	case "right", "l":
		m.headerSubfocus = subfocusValue
		return m.syncFocus(), nil
	}

	return m.updateFocusedInput(msg)
}

func (m Model) updateFocusedInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focus {
	case focusURL:
		m.urlInput, cmd = m.urlInput.Update(msg)
	case focusConcurrency:
		m.ccInput, cmd = m.ccInput.Update(msg)
	case focusHeaders:
		if len(m.headers) == 0 {
			break
		}
		if m.headerSubfocus == 0 {
			m.headers[m.selectedHead].Key, cmd = m.headers[m.selectedHead].Key.Update(msg)
		} else {
			m.headers[m.selectedHead].Value, cmd = m.headers[m.selectedHead].Value.Update(msg)
		}
	case focusBody:
		m.bodyInput, cmd = m.bodyInput.Update(msg)
	}
	return m, cmd
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
		m.status = strings.ToUpper(err.Error())
		return m, nil
	}

	if len(validated.Body) > maxTUIBodyBytes {
		m.status = "BODY TOO LARGE (MAX 1MB)"
		return m, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	hub := stream.NewHub()
	eventCh := make(chan model.Event, runconfig.MaxConcurrency)
	hub.Add(eventCh)

	m.running = true
	m.cancel = cancel
	m.eventCh = eventCh
	m.startedAt = time.Now()
	m.elapsed = 0
	m.results = nil
	m.selected = 0
	m.inspector = false
	m.status = "RUNNING"
	m.autoScroll = true
	m.summary = metrics.Summary{}
	m.latencyHead = 0
	m.latencyLen = 0

	go func() {
		defer hub.Remove(eventCh)
		engine.ExecuteConcurrent(ctx, validated, hub)
	}()

	return m, tea.Batch(waitForEvent(eventCh), tickCmd())
}

func (m Model) cancelRun() Model {
	if m.running && m.cancel != nil {
		m.cancel()
		m.status = "CANCELLED"
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
			return runFinishedMsg{}
		}
		return resultMsg{Result: result}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) moveFocus(delta int) Model {
	targets := []focusTarget{focusMethod, focusURL, focusConcurrency, focusPayload, focusResults}
	if m.showPayload {
		targets = []focusTarget{focusMethod, focusURL, focusConcurrency, focusPayload, focusHeaders, focusBody, focusResults}
	}

	current := 0
	for i, target := range targets {
		if target == m.focus {
			current = i
			break
		}
	}
	next := (current + delta + len(targets)) % len(targets)
	m.focus = targets[next]
	return m.syncFocus()
}

func (m Model) syncFocus() Model {
	m.urlInput.Blur()
	m.ccInput.Blur()
	m.bodyInput.Blur()
	for i := range m.headers {
		m.headers[i].Key.Blur()
		m.headers[i].Value.Blur()
	}

	switch m.focus {
	case focusURL:
		m.urlInput.Focus()
	case focusConcurrency:
		m.ccInput.Focus()
	case focusHeaders:
		if len(m.headers) > 0 {
			if m.selectedHead >= len(m.headers) {
				m.selectedHead = len(m.headers) - 1
			}
			if m.headerSubfocus == subfocusKey {
				m.headers[m.selectedHead].Key.Focus()
			} else {
				m.headers[m.selectedHead].Value.Focus()
			}
		}
	case focusBody:
		m.bodyInput.Focus()
	}
	return m
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
	return fmt.Sprintf("%.2fs", duration.Seconds())
}
