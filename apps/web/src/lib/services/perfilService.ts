import { apiRequest } from '@/lib/api'
import type { AtualizarPerfilPayload, PerfilDTO } from '@/types/perfil'

export function normalizarPerfil(perfil: PerfilDTO): PerfilDTO {
  let avatarUrl = perfil.avatar_url ?? perfil.foto_url ?? null
  if (avatarUrl && perfil.avatar_tipo === 'upload' && !avatarUrl.startsWith('http')) {
    const baseUrl = (import.meta.env.VITE_API_URL || '').replace(/\/+$/, '')
    const path = avatarUrl.startsWith('/') ? avatarUrl : `/${avatarUrl}`
    avatarUrl = `${baseUrl}${path}`
  }
  const jogandoDesde = perfil.jogando_desde ?? perfil.jogando_desde_ano ?? null
  return {
    ...perfil,
    avatar_url: avatarUrl,
    foto_url: avatarUrl,
    jogando_desde: jogandoDesde,
    jogando_desde_ano: jogandoDesde,
  }
}


export const perfilService = {
  async obter(token?: string, signal?: AbortSignal): Promise<PerfilDTO> {
    const perfil = await apiRequest<PerfilDTO>('/me/perfil', { method: 'GET', signal }, token)
    return normalizarPerfil(perfil)
  },

  async atualizar(payload: AtualizarPerfilPayload, token?: string): Promise<PerfilDTO> {
    const perfil = await apiRequest<PerfilDTO>(
      '/me/perfil',
      { method: 'PUT', body: JSON.stringify(payload) },
      token
    )
    return normalizarPerfil(perfil)
  },

  async enviarFoto(arquivo: File, token?: string): Promise<PerfilDTO> {
    const formData = new FormData()
    formData.append('arquivo', arquivo)
    const perfil = await apiRequest<PerfilDTO>(
      '/me/foto',
      { method: 'PUT', body: formData },
      token
    )
    return normalizarPerfil(perfil)
  },

  async usarCapa(jogoId: number, token?: string): Promise<PerfilDTO> {
    const perfil = await apiRequest<PerfilDTO>(
      '/me/foto/capa',
      { method: 'PUT', body: JSON.stringify({ jogo_id: jogoId }) },
      token
    )
    return normalizarPerfil(perfil)
  },

  async removerFoto(token?: string): Promise<PerfilDTO> {
    const perfil = await apiRequest<PerfilDTO>('/me/foto', { method: 'DELETE' }, token)
    return normalizarPerfil(perfil)
  },
}
