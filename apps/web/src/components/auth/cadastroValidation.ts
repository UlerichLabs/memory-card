import { VALIDATION_MESSAGES } from './authConstants'

export interface CadastroFormData {
  nome: string
  email: string
  senha: string
  confirmacaoSenha: string
}

export interface CadastroValidationErrors {
  nome?: string
  email?: string
  senha?: string
  confirmacaoSenha?: string
}

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const SENHA_MAIUSCULA_REGEX = /[A-Z]/
const SENHA_NUMERO_REGEX = /[0-9]/
const SENHA_ESPECIAL_REGEX = /[^A-Za-z0-9]/

export function validarCadastroForm(data: CadastroFormData): CadastroValidationErrors {
  const errors: CadastroValidationErrors = {}

  if (!data.nome.trim()) {
    errors.nome = VALIDATION_MESSAGES.nomeObrigatorio
  }

  if (!data.email.trim() || !EMAIL_REGEX.test(data.email.trim())) {
    errors.email = VALIDATION_MESSAGES.emailInvalido
  }

  const senha = data.senha
  const senhaValida =
    senha.length >= 8 &&
    SENHA_MAIUSCULA_REGEX.test(senha) &&
    SENHA_NUMERO_REGEX.test(senha) &&
    SENHA_ESPECIAL_REGEX.test(senha)

  if (!senhaValida) {
    errors.senha = VALIDATION_MESSAGES.senhaFraca
  }

  if (data.senha !== data.confirmacaoSenha) {
    errors.confirmacaoSenha = VALIDATION_MESSAGES.senhasDivergentes
  }

  return errors
}
