import { AuthApiError } from '@/services/authService'

export type EstadoTokenReset = 'invalido' | 'expirado' | 'utilizado'

export const RECUPERACAO_SENHA_MENSAGENS = {
  sucessoSolicitacao: 'Se este e-mail estiver cadastrado, você receberá as instruções em breve.',
  expirado: 'Este link de recuperação expirou. Solicite um novo.',
  utilizado: 'Este link já foi utilizado. Solicite um novo se necessário.',
  invalido: 'Este link de recuperação é inválido. Solicite um novo.',
  senhaFraca: 'A senha deve ter no mínimo 8 caracteres, incluindo maiúscula, número e caractere especial.',
  rateLimit: 'Muitas tentativas. Tente novamente em 1 hora.',
  falhaRedefinicao: 'Não foi possível redefinir sua senha. Tente novamente mais tarde.',
  falhaValidacao: 'Não foi possível validar este link. Tente novamente mais tarde.',
  sucessoRedefinicao: 'Senha redefinida com sucesso. Faça login para continuar.',
}

export function estadoTokenReset(error: unknown): EstadoTokenReset | null {
  if (!(error instanceof AuthApiError)) return null
  if (error.codigo === 'auth.password_reset.token_expired' || error.status === 410) return 'expirado'
  if (error.codigo === 'auth.password_reset.token_used' || error.status === 409) return 'utilizado'
  if (error.codigo === 'auth.password_reset.invalid_token' || error.status === 400) return 'invalido'
  return null
}

export function mensagemTokenReset(estado: EstadoTokenReset): string {
  return RECUPERACAO_SENHA_MENSAGENS[estado]
}
