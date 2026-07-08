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
	subfocusKey     = 0
	subfocusValue   = 1
	bodyFocus       = -1
	maxTUIBodyBytes = 1 << 20

	reqFieldMethod = 0
	reqFieldURL    = 1

	defaultURL = "https://httpbin.org/delay/1"

	zoneWhatHappened = 0
	zoneWhy          = 1
	zoneBody         = 2
)

type mode int

const (
	modeObserve mode = iota
	modeInspect
	modeCompare
)

type dialog int

const (
	dialogNone dialog = iota
	dialogRequest
	dialogConfirmQuit
)

type headerRow struct {
	Key   textinput.Model
	Value textinput.Model
}

type Model struct {
	shell     Shell
	workspace Workspace

	payloadGeometry payloadGeometry

	activeDomain     DomainType
	methodIndex      int
	requestField     int
	urlInput         textinput.Model
	concurrencyInput textinput.Model
	bodyInput        textarea.Model

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
	summary   metrics.Summary

	inspectZone       int
	inspectBodyOffset int
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
	url.Placeholder = defaultURL
	url.SetValue(defaultURL)
	url.Prompt = ""
	url.CharLimit = 2048

	cc := textinput.New()
	cc.SetValue(strconv.Itoa(runconfig.DefaultConcurrency))
	cc.Prompt = ""
	cc.CharLimit = 3
	cc.Width = 4

	geo := calculatePayloadGeometry(76)

	body := textarea.New()
	body.Placeholder = `{"name":"pulse"}`
	body.Prompt = ""
	body.CharLimit = 1 << 20
	body.SetHeight(geo.BodyHeight)
	body.SetWidth(geo.BodyWidth)

	return Model{
		shell:            NewShell(),
		workspace:        NewWorkspace(),
		payloadGeometry:  geo,
		activeDomain:     DomainRequest,
		methodIndex:      0,
		requestField:     reqFieldURL,
		urlInput:         url,
		concurrencyInput: cc,
		bodyInput:        body,
		status:           "SYSTEM READY",
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, startupTimeout())
}

func (m *Model) syncPayloadGeometry(contentWidth int) {
	geo := calculatePayloadGeometry(contentWidth)
	m.bodyInput.SetWidth(geo.BodyWidth)
	for i := range m.headers {
		m.headers[i].Key.Width = geo.KeyWidth
		m.headers[i].Value.Width = geo.ValueWidth
	}
	m.payloadGeometry = geo
}

func workspaceContentWidth(shellWidth, shellHeight int) int {
	layout := computeShellLayout(shellWidth, shellHeight)
	ws := layout.Workspace
	ws.Border = BorderFull
	ws.PaddingX = 1
	return ws.ContentRegion().Width
}

// payloadContentWidth returns the width available to the payload body editor
// at the given shell dimensions, accounting for context panel when visible.
func payloadContentWidth(shellWidth, shellHeight int) int {
	cw := workspaceContentWidth(shellWidth, shellHeight)
	if shellWidth < contextThreshold {
		return cw
	}
	ctxWidth := cw / 3
	if ctxWidth < contextMinWidth {
		ctxWidth = contextMinWidth
	}
	primaryWidth := cw - ctxWidth - 1
	if primaryWidth < 40 {
		return cw
	}
	return primaryWidth
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.shell.Resize(msg.Width, msg.Height)
		m.syncPayloadGeometry(payloadContentWidth(msg.Width, msg.Height))
		return m, nil
	case startupMsg:
		w, _ := m.shell.Dimensions()
		if w == 0 {
			m.shell.Resize(80, 24)
			m.syncPayloadGeometry(payloadContentWidth(80, 24))
		}
		return m, nil
	case tickMsg:
		if !m.running {
			return m, nil
		}
		m.elapsed = time.Since(m.startedAt)
		return m, tickCmd()
	case resultMsg:
		following := m.isFollowingTail()
		if len(m.results) < 10000 {
			m.results = append(m.results, msg.Result)
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
		m.errMsg = "ERROR: " + strings.ToUpper(msg.Err)
		m.status = "ERROR: " + strings.ToUpper(msg.Err)
		return m, nil
	case runFinishedMsg:
		m.elapsed = time.Since(m.startedAt)
		if m.running {
			m.status = "COMPLETE"
		}
		m.running = false
		m.errMsg = ""
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
	case error:
		m.errMsg = strings.ToUpper(msg.Error())
		m.status = "ERROR: " + strings.ToUpper(msg.Error())
		return m, nil
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.errMsg = ""

	switch m.workspace.mode {
	case modeCompare:
		return m.handleCompareKey(msg)
	case modeObserve:
		switch m.workspace.dialog {
		case dialogRequest:
			return m.handleRequestKey(msg)
		case dialogConfirmQuit:
			return m.handleConfirmQuitKey(msg)
		case dialogNone:
			return m.handleObserveKey(msg)
		}
	case modeInspect:
		switch m.workspace.dialog {
		case dialogConfirmQuit:
			return m.handleConfirmQuitKey(msg)
		case dialogNone:
			return m.handleInspectKey(msg)
		}
	}
	return m, nil
}

func (m Model) handleRequestKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.workspace.dialog = dialogNone
		m.blurAll()
		return m, nil
	case "tab":
		return m.advanceDomain(true)
	case "shift+tab":
		return m.advanceDomain(false)
	case "ctrl+r":
		return m.startRun()
	case "ctrl+x":
		return m.cancelRun(), nil
	}

	switch m.activeDomain {
	case DomainRequest:
		return m.handleRequestDomainKey(msg)
	case DomainPayload:
		return m.handlePayloadDomainKey(msg)
	case DomainExec:
		return m.handleExecDomainKey(msg)
	}
	return m, nil
}

