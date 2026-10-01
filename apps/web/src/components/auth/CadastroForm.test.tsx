import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { CadastroForm } from './CadastroForm'
import { authService, AuthApiError } from '@/services/authService'

const mockNavigate = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

function renderComponent() {
  return render(
    <MemoryRouter>
      <CadastroForm />
    </MemoryRouter>
  )
}

describe('CadastroForm', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renderiza o formulário corretamente com todos os campos', () => {
    renderComponent()

    expect(screen.getByLabelText(/^nome/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/^email/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/^senha$/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/confirmação de senha/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /cadastrar/i })).toBeInTheDocument()
  })

  it('não submete com campos vazios e exibe erro de validação por campo', async () => {
    const user = userEvent.setup()
    const cadastrarSpy = vi.spyOn(authService, 'cadastrar')
    renderComponent()

    await user.click(screen.getByRole('button', { name: /cadastrar/i }))

    expect(screen.getByText('O nome é obrigatório.')).toBeInTheDocument()
    expect(screen.getByText('Informe um email válido.')).toBeInTheDocument()
    expect(
      screen.getByText(
        'A senha deve ter no mínimo 8 caracteres, incluindo maiúscula, número e caractere especial.'
      )
    ).toBeInTheDocument()
    expect(cadastrarSpy).not.toHaveBeenCalled()
  })

  it('exibe erro de "senha e confirmação não conferem" quando os dois campos divergem', async () => {
    const user = userEvent.setup()
    const cadastrarSpy = vi.spyOn(authService, 'cadastrar')
    renderComponent()

    await user.type(screen.getByLabelText(/^nome/i), 'Lucas')
    await user.type(screen.getByLabelText(/^email/i), 'lucas@example.com')
    await user.type(screen.getByLabelText(/^senha$/i), 'SenhaForte@123')
    await user.type(screen.getByLabelText(/confirmação de senha/i), 'OutraSenha@456')

    await user.click(screen.getByRole('button', { name: /cadastrar/i }))

    expect(screen.getByText(/senha e confirmação não conferem/i)).toBeInTheDocument()
    expect(cadastrarSpy).not.toHaveBeenCalled()
  })

  const apiErrorCases = [
    {
      codigo: 'auth.register.email_taken',
      mensagemEsperada: 'Este email já está em uso.',
    },
    {
      codigo: 'auth.register.weak_password',
      mensagemEsperada:
        'A senha deve ter no mínimo 8 caracteres, incluindo maiúscula, número e caractere especial.',
    },
    {
      codigo: 'auth.register.invalid_email',
      mensagemEsperada: 'Informe um email válido.',
    },
    {
      codigo: 'auth.register.invalid_input',
      mensagemEsperada: 'Dados de entrada inválidos.',
    },
    {
      codigo: 'outro.erro.desconhecido',
      mensagemEsperada: 'Não foi possível realizar o cadastro. Tente novamente mais tarde.',
    },
  ]

  for (const tc of apiErrorCases) {
    it(`exibe erro inline mapeado para o código "${tc.codigo}" retornado pela API`, async () => {
      const user = userEvent.setup()
      vi.spyOn(authService, 'cadastrar').mockRejectedValue(
        new AuthApiError(tc.codigo, 'erro api')
      )
      renderComponent()

      await user.type(screen.getByLabelText(/^nome/i), 'Lucas')
      await user.type(screen.getByLabelText(/^email/i), 'lucas@example.com')
      await user.type(screen.getByLabelText(/^senha$/i), 'SenhaForte@123')
      await user.type(screen.getByLabelText(/confirmação de senha/i), 'SenhaForte@123')

      await user.click(screen.getByRole('button', { name: /cadastrar/i }))

      expect(await screen.findByText(tc.mensagemEsperada)).toBeInTheDocument()
    })
  }

  it('exibe mensagem de fallback em caso de erro genérico ou de rede', async () => {
    const user = userEvent.setup()
    vi.spyOn(authService, 'cadastrar').mockRejectedValue(new Error('Network failure'))
    renderComponent()

    await user.type(screen.getByLabelText(/^nome/i), 'Lucas')
    await user.type(screen.getByLabelText(/^email/i), 'lucas@example.com')
    await user.type(screen.getByLabelText(/^senha$/i), 'SenhaForte@123')
    await user.type(screen.getByLabelText(/confirmação de senha/i), 'SenhaForte@123')

    await user.click(screen.getByRole('button', { name: /cadastrar/i }))

    expect(
      await screen.findByText('Não foi possível realizar o cadastro. Tente novamente mais tarde.')
    ).toBeInTheDocument()
  })

  it('mantém o botão desabilitado e com estado de loading durante a requisição', async () => {
    const user = userEvent.setup()
    let resolvePromise: (val: unknown) => void = () => {}
    const pendingPromise = new Promise((resolve) => {
      resolvePromise = resolve
    })
    vi.spyOn(authService, 'cadastrar').mockReturnValue(
      pendingPromise as unknown as ReturnType<typeof authService.cadastrar>
    )
    renderComponent()

    await user.type(screen.getByLabelText(/^nome/i), 'Lucas')
    await user.type(screen.getByLabelText(/^email/i), 'lucas@example.com')
    await user.type(screen.getByLabelText(/^senha$/i), 'SenhaForte@123')
    await user.type(screen.getByLabelText(/confirmação de senha/i), 'SenhaForte@123')

    await user.click(screen.getByRole('button', { name: /cadastrar/i }))

    const submitBtn = screen.getByRole('button', { name: /cadastrando/i })
    expect(submitBtn).toBeDisabled()

    resolvePromise({
      id: 1,
      nome: 'Lucas',
      email: 'lucas@example.com',
      idioma: 'pt-BR',
      created_at: new Date().toISOString(),
    })

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/login')
    })
  })

  it('chama authService.cadastrar com os dados corretos sem a confirmação de senha', async () => {
    const user = userEvent.setup()
    const cadastrarSpy = vi.spyOn(authService, 'cadastrar').mockResolvedValue({
      id: 1,
      nome: 'Lucas',
      email: 'lucas@example.com',
      idioma: 'pt-BR',
      created_at: new Date().toISOString(),
    })
    renderComponent()

    await user.type(screen.getByLabelText(/^nome/i), '  Lucas  ')
    await user.type(screen.getByLabelText(/^email/i), '  lucas@example.com  ')
    await user.type(screen.getByLabelText(/^senha$/i), 'SenhaForte@123')
    await user.type(screen.getByLabelText(/confirmação de senha/i), 'SenhaForte@123')

    await user.click(screen.getByRole('button', { name: /cadastrar/i }))

    await waitFor(() => {
      expect(cadastrarSpy).toHaveBeenCalledTimes(1)
      expect(cadastrarSpy).toHaveBeenCalledWith({
        nome: 'Lucas',
        email: 'lucas@example.com',
        senha: 'SenhaForte@123',
      })
    })
  })

  it('redireciona para /login em caso de sucesso (201)', async () => {
    const user = userEvent.setup()
    vi.spyOn(authService, 'cadastrar').mockResolvedValue({
      id: 1,
      nome: 'Lucas',
      email: 'lucas@example.com',
      idioma: 'pt-BR',
      created_at: new Date().toISOString(),
    })
    renderComponent()

    await user.type(screen.getByLabelText(/^nome/i), 'Lucas')
    await user.type(screen.getByLabelText(/^email/i), 'lucas@example.com')
    await user.type(screen.getByLabelText(/^senha$/i), 'SenhaForte@123')
    await user.type(screen.getByLabelText(/confirmação de senha/i), 'SenhaForte@123')

    await user.click(screen.getByRole('button', { name: /cadastrar/i }))

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/login')
    })
  })
})
