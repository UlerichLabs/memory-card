import { afterEach, describe, expect, it, vi } from 'vitest'
import { act, renderHook } from '@testing-library/react'
import { createElement, type ReactNode } from 'react'
import { abandonadosService } from '@/lib/services/abandonadosService'
import { AbandonadosProvider, useAbandonadosStore } from './abandonadosStore'
import { JogandoContext, type JogandoStore } from './jogandoStore'
import type { JogoAbandonado } from '@/types/abandonados'

const abandonadoMock: JogoAbandonado = {
  id: 1,
  nome: 'Chrono Trigger',
  console: 'SNES',
  igdb_id: null,
  igdb_capa_url: null,
  abandonado_em: '2026-01-15',
  iniciado_em: null,
  tempo_jogado: 36000,
  motivo: 'Muito difícil',
  created_at: '2026-01-15T12:00:00Z',
  updated_at: '2026-01-15T12:00:00Z',
}

describe('abandonadosStore', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('inicia com estado padrão', () => {
    const { result } = renderHook(() => useAbandonadosStore(), { wrapper: AbandonadosProvider })
    expect(result.current.jogos).toEqual([])
    expect(result.current.meta).toEqual({ pagina: 1, por_pagina: 12, total: 0, total_paginas: 0 })
    expect(result.current.filtros).toEqual({ consoles: [] })
    expect(result.current.totalGeral).toBe(0)
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it('carrega lista e metadados de paginação', async () => {
    vi.spyOn(abandonadosService, 'listar').mockResolvedValue({
      data: [abandonadoMock],
      meta: { pagina: 1, por_pagina: 12, total: 1, total_paginas: 1 },
    })

    const { result } = renderHook(() => useAbandonadosStore(), { wrapper: AbandonadosProvider })

    await act(async () => {
      await result.current.carregarJogos()
    })

    expect(result.current.jogos).toEqual([abandonadoMock])
    expect(result.current.meta.total).toBe(1)
  })

  it('carrega filtros e total', async () => {
    vi.spyOn(abandonadosService, 'obterFiltros').mockResolvedValue({ consoles: ['SNES'] })
    vi.spyOn(abandonadosService, 'obterTotal').mockResolvedValue({ total: 10 })

    const { result } = renderHook(() => useAbandonadosStore(), { wrapper: AbandonadosProvider })

    await act(async () => {
      await result.current.carregarFiltros()
      await result.current.carregarTotal()
    })

    expect(result.current.filtros.consoles).toEqual(['SNES'])
    expect(result.current.totalGeral).toBe(10)
  })

  it('gerencia abertura e fechamento de modais', () => {
    const { result } = renderHook(() => useAbandonadosStore(), { wrapper: AbandonadosProvider })

    act(() => {
      result.current.abrirModalCriacao({
        origemFila: { listaNome: 'Favoritos' },
        valoresIniciais: { nome: 'Zelda' },
      })
    })
    expect(result.current.isModalOpen).toBe(true)
    expect(result.current.jogoEmEdicao).toBeNull()
    expect(result.current.modalOpcoes?.origemFila?.listaNome).toBe('Favoritos')

    act(() => {
      result.current.abrirModalEdicao(abandonadoMock)
    })
    expect(result.current.isModalOpen).toBe(true)
    expect(result.current.jogoEmEdicao).toEqual(abandonadoMock)

    act(() => {
      result.current.fecharModal()
    })
    expect(result.current.isModalOpen).toBe(false)
    expect(result.current.jogoEmEdicao).toBeNull()

    act(() => {
      result.current.abrirModalExcluir(abandonadoMock)
    })
    expect(result.current.isExcluirModalOpen).toBe(true)
    expect(result.current.jogoParaExcluir).toEqual(abandonadoMock)

    act(() => {
      result.current.fecharModalExcluir()
    })
    expect(result.current.isExcluirModalOpen).toBe(false)
    expect(result.current.jogoParaExcluir).toBeNull()
  })

  it('cria, atualiza e exclui jogo chamando serviço e recarregando', async () => {
    vi.spyOn(abandonadosService, 'criar').mockResolvedValue(abandonadoMock)
    vi.spyOn(abandonadosService, 'atualizar').mockResolvedValue(abandonadoMock)
    vi.spyOn(abandonadosService, 'excluir').mockResolvedValue(undefined)
    vi.spyOn(abandonadosService, 'listar').mockResolvedValue({
      data: [abandonadoMock],
      meta: { pagina: 1, por_pagina: 12, total: 1, total_paginas: 1 },
    })
    vi.spyOn(abandonadosService, 'obterFiltros').mockResolvedValue({ consoles: ['SNES'] })
    vi.spyOn(abandonadosService, 'obterTotal').mockResolvedValue({ total: 1 })

    const { result } = renderHook(() => useAbandonadosStore(), { wrapper: AbandonadosProvider })

    await act(async () => {
      await result.current.criarJogo({
        nome: 'Chrono Trigger',
        console: 'SNES',
        abandonado_em: '2026-01-15',
        tempo_jogado_horas: 10,
      })
    })
    expect(abandonadosService.criar).toHaveBeenCalled()

    await act(async () => {
      await result.current.atualizarJogo(1, {
        nome: 'Chrono Trigger',
        console: 'SNES',
        abandonado_em: '2026-01-15',
        tempo_jogado_horas: 15,
      })
    })
    expect(abandonadosService.atualizar).toHaveBeenCalledWith(1, expect.anything(), undefined)

    await act(async () => {
      await result.current.excluirJogo(1)
    })
    expect(abandonadosService.excluir).toHaveBeenCalledWith(1, undefined)
  })

  it('criarJogo recarrega Jogando agora após sucesso', async () => {
    vi.spyOn(abandonadosService, 'criar').mockResolvedValue(abandonadoMock)
    vi.spyOn(abandonadosService, 'listar').mockResolvedValue({ data: [], meta: { pagina: 1, por_pagina: 12, total: 0, total_paginas: 0 } })
    vi.spyOn(abandonadosService, 'obterFiltros').mockResolvedValue({ consoles: [] })
    vi.spyOn(abandonadosService, 'obterTotal').mockResolvedValue({ total: 0 })
    const carregarJogando = vi.fn().mockResolvedValue(undefined)
    const jogando = { carregar: carregarJogando } as unknown as JogandoStore
    const wrapper = ({ children }: { children: ReactNode }) => createElement(JogandoContext.Provider, { value: jogando }, createElement(AbandonadosProvider, null, children))
    const { result } = renderHook(() => useAbandonadosStore(), { wrapper })

    await act(async () => { await result.current.criarJogo({ nome: 'Chrono Trigger', console: 'SNES', abandonado_em: '2026-01-15', tempo_jogado_horas: 10 }) })

    expect(carregarJogando).toHaveBeenCalledTimes(1)
  })

  it('criarJogo não recarrega Jogando agora quando a criação falha', async () => {
    vi.spyOn(abandonadosService, 'criar').mockRejectedValue(new Error('Erro de conexão'))
    const carregarJogando = vi.fn()
    const jogando = { carregar: carregarJogando } as unknown as JogandoStore
    const wrapper = ({ children }: { children: ReactNode }) => createElement(JogandoContext.Provider, { value: jogando }, createElement(AbandonadosProvider, null, children))
    const { result } = renderHook(() => useAbandonadosStore(), { wrapper })

    await act(async () => { await expect(result.current.criarJogo({ nome: 'Chrono Trigger', console: 'SNES', abandonado_em: '2026-01-15', tempo_jogado_horas: 10 })).rejects.toThrow() })

    expect(carregarJogando).not.toHaveBeenCalled()
  })

  it('ignora falha ao recarregar Jogando agora após criar', async () => {
    vi.spyOn(abandonadosService, 'criar').mockResolvedValue(abandonadoMock)
    vi.spyOn(abandonadosService, 'listar').mockResolvedValue({ data: [], meta: { pagina: 1, por_pagina: 12, total: 0, total_paginas: 0 } })
    vi.spyOn(abandonadosService, 'obterFiltros').mockResolvedValue({ consoles: [] })
    vi.spyOn(abandonadosService, 'obterTotal').mockResolvedValue({ total: 0 })
    const carregarJogando = vi.fn().mockRejectedValue(new Error('Falha ao carregar'))
    const jogando = { carregar: carregarJogando } as unknown as JogandoStore
    const wrapper = ({ children }: { children: ReactNode }) => createElement(JogandoContext.Provider, { value: jogando }, createElement(AbandonadosProvider, null, children))
    const { result } = renderHook(() => useAbandonadosStore(), { wrapper })

    await act(async () => { await expect(result.current.criarJogo({ nome: 'Chrono Trigger', console: 'SNES', abandonado_em: '2026-01-15', tempo_jogado_horas: 10 })).resolves.toEqual(abandonadoMock) })

    expect(result.current.error).toBeNull()
    expect(carregarJogando).toHaveBeenCalledTimes(1)
  })
})
