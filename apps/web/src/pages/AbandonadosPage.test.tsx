import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { AuthProvider } from '@/store/authStore'
import { JogosProvider } from '@/stores/jogosStore'
import { AbandonadosProvider } from '@/stores/abandonadosStore'
import { GameFormDialog } from '@/components/jogos/GameForm/GameFormDialog'
import { AbandonarJogoDialog } from '@/components/abandonados/AbandonarJogoDialog'
import { AbandonadosPage } from './AbandonadosPage'
import { abandonadosService, AbandonadosApiError } from '@/lib/services/abandonadosService'
import { jogosService } from '@/lib/services/jogosService'
import type { JogoAbandonado } from '@/types/abandonados'

const abandonadoMock: JogoAbandonado = {
  id: 1,
  nome: 'Chrono Trigger',
  console: 'SNES',
  igdb_id: 100,
  igdb_capa_url: 'https://images.igdb.com/cover.jpg',
  tempo_jogado: 36000,
  motivo: 'Muito longo',
  abandonado_em: '2026-01-15',
  iniciado_em: null,
  created_at: '2026-01-15T10:00:00Z',
  updated_at: '2026-01-15T10:00:00Z',
}

function renderAbandonados(initialEntry = '/abandonados') {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <AuthProvider>
        <JogosProvider>
          <AbandonadosProvider>
            <GameFormDialog />
            <AbandonarJogoDialog />
            <Routes>
              <Route path="/abandonados" element={<AbandonadosPage />} />
            </Routes>
          </AbandonadosProvider>
        </JogosProvider>
      </AuthProvider>
    </MemoryRouter>
  )
}