func (m Model) advanceDomain(forward bool) (tea.Model, tea.Cmd) {
	if forward {
		return m.advanceDomainForward()
	}
	return m.advanceDomainBackward()
}

func (m Model) advanceDomainForward() (tea.Model, tea.Cmd) {
	switch m.activeDomain {
	case DomainRequest:
		if m.requestField == reqFieldURL {
			m.activeDomain = DomainPayload
			m.selectedHead = 0
			m.headerSubfocus = subfocusKey
			if len(m.headers) == 0 {
				m.headers = append(m.headers, newHeaderRow())
			}
			m.focusPayloadKey()
		} else {
			m.focusURL()
		}
	case DomainPayload:
		if m.selectedHead == bodyFocus {
			m.focusConcurrency()
		} else if m.headerSubfocus == subfocusKey {
			m.headerSubfocus = subfocusValue
			m.focusPayloadValue()
		} else if m.selectedHead < len(m.headers)-1 {
			m.selectedHead++
			m.headerSubfocus = subfocusKey
			m.focusPayloadKey()
		} else {
			m.selectedHead = bodyFocus
			m.focusPayloadBody()
		}
	case DomainExec:
		m.activeDomain = DomainRequest
		m.focusMethod()
	}
	return m, nil
}

func (m Model) advanceDomainBackward() (tea.Model, tea.Cmd) {
	switch m.activeDomain {
	case DomainRequest:
		if m.requestField == reqFieldMethod {
			m.focusConcurrency()
		} else {
			m.focusMethod()
		}
	case DomainPayload:
		if m.selectedHead == bodyFocus {
			if len(m.headers) > 0 {
				m.selectedHead = len(m.headers) - 1
				m.headerSubfocus = subfocusValue
				m.focusPayloadValue()
			} else {
				m.activeDomain = DomainRequest
				m.focusURL()
			}
		} else if m.headerSubfocus == subfocusValue {
			m.headerSubfocus = subfocusKey
			m.focusPayloadKey()
		} else if m.selectedHead > 0 {
			m.selectedHead--
			m.headerSubfocus = subfocusValue
			m.focusPayloadValue()
		} else {
			m.activeDomain = DomainRequest
			m.focusURL()
		}
	case DomainExec:
		m.activeDomain = DomainPayload
		m.selectedHead = bodyFocus
		m.focusPayloadBody()
	}
	return m, nil
}

func (m Model) handleRequestDomainKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up":
		if m.requestField == reqFieldURL {
			m.focusMethod()
		}
		return m, nil
	case "down":
		if m.requestField == reqFieldMethod {
			m.focusURL()
		} else {
			m.activeDomain = DomainPayload
			m.selectedHead = 0
			m.headerSubfocus = subfocusKey
			if len(m.headers) == 0 {
				m.headers = append(m.headers, newHeaderRow())
			}
			m.focusPayloadKey()
		}
		return m, nil
	case "left":
		if m.requestField == reqFieldMethod && m.methodIndex > 0 {
			m.methodIndex--
		}
		if m.requestField == reqFieldMethod {
			return m, nil
		}
	case "right":
		if m.requestField == reqFieldMethod {
			methods := runconfig.AllowedMethods()
			if m.methodIndex < len(methods)-1 {
				m.methodIndex++
			}
			return m, nil
		}
	}
	var cmd tea.Cmd
	if m.urlInput.Focused() {
		m.urlInput, cmd = m.urlInput.Update(msg)
	}
	return m, cmd
}

func (m Model) handlePayloadDomainKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.selectedHead == bodyFocus {
		return m.handlePayloadBodyKey(msg)
	}
	return m.handlePayloadHeaderKey(msg)
}

