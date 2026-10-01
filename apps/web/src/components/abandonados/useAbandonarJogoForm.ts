import { useEffect, useState, type FormEvent } from 'react'
import { abandonarJogoSchema, type AbandonarJogoFormData } from './abandonarJogo.schema'
import { ABANDONADOS_CAMPO_ERRO_MENSAGENS, ABANDONADOS_ERRO_GENERICO } from './abandonados.constants'
import { hojeIso } from './abandonados.utils'
import { AbandonadosApiError } from '@/lib/services/abandonadosService'
import type { JogoAbandonado } from '@/types/abandonados'

export interface UseAbandonarJogoFormProps {
  initialData?: Partial<JogoAbandonado> | null
  onSubmit: (data: AbandonarJogoFormData) => Promise<void>
}

export function useAbandonarJogoForm({ initialData, onSubmit }: UseAbandonarJogoFormProps) {
  const [nome, setNome] = useState(initialData?.nome ?? '')
  const [consoleName, setConsoleName] = useState(initialData?.console ?? '')
  const [abandonadoEm, setAbandonadoEm] = useState(initialData?.abandonado_em?.slice(0, 10) || hojeIso())
  const [iniciadoEm, setIniciadoEm] = useState(initialData?.iniciado_em?.slice(0, 10) || '')
  const [horas, setHoras] = useState('')
  const [minutos, setMinutos] = useState('')
  const [segundos, setSegundos] = useState('')
  const [motivo, setMotivo] = useState(initialData?.motivo ?? '')
  const [igdbId, setIgdbId] = useState<number | null>(initialData?.igdb_id ?? null)
  const [igdbCapaUrl, setIgdbCapaUrl] = useState(initialData?.igdb_capa_url ?? '')
  const [plataformas, setPlataformas] = useState<string[]>([])
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    const s = initialData?.tempo_jogado ?? 0
    setNome(initialData?.nome ?? ''); setConsoleName(initialData?.console ?? '')
    setAbandonadoEm(initialData?.abandonado_em?.slice(0, 10) || hojeIso())
    setIniciadoEm(initialData?.iniciado_em?.slice(0, 10) || '')
    setHoras(s >= 3600 ? String(Math.floor(s / 3600)) : '')
    setMinutos((s % 3600) >= 60 ? String(Math.floor((s % 3600) / 60)) : '')
    setSegundos((s % 60) > 0 ? String(s % 60) : '')
    setMotivo(initialData?.motivo ?? ''); setIgdbId(initialData?.igdb_id ?? null)
    setIgdbCapaUrl(initialData?.igdb_capa_url ?? ''); setErrors({})
  }, [initialData])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setErrors({})
    const res = abandonarJogoSchema.safeParse({
      igdb_id: igdbId, nome, console: consoleName, abandonado_em: abandonadoEm,
      tempo_jogado_horas: horas === '' ? 0 : parseInt(horas, 10),
      tempo_jogado_minutos: minutos === '' ? 0 : parseInt(minutos, 10),
      tempo_jogado_segundos: segundos === '' ? 0 : parseInt(segundos, 10),
      motivo: motivo || null, igdb_capa_url: igdbCapaUrl,
      iniciado_em: iniciadoEm || undefined,
    })
    if (!res.success) {
      const errMap: Record<string, string> = {}
      for (const i of res.error.issues) {
        const field = String(i.path[0])
        if (!errMap[field]) errMap[field] = i.message
      }
      setErrors(errMap); return
    }
    setIsSubmitting(true)
    try {
      await onSubmit(res.data)
    } catch (err) {
      if (err instanceof AbandonadosApiError) {
        if (err.status === 401) { window.location.assign('/login'); return }
        const campo = ABANDONADOS_CAMPO_ERRO_MENSAGENS[err.codigo]
        if (campo) setErrors({ [campo.campo]: campo.mensagem })
        else setErrors({ form: err.message || ABANDONADOS_ERRO_GENERICO })
      } else if (err instanceof Error) {
        setErrors({ form: err.message || ABANDONADOS_ERRO_GENERICO })
      } else {
        setErrors({ form: ABANDONADOS_ERRO_GENERICO })
      }
    } finally { setIsSubmitting(false) }
  }

  return {
    nome, setNome, consoleName, setConsoleName, abandonadoEm, setAbandonadoEm,
    horas, setHoras, minutos, setMinutos, segundos, setSegundos,
    motivo, setMotivo, igdbId, setIgdbId, igdbCapaUrl, setIgdbCapaUrl,
    iniciadoEm, setIniciadoEm,
    plataformas, setPlataformas, errors, isSubmitting, handleSubmit,
  }
}
