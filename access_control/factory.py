from typing import Any

from access_control.backend import AccessControlBackend

SUPPORTED_BACKENDS = ['evm', 'fabric']

def create_access_control_backend(
    backend_name: str,
    blockchain_config: dict[str, Any],
    fabric_gateway_url: str | None = None,
) -> AccessControlBackend:

    backend_name = backend_name.strip().lower()

    if backend_name == 'evm':
        # Import EVM dependencies only when EVM is selected.
        from access_control.evm.backend import EVMBackend

        return EVMBackend(
            blockchain_config=blockchain_config,
        )

    if backend_name == 'fabric':
        # Fabric has no dependency on Web3/Solidity.
        from access_control.fabric.backend import FabricBackend

        if not fabric_gateway_url:
            raise ValueError(
                'FABRIC_GATEWAY_URL is required '
                'when using the Fabric backend'
            )

        return FabricBackend(
            gateway_url=fabric_gateway_url,
        )

    raise ValueError(
        f"Unsupported access control backend "
        f"'{backend_name}'. "
        f"Supported backends: "
        f"{', '.join(SUPPORTED_BACKENDS)}"
    )