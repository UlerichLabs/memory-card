import type { FastifyInstance } from "fastify";
import { z } from "zod";

import { buscarDetalhes, buscarJogos, RawgNotFoundError } from "../services/rawg.service.js";

const buscaQuerySchema = z.object({
  q: z.string().trim().min(2)
});

const idParamsSchema = z.object({
  id: z.coerce.number().int().positive()
});

export async function igdbRoutes(app: FastifyInstance) {
  app.get("/igdb/jogos/busca", async (request, reply) => {
    const parsed = buscaQuerySchema.safeParse(request.query);

    if (!parsed.success) {
      return reply.code(400).send({ error: { message: "Informe q com pelo menos 2 caracteres" } });
    }

    const jogos = await buscarJogos(parsed.data.q);
    return { data: jogos };
  });

  app.get("/igdb/jogos/:id", async (request, reply) => {
    const params = idParamsSchema.parse(request.params);

    try {
      const jogo = await buscarDetalhes(params.id);
      return { data: jogo };
    } catch (error) {
      if (error instanceof RawgNotFoundError) {
        return reply.code(404).send({ error: { message: error.message } });
      }

      throw error;
    }
  });

  app.get("/igdb/covers/:id", async (request, reply) => {
    const params = idParamsSchema.parse(request.params);

    try {
      const jogo = await buscarDetalhes(params.id);
      return { data: { url: jogo.capaUrl } };
    } catch (error) {
      if (error instanceof RawgNotFoundError) {
        return reply.code(404).send({ error: { message: error.message } });
      }

      throw error;
    }
  });
}
