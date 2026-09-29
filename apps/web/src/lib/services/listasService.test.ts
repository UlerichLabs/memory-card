import { afterEach, describe, expect, it, vi } from 'vitest'
import * as apiModule from '@/lib/api'
import { listasService } from './listasService'
import type { CriarListaPayload, AtualizarListaPayload, AdicionarItemPayload } from '@/types/listas'

describe('listasService', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('listar chama GET /listas', async () => {
    const spy = vi.spyOn(apiModule, 'apiRequest').mockResolvedValue([])
    await listasService.listar('tok')
    expect(spy).toHaveBeenCalledWith('/listas', { method: 'GET', signal: undefined }, 'tok')
  })

  it('obterPorId chama GET /listas/:id', async () => {
    const spy = vi.spyOn(apiModule, 'apiRequest').mockResolvedValue({ id: 5 })
    await listasService.obterPorId(5, 'tok')
    expect(spy).toHaveBeenCalledWith('/listas/5', { method: 'GET', signal: undefined }, 'tok')
  })

  it('criar chama POST /listas com payload serializado', async () => {
    const spy = vi.spyOn(apiModule, 'apiRequest').mockResolvedValue({ id: 10 })
    const payload: CriarListaPayload = { tipo: 'fila', nome: 'Minha Fila' }
    await listasService.criar(payload, 'tok')
    expect(spy).toHaveBeenCalledWith('/listas', { method: 'POST', body: JSON.stringify(payload) }, 'tok')
  })

  it('atualizar chama PUT /listas/:id com payload', async () => {
    const spy = vi.spyOn(apiModule, 'apiRequest').mockResolvedValue({ id: 10 })
    const payload: AtualizarListaPayload = { nome: 'Novo Nome', meta: 10 }
    await listasService.atualizar(10, payload, 'tok')
    expect(spy).toHaveBeenCalledWith('/listas/10', { method: 'PUT', body: JSON.stringify(payload) }, 'tok')
  })

  it('excluir chama DELETE /listas/:id', async () => {
    const spy = vi.spyOn(apiModule, 'apiRequest').mockResolvedValue(undefined)
    await listasService.excluir(10, 'tok')
    expect(spy).toHaveBeenCalledWith('/listas/10', { method: 'DELETE' }, 'tok')
  })

  it('adicionarItem chama POST /listas/:id/itens', async () => {
    const spy = vi.spyOn(apiModule, 'apiRequest').mockResolvedValue({ id: 1 })
    const payload: AdicionarItemPayload = { nome: 'Chrono Trigger', igdb_id: 100 }
    await listasService.adicionarItem(7, payload, 'tok')
    expect(spy).toHaveBeenCalledWith('/listas/7/itens', { method: 'POST', body: JSON.stringify(payload) }, 'tok')
  })

  it('removerItem chama DELETE /listas/:id/itens/:itemId', async () => {
    const spy = vi.spyOn(apiModule, 'apiRequest').mockResolvedValue(undefined)
    await listasService.removerItem(7, 20, 'tok')
    expect(spy).toHaveBeenCalledWith('/listas/7/itens/20', { method: 'DELETE' }, 'tok')
  })

  it('reordenarItens chama PUT /listas/:id/ordem com item_ids', async () => {
    const spy = vi.spyOn(apiModule, 'apiRequest').mockResolvedValue([])
    await listasService.reordenarItens(7, [3, 2, 1], 'tok')
    expect(spy).toHaveBeenCalledWith('/listas/7/ordem', { method: 'PUT', body: JSON.stringify({ item_ids: [3, 2, 1] }) }, 'tok')
  })

  it('associarZeramento chama PUT /listas/:id/itens/:itemId/zeramento', async () => {
    const spy = vi.spyOn(apiModule, 'apiRequest').mockResolvedValue({ id: 20 })
    await listasService.associarZeramento(7, 20, 99, 'tok')
    expect(spy).toHaveBeenCalledWith('/listas/7/itens/20/zeramento', { method: 'PUT', body: JSON.stringify({ jogo_zerado_id: 99 }) }, 'tok')
  })

  it('sincronizarFranquia chama POST /listas/:id/sincronizar', async () => {
    const spy = vi.spyOn(apiModule, 'apiRequest').mockResolvedValue({ adicionados: 2 })
    await listasService.sincronizarFranquia(7, 'tok')
    expect(spy).toHaveBeenCalledWith('/listas/7/sincronizar', { method: 'POST' }, 'tok')
  })

  it('buscarFranquiasIGDB chama GET /igdb/franquias/busca', async () => {
    const spy = vi.spyOn(apiModule, 'apiRequest').mockResolvedValue([{ id: 1, name: 'Mario' }])
    const res = await listasService.buscarFranquiasIGDB('Mario', 'tok')
    expect(spy).toHaveBeenCalledWith('/igdb/franquias/busca?q=Mario', { method: 'GET', signal: undefined }, 'tok')
    expect(res).toEqual([{ id: 1, name: 'Mario' }])
  })

  it('buscarFranquiasIGDB retorna array vazio se sinal for abortado', async () => {
    const controller = new AbortController()
    controller.abort()
    vi.spyOn(apiModule, 'apiRequest').mockRejectedValue(new DOMException('aborted', 'AbortError'))
    const res = await listasService.buscarFranquiasIGDB('Mario', 'tok', controller.signal)
    expect(res).toEqual([])
  })
})
