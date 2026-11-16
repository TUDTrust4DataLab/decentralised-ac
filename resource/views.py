from flask import request
from flask_restx import Namespace, Resource

from api import settings
from access_control.authentication.factory import create_authentication_handler
from keycloak_interface.utils.functions import extract_header_token
from access_control.factory import create_access_control_backend
from access_control.context import ContextHandler

api: Namespace = Namespace('Resources', description='ExtremeXP Resource Management Endpoints')

access_control_backend = create_access_control_backend(
        backend_name=settings.ACCESS_CONTROL_BACKEND,
        blockchain_config=settings.BLOCKCHAIN_CONFIG,
        fabric_gateway_url=settings.FABRIC_GATEWAY_URL
    )

authentication_handler = create_authentication_handler(
    provider_name=settings.AUTHENTICATION_PROVIDER,
    keycloak_config=settings.KEYCLOAK_CONFIG,
)

context_handler = ContextHandler()

@api.route("/protect")
class ProtectResourceView(Resource):
    @api.doc('protect', params={
        'uri': 'Resource URI',
        'content_hash': 'Resource Content Hash (can be 0x0 if not applicable)',
        'policy_address': 'Policy address hash in the blockchain'
    })
    def post(self):
        token = extract_header_token(request)
        if not authentication_handler.validate_token(token):
            return {'error': 'Invalid or missing token'}, 401
        try:
            payload = request.json

            if not payload.get('uri') or not payload.get('content_hash') or not payload.get('policy_address'):
                return {'error': 'Missing required fields'}, 400

            response = (
                access_control_backend.protect_resource(
                    uri=payload.get('uri'),
                    content_hash=payload.get(
                        'content_hash'
                    ),
                    policy_reference=payload.get(
                        'policy_address'
                    ),
                )
            )

            return response, 201
        except Exception as e:
            return {'error': str(e)}, 401

    @api.doc('protect', params={
        'uri': 'Resource URI'
    }, methods=['GET'])
    def get(self, *args, **kwargs):
        token = extract_header_token(request)
        if not authentication_handler.validate_token(token):
            return {'error': 'Invalid or missing token'}, 401
        try:
            payload = dict(request.args)

            if not payload.get('uri'):
                return {'error': 'Missing required fields'}, 400

            # get resource from PAP
            policy_address = (
                access_control_backend
                .get_resource_policy(
                    payload.get('uri', '0x0')
                )
            )

            response = {
                'policy_address': policy_address
            }

            return response, 200 if int(policy_address, 16) else 404
        except Exception as e:
            return {'error': str(e)}, 401


@api.route("/access")
class AccessResourceView(Resource):
    @api.doc('access', params={
        'uri': 'Resource URI',
        'origin_ip': 'Origin IP Address',
        'scope': 'Scope or Action to be performed',
    })
    def post(self):
        token = extract_header_token(request)
        if not authentication_handler.validate_token(token):
            return {'error': 'Invalid or missing token'}, 401
        try:
            payload = request.json

            if not payload.get('uri') or not payload.get('origin_ip') or not payload.get('scope'):
                return {'error': 'Missing required fields'}, 400

            user_info = authentication_handler.get_user_info(token)

            context = context_handler.build(
                user_info=user_info,
                resource_id=payload.get('uri'),
                action=payload.get('scope'),
                origin_ip=payload.get('origin_ip'),
            )

            decision = access_control_backend.evaluate_access(context)

            return {"grant": decision.granted}, 200 if decision.granted else 403
        except Exception as e:
            return {'error': str(e)}, 403
