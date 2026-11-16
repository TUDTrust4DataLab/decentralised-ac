import type { FastifyInstance } from 'fastify';

import {
  AccessService,
  type RequestContext,
} from '../services/access.service.js';

const accessEvaluationSchema = {
  body: {
    type: 'object',
    required: [
      'subject',
      'resource',
      'action',
      'environment',
    ],
    properties: {
      subject: {
        type: 'object',
        required: [
          'attributes',
        ],
        properties: {
          id: {
            type: 'string',
          },

          attributes: {
            type: 'object',
            additionalProperties: true,
          },

          claims: {
            type: 'object',
            additionalProperties: {
              type: 'object',
              required: [
                'verified',
              ],
              properties: {
                verified: {
                  type: 'boolean',
                },
              },
            },
          },
        },
      },

      resource: {
        type: 'object',
        required: [
          'id',
        ],
        properties: {
          id: {
            type: 'string',
            minLength: 1,
          },

          attributes: {
            type: 'object',
            additionalProperties: true,
          },
        },
      },

      action: {
        type: 'object',
        required: [
          'name',
        ],
        properties: {
          name: {
            type: 'string',
            minLength: 1,
          },

          attributes: {
            type: 'object',
            additionalProperties: true,
          },
        },
      },

      environment: {
        type: 'object',
        properties: {
          attributes: {
            type: 'object',
            additionalProperties: true,
          },
        },
      },
    },
  },
};

export async function registerAccessRoutes(
  app: FastifyInstance,
  accessService: AccessService,
) {
  app.post<{
    Body: RequestContext;
  }>(
    '/access/evaluate',
    {
      schema: accessEvaluationSchema,
    },
    async (request, reply) => {

      try {
        const decision =
          await accessService.evaluate(
            request.body,
          );

        return decision;

      } catch (error) {
        request.log.error(error);

        return reply
          .code(500)
          .send({
            error:
              'Access evaluation failed',
            details:
              error instanceof Error
                  ? error.message
                  : String(error),
          });
      }
    },
  );
}