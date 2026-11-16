import 'dotenv/config';

import { buildApp } from './app.js';
import {
  createFabricConnection,
  closeFabricConnection,
} from './fabric/connection.js';
import { TransactionService } from './services/transaction.service.js';

const port = Number(
  process.env.PORT ?? 5524,
);

const host =
  process.env.HOST ?? '0.0.0.0';

const connection =
  await createFabricConnection();

const transactionService =
  new TransactionService(
    connection.gateway,
  );

const app = buildApp(
  transactionService,
);

app.addHook(
  'onClose',
  async () => {
    closeFabricConnection(
      connection,
    );
  },
);

try {
  await app.listen({
    port,
    host,
  });

} catch (error) {
  app.log.error(error);

  closeFabricConnection(
    connection,
  );

  process.exit(1);
}