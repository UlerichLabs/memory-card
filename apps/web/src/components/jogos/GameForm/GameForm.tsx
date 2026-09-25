import { useContext } from 'react'
import { jogosService, type Dificuldade, type JogoZeradoDTO, type SalvarJogoPayload } from '@/lib/services/jogosService'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { CustomSelect } from '@/components/ui/CustomSelect'
import { DatePicker } from '@/components/ui/DatePicker'
import { AuthContext } from '@/store/authStore'
import { formatarCapaIGDB } from '@/lib/utils'
import { GameFormAutocomplete } from './GameFormAutocomplete'
import { useGameForm } from './useGameForm'
import { DIFICULDADE_OPCOES, CONSOLES_PADRAO } from './GameForm.constants'

export interface GameFormProps {
  initialData?: Partial<JogoZeradoDTO>
  onSubmit: (payload: SalvarJogoPayload) => Promise<void>
  onCancel?: () => void
  isEditing?: boolean
}

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

  const consoleOptions = plataformas.length > 0 ? plataformas.map((p) => ({ value: p, label: p })) : CONSOLES_PADRAO

  return (
    <form noValidate onSubmit={handleSubmit} className="flex flex-col gap-4">
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
              if (!consoleName || !nomes.includes(consoleName)) setConsoleName(nomes[0])
            }
          } catch {}
        }}
        igdbCapaUrl={igdbCapaUrl}
        igdbDescricao={igdbDescricao}
        error={errors.nome}
      />
      <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="console" className="text-[var(--text-secondary)] text-sm">Console *</Label>
          <CustomSelect id="console" name="console" value={consoleName} onChange={setConsoleName} options={consoleOptions} placeholder="Selecione o console" error={!!errors.console} ariaLabel="Console" />
          {plataformas.length > 0 && <span className="text-[11px] text-[var(--text-muted)]">Disponível para este jogo no IGDB</span>}
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
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="iniciado_em" className="text-[var(--text-secondary)] text-sm">Iniciado em</Label>
          <DatePicker id="iniciado_em" name="iniciado_em" value={iniciadoEm} onChange={setIniciadoEm} ariaLabel="Iniciado em" />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="finalizado_em" className="text-[var(--text-secondary)] text-sm">Finalizado em *</Label>
          <DatePicker id="finalizado_em" name="finalizado_em" value={finalizadoEm} onChange={setFinalizadoEm} error={!!errors.finalizado_em} ariaLabel="Finalizado em" ariaInvalid={!!errors.finalizado_em} />
          {errors.finalizado_em && <span className="text-xs text-[var(--danger)]">{errors.finalizado_em}</span>}
        </div>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
        <div className="flex flex-col gap-1.5">
          <Label className="text-[var(--text-secondary)] text-sm">Tempo jogado</Label>
          <div className="flex h-8 items-center justify-between rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-2.5">
            <div className="flex items-center gap-1">
              <input id="tempo_jogado_horas" name="tempo_jogado_horas" type="text" inputMode="numeric" placeholder="0" value={horas} onChange={(e) => { if (/^\d*$/.test(e.target.value)) setHoras(e.target.value) }} className="w-8 bg-transparent text-center text-sm font-medium text-[var(--text-primary)] placeholder:text-[var(--text-muted)] outline-none" aria-label="Horas" />
              <span className="text-[11px] font-semibold text-[var(--text-muted)]">h</span>
            </div>
            <span className="text-[var(--border-subtle)] text-xs">|</span>
            <div className="flex items-center gap-1">
              <input id="tempo_jogado_minutos" name="tempo_jogado_minutos" type="text" inputMode="numeric" placeholder="0" value={minutos} onChange={(e) => { if (/^\d*$/.test(e.target.value)) setMinutos(e.target.value) }} className="w-8 bg-transparent text-center text-sm font-medium text-[var(--text-primary)] placeholder:text-[var(--text-muted)] outline-none" aria-label="Minutos" />
              <span className="text-[11px] font-semibold text-[var(--text-muted)]">m</span>
            </div>
            <span className="text-[var(--border-subtle)] text-xs">|</span>
            <div className="flex items-center gap-1">
              <input id="tempo_jogado_segundos" name="tempo_jogado_segundos" type="text" inputMode="numeric" placeholder="0" value={segundos} onChange={(e) => { if (/^\d*$/.test(e.target.value)) setSegundos(e.target.value) }} className="w-8 bg-transparent text-center text-sm font-medium text-[var(--text-primary)] placeholder:text-[var(--text-muted)] outline-none" aria-label="Segundos" />
              <span className="text-[11px] font-semibold text-[var(--text-muted)]">s</span>
            </div>
          </div>
          {(errors.tempo_jogado_minutos || errors.tempo_jogado_segundos) && <span className="text-xs text-[var(--danger)]">{errors.tempo_jogado_minutos || errors.tempo_jogado_segundos}</span>}
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="nota" className="text-[var(--text-secondary)] text-sm">Nota (1 a 11) *</Label>
          <Input id="nota" name="nota" type="number" min={1} max={11} value={nota} onChange={(e) => setNota(Number(e.target.value))} className="bg-[var(--bg-surface-alt)] border-[var(--border-subtle)] text-[var(--text-primary)] no-spinner" aria-invalid={!!errors.nota} />
          {errors.nota && <span className="text-xs text-[var(--danger)]">{errors.nota}</span>}
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="dificuldade" className="text-[var(--text-secondary)] text-sm">Dificuldade *</Label>
          <CustomSelect id="dificuldade" name="dificuldade" value={dificuldade} onChange={(v) => setDificuldade(v as Dificuldade)} options={DIFICULDADE_OPCOES} ariaLabel="Dificuldade" />
        </div>
      </div>
      <div className="flex flex-col gap-1.5">
        <div className="flex justify-between items-center">
          <Label htmlFor="condicao_zeramento" className="text-[var(--text-secondary)] text-sm">Condição de zeramento</Label>
          <span className="text-xs text-[var(--text-muted)]">{condicao.length}/500</span>
        </div>
        <textarea id="condicao_zeramento" name="condicao_zeramento" maxLength={500} rows={3} value={condicao} onChange={(e) => setCondicao(e.target.value)} placeholder="Ex: 100% de conquistas, final secreto, zerado no modo difícil" className="w-full rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-3 py-2 text-sm text-[var(--text-primary)] outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)] custom-scrollbar" />
        {errors.condicao_zeramento && <span className="text-xs text-[var(--danger)]">{errors.condicao_zeramento}</span>}
      </div>
      <div className="flex flex-col gap-1.5">
        <div className="flex items-center gap-3">
          <button id="destaque" type="button" role="switch" aria-checked={destaque} aria-label="Marcar como jogo destaque do ano" onClick={() => setDestaque(!destaque)} className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)] ${destaque ? 'bg-[var(--accent)]' : 'bg-[var(--border-subtle)]'}`}>
            <span className={`pointer-events-none inline-block h-4 w-4 transform rounded-full shadow ring-0 transition duration-200 ease-in-out ${destaque ? 'translate-x-4 bg-[#0E0F12]' : 'translate-x-0 bg-[var(--text-secondary)]'}`} />
          </button>
          <label htmlFor="destaque" className="text-sm font-medium text-[var(--text-primary)] cursor-pointer select-none">Marcar como jogo destaque do ano</label>
        </div>
        {destaqueError && <span className="text-xs text-[var(--danger)]">{destaqueError}</span>}
      </div>
      <div className="flex gap-3 justify-end pt-3 border-t border-[var(--border)]">
        {onCancel && <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting} className="border-[var(--border)] text-[var(--text-primary)] hover:bg-[var(--bg-surface-alt)]">Cancelar</Button>}
        <Button type="submit" disabled={isSubmitting} className="bg-[var(--accent)] text-[#0E0F12] font-bold hover:bg-[var(--accent)]/90">{isSubmitting ? 'Salvando...' : isEditing ? 'Atualizar registro' : 'Salvar registro'}</Button>
      </div>
    </form>
  )
}
