package tui

// DomainType identifies which domain is active within the Request workspace.
type DomainType int

const (
	DomainRequest DomainType = iota
	DomainPayload
	DomainExec
)

// Domain defines the autonomous behavioural unit. Each Domain encapsulates
// its own behaviour, validation, and action production. A Domain never
// knows about Shell, other Domains, or the surrounding workspace. A Domain
// never owns rendering, colours, typography, or layout.
type Domain interface {
	Actions(m Model) []Action
}

// RequestDomain represents the URL/method configuration domain within the
// Request workspace.
type RequestDomain struct{}

func (RequestDomain) Actions(m Model) []Action {
	return []Action{
		{ActionNextField, ConfigurationCategory, true},
		{ActionSwitchMethod, ConfigurationCategory, true},
	}
}

// PayloadDomain represents the headers/body configuration domain.
type PayloadDomain struct{}

func (PayloadDomain) Actions(m Model) []Action {
	return []Action{
		{ActionNextField, ConfigurationCategory, true},
		{ActionAddHeader, ConfigurationCategory, true},
		{ActionDeleteHeader, ConfigurationCategory, true},
	}
}

// RunDomain represents the concurrency/execution configuration domain.
type RunDomain struct{}

func (RunDomain) Actions(m Model) []Action {
	return []Action{
		{ActionAdjustConcurrency, ConfigurationCategory, true},
	}
}

// domainRegistry maps each DomainType to its Domain instance.
var domainRegistry = map[DomainType]Domain{
	DomainRequest: RequestDomain{},
	DomainPayload: PayloadDomain{},
	DomainExec:    RunDomain{},
}
