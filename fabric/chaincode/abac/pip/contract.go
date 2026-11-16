package pip

import (
	"encoding/json"
	"fmt"

	"cipher/fabric/chaincode/abac/model"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Contract struct {
	contractapi.Contract
}

func NewContract() *Contract {
	return &Contract{
		Contract: contractapi.Contract{
			Name: "pip",
		},
	}
}

func (c *Contract) CreateResource(
	ctx contractapi.TransactionContextInterface,
	resourceJSON string,
) error {

	var resource model.Resource

	if err := json.Unmarshal(
		[]byte(resourceJSON),
		&resource,
	); err != nil {
		return fmt.Errorf(
			"invalid resource JSON: %w",
			err,
		)
	}

	if resource.ID == "" {
		return fmt.Errorf("resource ID is required")
	}

	if resource.URI == "" {
		return fmt.Errorf("resource URI is required")
	}

	key := resourceKey(resource.ID)

	existing, err := ctx.GetStub().GetState(key)
	if err != nil {
		return fmt.Errorf(
			"failed to check resource existence: %w",
			err,
		)
	}

	if existing != nil {
		return fmt.Errorf(
			"resource %s already exists",
			resource.ID,
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

	resource.CreatedAt = timestamp
	resource.CreatedBy = actor

	resource.UpdatedAt = ""
	resource.UpdatedBy = ""

	data, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf(
			"failed to serialise resource: %w",
			err,
		)
	}

	if err := ctx.GetStub().PutState(key, data); err != nil {
		return fmt.Errorf(
			"failed to store resource: %w",
			err,
		)
	}

	return nil
}

func (c *Contract) GetResource(
	ctx contractapi.TransactionContextInterface,
	resourceID string,
) (*model.Resource, error) {

	data, err := ctx.GetStub().GetState(
		resourceKey(resourceID),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read resource %s: %w",
			resourceID,
			err,
		)
	}

	if data == nil {
		return nil, fmt.Errorf(
			"resource %s does not exist",
			resourceID,
		)
	}

	var resource model.Resource

	if err := json.Unmarshal(
		data,
		&resource,
	); err != nil {
		return nil, fmt.Errorf(
			"failed to deserialise resource %s: %w",
			resourceID,
			err,
		)
	}

	return &resource, nil
}

func (c *Contract) UpdateResource(
	ctx contractapi.TransactionContextInterface,
	resourceJSON string,
) error {

	var updated model.Resource

	if err := json.Unmarshal(
		[]byte(resourceJSON),
		&updated,
	); err != nil {
		return fmt.Errorf(
			"invalid resource JSON: %w",
			err,
		)
	}

	if updated.ID == "" {
		return fmt.Errorf("resource ID is required")
	}

	current, err := c.GetResource(
		ctx,
		updated.ID,
	)
	if err != nil {
		return err
	}

	if updated.URI == "" {
		return fmt.Errorf("resource URI is required")
	}

	actor, err := getClientIdentity(ctx)
	if err != nil {
		return err
	}

	timestamp, err := getTransactionTimestamp(ctx)
	if err != nil {
		return err
	}

	updated.CreatedAt = current.CreatedAt
	updated.CreatedBy = current.CreatedBy

	updated.UpdatedAt = timestamp
	updated.UpdatedBy = actor

	data, err := json.Marshal(updated)
	if err != nil {
		return fmt.Errorf(
			"failed to serialise resource: %w",
			err,
		)
	}

	if err := ctx.GetStub().PutState(
		resourceKey(updated.ID),
		data,
	); err != nil {
		return fmt.Errorf(
			"failed to update resource: %w",
			err,
		)
	}

	return nil
}
