import { afterEach, describe, expect, it, vi } from 'vitest'
import { JOGOS_CAMPO_ERRO_MENSAGENS, jogosService } from './jogosService'

describe('jogosService.buscarIGDB', () => {
  afterEach(() => vi.restoreAllMocks())

  it('ignora uma busca cancelada pelo cliente', async () => {
    const controller = new AbortController()
    controller.abort()
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new DOMException('aborted', 'AbortError')))

    await expect(jogosService.buscarIGDB('zel', undefined, controller.signal)).resolves.toEqual([])
  })

  it('mantém timeout como erro 503', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: { codigo: 'igdb.indisponivel' } }), { status: 503 })))

    await expect(jogosService.buscarIGDB('zel')).rejects.toMatchObject({ status: 503 })
  })

  it('trata resposta 499 cancelada pelo servidor sem corpo de erro', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('', { status: 499 })))

    await expect(jogosService.buscarIGDB('zel')).resolves.toEqual([])
  })
})

describe('erros de tamanho dos jogos', () => {
  it('mapeia cada código para o campo correspondente', () => {
    expect(Object.keys(JOGOS_CAMPO_ERRO_MENSAGENS)).toEqual([
      'jogos.nome_muito_longo', 'jogos.console_muito_longo', 'jogos.genero_muito_longo',
      'jogos.tipo_muito_longo', 'jogos.review_muito_longo',
    ])
    expect(JOGOS_CAMPO_ERRO_MENSAGENS['jogos.console_muito_longo'].campo).toBe('console')
  })
})
