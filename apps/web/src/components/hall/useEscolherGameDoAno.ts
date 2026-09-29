import { useEffect, useRef, useState, type KeyboardEvent } from 'react'
import { useHallDaFamaStore } from '@/stores/hallDaFamaStore'
import { JogosApiError } from '@/lib/services/jogosService'
import type { JogoZeradoDTO } from '@/types/jogos'

export interface UseEscolherGameDoAnoProps {
  open: boolean; ano: number; jogoAtual: JogoZeradoDTO | null; onSuccess?: () => void; onClose: () => void
}

export function useEscolherGameDoAno({ open, ano, jogoAtual, onSuccess, onClose }: UseEscolherGameDoAnoProps) {
  const { buscarJogosDoAno, definirGameDoAno, removerGameDoAno } = useHallDaFamaStore()
  const [jogos, setJogos] = useState<JogoZeradoDTO[]>([])
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const itemRefs = useRef<Map<number, HTMLDivElement>>(new Map())

  useEffect(() => {
    if (!open) return
    setSelectedId(jogoAtual ? jogoAtual.id : null)
    setErrorMessage(null)
    setIsLoading(true)
    const ctrl = new AbortController()
    buscarJogosDoAno(ano, ctrl.signal)
      .then((res) => { if (!ctrl.signal.aborted) setJogos(res) })
      .catch((err) => { if (!ctrl.signal.aborted) setErrorMessage(err instanceof Error ? err.message : 'Erro ao carregar jogos') })
      .finally(() => { if (!ctrl.signal.aborted) setIsLoading(false) })
    return () => ctrl.abort()
  }, [open, ano, jogoAtual, buscarJogosDoAno])

  const handleKeyDown = (e: KeyboardEvent<HTMLDivElement>, index: number) => {
    let nextIndex = -1
    if (e.key === 'ArrowDown' || e.key === 'ArrowRight') nextIndex = (index + 1) % jogos.length
    else if (e.key === 'ArrowUp' || e.key === 'ArrowLeft') nextIndex = (index - 1 + jogos.length) % jogos.length
    else if (e.key === ' ' || e.key === 'Enter') { setSelectedId(jogos[index].id); return }
    if (nextIndex >= 0) {
      e.preventDefault(); const nextJogo = jogos[nextIndex]
      setSelectedId(nextJogo.id); itemRefs.current.get(nextJogo.id)?.focus()
    }
  }

  const handleConfirmar = async () => {
    if (!selectedId || selectedId === jogoAtual?.id) return
    setIsSubmitting(true); setErrorMessage(null)
    try {
      await definirGameDoAno(selectedId); onSuccess?.(); onClose()
    } catch (err) {
      setErrorMessage(err instanceof JogosApiError && err.status === 409
        ? 'Este ano já possui outro Game do Ano definido.'
        : 'Não foi possível definir o Game do Ano. Tente novamente.')
    } finally { setIsSubmitting(false) }
  }

  const handleRemover = async () => {
    if (!jogoAtual) return
    setIsSubmitting(true); setErrorMessage(null)
    try {
      await removerGameDoAno(jogoAtual.id); onSuccess?.(); onClose()
    } catch {
      setErrorMessage('Não foi possível remover o Game do Ano. Tente novamente.')
    } finally { setIsSubmitting(false) }
  }

  return {
    jogos, selectedId, setSelectedId, isLoading, isSubmitting, errorMessage, itemRefs,
    handleKeyDown, handleConfirmar, handleRemover, modoTrocar: !!jogoAtual,
    desabilitarConfirmar: !selectedId || selectedId === jogoAtual?.id || isSubmitting,
    mostrarAvisoTroca: !!jogoAtual && selectedId !== null && selectedId !== jogoAtual.id,
  }
}
