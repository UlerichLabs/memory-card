import { afterEach, describe, expect, it, vi } from 'vitest'
import { act, renderHook } from '@testing-library/react'
import { authService, type SessaoDTO } from '@/services/authService'
import { AuthProvider, useAuthStore } from './authStore'

const sessao: SessaoDTO = {
  access_token: 'access', refresh_token: 'refresh',
  usuario: { id: 1, nome: 'Lucas', email: 'lucas@example.com', idioma: 'pt-BR', created_at: '' },
}
const payload = { email: 'lucas@example.com', senha: 'Senha@123' }
async function autenticado() {
  vi.spyOn(authService, 'login').mockResolvedValue(sessao)
  const hook = renderHook(useAuthStore, { wrapper: AuthProvider })
  await act(async () => hook.result.current.login(payload))
  return hook
}

describe('authStore', () => {
  afterEach(() => { vi.restoreAllMocks(); vi.unstubAllGlobals() })
  it('inicia sem sessão', () => {
    const { result } = renderHook(useAuthStore, { wrapper: AuthProvider })
    expect(result.current.sessao).toBeNull()
  })
  it('salva tokens e usuário apenas em memória', async () => {
    const local = vi.spyOn(Storage.prototype, 'setItem')
    const { result, unmount } = await autenticado()
    expect(result.current.sessao).toEqual(sessao)
    expect(local).not.toHaveBeenCalled()
    unmount()
    expect(renderHook(useAuthStore, { wrapper: AuthProvider }).result.current.sessao).toBeNull()
  })
  it.each(['auth.session.expired', 'auth.session.unauthorized'])('limpa sessão após 401 %s', async (codigo) => {
    const { result } = await autenticado()
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: { codigo } }), { status: 401 })))
    await act(async () => {
      await expect(result.current.request('/biblioteca')).rejects.toMatchObject({ codigo, status: 401 })
    })
    expect(result.current.sessao).toBeNull()
  })
  it('preserva sessão em falha que não representa invalidação', async () => {
    const { result } = await autenticado()
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('{}', { status: 500 })))
    await act(async () => { await expect(result.current.request('/biblioteca')).rejects.toThrow() })
    expect(result.current.sessao).toEqual(sessao)
  })
  it('envia Bearer e retorna o envelope data', async () => {
    const { result } = await autenticado()
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ data: [] })))
    vi.stubGlobal('fetch', fetchMock)
    await expect(result.current.request('/biblioteca')).resolves.toEqual([])
    expect(fetchMock.mock.calls[0][0]).toContain('/api/v1/biblioteca')
    expect(fetchMock.mock.calls[0][1].headers.get('Authorization')).toBe('Bearer access')
  })
  it('renova access token mantendo refresh e usuário', async () => {
    const { result } = await autenticado()
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ data: { access_token: 'novo' } })))
    vi.stubGlobal('fetch', fetchMock)
    await act(async () => result.current.refresh())
    expect(result.current.sessao).toEqual({ ...sessao, access_token: 'novo' })
    expect(fetchMock.mock.calls[0][0]).toContain('/api/v1/auth/refresh')
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({ refresh_token: 'refresh' })
  })
  it('limpa sessão quando refresh expira', async () => {
    const { result } = await autenticado()
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: { codigo: 'auth.session.expired' } }), { status: 401 })))
    await act(async () => { await expect(result.current.refresh()).rejects.toThrow() })
    expect(result.current.sessao).toBeNull()
  })
  it('login envia POST e preserva a senha', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ data: sessao })))
    vi.stubGlobal('fetch', fetchMock)
    const { result } = renderHook(useAuthStore, { wrapper: AuthProvider })
    await act(async () => result.current.login({ email: ' lucas@example.com ', senha: ' senha ' }))
    expect(fetchMock.mock.calls[0][0]).toContain('/api/v1/auth/login')
    expect(fetchMock.mock.calls[0][1].method).toBe('POST')
    expect(JSON.parse(fetchMock.mock.calls[0][1].body)).toEqual({ email: payload.email, senha: ' senha ' })
    expect(result.current.sessao).toEqual(sessao)
  })
})
