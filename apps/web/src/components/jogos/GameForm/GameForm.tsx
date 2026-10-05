import { useContext, useEffect, useRef } from 'react'
import { Crown, Gamepad2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { CustomSelect } from '@/components/ui/CustomSelect'
import { AuthContext } from '@/store/authStore'
import { formatarCapaIGDB } from '@/lib/utils'
import { jogosService, JogosApiError, type Dificuldade, type IGDBJogoSugestao, type JogoZeradoDTO, type SalvarJogoPayload } from '@/lib/services/jogosService'
import { GameFormAutocomplete } from './GameFormAutocomplete'
import { useGameForm } from './useGameForm'
import { CONSOLES_PADRAO } from './GameForm.constants'
import { DateInput } from './DateInput'
import { TimeInput } from './TimeInput'
import { RatingPicker } from './RatingPicker'
import { DifficultyPicker } from './DifficultyPicker'

export interface GameFormProps {
  initialData?: Partial<JogoZeradoDTO>
  onSubmit: (payload: SalvarJogoPayload) => Promise<void>
  onCancel?: () => void
  isEditing?: boolean
  aviso?: string
  textoSubmit?: string
}

function limitarLista(itens: string[], limite: number): string {
  return itens.reduce((resultado, item) => {
    const proximo = resultado ? `${resultado}, ${item}` : item
    return proximo.length <= limite ? proximo : resultado
  }, '')
}

export function GameForm({
  initialData,
  onSubmit,
  onCancel,
  isEditing = false,
  aviso,
  textoSubmit,
}: GameFormProps) {
  const auth = useContext(AuthContext)
  const token = auth?.sessao?.access_token
  const form = useGameForm({ initialData, onSubmit, isEditing })
  const detalheCarregado = useRef<number | null>(null)
  const selecaoAtual = useRef(0)
  const { nome, setNome, consoleName, setConsoleName, genero, setGenero, tipo, setTipo, iniciadoEm, setIniciadoEm, finalizadoEm, setFinalizadoEm, horas, setHoras, minutos, setMinutos, segundos, setSegundos, nota, setNota, dificuldade, setDificuldade, review, setReview, destaque, setDestaque, igdbId, setIgdbId, igdbCapaUrl, setIgdbCapaUrl, igdbDescricao, setIgdbDescricao, plataformas, setPlataformas, errors, destaqueError, isSubmitting, handleSubmit } = form

  useEffect(() => {
    if (!initialData?.igdb_id || detalheCarregado.current === initialData.igdb_id) return
    detalheCarregado.current = initialData.igdb_id
    jogosService.obterDetalhesIGDB(initialData.igdb_id, token).then((detalhes) => {
      if (detalhes.platforms?.length) setPlataformas(detalhes.platforms.map((item) => item.name))
      if (!genero && detalhes.genres?.length) setGenero(limitarLista(detalhes.genres.map((item) => item.name), 150))
      if (!igdbDescricao && detalhes.summary) setIgdbDescricao(detalhes.summary)
    }).catch(() => undefined)
  }, [initialData?.igdb_id, token, genero, igdbDescricao, setGenero, setPlataformas, setIgdbDescricao])

  function handleNomeChange(value: string) {
    selecaoAtual.current += 1
    if (igdbId !== null) { setIgdbId(null); setIgdbCapaUrl(''); setIgdbDescricao(''); setPlataformas([]); setConsoleName('') }
    setNome(value)
  }

  async function handleSelect(sugestao: IGDBJogoSugestao) {
    const selecao = ++selecaoAtual.current
    setNome(sugestao.name); setIgdbId(sugestao.id); setIgdbCapaUrl(formatarCapaIGDB(sugestao.cover?.url)); setIgdbDescricao(sugestao.summary || ''); setConsoleName('')
    if (!tipo) setTipo('Campanha')
    try {
      const detalhes = await jogosService.obterDetalhesIGDB(sugestao.id, token)
      if (selecao !== selecaoAtual.current) return
      if (detalhes.genres?.length) setGenero(limitarLista(detalhes.genres.map((item) => item.name), 150))
      if (detalhes.summary) setIgdbDescricao(detalhes.summary)
      const nomes = detalhes.platforms?.map((item) => item.name) || []
      setPlataformas(nomes)
      if (nomes.length === 1) setConsoleName(limitarLista(nomes, 100))
    } catch (error: unknown) { if (error instanceof JogosApiError && error.status === 401) window.location.assign('/login'); setPlataformas([]) }
  }

  const consoleOptions = plataformas.length ? plataformas.map((item) => ({ value: item, label: item })) : CONSOLES_PADRAO
  return <form noValidate onSubmit={handleSubmit} className="flex min-h-0 flex-1 flex-col gap-5 overflow-hidden">
    <div className="custom-scrollbar min-h-0 flex-1 overflow-y-auto overflow-x-hidden">
      <div className="flex flex-col gap-5">
        {aviso && (
          <div
            className={
              'rounded-lg border border-[var(--abandonado-banner-border)] ' +
              'bg-[var(--abandonado-banner-bg)] px-4 py-3 text-sm text-[var(--abandonado-banner-text)]'
            }
          >
            {aviso}
          </div>
        )}
        {errors.form && <div className="text-sm text-[var(--danger)]">{errors.form}</div>}
        <GameFormAutocomplete nome={nome} onChangeNome={handleNomeChange} onSelectSugestao={handleSelect} error={errors.nome} />
        <div className="grid min-w-0 gap-8 md:grid-cols-[240px_minmax(0,1fr)]">
          <div className="flex min-w-0 flex-col gap-4">
            <div className="relative mx-auto w-full max-w-[240px]">
              {igdbCapaUrl ? <img src={igdbCapaUrl} alt="Capa do jogo" className="aspect-[3/4] w-full rounded-md border border-[var(--border)] object-cover" /> : <div className="flex aspect-[3/4] items-center justify-center rounded-md border-2 border-dashed border-[var(--border-subtle)] bg-[linear-gradient(150deg,var(--bg-surface-alt),var(--bg-primary))] text-[var(--text-faint)]"><Gamepad2 className="size-12 opacity-40" /></div>}
              <button type="button" aria-label="Marcar como jogo do ano" aria-pressed={destaque} onClick={() => setDestaque(!destaque)} className={`absolute right-2 top-2 flex size-11 items-center justify-center rounded-full border focus-visible:ring-2 focus-visible:ring-[var(--accent)] ${destaque ? 'border-[var(--highlight-gold)] bg-[var(--highlight-gold)] text-[var(--bg-primary)]' : 'border-[var(--highlight-gold)] bg-[var(--bg-primary)]/80 text-[var(--highlight-gold)]'}`}><Crown className="size-5" fill={destaque ? 'currentColor' : 'none'} /></button>
              {destaque && <span className="absolute bottom-2 left-2 rounded-full bg-[var(--highlight-gold)] px-2 py-1 text-[10px] font-bold text-[var(--bg-primary)]">Jogo do ano</span>}
            </div>
            {destaqueError && <span className="text-xs text-[var(--danger)]">{destaqueError}</span>}
            <div className="flex flex-col gap-1.5"><Label htmlFor="console" className="text-sm text-[var(--text-secondary)]">Plataforma *</Label>{igdbId ? <CustomSelect id="console" name="console" value={consoleName} onChange={(value) => setConsoleName(value.slice(0, 100))} options={consoleOptions} placeholder="Selecione a plataforma" error={!!errors.console} ariaLabel="Plataforma" /> : <Input id="console" name="console" maxLength={100} value={consoleName} onChange={(event) => setConsoleName(event.target.value)} placeholder="Ex: PC, PlayStation" aria-invalid={!!errors.console} className="bg-[var(--bg-surface-alt)] text-[var(--text-primary)]" />}{errors.console && <span className="text-xs text-[var(--danger)]">{errors.console}</span>}</div>
            <div><Label htmlFor="genero" className="text-sm text-[var(--text-secondary)]">Gênero</Label><Input id="genero" maxLength={150} value={genero} onChange={(event) => setGenero(event.target.value)} aria-invalid={!!errors.genero} className="mt-1 bg-[var(--bg-surface-alt)] text-[var(--text-primary)]" />{errors.genero && <span className="text-xs text-[var(--danger)]">{errors.genero}</span>}</div>
            <div><Label htmlFor="tipo" className="text-sm text-[var(--text-secondary)]">Tipo</Label><Input id="tipo" maxLength={50} value={tipo} onChange={(event) => setTipo(event.target.value)} aria-invalid={!!errors.tipo} className="mt-1 bg-[var(--bg-surface-alt)] text-[var(--text-primary)]" />{errors.tipo && <span className="text-xs text-[var(--danger)]">{errors.tipo}</span>}</div>
          </div>
          <div className="flex min-w-0 flex-col gap-5">
            {igdbDescricao && <p id="igdb-descricao" className="line-clamp-3 text-sm leading-5 text-[var(--text-secondary)]">{igdbDescricao}</p>}
            <div className="grid min-w-0 gap-4 md:grid-cols-3"><DateInput key={iniciadoEm ? 'com-data' : 'sem-data'} id="iniciado_em" label="Iniciado em" value={iniciadoEm} onChange={setIniciadoEm} error={errors.iniciado_em} /><DateInput id="finalizado_em" label="Finalizado em *" value={finalizadoEm} onChange={setFinalizadoEm} error={errors.finalizado_em} /><div className="min-w-0"><Label className="text-sm text-[var(--text-secondary)]">Tempo jogado</Label><div className="mt-1"><TimeInput horas={horas} minutos={minutos} segundos={segundos} setHoras={setHoras} setMinutos={setMinutos} setSegundos={setSegundos} /></div>{(errors.tempo_jogado_minutos || errors.tempo_jogado_segundos) && <span className="text-xs text-[var(--danger)]">{errors.tempo_jogado_minutos || errors.tempo_jogado_segundos}</span>}</div></div>
            <DifficultyPicker value={dificuldade} onChange={(value) => setDificuldade(value as Dificuldade)} error={errors.dificuldade} />
            <RatingPicker value={nota} onChange={setNota} error={errors.nota} />
            <div><div className="flex items-center justify-between"><Label htmlFor="review" className="text-sm text-[var(--text-secondary)]">Review</Label><span className="text-xs text-[var(--text-muted)]">{review.length}/5000</span></div><textarea id="review" name="review" maxLength={5000} rows={4} value={review} onChange={(event) => setReview(event.target.value)} aria-invalid={!!errors.review} placeholder="O que achou do jogo? Final, dificuldade, como zerou…" className="mt-1 w-full rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)]" />{errors.review && <span className="text-xs text-[var(--danger)]">{errors.review}</span>}</div>
          </div>
        </div>
      </div>
    </div>
    <div className="flex shrink-0 justify-end gap-3 border-t border-[var(--border)] pt-4">
      {onCancel && <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting}>Cancelar</Button>}
      <Button type="submit" disabled={isSubmitting} className="btn-primario">
        {isSubmitting ? 'Salvando...' : isEditing ? 'Atualizar registro' : (textoSubmit ?? 'Salvar registro')}
      </Button>
    </div>
  </form>
}
