from typing import Any


class ContextHandler:
    """
    Builds a blockchain-independent access-control context.

    The Context Handler combines information obtained from
    identity providers with information from the incoming
    resource request.
    """

    def build(
        self,
        user_info: dict[str, Any],
        resource_id: str,
        action: str,
        origin_ip: str,
    ) -> dict[str, Any]:

        return {
            'subject': {
                'id': user_info.get('sub'),
                'attributes': {
                    # Common/Fabric attributes
                    'role':
                        user_info.get('role'),
                    'groups':
                        user_info.get(
                            'groups',
                            []
                        ),
                    # Legacy EVM attributes
                    'wallet_address':
                        user_info.get(
                            'user_wallet_address'
                        ),
                    'email':
                        user_info.get(
                            'email'
                        ),
                    'location_lat':
                        user_info.get(
                            'user_location_lat'
                        ),
                    'location_long':
                        user_info.get(
                            'user_location_long'
                        ),
                },
                'claims': {},
            },
            'resource': {
                'id': resource_id,
                'attributes': {},
            },
            'action': {
                'name': action,
                'attributes': {},
            },
            'environment': {
                'attributes': {
                    'ip': origin_ip,
                },
            },
        }