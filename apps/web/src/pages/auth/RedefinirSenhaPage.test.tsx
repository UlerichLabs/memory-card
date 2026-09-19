import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { AuthApiError, authService } from '@/services/authService'
import { RedefinirSenhaPage } from './RedefinirSenhaPage'

function renderPage(token = 'token-valido') {
  render(
    <MemoryRouter initialEntries={[`/redefinir-senha?token=${token}`]}>
      <Routes>
        <Route path="/redefinir-senha" element={<RedefinirSenhaPage />} />
        <Route path="/login" element={<p>Tela de login</p>} />
      </Routes>
    </MemoryRouter>
  )
}

async function preencherSenha(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText('Nova senha'), 'SenhaForte@123')
  await user.type(screen.getByLabelText('Confirmação de senha'), 'SenhaForte@123')
}

describe('RedefinirSenhaPage', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ data: {} }))))
  })
  afterEach(() => { vi.restoreAllMocks(); vi.unstubAllGlobals() })

  it('valida o token ao carregar e mostra o formulário quando ele é válido', async () => {
    renderPage()

    expect(await screen.findByLabelText('Nova senha')).toBeVisible()
    expect(globalThis.fetch).toHaveBeenCalledWith(
      expect.stringContaining('/api/v1/auth/validar-token-reset?token=token-valido'),
      expect.objectContaining({ method: 'GET' })
    )
  })

  it.each([
    ['auth.password_reset.invalid_token', 400, 'Este link de recuperação é inválido. Solicite um novo.'],
    ['auth.password_reset.token_expired', 410, 'Este link de recuperação expirou. Solicite um novo.'],
    ['auth.password_reset.token_used', 409, 'Este link já foi utilizado. Solicite um novo se necessário.'],
  ])('mostra erro e ação para solicitar outro link quando o token é %s', async (codigo, status, mensagem) => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({
      error: { codigo, mensagem },
    }), { status })))
    renderPage()

    expect(await screen.findByRole('alert')).toHaveTextContent(mensagem)
    expect(screen.queryByLabelText('Nova senha')).not.toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Solicitar novo link' })).toHaveAttribute('href', '/esqueci-senha')
  })

  it('valida a força da senha no cliente antes de submeter', async () => {
    const user = userEvent.setup()
    const redefinirSenha = vi.spyOn(authService, 'redefinirSenha')
    renderPage()
    await screen.findByLabelText('Nova senha')

    await user.type(screen.getByLabelText('Nova senha'), 'fraca')
    await user.type(screen.getByLabelText('Confirmação de senha'), 'fraca')
    await user.click(screen.getByRole('button', { name: 'Redefinir senha' }))

    expect(screen.getByRole('alert')).toHaveTextContent(
      'A senha deve ter no mínimo 8 caracteres, incluindo maiúscula, número e caractere especial.'
    )
    expect(redefinirSenha).not.toHaveBeenCalled()
  })

  it('envia a nova senha e redireciona para login após sucesso', async () => {
    const user = userEvent.setup()
    const redefinirSenha = vi.spyOn(authService, 'redefinirSenha').mockResolvedValue(undefined)
    renderPage()
    await screen.findByLabelText('Nova senha')

    await preencherSenha(user)
    await user.click(screen.getByRole('button', { name: 'Redefinir senha' }))

    expect(redefinirSenha).toHaveBeenCalledWith({ token: 'token-valido', senha: 'SenhaForte@123' })
    expect(await screen.findByText('Tela de login')).toBeVisible()
  })

  it.each([
    ['auth.password_reset.weak_password', 400, 'A senha deve ter no mínimo 8 caracteres, incluindo maiúscula, número e caractere especial.'],
    ['auth.password_reset.token_used', 409, 'Este link já foi utilizado. Solicite um novo se necessário.'],
    ['auth.password_reset.token_expired', 410, 'Este link de recuperação expirou. Solicite um novo.'],
  ])('trata o erro %s retornado na submissão', async (codigo, status, mensagem) => {
    const user = userEvent.setup()
    vi.spyOn(authService, 'redefinirSenha').mockRejectedValue(new AuthApiError(codigo, mensagem, status))
    renderPage()
    await screen.findByLabelText('Nova senha')

    await preencherSenha(user)
    await user.click(screen.getByRole('button', { name: 'Redefinir senha' }))

    expect(await screen.findByRole('alert')).toHaveTextContent(mensagem)
  })
})
