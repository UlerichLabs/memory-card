import { beforeEach, afterEach, describe, expect, it, vi } from 'vitest'
import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { authService, AuthApiError, type SessaoDTO } from '@/services/authService'
import { AuthProvider } from '@/store/authStore'
import { LoginForm } from './LoginForm'
import { PrivateRoute } from './PrivateRoute'

const sessao: SessaoDTO = {
  access_token: 'access', refresh_token: 'refresh',
  usuario: { id: 1, nome: 'Lucas', email: 'lucas@example.com', idioma: 'pt-BR', created_at: '' },
}
function setup() {
  render(<MemoryRouter initialEntries={['/login']}><AuthProvider><Routes>
    <Route path="/login" element={<LoginForm />} />
    <Route element={<PrivateRoute />}><Route path="/" element={<h1>Início autenticado</h1>} /></Route>
  </Routes></AuthProvider></MemoryRouter>)
  return userEvent.setup()
}
async function preencher(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText('Email'), 'lucas@example.com')
  await user.type(screen.getByLabelText('Senha'), 'Senha@123')
}

describe('LoginForm', () => {
  beforeEach(() => { vi.spyOn(authService, 'login').mockResolvedValue(sessao) })
  afterEach(() => vi.restoreAllMocks())

  it('renderiza campos e botão', () => {
    setup()
    expect(screen.getByLabelText('Email')).toBeVisible()
    expect(screen.getByLabelText('Senha')).toBeVisible()
    expect(screen.getByRole('button', { name: 'Entrar' })).toBeEnabled()
  })
  it('não submete campos vazios', async () => {
    const user = setup()
    await user.click(screen.getByRole('button', { name: 'Entrar' }))
    expect(authService.login).not.toHaveBeenCalled()
    expect(screen.getByText('Informe um email válido.')).toBeVisible()
    expect(screen.getByText('A senha é obrigatória.')).toBeVisible()
  })
  it('não submete email inválido', async () => {
    const user = setup()
    await user.type(screen.getByLabelText('Email'), 'invalido')
    await user.type(screen.getByLabelText('Senha'), 'senha')
    await user.click(screen.getByRole('button', { name: 'Entrar' }))
    expect(authService.login).not.toHaveBeenCalled()
  })
  it.each([
    { codigo: 'auth.login.invalid_credentials', mensagem: 'Email ou senha inválidos.' },
    { codigo: 'fallback', mensagem: 'Não foi possível entrar. Tente novamente mais tarde.' },
    { codigo: '', mensagem: 'Não foi possível entrar. Tente novamente mais tarde.' },
  ])('exibe erro inline sem detalhes internos: %s', async ({ codigo, mensagem }) => {
    vi.mocked(authService.login).mockRejectedValue(codigo
      ? new AuthApiError(codigo, 'detalhe interno', 401) : new Error('network failure'))
    const user = setup()
    await preencher(user)
    await user.click(screen.getByRole('button', { name: 'Entrar' }))
    expect(await screen.findByRole('alert')).toHaveTextContent(mensagem)
    expect(screen.getByRole('button', { name: 'Entrar' })).toBeEnabled()
  })
  it('desabilita durante o loading, envia dados corretos e redireciona após salvar sessão', async () => {
    let resolve!: (value: SessaoDTO) => void
    vi.mocked(authService.login).mockReturnValue(new Promise((done) => { resolve = done }))
    const user = setup()
    await preencher(user)
    await user.click(screen.getByRole('button', { name: 'Entrar' }))
    expect(screen.getByRole('button', { name: 'Entrando...' })).toBeDisabled()
    expect(screen.getByLabelText('Email')).toBeDisabled()
    expect(screen.getByLabelText('Senha')).toBeDisabled()
    expect(authService.login).toHaveBeenCalledExactlyOnceWith({ email: 'lucas@example.com', senha: 'Senha@123' })
    await act(async () => resolve(sessao))
    expect(await screen.findByRole('heading', { name: 'Início autenticado' })).toBeVisible()
  })
})
