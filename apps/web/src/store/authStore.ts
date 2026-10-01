import { createContext, createElement, useContext, useEffect, useState, type ReactNode } from 'react'
import { authService, AuthApiError, type LoginPayload, type SessaoDTO } from '@/services/authService'
import { setAccessToken, getStoredRefreshToken, setStoredRefreshToken, configureApi, apiRequest, ApiError } from '@/lib/api'

type AuthStore = {
  sessao: SessaoDTO | null; isCarregandoSessao?: boolean
  login: (p: LoginPayload) => Promise<void>; request: <T>(path: string, opts?: RequestInit) => Promise<T>
  refresh: () => Promise<void>; logout: () => Promise<void>
}

export const AuthContext = createContext<AuthStore | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [sessao, setSessao] = useState<SessaoDTO | null>(null)
  const [isCarregandoSessao, setIsCarregandoSessao] = useState<boolean>(() => Boolean(getStoredRefreshToken()))

  useEffect(() => {
    configureApi({
      onAuthFailure: () => { setSessao(null) },
      onTokenRefreshed: (novo) => { setSessao((p) => (p ? { ...p, access_token: novo } : null)) },
    })
  }, [])

  useEffect(() => {
    const refreshToken = getStoredRefreshToken()
    if (!refreshToken) return
    let cancelado = false
    async function restaurar() {
      try {
        const { access_token } = await authService.refresh(refreshToken!)
        setAccessToken(access_token); const usuario = await authService.me(access_token)
        if (!cancelado) setSessao({ access_token, refresh_token: refreshToken!, usuario })
      } catch {
        if (!cancelado) { setAccessToken(null); setStoredRefreshToken(null); setSessao(null) }
      } finally {
        if (!cancelado) setIsCarregandoSessao(false)
      }
    }
    restaurar(); return () => { cancelado = true }
  }, [])

  const store: AuthStore = {
    sessao, isCarregandoSessao,
    async login(payload) {
      const result = await authService.login({ ...payload, email: payload.email.trim() })
      setAccessToken(result.access_token); setStoredRefreshToken(result.refresh_token); setSessao(result)
    },
    async request<T>(path: string, options?: RequestInit): Promise<T> {
      try { return await apiRequest<T>(path, options, sessao?.access_token) }
      catch (e) { throw e instanceof ApiError ? new AuthApiError(e.codigo, e.message, e.status) : e }
    },
    async refresh() {
      const token = sessao?.refresh_token || getStoredRefreshToken()
      if (!token) throw new AuthApiError('auth.session.expired', '', 401)
      try {
        const res = await authService.refresh(token)
        setAccessToken(res.access_token)
        setSessao((prev) => (prev ? { ...prev, access_token: res.access_token } : null))
      } catch (err) {
        setAccessToken(null); setStoredRefreshToken(null); setSessao(null)
        throw err instanceof AuthApiError ? err : new AuthApiError('auth.session.expired', '', 401)
      }
    },
    async logout() {
      const current = sessao, stored = getStoredRefreshToken()
      setAccessToken(null); setStoredRefreshToken(null); setSessao(null)
      if (current || stored) {
        try { await authService.logout(current?.access_token ?? '', current?.refresh_token ?? stored ?? '') } catch {}
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
