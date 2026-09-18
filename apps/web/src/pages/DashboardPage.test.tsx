import { describe, expect, it, vi, afterEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { AuthProvider } from '@/store/authStore'
import * as authStoreModule from '@/store/authStore'
import { DashboardPage } from './DashboardPage'
import { ActiveChallenges } from '@/components/dashboard/ActiveChallenges'
import { StatsRow } from '@/components/dashboard/StatsRow'
import { calcularIntensidade, type DesafioAtivo, type EstatisticasGerais } from '@/mocks/dashboardData'

function renderDashboard() {
  return render(
    <MemoryRouter>
      <AuthProvider>
        <DashboardPage />
      </AuthProvider>
    </MemoryRouter>
  )
}

describe('DashboardPage', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('renderiza os 6 itens de navegação na topbar na ordem correta', () => {
    renderDashboard()
    const linksEsperados = ['Dashboard', 'Biblioteca', 'Abandonados', 'Desafios', 'Listas', 'Explorador']
    const nav = screen.getByRole('navigation', { name: 'Navegação Principal' })
    linksEsperados.forEach((label) => {
      expect(screen.getByRole('link', { name: label })).toBeInTheDocument()
      expect(nav).toHaveTextContent(label)
    })
  })

  it('exibe busca, sino de notificação e botão Registrar jogo na topbar', () => {
    renderDashboard()
    expect(screen.getByPlaceholderText('Buscar jogo ou usuário…')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Notificações' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /\+ Registrar jogo/i })).toBeInTheDocument()
  })

  it('exibe nome do usuário autenticado vindo do authStore e saudação', () => {
    vi.spyOn(authStoreModule, 'useAuthStore').mockReturnValue({
      sessao: {
        access_token: 'token-teste',
        refresh_token: 'refresh-teste',
        usuario: {
          id: 42,
          nome: 'Felipe Gamer',
          email: 'felipe@example.com',
          idioma: 'pt-BR',
          created_at: '2026-01-01',
        },
      },
      login: vi.fn(),
      request: vi.fn(),
      refresh: vi.fn(),
      logout: vi.fn(),
    })

    renderDashboard()
    expect(screen.getAllByText('Felipe Gamer').length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText(/Bem-vindo de volta,/)).toHaveTextContent('Felipe Gamer')
  })

  it('abre o dropdown do usuário ao clicar e exibe opções Conta e Sair', async () => {
    const user = userEvent.setup()
    renderDashboard()
    const trigger = screen.getByRole('button', { name: 'Perfil do usuário' })
    await user.click(trigger)
    expect(await screen.findByRole('menuitem', { name: /conta/i })).toBeInTheDocument()
    expect(screen.getByRole('menuitem', { name: /sair/i })).toBeInTheDocument()
  })

  it('redireciona para /login ao clicar em Sair', async () => {
    const logoutMock = vi.fn().mockResolvedValue(undefined)
    vi.spyOn(authStoreModule, 'useAuthStore').mockReturnValue({
      sessao: {
        access_token: 'token',
        refresh_token: 'refresh',
        usuario: { id: 1, nome: 'Lucas', email: 'lucas@example.com', idioma: 'pt-BR', created_at: '' },
      },
      login: vi.fn(),
      request: vi.fn(),
      refresh: vi.fn(),
      logout: logoutMock,
    })

    const user = userEvent.setup()
    render(
      <MemoryRouter initialEntries={['/']}>
        <Routes>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/login" element={<div>Tela de Login</div>} />
        </Routes>
      </MemoryRouter>
    )

    const trigger = screen.getByRole('button', { name: 'Perfil do usuário' })
    await user.click(trigger)
    const botaoSair = await screen.findByRole('menuitem', { name: /sair/i })
    await user.click(botaoSair)

    expect(logoutMock).toHaveBeenCalledTimes(1)
    expect(await screen.findByText('Tela de Login')).toBeInTheDocument()
  })

  it('renderiza todos os 8 blocos do dashboard na tela', () => {
    renderDashboard()

    expect(screen.getByRole('region', { name: 'Estatísticas Gerais' })).toBeInTheDocument()
    expect(screen.getByRole('region', { name: 'Jogo do Ano' })).toBeInTheDocument()
    expect(screen.getByRole('region', { name: '5 Jogos da Vida' })).toBeInTheDocument()
    expect(screen.getByRole('region', { name: 'Desafios Ativos' })).toBeInTheDocument()
    expect(screen.getByRole('region', { name: 'Atividade do Ano' })).toBeInTheDocument()
    expect(screen.getByRole('region', { name: 'Zerados Recentemente' })).toBeInTheDocument()
    expect(screen.getByRole('region', { name: 'Distribuição por Plataforma' })).toBeInTheDocument()
    expect(screen.getByRole('region', { name: 'Gêneros Mais Jogados' })).toBeInTheDocument()
  })
})

describe('StatsRow', () => {
  it('renderiza exatamente os 4 cards com os valores corretos do mock', () => {
    const estatisticasMock: EstatisticasGerais = {
      totalJogosZerados: 128,
      totalHorasJogadas: 2450,
      totalAbandonados: 14,
      notaMedia: 8.2,
    }

    render(<StatsRow estatisticas={estatisticasMock} />)

    expect(screen.getByText('Jogos zerados')).toBeInTheDocument()
    expect(screen.getByText('128')).toBeInTheDocument()

    expect(screen.getByText('Horas jogadas')).toBeInTheDocument()
    expect(screen.getByText('2.450h')).toBeInTheDocument()

    expect(screen.getByText('Abandonados')).toBeInTheDocument()
    expect(screen.getByText('14')).toBeInTheDocument()

    expect(screen.getByText('Nota média')).toBeInTheDocument()
    expect(screen.getByText('8.2')).toBeInTheDocument()
  })
})

describe('ActiveChallenges', () => {
  it('limita a exibição a no máximo 3 desafios mesmo quando houver mais no mock', () => {
    const desafiosMock: DesafioAtivo[] = [
      { id: '1', titulo: 'Desafio 1', descricao: 'Desc 1', progressoAtual: 1, progressoMeta: 5, unidade: 'jogos', prazo: '31/12' },
      { id: '2', titulo: 'Desafio 2', descricao: 'Desc 2', progressoAtual: 2, progressoMeta: 5, unidade: 'jogos', prazo: '31/12' },
      { id: '3', titulo: 'Desafio 3', descricao: 'Desc 3', progressoAtual: 3, progressoMeta: 5, unidade: 'jogos', prazo: '31/12' },
      { id: '4', titulo: 'Desafio 4', descricao: 'Desc 4', progressoAtual: 4, progressoMeta: 5, unidade: 'jogos', prazo: '31/12' },
    ]

    render(<ActiveChallenges desafios={desafiosMock} />)

    expect(screen.getByText('Desafio 1')).toBeInTheDocument()
    expect(screen.getByText('Desafio 2')).toBeInTheDocument()
    expect(screen.getByText('Desafio 3')).toBeInTheDocument()
    expect(screen.queryByText('Desafio 4')).not.toBeInTheDocument()
  })
})

describe('calcularIntensidade', () => {
  it('retorna os níveis corretos de intensidade com base nas horas jogadas', () => {
    expect(calcularIntensidade(0)).toBe(0)
    expect(calcularIntensidade(1)).toBe(1)
    expect(calcularIntensidade(2)).toBe(1)
    expect(calcularIntensidade(3)).toBe(2)
    expect(calcularIntensidade(4)).toBe(2)
    expect(calcularIntensidade(5)).toBe(3)
    expect(calcularIntensidade(6)).toBe(3)
    expect(calcularIntensidade(7)).toBe(4)
  })
})
