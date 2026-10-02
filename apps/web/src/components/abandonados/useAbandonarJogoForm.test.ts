import { act, renderHook } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import type { FormEvent } from 'react'
import { useAbandonarJogoForm } from './useAbandonarJogoForm'

const capaThumb = 'https://images.igdb.com/t_thumb/co1abc.jpg'
const capaGrande = 'https://images.igdb.com/t_cover_big/co1abc.jpg'

function eventoSubmit(): FormEvent {
  return { preventDefault: vi.fn() } as unknown as FormEvent
}

describe('useAbandonarJogoForm', () => {
  it('normaliza a capa inicial e mantém a URL normalizada no payload', async () => {
    const onSubmit = vi.fn().mockResolvedValue(undefined)
    const initialData = { nome: 'Chrono Trigger', console: 'SNES', abandonado_em: '2026-03-13', igdb_capa_url: capaThumb }
    const { result } = renderHook(() => useAbandonarJogoForm({ initialData, onSubmit }))

    expect(result.current.igdbCapaUrl).toBe(capaGrande)

    await act(() => result.current.handleSubmit(eventoSubmit()))

    expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({ igdb_capa_url: capaGrande }))
  })

  it.each([undefined, ''])('mantém capa vazia quando initialData.igdb_capa_url é %s', (igdb_capa_url) => {
    const initialData = { igdb_capa_url }
    const { result } = renderHook(() => useAbandonarJogoForm({ initialData, onSubmit: vi.fn() }))

    expect(result.current.igdbCapaUrl).toBe('')
  })

  it('renormaliza a capa quando initialData muda', () => {
    const { result, rerender } = renderHook(({ initialData }) => useAbandonarJogoForm({ initialData, onSubmit: vi.fn() }), {
      initialProps: { initialData: { igdb_capa_url: capaThumb } },
    })

    rerender({ initialData: { igdb_capa_url: '' } })

    expect(result.current.igdbCapaUrl).toBe('')
  })
})
