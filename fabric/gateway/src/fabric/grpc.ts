import { readFile } from 'node:fs/promises';

import * as grpc from '@grpc/grpc-js';

import { fabricConfig } from '../config/fabric.js';

export async function createGrpcClient(): Promise<grpc.Client> {
  const tlsRootCert = await readFile(
    fabricConfig.peer.tlsCertPath,
  );

  const tlsCredentials = grpc.credentials.createSsl(tlsRootCert);

  return new grpc.Client(
    fabricConfig.peer.endpoint,
    tlsCredentials,
    {
      'grpc.ssl_target_name_override': fabricConfig.peer.hostAlias,
    },
  );
}