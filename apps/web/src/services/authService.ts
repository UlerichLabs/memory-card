export interface CadastroPayload {
  nome: string
  email: string
  senha: string
}

export interface UsuarioDTO {
  id: number
  nome: string
  email: string
  idioma: string
  created_at: string
}

export class AuthApiError extends Error {
  readonly codigo: string

  constructor(codigo: string, message: string) {
    super(message)
    this.name = 'AuthApiError'
    this.codigo = codigo
  }
}

const baseUrl = (import.meta.env.VITE_API_URL || 'http://localhost:18080').replace(/\/+$/, '')

async function cadastrar(payload: CadastroPayload): Promise<UsuarioDTO> {
  const response = await fetch(`${baseUrl}/api/v1/auth/register`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json',
    },
    body: JSON.stringify(payload),
  })

  if (!response.ok) {
    let codigo = 'fallback'
    let mensagem = ''
    try {
      const data = await response.json()
      if (data?.error?.codigo) {
        codigo = data.error.codigo
        mensagem = data.error.mensagem || ''
      }
    } catch {
      codigo = 'fallback'
    }
    throw new AuthApiError(codigo, mensagem)
  }

  const json = await response.json()
  return json.data
}

export const authService = {
  cadastrar,
}
