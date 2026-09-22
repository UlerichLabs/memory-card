import { afterEach, describe, expect, it, vi } from 'vitest'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { AuthProvider } from '@/store/authStore'
import { JogosProvider } from '@/stores/jogosStore'
import { GameFormDialog } from '@/components/jogos/GameForm/GameFormDialog'
import { BibliotecaPage } from './BibliotecaPage'
import type { JogoZeradoDTO } from '@/lib/services/jogosService'

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
  condicao_zeramento: '100% dos finais',
  destaque: true,
  igdb_capa_url: 'https://images.igdb.com/cover.jpg',
}

function renderBiblioteca(initialJogos: JogoZeradoDTO[] = []) {
  return render(
    <MemoryRouter initialEntries={['/biblioteca']}>
      <AuthProvider>
        <JogosProvider initialJogos={initialJogos}>
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
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('renderiza estado vazio quando não há jogos cadastrados', () => {
    renderBiblioteca([])

    expect(screen.getByRole('heading', { name: 'Biblioteca' })).toBeInTheDocument()
    expect(screen.getByText('0 jogos')).toBeInTheDocument()
    expect(screen.getByText('Nenhum jogo registrado ainda')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Registrar primeiro jogo/i })).toBeInTheDocument()
  })

  it('abre modal de registro ao clicar em registrar jogo no estado vazio', async () => {
    const user = userEvent.setup()
    renderBiblioteca([])

    await user.click(screen.getByRole('button', { name: /Registrar primeiro jogo/i }))
    expect(await screen.findByRole('heading', { name: 'Registrar jogo' })).toBeVisible()
  })

  it('renderiza os cards de jogos quando existem itens na biblioteca', () => {
    renderBiblioteca([jogoMock])

    expect(screen.getByText('1 jogo')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Chrono Trigger' })).toBeInTheDocument()
    expect(screen.getByText('SNES')).toBeInTheDocument()
    expect(screen.getByText('JRPG')).toBeInTheDocument()
    expect(screen.getByText('Dif. A')).toBeInTheDocument()
    expect(screen.getByText('10')).toBeInTheDocument()
    expect(screen.getByText('Destaque')).toBeInTheDocument()
    expect(screen.getByText('20h jogados')).toBeInTheDocument()
  })

  it('filtra jogos por termo de busca no input de filtro', async () => {
    const jogo2: JogoZeradoDTO = {
      ...jogoMock,
      id: 2,
      nome: 'Super Mario World',
      console: 'SNES',
      genero: 'Plataforma',
    }
    const user = userEvent.setup()
    renderBiblioteca([jogoMock, jogo2])

    expect(screen.getByText('Chrono Trigger')).toBeInTheDocument()
    expect(screen.getByText('Super Mario World')).toBeInTheDocument()

    const filtroInput = screen.getByLabelText('Filtrar jogos na biblioteca')
    await user.type(filtroInput, 'Mario')

    expect(screen.queryByText('Chrono Trigger')).not.toBeInTheDocument()
    expect(screen.getByText('Super Mario World')).toBeInTheDocument()
  })

  it('abre modal de edição ao clicar no botão Editar', async () => {
    const user = userEvent.setup()
    renderBiblioteca([jogoMock])

    const btnEditar = screen.getByRole('button', { name: 'Editar Chrono Trigger' })
    await user.click(btnEditar)

    expect(await screen.findByRole('heading', { name: 'Editar registro' })).toBeVisible()
  })

  it('abre modal de exclusão ao clicar no botão Excluir e cancela sem remover', async () => {
    const user = userEvent.setup()
    renderBiblioteca([jogoMock])

    const btnExcluir = screen.getByRole('button', { name: 'Excluir Chrono Trigger' })
    await user.click(btnExcluir)

    expect(await screen.findByRole('heading', { name: 'Excluir registro' })).toBeVisible()
    expect(screen.getByText(/Tem certeza que deseja excluir o registro de "Chrono Trigger"\?/)).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Cancelar' }))
    expect(screen.getByText('Chrono Trigger')).toBeInTheDocument()
  })

  it('não renderiza botão duplicado de registrar jogo no header quando existem jogos', () => {
    renderBiblioteca([jogoMock])
    const pageHeader = screen.getByRole('heading', { name: 'Biblioteca' }).closest('header')
    expect(within(pageHeader!).queryByRole('button')).not.toBeInTheDocument()
  })

  it('alterna entre visualização em grade e lista e persiste no localStorage', async () => {
    const user = userEvent.setup()
    renderBiblioteca([jogoMock])

    const btnLista = screen.getByRole('button', { name: 'Visualização em lista' })
    await user.click(btnLista)

    expect(localStorage.getItem('biblioteca_view_mode')).toBe('list')
    expect(screen.getByRole('button', { name: 'Opções de Chrono Trigger' })).toBeInTheDocument()

    const btnGrade = screen.getByRole('button', { name: 'Visualização em grade' })
    await user.click(btnGrade)

    expect(localStorage.getItem('biblioteca_view_mode')).toBe('grid')
  })

  it('filtra jogos por console, gênero e nota mínima', async () => {
    const jogo2: JogoZeradoDTO = {
      ...jogoMock,
      id: 2,
      nome: 'God of War',
      console: 'PS5',
      genero: 'Ação',
      nota: 8,
    }
    const user = userEvent.setup()
    renderBiblioteca([jogoMock, jogo2])

    const selectConsole = screen.getByLabelText('Filtrar por console')
    await user.click(selectConsole)
    await user.click(screen.getByRole('option', { name: 'PS5' }))

    expect(screen.queryByText('Chrono Trigger')).not.toBeInTheDocument()
    expect(screen.getByText('God of War')).toBeInTheDocument()

    await user.click(selectConsole)
    await user.click(screen.getByRole('option', { name: 'Todos os consoles' }))

    const selectNota = screen.getByLabelText('Filtrar por nota mínima')
    await user.click(selectNota)
    await user.click(screen.getByRole('option', { name: 'Nota 10+' }))

    expect(screen.getByText('Chrono Trigger')).toBeInTheDocument()
    expect(screen.queryByText('God of War')).not.toBeInTheDocument()
  })
})
