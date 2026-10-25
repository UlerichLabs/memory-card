import { useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { jogosService, JogosApiError } from '@/lib/services/jogosService'
import { useJogandoStore } from '@/stores/jogandoStore'
import { encontrarJogoEmAndamento, hojeIso } from '@/lib/jogandoUtils'
import { formatarCapaIGDB } from '@/lib/utils'
import { ApiError } from '@/lib/api'
import { obterMensagemErroJogando } from '@/lib/jogandoErros'
import type { IGDBJogoSugestao } from '@/types/jogos'

export function useJogoCatalogo(igdbId: string | undefined) {
  const navigate = useNavigate()
  const { jogos, criar } = useJogandoStore()
  const numId = Number(igdbId)
  const idInvalido = !igdbId || Number.isNaN(numId) || numId <= 0

  const [jogo, setJogo] = useState<IGDBJogoSugestao | null>(null)
  const [carregando, setCarregando] = useState(!idInvalido)
  const [erroStatus, setErroStatus] = useState<'notFound' | 'network' | null>(idInvalido ? 'notFound' : null)
  const [isIniciando, setIsIniciando] = useState(false)
  const [sucesso, setSucesso] = useState(false)
  const [erroAcao, setErroAcao] = useState<string | null>(null)
  const [trigger, setTrigger] = useState(0)

  useEffect(() => {
    if (idInvalido) return
    const ctrl = new AbortController()
    jogosService.obterDetalhesIGDB(numId).then((dados) => {
      if (!ctrl.signal.aborted) {
        setJogo(dados)
        setCarregando(false)
      }
    }).catch((err) => {
      if (!ctrl.signal.aborted) {
        const nf = err instanceof JogosApiError &&
          (err.status === 404 || err.status === 400 || err.codigo === 'jogos.not_found' || err.codigo === 'jogos.invalid_id')
        setErroStatus(nf ? 'notFound' : 'network')
        setCarregando(false)
      }
    })
    return () => ctrl.abort()
  }, [idInvalido, numId, trigger])

  const recarregar = useCallback(() => {
    setCarregando(true)
    setErroStatus(null)
    setTrigger((v) => v + 1)
  }, [])

  const emAndamento = sucesso || Boolean(jogo && encontrarJogoEmAndamento(jogos, { igdbId: jogo.id, nome: jogo.name }))

  const jogarAgora = useCallback(async () => {
    if (!jogo || isIniciando || emAndamento) return
    setIsIniciando(true)
    setErroAcao(null)
    try {
      await criar({
        nome: jogo.name,
        igdb_id: jogo.id,
        igdb_capa_url: formatarCapaIGDB(jogo.cover?.url) || null,
        iniciado_em: hojeIso(),
      })
      setSucesso(true)
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        navigate('/login')
        return
      }
      setErroAcao(obterMensagemErroJogando(err))
    } finally {
      setIsIniciando(false)
    }
  }, [jogo, isIniciando, emAndamento, criar, navigate])

  return {
    jogo, carregando, erroStatus, recarregar, jogarAgora, isIniciando, emAndamento, sucesso, erroAcao,
  }
}
