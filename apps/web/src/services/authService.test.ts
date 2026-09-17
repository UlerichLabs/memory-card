import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { authService, AuthApiError } from './authService'

describe('authService', () => {
  const originalFetch = globalThis.fetch

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    globalThis.fetch = originalFetch
  })

  it('envia POST para /api/v1/auth/register com os dados corretos e retorna o usuário', async () => {
    const mockUser = {
      id: 1,
      nome: 'Lucas',
      email: 'lucas@example.com',
      idioma: 'pt-BR',
      created_at: '2026-09-17T12:00:00Z',
    }

    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 201,
      json: async () => ({ data: mockUser }),
    })
    globalThis.fetch = mockFetch

    const result = await authService.cadastrar({
      nome: 'Lucas',
      email: 'lucas@example.com',
      senha: 'SenhaForte@123',
    })

    expect(result).toEqual(mockUser)
    expect(mockFetch).toHaveBeenCalledTimes(1)
    const [url, options] = mockFetch.mock.calls[0]
    expect(url).toContain('/api/v1/auth/register')
    expect(options.method).toBe('POST')
    expect(JSON.parse(options.body)).toEqual({
      nome: 'Lucas',
      email: 'lucas@example.com',
      senha: 'SenhaForte@123',
    })
  })

  it('lança AuthApiError com o código da API quando a resposta não é ok', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 409,
      json: async () => ({
        error: {
          codigo: 'auth.register.email_taken',
          mensagem: 'Este email já está em uso.',
        },
      }),
    })
    globalThis.fetch = mockFetch

    await expect(
      authService.cadastrar({
        nome: 'Lucas',
        email: 'existente@example.com',
        senha: 'SenhaForte@123',
      })
    ).rejects.toThrow(AuthApiError)

    try {
      await authService.cadastrar({
        nome: 'Lucas',
        email: 'existente@example.com',
        senha: 'SenhaForte@123',
      })
    } catch (err) {
      expect(err).toBeInstanceOf(AuthApiError)
      expect((err as AuthApiError).codigo).toBe('auth.register.email_taken')
    }
  })

  it('lança AuthApiError com código fallback quando o json da resposta é inválido', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      json: async () => {
        throw new Error('invalid json')
      },
    })
    globalThis.fetch = mockFetch

    try {
      await authService.cadastrar({
        nome: 'Lucas',
        email: 'lucas@example.com',
        senha: 'SenhaForte@123',
      })
      expect.fail('deveria ter lançado erro')
    } catch (err) {
      expect(err).toBeInstanceOf(AuthApiError)
      expect((err as AuthApiError).codigo).toBe('fallback')
    }
  })
})
