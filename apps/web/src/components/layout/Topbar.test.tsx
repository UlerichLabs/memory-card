import { afterEach, describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { AuthProvider, useAuthStore } from '@/store/authStore'
import { authService, type SessaoDTO } from '@/services/authService'
import { Topbar } from './Topbar'

const mockSessao: SessaoDTO = {
  access_token: 'access-123',
  refresh_token: 'refresh-123',
  usuario: {
    id: 1,
    nome: 'Lucas Tester',
    email: 'lucas@example.com',
    idioma: 'pt-BR',
    created_at: '2026-09-18T10:00:00Z',
  },
}

function TopbarWrapper() {
  const { login } = useAuthStore()
  return (
    <>
      <button
        type="button"
        onClick={async () => {
          await login({ email: 'lucas@example.com', senha: 'password' })
        }}
      >
        Login Teste
      </button>
      <Topbar />
    </>
  )
}

function renderTopbar(initialPath = '/') {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <AuthProvider>
        <Routes>
          <Route path="/" element={<TopbarWrapper />} />
          <Route path="/login" element={<h1>Tela de Login</h1>} />
          <Route path="/conta" element={<h1>Tela de Conta</h1>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>
  )
}

describe('Topbar - Menu de usuário', () => {
  afterEach(() => {
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
  })

  it('abre dropdown com as opcoes Conta e Sair ao clicar no perfil', async () => {
    vi.spyOn(authService, 'login').mockResolvedValue(mockSessao)
    const user = userEvent.setup()
    renderTopbar()

    await user.click(screen.getByRole('button', { name: 'Login Teste' }))
    expect(await screen.findByText('Lucas Tester')).toBeInTheDocument()

    const trigger = screen.getByRole('button', { name: 'Perfil do usuário' })
    await user.click(trigger)

    expect(await screen.findByRole('menuitem', { name: /conta/i })).toBeVisible()
    expect(screen.getByRole('menuitem', { name: /sair/i })).toBeVisible()
  })

  it('redireciona para /login ao clicar em Sair', async () => {
    vi.spyOn(authService, 'login').mockResolvedValue(mockSessao)
    const logoutSpy = vi.spyOn(authService, 'logout').mockResolvedValue(undefined)
    const user = userEvent.setup()
    renderTopbar()

    await user.click(screen.getByRole('button', { name: 'Login Teste' }))
    expect(await screen.findByText('Lucas Tester')).toBeInTheDocument()

    const trigger = screen.getByRole('button', { name: 'Perfil do usuário' })
    await user.click(trigger)

    const botaoSair = await screen.findByRole('menuitem', { name: /sair/i })
    await user.click(botaoSair)

    expect(logoutSpy).toHaveBeenCalledWith('access-123', 'refresh-123')
    expect(await screen.findByRole('heading', { name: 'Tela de Login' })).toBeVisible()
  })
})
