import 'dotenv/config';

import { TextDecoder } from 'node:util';

import { createGrpcClient } from './fabric/grpc.js';
import { createGateway } from './fabric/gateway.js';
import { TransactionService } from './services/transaction.service.js';

const utf8Decoder = new TextDecoder();

const client = await createGrpcClient();
const gateway = await createGateway(client);

try {
  const transactions = new TransactionService(gateway);

  console.log('Connecting to Fabric...');

  console.log('Creating asset...');

  await transactions.submit(
    'mychannel',
    'basic',
    'CreateAsset',
    [
      'asset1',
      'blue',
      '10',
      'Yago',
      '500',
    ],
  );

  console.log('Asset created successfully.');

  console.log('Reading asset...');

  const result = await transactions.evaluate(
    'mychannel',
    'basic',
    'ReadAsset',
    ['asset1'],
  );

  console.log('Asset retrieved successfully:');
  console.log(utf8Decoder.decode(result));
} finally {
  gateway.close();
  client.close();
}