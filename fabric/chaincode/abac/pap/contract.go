package pap

import (
	"encoding/json"
	"fmt"

	"cipher/fabric/chaincode/abac/model"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	policyKeyPrefix         = "policy:"
	resourcePolicyKeyPrefix = "resource-policy:"
)

type Contract struct {
	contractapi.Contract
}

func NewContract() *Contract {
	return &Contract{
		Contract: contractapi.Contract{
			Name: "pap",
		},
	}
}

func (c *Contract) CreatePolicy(
	ctx contractapi.TransactionContextInterface,
	policyJSON string,
) error {

	var policy model.Policy

	if err := json.Unmarshal([]byte(policyJSON), &policy); err != nil {
		return fmt.Errorf("invalid policy JSON: %w", err)
	}

	if policy.ID == "" {
		return fmt.Errorf("policy ID is required")
	}

	key := policyKey(policy.ID)

	existing, err := ctx.GetStub().GetState(key)
	if err != nil {
		return fmt.Errorf("failed to check policy existence: %w", err)
	}

	if existing != nil {
		return fmt.Errorf("policy %s already exists", policy.ID)
	}

	actor, err := getClientIdentity(ctx)
	if err != nil {
		return err
	}

	timestamp, err := getTransactionTimestamp(ctx)
	if err != nil {
		return err
	}

	policy.Status = model.PolicyStatusActive
	policy.CreatedBy = actor
	policy.CreatedAt = timestamp

	// Lifecycle metadata is controlled by the chaincode.
	policy.UpdatedBy = ""
	policy.UpdatedAt = ""
	policy.DeactivatedBy = ""
	policy.DeactivatedAt = ""

	if policy.Version <= 0 {
		policy.Version = 1
	}

	data, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to serialise policy: %w", err)
	}

	if err := ctx.GetStub().PutState(key, data); err != nil {
		return fmt.Errorf("failed to store policy: %w", err)
	}

	return nil
}

func (c *Contract) GetPolicy(
	ctx contractapi.TransactionContextInterface,
	policyID string,
) (*model.Policy, error) {

	data, err := ctx.GetStub().GetState(policyKey(policyID))
	if err != nil {
		return nil, fmt.Errorf("failed to read policy %s: %w", policyID, err)
	}

	if data == nil {
		return nil, fmt.Errorf("policy %s does not exist", policyID)
	}

	var policy model.Policy

	if err := json.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("failed to deserialise policy %s: %w", policyID, err)
	}

	return &policy, nil
}

func (c *Contract) UpdatePolicy(
	ctx contractapi.TransactionContextInterface,
	policyJSON string,
) error {

	var updated model.Policy

	if err := json.Unmarshal([]byte(policyJSON), &updated); err != nil {
		return fmt.Errorf("invalid policy JSON: %w", err)
	}

	if updated.ID == "" {
		return fmt.Errorf("policy ID is required")
	}

	current, err := c.GetPolicy(ctx, updated.ID)
	if err != nil {
		return err
	}

	if current.Status != model.PolicyStatusActive {
		return fmt.Errorf(
			"policy %s is inactive and cannot be updated",
			updated.ID,
		)
	}

	actor, err := getClientIdentity(ctx)
	if err != nil {
		return err
	}

	timestamp, err := getTransactionTimestamp(ctx)
	if err != nil {
		return err
	}

	// Preserve chaincode-controlled lifecycle information.
	updated.Status = current.Status

	updated.CreatedAt = current.CreatedAt
	updated.CreatedBy = current.CreatedBy

	updated.UpdatedAt = timestamp
	updated.UpdatedBy = actor

	updated.DeactivatedAt = current.DeactivatedAt
	updated.DeactivatedBy = current.DeactivatedBy

	// Versioning is controlled by the chaincode.
	updated.Version = current.Version + 1

	data, err := json.Marshal(updated)
	if err != nil {
		return fmt.Errorf("failed to serialise policy: %w", err)
	}

	if err := ctx.GetStub().PutState(policyKey(updated.ID), data); err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	return nil
}

func (c *Contract) DeletePolicy(
	ctx contractapi.TransactionContextInterface,
	policyID string,
) error {

	policy, err := c.GetPolicy(ctx, policyID)
	if err != nil {
		return err
	}

	if policy.Status == model.PolicyStatusInactive {
		return fmt.Errorf(
			"policy %s is already inactive",
			policyID,
		)
	}

	actor, err := getClientIdentity(ctx)
	if err != nil {
		return err
	}

	timestamp, err := getTransactionTimestamp(ctx)
	if err != nil {
		return err
	}

	policy.Status = model.PolicyStatusInactive
	policy.DeactivatedBy = actor
	policy.DeactivatedAt = timestamp

	// Inactivation is itself a state change.
	policy.Version++
	policy.UpdatedBy = actor
	policy.UpdatedAt = timestamp

	data, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to serialise policy: %w", err)
	}

	if err := ctx.GetStub().PutState(policyKey(policy.ID), data); err != nil {
		return fmt.Errorf("failed to deactivate policy: %w", err)
	}

	return nil
}

func (c *Contract) AssignPolicyToResource(
	ctx contractapi.TransactionContextInterface,
	resourceID string,
	policyID string,
) error {

	if resourceID == "" {
		return fmt.Errorf("resource ID is required")
	}

	policy, err := c.GetPolicy(ctx, policyID)
	if err != nil {
		return err
	}

	if policy.Status != model.PolicyStatusActive {
		return fmt.Errorf(
			"cannot assign inactive policy %s",
			policyID,
		)
	}

	if err := ctx.GetStub().PutState(
		resourcePolicyKey(resourceID),
		[]byte(policyID),
	); err != nil {
		return fmt.Errorf(
			"failed to assign policy %s to resource %s: %w",
			policyID,
			resourceID,
			err,
		)
	}

	return nil
}

func (c *Contract) GetPolicyForResource(
	ctx contractapi.TransactionContextInterface,
	resourceID string,
) (*model.Policy, error) {

	policyID, err := ctx.GetStub().GetState(
		resourcePolicyKey(resourceID),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read policy assignment for resource %s: %w",
			resourceID,
			err,
		)
	}

	if policyID == nil {
		return nil, fmt.Errorf(
			"resource %s has no assigned policy",
			resourceID,
		)
	}

	return c.GetPolicy(ctx, string(policyID))
}
