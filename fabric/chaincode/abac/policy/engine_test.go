package policy

import (
	"testing"

	"cipher/fabric/chaincode/abac/model"
)

func TestEvaluateRolePermit(t *testing.T) {
	engine := NewEngine()

	context := model.RequestContext{
		Subject: model.SubjectContext{
			Attributes: map[string]interface{}{
				"role": "analyst",
			},
		},
		Resource: model.ResourceContext{
			ID: "ddm-resource-1",
		},
	}

	policy := model.Policy{
		ID:                 "ddm-policy",
		Version:            1,
		Status:             model.PolicyStatusActive,
		CombiningAlgorithm: model.CombiningAllOf,
		Rules: []model.Rule{
			{
				ID:        "required-role",
				Source:    model.SourceSubject,
				Attribute: "role",
				Operator:  model.OperatorIn,
				Value: []string{
					"analyst",
					"admin",
				},
			},
		},
	}

	decision, err := engine.Evaluate(policy, context)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.Decision != model.DecisionPermit {
		t.Fatalf(
			"expected permit, got %s",
			decision.Decision,
		)
	}
}

func TestEvaluateRoleDeny(t *testing.T) {
	engine := NewEngine()

	context := model.RequestContext{
		Subject: model.SubjectContext{
			Attributes: map[string]interface{}{
				"role": "guest",
			},
		},
		Resource: model.ResourceContext{
			ID: "ddm-resource-1",
		},
	}

	policy := model.Policy{
		ID:                 "ddm-policy",
		Version:            1,
		Status:             model.PolicyStatusActive,
		CombiningAlgorithm: model.CombiningAllOf,
		Rules: []model.Rule{
			{
				ID:        "required-role",
				Source:    model.SourceSubject,
				Attribute: "role",
				Operator:  model.OperatorIn,
				Value: []string{
					"analyst",
					"admin",
				},
			},
		},
	}

	decision, err := engine.Evaluate(policy, context)

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

func TestEvaluateVerifiedClaim(t *testing.T) {
	engine := NewEngine()

	context := model.RequestContext{
		Subject: model.SubjectContext{
			Claims: map[string]model.VerifiedClaim{
				"ageOver18": {
					Verified: true,
				},
			},
		},
		Resource: model.ResourceContext{
			ID: "ddm-resource-1",
		},
	}

	policy := model.Policy{
		ID:                 "adult-access-policy",
		Version:            1,
		Status:             model.PolicyStatusActive,
		CombiningAlgorithm: model.CombiningAllOf,
		Rules: []model.Rule{
			{
				ID:        "adult-user",
				Source:    model.SourceClaim,
				Attribute: "ageOver18",
				Operator:  model.OperatorVerified,
			},
		},
	}

	decision, err := engine.Evaluate(policy, context)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decision.Decision != model.DecisionPermit {
		t.Fatalf(
			"expected permit, got %s",
			decision.Decision,
		)
	}
}
