import { afterEach, describe, expect, it, vi } from 'vitest'
import { act, renderHook } from '@testing-library/react'
import { jogosService, JogosApiError, type JogoZeradoDTO, type IGDBJogoSugestao } from '@/lib/services/jogosService'
import { JogosProvider, useJogosStore } from './jogosStore'

const jogoMock: JogoZeradoDTO = {
  id: 1,
  usuario_id: 10,
  nome: 'Chrono Trigger',
  console: 'SNES',
  genero: 'RPG',
  tipo: 'Campanha',
  iniciado_em: '2026-01-01',
  finalizado_em: '2026-01-15',
  tempo_jogado: 72000,
  nota: 10,
  dificuldade: 'A',
  review: '100% dos finais',
  destaque: false,
}

describe('jogosStore', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('inicia com estado padrão', () => {
    const { result } = renderHook(() => useJogosStore(), { wrapper: JogosProvider })
    expect(result.current.jogos).toEqual([])
    expect(result.current.meta).toEqual({ pagina: 1, por_pagina: 24, total: 0, total_paginas: 0 })
    expect(result.current.filtros).toEqual({ consoles: [], generos: [], tipos: [], anos: [] })
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it('guarda a lista vinda da api e os metadados de paginação', async () => {
    const mockResposta = {
      data: [jogoMock],
      meta: { pagina: 2, por_pagina: 24, total: 25, total_paginas: 2 },
    }
    vi.spyOn(jogosService, 'listar').mockResolvedValue(mockResposta)

    const { result } = renderHook(() => useJogosStore(), { wrapper: JogosProvider })

    await act(async () => {
      await result.current.carregarJogos({ pagina: 2 })
    })

    expect(result.current.jogos).toEqual([jogoMock])
    expect(result.current.meta).toEqual(mockResposta.meta)
    expect(result.current.isLoading).toBe(false)
  })

  it('limpa a lista e reseta meta ao chamar limparBiblioteca', () => {
    const { result } = renderHook(() => useJogosStore(), {
      wrapper: ({ children }) => JogosProvider({ children, initialJogos: [jogoMock] }),
    })

    expect(result.current.jogos).toHaveLength(1)
    act(() => result.current.limparBiblioteca())
    expect(result.current.jogos).toEqual([])
    expect(result.current.meta.total).toBe(0)
  })

  it('limpa o registro em edição ao fechar o modal', () => {
    const { result } = renderHook(() => useJogosStore(), {
      wrapper: ({ children }) => JogosProvider({ children, initialJogos: [jogoMock] }),
    })
    act(() => result.current.abrirModalEdicao(jogoMock))
    expect(result.current.jogoEmEdicao).toEqual(jogoMock)
    act(() => result.current.fecharModal())
    expect(result.current.isModalOpen).toBe(false)
    expect(result.current.jogoEmEdicao).toBeNull()
  })

  it('lança erro ao ser usado fora de JogosProvider', () => {
    expect(() => renderHook(() => useJogosStore())).toThrow(
      'useJogosStore deve ser utilizado dentro de um JogosProvider'
    )
  })

  it('criarJogo chama api e recarrega listagem e filtros', async () => {
    vi.spyOn(jogosService, 'criar').mockResolvedValue(jogoMock)
    const spyListar = vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [jogoMock],
      meta: { pagina: 1, por_pagina: 24, total: 1, total_paginas: 1 },
    })
    const spyFiltros = vi.spyOn(jogosService, 'obterFiltros').mockResolvedValue({
      consoles: ['SNES'],
      generos: ['RPG'],
      tipos: ['Campanha'],
      anos: [2026],
    })

    const { result } = renderHook(() => useJogosStore(), { wrapper: JogosProvider })

    await act(async () => {
      const criado = await result.current.criarJogo({
        nome: 'Chrono Trigger',
        console: 'SNES',
        finalizado_em: '2026-01-15',
        nota: 10,
        dificuldade: 'A',
        destaque: false,
      })
      expect(criado).toEqual(jogoMock)
    })

    expect(spyListar).toHaveBeenCalled()
    expect(spyFiltros).toHaveBeenCalled()
    expect(result.current.jogos).toEqual([jogoMock])
  })

  it('criarJogo atualiza error e lança exceção em caso de erro', async () => {
    vi.spyOn(jogosService, 'criar').mockRejectedValue(new Error('Erro de conexão'))
    const { result } = renderHook(() => useJogosStore(), { wrapper: JogosProvider })

    await act(async () => {
      await expect(
        result.current.criarJogo({
          nome: 'Chrono Trigger',
          console: 'SNES',
          finalizado_em: '2026-01-15',
          nota: 10,
          dificuldade: 'A',
          destaque: false,
        })
      ).rejects.toThrow('Erro de conexão')
    })

    expect(result.current.jogos).toEqual([])
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBe('Erro de conexão')
  })

  it('atualizarJogo altera registro e recarrega listagem', async () => {
    const atualizado: JogoZeradoDTO = { ...jogoMock, id: 1, nome: 'Versao Nova' }
    vi.spyOn(jogosService, 'atualizar').mockResolvedValue(atualizado)
    const spyListar = vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [atualizado],
      meta: { pagina: 1, por_pagina: 24, total: 1, total_paginas: 1 },
    })
    vi.spyOn(jogosService, 'obterFiltros').mockResolvedValue({ consoles: [], generos: [], tipos: [], anos: [] })

    const { result } = renderHook(() => useJogosStore(), {
      wrapper: ({ children }) => JogosProvider({ children, initialJogos: [jogoMock] }),
    })

    await act(async () => {
      const res = await result.current.atualizarJogo(1, {
        nome: 'Versao Nova',
        console: 'SNES',
        finalizado_em: '2026-01-15',
        nota: 10,
        dificuldade: 'A',
        destaque: false,
      })
      expect(res).toEqual(atualizado)
    })

    expect(spyListar).toHaveBeenCalled()
    expect(result.current.jogos[0].nome).toBe('Versao Nova')
  })

  it('atualizarJogo trata erro 409 de conflito', async () => {
    const erro409 = new JogosApiError('jogos.destaque_ano_conflito', 'já existe um destaque para este ano', 409)
    vi.spyOn(jogosService, 'atualizar').mockRejectedValue(erro409)

    const { result } = renderHook(() => useJogosStore(), {
      wrapper: ({ children }) => JogosProvider({ children, initialJogos: [jogoMock] }),
    })

    await act(async () => {
      await expect(
        result.current.atualizarJogo(1, {
          nome: 'Chrono Trigger',
          console: 'SNES',
          finalizado_em: '2026-01-15',
          nota: 10,
          dificuldade: 'A',
          destaque: true,
        })
      ).rejects.toMatchObject({ status: 409, codigo: 'jogos.destaque_ano_conflito' })
    })

    expect(result.current.error).toBe('já existe um destaque para este ano')
    expect(result.current.isLoading).toBe(false)
  })

  it('excluirJogo chama api e recarrega listagem e filtros', async () => {
    vi.spyOn(jogosService, 'excluir').mockResolvedValue(undefined)
    const spyListar = vi.spyOn(jogosService, 'listar').mockResolvedValue({
      data: [],
      meta: { pagina: 1, por_pagina: 24, total: 0, total_paginas: 0 },
    })
    const spyFiltros = vi.spyOn(jogosService, 'obterFiltros').mockResolvedValue({ consoles: [], generos: [], tipos: [], anos: [] })

    const { result } = renderHook(() => useJogosStore(), {
      wrapper: ({ children }) => JogosProvider({ children, initialJogos: [jogoMock] }),
    })

    await act(async () => {
      await result.current.excluirJogo(1)
    })

    expect(spyListar).toHaveBeenCalled()
    expect(spyFiltros).toHaveBeenCalled()
    expect(result.current.jogos).toEqual([])
  })

  it('buscarIGDB chama o serviço e retorna sugestões', async () => {
    const sugestoes: IGDBJogoSugestao[] = [
      { id: 123, name: 'Chrono Trigger', cover: { url: '//images.igdb.com/123.jpg' } },
    ]
    vi.spyOn(jogosService, 'buscarIGDB').mockResolvedValue(sugestoes)

    const { result } = renderHook(() => useJogosStore(), { wrapper: JogosProvider })

    let resultado: IGDBJogoSugestao[] = []
    await act(async () => {
      resultado = await result.current.buscarIGDB('Chrono')
    })

    expect(resultado).toEqual(sugestoes)
  })

  it('armazena valoresIniciais e onSalvo em modalRegistroOpcoes e limpa ao fecharModal', () => {
    const onSalvo = vi.fn()
    const { result } = renderHook(() => useJogosStore(), { wrapper: JogosProvider })

    act(() => {
      result.current.abrirModalRegistro({
        valoresIniciais: { nome: 'Super Mario 64', igdb_id: 1070 },
        onSalvo,
      })
    })

    expect(result.current.isModalOpen).toBe(true)
    expect(result.current.modalRegistroOpcoes?.valoresIniciais).toEqual({ nome: 'Super Mario 64', igdb_id: 1070 })
    expect(result.current.modalRegistroOpcoes?.onSalvo).toBe(onSalvo)

    act(() => {
      result.current.fecharModal()
    })

    expect(result.current.isModalOpen).toBe(false)
    expect(result.current.modalRegistroOpcoes).toBeNull()
  })
})
