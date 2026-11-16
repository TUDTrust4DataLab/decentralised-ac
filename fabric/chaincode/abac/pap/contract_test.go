package pap

import (
	"encoding/json"
	"testing"

	"cipher/fabric/chaincode/abac/model"
	"cipher/fabric/chaincode/abac/util"
)

func testPolicy() model.Policy {
	return model.Policy{
		ID:                 "policy-001",
		Name:               "DDM Access Policy",
		Description:        "Policy used to protect DDM resources",
		CombiningAlgorithm: model.CombiningAllOf,
		Rules: []model.Rule{
			{
				ID:        "role-rule",
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
}

func policyJSON(t *testing.T, policy model.Policy) string {
	t.Helper()

	data, err := json.Marshal(policy)
	if err != nil {
		t.Fatalf("failed to marshal policy: %v", err)
	}

	return string(data)
}

func TestCreatePolicy(t *testing.T) {
	ctx := util.NewTestContext()
	contract := new(Contract)

	policy := testPolicy()

	util.StartTransaction(ctx, "tx-create-policy")

	err := contract.CreatePolicy(
		ctx,
		policyJSON(t, policy),
	)

	util.EndTransaction(ctx, "tx-create-policy")

	if err != nil {
		t.Fatalf("CreatePolicy returned error: %v", err)
	}

	stored, err := contract.GetPolicy(
		ctx,
		policy.ID,
	)

	if err != nil {
		t.Fatalf("GetPolicy returned error: %v", err)
	}

	if stored.ID != policy.ID {
		t.Fatalf(
			"expected policy ID %s, got %s",
			policy.ID,
			stored.ID,
		)
	}

	if stored.Status != model.PolicyStatusActive {
		t.Fatalf(
			"expected active policy, got %s",
			stored.Status,
		)
	}

	if stored.Version != 1 {
		t.Fatalf(
			"expected version 1, got %d",
			stored.Version,
		)
	}

	if stored.CreatedBy == "" {
		t.Fatal("expected CreatedBy to be populated")
	}

	if stored.CreatedAt == "" {
		t.Fatal("expected CreatedAt to be populated")
	}
}

func TestCreatePolicyDuplicate(t *testing.T) {
	ctx := util.NewTestContext()
	contract := new(Contract)

	policy := testPolicy()

	util.StartTransaction(ctx, "tx-create-policy")

	err := contract.CreatePolicy(
		ctx,
		policyJSON(t, policy),
	)

	util.EndTransaction(ctx, "tx-create-policy")

	if err != nil {
		t.Fatalf("unexpected creation error: %v", err)
	}

	util.StartTransaction(ctx, "tx-create-policy-duplicate")

	err = contract.CreatePolicy(
		ctx,
		policyJSON(t, policy),
	)

	util.EndTransaction(ctx, "tx-create-policy-duplicate")

	if err == nil {
		t.Fatal("expected duplicate creation to fail")
	}
}

func TestUpdatePolicy(t *testing.T) {
	ctx := util.NewTestContext()
	contract := new(Contract)

	policy := testPolicy()

	util.StartTransaction(ctx, "tx-create-policy")

	err := contract.CreatePolicy(
		ctx,
		policyJSON(t, policy),
	)

	util.EndTransaction(ctx, "tx-create-policy")

	if err != nil {
		t.Fatalf("unexpected creation error: %v", err)
	}

	policy.Name = "Updated DDM Access Policy"

	util.StartTransaction(ctx, "tx-update-policy")

	err = contract.UpdatePolicy(
		ctx,
		policyJSON(t, policy),
	)

	util.EndTransaction(ctx, "tx-update-policy")

	if err != nil {
		t.Fatalf("UpdatePolicy returned error: %v", err)
	}

	updated, err := contract.GetPolicy(
		ctx,
		policy.ID,
	)

	if err != nil {
		t.Fatalf("GetPolicy returned error: %v", err)
	}

	if updated.Name != "Updated DDM Access Policy" {
		t.Fatalf(
			"expected updated name, got %s",
			updated.Name,
		)
	}

	if updated.Version != 2 {
		t.Fatalf(
			"expected version 2, got %d",
			updated.Version,
		)
	}

	if updated.UpdatedBy == "" {
		t.Fatal("expected UpdatedBy to be populated")
	}
}

func TestAssignPolicyToResource(t *testing.T) {
	ctx := util.NewTestContext()
	contract := new(Contract)

	policy := testPolicy()

	ctx.Stub.MockTransactionStart("tx-create-policy")

	err := contract.CreatePolicy(
		ctx,
		policyJSON(t, policy),
	)

	ctx.Stub.MockTransactionEnd("tx-create-policy")

	if err != nil {
		t.Fatalf("unexpected creation error: %v", err)
	}

	ctx.Stub.MockTransactionStart("tx-assign-policy")

	err = contract.AssignPolicyToResource(
		ctx,
		"ddm-resource-1",
		policy.ID,
	)

	ctx.Stub.MockTransactionEnd("tx-assign-policy")

	if err != nil {
		t.Fatalf(
			"AssignPolicyToResource returned error: %v",
			err,
		)
	}

	resolved, err := contract.GetPolicyForResource(
		ctx,
		"ddm-resource-1",
	)

	if err != nil {
		t.Fatalf(
			"GetPolicyForResource returned error: %v",
			err,
		)
	}

	if resolved.ID != policy.ID {
		t.Fatalf(
			"expected policy %s, got %s",
			policy.ID,
			resolved.ID,
		)
	}
}

func TestDeletePolicyDeactivatesPolicy(t *testing.T) {
	ctx := util.NewTestContext()
	contract := new(Contract)

	policy := testPolicy()

	ctx.Stub.MockTransactionStart("tx-create-policy")

	err := contract.CreatePolicy(
		ctx,
		policyJSON(t, policy),
	)

	ctx.Stub.MockTransactionEnd("tx-create-policy")

	if err != nil {
		t.Fatalf("unexpected creation error: %v", err)
	}

	ctx.Stub.MockTransactionStart("tx-delete-policy")

	err = contract.DeletePolicy(
		ctx,
		policy.ID,
	)

	ctx.Stub.MockTransactionEnd("tx-delete-policy")

	if err != nil {
		t.Fatalf("DeletePolicy returned error: %v", err)
	}

	stored, err := contract.GetPolicy(
		ctx,
		policy.ID,
	)

	if err != nil {
		t.Fatalf(
			"policy should still exist after deletion: %v",
			err,
		)
	}

	if stored.Status != model.PolicyStatusInactive {
		t.Fatalf(
			"expected inactive policy, got %s",
			stored.Status,
		)
	}

	if stored.DeactivatedAt == "" {
		t.Fatal("expected DeactivatedAt to be populated")
	}

	if stored.DeactivatedBy == "" {
		t.Fatal("expected DeactivatedBy to be populated")
	}

	if stored.Version != 2 {
		t.Fatalf(
			"expected version 2 after deactivation, got %d",
			stored.Version,
		)
	}
}

func TestUpdateInactivePolicyFails(t *testing.T) {
	ctx := util.NewTestContext()
	contract := new(Contract)

	policy := testPolicy()

	ctx.Stub.MockTransactionStart("tx-create-policy")

	err := contract.CreatePolicy(
		ctx,
		policyJSON(t, policy),
	)

	ctx.Stub.MockTransactionEnd("tx-create-policy")

	if err != nil {
		t.Fatalf("unexpected creation error: %v", err)
	}

	ctx.Stub.MockTransactionStart("tx-delete-policy")

	err = contract.DeletePolicy(
		ctx,
		policy.ID,
	)

	ctx.Stub.MockTransactionEnd("tx-delete-policy")

	if err != nil {
		t.Fatalf("unexpected deletion error: %v", err)
	}

	policy.Name = "Should not work"

	ctx.Stub.MockTransactionStart("tx-update-inactive-policy")

	err = contract.UpdatePolicy(
		ctx,
		policyJSON(t, policy),
	)

	ctx.Stub.MockTransactionEnd("tx-update-inactive-policy")

	if err == nil {
		t.Fatal(
			"expected update of inactive policy to fail",
		)
	}
}

func TestAssignInactivePolicyFails(t *testing.T) {
	ctx := util.NewTestContext()
	contract := new(Contract)

	policy := testPolicy()

	ctx.Stub.MockTransactionStart("tx-create-policy")

	err := contract.CreatePolicy(
		ctx,
		policyJSON(t, policy),
	)

	ctx.Stub.MockTransactionEnd("tx-create-policy")

	if err != nil {
		t.Fatalf("unexpected creation error: %v", err)
	}

	ctx.Stub.MockTransactionStart("tx-delete-policy")

	err = contract.DeletePolicy(
		ctx,
		policy.ID,
	)

	ctx.Stub.MockTransactionEnd("tx-delete-policy")

	if err != nil {
		t.Fatalf("unexpected deletion error: %v", err)
	}

	ctx.Stub.MockTransactionStart("tx-assign-inactive-policy")

	err = contract.AssignPolicyToResource(
		ctx,
		"ddm-resource-1",
		policy.ID,
	)

	ctx.Stub.MockTransactionEnd("tx-assign-inactive-policy")

	if err == nil {
		t.Fatal(
			"expected assignment of inactive policy to fail",
		)
	}
}
