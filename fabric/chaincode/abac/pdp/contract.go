package pdp

import (
	"encoding/json"
	"fmt"

	"cipher/fabric/chaincode/abac/model"
	"cipher/fabric/chaincode/abac/pap"
	"cipher/fabric/chaincode/abac/pip"
	"cipher/fabric/chaincode/abac/policy"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Contract struct {
	contractapi.Contract
}

func NewContract() *Contract {
	return &Contract{
		Contract: contractapi.Contract{
			Name: "pdp",
		},
	}
}

func (c *Contract) EvaluateRequest(
	ctx contractapi.TransactionContextInterface,
	requestJSON string,
) (*model.AccessDecision, error) {

	var requestContext model.RequestContext

	if err := json.Unmarshal(
		[]byte(requestJSON),
		&requestContext,
	); err != nil {
		return nil, fmt.Errorf(
			"invalid request context JSON: %w",
			err,
		)
	}

	if requestContext.Resource.ID == "" {
		return nil, fmt.Errorf(
			"resource ID is required",
		)
	}

	//
	// 1. Resolve the policy through the PAP.
	//

	papContract := new(pap.Contract)

	applicablePolicy, err := papContract.GetPolicyForResource(
		ctx,
		requestContext.Resource.ID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to resolve policy for resource %s: %w",
			requestContext.Resource.ID,
			err,
		)
	}

	//
	// 2. Fail closed if the policy is inactive.
	//

	if applicablePolicy.Status != model.PolicyStatusActive {
		return &model.AccessDecision{
			Decision:      model.DecisionDeny,
			PolicyID:      applicablePolicy.ID,
			PolicyVersion: applicablePolicy.Version,
			ResourceID:    requestContext.Resource.ID,
			Evaluations:   []model.RuleEvaluation{},
		}, nil
	}

	//
	// 3. Retrieve trusted resource information from PIP.
	//

	pipContract := new(pip.Contract)

	resource, err := pipContract.GetResource(
		ctx,
		requestContext.Resource.ID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to retrieve resource %s: %w",
			requestContext.Resource.ID,
			err,
		)
	}

	//
	// 4. Replace client-provided resource attributes with
	//    trusted resource information from Fabric.
	//

	requestContext.Resource.Attributes =
		buildTrustedResourceAttributes(resource)

	//
	// 5. Evaluate the policy.
	//

	engine := policy.NewEngine()

	decision, err := engine.Evaluate(
		*applicablePolicy,
		requestContext,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to evaluate policy %s: %w",
			applicablePolicy.ID,
			err,
		)
	}

	return &decision, nil
}

func buildTrustedResourceAttributes(
	resource *model.Resource,
) map[string]interface{} {

	attributes := make(map[string]interface{})

	for key, value := range resource.Attributes {
		attributes[key] = value
	}

	attributes["uri"] = resource.URI
	attributes["owner"] = resource.Owner
	attributes["contentHash"] = resource.ContentHash

	return attributes
}
