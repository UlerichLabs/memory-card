import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/lib/api'
import { jogandoService } from './jogandoService'

const jogo = { id: 1, nome: 'Hades', igdb_id: 10, igdb_capa_url: 'capa', iniciado_em: '2026-09-20T00:00:00Z' }

describe('jogandoService', () => {
  afterEach(() => vi.restoreAllMocks())

  it('lista com envelope e autorização', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({ data: [jogo] }), { status: 200 }))
    await expect(jogandoService.listar('token')).resolves.toEqual([jogo])
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/jogando', expect.objectContaining({ method: 'GET', headers: expect.any(Headers) }))
    const headers = fetchMock.mock.calls[0][1]?.headers as Headers
    expect(headers.get('Authorization')).toBe('Bearer token')
  })

  it('cria com payload JSON e remove aceitando 204', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(new Response(JSON.stringify({ data: jogo }), { status: 201 }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }))
    await expect(jogandoService.criar({ nome: 'Hades', iniciado_em: '2026-09-20' }, 'token')).resolves.toEqual(jogo)
    await expect(jogandoService.remover(1, 'token')).resolves.toBeUndefined()
    expect(fetchMock.mock.calls[0][1]?.body).toBe(JSON.stringify({ nome: 'Hades', iniciado_em: '2026-09-20' }))
    expect(fetchMock.mock.calls[1][0]).toBe('/api/v1/jogando/1')
  })

  it('converte erro da API em ApiError por codigo', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({ error: { codigo: 'jogando.nome_obrigatorio', mensagem: 'texto' } }), { status: 400 }))
    await expect(jogandoService.listar()).rejects.toSatisfy((error: unknown) => error instanceof ApiError && error.codigo === 'jogando.nome_obrigatorio' && error.status === 400)
  })
})
