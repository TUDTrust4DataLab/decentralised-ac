from abc import ABC, abstractmethod
from typing import Any


class AuthenticationHandler(ABC):

    @abstractmethod
    def validate_token(
        self,
        token: str,
    ) -> bool:
        """
        Validate an authentication token.
        """
        raise NotImplementedError

    @abstractmethod
    def get_user_info(
        self,
        token: str,
    ) -> dict[str, Any]:
        """
        Retrieve trusted identity information associated
        with an authenticated subject.
        """
        raise NotImplementedError