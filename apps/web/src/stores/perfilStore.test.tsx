import { act, renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { PerfilProvider, usePerfilStore } from './perfilStore'
import { perfilService } from '@/lib/services/perfilService'
import type { PerfilDTO } from '@/types/perfil'

const mockPerfil: PerfilDTO = {
  nome: 'Lucas Ulerich',
  email: 'lucas@example.com',
  username: 'lucas_u',
  bio: 'Bio teste',
  avatar_url: 'https://example.com/foto.jpg',
  avatar_tipo: 'upload',
  jogo_favorito: null,
  console_favorito: null,
  jogando_desde: 2000,
  jogando_desde_efetivo: 2000,
}

function wrapper({ children }: { children: ReactNode }) {
  return <PerfilProvider token="test-token">{children}</PerfilProvider>
}

describe('perfilStore', () => {
  it('carrega o perfil no mount com o estado inicial correto', async () => {
    vi.spyOn(perfilService, 'obter').mockResolvedValue(mockPerfil)

    const { result } = renderHook(() => usePerfilStore(), { wrapper })

    await waitFor(() => {
      expect(result.current.perfil).toEqual(mockPerfil)
    })
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it('atualiza o perfil após chamada de sucesso em salvar()', async () => {
    vi.spyOn(perfilService, 'obter').mockResolvedValue(mockPerfil)
    const atualizado: PerfilDTO = { ...mockPerfil, nome: 'Lucas Alterado' }
    vi.spyOn(perfilService, 'atualizar').mockResolvedValue(atualizado)

    const { result } = renderHook(() => usePerfilStore(), { wrapper })
    await waitFor(() => expect(result.current.perfil).toEqual(mockPerfil))

    await act(async () => {
      await result.current.salvar({
        nome: 'Lucas Alterado',
        username: 'lucas_u',
        bio: 'Bio teste',
        jogo_favorito_id: null,
        console_favorito: null,
        jogando_desde: 2000,
      })
    })

    expect(result.current.perfil?.nome).toBe('Lucas Alterado')
  })

  it('atualiza avatar após enviarFoto, usarCapa e removerFoto', async () => {
    vi.spyOn(perfilService, 'obter').mockResolvedValue(mockPerfil)
    const perfilComNovaFoto: PerfilDTO = { ...mockPerfil, avatar_url: 'https://example.com/nova.jpg' }
    const perfilComCapa: PerfilDTO = { ...mockPerfil, avatar_url: 'https://example.com/capa.jpg', avatar_tipo: 'jogo' }
    const perfilSemFoto: PerfilDTO = { ...mockPerfil, avatar_url: null, avatar_tipo: null }

    vi.spyOn(perfilService, 'enviarFoto').mockResolvedValue(perfilComNovaFoto)
    vi.spyOn(perfilService, 'usarCapa').mockResolvedValue(perfilComCapa)
    vi.spyOn(perfilService, 'removerFoto').mockResolvedValue(perfilSemFoto)

    const { result } = renderHook(() => usePerfilStore(), { wrapper })
    await waitFor(() => expect(result.current.perfil).toEqual(mockPerfil))

    await act(async () => {
      await result.current.enviarFoto(new File(['x'], 'avatar.png', { type: 'image/png' }))
    })
    expect(result.current.perfil?.avatar_url).toBe('https://example.com/nova.jpg')

    await act(async () => {
      await result.current.usarCapa(10)
    })
    expect(result.current.perfil?.avatar_tipo).toBe('jogo')

    await act(async () => {
      await result.current.removerFoto()
    })
    expect(result.current.perfil?.avatar_url).toBeNull()
  })

  it('erro em salvar() define erro e não corrompe o estado prévio', async () => {
    vi.spyOn(perfilService, 'obter').mockResolvedValue(mockPerfil)
    vi.spyOn(perfilService, 'atualizar').mockRejectedValue(new Error('Falha no servidor'))

    const { result } = renderHook(() => usePerfilStore(), { wrapper })
    await waitFor(() => expect(result.current.perfil).toEqual(mockPerfil))

    await act(async () => {
      try {
        await result.current.salvar({
          nome: 'Invalido',
          username: null,
          bio: null,
          jogo_favorito_id: null,
          console_favorito: null,
          jogando_desde: null,
        })
      } catch {}
    })

    expect(result.current.error).toBe('Falha no servidor')
    expect(result.current.perfil).toEqual(mockPerfil)
  })
})
