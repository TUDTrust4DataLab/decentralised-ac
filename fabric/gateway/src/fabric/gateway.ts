import {
  connect,
  type Gateway,
} from '@hyperledger/fabric-gateway';

import * as grpc from '@grpc/grpc-js';

import {
  createIdentity,
  createSigner,
} from './identity.js';

export async function createGateway(
  client: grpc.Client,
): Promise<Gateway> {
  const identity = await createIdentity();
  const signer = await createSigner();

  return connect({
    client,
    identity,
    signer,
  });
}