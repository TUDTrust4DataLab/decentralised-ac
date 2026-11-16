from typing import Any

from access_control.authentication.handler import (
    AuthenticationHandler,
)

from keycloak_interface.keycloakInterface import (
    KeycloakInterface,
)


class KeycloakAuthenticationHandler(
    AuthenticationHandler
):

    def __init__(
        self,
        keycloak_config: dict[str, Any],
    ):
        self.keycloak = KeycloakInterface(
            server_url=keycloak_config.get('KEYCLOAK_SERVER_URL'),
            realm_name=keycloak_config.get('KEYCLOAK_REALM'),
            client_id=keycloak_config.get('KEYCLOAK_CLIENT_ID'),
            client_secret_key=keycloak_config.get('KEYCLOAK_CLIENT_SECRET_KEY'),
        )

    def validate_token(
        self,
        token: str,
    ) -> bool:

        return self.keycloak.validate_request_token(
            token
        )

    def get_user_info(
        self,
        token: str,
    ) -> dict[str, Any]:

        user_info, status_code = (self.keycloak.userinfo(token))

        if status_code != 200:
            raise ValueError(
                'Unable to retrieve authenticated '
                'user information'
            )

        return user_info