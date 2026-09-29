import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes, useLocation, useNavigate } from 'react-router-dom'
import { AuthProvider, useAuthStore } from '@/store/authStore'
import { authService } from '@/services/authService'
import { resetApiState } from '@/lib/api'
import { PrivateRoute } from './PrivateRoute'

function Login() {
  const { login } = useAuthStore()
  const navigate = useNavigate()
  const location = useLocation()
  const from = (location.state as { from?: { pathname: string; search?: string } } | null)?.from
  const fromStr = from ? `${from.pathname}${from.search || ''}` : ''
  return (
    <div>
      <h1>Login</h1>
      <p data-testid="from">{fromStr}</p>
      <button onClick={async () => {
        await login({ email: 'lucas@example.com', senha: 'senha' })
        navigate('/')
      }}>Autenticar</button>
    </div>
  )
}
function Conteudo() {
  const { request, sessao } = useAuthStore()
  return (
    <div>
      <span data-testid="user">{sessao?.usuario?.nome}</span>
      <button onClick={() => void request('/biblioteca').catch(() => undefined)}>Biblioteca privada</button>
    </div>
  )
}
function setup(path: string) {
  render(<MemoryRouter initialEntries={[path]}><AuthProvider>
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route element={<PrivateRoute />}><Route path="*" element={<Conteudo />} /></Route>
    </Routes>
  </AuthProvider></MemoryRouter>)
}

describe('PrivateRoute', () => {
  beforeEach(() => {
    localStorage.clear()
    resetApiState()
  })
  afterEach(() => {
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
    localStorage.clear()
    resetApiState()
  })

  it('redireciona visitante sem sessão para login guardando rota', () => {
    setup('/biblioteca?console=SNES')
    expect(screen.getByRole('heading', { name: 'Login' })).toBeVisible()
    expect(screen.queryByText('Biblioteca privada')).not.toBeInTheDocument()
    expect(screen.getByTestId('from').textContent).toContain('/biblioteca?console=SNES')
  })

  it('carga do app com refresh válido renderiza rota privada sem passar por login', async () => {
    localStorage.setItem('refresh_token', 'refresh-valido')
    vi.spyOn(authService, 'refresh').mockResolvedValue({ access_token: 'novo-access' })
    vi.spyOn(authService, 'me').mockResolvedValue({
      id: 1, nome: 'Lucas', email: 'lucas@example.com', idioma: 'pt-BR', created_at: '',
    })

    setup('/biblioteca')
    expect(screen.getByRole('status', { name: 'Carregando sessão' })).toBeInTheDocument()
    expect(await screen.findByRole('button', { name: 'Biblioteca privada' })).toBeVisible()
    expect(screen.getByTestId('user').textContent).toBe('Lucas')
    expect(screen.queryByRole('heading', { name: 'Login' })).not.toBeInTheDocument()
  })

  it('carga do app com refresh inválido limpa estado e redireciona para login guardando rota', async () => {
    localStorage.setItem('refresh_token', 'refresh-expirado')
    vi.spyOn(authService, 'refresh').mockRejectedValue(new Error('expirado'))

    setup('/biblioteca?busca=mario')
    expect(await screen.findByRole('heading', { name: 'Login' })).toBeVisible()
    expect(localStorage.getItem('refresh_token')).toBeNull()
    expect(screen.queryByText('Biblioteca privada')).not.toBeInTheDocument()
    expect(screen.getByTestId('from').textContent).toContain('/biblioteca?busca=mario')
  })

  it.each(['auth.session.unauthorized', 'auth.session.expired'])('renderiza autenticado e redireciona após 401 %s', async (codigo) => {
    vi.spyOn(authService, 'login').mockResolvedValue({
      access_token: 'access', refresh_token: 'refresh',
      usuario: { id: 1, nome: 'Lucas', email: 'lucas@example.com', idioma: 'pt-BR', created_at: '' },
    })
    const user = userEvent.setup()
    setup('/login')
    await user.click(screen.getByRole('button', { name: 'Autenticar' }))
    expect(await screen.findByRole('button', { name: 'Biblioteca privada' })).toBeVisible()
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: { codigo } }), { status: 401 })))
    await user.click(screen.getByRole('button', { name: 'Biblioteca privada' }))
    expect(await screen.findByRole('heading', { name: 'Login' })).toBeVisible()
    expect(screen.queryByText('Biblioteca privada')).not.toBeInTheDocument()
  })
})
