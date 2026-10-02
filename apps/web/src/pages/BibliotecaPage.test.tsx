import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { AuthProvider } from '@/store/authStore'
import { JogosProvider } from '@/stores/jogosStore'
import { GameFormDialog } from '@/components/jogos/GameForm/GameFormDialog'
import { BibliotecaPage } from './BibliotecaPage'
import { jogosService } from '@/lib/services/jogosService'
import type { JogoZeradoDTO, OpcoesFiltrosDTO } from '@/types/jogos'

const jogoMock: JogoZeradoDTO = {
  id: 1,
  usuario_id: 10,
  nome: 'Chrono Trigger',
  console: 'SNES',
  genero: 'JRPG',
  tipo: 'Campanha',
  iniciado_em: '2026-01-01',
  finalizado_em: '2026-01-15',
  tempo_jogado: 72000,
  nota: 10,
  dificuldade: 'A',
  review: '100% dos finais',
  destaque: true,
  igdb_capa_url: 'https://images.igdb.com/cover.jpg',
}

const filtrosMock: OpcoesFiltrosDTO = {
  consoles: ['SNES', 'PS5'],
  generos: ['JRPG', 'Ação'],
  tipos: ['Campanha'],
  anos: [2026, 2025],
}

function renderBiblioteca(initialEntry = '/biblioteca') {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <AuthProvider>
        <JogosProvider>
          <GameFormDialog />
          <Routes>
            <Route path="/biblioteca" element={<BibliotecaPage />} />
          </Routes>
        </JogosProvider>
      </AuthProvider>
    </MemoryRouter>
  )
}

