from typing import Any

from access_control.backend import AccessControlBackend
from access_control.models import AccessDecision

from blockchain_interface.helpers.utils import transaction_to_dict
from blockchain_interface.interfaces.ABACContracts.PAP import (
    PolicyAdministrationPoint,
)
from blockchain_interface.interfaces.ABACContracts.PDP import (
    PolicyDecisionPoint,
)
from blockchain_interface.interfaces.ABACContracts.PIP import (
    PolicyInformationPoint,
)
from blockchain_interface.interfaces.HyperledgerBesu import (
    HyperledgerBesu,
)


class EVMBackend(AccessControlBackend):

    def __init__(
            self,
            blockchain_config: dict[str, Any],
    ):
        self.evm_interface = HyperledgerBesu(
            blockchain_config.get(
                'BLOCKCHAIN_RPC_URL',
                ''
            )
        ).connect(
            blockchain_config.get(
                'BLOCKCHAIN_PRIVATE_KEY',
                ''
            )
        )

        contracts_root = blockchain_config.get(
            'BLOCKCHAIN_CONTRACTS_ROOT_PATH',
            ''
        )

        self.pip = PolicyInformationPoint(
            self.evm_interface,
            blockchain_config.get(
                'POLICY_INFORMATION_POINT_ADDRESS',
                ''
            ),
            f'{contracts_root}/PIP.sol',
        ).load()

        self.pap = PolicyAdministrationPoint(
            self.evm_interface,
            blockchain_config.get(
                'POLICY_ADMINISTRATION_POINT_ADDRESS',
                ''
            ),
            f'{contracts_root}/PAP.sol',
        ).load()

        self.pdp = PolicyDecisionPoint(
            self.evm_interface,
            blockchain_config.get(
                'POLICY_DECISION_POINT_ADDRESS',
                ''
            ),
            f'{contracts_root}/PDP.sol',
        ).load()

    def protect_resource(
            self,
            uri: str,
            content_hash: str,
            policy_reference: str,
    ) -> dict[str, Any]:
        pip_tx = self.pip.add_resource(
            uri,
            content_hash,
        )

        pap_tx = self.pap.register_resource(
            uri,
            policy_reference,
        )

        return {
            'PIP_transaction':
                transaction_to_dict(pip_tx),

            'PAP_transaction':
                transaction_to_dict(pap_tx),
        }

    def get_resource_policy(
            self,
            uri: str,
    ) -> str:
        return self.pap.get_resource_policy(
            uri
        )

    def evaluate_access(
            self,
            context: dict[str, Any],
    ) -> AccessDecision:
        subject = context['subject']
        subject_attributes = subject['attributes']

        resource = context['resource']
        action = context['action']
        environment = context['environment']

        granted = self.pdp.evaluate_request(
            subject_attributes.get(
                'wallet_address'
            ),
            subject_attributes.get(
                'email'
            ),
            environment['attributes'].get(
                'ip'
            ),
            action.get(
                'name'
            ),
            subject_attributes.get(
                'location_lat'
            ),
            subject_attributes.get(
                'location_long'
            ),
            resource.get(
                'id'
            ),
        )

        return AccessDecision(
            decision=(
                'permit'
                if granted
                else 'deny'
            ),
            resource_id=resource.get('id'),
        )
