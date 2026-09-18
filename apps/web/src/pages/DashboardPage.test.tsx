import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { AuthProvider } from '@/store/authStore'
import { DashboardPage } from './DashboardPage'
import { ActiveChallenges } from '@/components/dashboard/ActiveChallenges'
import { calcularIntensidade, type DesafioAtivo } from '@/mocks/dashboardData'

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
  it('renderiza sem erro e exibe a navegação topbar', () => {
    renderDashboard()
    expect(screen.getByRole('link', { name: /Memory Card/i })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Dashboard' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Biblioteca' })).toBeInTheDocument()
  })

  it('renderiza todos os 8 blocos do dashboard na tela', () => {
    renderDashboard()

    expect(screen.getByRole('region', { name: 'Estatísticas Gerais' })).toBeInTheDocument()
    expect(screen.getByText('Jogos Zerados')).toBeInTheDocument()
    expect(screen.getByText('Horas Jogadas')).toBeInTheDocument()

    expect(screen.getByRole('region', { name: 'Jogo do Ano' })).toBeInTheDocument()
    expect(screen.getByText(/Jogo do Ano 2025/i)).toBeInTheDocument()

    expect(screen.getByRole('region', { name: '5 Jogos da Vida' })).toBeInTheDocument()
    expect(screen.getByText('Chrono Trigger')).toBeInTheDocument()

    expect(screen.getByRole('region', { name: 'Desafios Ativos' })).toBeInTheDocument()

    expect(screen.getByRole('region', { name: 'Atividade do Ano' })).toBeInTheDocument()
    expect(screen.getByText(/Atividade em 2026/i)).toBeInTheDocument()

    expect(screen.getByRole('region', { name: 'Zerados Recentemente' })).toBeInTheDocument()
    expect(screen.getByText('Metroid Prime Remastered')).toBeInTheDocument()

    expect(screen.getByRole('region', { name: 'Distribuição por Plataforma' })).toBeInTheDocument()
    expect(screen.getByText('Por Plataforma')).toBeInTheDocument()

    expect(screen.getByRole('region', { name: 'Gêneros Mais Jogados' })).toBeInTheDocument()
    expect(screen.getByText('Gêneros Mais Jogados')).toBeInTheDocument()
  })
})

describe('ActiveChallenges', () => {
  it('limita a exibição a no máximo 3 desafios mesmo quando houver mais no mock', () => {
    const desafiosMock: DesafioAtivo[] = [
      { id: '1', titulo: 'Desafio 1', descricao: 'Desc 1', progressoAtual: 1, progressoMeta: 5, unidade: 'jogos', prazo: '31/12' },
      { id: '2', titulo: 'Desafio 2', descricao: 'Desc 2', progressoAtual: 2, progressoMeta: 5, unidade: 'jogos', prazo: '31/12' },
      { id: '3', titulo: 'Desafio 3', descricao: 'Desc 3', progressoAtual: 3, progressoMeta: 5, unidade: 'jogos', prazo: '31/12' },
      { id: '4', titulo: 'Desafio 4', descricao: 'Desc 4', progressoAtual: 4, progressoMeta: 5, unidade: 'jogos', prazo: '31/12' },
      { id: '5', titulo: 'Desafio 5', descricao: 'Desc 5', progressoAtual: 5, progressoMeta: 5, unidade: 'jogos', prazo: '31/12' },
    ]

    render(<ActiveChallenges desafios={desafiosMock} />)

    expect(screen.getByText('Desafio 1')).toBeInTheDocument()
    expect(screen.getByText('Desafio 2')).toBeInTheDocument()
    expect(screen.getByText('Desafio 3')).toBeInTheDocument()
    expect(screen.queryByText('Desafio 4')).not.toBeInTheDocument()
    expect(screen.queryByText('Desafio 5')).not.toBeInTheDocument()
  })
})

describe('calcularIntensidade', () => {
  it('retorna os níveis corretos de intensidade com base nas horas jogadas', () => {
    expect(calcularIntensidade(0)).toBe(0)
    expect(calcularIntensidade(-1)).toBe(0)
    expect(calcularIntensidade(1)).toBe(1)
    expect(calcularIntensidade(2)).toBe(1)
    expect(calcularIntensidade(3)).toBe(2)
    expect(calcularIntensidade(4)).toBe(2)
    expect(calcularIntensidade(5)).toBe(3)
    expect(calcularIntensidade(6)).toBe(3)
    expect(calcularIntensidade(7)).toBe(4)
    expect(calcularIntensidade(12)).toBe(4)
  })
})
