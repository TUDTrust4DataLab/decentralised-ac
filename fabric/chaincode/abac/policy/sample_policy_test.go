package policy

import (
	"testing"

	"cipher/fabric/chaincode/abac/model"
)

func samplePolicy() model.Policy {
	return model.Policy{
		ID:                 "sample-policy",
		Name:               "Sample ABAC Policy",
		Description:        "Fabric equivalent of the Solidity SamplePolicy",
		Version:            1,
		Status:             model.PolicyStatusActive,
		CombiningAlgorithm: model.CombiningAllOf,

		Rules: []model.Rule{
			{
				ID:          "allowed-role",
				Description: "Subject must have an allowed role",
				Source:      model.SourceSubject,
				Attribute:   "role",
				Operator:    model.OperatorIn,
				Value: []string{
					"analyst",
					"admin",
				},
			},
			{
				ID:          "allowed-group",
				Description: "Subject must belong to an allowed group",
				Source:      model.SourceSubject,
				Attribute:   "groups",
				Operator:    model.OperatorAnyOf,
				Value: []string{
					"cipher",
					"researchers",
				},
			},
			{
				ID:          "allowed-ip",
				Description: "Request must originate from an allowed IP",
				Source:      model.SourceEnvironment,
				Attribute:   "ip",
				Operator:    model.OperatorIn,
				Value: []string{
					"192.168.1.10",
					"192.168.1.20",
				},
			},
			{
				ID:          "resource-integrity",
				Description: "Resource content hash must match the expected hash",
				Source:      model.SourceResource,
				Attribute:   "contentHash",
				Operator:    model.OperatorEquals,
				Value:       "sha256:abc123",
			},
		},
	}
}

func validSampleContext() model.RequestContext {
	return model.RequestContext{
		Subject: model.SubjectContext{
			Attributes: map[string]interface{}{
				"role": "analyst",
				"groups": []string{
					"cipher",
					"developers",
				},
			},
		},

		Resource: model.ResourceContext{
			ID: "ddm-resource-1",
			Attributes: map[string]interface{}{
				"contentHash": "sha256:abc123",
			},
		},

		Action: model.ActionContext{
			Name: "read",
		},

		Environment: model.EnvironmentContext{
			Attributes: map[string]interface{}{
				"ip": "192.168.1.10",
			},
		},
	}
}

func TestSamplePolicyPermit(t *testing.T) {
	engine := NewEngine()

	decision, err := engine.Evaluate(
		samplePolicy(),
		validSampleContext(),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.Decision != model.DecisionPermit {
		t.Fatalf(
			"expected permit, got %s",
			decision.Decision,
		)
	}

	for _, evaluation := range decision.Evaluations {
		if !evaluation.Result {
			t.Errorf(
				"expected rule %s to pass",
				evaluation.RuleID,
			)
		}
	}
}

func TestSamplePolicyDenyWhenRoleIsInsufficient(t *testing.T) {
	engine := NewEngine()

	context := validSampleContext()

	policy := samplePolicy()

	policy.Rules[0].Value = []string{
		"analyst-senior",
		"admin",
	}

	decision, err := engine.Evaluate(
		policy,
		context,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.Decision != model.DecisionDeny {
		t.Fatalf(
			"expected deny, got %s",
			decision.Decision,
		)
	}
}

func TestSamplePolicyDenyWhenGroupIsNotAllowed(t *testing.T) {
	engine := NewEngine()

	context := validSampleContext()

	context.Subject.Attributes["groups"] = []string{
		"external-users",
	}

	decision, err := engine.Evaluate(
		samplePolicy(),
		context,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.Decision != model.DecisionDeny {
		t.Fatalf(
			"expected deny, got %s",
			decision.Decision,
		)
	}
}
