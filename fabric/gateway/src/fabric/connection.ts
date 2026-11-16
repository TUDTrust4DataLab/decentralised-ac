import type { Gateway } from '@hyperledger/fabric-gateway';
import type * as grpc from '@grpc/grpc-js';

import { createGrpcClient } from './grpc.js';
import { createGateway } from './gateway.js';

export interface FabricConnection {
  gateway: Gateway;
  grpcClient: grpc.Client;
}

export async function createFabricConnection():
  Promise<FabricConnection> {

  const grpcClient = await createGrpcClient();

  const gateway = await createGateway(
    grpcClient,
  );

  return {
    gateway,
    grpcClient,
  };
}

export function closeFabricConnection(
  connection: FabricConnection,
): void {

  connection.gateway.close();
  connection.grpcClient.close();
}