func (m Model) handlePayloadBodyKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.bodyInput, cmd = m.bodyInput.Update(msg)
	return m, cmd
}

func (m Model) handlePayloadHeaderKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.headers) == 0 {
		m.headers = append(m.headers, newHeaderRow())
	}

	switch msg.String() {
	case "ctrl+n":
		m.headers = append(m.headers, newHeaderRow())
		m.selectedHead = len(m.headers) - 1
		m.headerSubfocus = subfocusKey
		m.focusPayloadKey()
		return m, nil
	case "ctrl+d":
		if len(m.headers) > 0 && m.selectedHead >= 0 {
			m.headers = append(m.headers[:m.selectedHead], m.headers[m.selectedHead+1:]...)
			if len(m.headers) == 0 {
				m.selectedHead = bodyFocus
				m.focusPayloadBody()
			} else {
				if m.selectedHead >= len(m.headers) {
					m.selectedHead = len(m.headers) - 1
				}
				if m.headerSubfocus == subfocusKey {
					m.focusPayloadKey()
				} else {
					m.focusPayloadValue()
				}
			}
		}
		return m, nil
	case "up":
		if m.selectedHead == 0 {
			m.activeDomain = DomainRequest
			m.focusURL()
		} else {
			m.selectedHead--
		}
		return m, nil
	case "down":
		if m.selectedHead == len(m.headers)-1 {
			m.selectedHead = bodyFocus
			m.focusPayloadBody()
		} else {
			m.selectedHead++
		}
		return m, nil
	case "left":
		if m.headerSubfocus == subfocusValue {
			m.headerSubfocus = subfocusKey
			m.focusPayloadKey()
		}
		return m, nil
	case "right":
		if m.headerSubfocus == subfocusKey {
			m.headerSubfocus = subfocusValue
			m.focusPayloadValue()
		}
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

func (m Model) handleExecDomainKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up":
		m.setConcurrency(m.concurrency() + 1)
		return m, nil
	case "down":
		m.setConcurrency(m.concurrency() - 1)
		return m, nil
	}
	var cmd tea.Cmd
	if m.concurrencyInput.Focused() {
		m.concurrencyInput, cmd = m.concurrencyInput.Update(msg)
	}
	return m, cmd
}

