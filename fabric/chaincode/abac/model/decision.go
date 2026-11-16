package model

type Decision string

const (
	DecisionPermit Decision = "permit"
	DecisionDeny   Decision = "deny"
)

type AccessDecision struct {
	Decision      Decision         `json:"decision"`
	PolicyID      string           `json:"policyId"`
	PolicyVersion int              `json:"policyVersion"`
	ResourceID    string           `json:"resourceId"`
	Evaluations   []RuleEvaluation `json:"evaluations"`
}

type RuleEvaluation struct {
	RuleID string `json:"ruleId"`
	Result bool   `json:"result"`
}
