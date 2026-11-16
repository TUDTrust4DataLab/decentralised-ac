import { fabricConfig } from '../config/fabric.js';
import { TransactionService } from './transaction.service.js';

export interface RequestContext {
  subject: {
    id?: string;
    attributes?: Record<string, unknown>;
    claims?: Record<
      string,
      {
        verified: boolean;
      }
    >;
  };

  resource: {
    id: string;
    attributes?: Record<string, unknown>;
  };

  action: {
    name: string;
    attributes?: Record<string, unknown>;
  };

  environment: {
    attributes?: Record<string, unknown>;
  };
}

export interface RuleEvaluation {
  ruleId: string;
  result: boolean;
}

export interface AccessDecision {
  decision: 'permit' | 'deny';
  policyId: string;
  policyVersion: number;
  resourceId: string;
  evaluations: RuleEvaluation[];
}

export class AccessService {
  constructor(
    private readonly transactionService: TransactionService,
  ) {}

  async evaluate(
    context: RequestContext,
  ): Promise<AccessDecision> {

    const result =
      await this.transactionService.evaluate(
        fabricConfig.channelName,
        fabricConfig.chaincodeName,
        'pdp',
        'EvaluateRequest',
        [
          JSON.stringify(context),
        ],
      );

    const text =
      Buffer.from(result).toString('utf8');

    return JSON.parse(text) as AccessDecision;
  }
}