describe('AbandonadosPage', () => {
  beforeEach(() => {
    vi.spyOn(abandonadosService, 'obterFiltros').mockResolvedValue({ consoles: ['SNES', 'GBA'] })
    vi.spyOn(abandonadosService, 'obterTotal').mockResolvedValue({ total: 1 })
    vi.spyOn(abandonadosService, 'listar').mockResolvedValue({
      data: [abandonadoMock],
      meta: { pagina: 1, por_pagina: 12, total: 1, total_paginas: 1 },
    })
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('carrega dados e exibe título, pill de total e card do jogo', async () => {
    renderAbandonados()

    expect(await screen.findByRole('heading', { name: 'Abandonados' })).toBeInTheDocument()
    expect(screen.getByText('1 jogo')).toBeInTheDocument()
    expect(screen.getByText('Chrono Trigger')).toBeInTheDocument()
    expect(screen.getByText('SNES')).toBeInTheDocument()
    expect(screen.getByText('Abandonado em 15/01/2026')).toBeInTheDocument()
    expect(screen.getByText('"Muito longo"')).toBeInTheDocument()
  })

  it('exibe estado vazio quando não há jogos abandonados', async () => {
    vi.spyOn(abandonadosService, 'obterTotal').mockResolvedValue({ total: 0 })
    vi.spyOn(abandonadosService, 'listar').mockResolvedValue({
      data: [],
      meta: { pagina: 1, por_pagina: 12, total: 0, total_paginas: 0 },
    })

    renderAbandonados()

    expect(await screen.findByText('Nenhum jogo abandonado')).toBeInTheDocument()
  })

  it('exibe estado de busca sem resultados com botão de limpar filtros', async () => {
    vi.spyOn(abandonadosService, 'listar').mockResolvedValue({
      data: [],
      meta: { pagina: 1, por_pagina: 12, total: 0, total_paginas: 0 },
    })

    renderAbandonados('/abandonados?busca=Inexistente')

    expect(await screen.findByText('Nenhum jogo encontrado com os filtros aplicados')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Limpar filtros' })).toBeInTheDocument()
  })

  it('remove o botão duplicado e orienta usar o menu Novo no vazio', async () => {
    vi.spyOn(abandonadosService, 'obterTotal').mockResolvedValue({ total: 0 })
    vi.spyOn(abandonadosService, 'listar').mockResolvedValue({ data: [], meta: { pagina: 1, por_pagina: 12, total: 0, total_paginas: 0 } })
    renderAbandonados()
    expect(await screen.findByText('Use o botão Novo no topo e escolha Abandonar jogo.')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Abandonar jogo' })).not.toBeInTheDocument()
  })

  it('abre o modal de criação pelo menu Novo', async () => {
    const user = userEvent.setup()
    renderAbandonados()

    await screen.findByText('Chrono Trigger')
    await user.click(screen.getByRole('button', { name: 'Novo' }))
    await user.click(screen.getByRole('menuitem', { name: /Abandonar jogo/ }))

    expect(await screen.findByRole('heading', { name: 'Abandonar jogo' })).toBeInTheDocument()
    expect(screen.getByLabelText('Nome do jogo *')).toBeInTheDocument()
  })

  it('abre o modal de edição ao clicar no botão de editar', async () => {
    const user = userEvent.setup()
    renderAbandonados()

    await screen.findByText('Chrono Trigger')
    const btnEditar = screen.getByRole('button', { name: 'Editar Chrono Trigger' })
    await user.click(btnEditar)

    expect(await screen.findByRole('heading', { name: 'Editar abandono' })).toBeInTheDocument()
    expect(screen.getByDisplayValue('Chrono Trigger')).toBeInTheDocument()
  })

  it('abre confirmação de exclusão e exclui o jogo', async () => {
    const user = userEvent.setup()
    vi.spyOn(abandonadosService, 'excluir').mockResolvedValue(undefined)
    renderAbandonados()

    await screen.findByText('Chrono Trigger')
    const btnExcluir = screen.getByRole('button', { name: 'Excluir Chrono Trigger' })
    await user.click(btnExcluir)

    expect(await screen.findByRole('heading', { name: 'Excluir jogo abandonado' })).toBeInTheDocument()
    const btnConfirmar = screen.getByRole('button', { name: 'Excluir jogo' })
    await user.click(btnConfirmar)

    await waitFor(() => {
      expect(abandonadosService.excluir).toHaveBeenCalledWith(1, undefined)
    })
  })

  it('ao clicar em Retomar abre o modal de registro com banner e botão Salvar zeramento', async () => {
    const user = userEvent.setup()
    renderAbandonados()

    await screen.findByText('Chrono Trigger')
    const btnRetomar = screen.getByRole('button', { name: 'Retomar Chrono Trigger' })
    await user.click(btnRetomar)

    expect(await screen.findByRole('heading', { name: 'Registrar jogo' })).toBeInTheDocument()
    expect(screen.getByText(/Retomando um jogo abandonado em 15\/01\/2026/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Salvar zeramento' })).toBeInTheDocument()
  })

  it('exibe erro inline no dialog de exclusão quando o backend falha e mantém o dialog aberto', async () => {
    const user = userEvent.setup()
    vi.spyOn(abandonadosService, 'excluir').mockRejectedValue(
      new AbandonadosApiError('abandonados.nao_encontrado', 'Não encontrado', 404)
    )
    renderAbandonados()

    await screen.findByText('Chrono Trigger')
    await user.click(screen.getByRole('button', { name: 'Excluir Chrono Trigger' }))

    expect(await screen.findByRole('heading', { name: 'Excluir jogo abandonado' })).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Excluir jogo' }))

    expect(await screen.findByText('Jogo não encontrado.')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Excluir jogo abandonado' })).toBeInTheDocument()
  })

  it('ao retomar jogo com sucesso, registra zeramento, exclui o jogo abandonado e recarrega a lista', async () => {
    const user = userEvent.setup()
    vi.spyOn(abandonadosService, 'excluir').mockResolvedValue(undefined)
    vi.spyOn(jogosService, 'criar').mockResolvedValue({
      id: 99,
      usuario_id: 1,
      nome: 'Chrono Trigger',
      console: 'SNES',
      genero: 'RPG',
      finalizado_em: '2026-09-30',
      tempo_jogado: 36000,
      nota: 10,
      dificuldade: 'A',
      destaque: false,
    })

    renderAbandonados()

    await screen.findByText('Chrono Trigger')
    await user.click(screen.getByRole('button', { name: 'Retomar Chrono Trigger' }))

    await screen.findByRole('heading', { name: 'Registrar jogo' })
    fireEvent.change(screen.getByLabelText(/Finalizado em/i), { target: { value: '13/03/2026' } })
    await user.click(screen.getByRole('button', { name: '10' }))
    await user.click(screen.getByRole('button', { name: 'Normal' }))
    await user.click(screen.getByRole('button', { name: 'Salvar zeramento' }))

    await waitFor(() => {
      expect(abandonadosService.excluir).toHaveBeenCalledWith(1, undefined)
      expect(abandonadosService.listar).toHaveBeenCalled()
    })
  })

  it('ao retomar jogo, se excluir falhar exibe aviso inline e mantém o item na lista', async () => {
    const user = userEvent.setup()
    vi.spyOn(abandonadosService, 'excluir').mockRejectedValue(new Error('Falha ao excluir'))
    vi.spyOn(jogosService, 'criar').mockResolvedValue({
      id: 99,
      usuario_id: 1,
      nome: 'Chrono Trigger',
      console: 'SNES',
      genero: 'RPG',
      finalizado_em: '2026-09-30',
      tempo_jogado: 36000,
      nota: 10,
      dificuldade: 'A',
      destaque: false,
    })

    renderAbandonados()

    await screen.findByText('Chrono Trigger')
    await user.click(screen.getByRole('button', { name: 'Retomar Chrono Trigger' }))

    await screen.findByRole('heading', { name: 'Registrar jogo' })
    fireEvent.change(screen.getByLabelText(/Finalizado em/i), { target: { value: '13/03/2026' } })
    await user.click(screen.getByRole('button', { name: '10' }))
    await user.click(screen.getByRole('button', { name: 'Normal' }))
    await user.click(screen.getByRole('button', { name: 'Salvar zeramento' }))

    expect(await screen.findByText(
      'Zeramento registrado, mas não foi possível remover este jogo dos abandonados. Exclua-o manualmente.'
    )).toBeInTheDocument()
    expect(screen.getByText('Chrono Trigger')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Fechar aviso' }))
    expect(screen.queryByText(
      'Zeramento registrado, mas não foi possível remover este jogo dos abandonados. Exclua-o manualmente.'
    )).not.toBeInTheDocument()
  })
})
