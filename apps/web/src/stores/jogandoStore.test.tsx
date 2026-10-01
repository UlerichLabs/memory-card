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

  it('guarda o codigo no erro e cria no início da lista', async () => {
    service.listar.mockRejectedValue(new ApiError('jogando.entrada_invalida', '', 400))
    const criado = { ...jogos[0], id: 3, nome: 'Tunic' }
    service.criar.mockResolvedValue(criado)
    const { result } = renderHook(() => useJogandoStore(), { wrapper })
    await act(async () => { await result.current.carregar() })
    expect(result.current.error).toBe('jogando.entrada_invalida')
    await act(async () => { await result.current.criar({ nome: 'Tunic', iniciado_em: '2026-09-20' }) })
    expect(result.current.jogos[0]).toEqual(criado)
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
