import type { Usuario } from "@memory-card/types";
import { api } from "@/lib/api";

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

export function loginRequest(email: string, senha: string) {
  return api.post<LoginResponse>("/api/v1/auth/login", { email, senha });
}

export function meRequest(accessToken: string) {
  return api.get<Usuario>("/api/v1/auth/me", accessToken);
}

export function registerRequest(input: RegisterInput) {
  return api.post<Usuario>("/api/v1/auth/register", input);
}
