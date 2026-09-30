import { afterEach, describe, expect, it, vi } from 'vitest'
import { AbandonadosApiError, abandonadosService } from './abandonadosService'
import {
  ehCancelado,
  formatarTempoAbandonado,
  montarQueryStringAbandonados,
  segundosParaHms,
} from '@/components/abandonados/abandonados.utils'

describe('abandonados.utils', () => {
  it('monta query string corretamente', () => {
    expect(montarQueryStringAbandonados()).toBe('')
    expect(montarQueryStringAbandonados({ pagina: 1, por_pagina: 12, ordenar: 'recentes' })).toBe('')
    expect(
      montarQueryStringAbandonados({
        pagina: 2,
        por_pagina: 24,
        busca: 'Zelda',
        console: 'Switch',
        ordenar: 'tempo',
      })
    ).toBe('pagina=2&por_pagina=24&busca=Zelda&console=Switch&ordenar=tempo')
  })

  it('formata tempo abandonado em horas e minutos', () => {
    expect(formatarTempoAbandonado(0)).toBe('0h')
    expect(formatarTempoAbandonado(3600)).toBe('1h')
    expect(formatarTempoAbandonado(5400)).toBe('1h 30m')
    expect(formatarTempoAbandonado(1800)).toBe('30m')
    expect(formatarTempoAbandonado(45)).toBe('45s')
  })

  it('converte segundos para hms', () => {
    expect(segundosParaHms(0)).toEqual({ horas: '', minutos: '', segundos: '' })
    expect(segundosParaHms(3665)).toEqual({ horas: '1', minutos: '1', segundos: '5' })
  })

  it('identifica requisicao cancelada', () => {
    expect(ehCancelado(new DOMException('Aborted', 'AbortError'))).toBe(true)
    expect(ehCancelado(new AbandonadosApiError('cancel', '', 499))).toBe(true)
    expect(ehCancelado(new Error('Outro erro'))).toBe(false)
  })
})

describe('abandonadosService', () => {
  afterEach(() => vi.restoreAllMocks())

  it('lista jogos abandonados com sucesso', async () => {
    const mockResposta = {
      data: [{ id: 1, nome: 'Metroid', console: 'NES', tempo_jogado: 3600 }],
      meta: { pagina: 1, por_pagina: 12, total: 1, total_paginas: 1 },
    }
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: async () => mockResposta,
    } as Response)

    const resultado = await abandonadosService.listar({ busca: 'Metroid' })
    expect(resultado.data).toHaveLength(1)
    expect(resultado.data[0].nome).toBe('Metroid')
  })

  it('lida com erro de API lancando AbandonadosApiError', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce({
      ok: false,
      status: 400,
      json: async () => ({ error: { codigo: 'abandonados.tempo_invalido', mensagem: 'Tempo inválido' } }),
    } as Response)

    await expect(
      abandonadosService.criar({ nome: 'X', console: 'Y', abandonado_em: '2026-01-01', tempo_jogado_horas: 999999 })
    ).rejects.toMatchObject({
      name: 'AbandonadosApiError',
      codigo: 'abandonados.tempo_invalido',
      status: 400,
    })
  })

  it('retorna fallback vazio se listagem for abortada', async () => {
    vi.spyOn(globalThis, 'fetch').mockRejectedValueOnce(new DOMException('Aborted', 'AbortError'))
    const controller = new AbortController()
    controller.abort()

    const resultado = await abandonadosService.listar({ pagina: 2 }, undefined, controller.signal)
    expect(resultado.data).toEqual([])
    expect(resultado.meta.pagina).toBe(2)
  })

  it('obtem filtros e total', async () => {
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ data: { consoles: ['SNES', 'PS1'] } }),
      } as Response)
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: async () => ({ data: { total: 5 } }),
      } as Response)

    const filtros = await abandonadosService.obterFiltros()
    expect(filtros.consoles).toEqual(['SNES', 'PS1'])

    const total = await abandonadosService.obterTotal()
    expect(total.total).toBe(5)
  })
})
