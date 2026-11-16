from typing import Any

from access_control.backend import (
    AccessControlBackend,
)
from access_control.fabric.gateway_client import (
    FabricGatewayClient,
)
from access_control.models import (
    AccessDecision,
    RuleEvaluation,
)


class FabricBackend(AccessControlBackend):

    def __init__(
        self,
        gateway_url: str,
    ):
        self.gateway = FabricGatewayClient(
            base_url=gateway_url,
        )

    def evaluate_access(
        self,
        context: dict[str, Any],
    ) -> AccessDecision:

        response = self.gateway.evaluate_access(
            context
        )

        evaluations = [
            RuleEvaluation(
                rule_id=evaluation['ruleId'],
                result=evaluation['result'],
            )
            for evaluation
            in response.get('evaluations', [])
        ]

        return AccessDecision(
            decision=response['decision'],
            resource_id=response.get(
                'resourceId'
            ),
            policy_reference=response.get(
                'policyId'
            ),
            policy_version=response.get(
                'policyVersion'
            ),
            evaluations=evaluations,
        )

    def protect_resource(
        self,
        uri: str,
        content_hash: str,
        policy_reference: str,
    ) -> dict[str, Any]:

        raise NotImplementedError(
            'Fabric resource protection '
            'is not implemented yet'
        )

    def get_resource_policy(
        self,
        uri: str,
    ) -> str:

        raise NotImplementedError(
            'Fabric resource policy retrieval '
            'is not implemented yet'
        )