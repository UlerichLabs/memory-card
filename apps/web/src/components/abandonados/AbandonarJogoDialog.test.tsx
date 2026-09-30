import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createElement, type ReactNode } from 'react'
import { AbandonarJogoDialog } from './AbandonarJogoDialog'
import { AbandonadosContext, type AbandonadosStore } from '@/stores/abandonadosStore'
import { AbandonadosApiError } from '@/lib/services/abandonadosService'
import { ERRO_FALHA_REMOCAO_FILA } from './abandonados.constants'
import type { JogoAbandonado } from '@/types/abandonados'

const mockJogoEdicao: JogoAbandonado = {
  id: 1,
  nome: 'Chrono Trigger',
  console: 'SNES',
  igdb_id: 100,
  igdb_capa_url: null,
  tempo_jogado: 3600,
  motivo: 'Demorado',
  abandonado_em: '2025-05-10',
  created_at: '2025-05-10T10:00:00Z',
  updated_at: '2025-05-10T10:00:00Z',
}

function renderDialog(storeOverrides: Partial<AbandonadosStore> = {}) {
  const store: AbandonadosStore = {
    jogos: [],
    meta: { pagina: 1, por_pagina: 12, total: 0, total_paginas: 0 },
    filtros: { consoles: [] },
    totalGeral: 0,
    isLoading: false,
    carregado: true,
    error: null,
    aviso: null,
    isModalOpen: true,
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
    criarJogo: vi.fn().mockResolvedValue({ id: 99, nome: 'Novo', console: 'PC', tempo_jogado: 0 }),
    atualizarJogo: vi.fn().mockResolvedValue(mockJogoEdicao),
    excluirJogo: vi.fn(),
    limparErro: vi.fn(),
    definirAviso: vi.fn(),
    limparAviso: vi.fn(),
    ...storeOverrides,
  }

  const wrapper = ({ children }: { children: ReactNode }) =>
    createElement(AbandonadosContext.Provider, { value: store }, children)

  return {
    ...render(<AbandonarJogoDialog />, { wrapper }),
    store,
  }
}

describe('AbandonarJogoDialog', () => {
  it('renderiza título e botão corretos no modo criar', () => {
    renderDialog()

    expect(screen.getByRole('heading', { name: 'Abandonar jogo' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Salvar abandono' })).toBeInTheDocument()
  })

  it('renderiza título e botão no modo editar e reenvia abandonado_em original', async () => {
    const user = userEvent.setup()
    const mockAtualizar = vi.fn().mockResolvedValue(mockJogoEdicao)
    renderDialog({
      jogoEmEdicao: mockJogoEdicao,
      atualizarJogo: mockAtualizar,
    })

    expect(screen.getByRole('heading', { name: 'Editar abandono' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Salvar alterações' })).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Salvar alterações' }))

    expect(mockAtualizar).toHaveBeenCalledTimes(1)
    const callArgs = mockAtualizar.mock.calls[0]
    expect(callArgs[0]).toBe(1)
    expect(callArgs[1].abandonado_em).toBe('2025-05-10')
  })

  it('fila: POST ok + remoção ok fecha o dialog', async () => {
    const user = userEvent.setup()
    const onSalvo = vi.fn().mockResolvedValue(undefined)
    const { store } = renderDialog({
      modalOpcoes: {
        valoresIniciais: { nome: 'Super Mario', console: 'SNES' },
        origemFila: { listaNome: 'Backlog', onSalvo },
      },
    })

    expect(screen.getByText(/Vindo da fila Backlog/)).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Salvar abandono' }))

    expect(store.criarJogo).toHaveBeenCalledTimes(1)
    expect(onSalvo).toHaveBeenCalledTimes(1)
    expect(store.fecharModal).toHaveBeenCalledTimes(1)
  })

  it('fila: POST falha mantém item na fila sem chamar remoção', async () => {
    const user = userEvent.setup()
    const onSalvo = vi.fn().mockResolvedValue(undefined)
    const { store } = renderDialog({
      modalOpcoes: {
        valoresIniciais: { nome: 'Super Mario', console: 'SNES' },
        origemFila: { listaNome: 'Backlog', onSalvo },
      },
      criarJogo: vi.fn().mockRejectedValue(new AbandonadosApiError('abandonados.entrada_invalida', 'Dados inválidos')),
    })

    await user.click(screen.getByRole('button', { name: 'Salvar abandono' }))

    expect(store.criarJogo).toHaveBeenCalledTimes(1)
    expect(onSalvo).not.toHaveBeenCalled()
    expect(store.fecharModal).not.toHaveBeenCalled()
    expect(await screen.findByText('Dados inválidos.')).toBeInTheDocument()
  })

  it('fila: POST ok + remoção falha exibe erro específico e segundo clique não chama criar de novo', async () => {
    const user = userEvent.setup()
    const onSalvo = vi
      .fn()
      .mockRejectedValueOnce(new Error('Falha ao remover da fila'))
      .mockResolvedValueOnce(undefined)

    const { store } = renderDialog({
      modalOpcoes: {
        valoresIniciais: { nome: 'Super Mario', console: 'SNES' },
        origemFila: { listaNome: 'Backlog', onSalvo },
      },
    })

    await user.click(screen.getByRole('button', { name: 'Salvar abandono' }))

    expect(store.criarJogo).toHaveBeenCalledTimes(1)
    expect(onSalvo).toHaveBeenCalledTimes(1)
    expect(store.fecharModal).not.toHaveBeenCalled()
    expect(await screen.findByText(ERRO_FALHA_REMOCAO_FILA)).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Salvar abandono' }))

    expect(store.criarJogo).toHaveBeenCalledTimes(1)
    expect(onSalvo).toHaveBeenCalledTimes(2)
    expect(store.fecharModal).toHaveBeenCalledTimes(1)
  })
})
