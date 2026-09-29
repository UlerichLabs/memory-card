import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import {
  apiRequest,
  configureApi,
  getAccessToken,
  getStoredRefreshToken,
  resetApiState,
  setAccessToken,
  setStoredRefreshToken,
} from './api'

describe('lib/api', () => {
  beforeEach(() => {
    localStorage.clear()
    resetApiState()
  })

  afterEach(() => {
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
    localStorage.clear()
    resetApiState()
  })

  it('anexa header Authorization quando access token existe', async () => {
    setAccessToken('token-valido')
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ data: { ok: true } }), { status: 200 }))
    vi.stubGlobal('fetch', fetchMock)

    const res = await apiRequest<{ ok: boolean }>('/jogos')
    expect(res).toEqual({ ok: true })
    expect(fetchMock.mock.calls[0][1].headers.get('Authorization')).toBe('Bearer token-valido')
  })

  it('401 numa chamada tenta refresh uma vez e repete a chamada original com sucesso', async () => {
    setAccessToken('token-expirado')
    setStoredRefreshToken('refresh-valido')

    let chamadaOriginal = 0
    const fetchMock = vi.fn().mockImplementation(async (url: string) => {
      if (url.includes('/auth/refresh')) {
        return new Response(JSON.stringify({ data: { access_token: 'token-renovado' } }), { status: 200 })
      }
      if (url.includes('/jogos')) {
        chamadaOriginal++
        if (chamadaOriginal === 1) {
          return new Response(JSON.stringify({ error: { codigo: 'auth.session.unauthorized' } }), { status: 401 })
        }
        return new Response(JSON.stringify({ data: [{ id: 10, nome: 'Metroid' }] }), { status: 200 })
      }
      return new Response('{}', { status: 404 })
    })
    vi.stubGlobal('fetch', fetchMock)

    const resultado = await apiRequest<Array<{ id: number; nome: string }>>('/jogos')
    expect(resultado).toEqual([{ id: 10, nome: 'Metroid' }])
    expect(getAccessToken()).toBe('token-renovado')
    const chamadasRefresh = fetchMock.mock.calls.filter((c) => String(c[0]).includes('/auth/refresh'))
    expect(chamadasRefresh).toHaveLength(1)
  })

  it('três chamadas com 401 simultâneas disparam um único refresh e as três são repetidas', async () => {
    setAccessToken('token-antigo')
    setStoredRefreshToken('refresh-valido')

    let refreshContador = 0

    const fetchMock = vi.fn().mockImplementation(async (url: string) => {
      if (url.includes('/auth/refresh')) {
        refreshContador++
        await new Promise((resolve) => setTimeout(resolve, 30))
        return new Response(JSON.stringify({ data: { access_token: 'token-novo' } }), { status: 200 })
      }
      if (url.includes('/endpoint-')) {
        const auth = (fetchMock.mock.calls[fetchMock.mock.calls.length - 1][1] as RequestInit)?.headers
        const authHeader = auth instanceof Headers ? auth.get('Authorization') : null
        if (authHeader === 'Bearer token-novo') {
          return new Response(JSON.stringify({ data: `resposta-${url}` }), { status: 200 })
        }
        return new Response(JSON.stringify({ error: { codigo: 'auth.session.unauthorized' } }), { status: 401 })
      }
      return new Response('{}', { status: 404 })
    })
    vi.stubGlobal('fetch', fetchMock)

    const [res1, res2, res3] = await Promise.all([
      apiRequest<string>('/endpoint-1'),
      apiRequest<string>('/endpoint-2'),
      apiRequest<string>('/endpoint-3'),
    ])

    expect(res1).toContain('/endpoint-1')
    expect(res2).toContain('/endpoint-2')
    expect(res3).toContain('/endpoint-3')
    expect(refreshContador).toBe(1)
    expect(getAccessToken()).toBe('token-novo')
  })

  it('refresh falhando após 401 limpa os tokens e executa callback de falha', async () => {
    setAccessToken('token-invalido')
    setStoredRefreshToken('refresh-expirado')

    const onAuthFailure = vi.fn()
    configureApi({ onAuthFailure })

    const fetchMock = vi.fn().mockImplementation(async (url: string) => {
      if (url.includes('/auth/refresh')) {
        return new Response(JSON.stringify({ error: { codigo: 'auth.session.expired' } }), { status: 401 })
      }
      return new Response(JSON.stringify({ error: { codigo: 'auth.session.unauthorized' } }), { status: 401 })
    })
    vi.stubGlobal('fetch', fetchMock)

    await expect(apiRequest('/jogos')).rejects.toMatchObject({ codigo: 'auth.session.unauthorized', status: 401 })
    expect(getAccessToken()).toBeNull()
    expect(getStoredRefreshToken()).toBeNull()
    expect(onAuthFailure).toHaveBeenCalledTimes(1)
  })
})
