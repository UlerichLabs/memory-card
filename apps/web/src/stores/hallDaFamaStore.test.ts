import { afterEach, describe, expect, it, vi } from 'vitest'
import { act, renderHook } from '@testing-library/react'
import { jogosService, type JogoZeradoDTO, type ResumoGameDoAnoItem } from '@/lib/services/jogosService'
import { HallDaFamaProvider, useHallDaFamaStore } from './hallDaFamaStore'

const jogoMock: JogoZeradoDTO = {
  id: 1,
  usuario_id: 10,
  nome: 'Chrono Trigger',
  console: 'SNES',
  finalizado_em: '2026-01-15',
  tempo_jogado: 72000,
  nota: 11,
  dificuldade: 'A',
  destaque: true,
}

const resumoMock: ResumoGameDoAnoItem[] = [
  { ano: 2026, total_jogos: 5, game_do_ano: jogoMock },
]

describe('hallDaFamaStore', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('inicia com valores padrão', () => {
    const { result } = renderHook(() => useHallDaFamaStore(), { wrapper: HallDaFamaProvider })
    expect(result.current.resumo).toEqual([])
    expect(result.current.gamesDaVida).toEqual([])
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it('carrega resumo e games da vida simultaneamente', async () => {
    vi.spyOn(jogosService, 'obterResumoGameDoAno').mockResolvedValue(resumoMock)
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 1, por_pagina: 100, total: 1, total_paginas: 1 },
    })

    const { result } = renderHook(() => useHallDaFamaStore(), { wrapper: HallDaFamaProvider })

    await act(async () => {
      await result.current.carregarHallDaFama()
    })

    expect(result.current.resumo).toEqual(resumoMock)
    expect(result.current.gamesDaVida).toEqual([jogoMock])
    expect(result.current.isLoading).toBe(false)
  })

  it('trata erro no carregamento do Hall da Fama', async () => {
    vi.spyOn(jogosService, 'obterResumoGameDoAno').mockRejectedValue(new Error('Falha de rede'))
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [],
      meta: { pagina: 1, por_pagina: 100, total: 0, total_paginas: 0 },
    })

    const { result } = renderHook(() => useHallDaFamaStore(), { wrapper: HallDaFamaProvider })

    await act(async () => {
      await expect(result.current.carregarHallDaFama()).rejects.toThrow('Falha de rede')
    })

    expect(result.current.error).toBe('Falha de rede')
    expect(result.current.isLoading).toBe(false)
  })

  it('definirGameDoAno chama service e recarrega dados', async () => {
    const spyDefinir = vi.spyOn(jogosService, 'definirGameDoAno').mockResolvedValue({
      ano: 2026,
      anterior_id: null,
      game_do_ano: jogoMock,
    })
    const spyResumo = vi.spyOn(jogosService, 'obterResumoGameDoAno').mockResolvedValue(resumoMock)
    const spyListar = vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 1, por_pagina: 100, total: 1, total_paginas: 1 },
    })

    const { result } = renderHook(() => useHallDaFamaStore(), { wrapper: HallDaFamaProvider })

    await act(async () => {
      const resp = await result.current.definirGameDoAno(1)
      expect(resp.ano).toBe(2026)
    })

    expect(spyDefinir).toHaveBeenCalledWith(1, undefined)
    expect(spyResumo).toHaveBeenCalled()
    expect(spyListar).toHaveBeenCalled()
  })

  it('removerGameDoAno chama service e recarrega dados', async () => {
    const spyRemover = vi.spyOn(jogosService, 'removerGameDoAno').mockResolvedValue(undefined)
    const spyResumo = vi.spyOn(jogosService, 'obterResumoGameDoAno').mockResolvedValue([])
    const spyListar = vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [],
      meta: { pagina: 1, por_pagina: 100, total: 0, total_paginas: 0 },
    })

    const { result } = renderHook(() => useHallDaFamaStore(), { wrapper: HallDaFamaProvider })

    await act(async () => {
      await result.current.removerGameDoAno(1)
    })

    expect(spyRemover).toHaveBeenCalledWith(1, undefined)
    expect(spyResumo).toHaveBeenCalled()
    expect(spyListar).toHaveBeenCalled()
  })

  it('buscarJogosDoAno busca jogos do ano com ordenar=nota', async () => {
    const spyListar = vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 1, por_pagina: 100, total: 1, total_paginas: 1 },
    })

    const { result } = renderHook(() => useHallDaFamaStore(), { wrapper: HallDaFamaProvider })

    let jogos: JogoZeradoDTO[] = []
    await act(async () => {
      jogos = await result.current.buscarJogosDoAno(2026)
    })

    expect(spyListar).toHaveBeenCalledWith({ ano: 2026, ordenar: 'nota', por_pagina: 100 }, undefined, undefined)
    expect(jogos).toEqual([jogoMock])
  })
})
