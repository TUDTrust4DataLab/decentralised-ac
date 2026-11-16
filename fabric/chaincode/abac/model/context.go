package model

type RequestContext struct {
	Subject     SubjectContext     `json:"subject"`
	Resource    ResourceContext    `json:"resource"`
	Action      ActionContext      `json:"action"`
	Environment EnvironmentContext `json:"environment"`
}

type SubjectContext struct {
	ID         string                   `json:"id"`
	Attributes map[string]interface{}   `json:"attributes"`
	Claims     map[string]VerifiedClaim `json:"claims"`
}

type ResourceContext struct {
	ID         string                 `json:"id"`
	Attributes map[string]interface{} `json:"attributes"`
}

type ActionContext struct {
	Name       string                 `json:"name"`
	Attributes map[string]interface{} `json:"attributes"`
}

type EnvironmentContext struct {
	Attributes map[string]interface{} `json:"attributes"`
}

type VerifiedClaim struct {
	Verified bool `json:"verified"`
}
