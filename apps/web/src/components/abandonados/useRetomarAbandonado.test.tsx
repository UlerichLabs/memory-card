import { describe, expect, it, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { createElement, type ReactNode } from 'react'
import { useRetomarAbandonado } from './useRetomarAbandonado'
import { JogosContext, type JogosStore } from '@/stores/jogosStore'
import { AbandonadosContext, type AbandonadosStore } from '@/stores/abandonadosStore'
import { AVISO_FALHA_REMOCAO_RETOMAR } from './abandonados.constants'
import type { JogoAbandonado } from '@/types/abandonados'

const mockJogo: JogoAbandonado = {
  id: 10,
  nome: 'Chrono Trigger',
  console: 'SNES',
  igdb_id: 100,
  igdb_capa_url: 'https://images.igdb.com/cover.jpg',
  tempo_jogado: 36000,
  motivo: 'Muito difícil',
  abandonado_em: '2026-01-15',
  iniciado_em: null,
  created_at: '2026-01-15T10:00:00Z',
  updated_at: '2026-01-15T10:00:00Z',
}

function createWrappers(
  jogosOverrides: Partial<JogosStore> = {},
  abandonadosOverrides: Partial<AbandonadosStore> = {}
) {
  const mockAbrirModalRegistro = vi.fn()
  const mockExcluirJogo = vi.fn().mockResolvedValue(undefined)
  const mockDefinirAviso = vi.fn()

  const jogosStore: JogosStore = {
    jogos: [],
    meta: { pagina: 1, por_pagina: 24, total: 0, total_paginas: 0 },
    filtros: { consoles: [], generos: [], tipos: [], anos: [] },
    isLoading: false,
    error: null,
    isModalOpen: false,
    jogoEmEdicao: null,
    modalRegistroOpcoes: null,
    abrirModalRegistro: mockAbrirModalRegistro,
    abrirModalEdicao: vi.fn(),
    fecharModal: vi.fn(),
    carregarJogos: vi.fn(),
    carregarFiltros: vi.fn(),
    limparBiblioteca: vi.fn(),
    criarJogo: vi.fn(),
    atualizarJogo: vi.fn(),
    excluirJogo: vi.fn(),
    buscarIGDB: vi.fn(),
    obterDetalhesIGDB: vi.fn(),
    obterJogoPorId: vi.fn(),
    setJogos: vi.fn(),
    limparErro: vi.fn(),
    ...jogosOverrides,
  }

  const abandonadosStore: AbandonadosStore = {
    jogos: [mockJogo],
    meta: { pagina: 1, por_pagina: 12, total: 1, total_paginas: 1 },
    filtros: { consoles: ['SNES'] },
    totalGeral: 1,
    isLoading: false,
    carregado: true,
    error: null,
    aviso: null,
    isModalOpen: false,
    jogoEmEdicao: null,
    modalOpcoes: null,
    isExcluirModalOpen: false,
    jogoParaExcluir: null,
    abrirModalCriacao: vi.fn(),
    abrirModalEdicao: vi.fn(),
    fecharModal: vi.fn(),
    abrirModalExcluir: vi.fn(),
    fecharModalExcluir: vi.fn(),
    carregarJogos: vi.fn(),
    carregarFiltros: vi.fn(),
    carregarTotal: vi.fn(),
    criarJogo: vi.fn(),
    atualizarJogo: vi.fn(),
    excluirJogo: mockExcluirJogo,
    limparErro: vi.fn(),
    definirAviso: mockDefinirAviso,
    limparAviso: vi.fn(),
    ...abandonadosOverrides,
  }

  const wrapper = ({ children }: { children: ReactNode }) =>
    createElement(
      JogosContext.Provider,
      { value: jogosStore },
      createElement(AbandonadosContext.Provider, { value: abandonadosStore }, children)
    )

  return { wrapper, mockAbrirModalRegistro, mockExcluirJogo, mockDefinirAviso }
}

describe('useRetomarAbandonado', () => {
  it('abre modal de registro com valores preenchidos e exclui abandono no sucesso do onSalvo', async () => {
    const { wrapper, mockAbrirModalRegistro, mockExcluirJogo } = createWrappers()
    const { result } = renderHook(() => useRetomarAbandonado(), { wrapper })

    act(() => {
      result.current.retomarJogo(mockJogo)
    })

    expect(mockAbrirModalRegistro).toHaveBeenCalledTimes(1)
    const opcoes = mockAbrirModalRegistro.mock.calls[0][0]
    expect(opcoes.valoresIniciais).toEqual({
      nome: 'Chrono Trigger',
      console: 'SNES',
      igdb_id: 100,
      igdb_capa_url: 'https://images.igdb.com/cover.jpg',
      tempo_jogado: 36000,
    })
    expect(opcoes.textoSubmit).toBe('Salvar zeramento')
    expect(opcoes.aviso).toContain('Retomando um jogo abandonado em 15/01/2026')

    await act(async () => {
      await opcoes.onSalvo()
    })

    expect(mockExcluirJogo).toHaveBeenCalledWith(10)
  })

  it('quando exclusão falha no onSalvo, define aviso no store e não propaga erro', async () => {
    const { wrapper, mockAbrirModalRegistro, mockDefinirAviso } = createWrappers({
      ...{},
    }, {
      excluirJogo: vi.fn().mockRejectedValue(new Error('Erro ao excluir')),
    })

    const { result } = renderHook(() => useRetomarAbandonado(), { wrapper })

    act(() => {
      result.current.retomarJogo(mockJogo)
    })

    const opcoes = mockAbrirModalRegistro.mock.calls[0][0]

    await act(async () => {
      await opcoes.onSalvo()
    })

    expect(mockDefinirAviso).toHaveBeenCalledWith(AVISO_FALHA_REMOCAO_RETOMAR)
  })
})
