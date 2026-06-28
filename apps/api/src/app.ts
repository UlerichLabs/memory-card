import "./env.js";

import cors from "@fastify/cors";
import Fastify from "fastify";
import { ZodError } from "zod";

import { pool } from "./db.js";
import { authPlugin } from "./plugins/auth.js";
import { authRoutes } from "./routes/auth.routes.js";
import { dashboardRoutes } from "./routes/dashboard.routes.js";
import { igdbRoutes } from "./routes/igdb.routes.js";
import { jogosRoutes } from "./routes/jogos.routes.js";
import { hasRawgApiKey } from "./services/rawg.service.js";

type BuildServerOptions = {
  logger?: boolean;
};

export async function buildServer({ logger = true }: BuildServerOptions = {}) {
  const server = Fastify({ logger });

  server.setErrorHandler((error, request, reply) => {
    if (error instanceof ZodError) {
      void reply.code(400).send({
        error: {
          message: "Dados inválidos"
        }
      });
      return;
    }

    request.log.error(error);
    void reply.send(error);
  });

  await server.register(cors, {
    origin: true,
    methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"]
  });

  await server.register(authPlugin);
  await server.register(authRoutes, { prefix: "/api/v1" });
  await server.register(dashboardRoutes, { prefix: "/api/v1" });
  await server.register(igdbRoutes, { prefix: "/api/v1" });
  await server.register(jogosRoutes, { prefix: "/api/v1" });

  if (!hasRawgApiKey()) {
    server.log.warn("RAWG_API_KEY não configurada; busca de jogos retornará vazia");
  }

  server.get("/api/v1/health", async () => {
    await pool.query("select 1");

    return {
      status: "ok"
    };
  });

  return server;
}
