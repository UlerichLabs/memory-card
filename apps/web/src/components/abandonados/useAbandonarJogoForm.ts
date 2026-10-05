import { useContext, useEffect, useRef, useState, type FormEvent } from 'react'
import { abandonarJogoSchema, type AbandonarJogoFormData } from './abandonarJogo.schema'
import { ABANDONADOS_CAMPO_ERRO_MENSAGENS, ABANDONADOS_ERRO_GENERICO } from './abandonados.constants'; import { hojeIso } from './abandonados.utils'
import { AbandonadosApiError } from '@/lib/services/abandonadosService'; import type { JogoAbandonado } from '@/types/abandonados'
import { formatarCapaIGDB } from '@/lib/utils'; import { JogandoContext } from '@/stores/jogandoStore'; import { encontrarJogoEmAndamento } from '@/lib/jogandoUtils'

export interface UseAbandonarJogoFormProps {
  initialData?: Partial<JogoAbandonado> | null
  onSubmit: (data: AbandonarJogoFormData) => Promise<void>
  isEditing?: boolean
}

export function useAbandonarJogoForm(props: UseAbandonarJogoFormProps) {
  const { initialData, onSubmit } = props
  const isEditing = props.isEditing ?? Boolean(initialData && 'id' in initialData && initialData.id)
  const jogando = useContext(JogandoContext), jogos = jogando?.jogos
  const ultimoMatchIdRef = useRef<number | null>(null), [nome, setNome] = useState(initialData?.nome ?? ''), [consoleName, setConsoleName] = useState(initialData?.console ?? '')
  const [abandonadoEm, setAbandonadoEm] = useState(initialData?.abandonado_em?.slice(0, 10) || hojeIso()), [iniciadoEm, setIniciadoEm] = useState(initialData?.iniciado_em?.slice(0, 10) || '')
  const [horas, setHoras] = useState(''), [minutos, setMinutos] = useState(''), [segundos, setSegundos] = useState('')
  const [motivo, setMotivo] = useState(initialData?.motivo ?? ''), [igdbId, setIgdbId] = useState<number | null>(initialData?.igdb_id ?? null)
  const [igdbCapaUrl, setIgdbCapaUrl] = useState(formatarCapaIGDB(initialData?.igdb_capa_url ?? undefined))
  const [plataformas, setPlataformas] = useState<string[]>([]), [errors, setErrors] = useState<Record<string, string>>({}), [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    const s = initialData?.tempo_jogado ?? 0
    setNome(initialData?.nome ?? ''); setConsoleName(initialData?.console ?? '')
    setAbandonadoEm(initialData?.abandonado_em?.slice(0, 10) || hojeIso()); setIniciadoEm(initialData?.iniciado_em?.slice(0, 10) || '')
    setHoras(s >= 3600 ? String(Math.floor(s / 3600)) : ''); setMinutos((s % 3600) >= 60 ? String(Math.floor((s % 3600) / 60)) : ''); setSegundos((s % 60) > 0 ? String(s % 60) : '')
    setMotivo(initialData?.motivo ?? ''); setIgdbId(initialData?.igdb_id ?? null); setIgdbCapaUrl(formatarCapaIGDB(initialData?.igdb_capa_url ?? undefined))
    setErrors({}); ultimoMatchIdRef.current = null
  }, [initialData])

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
    e.preventDefault(); setErrors({})
    const res = abandonarJogoSchema.safeParse({
      igdb_id: igdbId, nome, console: consoleName, abandonado_em: abandonadoEm,
      tempo_jogado_horas: horas === '' ? 0 : parseInt(horas, 10), tempo_jogado_minutos: minutos === '' ? 0 : parseInt(minutos, 10), tempo_jogado_segundos: segundos === '' ? 0 : parseInt(segundos, 10),
      motivo: motivo || null, igdb_capa_url: igdbCapaUrl, iniciado_em: iniciadoEm || undefined,
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
      if (err instanceof AbandonadosApiError) {
        if (err.status === 401) { window.location.assign('/login'); return }
        const campo = ABANDONADOS_CAMPO_ERRO_MENSAGENS[err.codigo]
        setErrors(campo ? { [campo.campo]: campo.mensagem } : { form: err.message || ABANDONADOS_ERRO_GENERICO })
      } else {
        setErrors({ form: err instanceof Error ? err.message : ABANDONADOS_ERRO_GENERICO })
      }
    } finally { setIsSubmitting(false) }
  }

  return {
    nome, setNome, consoleName, setConsoleName, abandonadoEm, setAbandonadoEm, horas, setHoras,
    minutos, setMinutos, segundos, setSegundos, motivo, setMotivo, igdbId, setIgdbId, igdbCapaUrl, setIgdbCapaUrl,
    iniciadoEm, setIniciadoEm, plataformas, setPlataformas, errors, isSubmitting, handleSubmit,
  }
}
