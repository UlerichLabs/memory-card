import { act, renderHook, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { createElement, type FormEvent, type ReactNode } from 'react'
import { JogandoContext, type JogandoStore } from '@/stores/jogandoStore'
import type { JogoEmAndamento } from '@/types/jogando'
import { useGameForm } from './useGameForm'

const capaThumb = 'https://images.igdb.com/t_thumb/co1abc.jpg'
const capaGrande = 'https://images.igdb.com/t_cover_big/co1abc.jpg'

function eventoSubmit(): FormEvent {
  return { preventDefault: vi.fn() } as unknown as FormEvent
}

function criarJogandoWrapper(jogos: JogoEmAndamento[]) {
  const store: JogandoStore = {
    jogos,
    isLoading: false,
    carregado: true,
    error: null,
    aviso: null,
    isModalOpen: false,
    carregar: vi.fn().mockResolvedValue(undefined),
    criar: vi.fn(),
    remover: vi.fn(),
    abrirModalIniciar: vi.fn(),
    fecharModal: vi.fn(),
    definirAviso: vi.fn(),
    limparAviso: vi.fn(),
  }
  return ({ children }: { children: ReactNode }) =>
    createElement(JogandoContext.Provider, { value: store }, children)
}

describe('useGameForm', () => {
  it('normaliza a capa inicial e mantém a URL normalizada no payload', async () => {
    const onSubmit = vi.fn().mockResolvedValue(undefined)
    const initialData = { nome: 'Chrono Trigger', console: 'SNES', finalizado_em: '2026-03-13', nota: 10, dificuldade: 'A' as const, igdb_capa_url: capaThumb }
    const { result } = renderHook(() => useGameForm({ initialData, onSubmit }))

    expect(result.current.igdbCapaUrl).toBe(capaGrande)

    await act(() => result.current.handleSubmit(eventoSubmit()))

    expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({ igdb_capa_url: capaGrande }))
  })

  it.each([undefined, ''])('mantém capa vazia quando initialData.igdb_capa_url é %s', (igdb_capa_url) => {
    const { result } = renderHook(() => useGameForm({ initialData: { igdb_capa_url }, onSubmit: vi.fn() }))

    expect(result.current.igdbCapaUrl).toBe('')
  })

  it('renormaliza a capa quando initialData muda', () => {
    const { result, rerender } = renderHook(({ initialData }) => useGameForm({ initialData, onSubmit: vi.fn() }), {
      initialProps: { initialData: { igdb_capa_url: capaThumb } },
    })

    rerender({ initialData: { igdb_capa_url: '' } })

    expect(result.current.igdbCapaUrl).toBe('')
  })

  it('selecionar sugestão correspondente preenche iniciado_em', () => {
    const jogo: JogoEmAndamento = { id: 1, nome: 'Chrono Trigger', igdb_id: 42, igdb_capa_url: null, iniciado_em: '2026-09-01T00:00:00Z' }
    const wrapper = criarJogandoWrapper([jogo])
    const { result } = renderHook(() => useGameForm({ onSubmit: vi.fn() }), { wrapper })

    act(() => {
      result.current.setNome('Chrono Trigger')
      result.current.setIgdbId(42)
    })

    expect(result.current.iniciadoEm).toBe('2026-09-01')
  })

  it('digitar nome correspondente preenche iniciado_em com debounce', async () => {
    const jogo: JogoEmAndamento = { id: 2, nome: 'Chrono Trigger', igdb_id: null, igdb_capa_url: null, iniciado_em: '2026-09-01T00:00:00Z' }
    const wrapper = criarJogandoWrapper([jogo])
    const { result } = renderHook(() => useGameForm({ onSubmit: vi.fn() }), { wrapper })

    act(() => {
      result.current.setNome('  chrono trigger  ')
    })

    await waitFor(() => expect(result.current.iniciadoEm).toBe('2026-09-01'))
  })

  it('iniciado_em já preenchido não é sobrescrito', () => {
    const jogo: JogoEmAndamento = { id: 3, nome: 'Chrono Trigger', igdb_id: 42, igdb_capa_url: null, iniciado_em: '2026-09-01T00:00:00Z' }
    const wrapper = criarJogandoWrapper([jogo])
    const { result } = renderHook(() => useGameForm({ onSubmit: vi.fn() }), { wrapper })

    act(() => {
      result.current.setIniciadoEm('2026-10-01')
      result.current.setNome('Chrono Trigger')
      result.current.setIgdbId(42)
    })

    expect(result.current.iniciadoEm).toBe('2026-10-01')
  })

  it('usuário apagar a data: não preenche de novo até mudar o jogo', async () => {
    const jogo1: JogoEmAndamento = { id: 1, nome: 'Chrono Trigger', igdb_id: 42, igdb_capa_url: null, iniciado_em: '2026-09-01T00:00:00Z' }
    const jogo2: JogoEmAndamento = { id: 2, nome: 'Super Mario 64', igdb_id: 99, igdb_capa_url: null, iniciado_em: '2026-08-15T00:00:00Z' }
    const wrapper = criarJogandoWrapper([jogo1, jogo2])
    const { result } = renderHook(() => useGameForm({ onSubmit: vi.fn() }), { wrapper })

    act(() => {
      result.current.setNome('Chrono Trigger')
      result.current.setIgdbId(42)
    })
    expect(result.current.iniciadoEm).toBe('2026-09-01')

    act(() => {
      result.current.setIniciadoEm('')
    })
    expect(result.current.iniciadoEm).toBe('')

    act(() => {
      result.current.setNome('Chrono Trigger ')
    })
    await new Promise((resolve) => setTimeout(resolve, 200))
    expect(result.current.iniciadoEm).toBe('')

    act(() => {
      result.current.setNome('Super Mario 64')
      result.current.setIgdbId(99)
    })
    expect(result.current.iniciadoEm).toBe('2026-08-15')
  })

  it('sem correspondência: iniciado_em continua vazio', () => {
    const jogo: JogoEmAndamento = { id: 1, nome: 'Chrono Trigger', igdb_id: 42, igdb_capa_url: null, iniciado_em: '2026-09-01T00:00:00Z' }
    const wrapper = criarJogandoWrapper([jogo])
    const { result } = renderHook(() => useGameForm({ onSubmit: vi.fn() }), { wrapper })

    act(() => {
      result.current.setNome('Jogo Qualquer')
      result.current.setIgdbId(999)
    })

    expect(result.current.iniciadoEm).toBe('')
  })

  it('sem JogandoProvider continua funcionando normalmente', () => {
    const { result } = renderHook(() => useGameForm({ onSubmit: vi.fn() }))

    act(() => {
      result.current.setNome('Chrono Trigger')
      result.current.setIgdbId(42)
    })

    expect(result.current.iniciadoEm).toBe('')
  })
})
