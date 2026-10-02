import { act, renderHook } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import type { FormEvent } from 'react'
import { useGameForm } from './useGameForm'

const capaThumb = 'https://images.igdb.com/t_thumb/co1abc.jpg'
const capaGrande = 'https://images.igdb.com/t_cover_big/co1abc.jpg'

function eventoSubmit(): FormEvent {
  return { preventDefault: vi.fn() } as unknown as FormEvent
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
})
