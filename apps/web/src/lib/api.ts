import type {
  AtualizarJogoDto,
  CriarJogoDto,
  Dificuldade,
  FiltrosJogoDto,
  IgdbJogo,
  IgdbJogoDetalhe,
  JogoZerado,
  Nota,
  PaginacaoResult
} from "@memory-card/types";

type ApiSuccess<T> = {
  data: T;
};

type ApiError = {
  error: {
    message: string;
    code?: string;
  };
};

export type DashboardJogo = {
  id: number;
  nome: string;
  console: string;
  nota: Nota;
  dificuldade: Dificuldade;
  finalizadoEm: Date;
  tempoJogado: number;
  igdbCapaUrl: string | null;
};

export type DashboardTotais = {
  totalJogos: number;
  totalSegundos: number;
  notaMedia: number;
  totalConsoles: number;
  primeiroZeramento: Date | null;
};

export type DashboardPorAno = {
  ano: number;
  totalJogos: number;
  totalSegundos: number;
  destaque: DashboardJogo | null;
};

export type DashboardPorPlataforma = {
  console: string;
  total: number;
};

export type DashboardRecordes = {
  maisLongo: DashboardJogo | null;
  maisCurto: DashboardJogo | null;
};

const apiBaseUrl = import.meta.env.VITE_API_URL ?? "http://localhost:3001";

export class ApiRequestError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

function isApiError(value: unknown): value is ApiError {
  return typeof value === "object" && value !== null && "error" in value;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const method = init?.method?.toUpperCase();
  const headers = new Headers(init?.headers);
  const token = localStorage.getItem("mc_token");

  if (method === "POST" || method === "PUT") {
    headers.set("Content-Type", "application/json");
  }

  if (token && !headers.has("Authorization")) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${apiBaseUrl}${path}`, {
    ...init,
    headers
  });
  const payload: unknown = response.status === 204 ? null : await response.json();

  if (!response.ok || isApiError(payload)) {
    if (response.status === 401) {
      localStorage.removeItem("mc_token");
      window.location.assign("/login");
    }

    throw new ApiRequestError(isApiError(payload) ? payload.error.message : "Erro na API", response.status);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (payload as ApiSuccess<T>).data;
}

function authHeaders(token = localStorage.getItem("mc_token")) {
  return token ? { Authorization: `Bearer ${token}` } : undefined;
}

export const api = {
  get: <T>(path: string, token?: string) =>
    request<T>(path, {
      headers: authHeaders(token)
    }),
  post: <T>(path: string, body: unknown, token?: string) =>
    request<T>(path, {
      method: "POST",
      body: JSON.stringify(body),
      headers: authHeaders(token)
    }),
  put: <T>(path: string, body: unknown, token?: string) =>
    request<T>(path, {
      method: "PUT",
      body: JSON.stringify(body),
      headers: authHeaders(token)
    }),
  delete: <T>(path: string, token?: string) =>
    request<T>(path, {
      method: "DELETE",
      headers: authHeaders(token)
    })
};

function queryString(filtros?: Partial<FiltrosJogoDto>) {
  const params = new URLSearchParams();

  Object.entries(filtros ?? {}).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== "") {
      params.set(key, String(value));
    }
  });

  const query = params.toString();
  return query ? `?${query}` : "";
}

export async function buscarJogosIgdb(q: string): Promise<IgdbJogo[]> {
  return api.get<IgdbJogo[]>(`/api/v1/igdb/jogos/busca?q=${encodeURIComponent(q)}`);
}

export async function buscarJogoIgdbPorId(id: number): Promise<IgdbJogoDetalhe> {
  return api.get<IgdbJogoDetalhe>(`/api/v1/igdb/jogos/${id}`);
}

export async function criarJogo(dados: CriarJogoDto): Promise<JogoZerado> {
  return api.post<JogoZerado>("/api/v1/jogos", dados);
}

export async function listarJogos(
  filtros?: Partial<FiltrosJogoDto>
): Promise<PaginacaoResult<JogoZerado>> {
  return api.get<PaginacaoResult<JogoZerado>>(`/api/v1/jogos${queryString(filtros)}`);
}

export async function buscarJogo(id: number): Promise<JogoZerado> {
  return api.get<JogoZerado>(`/api/v1/jogos/${id}`);
}

export async function atualizarJogo(id: number, dados: AtualizarJogoDto): Promise<JogoZerado> {
  return api.put<JogoZerado>(`/api/v1/jogos/${id}`, dados);
}

export async function deletarJogo(id: number): Promise<void> {
  return api.delete<void>(`/api/v1/jogos/${id}`);
}

export async function buscarDashboardTotais(): Promise<DashboardTotais> {
  return api.get<DashboardTotais>("/api/v1/dashboard/totais");
}

export async function buscarDashboardPorAno(): Promise<DashboardPorAno[]> {
  return api.get<DashboardPorAno[]>("/api/v1/dashboard/por-ano");
}

export async function buscarDashboardPorPlataforma(): Promise<DashboardPorPlataforma[]> {
  return api.get<DashboardPorPlataforma[]>("/api/v1/dashboard/por-plataforma");
}

export async function buscarDashboardRecordes(): Promise<DashboardRecordes> {
  return api.get<DashboardRecordes>("/api/v1/dashboard/recordes");
}
