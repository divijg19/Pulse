package tui

type DomainType int

const (
	DomainRequest DomainType = iota
	DomainPayload
	DomainExec
)

type Domain interface {
	Actions(m Model) []Action
}

type RequestDomain struct{}

func (RequestDomain) Actions(m Model) []Action {
	return []Action{
		{ActionNextField, ConfigurationCategory, true},
		{ActionSwitchMethod, ConfigurationCategory, true},
	}
}

type PayloadDomain struct{}

func (PayloadDomain) Actions(m Model) []Action {
	return []Action{
		{ActionNextField, ConfigurationCategory, true},
		{ActionAddHeader, ConfigurationCategory, true},
		{ActionDeleteHeader, ConfigurationCategory, true},
	}
}

type ExecDomain struct{}

func (ExecDomain) Actions(m Model) []Action {
	return []Action{
		{ActionAdjustConcurrency, ConfigurationCategory, true},
	}
}

var domainRegistry = map[DomainType]Domain{
	DomainRequest: RequestDomain{},
	DomainPayload: PayloadDomain{},
	DomainExec:    ExecDomain{},
}
