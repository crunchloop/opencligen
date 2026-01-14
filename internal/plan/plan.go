package plan

// Plan represents the full command plan for the generated CLI
type Plan struct {
	AppName    string
	ModuleName string
	Groups     []GroupPlan
}

// GroupPlan represents a command group (typically one per tag)
type GroupPlan struct {
	Name        string
	Description string
	Operations  []OpPlan
}

// OpPlan represents a single operation/command plan
type OpPlan struct {
	CommandPath   []string // e.g. ["tasks", "create"]
	Method        string
	Path          string
	OperationID   string
	Summary       string
	Description   string
	Positionals   []ParamPlan
	Flags         []ParamPlan
	HasJSONBody   bool
	IsEventStream bool
	Hidden        bool
	Aliases       []string
}

// ParamPlan represents a parameter plan for a command
type ParamPlan struct {
	Name        string
	FlagName    string
	Shorthand   string
	Type        string
	Format      string
	Required    bool
	Default     interface{}
	Min         *float64
	Max         *float64
	Description string
	EnvVar      string
	ConfigKey   string
	In          string // path, query, header
}
