export const AUTH_API_ERROR_MESSAGES: Record<string, string> = {
  'auth.register.email_taken': 'Este email já está em uso.',
  'auth.register.weak_password':
    'A senha deve ter no mínimo 8 caracteres, incluindo maiúscula, número e caractere especial.',
  'auth.register.invalid_email': 'Informe um email válido.',
  'auth.register.invalid_input': 'Dados de entrada inválidos.',
  'auth.login.invalid_credentials': 'Email ou senha inválidos.',
  'auth.session.unauthorized': 'Não autorizado. Faça login novamente.',
  'auth.session.expired': 'Sessão expirada. Faça login novamente.',
  'auth.login.invalid_input': 'Dados de entrada inválidos.',
  loginFallback: 'Não foi possível entrar. Tente novamente mais tarde.',
  fallback: 'Não foi possível realizar o cadastro. Tente novamente mais tarde.',
}

export const VALIDATION_MESSAGES = {
  senhaObrigatoria: 'A senha é obrigatória.',
  nomeObrigatorio: 'O nome é obrigatório.',
  emailInvalido: 'Informe um email válido.',
  senhaFraca:
    'A senha deve ter no mínimo 8 caracteres, incluindo maiúscula, número e caractere especial.',
  senhasDivergentes: 'Senha e confirmação não conferem.',
}
