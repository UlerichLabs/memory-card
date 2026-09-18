import { createContext, createElement, useContext, useState, type ReactNode } from 'react'
import { authService, AuthApiError, type LoginPayload, type SessaoDTO } from '@/services/authService'

type AuthStore = {
  sessao: SessaoDTO | null
  login: (payload: LoginPayload) => Promise<void>
  request: <T>(path: string, options?: RequestInit) => Promise<T>
  refresh: () => Promise<void>
  logout: () => Promise<void>
}
const AuthContext = createContext<AuthStore | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [sessao, setSessao] = useState<SessaoDTO | null>(null)
  function limparSessao() {
    setSessao(null)
  }
  async function authenticated<T>(operation: (current: SessaoDTO) => Promise<T>): Promise<T> {
    if (!sessao) throw new AuthApiError('auth.session.unauthorized', '', 401)
    try {
      return await operation(sessao)
    } catch (error) {
      if (error instanceof AuthApiError && error.status === 401 &&
        ['auth.session.unauthorized', 'auth.session.expired'].includes(error.codigo)) {
        limparSessao()
      }
      throw error
    }
  }
  const store: AuthStore = {
    sessao,
    async login(payload) {
      const result = await authService.login({ ...payload, email: payload.email.trim() })
      setSessao(result)
    },
    request: <T,>(path: string, options?: RequestInit) => authenticated((current) =>
      authService.authenticatedRequest<T>(path, current.access_token, options)),
    async refresh() {
      const result = await authenticated((current) => authService.refresh(current.refresh_token))
      setSessao((current) => current === sessao && current ? { ...current, ...result } : current)
    },
    async logout() {
      const current = sessao
      limparSessao()
      if (current) {
        try {
          await authService.logout(current.access_token, current.refresh_token)
        } catch {}
      }
    },
  }
  return createElement(AuthContext.Provider, { value: store }, children)
}

export function useAuthStore(): AuthStore {
  const store = useContext(AuthContext)
  if (!store) throw new AuthApiError('auth.session.unauthorized', '', 401)
  return store
}
