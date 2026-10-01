import { useCallback, useState, type FormEvent } from 'react'
import { ApiError } from '@/lib/api'
import { hojeIso } from '@/lib/jogandoUtils'
import { iniciarJogoSchema, type IniciarJogoFormData } from './iniciarJogo.schema'

const ERROS: Record<string, { campo: string; mensagem: string }> = {
  'jogando.nome_obrigatorio': { campo: 'nome', mensagem: 'O nome do jogo é obrigatório.' },
  'jogando.nome_muito_longo': { campo: 'nome', mensagem: 'O nome deve ter no máximo 200 caracteres.' },
  'jogando.iniciado_em_obrigatorio': { campo: 'iniciado_em', mensagem: 'A data de início é obrigatória.' },
  'jogando.iniciado_em_invalido': { campo: 'iniciado_em', mensagem: 'Data de início inválida.' },
  'jogando.iniciado_em_futuro': { campo: 'iniciado_em', mensagem: 'Data de início não pode ser futura.' },
}

export function useIniciarJogoForm(onSubmit: (data: IniciarJogoFormData) => Promise<void>) {
  const [nome, setNome] = useState('')
  const [iniciadoEm, setIniciadoEm] = useState(hojeIso())
  const [igdbId, setIgdbId] = useState<number | null>(null)
  const [igdbCapaUrl, setIgdbCapaUrl] = useState<string | null>(null)
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)
  const reset = useCallback(() => {
    setNome('')
    setIniciadoEm(hojeIso())
    setIgdbId(null)
    setIgdbCapaUrl(null)
    setErrors({})
    setIsSubmitting(false)
  }, [])

  function escolherSugestao(id: number, capa: string | null, nomeEscolhido: string) {
    setIgdbId(id)
    setIgdbCapaUrl(capa)
    setNome(nomeEscolhido)
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setErrors({})
    const resultado = iniciarJogoSchema.safeParse({ nome, iniciado_em: iniciadoEm, igdb_id: igdbId, igdb_capa_url: igdbCapaUrl })
    if (!resultado.success) {
      const proximos: Record<string, string> = {}
      resultado.error.issues.forEach((issue) => { const campo = String(issue.path[0]); if (!proximos[campo]) proximos[campo] = issue.message })
      setErrors(proximos)
      return
    }
    setIsSubmitting(true)
    try {
      await onSubmit(resultado.data)
    } catch (error) {
      if (error instanceof ApiError) {
        if (error.status === 401) { window.location.assign('/login'); return }
        const mapeado = ERROS[error.codigo]
        setErrors(mapeado ? { [mapeado.campo]: mapeado.mensagem } : { form: 'Não foi possível iniciar o jogo. Tente novamente.' })
      } else {
        setErrors({ form: 'Não foi possível iniciar o jogo. Tente novamente.' })
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  return { nome, setNome, iniciadoEm, setIniciadoEm, igdbId, igdbCapaUrl, escolherSugestao, errors, isSubmitting, handleSubmit, reset }
}
