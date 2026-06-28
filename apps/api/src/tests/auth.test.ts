import type { FastifyInstance } from "fastify";
import request, { type Response } from "supertest";

import type { JwtPayload, Usuario } from "@memory-card/types";

import { buildServer } from "../app.js";
import { pool } from "../db.js";

type LoginResponse = {
  usuario: Usuario;
  accessToken: string;
};

type RegisterInput = {
  nome: string;
  username: string;
  email: string;
  senha: string;
};

type ApiSuccess<T> = {
  data: T;
};

const password = "senha1234";
const runId = Math.random().toString(36).slice(2, 10);

let app: FastifyInstance;

function userInput(tag: string): RegisterInput {
  const usernameTag = tag.replace(/_/g, "").slice(0, 24);

  return {
    nome: `Usuario ${tag}`,
    username: `user_${usernameTag}_${runId}`,
    email: `user.${tag}.${runId}@test.com`,
    senha: password
  };
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function dataFrom<T>(response: Response): T {
  const body: unknown = response.body;

  if (!isRecord(body) || !("data" in body)) {
    throw new Error("Resposta sem data");
  }

  return (body as ApiSuccess<T>).data;
}

async function registerUser(input: RegisterInput) {
  return request(app.server).post("/api/v1/auth/register").send(input);
}

async function loginUser(email: string, senha: string) {
  return request(app.server).post("/api/v1/auth/login").send({ email, senha });
}

beforeAll(async () => {
  app = await buildServer({ logger: false });
  await app.ready();
  await pool.query("delete from usuarios where email like '%@test.com%'");
});

afterAll(async () => {
  await pool.query("delete from usuarios where email like '%@test.com%'");
  await app.close();
  await pool.end();
});

describe("POST /api/v1/auth/register", () => {
  it("cria usuário com dados válidos e retorna 201", async () => {
    const response = await registerUser(userInput("register_valid"));

    expect(response.status).toBe(201);
    expect(dataFrom<Usuario>(response)).toMatchObject({
      nome: "Usuario register_valid",
      email: expect.stringContaining("@test.com")
    });
  });

  it("retorna 409 se email já cadastrado", async () => {
    const input = userInput("register_email_duplicate");
    await registerUser(input);

    const response = await registerUser({
      ...userInput("register_email_duplicate_other"),
      email: input.email
    });

    expect(response.status).toBe(409);
  });

  it("retorna 409 se username já cadastrado", async () => {
    const input = userInput("register_username_duplicate");
    await registerUser(input);

    const response = await registerUser({
      ...userInput("register_username_duplicate_other"),
      username: input.username
    });

    expect(response.status).toBe(409);
  });

  it("retorna 400 se senha menor que 8 caracteres", async () => {
    const response = await registerUser({
      ...userInput("register_short_password"),
      senha: "1234567"
    });

    expect(response.status).toBe(400);
  });

  it("retorna 400 se email inválido", async () => {
    const response = await registerUser({
      ...userInput("register_invalid_email"),
      email: "email-invalido"
    });

    expect(response.status).toBe(400);
  });

  it("não retorna senha_hash no response", async () => {
    const response = await registerUser(userInput("register_no_hash"));
    const usuario = dataFrom<Usuario>(response);

    expect(response.status).toBe(201);
    expect(usuario).not.toHaveProperty("senhaHash");
    expect(usuario).not.toHaveProperty("senha_hash");
  });
});

describe("POST /api/v1/auth/login", () => {
  it("retorna JWT válido com credenciais corretas", async () => {
    const input = userInput("login_valid");
    await registerUser(input);

    const response = await loginUser(input.email, input.senha);
    const login = dataFrom<LoginResponse>(response);

    expect(response.status).toBe(200);
    expect(() => app.jwt.verify(login.accessToken)).not.toThrow();
  });

  it("retorna 401 com senha errada", async () => {
    const input = userInput("login_wrong_password");
    await registerUser(input);

    const response = await loginUser(input.email, "senha-incorreta");

    expect(response.status).toBe(401);
  });

  it("retorna 401 com email não cadastrado", async () => {
    const response = await loginUser(`missing.${runId}@test.com`, password);

    expect(response.status).toBe(401);
  });

  it("JWT contém usuario_id e email no payload", async () => {
    const input = userInput("login_payload");
    await registerUser(input);

    const response = await loginUser(input.email, input.senha);
    const login = dataFrom<LoginResponse>(response);
    const payload = app.jwt.decode<JwtPayload>(login.accessToken);

    expect(payload).toMatchObject({
      usuario_id: login.usuario.id,
      email: input.email
    });
  });
});

describe("GET /api/v1/auth/me", () => {
  it("retorna dados do usuário com token válido", async () => {
    const input = userInput("me_valid");
    await registerUser(input);
    const login = dataFrom<LoginResponse>(await loginUser(input.email, input.senha));

    const response = await request(app.server)
      .get("/api/v1/auth/me")
      .set("Authorization", `Bearer ${login.accessToken}`);

    expect(response.status).toBe(200);
    expect(dataFrom<Usuario>(response)).toMatchObject({
      id: login.usuario.id,
      email: input.email
    });
  });

  it("retorna 401 sem token", async () => {
    const response = await request(app.server).get("/api/v1/auth/me");

    expect(response.status).toBe(401);
  });

  it("retorna 401 com token expirado", async () => {
    const token = app.jwt.sign(
      {
        sub: 1,
        usuario_id: 1,
        email: `expired.${runId}@test.com`,
        username: "expired"
      },
      { expiresIn: "-1s" }
    );

    const response = await request(app.server)
      .get("/api/v1/auth/me")
      .set("Authorization", `Bearer ${token}`);

    expect(response.status).toBe(401);
  });

  it("retorna 401 com token malformado", async () => {
    const response = await request(app.server)
      .get("/api/v1/auth/me")
      .set("Authorization", "Bearer token-malformado");

    expect(response.status).toBe(401);
  });
});
