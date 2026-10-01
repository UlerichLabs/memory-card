import { afterEach, describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { AuthProvider, useAuthStore } from '@/store/authStore'
import { JogosProvider } from '@/stores/jogosStore'
import { JogandoProvider } from '@/stores/jogandoStore'
import { GameFormDialog } from '@/components/jogos/GameForm/GameFormDialog'
import { IniciarJogoDialog } from '@/components/jogando/IniciarJogoDialog'
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
        <JogosProvider>
          <JogandoProvider>
          <GameFormDialog />
          <IniciarJogoDialog />
          <Routes>
            <Route path="/login" element={<h1>Tela de Login</h1>} />
            <Route path="/conta" element={<h1>Tela de Conta</h1>} />
            <Route path="*" element={<TopbarWrapper />} />
          </Routes>
          </JogandoProvider>
        </JogosProvider>
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

  it('abre o menu Novo e o modal de registro de jogo zerado', async () => {
    const user = userEvent.setup()
    renderTopbar()

    await user.click(screen.getByRole('button', { name: /\+ Novo/ }))
    await user.click(await screen.findByRole('menuitem', { name: /Registrar jogo zerado/ }))

    expect(await screen.findByRole('heading', { name: 'Registrar jogo' })).toBeVisible()
  })

  it('abre o dialogo de iniciar jogo pelo menu Novo', async () => {
    const user = userEvent.setup()
    renderTopbar()
    await user.click(screen.getByRole('button', { name: /\+ Novo/ }))
    await user.click(await screen.findByRole('menuitem', { name: /Iniciar jogo/ }))
    expect(await screen.findByRole('heading', { name: 'Iniciar jogo' })).toBeVisible()
  })

  it('exibe o item Hall da Fama no menu de navegação apontando para /hall-da-fama', () => {
    renderTopbar()
    const link = screen.getByRole('link', { name: 'Hall da Fama' })
    expect(link).toBeInTheDocument()
    expect(link).toHaveAttribute('href', '/hall-da-fama')
  })

  it('exibe o item Hall da Fama com estado ativo quando na rota /hall-da-fama', () => {
    renderTopbar('/hall-da-fama')
    const link = screen.getByRole('link', { name: 'Hall da Fama' })
    expect(link).toHaveClass('border-[var(--nav-link-active-border)]')
    expect(link).toHaveClass('text-[var(--nav-link-active-text)]')

    const linkBiblioteca = screen.getByRole('link', { name: 'Biblioteca' })
    expect(linkBiblioteca).toHaveClass('border-transparent')
  })

  it('exibe item único Listas e Desafios apontando para /listas e ativo em /listas/5', () => {
    renderTopbar('/listas/5')
    const link = screen.getByRole('link', { name: 'Listas e Desafios' })
    expect(link).toBeInTheDocument()
    expect(link).toHaveAttribute('href', '/listas')
    expect(link).toHaveClass('border-[var(--nav-link-active-border)]')

    expect(screen.queryByRole('link', { name: /^Desafios$/ })).not.toBeInTheDocument()
    expect(screen.queryByRole('link', { name: /^Listas$/ })).not.toBeInTheDocument()
  })
})
