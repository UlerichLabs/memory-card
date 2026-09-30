import { useContext, useEffect, useRef } from 'react'
import { Button } from '@/components/ui/button'
import { AuthContext } from '@/store/authStore'
import { formatarCapaIGDB } from '@/lib/utils'
import { jogosService, type IGDBJogoSugestao } from '@/lib/services/jogosService'
import { CONSOLES_PADRAO } from '@/components/jogos/GameForm/GameForm.constants'
import { AbandonarJogoCampos } from './AbandonarJogoCampos'
import { useAbandonarJogoForm } from './useAbandonarJogoForm'
import type { OrigemFilaInfo } from '@/stores/abandonadosStore'
import type { JogoAbandonado } from '@/types/abandonados'
import type { AbandonarJogoFormData } from './abandonarJogo.schema'

export interface AbandonarJogoFormProps {
  initialData?: Partial<JogoAbandonado> | null
  origemFila?: OrigemFilaInfo
  isEditing?: boolean
  onSubmit: (data: AbandonarJogoFormData) => Promise<void>
  onCancel: () => void
}

export function AbandonarJogoForm({
  initialData,
  origemFila,
  isEditing = false,
  onSubmit,
  onCancel,
}: AbandonarJogoFormProps) {
  const auth = useContext(AuthContext)
  const token = auth?.sessao?.access_token

  const {
    nome, setNome, consoleName, setConsoleName, abandonadoEm, setAbandonadoEm,
    horas, setHoras, minutos, setMinutos, segundos, setSegundos,
    motivo, setMotivo, igdbId, setIgdbId, setIgdbCapaUrl,
    plataformas, setPlataformas,
    errors, isSubmitting, handleSubmit,
  } = useAbandonarJogoForm({ initialData, onSubmit })

  const selecaoAtual = useRef(0)

  useEffect(() => {
    if (!initialData?.igdb_id) return
    jogosService.obterDetalhesIGDB(initialData.igdb_id, token).then((detalhes) => {
      if (detalhes.platforms?.length) setPlataformas(detalhes.platforms.map((p) => p.name))
    }).catch(() => undefined)
  }, [initialData?.igdb_id, token, setPlataformas])

  function handleNomeChange(value: string) {
    selecaoAtual.current += 1
    if (igdbId !== null) {
      setIgdbId(null); setIgdbCapaUrl(''); setPlataformas([]); setConsoleName('')
    }
    setNome(value)
  }

  async function handleSelect(sugestao: IGDBJogoSugestao) {
    const selecao = ++selecaoAtual.current
    setNome(sugestao.name); setIgdbId(sugestao.id); setConsoleName('')
    setIgdbCapaUrl(formatarCapaIGDB(sugestao.cover?.url))
    try {
      const detalhes = await jogosService.obterDetalhesIGDB(sugestao.id, token)
      if (selecao !== selecaoAtual.current) return
      const nomes = detalhes.platforms?.map((p) => p.name) || []
      setPlataformas(nomes)
      if (nomes.length === 1) setConsoleName(nomes[0].slice(0, 100))
    } catch {
      setPlataformas([])
    }
  }

  const consoleOptions = plataformas.length
    ? plataformas.map((item) => ({ value: item, label: item }))
    : CONSOLES_PADRAO

  return (
    <form noValidate onSubmit={handleSubmit} className="flex min-h-0 flex-1 flex-col gap-5 overflow-hidden">
      <div className="custom-scrollbar min-h-0 flex-1 overflow-y-auto overflow-x-hidden pr-1">
        <AbandonarJogoCampos
          origemFila={origemFila}
          errors={errors}
          nome={nome}
          onChangeNome={handleNomeChange}
          onSelectSugestao={handleSelect}
          consoleName={consoleName}
          setConsoleName={setConsoleName}
          igdbId={igdbId}
          consoleOptions={consoleOptions}
          abandonadoEm={abandonadoEm}
          setAbandonadoEm={setAbandonadoEm}
          horas={horas}
          minutos={minutos}
          segundos={segundos}
          setHoras={setHoras}
          setMinutos={setMinutos}
          setSegundos={setSegundos}
          motivo={motivo}
          setMotivo={setMotivo}
        />
      </div>
      <div className="flex shrink-0 justify-end gap-3 border-t border-[var(--border)] pt-4">
        <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting}>
          Cancelar
        </Button>
        <Button
          type="submit"
          disabled={isSubmitting}
          className="bg-[var(--accent)] font-bold text-[var(--accent-foreground)]"
        >
          {isSubmitting ? 'Salvando...' : isEditing ? 'Salvar alterações' : 'Salvar abandono'}
        </Button>
      </div>
    </form>
  )
}
