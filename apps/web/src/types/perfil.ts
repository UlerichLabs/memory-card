export type AvatarTipo = 'upload' | 'jogo' | 'iniciais' | null

export interface JogoFavoritoDTO {
  id: number
  nome: string
  capa_url: string | null
}

export interface PerfilDTO {
  id?: number
  nome: string
  email?: string
  username: string | null
  bio: string | null
  avatar_url: string | null
  foto_url?: string | null
  avatar_tipo: AvatarTipo
  jogo_favorito: JogoFavoritoDTO | null
  console_favorito: string | null
  jogando_desde: number | null
  jogando_desde_ano?: number | null
  jogando_desde_efetivo: number | null
  created_at?: string
  updated_at?: string
}

export interface AtualizarPerfilPayload {
  nome: string
  username: string | null
  bio: string | null
  jogo_favorito_id: number | null
  console_favorito: string | null
  jogando_desde: number | null
}
