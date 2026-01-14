package spec

// Spec represents a normalized OpenAPI specification
type Spec struct {
	Title       string
	Version     string
	Description string
	Operations  []Operation
	GlobalCli   *CliOverrides
}

// Operation represents a single API operation extracted from the spec
type Operation struct {
	Tag         string
	Method      string
	Path        string
	OperationID string
	Summary     string
	Description string
	Params      []Param
	RequestBody *RequestBody
	Responses   []Response
	Cli         *CliOverrides
}

// Param represents a parameter for an operation
type Param struct {
	Name        string
	In          string // path, query, header
	Required    bool
	Type        string
	Format      string
	Default     interface{}
	Min         *float64
	Max         *float64
	Description string
	Cli         *ParamCliOverrides
}

// RequestBody represents a request body for an operation
type RequestBody struct {
	Required     bool
	ContentTypes []string
	Description  string
}

// Response represents a response from an operation
type Response struct {
	StatusCode   string
	Description  string
	ContentTypes []string
}

// CliOverrides represents x-cli overrides at the operation level
type CliOverrides struct {
	Name    string   `json:"name,omitempty" yaml:"name,omitempty"`
	Group   string   `json:"group,omitempty" yaml:"group,omitempty"`
	Aliases []string `json:"aliases,omitempty" yaml:"aliases,omitempty"`
	Hidden  bool     `json:"hidden,omitempty" yaml:"hidden,omitempty"`
}

// ParamCliOverrides represents x-cli overrides at the parameter level
type ParamCliOverrides struct {
	Flag       string `json:"flag,omitempty" yaml:"flag,omitempty"`
	Shorthand  string `json:"shorthand,omitempty" yaml:"shorthand,omitempty"`
	Env        string `json:"env,omitempty" yaml:"env,omitempty"`
	ConfigKey  string `json:"config,omitempty" yaml:"config,omitempty"`
	Positional *bool  `json:"positional,omitempty" yaml:"positional,omitempty"`
}
