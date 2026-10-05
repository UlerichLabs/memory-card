import { useContext, useEffect, useRef, useState, type FormEvent } from 'react'
import { gameFormSchema } from './GameForm.schema'
import { JOGOS_CAMPO_ERRO_MENSAGENS, JOGOS_ERRO_GENERICO, JogosApiError, type Dificuldade, type JogoZeradoDTO, type SalvarJogoPayload } from '@/lib/services/jogosService'
import { formatarCapaIGDB } from '@/lib/utils'; import { JogandoContext } from '@/stores/jogandoStore'; import { encontrarJogoEmAndamento } from '@/lib/jogandoUtils'

export interface UseGameFormProps {
  initialData?: Partial<JogoZeradoDTO>
  onSubmit: (payload: SalvarJogoPayload) => Promise<void>
  isEditing?: boolean
}

export function useGameForm(props: UseGameFormProps) {
  const { initialData, onSubmit } = props
  const isEditing = props.isEditing ?? Boolean(initialData && 'id' in initialData && initialData.id)
  const jogando = useContext(JogandoContext), jogos = jogando?.jogos
  const ultimoMatchIdRef = useRef<number | null>(null), s = initialData?.tempo_jogado ?? 0
  const [nome, setNome] = useState(initialData?.nome ?? ''), [consoleName, setConsoleName] = useState(initialData?.console ?? '')
  const [genero, setGenero] = useState(initialData?.genero ?? ''), [tipo, setTipo] = useState(initialData?.tipo ?? '')
  const [iniciadoEm, setIniciadoEm] = useState(initialData?.iniciado_em?.slice(0, 10) ?? ''), [finalizadoEm, setFinalizadoEm] = useState(initialData?.finalizado_em?.slice(0, 10) ?? '')
  const [horas, setHoras] = useState(s >= 3600 ? String(Math.floor(s / 3600)) : ''), [minutos, setMinutos] = useState((s % 3600) >= 60 ? String(Math.floor((s % 3600) / 60)) : ''), [segundos, setSegundos] = useState((s % 60) > 0 ? String(s % 60) : '')
  const [nota, setNota] = useState<number | undefined>(initialData?.nota), [dificuldade, setDificuldade] = useState<Dificuldade | undefined>(initialData?.dificuldade)
  const [review, setReview] = useState(initialData?.review ?? ''), [destaque, setDestaque] = useState(initialData?.destaque ?? false)
  const [igdbId, setIgdbId] = useState<number | null>(initialData?.igdb_id ?? null), [igdbCapaUrl, setIgdbCapaUrl] = useState(formatarCapaIGDB(initialData?.igdb_capa_url ?? undefined))
  const [igdbDescricao, setIgdbDescricao] = useState(initialData?.igdb_descricao ?? ''), [plataformas, setPlataformas] = useState<string[]>([])
  const [errors, setErrors] = useState<Record<string, string>>({}), [destaqueError, setDestaqueError] = useState<string | null>(null), [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => { setIgdbCapaUrl(formatarCapaIGDB(initialData?.igdb_capa_url ?? undefined)) }, [initialData])

  useEffect(() => {
    if (isEditing) return
    const preencher = () => {
      const match = encontrarJogoEmAndamento(jogos ?? [], { nome, igdbId })
      if (!match) { ultimoMatchIdRef.current = null; return }
      if (ultimoMatchIdRef.current !== match.id) {
        ultimoMatchIdRef.current = match.id
        setIniciadoEm((atual) => (!atual ? match.iniciado_em.slice(0, 10) : atual))
      }
    }
    if (igdbId !== null) { preencher(); return }
    const timer = setTimeout(preencher, 150)
    return () => clearTimeout(timer)
  }, [nome, igdbId, jogos, isEditing])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault(); setErrors({}); setDestaqueError(null)
    const res = gameFormSchema.safeParse({
      igdb_id: igdbId, nome, console: consoleName, genero, tipo, iniciado_em: iniciadoEm || undefined, finalizado_em: finalizadoEm,
      tempo_jogado_horas: horas === '' ? 0 : parseInt(horas, 10), tempo_jogado_minutos: minutos === '' ? 0 : parseInt(minutos, 10), tempo_jogado_segundos: segundos === '' ? 0 : parseInt(segundos, 10),
      nota, dificuldade, review: review || null, destaque, igdb_capa_url: igdbCapaUrl, igdb_descricao: igdbDescricao,
    })
    if (!res.success) {
      const errMap: Record<string, string> = {}
      for (const i of res.error.issues) { const f = String(i.path[0]); if (!errMap[f]) errMap[f] = i.message }
      setErrors(errMap); return
    }
    setIsSubmitting(true)
    try {
      await onSubmit(res.data)
    } catch (err) {
      if (err instanceof JogosApiError && err.status === 401) { window.location.assign('/login'); return }
      if (err instanceof JogosApiError && (err.status === 409 || err.codigo === 'jogos.destaque_ano_conflito')) {
        setDestaque(false); setDestaqueError(err.message || 'já existe um destaque para este ano')
      } else {
        const campo = err instanceof JogosApiError ? JOGOS_CAMPO_ERRO_MENSAGENS[err.codigo] : undefined
        if (campo) setErrors({ [campo.campo]: campo.mensagem })
        else setErrors({ form: JOGOS_ERRO_GENERICO })
      }
    } finally { setIsSubmitting(false) }
  }

  return {
    nome, setNome, consoleName, setConsoleName, genero, setGenero, tipo, setTipo, iniciadoEm, setIniciadoEm,
    finalizadoEm, setFinalizadoEm, horas, setHoras, minutos, setMinutos, segundos, setSegundos,
    nota, setNota, dificuldade, setDificuldade, review, setReview, destaque, setDestaque,
    igdbId, setIgdbId, igdbCapaUrl, setIgdbCapaUrl, igdbDescricao, setIgdbDescricao, plataformas, setPlataformas,
    errors, destaqueError, isSubmitting, handleSubmit,
  }
}
