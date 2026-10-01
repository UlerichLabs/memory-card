import { useCallback, useContext } from 'react'
import { JogosContext } from '@/stores/jogosStore'
import { useAbandonadosStore } from '@/stores/abandonadosStore'
import { formatarDataBrasileira } from './abandonados.utils'
import { AVISO_FALHA_REMOCAO_RETOMAR } from './abandonados.constants'
import type { JogoAbandonado } from '@/types/abandonados'

export function useRetomarAbandonado() {
  const jogos = useContext(JogosContext)
  const { excluirJogo, definirAviso } = useAbandonadosStore()

  const retomarJogo = useCallback((jogo: JogoAbandonado) => {
    if (!jogos) return
    const dataPt = formatarDataBrasileira(jogo.abandonado_em)
    const aviso = `Retomando um jogo abandonado em ${dataPt}. Ao salvar o zeramento, ele sai da lista de Abandonados.`

    jogos.abrirModalRegistro({
      valoresIniciais: {
        nome: jogo.nome,
        console: jogo.console,
        igdb_id: jogo.igdb_id ?? undefined,
        igdb_capa_url: jogo.igdb_capa_url ?? undefined,
        tempo_jogado: jogo.tempo_jogado,
      },
      aviso,
      textoSubmit: 'Salvar zeramento',
      onSalvo: async () => {
        try {
          await excluirJogo(jogo.id)
        } catch {
          definirAviso(AVISO_FALHA_REMOCAO_RETOMAR)
        }
      },
    })
  }, [jogos, excluirJogo, definirAviso])

  return { retomarJogo }
}
