package policy

import (
	"fmt"
	"reflect"

	"cipher/fabric/chaincode/abac/model"
)

func evaluateRule(
	rule model.Rule,
	context model.RequestContext,
) (bool, error) {

	actualValue, exists, err := resolveAttribute(
		rule.Source,
		rule.Attribute,
		context,
	)

	if err != nil {
		return false, err
	}

	// Missing attributes fail the rule.
	if !exists {
		return false, nil
	}

	switch rule.Operator {

	case model.OperatorEquals:
		return reflect.DeepEqual(actualValue, rule.Value), nil

	case model.OperatorNotEquals:
		return !reflect.DeepEqual(actualValue, rule.Value), nil

	case model.OperatorIn:
		return evaluateIn(actualValue, rule.Value)

	case model.OperatorAnyOf:
		return evaluateAnyOf(actualValue, rule.Value)

	case model.OperatorAllOf:
		return evaluateAllOf(actualValue, rule.Value)

	case model.OperatorVerified:
		claim, ok := actualValue.(model.VerifiedClaim)

		if !ok {
			return false, fmt.Errorf(
				"operator %q requires a verified claim",
				rule.Operator,
			)
		}

		return claim.Verified, nil

	default:
		return false, fmt.Errorf(
			"unsupported operator: %s",
			rule.Operator,
		)
	}
}

func evaluateIn(actual interface{}, expected interface{}) (bool, error) {
	expectedValues, err := toStringSlice(expected)
	if err != nil {
		return false, fmt.Errorf(
			"in operator: %w",
			err,
		)
	}

	actualString, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf(
			"in operator expects string context value",
		)
	}

	for _, value := range expectedValues {
		if actualString == value {
			return true, nil
		}
	}

	return false, nil
}

func evaluateAnyOf(actual interface{}, expected interface{}) (bool, error) {
	actualValues, err := toStringSlice(actual)
	if err != nil {
		return false, fmt.Errorf(
			"anyOf actual value: %w",
			err,
		)
	}

	expectedValues, err := toStringSlice(expected)
	if err != nil {
		return false, fmt.Errorf(
			"anyOf expected value: %w",
			err,
		)
	}

	for _, actualValue := range actualValues {
		for _, expectedValue := range expectedValues {
			if actualValue == expectedValue {
				return true, nil
			}
		}
	}

	return false, nil
}

func evaluateAllOf(actual interface{}, expected interface{}) (bool, error) {
	actualValues, err := toStringSlice(actual)
	if err != nil {
		return false, fmt.Errorf(
			"allOf actual value: %w",
			err,
		)
	}

	expectedValues, err := toStringSlice(expected)
	if err != nil {
		return false, fmt.Errorf(
			"allOf expected value: %w",
			err,
		)
	}

	for _, expectedValue := range expectedValues {
		found := false

		for _, actualValue := range actualValues {
			if actualValue == expectedValue {
				found = true
				break
			}
		}

		if !found {
			return false, nil
		}
	}

	return true, nil
}

func toStringSlice(value interface{}) ([]string, error) {
	switch values := value.(type) {

	case []string:
		return values, nil

	case []interface{}:
		result := make([]string, 0, len(values))

		for _, item := range values {
			stringValue, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf(
					"expected string array, found %T",
					item,
				)
			}

			result = append(result, stringValue)
		}

		return result, nil

	default:
		return nil, fmt.Errorf(
			"expected string array, found %T",
			value,
		)
	}
}
