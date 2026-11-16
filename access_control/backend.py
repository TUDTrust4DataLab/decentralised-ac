from abc import ABC, abstractmethod
from typing import Any

from access_control.models import AccessDecision


class AccessControlBackend(ABC):

    @abstractmethod
    def protect_resource(
        self,
        uri: str,
        content_hash: str,
        policy_reference: str,
    ) -> dict[str, Any]:
        """
        Register a protected resource and associate it with a policy.
        """
        raise NotImplementedError

    @abstractmethod
    def get_resource_policy(
        self,
        uri: str,
    ) -> str:
        """
        Retrieve the policy associated with a resource.
        """
        raise NotImplementedError

    @abstractmethod
    def evaluate_access(
            self,
            context: dict[str, Any],
    ) -> AccessDecision:
        """
        Evaluate an access request and return a
        backend-independent decision.
        """
        raise NotImplementedError