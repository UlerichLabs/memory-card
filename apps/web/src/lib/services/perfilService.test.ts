import { afterEach, describe, expect, it, vi } from 'vitest'
import { perfilService } from './perfilService'
import * as apiModule from '@/lib/api'
import type { AtualizarPerfilPayload, PerfilDTO } from '@/types/perfil'

const mockPerfilBase: PerfilDTO = {
  nome: 'Lucas Ulerich',
  email: 'lucas@example.com',
  username: 'lucas_u',
  bio: 'Gamer',
  avatar_url: '/api/v1/avatares/hash123.jpg',
  avatar_tipo: 'upload',
  jogo_favorito: { id: 1, nome: 'Chrono Trigger', capa_url: 'https://igdb.com/cover.jpg' },
  console_favorito: 'SNES',
  jogando_desde: 1995,
  jogando_desde_efetivo: 1995,
}

describe('perfilService', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('obter busca /me/perfil via GET e converte avatar_url relativa para absoluta', async () => {
    vi.stubEnv('VITE_API_URL', 'http://localhost:8080')
    const spy = vi.spyOn(apiModule, 'apiRequest').mockResolvedValue(mockPerfilBase)

    const res = await perfilService.obter('token-123')

    expect(spy).toHaveBeenCalledWith('/me/perfil', { method: 'GET', signal: undefined }, 'token-123')
    expect(res.avatar_url).toBe('http://localhost:8080/api/v1/avatares/hash123.jpg')
  })

  it('mantém avatar_url absoluta inalterada quando for do tipo jogo', async () => {
    vi.spyOn(apiModule, 'apiRequest').mockResolvedValue({
      ...mockPerfilBase,
      avatar_url: 'https://images.igdb.com/cover456.jpg',
      avatar_tipo: 'jogo',
    })

    const res = await perfilService.obter()
    expect(res.avatar_url).toBe('https://images.igdb.com/cover456.jpg')
  })

  it('atualizar envia PUT /me/perfil com payload em JSON', async () => {
    const payload: AtualizarPerfilPayload = {
      nome: 'Lucas Novo',
      username: 'novo_user',
      bio: 'Nova bio',
      jogo_favorito_id: 10,
      console_favorito: 'PS5',
      jogando_desde: 2000,
    }

    const spy = vi.spyOn(apiModule, 'apiRequest').mockResolvedValue({
      ...mockPerfilBase,
      nome: payload.nome,
      username: payload.username,
    })

    const res = await perfilService.atualizar(payload, 'token-123')

    expect(spy).toHaveBeenCalledWith(
      '/me/perfil',
      { method: 'PUT', body: JSON.stringify(payload) },
      'token-123'
    )
    expect(res.nome).toBe('Lucas Novo')
  })

  it('enviarFoto envia PUT /me/foto com FormData contendo o campo arquivo', async () => {
    const spy = vi.spyOn(apiModule, 'apiRequest').mockResolvedValue(mockPerfilBase)
    const file = new File(['conteudo'], 'avatar.jpg', { type: 'image/jpeg' })

    await perfilService.enviarFoto(file, 'token-123')

    expect(spy).toHaveBeenCalledWith(
      '/me/foto',
      expect.objectContaining({
        method: 'PUT',
        body: expect.any(FormData),
      }),
      'token-123'
    )

    const callBody = spy.mock.calls[0][1]?.body as FormData
    expect(callBody.get('arquivo')).toBe(file)
  })

  it('usarCapa envia PUT /me/foto/capa com jogo_id no body', async () => {
    const spy = vi.spyOn(apiModule, 'apiRequest').mockResolvedValue(mockPerfilBase)

    await perfilService.usarCapa(42, 'token-123')

    expect(spy).toHaveBeenCalledWith(
      '/me/foto/capa',
      { method: 'PUT', body: JSON.stringify({ jogo_id: 42 }) },
      'token-123'
    )
  })

  it('removerFoto envia DELETE /me/foto', async () => {
    const spy = vi.spyOn(apiModule, 'apiRequest').mockResolvedValue({
      ...mockPerfilBase,
      avatar_url: null,
      avatar_tipo: null,
    })

    const res = await perfilService.removerFoto('token-123')

    expect(spy).toHaveBeenCalledWith('/me/foto', { method: 'DELETE' }, 'token-123')
    expect(res.avatar_url).toBeNull()
  })
})
