import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { AuthApiError, authService, type SessaoDTO } from '@/services/authService'
import { AuthProvider, useAuthStore } from '@/store/authStore'
import { TrocarSenhaForm } from './TrocarSenhaForm'

const sessao: SessaoDTO = {
  access_token: 'access',
  refresh_token: 'refresh',
  usuario: { id: 1, nome: 'Lucas', email: 'lucas@example.com', idioma: 'pt-BR', created_at: '' },
}

function ConteudoAutenticado() {
  const { login } = useAuthStore()
  return <><button onClick={() => login({ email: 'lucas@example.com', senha: 'Senha@123' })}>Autenticar</button><TrocarSenhaForm /></>
}

function setup() {
  vi.spyOn(authService, 'login').mockResolvedValue(sessao)
  render(<MemoryRouter><AuthProvider><ConteudoAutenticado /></AuthProvider></MemoryRouter>)
  return userEvent.setup()
}

async function autenticar(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole('button', { name: 'Autenticar' }))
}

async function preencher(user: ReturnType<typeof userEvent.setup>, novaSenha = 'SenhaNova@123') {
  await user.type(screen.getByLabelText('Senha atual'), 'SenhaAtual@123')
  await user.type(screen.getByLabelText('Nova senha'), novaSenha)
  await user.type(screen.getByLabelText('Confirmação da nova senha'), novaSenha)
}

describe('TrocarSenhaForm', () => {
  beforeEach(() => {
    vi.spyOn(authService, 'login').mockResolvedValue(sessao)
    vi.spyOn(authService, 'trocarSenha').mockResolvedValue({ mensagem: 'ok' })
  })
  afterEach(() => { vi.restoreAllMocks() })

  it('valida confirmação divergente sem chamar a API', async () => {
    const user = setup()
    await autenticar(user)
    await user.type(screen.getByLabelText('Senha atual'), 'SenhaAtual@123')
    await user.type(screen.getByLabelText('Nova senha'), 'SenhaNova@123')
    await user.type(screen.getByLabelText('Confirmação da nova senha'), 'OutraSenha@123')
    await user.click(screen.getByRole('button', { name: 'Alterar senha' }))

    expect(screen.getByRole('alert')).toHaveTextContent('Senha e confirmação não conferem.')
    expect(authService.trocarSenha).not.toHaveBeenCalled()
  })

  it('valida força da nova senha sem chamar a API', async () => {
    const user = setup()
    await autenticar(user)
    await preencher(user, 'fraca')
    await user.click(screen.getByRole('button', { name: 'Alterar senha' }))

    expect(screen.getByRole('alert')).toHaveTextContent(
      'A senha deve ter no mínimo 8 caracteres, incluindo maiúscula, número e caractere especial.'
    )
    expect(authService.trocarSenha).not.toHaveBeenCalled()
  })

  it('envia dados, mostra sucesso e limpa o formulário', async () => {
    const user = setup()
    await autenticar(user)
    const trocarSenha = vi.mocked(authService.trocarSenha)
    await preencher(user)
    await user.click(screen.getByRole('button', { name: 'Alterar senha' }))

    expect(trocarSenha).toHaveBeenCalledWith('access', { senha_atual: 'SenhaAtual@123', nova_senha: 'SenhaNova@123' })
    expect(await screen.findByRole('status')).toHaveTextContent('Senha alterada com sucesso.')
    expect(screen.getByLabelText('Senha atual')).toHaveValue('')
    expect(screen.getByLabelText('Nova senha')).toHaveValue('')
  })

  it.each([
    ['auth.password_change.current_password_invalid', 'Senha atual incorreta.'],
    ['auth.password_change.weak_password', 'A senha deve ter no mínimo 8 caracteres, incluindo maiúscula, número e caractere especial.'],
    ['auth.password_change.same_password', 'A nova senha deve ser diferente da atual.'],
  ])('exibe o erro do servidor %s', async (codigo, mensagem) => {
    const user = setup()
    await autenticar(user)
    vi.mocked(authService.trocarSenha).mockRejectedValue(new AuthApiError(codigo, mensagem, 400))
    await preencher(user)
    await user.click(screen.getByRole('button', { name: 'Alterar senha' }))

    expect(await screen.findByRole('alert')).toHaveTextContent(mensagem)
  })

  it('desabilita o formulário durante o loading', async () => {
    const user = setup()
    await autenticar(user)
    let resolve!: (value: { mensagem: string }) => void
    vi.mocked(authService.trocarSenha).mockReturnValue(new Promise((done) => { resolve = done }))
    await preencher(user)
    await user.click(screen.getByRole('button', { name: 'Alterar senha' }))

    expect(screen.getByRole('button', { name: 'Alterando...' })).toBeDisabled()
    expect(screen.getByLabelText('Senha atual')).toBeDisabled()
    await act(async () => resolve({ mensagem: 'ok' }))
  })
})
