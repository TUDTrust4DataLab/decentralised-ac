import type {FastifyInstance} from 'fastify';

import {
  PolicyService,
  type Policy,
} from '../services/policy.service.js';

const policySchema = {
  type: 'object',
  required: [
    'id',
    'name',
    'combiningAlgorithm',
    'rules',
  ],
  properties: {
    id: {
      type: 'string',
      minLength: 1,
    },

    name: {
      type: 'string',
      minLength: 1,
    },

    description: {
      type: 'string',
    },

    combiningAlgorithm: {
      type: 'string',
      enum: [
        'allOf',
        'anyOf',
      ],
    },

    rules: {
      type: 'array',
      minItems: 1,

      items: {
        type: 'object',

        required: [
          'id',
          'source',
          'attribute',
          'operator',
        ],

        properties: {
          id: {
            type: 'string',
            minLength: 1,
          },

          description: {
            type: 'string',
          },

          source: {
            type: 'string',
            enum: [
              'subject',
              'claim',
              'resource',
              'action',
              'environment',
            ],
          },

          attribute: {
            type: 'string',
            minLength: 1,
          },

          operator: {
            type: 'string',
            enum: [
              'equals',
              'notEquals',
              'in',
              'notIn',
              'anyOf',
              'allOf',
              'greaterThan',
              'lessThan',
              'between',
              'verified',
            ],
          },

          value: {},
        },
      },
    },
  },
};

export async function registerPolicyRoutes(
  app: FastifyInstance,
  policyService: PolicyService,
) {

  app.get<{
    Params: {
      id: string;
    };
  }>(
    '/policies/:id',
    async (request) => {
      return policyService.get(
        request.params.id,
      );
    },
  );

  app.post<{
    Body: Policy;
  }>(
    '/policies',
    {
      schema: {
        body: policySchema,
      },
    },
    async (request, reply) => {

      await policyService.create(
        request.body,
      );

      return reply
        .code(201)
        .send({
          status: 'created',
        });
    },
  );

  app.put<{
    Params: {
      id: string;
    };
    Body: Policy;
  }>(
    '/policies/:id',
    {
      schema: {
        body: policySchema,
      },
    },
    async (request, reply) => {

      if (
        request.body.id !==
        request.params.id
      ) {
        return reply
          .code(400)
          .send({
            error:
              'Policy ID does not match URL',
          });
      }
      await policyService.update(
        request.body,
      );

      return {
        status: 'updated',
      };
    },
  );

  app.delete<{
    Params: {
      id: string;
    };
  }>(
    '/policies/:id',
    async (request) => {

      await policyService.delete(
        request.params.id,
      );

      return {
        status: 'deactivated',
      };
    },
  );
}