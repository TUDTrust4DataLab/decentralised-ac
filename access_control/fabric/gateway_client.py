from typing import Any
from urllib.parse import quote
import requests


class FabricGatewayClient:

    def __init__(
        self,
        base_url: str,
        timeout: int = 10,
    ):
        self.base_url = base_url.rstrip('/')
        self.timeout = timeout

    def _request(
        self,
        method: str,
        path: str,
        **kwargs,
    ) -> dict[str, Any]:

        response = requests.request(
            method=method,
            url=f'{self.base_url}{path}',
            timeout=self.timeout,
            **kwargs,
        )

        if not response.ok:
            try:
                details = response.json()
            except ValueError:
                details = response.text

            raise RuntimeError(
                f'Fabric Gateway request failed '
                f'({response.status_code}): '
                f'{details}'
            )

        if not response.content:
            return {}

        return response.json()

    def evaluate_access(
        self,
        context: dict[str, Any],
    ) -> dict[str, Any]:

        return self._request(
            method='POST',
            path='/access/evaluate',
            json=context,
        )

    def create_resource(
            self,
            resource: dict[str, Any],
    ) -> dict[str, Any]:

        return self._request(
            method='POST',
            path='/resources',
            json=resource,
        )

    def assign_resource_policy(
            self,
            resource_id: str,
            policy_id: str,
    ) -> dict[str, Any]:

        resource_id = quote(
            resource_id,
            safe='',
        )

        policy_id = quote(
            policy_id,
            safe='',
        )

        return self._request(
            method='PUT',
            path=(
                f'/resources/{resource_id}'
                f'/policy/{policy_id}'
            ),
        )

    def get_resource_policy(
            self,
            resource_id: str,
    ) -> dict[str, Any]:

        resource_id = quote(
            resource_id,
            safe='',
        )

        return self._request(
            method='GET',
            path=(
                f'/resources/{resource_id}/policy'
            ),
        )