import type { FastifyInstance } from "fastify";
import { z } from "zod";
import type { JwtPayload } from "@memory-card/types";
import {
  AuthConflictError,
  AuthInvalidCredentialsError,
  autenticarUsuario,
  buscarUsuarioPorId,
  registrarUsuario
} from "../services/auth.service.js";

const registerSchema = z.object({
  nome: z.string().min(1).max(100),
  username: z.string().min(1).max(50),
  email: z.string().email().max(255),
  senha: z.string().min(8)
});

const loginSchema = z.object({
  email: z.string().email(),
  senha: z.string().min(1)
});

export async function authRoutes(app: FastifyInstance) {
  app.post("/auth/register", async (request, reply) => {
    const input = registerSchema.parse(request.body);

    try {
      const usuario = await registrarUsuario(input);
      return reply.code(201).send({ data: usuario });
    } catch (error) {
      if (error instanceof AuthConflictError) {
        return reply.code(409).send({ error: { message: error.message } });
      }

      throw error;
    }
  });

  app.post("/auth/login", async (request, reply) => {
    const input = loginSchema.parse(request.body);

    try {
      const usuario = await autenticarUsuario(input.email, input.senha);
      const payload: JwtPayload = {
        sub: usuario.id,
        usuario_id: usuario.id,
        email: usuario.email,
        username: usuario.username
      };
      const accessToken = app.jwt.sign(payload, { expiresIn: "7d" });

      return { data: { usuario, accessToken } };
    } catch (error) {
      if (error instanceof AuthInvalidCredentialsError) {
        return reply.code(401).send({ error: { message: error.message } });
      }

      throw error;
    }
  });

  app.get("/auth/me", { preHandler: app.authenticate }, async (request, reply) => {
    const usuario = await buscarUsuarioPorId(request.user.sub);

    if (!usuario) {
      return reply.code(401).send({ error: { message: "Token inválido ou ausente" } });
    }

    return { data: usuario };
  });
}
