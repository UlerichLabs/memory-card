import { afterEach, describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { AuthApiError, authService } from '@/services/authService'
import { SolicitarResetForm } from './SolicitarResetForm'

function renderForm() {
  render(<MemoryRouter><SolicitarResetForm /></MemoryRouter>)
}

describe('SolicitarResetForm', () => {
  afterEach(() => vi.restoreAllMocks())

  it('envia email válido e exibe a mensagem genérica de sucesso', async () => {
    const user = userEvent.setup()
    const solicitarReset = vi.spyOn(authService, 'solicitarReset').mockResolvedValue(undefined)
    renderForm()

    await user.type(screen.getByLabelText('Email'), '  lucas@example.com  ')
    await user.click(screen.getByRole('button', { name: 'Enviar instruções' }))

    expect(solicitarReset).toHaveBeenCalledWith({ email: 'lucas@example.com' })
    expect(await screen.findByRole('status')).toHaveTextContent(
      'Se este e-mail estiver cadastrado, você receberá as instruções em breve.'
    )
  })

  it('exibe a mensagem de limite de solicitações para erro 429', async () => {
    const user = userEvent.setup()
    vi.spyOn(authService, 'solicitarReset').mockRejectedValue(new AuthApiError('auth.password_reset.rate_limited', '', 429))
    renderForm()

    await user.type(screen.getByLabelText('Email'), 'lucas@example.com')
    await user.click(screen.getByRole('button', { name: 'Enviar instruções' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('Muitas tentativas. Tente novamente em 1 hora.')
  })

  it('mantém o botão e o campo desabilitados enquanto envia a solicitação', async () => {
    const user = userEvent.setup()
    let resolver: () => void = () => undefined
    vi.spyOn(authService, 'solicitarReset').mockReturnValue(new Promise<void>((resolve) => { resolver = resolve }))
    renderForm()

    await user.type(screen.getByLabelText('Email'), 'lucas@example.com')
    await user.click(screen.getByRole('button', { name: 'Enviar instruções' }))

    expect(screen.getByRole('button', { name: 'Enviando...' })).toBeDisabled()
    expect(screen.getByLabelText('Email')).toBeDisabled()
    resolver()
  })
})