func (m Model) handleConfirmQuitKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.workspace.dialog = dialogNone
	switch msg.String() {
	case "ctrl+c", "q", "enter":
		if m.running && m.cancel != nil {
			m.cancel()
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) effectiveMethod(result model.Result) string {
	if result.RequestMethod != "" {
		return result.RequestMethod
	}
	return runconfig.AllowedMethods()[m.methodIndex]
}

func (m Model) effectiveURL(result model.Result) string {
	if result.RequestURL != "" {
		return result.RequestURL
	}
	return m.urlInput.Value()
}

func (m *Model) blurAll() {
	m.urlInput.Blur()
	m.concurrencyInput.Blur()
	m.bodyInput.Blur()
	for i := range m.headers {
		m.headers[i].Key.Blur()
		m.headers[i].Value.Blur()
	}
}

func (m *Model) focusMethod() {
	m.requestField = reqFieldMethod
	m.blurAll()
}

func (m *Model) focusURL() {
	m.requestField = reqFieldURL
	m.blurAll()
	m.urlInput.Focus()
}

func (m *Model) focusConcurrency() {
	m.activeDomain = DomainExec
	m.blurAll()
	m.concurrencyInput.Focus()
}

func (m *Model) focusPayloadKey() {
	m.blurAll()
	if m.selectedHead < 0 || m.selectedHead >= len(m.headers) {
		return
	}
	m.headers[m.selectedHead].Key.Focus()
}

func (m *Model) focusPayloadValue() {
	m.blurAll()
	if m.selectedHead < 0 || m.selectedHead >= len(m.headers) {
		return
	}
	m.headers[m.selectedHead].Value.Focus()
}

func (m *Model) focusPayloadBody() {
	m.blurAll()
	m.bodyInput.Focus()
}

// computeComparisonAnalysis runs the Comparison Engine using the session's
// baseline and candidate results. It uses PinnedBaseline as the baseline when
// available, falling back to m.results[BaselineIndex].
func (m Model) computeComparisonAnalysis() *ComparisonAnalysis {
	if m.workspace.compare.Session.State != SessionComparing {
		return nil
	}
	if m.workspace.compare.Session.CandidateIndex < 0 || m.workspace.compare.Session.CandidateIndex >= len(m.results) {
		return nil
	}
	candidate := m.results[m.workspace.compare.Session.CandidateIndex]

	var baseline model.Result
	if m.workspace.compare.PinnedBaseline != nil {
		baseline = *m.workspace.compare.PinnedBaseline
	} else if m.workspace.compare.Session.BaselineIndex >= 0 && m.workspace.compare.Session.BaselineIndex < len(m.results) {
		baseline = m.results[m.workspace.compare.Session.BaselineIndex]
	} else {
		return nil
	}

	analysis := AnalyzeComparison(baseline, candidate)
	return &analysis
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
			_, h := m.shell.Dimensions()
			pageSize := max(5, h/2)
			m.selected = max(0, m.selected-pageSize)
		}
		return m, nil
	case "pgdown":
		if len(m.results) > 0 {
			_, h := m.shell.Dimensions()
			pageSize := max(5, h/2)
			m.selected = min(len(m.results)-1, m.selected+pageSize)
		}
		return m, nil
	case "enter":
		if len(m.results) > 0 {
			m.workspace.mode = modeInspect
			m.inspectBodyOffset = 0
			m.inspectZone = zoneWhatHappened
		}
		return m, nil
	case "tab", "shift+tab", "left", "right":
		return m, nil
	case "[":
		m.workspace.view = TimelineView
		return m, nil
	case "]":
		m.workspace.view = LogsView
		return m, nil
	case "c":
		if len(m.results) == 0 {
			return m, nil
		}
		switch m.workspace.compare.Session.State {
		case SessionIdle:
			m.workspace.compare.Session.BaselineIndex = m.selected
			m.workspace.compare.Session.State = SessionBaselineMarked
			m.status = "Baseline marked"
			m.workspace.mode = modeObserve
		case SessionBaselineMarked:
			if m.selected != m.workspace.compare.Session.BaselineIndex {
				m.workspace.compare.Session.CandidateIndex = m.selected
				m.workspace.compare.Session.State = SessionComparing
				m.workspace.compare.Session.Analysis = m.computeComparisonAnalysis()
				m.workspace.mode = modeCompare
				m.inspectZone = zoneWhatHappened
				m.inspectBodyOffset = 0
			}
		case SessionComparing:
			m.workspace.compare.Session.CandidateIndex = m.selected
			m.inspectZone = zoneWhatHappened
			m.inspectBodyOffset = 0
			m.workspace.compare.Session.Analysis = m.computeComparisonAnalysis()
		}
		return m, nil
	case "e":
		m.workspace.dialog = dialogRequest
		m.activeDomain = DomainRequest
		m.focusURL()
		w, _ := m.shell.Dimensions()
		m.syncPayloadGeometry(payloadContentWidth(w, 24))
		return m, nil
	case "ctrl+r":
		return m.startRun()
	case "ctrl+x":
		return m.cancelRun(), nil
	case "q", "ctrl+c":
		m.workspace.dialog = dialogConfirmQuit
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
		errStr := strings.ToUpper(err.Error())
		m.errMsg = errStr
		m.status = errStr
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

	m.workspace.mode = modeObserve
	m.workspace.dialog = dialogNone
	if m.workspace.compare.Session.State >= SessionBaselineMarked && len(m.results) > m.workspace.compare.Session.BaselineIndex {
		result := m.results[m.workspace.compare.Session.BaselineIndex]
		m.workspace.compare.PinnedBaseline = &result
	}
	m.workspace.compare.Session = ComparisonSession{BaselineIndex: -1, CandidateIndex: -1, State: SessionIdle, Analysis: nil}
	m.running = true
	m.cancel = cancel
	m.eventCh = eventCh
	m.startedAt = time.Now()
	m.elapsed = 0
	m.results = nil
	m.selected = 0
	m.errMsg = ""
	m.status = "RUNNING"
	m.summary = metrics.Summary{}
	m.inspectBodyOffset = 0
	m.inspectZone = zoneWhatHappened

	go func() {
		defer hub.Remove(eventCh)
		defer func() {
			if r := recover(); r != nil {
				select {
				case eventCh <- model.Event{
					Type: fmt.Sprintf("engine panic: %v", r),
				}:
				default:
				}
			}
		}()
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
		m.blurAll()
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
	value, err := strconv.Atoi(strings.TrimSpace(m.concurrencyInput.Value()))
	if err != nil {
		return runconfig.DefaultConcurrency
	}
	return value
}

func (m *Model) setConcurrency(value int) {
	m.concurrencyInput.SetValue(strconv.Itoa(value))
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
	geo := calculatePayloadGeometry(76)

	key := textinput.New()
	key.Prompt = ""
	key.Placeholder = "Header"
	key.CharLimit = 256
	key.Width = geo.KeyWidth

	value := textinput.New()
	value.Prompt = ""
	value.Placeholder = "Value"
	value.CharLimit = 2048
	value.Width = geo.ValueWidth

	return headerRow{Key: key, Value: value}
}
