import type { FastifyInstance } from "fastify";
import { z } from "zod";

import { AtualizarJogoSchema, CriarJogoSchema, FiltrosJogoSchema } from "@memory-card/types";

import {
  atualizarJogo,
  buscarJogoPorId,
  criarJogo,
  deletarJogo,
  DestaqueDuplicadoError,
  JogoNotFoundError,
  listarJogos
} from "../services/jogos.service.js";

const idParamsSchema = z.object({
  id: z.coerce.number().int().positive()
});

function usuarioId(user: { sub: number }) {
  return user.sub;
}

export async function jogosRoutes(app: FastifyInstance) {
  app.post("/jogos", { preHandler: app.authenticate }, async (request, reply) => {
    const input = CriarJogoSchema.parse(request.body);

    try {
      const jogo = await criarJogo(usuarioId(request.user), input);
      return reply.code(201).send({ data: jogo });
    } catch (error) {
      if (error instanceof DestaqueDuplicadoError) {
        return reply.code(409).send({ error: { message: error.message } });
      }

      throw error;
    }
  });

  app.get("/jogos", { preHandler: app.authenticate }, async (request) => {
    const filtros = FiltrosJogoSchema.parse(request.query);
    const jogos = await listarJogos(usuarioId(request.user), filtros);
    return { data: jogos };
  });

  app.get("/jogos/:id", { preHandler: app.authenticate }, async (request, reply) => {
    const params = idParamsSchema.parse(request.params);

    try {
      const jogo = await buscarJogoPorId(usuarioId(request.user), params.id);
      return { data: jogo };
    } catch (error) {
      if (error instanceof JogoNotFoundError) {
        return reply.code(404).send({ error: { message: error.message } });
      }

      throw error;
    }
  });

  app.put("/jogos/:id", { preHandler: app.authenticate }, async (request, reply) => {
    const params = idParamsSchema.parse(request.params);
    const input = AtualizarJogoSchema.parse(request.body);

    try {
      const jogo = await atualizarJogo(usuarioId(request.user), params.id, input);
      return { data: jogo };
    } catch (error) {
      if (error instanceof JogoNotFoundError) {
        return reply.code(404).send({ error: { message: error.message } });
      }

      if (error instanceof DestaqueDuplicadoError) {
        return reply.code(409).send({ error: { message: error.message } });
      }

      throw error;
    }
  });

  app.delete("/jogos/:id", { preHandler: app.authenticate }, async (request, reply) => {
    const params = idParamsSchema.parse(request.params);

    try {
      await deletarJogo(usuarioId(request.user), params.id);
      return reply.code(204).send();
    } catch (error) {
      if (error instanceof JogoNotFoundError) {
        return reply.code(404).send({ error: { message: error.message } });
      }

      throw error;
    }
  });
}
