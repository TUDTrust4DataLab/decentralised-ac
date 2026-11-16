import type { FastifyInstance } from 'fastify';

import { fabricConfig } from '../config/fabric.js';
import { TransactionService } from '../services/transaction.service.js';

interface TransactionBody {
  contract: string;
  transaction: string;
  args?: string[];
}

function decodeResult(result: Uint8Array): unknown {
  if (result.length === 0) {
    return null;
  }

  const text = Buffer.from(result).toString('utf8');

  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

export async function registerTransactionRoutes(
  app: FastifyInstance,
  transactionService: TransactionService,
) {
  app.post<{
    Body: TransactionBody;
  }>(
    '/transactions/evaluate',
    async (request, reply) => {

      const {
        contract,
        transaction,
        args = [],
      } = request.body;

      try {
        const result =
          await transactionService.evaluate(
            fabricConfig.channelName,
            fabricConfig.chaincodeName,
            contract,
            transaction,
            args,
          );

        return {
          result: decodeResult(result),
        };

      } catch (error) {
        request.log.error(error);

        return reply
          .code(500)
          .send({
            error: 'Fabric transaction evaluation failed',
          });
      }
    },
  );

  app.post<{
    Body: TransactionBody;
  }>(
    '/transactions/submit',
    async (request, reply) => {

      const {
        contract,
        transaction,
        args = [],
      } = request.body;

      try {
        const result =
          await transactionService.submit(
            fabricConfig.channelName,
            fabricConfig.chaincodeName,
            contract,
            transaction,
            args,
          );

        return {
          result: decodeResult(result),
        };

      } catch (error) {
        request.log.error(error);

        return reply
          .code(500)
          .send({
            error: 'Fabric transaction submission failed',
          });
      }
    },
  );
}