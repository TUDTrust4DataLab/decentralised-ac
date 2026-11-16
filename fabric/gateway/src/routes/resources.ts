import type { FastifyInstance } from 'fastify';

import {
  ResourceService,
  type Resource,
} from '../services/resource.service.js';

export async function registerResourceRoutes(
  app: FastifyInstance,
  resourceService: ResourceService,
) {

  app.get<{
    Params: {
      id: string;
    };
  }>(
    '/resources/:id',
    async (request, reply) => {

      try {
        return await resourceService.get(
          request.params.id,
        );

      } catch (error) {
        request.log.error(error);

        return reply
          .code(500)
          .send({
            error:
              'Failed to retrieve resource',
          });
      }
    },
  );

  app.post<{
    Body: Resource;
  }>(
    '/resources',
    async (request, reply) => {

      try {
        await resourceService.create(
          request.body,
        );

        return reply
          .code(201)
          .send({
            status: 'created',
          });

      } catch (error) {
        request.log.error(error);

        return reply
          .code(500)
          .send({
            error:
              'Failed to create resource',
          });
      }
    },
  );

  app.put<{
    Params: {
      id: string;
    };

    Body: Resource;
  }>(
    '/resources/:id',
    async (request, reply) => {

      if (
        request.body.id !==
        request.params.id
      ) {
        return reply
          .code(400)
          .send({
            error:
              'Resource ID does not match URL',
          });
      }

      try {
        await resourceService.update(
          request.body,
        );

        return {
          status: 'updated',
        };

      } catch (error) {
        request.log.error(error);

        return reply
          .code(500)
          .send({
            error:
              'Failed to update resource',
          });
      }
    },
  );

  app.put<{
    Params: {
      id: string;
      policyId: string;
    };
  }>(
    '/resources/:id/policy/:policyId',
    async (request, reply) => {

      try {
        await resourceService.assignPolicy(
          request.params.id,
          request.params.policyId,
        );

        return {
          status: 'assigned',
        };

      } catch (error) {
        request.log.error(error);

        return reply
          .code(500)
          .send({
            error:
              'Failed to assign policy to resource',
          });
      }
    },
  );

  app.get<{
    Params: {
      id: string;
    };
  }>(
    '/resources/:id/policy',
    async (request, reply) => {

      try {
        return await resourceService.getPolicy(
          request.params.id,
        );

      } catch (error) {
        request.log.error(error);

        return reply
          .code(500)
          .send({
            error:
              'Failed to retrieve resource policy',
          });
      }
    },
  );
}