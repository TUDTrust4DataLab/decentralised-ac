package policy

import (
	"fmt"

	"cipher/fabric/chaincode/abac/model"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) Evaluate(
	policy model.Policy,
	context model.RequestContext,
) (model.AccessDecision, error) {

	if policy.Status != model.PolicyStatusActive {
		return model.AccessDecision{
			Decision:      model.DecisionDeny,
			PolicyID:      policy.ID,
			PolicyVersion: policy.Version,
			ResourceID:    context.Resource.ID,
		}, nil
	}

	evaluations := make([]model.RuleEvaluation, 0, len(policy.Rules))

	for _, rule := range policy.Rules {
		result, err := evaluateRule(rule, context)
		if err != nil {
			return model.AccessDecision{}, fmt.Errorf(
				"evaluate rule %q: %w",
				rule.ID,
				err,
			)
		}

		evaluations = append(evaluations, model.RuleEvaluation{
			RuleID: rule.ID,
			Result: result,
		})
	}

	decision, err := combine(
		policy.CombiningAlgorithm,
		evaluations,
	)
	if err != nil {
		return model.AccessDecision{}, err
	}

	return model.AccessDecision{
		Decision:      decision,
		PolicyID:      policy.ID,
		PolicyVersion: policy.Version,
		ResourceID:    context.Resource.ID,
		Evaluations:   evaluations,
	}, nil
}

func combine(
	algorithm model.CombiningAlgorithm,
	evaluations []model.RuleEvaluation,
) (model.Decision, error) {

	switch algorithm {

	case model.CombiningAllOf:
		for _, evaluation := range evaluations {
			if !evaluation.Result {
				return model.DecisionDeny, nil
			}
		}

		return model.DecisionPermit, nil

	case model.CombiningAnyOf:
		for _, evaluation := range evaluations {
			if evaluation.Result {
				return model.DecisionPermit, nil
			}
		}

		return model.DecisionDeny, nil

	default:
		return model.DecisionDeny, fmt.Errorf(
			"unsupported combining algorithm: %s",
			algorithm,
		)
	}
}
