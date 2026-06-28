import type { IgdbJogo, IgdbJogoDetalhe } from "@memory-card/types";

const rawgBaseUrl = "https://api.rawg.io/api";
const steamCdnBaseUrl = "https://cdn.cloudflare.steamstatic.com/steam/apps";
const cacheTtlMs = 5 * 60 * 1000;
const cache = new Map<string, { data: unknown; expira: number }>();

type RawgPlatform = {
  platform?: {
    name?: string;
  };
};

type RawgGenre = {
  name?: string;
};

type RawgStore = {
  store?: {
    slug?: string;
  };
  store_id?: number;
  url?: string | null;
};

type RawgGame = {
  id: number;
  name?: string;
  background_image?: string | null;
  released?: string | null;
  platforms?: RawgPlatform[];
  genres?: RawgGenre[];
  stores?: RawgStore[];
  description_raw?: string | null;
};

type RawgSearchResponse = {
  results?: RawgGame[];
};

type RawgStoresResponse = {
  results?: RawgStore[];
};

export class RawgNotFoundError extends Error {
  constructor() {
    super("Jogo não encontrado na RAWG");
  }
}

export function hasRawgApiKey() {
  return Boolean(process.env.RAWG_API_KEY);
}

function getCached<T>(key: string): T | null {
  const entry = cache.get(key);

  if (!entry || Date.now() > entry.expira) {
    return null;
  }

  return entry.data as T;
}

function setCached(key: string, data: unknown): void {
  cache.set(key, { data, expira: Date.now() + cacheTtlMs });
}

function getRawgApiKey() {
  return process.env.RAWG_API_KEY ?? "";
}

function getReleasedYear(released?: string | null) {
  if (!released) {
    return null;
  }

  const year = new Date(`${released}T00:00:00.000Z`).getUTCFullYear();
  return Number.isNaN(year) ? null : year;
}

function getPlatforms(platforms?: RawgPlatform[]) {
  return platforms?.map((item) => item.platform?.name).filter((name): name is string => Boolean(name)) ?? [];
}

function getGenres(genres?: RawgGenre[]) {
  return genres?.map((genre) => genre.name).filter((name): name is string => Boolean(name)) ?? [];
}

function getSteamCover(stores?: RawgStore[]) {
  const steam = stores?.find((store) => store.store?.slug === "steam" || store.store_id === 1);

  if (!steam?.url) {
    return null;
  }

  const match = steam.url.match(/\/app\/(\d+)/);

  if (!match) {
    return null;
  }

  return `${steamCdnBaseUrl}/${match[1]}/library_600x900.jpg`;
}

function mapGame(game: RawgGame, capaUrl = game.background_image ?? null): IgdbJogo {
  return {
    igdbId: game.id,
    nome: game.name ?? "Sem nome",
    capaUrl,
    ano: getReleasedYear(game.released),
    generos: getGenres(game.genres),
    plataformas: getPlatforms(game.platforms)
  };
}

async function requestRawg<T>(path: string, params?: Record<string, string>): Promise<T> {
  const searchParams = new URLSearchParams({
    key: getRawgApiKey(),
    ...params
  });
  const response = await fetch(`${rawgBaseUrl}${path}?${searchParams.toString()}`);

  if (response.status === 404) {
    throw new RawgNotFoundError();
  }

  if (!response.ok) {
    throw new Error("Falha ao consultar a RAWG");
  }

  return (await response.json()) as T;
}

async function buscarCapaSteam(id: number) {
  try {
    const data = await requestRawg<RawgStoresResponse>(`/games/${id}/stores`);
    return getSteamCover(data.results);
  } catch {
    return null;
  }
}

export async function buscarJogos(query: string): Promise<IgdbJogo[]> {
  const key = `search:${query.trim().toLowerCase()}`;
  const cached = getCached<IgdbJogo[]>(key);

  if (cached) {
    return cached;
  }

  try {
    const data = await requestRawg<RawgSearchResponse>("/games", {
      search: query,
      page_size: "8"
    });
    const jogos = data.results?.map((game) => mapGame(game)) ?? [];
    setCached(key, jogos);

    return jogos;
  } catch {
    return [];
  }
}

export async function buscarDetalhes(id: number): Promise<IgdbJogoDetalhe> {
  const key = `details:${id}`;
  const cached = getCached<IgdbJogoDetalhe>(key);

  if (cached) {
    return cached;
  }

  try {
    const game = await requestRawg<RawgGame>(`/games/${id}`);
    const capaUrl = getSteamCover(game.stores) ?? (await buscarCapaSteam(id)) ?? game.background_image ?? null;
    const detalhe: IgdbJogoDetalhe = {
      ...mapGame(game, capaUrl),
      descricao: game.description_raw ?? null,
      desenvolvedor: null
    };
    setCached(key, detalhe);

    return detalhe;
  } catch (error) {
    if (error instanceof RawgNotFoundError) {
      throw error;
    }

    throw new RawgNotFoundError();
  }
}
