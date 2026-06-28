import type { FastifyInstance } from "fastify";

import { getPorAno, getPorPlataforma, getRecordes, getTotais } from "../services/dashboard.service.js";

function usuarioId(user: { sub: number }) {
  return user.sub;
}

export async function dashboardRoutes(app: FastifyInstance) {
  app.get("/dashboard/totais", { preHandler: app.authenticate }, async (request) => {
    const totais = await getTotais(usuarioId(request.user));
    return { data: totais };
  });

  app.get("/dashboard/por-ano", { preHandler: app.authenticate }, async (request) => {
    const porAno = await getPorAno(usuarioId(request.user));
    return { data: porAno };
  });

  app.get("/dashboard/por-plataforma", { preHandler: app.authenticate }, async (request) => {
    const porPlataforma = await getPorPlataforma(usuarioId(request.user));
    return { data: porPlataforma };
  });

  app.get("/dashboard/recordes", { preHandler: app.authenticate }, async (request) => {
    const recordes = await getRecordes(usuarioId(request.user));
    return { data: recordes };
  });
}
