import { readFile } from 'node:fs/promises';
import * as crypto from 'node:crypto';

import {
  type Identity,
  type Signer,
  signers,
} from '@hyperledger/fabric-gateway';

import { fabricConfig } from '../config/fabric.js';

export async function createIdentity(): Promise<Identity> {
  const credentials = await readFile(
    fabricConfig.identity.certPath,
  );

  return {
    mspId: fabricConfig.mspId,
    credentials,
  };
}

export async function createSigner(): Promise<Signer> {
  const privateKeyPem = await readFile(
    fabricConfig.identity.keyPath,
  );

  const privateKey = crypto.createPrivateKey(privateKeyPem);

  return signers.newPrivateKeySigner(privateKey);
}