import Fastify from 'fastify';

import type {TransactionService} from './services/transaction.service.js';
import {AccessService} from './services/access.service.js';
import {PolicyService} from "./services/policy.service.js";
import {ResourceService} from './services/resource.service.js';

import {registerTransactionRoutes} from './routes/transactions.js';
import {registerAccessRoutes} from './routes/access.js';
import {registerResourceRoutes} from './routes/resources.js';
import {registerPolicyRoutes} from "./routes/policies.js";

import {HttpError} from './errors/http-error.js';

export function buildApp(
  transactionService: TransactionService,
) {
  const app = Fastify({
    logger: true,
  });

  app.setErrorHandler(
    (error, request, reply) => {

      request.log.error(error);

      if (error instanceof HttpError) {
        return reply
          .code(error.statusCode)
          .send({
            error: error.message,
          });
      }

      if (error.validation) {
        return reply
          .code(400)
          .send({
            error: 'Invalid request',
            details: error.validation,
          });
      }

      return reply
        .code(500)
        .send({
          error: 'Internal server error',
        });
    },
  );

  const accessService =
    new AccessService(
      transactionService,
    );

  const policyService =
    new PolicyService(
      transactionService,
    );

  const resourceService =
    new ResourceService(
      transactionService,
    );

  app.get('/health', async () => {
    return {
      status: 'ok',
    };
  });

  if (
    process.env.ENABLE_GENERIC_TRANSACTIONS ===
    'true'
  ) {
    registerTransactionRoutes(
      app,
      transactionService,
    );
  }

  registerAccessRoutes(
    app,
    accessService,
  );

  registerPolicyRoutes(
    app,
    policyService,
  );

  registerResourceRoutes(
    app,
    resourceService,
  );

  return app;
}