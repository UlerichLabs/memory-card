import { beforeEach, describe, expect, it, vi } from 'vitest'
import { dashboardService } from './dashboardService'

describe('dashboardService', () => {
  beforeEach(() => vi.restoreAllMocks())
  it('lê respostas diretas e envia autenticação', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({ total_jogos: 1 }), { status: 200 }))
    await expect(dashboardService.resumo('token')).resolves.toEqual({ total_jogos: 1 })
    expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining('/api/v1/dashboard/resumo'), expect.objectContaining({ headers: expect.any(Headers) }))
  })
  it('codifica gênero e converte erro tipado', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({ error: { codigo: 'dashboard.genero_obrigatorio', mensagem: 'obrigatório' } }), { status: 400 }))
    await expect(dashboardService.tipos('Ação e RPG')).rejects.toMatchObject({ codigo: 'dashboard.genero_obrigatorio', status: 400 })
    expect(globalThis.fetch).toHaveBeenCalledWith(expect.stringContaining('genero=A%C3%A7%C3%A3o%20e%20RPG'), expect.anything())
  })
})
