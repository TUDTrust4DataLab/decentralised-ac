import { fabricConfig } from '../config/fabric.js';
import { TransactionService } from './transaction.service.js';
import { translateFabricError } from '../errors/fabric-error.js';

export interface Policy {
  id: string;
  name: string;
  description?: string;
  version?: number;
  status?: string;
  combiningAlgorithm: string;
  rules: unknown[];

  createdAt?: string;
  createdBy?: string;
  updatedAt?: string;
  updatedBy?: string;
  deactivatedAt?: string;
  deactivatedBy?: string;
}

export class PolicyService {
  constructor(
    private readonly transactionService: TransactionService,
  ) {}

  async get(
    policyId: string,
  ): Promise<Policy> {

    try {
      const result =
        await this.transactionService.evaluate(
          fabricConfig.channelName,
          fabricConfig.chaincodeName,
          'pap',
          'GetPolicy',
          [policyId],
        );

      return JSON.parse(
        Buffer.from(result).toString('utf8'),
      ) as Policy;

    } catch (error) {
      throw translateFabricError(error);
    }
  }

  async create(
    policy: Policy,
  ): Promise<void> {

    try {
      await this.transactionService.submit(
        fabricConfig.channelName,
        fabricConfig.chaincodeName,
        'pap',
        'CreatePolicy',
        [
          JSON.stringify(policy),
        ],
      );

    } catch (error) {
      throw translateFabricError(error);
    }
  }

  async update(
    policy: Policy,
  ): Promise<void> {

     try {
      await this.transactionService.submit(
        fabricConfig.channelName,
        fabricConfig.chaincodeName,
        'pap',
        'UpdatePolicy',
        [
          JSON.stringify(policy),
        ],
      );
    } catch (error) {
      throw translateFabricError(error);
    }
  }

  async delete(
    policyId: string,
  ): Promise<void> {
    try{
      await this.transactionService.submit(
        fabricConfig.channelName,
        fabricConfig.chaincodeName,
        'pap',
        'DeletePolicy',
        [
          policyId,
        ],
      );
    }catch (error){
      throw translateFabricError(error);
    }
  }
}