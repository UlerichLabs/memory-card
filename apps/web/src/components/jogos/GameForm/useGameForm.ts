import { useState, type FormEvent } from 'react'
import { gameFormSchema } from './GameForm.schema'
import { JogosApiError, type Dificuldade, type JogoZeradoDTO, type SalvarJogoPayload } from '@/lib/services/jogosService'

export interface UseGameFormProps {
  initialData?: Partial<JogoZeradoDTO>
  onSubmit: (payload: SalvarJogoPayload) => Promise<void>
}

export function useGameForm({ initialData, onSubmit }: UseGameFormProps) {
  const s = initialData?.tempo_jogado ?? 0
  const [nome, setNome] = useState(initialData?.nome ?? '')
  const [consoleName, setConsoleName] = useState(initialData?.console ?? '')
  const [genero, setGenero] = useState(initialData?.genero ?? '')
  const [tipo, setTipo] = useState(initialData?.tipo ?? '')
  const [iniciadoEm, setIniciadoEm] = useState(initialData?.iniciado_em?.slice(0, 10) ?? '')
  const [finalizadoEm, setFinalizadoEm] = useState(initialData?.finalizado_em?.slice(0, 10) ?? '')
  const [horas, setHoras] = useState(s >= 3600 ? String(Math.floor(s / 3600)) : '')
  const [minutos, setMinutos] = useState((s % 3600) >= 60 ? String(Math.floor((s % 3600) / 60)) : '')
  const [segundos, setSegundos] = useState((s % 60) > 0 ? String(s % 60) : '')
  const [nota, setNota] = useState<number>(initialData?.nota ?? 10)
  const [dificuldade, setDificuldade] = useState<Dificuldade>(initialData?.dificuldade ?? 'A')
  const [condicao, setCondicao] = useState(initialData?.condicao_zeramento ?? '')
  const [destaque, setDestaque] = useState(initialData?.destaque ?? false)
  const [igdbId, setIgdbId] = useState<number | null>(initialData?.igdb_id ?? null)
  const [igdbCapaUrl, setIgdbCapaUrl] = useState(initialData?.igdb_capa_url ?? '')
  const [igdbDescricao, setIgdbDescricao] = useState(initialData?.igdb_descricao ?? '')
  const [plataformas, setPlataformas] = useState<string[]>([])
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [destaqueError, setDestaqueError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setErrors({}); setDestaqueError(null)
    const res = gameFormSchema.safeParse({
      igdb_id: igdbId, nome, console: consoleName, genero, tipo,
      iniciado_em: iniciadoEm || undefined, finalizado_em: finalizadoEm,
      tempo_jogado_horas: horas === '' ? 0 : parseInt(horas, 10),
      tempo_jogado_minutos: minutos === '' ? 0 : parseInt(minutos, 10),
      tempo_jogado_segundos: segundos === '' ? 0 : parseInt(segundos, 10),
      nota, dificuldade, condicao_zeramento: condicao, destaque,
      igdb_capa_url: igdbCapaUrl, igdb_descricao: igdbDescricao,
    })
    if (!res.success) {
      const errMap: Record<string, string> = {}
      for (const i of res.error.issues) {
        const field = String(i.path[0])
        if (!errMap[field]) errMap[field] = i.message
      }
      setErrors(errMap)
      return
    }
    setIsSubmitting(true)
    try {
      await onSubmit(res.data)
    } catch (err) {
      if (err instanceof JogosApiError && (err.status === 409 || err.codigo === 'jogos.destaque_ano_conflito')) {
        setDestaque(false)
        setDestaqueError(err.message || 'já existe um destaque para este ano')
      } else {
        setErrors({ form: err instanceof Error ? err.message : 'Erro ao salvar jogo' })
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  return {
    nome, setNome, consoleName, setConsoleName, genero, setGenero, tipo, setTipo,
    iniciadoEm, setIniciadoEm, finalizadoEm, setFinalizadoEm, horas, setHoras,
    minutos, setMinutos, segundos, setSegundos, nota, setNota, dificuldade,
    setDificuldade, condicao, setCondicao, destaque, setDestaque, igdbId, setIgdbId,
    igdbCapaUrl, setIgdbCapaUrl, igdbDescricao, setIgdbDescricao, plataformas, setPlataformas,
    errors, destaqueError, isSubmitting, handleSubmit,
  }
}
