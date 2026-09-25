import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { JogosContext, type JogosStore } from '@/stores/jogosStore'
import { GameFormDialog } from './GameFormDialog'
import type { JogoZeradoDTO } from '@/lib/services/jogosService'

const mockJogo: JogoZeradoDTO = {
  id: 1,
  usuario_id: 1,
  nome: 'Chrono Trigger',
  console: 'SNES',
  genero: 'JRPG',
  finalizado_em: '2026-01-01',
  tempo_jogado: 3600,
  nota: 10,
  dificuldade: 'A',
  destaque: false,
}

function renderDialog(storeOverrides: Partial<JogosStore> = {}) {
  const store: JogosStore = {
    jogos: [],
    isLoading: false,
    error: null,
    isModalOpen: true,
    jogoEmEdicao: null,
    abrirModalRegistro: vi.fn(),
    abrirModalEdicao: vi.fn(),
    fecharModal: vi.fn(),
    criarJogo: vi.fn(),
    atualizarJogo: vi.fn(),
    excluirJogo: vi.fn(),
    buscarIGDB: vi.fn().mockResolvedValue([]),
    obterDetalhesIGDB: vi.fn(),
    setJogos: vi.fn(),
    limparErro: vi.fn(),
    ...storeOverrides,
  }

  return {
    ...render(
      <JogosContext.Provider value={store}>
        <GameFormDialog />
      </JogosContext.Provider>
    ),
    store,
  }
}

describe('GameFormDialog', () => {
  it('não renderiza conteúdo quando isModalOpen for falso', () => {
    renderDialog({ isModalOpen: false })
    expect(screen.queryByRole('heading', { name: 'Registrar jogo' })).not.toBeInTheDocument()
  })

  it('renderiza título Registrar jogo no modo de criação', () => {
    renderDialog({ isModalOpen: true, jogoEmEdicao: null })
    expect(screen.getByRole('heading', { name: 'Registrar jogo' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Salvar registro' })).toBeInTheDocument()
  })

  it('renderiza título Editar registro e preenche dados no modo de edição', () => {
    renderDialog({ isModalOpen: true, jogoEmEdicao: mockJogo })
    expect(screen.getByRole('heading', { name: 'Editar registro' })).toBeInTheDocument()
    expect(screen.getByLabelText(/Nome do jogo/i)).toHaveValue('Chrono Trigger')
    expect(screen.getByLabelText(/Console/i)).toHaveValue('SNES')
    expect(screen.getByRole('button', { name: 'Atualizar registro' })).toBeInTheDocument()
  })

  it('chama fecharModal ao clicar em Cancelar', async () => {
    const user = userEvent.setup()
    const { store } = renderDialog({ isModalOpen: true })

    const btnCancelar = screen.getByRole('button', { name: 'Cancelar' })
    await user.click(btnCancelar)

    expect(store.fecharModal).toHaveBeenCalledTimes(1)
  })
})
