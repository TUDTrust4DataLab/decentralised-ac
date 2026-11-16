import {fabricConfig} from '../config/fabric.js';
import {TransactionService} from './transaction.service.js';
import {translateFabricError} from "../errors/fabric-error.js";

export interface Resource {
  id: string;
  uri: string;
  owner: string;
  contentHash: string;
  attributes: Record<string, unknown>;

  createdAt?: string;
  createdBy?: string;
  updatedAt?: string;
  updatedBy?: string;
}

export interface Policy {
  id: string;
  name: string;
  description: string;
  version: number;
  status: string;
  combiningAlgorithm: string;
  rules: unknown[];

  createdAt: string;
  createdBy: string;
  updatedAt: string;
  updatedBy: string;
  deactivatedAt: string;
  deactivatedBy: string;
}

export class ResourceService {
  constructor(
    private readonly transactionService: TransactionService,
  ) {
  }

  async get(
    resourceId: string,
  ): Promise<Resource> {
    try {
      const result =
        await this.transactionService.evaluate(
          fabricConfig.channelName,
          fabricConfig.chaincodeName,
          'pip',
          'GetResource',
          [
            resourceId,
          ],
        );

      return JSON.parse(
        Buffer.from(result).toString('utf8'),
      ) as Resource;
    } catch (error) {
      throw translateFabricError(error);
    }
  }

  async create(
    resource: Resource,
  ): Promise<void> {
    try {
      await this.transactionService.submit(
        fabricConfig.channelName,
        fabricConfig.chaincodeName,
        'pip',
        'CreateResource',
        [
          JSON.stringify(resource),
        ],
      );
    } catch (error) {
      throw translateFabricError(error);
    }
  }

  async update(
    resource: Resource,
  ): Promise<void> {
    try {
      await this.transactionService.submit(
        fabricConfig.channelName,
        fabricConfig.chaincodeName,
        'pip',
        'UpdateResource',
        [
          JSON.stringify(resource),
        ],
      );
    } catch (error) {
      throw translateFabricError(error);
    }

  }

  async assignPolicy(
    resourceId: string,
    policyId: string,
  ): Promise<void> {
    try {
      await this.transactionService.submit(
        fabricConfig.channelName,
        fabricConfig.chaincodeName,
        'pap',
        'AssignPolicyToResource',
        [
          resourceId,
          policyId,
        ],
      );
    } catch (error) {
      throw translateFabricError(error);
    }
  }

  async getPolicy(
    resourceId: string,
  ): Promise<Policy> {
    try {
      const result =
        await this.transactionService.evaluate(
          fabricConfig.channelName,
          fabricConfig.chaincodeName,
          'pap',
          'GetPolicyForResource',
          [
            resourceId,
          ],
        );

      return JSON.parse(
        Buffer.from(result).toString('utf8'),
      ) as Policy;
    } catch (error) {
      throw translateFabricError(error);
    }
  }
}