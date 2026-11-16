package model

type Policy struct {
	ID                 string             `json:"id"`
	Name               string             `json:"name"`
	Description        string             `json:"description"`
	Version            int                `json:"version"`
	Status             PolicyStatus       `json:"status"`
	CombiningAlgorithm CombiningAlgorithm `json:"combiningAlgorithm"`
	Rules              []Rule             `json:"rules"`

	CreatedAt     string `json:"createdAt"`
	CreatedBy     string `json:"createdBy"`
	UpdatedAt     string `json:"updatedAt"`
	UpdatedBy     string `json:"updatedBy"`
	DeactivatedAt string `json:"deactivatedAt"`
	DeactivatedBy string `json:"deactivatedBy"`
}

type PolicyStatus string

const (
	PolicyStatusActive   PolicyStatus = "active"
	PolicyStatusInactive PolicyStatus = "inactive"
)

type Rule struct {
	ID          string `json:"id"`
	Description string `json:"description"`

	Source    AttributeSource `json:"source"`
	Attribute string          `json:"attribute"`
	Operator  Operator        `json:"operator"`

	Value interface{} `json:"value"`
}

type AttributeSource string

const (
	SourceSubject     AttributeSource = "subject"
	SourceClaim       AttributeSource = "claim"
	SourceResource    AttributeSource = "resource"
	SourceAction      AttributeSource = "action"
	SourceEnvironment AttributeSource = "environment"
)

type Operator string

const (
	OperatorEquals    Operator = "equals"
	OperatorNotEquals Operator = "notEquals"

	OperatorIn    Operator = "in"
	OperatorNotIn Operator = "notIn"
	OperatorAnyOf Operator = "anyOf"
	OperatorAllOf Operator = "allOf"

	OperatorGreaterThan Operator = "greaterThan"
	OperatorLessThan    Operator = "lessThan"
	OperatorBetween     Operator = "between"

	OperatorVerified Operator = "verified"
)

type CombiningAlgorithm string

const (
	CombiningAllOf CombiningAlgorithm = "allOf"
	CombiningAnyOf CombiningAlgorithm = "anyOf"
)