describe('BibliotecaPage', () => {
  beforeEach(() => {
    vi.spyOn(jogosService, 'obterFiltros').mockResolvedValue(filtrosMock)
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('carregamento inicial chama api com pagina=1 e por_pagina=100 e filtros', async () => {
    const spyListar = vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 1, por_pagina: 24, total: 1, total_paginas: 1 },
    })

    renderBiblioteca()

    await waitFor(() => {
      expect(spyListar).toHaveBeenCalledWith(
        expect.objectContaining({ pagina: 1, por_pagina: 100 }),
        undefined,
        expect.any(AbortSignal)
      )
    })
    expect(jogosService.obterFiltros).toHaveBeenCalled()
    expect(await screen.findByRole('heading', { name: 'Chrono Trigger' })).toBeInTheDocument()
  })

  it('renderiza os cards no modo grade com título, console, ano, tempo, dificuldade, nota e capa', async () => {
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 1, por_pagina: 24, total: 1, total_paginas: 1 },
    })

    renderBiblioteca()

    expect(await screen.findByRole('heading', { name: 'Chrono Trigger' })).toBeInTheDocument()
    const card = screen.getByRole('article')
    expect(within(card).getByText('SNES')).toBeInTheDocument()
    expect(within(card).getByText('15/01/2026 · 20h')).toBeInTheDocument()
    expect(within(card).getByText('Normal')).toBeInTheDocument()
    expect(within(card).getByText('10')).toBeInTheDocument()
    expect(within(card).getByText('Jogo do ano')).toBeInTheDocument()
    const capa = screen.getByAltText('Chrono Trigger')
    expect(capa).toHaveAttribute('src', 'https://images.igdb.com/cover.jpg')
  })

  it('alterna entre modo grade e lista pelo toggle', async () => {
    const user = userEvent.setup()
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 1, por_pagina: 24, total: 1, total_paginas: 1 },
    })

    renderBiblioteca()
    expect(await screen.findByRole('heading', { name: 'Chrono Trigger' })).toBeInTheDocument()

    const btnLista = screen.getByRole('button', { name: 'Visualização em lista' })
    await user.click(btnLista)
    expect(await screen.findByTestId('biblioteca-lista')).toBeInTheDocument()

    const btnGrade = screen.getByRole('button', { name: 'Visualização em grade' })
    await user.click(btnGrade)
    expect(await screen.findByTestId('biblioteca-grade')).toBeInTheDocument()
  })

  it('busca com debounce de 300ms, ignora espaços extras nas pontas e limpar a busca volta a listar tudo', async () => {
    const user = userEvent.setup()
    const spyListar = vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 1, por_pagina: 24, total: 1, total_paginas: 1 },
    })

    renderBiblioteca()
    await screen.findByRole('heading', { name: 'Chrono Trigger' })
    spyListar.mockClear()

    const input = screen.getByLabelText('Filtrar jogos na biblioteca')
    await user.type(input, 'Zelda')

    await waitFor(() => {
      expect(spyListar).toHaveBeenCalledWith(
        expect.objectContaining({ busca: 'Zelda', pagina: 1 }),
        undefined,
        expect.any(AbortSignal)
      )
    }, { timeout: 1000 })
    expect(spyListar).toHaveBeenCalledTimes(1)

    const btnLimpar = screen.getByRole('button', { name: 'Limpar busca' })
    await user.click(btnLimpar)

    await waitFor(() => {
      expect(spyListar).toHaveBeenLastCalledWith(
        expect.not.objectContaining({ busca: 'Zelda' }),
        undefined,
        expect.any(AbortSignal)
      )
    })
  })

  it('filtra por select atualizando a requisição e resetando pagina', async () => {
    const user = userEvent.setup()
    const spyListar = vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 1, por_pagina: 24, total: 1, total_paginas: 1 },
    })

    renderBiblioteca()
    await screen.findByRole('heading', { name: 'Chrono Trigger' })
    spyListar.mockClear()

    const selectConsole = screen.getByLabelText('Filtrar por console')
    await user.click(selectConsole)
    await user.click(screen.getByRole('option', { name: 'SNES' }))

    await waitFor(() => {
      expect(spyListar).toHaveBeenCalledWith(
        expect.objectContaining({ console: 'SNES' }),
        undefined,
        expect.any(AbortSignal)
      )
    })
  })

  it('filtra por genero, tipo e ano e mudar qualquer filtro quando pagina > 1 reseta para pagina 1', async () => {
    const user = userEvent.setup()
    const spyListar = vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 2, por_pagina: 24, total: 48, total_paginas: 2 },
    })

    renderBiblioteca('/biblioteca?pagina=2')
    await screen.findByRole('heading', { name: 'Chrono Trigger' })
    spyListar.mockClear()

    const selectGenero = screen.getByLabelText('Filtrar por gênero')
    await user.click(selectGenero)
    await user.click(screen.getByRole('option', { name: 'JRPG' }))

    await waitFor(() => {
      expect(spyListar).toHaveBeenCalledWith(
        expect.objectContaining({ genero: 'JRPG', pagina: 1 }),
        undefined,
        expect.any(AbortSignal)
      )
    })
    spyListar.mockClear()

    const selectTipo = screen.getByLabelText('Filtrar por tipo')
    await user.click(selectTipo)
    await user.click(screen.getByRole('option', { name: 'Campanha' }))

    await waitFor(() => {
      expect(spyListar).toHaveBeenCalledWith(
        expect.objectContaining({ tipo: 'Campanha', pagina: 1 }),
        undefined,
        expect.any(AbortSignal)
      )
    })
    spyListar.mockClear()

    const selectAno = screen.getByLabelText('Filtrar por ano')
    await user.click(selectAno)
    await user.click(screen.getByRole('option', { name: '2026' }))

    await waitFor(() => {
      expect(spyListar).toHaveBeenCalledWith(
        expect.objectContaining({ ano: 2026, pagina: 1 }),
        undefined,
        expect.any(AbortSignal)
      )
    })
  })

  it('monta a página com todos os parâmetros na URL e reproduz filtros, página e modo', async () => {
    const spyListar = vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 2, por_pagina: 24, total: 48, total_paginas: 2 },
    })

    const urlCompleta = '/biblioteca?busca=Chrono&console=SNES&genero=JRPG&tipo=Campanha&ano=2026&nota_min=8&nota_max=10&dificuldade=A&pagina=2&modo=list'
    renderBiblioteca(urlCompleta)

    await waitFor(() => {
      expect(spyListar).toHaveBeenCalledWith(
        expect.objectContaining({
          busca: 'Chrono',
          console: 'SNES',
          genero: 'JRPG',
          tipo: 'Campanha',
          ano: 2026,
          nota_min: 8,
          nota_max: 10,
          dificuldade: 'A',
          pagina: 2,
        }),
        undefined,
        expect.any(AbortSignal)
      )
    })
    expect(await screen.findByTestId('biblioteca-lista')).toBeInTheDocument()
  })

  it('ajusta faixa de nota quando min > max', async () => {
    const user = userEvent.setup()
    const spyListar = vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 1, por_pagina: 24, total: 1, total_paginas: 1 },
    })

    renderBiblioteca()
    await screen.findByRole('heading', { name: 'Chrono Trigger' })
    spyListar.mockClear()

    const selectMin = screen.getByLabelText('Nota mínima')
    await user.click(selectMin)
    const options = screen.getAllByRole('option', { name: '10' })
    await user.click(options[0])

    await waitFor(() => {
      expect(spyListar).toHaveBeenCalledWith(
        expect.objectContaining({ nota_min: 10 }),
        undefined,
        expect.any(AbortSignal)
      )
    })
  })

  it('filtra por dificuldade e toggle para desmarcar', async () => {
    const user = userEvent.setup()
    const spyListar = vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 1, por_pagina: 24, total: 1, total_paginas: 1 },
    })

    renderBiblioteca()
    await screen.findByRole('heading', { name: 'Chrono Trigger' })
    spyListar.mockClear()

    const btnDifA = screen.getByRole('button', { name: 'Normal' })
    await user.click(btnDifA)

    await waitFor(() => {
      expect(spyListar).toHaveBeenCalledWith(
        expect.objectContaining({ dificuldade: 'A' }),
        undefined,
        expect.any(AbortSignal)
      )
    })

    await user.click(btnDifA)
    await waitFor(() => {
      expect(spyListar).toHaveBeenLastCalledWith(
        expect.not.objectContaining({ dificuldade: 'A' }),
        undefined,
        expect.any(AbortSignal)
      )
    })
  })

  it('exibe contador de filtros ativos e botão limpar filtros', async () => {
    const user = userEvent.setup()
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 1, por_pagina: 24, total: 1, total_paginas: 1 },
    })

    renderBiblioteca()
    await screen.findByRole('heading', { name: 'Chrono Trigger' })

    const btnDifAA = screen.getByRole('button', { name: 'Difícil' })
    await user.click(btnDifAA)

    expect(await screen.findByText('filtro ativo')).toBeInTheDocument()
    const btnLimpar = screen.getByRole('button', { name: 'Limpar filtros' })
    await user.click(btnLimpar)

    expect(screen.queryByText('filtro ativo')).not.toBeInTheDocument()
  })

  it('navega pelas páginas na paginação', async () => {
    const user = userEvent.setup()
    const spyListar = vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 1, por_pagina: 24, total: 48, total_paginas: 2 },
    })

    renderBiblioteca()
    expect(await screen.findByText(/Mostrando/)).toBeInTheDocument()

    const btnPagina2 = screen.getByRole('button', { name: 'Página 2' })
    await user.click(btnPagina2)

    await waitFor(() => {
      expect(spyListar).toHaveBeenCalledWith(
        expect.objectContaining({ pagina: 2 }),
        undefined,
        expect.any(AbortSignal)
      )
    })
  })

  it('exibe onboarding quando biblioteca estiver vazia sem filtros', async () => {
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [],
      meta: { pagina: 1, por_pagina: 24, total: 0, total_paginas: 0 },
    })

    renderBiblioteca()

    expect(await screen.findByText('Nenhum jogo registrado ainda')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Registrar primeiro jogo/i })).toBeInTheDocument()
  })

  it('exibe mensagem de sem resultados quando houver filtros ativos sem retorno', async () => {
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [],
      meta: { pagina: 1, por_pagina: 24, total: 0, total_paginas: 0 },
    })

    renderBiblioteca('/biblioteca?busca=Inexistente')

    expect(await screen.findByText('Nenhum jogo encontrado com os filtros aplicados')).toBeInTheDocument()
    expect(screen.getAllByRole('button', { name: 'Limpar filtros' }).length).toBeGreaterThanOrEqual(1)
  })

  it('abre modal de edição a partir do menu do card', async () => {
    const user = userEvent.setup()
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 1, por_pagina: 24, total: 1, total_paginas: 1 },
    })

    renderBiblioteca()
    await screen.findByRole('heading', { name: 'Chrono Trigger' })

    const btnMenu = screen.getByRole('button', { name: 'Opções de Chrono Trigger' })
    await user.click(btnMenu)

    const btnEditar = screen.getByRole('button', { name: 'Editar Chrono Trigger' })
    await user.click(btnEditar)

    expect(await screen.findByRole('heading', { name: 'Editar registro' })).toBeVisible()
  })

  it('abre modal de exclusão a partir do menu do card', async () => {
    const user = userEvent.setup()
    vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 1, por_pagina: 24, total: 1, total_paginas: 1 },
    })

    renderBiblioteca()
    await screen.findByRole('heading', { name: 'Chrono Trigger' })

    const btnMenu = screen.getByRole('button', { name: 'Opções de Chrono Trigger' })
    await user.click(btnMenu)

    const btnExcluir = screen.getByRole('button', { name: 'Excluir Chrono Trigger' })
    await user.click(btnExcluir)

    expect(await screen.findByRole('heading', { name: 'Excluir registro' })).toBeVisible()
  })
})
