import { describe, expect, it, vi, afterEach } from 'vitest'
import { renderHook, waitFor, act } from '@testing-library/react'
import type { ReactNode } from 'react'
import { MemoryRouter } from 'react-router-dom'
import { AuthProvider } from '@/store/authStore'
import { JogandoProvider } from '@/stores/jogandoStore'
import { useJogoCatalogo } from './useJogoCatalogo'
import { jogosService } from '@/lib/services/jogosService'
import { jogandoService } from '@/lib/services/jogandoService'
import { ApiError } from '@/lib/api'
import type { IGDBJogoSugestao } from '@/types/jogos'

const jogoMock: IGDBJogoSugestao = {
  id: 1029,
  name: 'Chrono Trigger',
  cover: { id: 1, url: '//images.igdb.com/igdb/image/upload/t_thumb/co123.jpg' },
  first_release_date: 794966400,
  summary: 'A classic RPG.',
}

function wrapper({ children }: { children: ReactNode }) {
  return (
    <MemoryRouter initialEntries={['/catalogo/1029']}>
      <AuthProvider>
        <JogandoProvider>
          {children}
        </JogandoProvider>
      </AuthProvider>
    </MemoryRouter>
  )
}

describe('useJogoCatalogo', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('carrega detalhes do jogo com sucesso', async () => {
    vi.spyOn(jogosService, 'obterDetalhesIGDB').mockResolvedValue(jogoMock)

    const { result } = renderHook(() => useJogoCatalogo('1029'), { wrapper })

    expect(result.current.carregando).toBe(true)

    await waitFor(() => {
      expect(result.current.carregando).toBe(false)
    })

    expect(result.current.jogo).toEqual(jogoMock)
    expect(result.current.erroStatus).toBeNull()
  })

  it('marca erroStatus como notFound para id invalido', async () => {
    const { result } = renderHook(() => useJogoCatalogo('invalido'), { wrapper })

    expect(result.current.carregando).toBe(false)
    expect(result.current.erroStatus).toBe('notFound')
    expect(result.current.jogo).toBeNull()
  })

  it('lida com erro de rede e recupera via recarregar', async () => {
    const spy = vi
      .spyOn(jogosService, 'obterDetalhesIGDB')
      .mockRejectedValueOnce(new Error('Falha de rede'))
      .mockResolvedValueOnce(jogoMock)

    const { result } = renderHook(() => useJogoCatalogo('1029'), { wrapper })

    await waitFor(() => {
      expect(result.current.erroStatus).toBe('network')
    })
    expect(result.current.carregando).toBe(false)

    act(() => {
      result.current.recarregar()
    })

    await waitFor(() => {
      expect(result.current.carregando).toBe(false)
    })

    expect(result.current.jogo).toEqual(jogoMock)
    expect(result.current.erroStatus).toBeNull()
    expect(spy).toHaveBeenCalledTimes(2)
  })

  it('inicia jogo com sucesso ao chamar jogarAgora', async () => {
    vi.spyOn(jogosService, 'obterDetalhesIGDB').mockResolvedValue(jogoMock)
    const spyCriar = vi.spyOn(jogandoService, 'criar').mockResolvedValue({
      id: 50,
      nome: 'Chrono Trigger',
      igdb_id: 1029,
      igdb_capa_url: 'https://images.igdb.com/igdb/image/upload/t_cover_big/co123.jpg',
      iniciado_em: '2026-10-24',
    })

    const { result } = renderHook(() => useJogoCatalogo('1029'), { wrapper })

    await waitFor(() => {
      expect(result.current.jogo).toEqual(jogoMock)
    })

    await act(async () => {
      await result.current.jogarAgora()
    })

    expect(spyCriar).toHaveBeenCalledWith(
      expect.objectContaining({
        nome: 'Chrono Trigger',
        igdb_id: 1029,
      }),
      undefined,
    )
    expect(result.current.sucesso).toBe(true)
    expect(result.current.emAndamento).toBe(true)
  })

  it('registra erro amigavel resolvido pelo codigo em falha de jogarAgora', async () => {
    vi.spyOn(jogosService, 'obterDetalhesIGDB').mockResolvedValue(jogoMock)
    vi.spyOn(jogandoService, 'criar').mockRejectedValue(
      new ApiError('jogando.item_duplicado', 'Mensagem ignorada', 409),
    )

    const { result } = renderHook(() => useJogoCatalogo('1029'), { wrapper })

    await waitFor(() => {
      expect(result.current.jogo).toEqual(jogoMock)
    })

    await act(async () => {
      await result.current.jogarAgora()
    })

    expect(result.current.erroAcao).toBe('Este jogo já está em andamento.')
    expect(result.current.sucesso).toBe(false)
  })
})
