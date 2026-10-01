import { afterEach, describe, expect, it, vi } from 'vitest'
import { act, renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { ApiError } from '@/lib/api'
import { jogandoService } from '@/lib/services/jogandoService'
import { JogandoProvider, useJogandoStore } from './jogandoStore'

vi.mock('@/lib/services/jogandoService', () => ({ jogandoService: { listar: vi.fn(), criar: vi.fn(), remover: vi.fn() } }))

const service = vi.mocked(jogandoService)
const jogos = [
  { id: 1, nome: 'Hades', igdb_id: null, igdb_capa_url: null, iniciado_em: '2026-09-20T00:00:00Z' },
  { id: 2, nome: 'Celeste', igdb_id: null, igdb_capa_url: null, iniciado_em: '2026-09-19T00:00:00Z' },
]
const wrapper = ({ children }: { children: ReactNode }) => <JogandoProvider token="token">{children}</JogandoProvider>

describe('JogandoProvider', () => {
  afterEach(() => vi.resetAllMocks())

  it('começa vazio e carrega os jogos', async () => {
    service.listar.mockResolvedValue(jogos)
    const { result } = renderHook(() => useJogandoStore(), { wrapper })
    expect(result.current.jogos).toEqual([])
    await act(async () => { await result.current.carregar() })
    expect(result.current.jogos).toEqual(jogos)
    expect(result.current.carregado).toBe(true)
  })

  it('mantém o estado da lista quando criar falha', async () => {
    service.listar.mockRejectedValue(new ApiError('jogando.entrada_invalida', '', 400))
    const { result } = renderHook(() => useJogandoStore(), { wrapper })
    await act(async () => { await result.current.carregar() })
    expect(result.current.error).toBe('jogando.entrada_invalida')
    const antes = result.current.jogos
    service.criar.mockRejectedValue(new ApiError('jogando.entrada_invalida', '', 400))
    await act(async () => { await expect(result.current.criar({ nome: 'Tunic', iniciado_em: '2026-09-20' })).rejects.toBeInstanceOf(ApiError) })
    expect(result.current.error).toBe('jogando.entrada_invalida')
    expect(result.current.jogos).toEqual(antes)
    expect(result.current.isLoading).toBe(false)
  })

  it('cria sem alterar carregamento e ordena por data e id', async () => {
    service.listar.mockResolvedValue([jogos[0]])
    const criado = { ...jogos[1], id: 3, nome: 'Tunic' }
    service.criar.mockResolvedValue(criado)
    const { result } = renderHook(() => useJogandoStore(), { wrapper })
    await act(async () => { await result.current.carregar() })
    await act(async () => { await result.current.criar({ nome: 'Tunic', iniciado_em: '2026-09-19' }) })
    expect(result.current.jogos.map((jogo) => jogo.id)).toEqual([1, 3])
    expect(result.current.isLoading).toBe(false)
  })

  it('mantém os cards durante um novo carregamento', async () => {
    service.listar.mockResolvedValueOnce(jogos).mockImplementation(() => new Promise(() => undefined))
    const { result } = renderHook(() => useJogandoStore(), { wrapper })
    await act(async () => { await result.current.carregar() })
    act(() => { void result.current.carregar() })
    expect(result.current.carregado).toBe(true)
    expect(result.current.jogos).toEqual(jogos)
    expect(result.current.isLoading).toBe(true)
  })

  it('remove localmente com sucesso ou 404 e mantém o aviso', async () => {
    service.listar.mockResolvedValue(jogos)
    service.remover.mockResolvedValue(undefined)
    const { result } = renderHook(() => useJogandoStore(), { wrapper })
    await act(async () => { await result.current.carregar() })
    await act(async () => { await result.current.remover(1) })
    expect(result.current.jogos).toHaveLength(1)
    service.remover.mockRejectedValue(new ApiError('jogando.nao_encontrado', '', 404))
    await act(async () => { await expect(result.current.remover(2)).rejects.toBeInstanceOf(ApiError) })
    expect(result.current.jogos).toEqual([])
    act(() => result.current.definirAviso('aviso'))
    await waitFor(() => expect(result.current.aviso).toBe('aviso'))
  })
})
