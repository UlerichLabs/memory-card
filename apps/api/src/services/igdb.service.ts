import type { IgdbJogo, IgdbJogoDetalhe } from "@memory-card/types";

type TokenState = {
  accessToken: string;
  expiresAt: number;
};

type IgdbImage = {
  url?: string;
};

type IgdbNamed = {
  name?: string;
};

type IgdbCompany = {
  company?: IgdbNamed;
  developer?: boolean;
};

type IgdbGameResponse = {
  id: number;
  name?: string;
  cover?: IgdbImage;
  summary?: string;
  first_release_date?: number;
  genres?: IgdbNamed[];
  platforms?: IgdbNamed[];
  involved_companies?: IgdbCompany[];
};

type TwitchTokenResponse = {
  access_token: string;
  expires_in: number;
};

const tokenRenewMarginMs = 60_000;
const cacheTtlMs = 5 * 60_000;
const cache = new Map<string, { data: IgdbJogo[]; expiresAt: number }>();

let tokenState: TokenState | null = null;

function normalizeCoverUrl(url?: string, size = "t_cover_big") {
  if (!url) {
    return null;
  }

  const fullUrl = url.startsWith("//") ? `https:${url}` : url;
  return fullUrl.replace("t_thumb", size);
}

function mapGame(game: IgdbGameResponse): IgdbJogo {
  return {
    igdbId: game.id,
    nome: game.name ?? "Sem nome",
    capaUrl: normalizeCoverUrl(game.cover?.url),
    ano: game.first_release_date ? new Date(game.first_release_date * 1000).getUTCFullYear() : null,
    generos: game.genres?.map((genre) => genre.name).filter((name): name is string => Boolean(name)) ?? [],
    plataformas:
      game.platforms?.map((platform) => platform.name).filter((name): name is string => Boolean(name)) ?? []
  };
}

function getIgdbCredentials() {
  const clientId = process.env.IGDB_CLIENT_ID;
  const clientSecret = process.env.IGDB_CLIENT_SECRET;

  if (!clientId || !clientSecret) {
    throw new Error("Credenciais IGDB não configuradas");
  }

  return { clientId, clientSecret };
}

async function getAccessToken() {
  if (tokenState && tokenState.expiresAt - tokenRenewMarginMs > Date.now()) {
    return tokenState.accessToken;
  }

  const { clientId, clientSecret } = getIgdbCredentials();
  const params = new URLSearchParams({
    client_id: clientId,
    client_secret: clientSecret,
    grant_type: "client_credentials"
  });

  const response = await fetch(`https://id.twitch.tv/oauth2/token?${params.toString()}`, {
    method: "POST"
  });

  if (!response.ok) {
    throw new Error("Falha ao autenticar na IGDB");
  }

  const payload = (await response.json()) as TwitchTokenResponse;
  tokenState = {
    accessToken: payload.access_token,
    expiresAt: Date.now() + payload.expires_in * 1000
  };

  return tokenState.accessToken;
}

async function queryIgdb<T>(endpoint: string, body: string): Promise<T[]> {
  const { clientId } = getIgdbCredentials();
  const accessToken = await getAccessToken();
  const response = await fetch(`https://api.igdb.com/v4/${endpoint}`, {
    method: "POST",
    headers: {
      "Client-ID": clientId,
      Authorization: `Bearer ${accessToken}`,
      "Content-Type": "text/plain"
    },
    body
  });

  if (!response.ok) {
    throw new Error("Falha ao consultar a IGDB");
  }

  return (await response.json()) as T[];
}

export async function buscarJogos(q: string): Promise<IgdbJogo[]> {
  const key = q.trim().toLowerCase();
  const cached = cache.get(key);

  if (cached && cached.expiresAt > Date.now()) {
    return cached.data;
  }

  const data = await queryIgdb<IgdbGameResponse>(
    "games",
    `fields id,name,cover.url,first_release_date,genres.name,platforms.name; search "${q.replaceAll('"', '\\"')}"; limit 8;`
  );
  const jogos = data.map(mapGame);
  cache.set(key, { data: jogos, expiresAt: Date.now() + cacheTtlMs });

  return jogos;
}

export async function buscarJogoPorId(igdbId: number): Promise<IgdbJogoDetalhe> {
  const data = await queryIgdb<IgdbGameResponse>(
    "games",
    `fields id,name,cover.url,summary,first_release_date,genres.name,platforms.name,involved_companies.company.name,involved_companies.developer; where id = ${igdbId}; limit 1;`
  );
  const game = data[0];

  if (!game) {
    throw new Error("Jogo não encontrado na IGDB");
  }

  const desenvolvedor =
    game.involved_companies?.find((company) => company.developer)?.company?.name ?? null;

  return {
    ...mapGame(game),
    descricao: game.summary ?? null,
    desenvolvedor
  };
}

export async function buscarCapa(coverId: number): Promise<string> {
  const data = await queryIgdb<IgdbImage>("covers", `fields url; where id = ${coverId}; limit 1;`);
  const url = normalizeCoverUrl(data[0]?.url, "t_cover_big_2x");

  if (!url) {
    throw new Error("Capa não encontrada na IGDB");
  }

  return url;
}
