import path from 'node:path';

function required(name: string): string {
  const value = process.env[name];

  if (!value) {
    throw new Error(`Missing required environment variable: ${name}`);
  }

  return value;
}

function resolvePath(filePath: string): string {
  return path.resolve(process.cwd(), filePath);
}

export const fabricConfig = {
  mspId: required('FABRIC_MSP_ID'),

  channelName: required('FABRIC_CHANNEL_NAME'),
  chaincodeName: required('FABRIC_CHAINCODE_NAME'),

  peer: {
    endpoint: required('FABRIC_PEER_ENDPOINT'),
    hostAlias: required('FABRIC_PEER_HOST_ALIAS'),
    tlsCertPath: resolvePath(
      required('FABRIC_TLS_CERT_PATH'),
    ),
  },

  identity: {
    certPath: resolvePath(
      required('FABRIC_CERT_PATH'),
    ),
    keyPath: resolvePath(
      required('FABRIC_KEY_PATH'),
    ),
  },
};