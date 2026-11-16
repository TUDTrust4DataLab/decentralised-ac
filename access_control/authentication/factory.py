from typing import Any

from access_control.authentication.handler import (
    AuthenticationHandler,
)
from access_control.authentication.keycloak import (
    KeycloakAuthenticationHandler,
)


def create_authentication_handler(
    provider_name: str,
    keycloak_config: dict[str, Any],
) -> AuthenticationHandler:

    provider_name = provider_name.strip().lower()

    if provider_name == 'keycloak':
        return KeycloakAuthenticationHandler(
            keycloak_config=keycloak_config,
        )

    supported_providers = [
        'keycloak',
    ]

    raise ValueError(
        f"Unsupported authentication provider "
        f"'{provider_name}'. "
        f"Supported providers: "
        f"{', '.join(supported_providers)}"
    )