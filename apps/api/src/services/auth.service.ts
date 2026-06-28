import bcrypt from "bcrypt";
import { eq, or } from "drizzle-orm";
import { usuarios } from "@memory-card/db/schema";
import type { Usuario } from "@memory-card/types";
import { db } from "../db.js";

export class AuthConflictError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "AuthConflictError";
  }
}

export class AuthInvalidCredentialsError extends Error {
  constructor() {
    super("Email ou senha inválidos");
    this.name = "AuthInvalidCredentialsError";
  }
}

type RegisterInput = {
  nome: string;
  username: string;
  email: string;
  senha: string;
};

const usuarioColumns = {
  id: usuarios.id,
  nome: usuarios.nome,
  username: usuarios.username,
  email: usuarios.email,
  avatar: usuarios.avatar,
  bio: usuarios.bio,
  createdAt: usuarios.createdAt,
  updatedAt: usuarios.updatedAt
};

export async function registrarUsuario(input: RegisterInput): Promise<Usuario> {
  const [existente] = await db
    .select({
      email: usuarios.email,
      username: usuarios.username
    })
    .from(usuarios)
    .where(or(eq(usuarios.email, input.email), eq(usuarios.username, input.username)))
    .limit(1);

  if (existente?.email === input.email) {
    throw new AuthConflictError("Email já cadastrado");
  }

  if (existente?.username === input.username) {
    throw new AuthConflictError("Username já cadastrado");
  }

  const senhaHash = await bcrypt.hash(input.senha, 12);
  const [usuario] = await db
    .insert(usuarios)
    .values({
      nome: input.nome,
      username: input.username,
      email: input.email,
      senhaHash
    })
    .returning(usuarioColumns);

  return usuario;
}

export async function autenticarUsuario(email: string, senha: string): Promise<Usuario> {
  const [usuarioComSenha] = await db
    .select()
    .from(usuarios)
    .where(eq(usuarios.email, email))
    .limit(1);

  if (!usuarioComSenha) {
    throw new AuthInvalidCredentialsError();
  }

  const senhaValida = await bcrypt.compare(senha, usuarioComSenha.senhaHash);

  if (!senhaValida) {
    throw new AuthInvalidCredentialsError();
  }

  return {
    id: usuarioComSenha.id,
    nome: usuarioComSenha.nome,
    username: usuarioComSenha.username,
    email: usuarioComSenha.email,
    avatar: usuarioComSenha.avatar,
    bio: usuarioComSenha.bio,
    createdAt: usuarioComSenha.createdAt,
    updatedAt: usuarioComSenha.updatedAt
  };
}

export async function buscarUsuarioPorId(id: number): Promise<Usuario | null> {
  const [usuario] = await db
    .select(usuarioColumns)
    .from(usuarios)
    .where(eq(usuarios.id, id))
    .limit(1);

  return usuario ?? null;
}
