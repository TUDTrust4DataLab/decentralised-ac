package pdp

import (
	"encoding/json"
	"testing"

	"cipher/fabric/chaincode/abac/model"
	"cipher/fabric/chaincode/abac/pap"
	"cipher/fabric/chaincode/abac/pip"
	"cipher/fabric/chaincode/abac/util"
)

func policyJSON(
	t *testing.T,
	policy model.Policy,
) string {
	t.Helper()

	data, err := json.Marshal(policy)
	if err != nil {
		t.Fatalf(
			"failed to marshal policy: %v",
			err,
		)
	}

	return string(data)
}

func resourceJSON(
	t *testing.T,
	resource model.Resource,
) string {
	t.Helper()

	data, err := json.Marshal(resource)
	if err != nil {
		t.Fatalf(
			"failed to marshal resource: %v",
			err,
		)
	}

	return string(data)
}

func requestJSON(
	t *testing.T,
	request model.RequestContext,
) string {
	t.Helper()

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf(
			"failed to marshal request: %v",
			err,
		)
	}

	return string(data)
}

func testPolicy() model.Policy {
	return model.Policy{
		ID:                 "policy-001",
		Name:               "DDM Access Policy",
		Description:        "Policy used to protect a DDM resource",
		CombiningAlgorithm: model.CombiningAllOf,

		Rules: []model.Rule{
			{
				ID:          "role-rule",
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
				ID:          "group-rule",
				Description: "Subject must belong to an allowed group",
				Source:      model.SourceSubject,
				Attribute:   "groups",
				Operator:    model.OperatorAnyOf,
				Value: []string{
					"cipher",
				},
			},
			{
				ID:          "ip-rule",
				Description: "Request must originate from an allowed IP",
				Source:      model.SourceEnvironment,
				Attribute:   "ip",
				Operator:    model.OperatorIn,
				Value: []string{
					"192.168.1.10",
				},
			},
			{
				ID:          "resource-hash-rule",
				Description: "Resource hash must match the expected value",
				Source:      model.SourceResource,
				Attribute:   "contentHash",
				Operator:    model.OperatorEquals,
				Value:       "sha256:abc123",
			},
		},
	}
}

func testResource() model.Resource {
	return model.Resource{
		ID:          "ddm-resource-1",
		URI:         "/ddm/models/1",
		Owner:       "CIPHER",
		ContentHash: "sha256:abc123",

		Attributes: map[string]interface{}{
			"type": "ddm-model",
		},
	}
}

func validRequest() model.RequestContext {
	return model.RequestContext{
		Subject: model.SubjectContext{
			ID: "user-001",
			Attributes: map[string]interface{}{
				"role": "analyst",
				"groups": []string{
					"cipher",
				},
			},
			Claims: map[string]model.VerifiedClaim{},
		},

		Resource: model.ResourceContext{
			ID:         "ddm-resource-1",
			Attributes: map[string]interface{}{},
		},

		Action: model.ActionContext{
			Name:       "read",
			Attributes: map[string]interface{}{},
		},

		Environment: model.EnvironmentContext{
			Attributes: map[string]interface{}{
				"ip": "192.168.1.10",
			},
		},
	}
}

func prepareLedger(
	t *testing.T,
	ctx *util.MockTransactionContext,
) {
	t.Helper()

	papContract := new(pap.Contract)
	pipContract := new(pip.Contract)

	policy := testPolicy()
	resource := testResource()

	//
	// Transaction 1:
	// Create resource
	//

	util.StartTransaction(
		ctx,
		"tx-create-resource",
	)

	err := pipContract.CreateResource(
		ctx,
		resourceJSON(t, resource),
	)

	util.EndTransaction(
		ctx,
		"tx-create-resource",
	)

	if err != nil {
		t.Fatalf(
			"failed to create resource: %v",
			err,
		)
	}

	//
	// Transaction 2:
	// Create policy
	//

	util.StartTransaction(
		ctx,
		"tx-create-policy",
	)

	err = papContract.CreatePolicy(
		ctx,
		policyJSON(t, policy),
	)

	util.EndTransaction(
		ctx,
		"tx-create-policy",
	)

	if err != nil {
		t.Fatalf(
			"failed to create policy: %v",
			err,
		)
	}

	//
	// Transaction 3:
	// Assign policy to resource
	//

	util.StartTransaction(
		ctx,
		"tx-assign-policy",
	)

	err = papContract.AssignPolicyToResource(
		ctx,
		resource.ID,
		policy.ID,
	)

	util.EndTransaction(
		ctx,
		"tx-assign-policy",
	)

	if err != nil {
		t.Fatalf(
			"failed to assign policy to resource: %v",
			err,
		)
	}
}

