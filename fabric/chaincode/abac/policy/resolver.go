package policy

import (
	"fmt"

	"cipher/fabric/chaincode/abac/model"
)

func resolveAttribute(
	source model.AttributeSource,
	attribute string,
	context model.RequestContext,
) (interface{}, bool, error) {

	switch source {

	case model.SourceSubject:
		value, exists := context.Subject.Attributes[attribute]
		return value, exists, nil

	case model.SourceResource:
		value, exists := context.Resource.Attributes[attribute]
		return value, exists, nil

	case model.SourceAction:
		if attribute == "name" {
			return context.Action.Name, true, nil
		}

		value, exists := context.Action.Attributes[attribute]
		return value, exists, nil

	case model.SourceEnvironment:
		value, exists := context.Environment.Attributes[attribute]
		return value, exists, nil

	case model.SourceClaim:
		claim, exists := context.Subject.Claims[attribute]

		if !exists {
			return nil, false, nil
		}

		return claim, true, nil

	default:
		return nil, false, fmt.Errorf(
			"unsupported attribute source: %s",
			source,
		)
	}
}
