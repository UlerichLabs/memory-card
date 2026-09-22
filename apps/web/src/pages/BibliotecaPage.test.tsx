import { afterEach, describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { AuthProvider } from '@/store/authStore'
import { JogosProvider } from '@/stores/jogosStore'
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
          <Routes>
            <Route path="/biblioteca" element={<BibliotecaPage />} />
            <Route path="/jogos/novo" element={<h1>Tela Novo Jogo</h1>} />
            <Route path="/jogos/:id/editar" element={<h1>Tela Editar Jogo</h1>} />
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

  it('redireciona para /jogos/novo ao clicar em registrar jogo no estado vazio', async () => {
    const user = userEvent.setup()
    renderBiblioteca([])

    await user.click(screen.getByRole('button', { name: /Registrar primeiro jogo/i }))
    expect(await screen.findByRole('heading', { name: 'Tela Novo Jogo' })).toBeVisible()
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

  it('navega para /jogos/:id/editar ao clicar no botão Editar', async () => {
    const user = userEvent.setup()
    renderBiblioteca([jogoMock])

    const btnEditar = screen.getByRole('button', { name: 'Editar Chrono Trigger' })
    await user.click(btnEditar)

    expect(await screen.findByRole('heading', { name: 'Tela Editar Jogo' })).toBeVisible()
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
})
