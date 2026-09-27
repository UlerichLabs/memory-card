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
}

export function GameForm({ initialData, onSubmit, onCancel, isEditing = false }: GameFormProps) {
  const auth = useContext(AuthContext)
  const token = auth?.sessao?.access_token
  const form = useGameForm({ initialData, onSubmit })
  const detalheCarregado = useRef<number | null>(null)
  const selecaoAtual = useRef(0)
  const { nome, setNome, consoleName, setConsoleName, genero, setGenero, tipo, setTipo, iniciadoEm, setIniciadoEm, finalizadoEm, setFinalizadoEm, horas, setHoras, minutos, setMinutos, segundos, setSegundos, nota, setNota, dificuldade, setDificuldade, condicao, setCondicao, destaque, setDestaque, igdbId, setIgdbId, igdbCapaUrl, setIgdbCapaUrl, igdbDescricao, setIgdbDescricao, plataformas, setPlataformas, errors, destaqueError, isSubmitting, handleSubmit } = form

  useEffect(() => {
    if (!initialData?.igdb_id || detalheCarregado.current === initialData.igdb_id) return
    detalheCarregado.current = initialData.igdb_id
    jogosService.obterDetalhesIGDB(initialData.igdb_id, token).then((detalhes) => {
      if (detalhes.platforms?.length) setPlataformas(detalhes.platforms.map((item) => item.name))
      if (!genero && detalhes.genres?.length) setGenero(detalhes.genres.map((item) => item.name).join(', '))
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
      if (detalhes.genres?.length) setGenero(detalhes.genres.map((item) => item.name).join(', '))
      if (detalhes.summary) setIgdbDescricao(detalhes.summary)
      const nomes = detalhes.platforms?.map((item) => item.name) || []
      setPlataformas(nomes)
      if (nomes.length === 1) setConsoleName(nomes[0])
    } catch (error: unknown) { if (error instanceof JogosApiError && error.status === 401) window.location.assign('/login'); setPlataformas([]) }
  }

  const consoleOptions = plataformas.length ? plataformas.map((item) => ({ value: item, label: item })) : CONSOLES_PADRAO
  return <form noValidate onSubmit={handleSubmit} className="flex flex-col gap-5">
    {errors.form && <div className="text-sm text-[var(--danger)]">{errors.form}</div>}
    <GameFormAutocomplete nome={nome} onChangeNome={handleNomeChange} onSelectSugestao={handleSelect} error={errors.nome} />
    <div className="grid gap-6 md:grid-cols-[240px_minmax(0,1fr)]">
      <div className="flex flex-col gap-3">
        <div className="relative mx-auto w-full max-w-[240px]">
          {igdbCapaUrl ? <img src={igdbCapaUrl} alt="Capa do jogo" className="aspect-[3/4] w-full rounded-md border border-[var(--border)] object-cover" /> : <div className="flex aspect-[3/4] items-center justify-center rounded-md border-2 border-dashed border-[var(--border-subtle)] bg-[linear-gradient(150deg,var(--bg-surface-alt),var(--bg-primary))] text-[var(--text-faint)]"><Gamepad2 className="size-12 opacity-40" /></div>}
          <button type="button" aria-label="Marcar como jogo do ano" aria-pressed={destaque} onClick={() => setDestaque(!destaque)} className={`absolute right-2 top-2 flex size-11 items-center justify-center rounded-full border focus-visible:ring-2 focus-visible:ring-[var(--accent)] ${destaque ? 'border-[var(--highlight-gold)] bg-[var(--highlight-gold)] text-[var(--bg-primary)]' : 'border-[var(--highlight-gold)] bg-[var(--bg-primary)]/80 text-[var(--highlight-gold)]'}`}><Crown className="size-5" fill={destaque ? 'currentColor' : 'none'} /></button>
          {destaque && <span className="absolute bottom-2 left-2 rounded-full bg-[var(--highlight-gold)] px-2 py-1 text-[10px] font-bold text-[var(--bg-primary)]">Jogo do ano</span>}
        </div>
        {destaqueError && <span className="text-xs text-[var(--danger)]">{destaqueError}</span>}
        <div className="flex flex-col gap-1.5"><Label htmlFor="console" className="text-sm text-[var(--text-secondary)]">Plataforma *</Label>{igdbId ? <CustomSelect id="console" name="console" value={consoleName} onChange={setConsoleName} options={consoleOptions} placeholder="Selecione a plataforma" error={!!errors.console} ariaLabel="Plataforma" /> : <Input id="console" name="console" value={consoleName} onChange={(event) => setConsoleName(event.target.value)} placeholder="Ex: PC, PlayStation" aria-invalid={!!errors.console} className="bg-[var(--bg-surface-alt)] text-[var(--text-primary)]" />}{errors.console && <span className="text-xs text-[var(--danger)]">{errors.console}</span>}</div>
        <div><Label htmlFor="genero" className="text-sm text-[var(--text-secondary)]">Gênero</Label><Input id="genero" value={genero} onChange={(event) => setGenero(event.target.value)} className="mt-1 bg-[var(--bg-surface-alt)] text-[var(--text-primary)]" /></div>
        <div><Label htmlFor="tipo" className="text-sm text-[var(--text-secondary)]">Tipo</Label><Input id="tipo" value={tipo} onChange={(event) => setTipo(event.target.value)} className="mt-1 bg-[var(--bg-surface-alt)] text-[var(--text-primary)]" /></div>
      </div>
      <div className="flex min-w-0 flex-col gap-4">
        <div><Label htmlFor="igdb-descricao" className="text-sm text-[var(--text-secondary)]">Resumo</Label><p id="igdb-descricao" className="mt-1 line-clamp-3 min-h-14 text-sm leading-5 text-[var(--text-secondary)]">{igdbDescricao || 'Nenhum resumo disponível.'}</p></div>
        <div className="grid gap-3 md:grid-cols-3"><DateInput id="iniciado_em" label="Iniciado em" value={iniciadoEm} onChange={setIniciadoEm} error={errors.iniciado_em} /><DateInput id="finalizado_em" label="Finalizado em *" value={finalizadoEm} onChange={setFinalizadoEm} error={errors.finalizado_em} /><div><Label className="text-sm text-[var(--text-secondary)]">Tempo jogado</Label><div className="mt-1"><TimeInput horas={horas} minutos={minutos} segundos={segundos} setHoras={setHoras} setMinutos={setMinutos} setSegundos={setSegundos} /></div>{(errors.tempo_jogado_minutos || errors.tempo_jogado_segundos) && <span className="text-xs text-[var(--danger)]">{errors.tempo_jogado_minutos || errors.tempo_jogado_segundos}</span>}</div></div>
        <DifficultyPicker value={dificuldade} onChange={(value) => setDificuldade(value as Dificuldade)} error={errors.dificuldade} />
        <RatingPicker value={nota} onChange={setNota} error={errors.nota} />
        <div><div className="flex items-center justify-between"><Label htmlFor="condicao_zeramento" className="text-sm text-[var(--text-secondary)]">Review</Label><span className="text-xs text-[var(--text-muted)]">{condicao.length}/500</span></div><textarea id="condicao_zeramento" name="condicao_zeramento" maxLength={500} rows={4} value={condicao} onChange={(event) => setCondicao(event.target.value)} placeholder="O que achou do jogo? Final, dificuldade, como zerou…" className="mt-1 w-full rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)]" />{errors.condicao_zeramento && <span className="text-xs text-[var(--danger)]">{errors.condicao_zeramento}</span>}</div>
      </div>
    </div>
    <div className="flex justify-end gap-3 border-t border-[var(--border)] pt-4">{onCancel && <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting}>Cancelar</Button>}<Button type="submit" disabled={isSubmitting} className="bg-[var(--accent)] font-bold text-[var(--accent-foreground)]">{isSubmitting ? 'Salvando...' : isEditing ? 'Atualizar registro' : 'Salvar registro'}</Button></div>
  </form>
}
