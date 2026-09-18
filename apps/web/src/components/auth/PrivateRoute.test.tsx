import { afterEach, describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes, useNavigate } from 'react-router-dom'
import { AuthProvider, useAuthStore } from '@/store/authStore'
import { authService } from '@/services/authService'
import { PrivateRoute } from './PrivateRoute'

function Login() {
  const { login } = useAuthStore()
  const navigate = useNavigate()
  return <button onClick={async () => {
    await login({ email: 'lucas@example.com', senha: 'senha' })
    navigate('/')
  }}>Autenticar</button>
}
function Conteudo() {
  const { request } = useAuthStore()
  return <button onClick={() => void request('/biblioteca').catch(() => undefined)}>Biblioteca privada</button>
}
function setup(path: string) {
  render(<MemoryRouter initialEntries={[path]}><AuthProvider>
    <Login /><Routes>
      <Route path="/login" element={<h1>Login</h1>} />
      <Route element={<PrivateRoute />}><Route path="/" element={<Conteudo />} /></Route>
    </Routes>
  </AuthProvider></MemoryRouter>)
}

describe('PrivateRoute', () => {
  afterEach(() => { vi.restoreAllMocks(); vi.unstubAllGlobals() })
  it('redireciona visitante sem sessão para login', () => {
    setup('/')
    expect(screen.getByRole('heading', { name: 'Login' })).toBeVisible()
    expect(screen.queryByText('Biblioteca privada')).not.toBeInTheDocument()
  })
  it.each(['auth.session.unauthorized', 'auth.session.expired'])('renderiza autenticado e redireciona após 401 %s', async (codigo) => {
    vi.spyOn(authService, 'login').mockResolvedValue({
      access_token: 'access', refresh_token: 'refresh',
      usuario: { id: 1, nome: 'Lucas', email: 'lucas@example.com', idioma: 'pt-BR', created_at: '' },
    })
    const user = userEvent.setup()
    setup('/')
    await user.click(screen.getByRole('button', { name: 'Autenticar' }))
    expect(await screen.findByRole('button', { name: 'Biblioteca privada' })).toBeVisible()
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: { codigo } }), { status: 401 })))
    await user.click(screen.getByRole('button', { name: 'Biblioteca privada' }))
    expect(await screen.findByRole('heading', { name: 'Login' })).toBeVisible()
    expect(screen.queryByText('Biblioteca privada')).not.toBeInTheDocument()
  })
})
