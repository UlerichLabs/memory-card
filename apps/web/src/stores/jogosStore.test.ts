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
  condicao_zeramento: '100% dos finais',
  destaque: false,
}

describe('jogosStore', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('inicia com estado padrão', () => {
    const { result } = renderHook(() => useJogosStore(), { wrapper: JogosProvider })
    expect(result.current.jogos).toEqual([])
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it('lança erro ao ser usado fora de JogosProvider', () => {
    expect(() => renderHook(() => useJogosStore())).toThrow(
      'useJogosStore deve ser utilizado dentro de um JogosProvider'
    )
  })

  it('criarJogo adiciona o jogo criado à lista com sucesso', async () => {
    vi.spyOn(jogosService, 'criar').mockResolvedValue(jogoMock)
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

    expect(result.current.jogos).toContainEqual(jogoMock)
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBeNull()
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

  it('atualizarJogo altera o item correspondente na lista', async () => {
    const inicial: JogoZeradoDTO = { ...jogoMock, id: 1, nome: 'Versao Antiga' }
    const atualizado: JogoZeradoDTO = { ...jogoMock, id: 1, nome: 'Versao Nova' }
    vi.spyOn(jogosService, 'atualizar').mockResolvedValue(atualizado)

    const { result } = renderHook(() => useJogosStore(), {
      wrapper: ({ children }) => JogosProvider({ children, initialJogos: [inicial] }),
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

    expect(result.current.jogos[0].nome).toBe('Versao Nova')
    expect(result.current.isLoading).toBe(false)
    expect(result.current.error).toBeNull()
  })

  it('atualizarJogo trata erro 409 de conflito', async () => {
    const inicial: JogoZeradoDTO = { ...jogoMock, id: 1 }
    const erro409 = new JogosApiError('jogos.destaque_ano_conflito', 'já existe um destaque para este ano', 409)
    vi.spyOn(jogosService, 'atualizar').mockRejectedValue(erro409)

    const { result } = renderHook(() => useJogosStore(), {
      wrapper: ({ children }) => JogosProvider({ children, initialJogos: [inicial] }),
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

  it('excluirJogo remove o item correspondente da lista', async () => {
    const jogo2: JogoZeradoDTO = { ...jogoMock, id: 2, nome: 'Super Mario World' }
    vi.spyOn(jogosService, 'excluir').mockResolvedValue(undefined)

    const { result } = renderHook(() => useJogosStore(), {
      wrapper: ({ children }) => JogosProvider({ children, initialJogos: [jogoMock, jogo2] }),
    })

    await act(async () => {
      await result.current.excluirJogo(1)
    })

    expect(result.current.jogos).toHaveLength(1)
    expect(result.current.jogos[0].id).toBe(2)
    expect(result.current.isLoading).toBe(false)
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
})