func TestEvaluateRequestPermit(t *testing.T) {
	ctx := util.NewTestContext()

	prepareLedger(
		t,
		ctx,
	)

	contract := NewContract()

	request := validRequest()

	decision, err := contract.EvaluateRequest(
		ctx,
		requestJSON(t, request),
	)

	if err != nil {
		t.Fatalf(
			"EvaluateRequest returned error: %v",
			err,
		)
	}

	if decision.Decision != model.DecisionPermit {
		t.Fatalf(
			"expected PERMIT, got %s",
			decision.Decision,
		)
	}

	if decision.PolicyID != "policy-001" {
		t.Fatalf(
			"expected policy-001, got %s",
			decision.PolicyID,
		)
	}

	if decision.PolicyVersion != 1 {
		t.Fatalf(
			"expected policy version 1, got %d",
			decision.PolicyVersion,
		)
	}

	if decision.ResourceID != "ddm-resource-1" {
		t.Fatalf(
			"expected resource ddm-resource-1, got %s",
			decision.ResourceID,
		)
	}

	if len(decision.Evaluations) != 4 {
		t.Fatalf(
			"expected 4 rule evaluations, got %d",
			len(decision.Evaluations),
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

func TestEvaluateRequestDenyInvalidRole(t *testing.T) {
	ctx := util.NewTestContext()

	prepareLedger(
		t,
		ctx,
	)

	contract := NewContract()

	request := validRequest()
	request.Subject.Attributes["role"] = "guest"

	decision, err := contract.EvaluateRequest(
		ctx,
		requestJSON(t, request),
	)

	if err != nil {
		t.Fatalf(
			"EvaluateRequest returned error: %v",
			err,
		)
	}

	if decision.Decision != model.DecisionDeny {
		t.Fatalf(
			"expected DENY, got %s",
			decision.Decision,
		)
	}

	assertRuleResult(
		t,
		decision,
		"role-rule",
		false,
	)
}

func TestEvaluateRequestDenyInvalidGroup(t *testing.T) {
	ctx := util.NewTestContext()

	prepareLedger(
		t,
		ctx,
	)

	contract := NewContract()

	request := validRequest()

	request.Subject.Attributes["groups"] = []string{
		"external-users",
	}

	decision, err := contract.EvaluateRequest(
		ctx,
		requestJSON(t, request),
	)

	if err != nil {
		t.Fatalf(
			"EvaluateRequest returned error: %v",
			err,
		)
	}

	if decision.Decision != model.DecisionDeny {
		t.Fatalf(
			"expected DENY, got %s",
			decision.Decision,
		)
	}

	assertRuleResult(
		t,
		decision,
		"group-rule",
		false,
	)
}

func TestEvaluateRequestDenyInvalidIP(t *testing.T) {
	ctx := util.NewTestContext()

	prepareLedger(
		t,
		ctx,
	)

	contract := NewContract()

	request := validRequest()

	request.Environment.Attributes["ip"] =
		"10.0.0.55"

	decision, err := contract.EvaluateRequest(
		ctx,
		requestJSON(t, request),
	)

	if err != nil {
		t.Fatalf(
			"EvaluateRequest returned error: %v",
			err,
		)
	}

	if decision.Decision != model.DecisionDeny {
		t.Fatalf(
			"expected DENY, got %s",
			decision.Decision,
		)
	}

	assertRuleResult(
		t,
		decision,
		"ip-rule",
		false,
	)
}

func TestEvaluateRequestDenyInvalidResourceHash(t *testing.T) {
	ctx := util.NewTestContext()

	prepareLedger(
		t,
		ctx,
	)

	pipContract := new(pip.Contract)

	//
	// Simulate a change in the trusted resource.
	//
	// The policy still expects:
	// sha256:abc123
	//
	// but the PIP now stores:
	// sha256:def456
	//

	resource := testResource()
	resource.ContentHash = "sha256:def456"

	util.StartTransaction(
		ctx,
		"tx-update-resource",
	)

	err := pipContract.UpdateResource(
		ctx,
		resourceJSON(t, resource),
	)

	util.EndTransaction(
		ctx,
		"tx-update-resource",
	)

	if err != nil {
		t.Fatalf(
			"failed to update resource: %v",
			err,
		)
	}

	contract := NewContract()

	request := validRequest()

	decision, err := contract.EvaluateRequest(
		ctx,
		requestJSON(t, request),
	)

	if err != nil {
		t.Fatalf(
			"EvaluateRequest returned error: %v",
			err,
		)
	}

	if decision.Decision != model.DecisionDeny {
		t.Fatalf(
			"expected DENY, got %s",
			decision.Decision,
		)
	}

	assertRuleResult(
		t,
		decision,
		"resource-hash-rule",
		false,
	)
}

func TestEvaluateRequestDenyInactivePolicy(t *testing.T) {
	ctx := util.NewTestContext()

	prepareLedger(
		t,
		ctx,
	)

	papContract := new(pap.Contract)

	util.StartTransaction(
		ctx,
		"tx-delete-policy",
	)

	err := papContract.DeletePolicy(
		ctx,
		"policy-001",
	)

	util.EndTransaction(
		ctx,
		"tx-delete-policy",
	)

	if err != nil {
		t.Fatalf(
			"failed to deactivate policy: %v",
			err,
		)
	}

	contract := NewContract()

	request := validRequest()

	decision, err := contract.EvaluateRequest(
		ctx,
		requestJSON(t, request),
	)

	if err != nil {
		t.Fatalf(
			"EvaluateRequest returned error: %v",
			err,
		)
	}

	if decision.Decision != model.DecisionDeny {
		t.Fatalf(
			"expected DENY, got %s",
			decision.Decision,
		)
	}

	if decision.PolicyID != "policy-001" {
		t.Fatalf(
			"expected policy-001, got %s",
			decision.PolicyID,
		)
	}

	if len(decision.Evaluations) != 0 {
		t.Fatalf(
			"expected no rule evaluations for inactive policy, got %d",
			len(decision.Evaluations),
		)
	}
}

func TestEvaluateRequestWithoutPolicyFails(t *testing.T) {
	ctx := util.NewTestContext()

	pipContract := new(pip.Contract)

	resource := testResource()

	util.StartTransaction(
		ctx,
		"tx-create-resource",
	)

	err := pipContract.CreateResource(
		ctx,
		resourceJSON(t, resource),
	)

	util.EndTransaction(
		ctx,
		"tx-create-resource",
	)

	if err != nil {
		t.Fatalf(
			"failed to create resource: %v",
			err,
		)
	}

	contract := NewContract()

	request := validRequest()

	_, err = contract.EvaluateRequest(
		ctx,
		requestJSON(t, request),
	)

	if err == nil {
		t.Fatal(
			"expected request without assigned policy to fail",
		)
	}
}

func assertRuleResult(
	t *testing.T,
	decision *model.AccessDecision,
	ruleID string,
	expected bool,
) {
	t.Helper()

	for _, evaluation := range decision.Evaluations {
		if evaluation.RuleID == ruleID {

			if evaluation.Result != expected {
				t.Fatalf(
					"expected rule %s result %t, got %t",
					ruleID,
					expected,
					evaluation.Result,
				)
			}

			return
		}
	}

	t.Fatalf(
		"rule evaluation %s was not found",
		ruleID,
	)
}
