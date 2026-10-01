import { afterEach, describe, expect, it, vi } from 'vitest'
import { JOGOS_CAMPO_ERRO_MENSAGENS, jogosService, montarQueryString } from './jogosService'

describe('montarQueryString', () => {
  it('retorna string vazia quando não há parâmetros', () => {
    expect(montarQueryString()).toBe('')
    expect(montarQueryString({})).toBe('')
  })

  it('omite parâmetros de paginação padrão', () => {
    expect(montarQueryString({ pagina: 1, por_pagina: 24 })).toBe('')
  })

  it('inclui página e por_pagina quando não forem padrão', () => {
    const qs = montarQueryString({ pagina: 2, por_pagina: 10 })
    expect(qs).toContain('pagina=2')
    expect(qs).toContain('por_pagina=10')
  })

  it('monta e codifica filtros textuais e numéricos sem campos vazios', () => {
    const qs = montarQueryString({
      busca: 'Chrono Trigger',
      console: 'Super Nintendo',
      genero: 'RPG, Aventura',
      tipo: 'Campanha',
      nota_min: 8,
      nota_max: 10,
      ano: 2026,
      dificuldade: 'A',
    })
    expect(qs).toContain('busca=Chrono+Trigger')
    expect(qs).toContain('console=Super+Nintendo')
    expect(qs).toContain('genero=RPG%2C+Aventura')
    expect(qs).toContain('tipo=Campanha')
    expect(qs).toContain('nota_min=8')
    expect(qs).toContain('nota_max=10')
    expect(qs).toContain('ano=2026')
    expect(qs).toContain('dificuldade=A')
  })

  it('adiciona parâmetro ordenar quando especificado', () => {
    expect(montarQueryString({ ordenar: 'nota' })).toBe('ordenar=nota')
    expect(montarQueryString({ ordenar: 'recentes' })).toBe('ordenar=recentes')
  })

  it('ignora parâmetros vazios ou com espaços em branco', () => {
    const qs = montarQueryString({
      busca: '   ',
      console: '',
      genero: undefined,
    })
    expect(qs).toBe('')
  })
})

describe('jogosService.listar', () => {
  afterEach(() => vi.restoreAllMocks())

  it('chama endpoint correto e retorna data e meta', async () => {
    const mockResposta = {
      data: [{ id: 1, nome: 'Zelda' }],
      meta: { pagina: 1, por_pagina: 24, total: 1, total_paginas: 1 },
    }
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify(mockResposta), { status: 200 })))

    const resultado = await jogosService.listar({ busca: 'Zelda' })
    expect(resultado).toEqual(mockResposta)
  })

  it('ignora cancelamento via AbortSignal sem lançar erro', async () => {
    const controller = new AbortController()
    controller.abort()
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new DOMException('aborted', 'AbortError')))

    await expect(jogosService.listar({}, undefined, controller.signal)).resolves.toEqual({
      data: [],
      meta: { pagina: 1, por_pagina: 24, total: 0, total_paginas: 0 },
    })
  })

  it('trata erro 499 do servidor como cancelamento silencioso', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('', { status: 499 })))

    await expect(jogosService.listar()).resolves.toEqual({
      data: [],
      meta: { pagina: 1, por_pagina: 24, total: 0, total_paginas: 0 },
    })
  })
})

describe('jogosService.obterFiltros', () => {
  afterEach(() => vi.restoreAllMocks())

  it('retorna opções de filtros da api', async () => {
    const mockFiltros = {
      consoles: ['SNES', 'PS1'],
      generos: ['RPG'],
      tipos: ['Campanha'],
      anos: [2026, 2025],
    }
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ data: mockFiltros }), { status: 200 })))

    const resultado = await jogosService.obterFiltros()
    expect(resultado).toEqual(mockFiltros)
  })
})

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

describe('jogosService - game do ano', () => {
  afterEach(() => vi.restoreAllMocks())

  it('obterResumoGameDoAno busca resumo na rota /jogos/game-do-ano', async () => {
    const mockResumo = [{ ano: 2026, total_jogos: 5, game_do_ano: null }]
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ data: mockResumo }), { status: 200 })))

    const res = await jogosService.obterResumoGameDoAno('fake-token')
    expect(res).toEqual(mockResumo)
  })

  it('definirGameDoAno envia PUT para /jogos/:id/game-do-ano', async () => {
    const mockResp = { ano: 2026, anterior_id: null, game_do_ano: { id: 10, nome: 'Zelda' } }
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ data: mockResp }), { status: 200 })))

    const res = await jogosService.definirGameDoAno(10, 'fake-token')
    expect(res).toEqual(mockResp)
  })

  it('removerGameDoAno envia DELETE para /jogos/:id/game-do-ano', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(null, { status: 204 })))

    await expect(jogosService.removerGameDoAno(10, 'fake-token')).resolves.toBeUndefined()
  })
})
