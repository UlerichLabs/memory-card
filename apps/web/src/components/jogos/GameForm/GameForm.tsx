import { useContext } from 'react'
import { jogosService, type Dificuldade, type JogoZeradoDTO, type SalvarJogoPayload } from '@/lib/services/jogosService'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { AuthContext } from '@/store/authStore'
import { formatarCapaIGDB } from '@/lib/utils'
import { GameFormAutocomplete } from './GameFormAutocomplete'
import { useGameForm } from './useGameForm'

export interface GameFormProps {
  initialData?: Partial<JogoZeradoDTO>
  onSubmit: (payload: SalvarJogoPayload) => Promise<void>
  onCancel?: () => void
  isEditing?: boolean
}

const DIFICULDADES: Dificuldade[] = ['C', 'B', 'A', 'AA', 'AAA']

export function GameForm({ initialData, onSubmit, onCancel, isEditing = false }: GameFormProps) {
  const auth = useContext(AuthContext)
  const token = auth?.sessao?.access_token
  const {
    nome, setNome, consoleName, setConsoleName, genero, setGenero, tipo, setTipo,
    iniciadoEm, setIniciadoEm, finalizadoEm, setFinalizadoEm, horas, setHoras,
    minutos, setMinutos, segundos, setSegundos, nota, setNota, dificuldade, setDificuldade,
    condicao, setCondicao, destaque, setDestaque, setIgdbId, igdbCapaUrl, setIgdbCapaUrl,
    igdbDescricao, setIgdbDescricao, plataformas, setPlataformas, errors, destaqueError,
    isSubmitting, handleSubmit,
  } = useGameForm({ initialData, onSubmit })

  return (
    <form noValidate onSubmit={handleSubmit} className="flex flex-col gap-5 bg-[var(--bg-surface)] p-6 rounded-xl border border-[var(--border)]">
      {errors.form && <div className="text-sm text-[var(--danger)]">{errors.form}</div>}
      <GameFormAutocomplete
        nome={nome}
        onChangeNome={setNome}
        onSelectSugestao={async (sugestao) => {
          setNome(sugestao.name); setIgdbId(sugestao.id); setIgdbCapaUrl(formatarCapaIGDB(sugestao.cover?.url))
          if (sugestao.summary) setIgdbDescricao(sugestao.summary)
          if (!tipo) setTipo('Campanha')
          try {
            const detalhes = await jogosService.obterDetalhesIGDB(sugestao.id, token)
            if (detalhes?.genres?.length) setGenero(detalhes.genres.map((g) => g.name).join(', '))
            if (detalhes?.summary) setIgdbDescricao(detalhes.summary)
            if (detalhes?.platforms?.length) {
              const nomes = detalhes.platforms.map((p) => p.name)
              setPlataformas(nomes)
              if (!consoleName) setConsoleName(nomes[0])
            }
          } catch {}
        }}
        igdbCapaUrl={igdbCapaUrl}
        error={errors.nome}
      />
      {igdbCapaUrl && (
        <div className="flex items-center gap-3.5 rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] p-3">
          <img src={igdbCapaUrl} alt="Prévia da capa" className="h-20 w-15 shrink-0 rounded border border-[var(--border)] object-cover shadow-sm" />
          <div className="min-w-0 flex-1 space-y-0.5">
            <span className="text-[10px] font-bold uppercase tracking-wider text-[var(--accent)]">Identificado no IGDB</span>
            <p className="truncate text-sm font-bold text-[var(--text-primary)]">{nome}</p>
            {igdbDescricao && <p className="line-clamp-2 text-xs text-[var(--text-muted)]">{igdbDescricao}</p>}
          </div>
        </div>
      )}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="console" className="text-[var(--text-secondary)] text-sm">Console *</Label>
          <Input id="console" name="console" list="consoles-lista" value={consoleName} onChange={(e) => setConsoleName(e.target.value)} placeholder="Ex: Super Nintendo, PS5" className="bg-[var(--bg-surface-alt)] border-[var(--border-subtle)] text-[var(--text-primary)]" aria-invalid={!!errors.console} />
          <datalist id="consoles-lista">{plataformas.map((p) => <option key={p} value={p} />)}</datalist>
          {plataformas.length > 0 && (
            <div className="flex flex-wrap gap-1 pt-1">
              {plataformas.slice(0, 4).map((p) => (
                <button key={p} type="button" onClick={() => setConsoleName(p)} className={`rounded-full border px-2 py-0.5 text-[10px] font-medium transition-colors ${consoleName === p ? 'border-[var(--accent)] bg-[var(--accent)] text-[#0E0F12]' : 'border-[var(--border-subtle)] bg-[var(--bg-surface)] text-[var(--text-secondary)] hover:text-[var(--text-primary)]'}`}>{p}</button>
              ))}
            </div>
          )}
          {errors.console && <span className="text-xs text-[var(--danger)]">{errors.console}</span>}
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="genero" className="text-[var(--text-secondary)] text-sm">Gênero</Label>
          <Input id="genero" name="genero" value={genero} onChange={(e) => setGenero(e.target.value)} placeholder="Ex: JRPG, Ação" className="bg-[var(--bg-surface-alt)] border-[var(--border-subtle)] text-[var(--text-primary)]" />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="tipo" className="text-[var(--text-secondary)] text-sm">Tipo</Label>
          <Input id="tipo" name="tipo" value={tipo} onChange={(e) => setTipo(e.target.value)} placeholder="Ex: Campanha, DLC" className="bg-[var(--bg-surface-alt)] border-[var(--border-subtle)] text-[var(--text-primary)]" />
        </div>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="iniciado_em" className="text-[var(--text-secondary)] text-sm">Iniciado em</Label>
          <Input id="iniciado_em" name="iniciado_em" type="date" value={iniciadoEm} onChange={(e) => setIniciadoEm(e.target.value)} className="bg-[var(--bg-surface-alt)] border-[var(--border-subtle)] text-[var(--text-primary)]" />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="finalizado_em" className="text-[var(--text-secondary)] text-sm">Finalizado em *</Label>
          <Input id="finalizado_em" name="finalizado_em" type="date" value={finalizadoEm} onChange={(e) => setFinalizadoEm(e.target.value)} className="bg-[var(--bg-surface-alt)] border-[var(--border-subtle)] text-[var(--text-primary)]" aria-invalid={!!errors.finalizado_em} />
          {errors.finalizado_em && <span className="text-xs text-[var(--danger)]">{errors.finalizado_em}</span>}
        </div>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="flex flex-col gap-1.5">
          <Label className="text-[var(--text-secondary)] text-sm">Tempo jogado</Label>
          <div className="flex gap-2 items-center">
            <Input id="tempo_jogado_horas" name="tempo_jogado_horas" type="number" min={0} value={horas} onChange={(e) => setHoras(Number(e.target.value))} placeholder="h" className="bg-[var(--bg-surface-alt)] border-[var(--border-subtle)] text-[var(--text-primary)] text-center" aria-label="Horas" />
            <span className="text-[var(--text-muted)]">:</span>
            <Input id="tempo_jogado_minutos" name="tempo_jogado_minutos" type="number" min={0} max={59} value={minutos} onChange={(e) => setMinutos(Number(e.target.value))} placeholder="m" className="bg-[var(--bg-surface-alt)] border-[var(--border-subtle)] text-[var(--text-primary)] text-center" aria-label="Minutos" />
            <span className="text-[var(--text-muted)]">:</span>
            <Input id="tempo_jogado_segundos" name="tempo_jogado_segundos" type="number" min={0} max={59} value={segundos} onChange={(e) => setSegundos(Number(e.target.value))} placeholder="s" className="bg-[var(--bg-surface-alt)] border-[var(--border-subtle)] text-[var(--text-primary)] text-center" aria-label="Segundos" />
          </div>
          {(errors.tempo_jogado_minutos || errors.tempo_jogado_segundos) && (
            <span className="text-xs text-[var(--danger)]">{errors.tempo_jogado_minutos || errors.tempo_jogado_segundos}</span>
          )}
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="nota" className="text-[var(--text-secondary)] text-sm">Nota (1 a 11) *</Label>
          <Input id="nota" name="nota" type="number" min={1} max={11} value={nota} onChange={(e) => setNota(Number(e.target.value))} className="bg-[var(--bg-surface-alt)] border-[var(--border-subtle)] text-[var(--text-primary)]" aria-invalid={!!errors.nota} />
          {errors.nota && <span className="text-xs text-[var(--danger)]">{errors.nota}</span>}
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="dificuldade" className="text-[var(--text-secondary)] text-sm">Dificuldade *</Label>
          <select id="dificuldade" name="dificuldade" value={dificuldade} onChange={(e) => setDificuldade(e.target.value as Dificuldade)} className="h-8 rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-2.5 py-1 text-sm text-[var(--text-primary)] outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)]">
            {DIFICULDADES.map((d) => <option key={d} value={d}>{d}</option>)}
          </select>
        </div>
      </div>
      <div className="flex flex-col gap-1.5">
        <div className="flex justify-between items-center">
          <Label htmlFor="condicao_zeramento" className="text-[var(--text-secondary)] text-sm">Condição de zeramento</Label>
          <span className="text-xs text-[var(--text-muted)]">{condicao.length}/500</span>
        </div>
        <textarea id="condicao_zeramento" name="condicao_zeramento" maxLength={500} rows={3} value={condicao} onChange={(e) => setCondicao(e.target.value)} placeholder="Ex: 100% de conquistas, final secreto, zerado no modo difícil" className="w-full rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)]" />
        {errors.condicao_zeramento && <span className="text-xs text-[var(--danger)]">{errors.condicao_zeramento}</span>}
      </div>
      <label className="flex items-center gap-2 cursor-pointer select-none">
        <input id="destaque" name="destaque" type="checkbox" checked={destaque} onChange={(e) => setDestaque(e.target.checked)} className="h-4 w-4 rounded border-[var(--border-subtle)] accent-[var(--accent)]" />
        <span className="text-sm font-medium text-[var(--text-primary)]">Marcar como jogo destaque do ano</span>
      </label>
      {destaqueError && <span className="text-xs text-[var(--danger)]">{destaqueError}</span>}
      <div className="flex gap-3 justify-end pt-2">
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting} className="border-[var(--border)] text-[var(--text-primary)] hover:bg-[var(--bg-surface-alt)]">Cancelar</Button>
        )}
        <Button type="submit" disabled={isSubmitting} className="bg-[var(--accent)] text-[#0E0F12] font-bold hover:bg-[var(--accent)]/90">
          {isSubmitting ? 'Salvando...' : isEditing ? 'Atualizar registro' : 'Salvar registro'}
        </Button>
      </div>
    </form>
  )
}
