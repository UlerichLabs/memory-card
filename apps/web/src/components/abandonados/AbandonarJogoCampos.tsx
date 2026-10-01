import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { CustomSelect } from '@/components/ui/CustomSelect'
import { DateInput } from '@/components/jogos/GameForm/DateInput'
import { TimeInput } from '@/components/jogos/GameForm/TimeInput'
import { GameFormAutocomplete } from '@/components/jogos/GameForm/GameFormAutocomplete'
import type { OrigemFilaInfo } from '@/stores/abandonadosStore'
import type { IGDBJogoSugestao } from '@/lib/services/jogosService'

export interface AbandonarJogoCamposProps {
  origemFila?: OrigemFilaInfo
  aviso?: string
  errors: Record<string, string>
  nome: string
  onChangeNome: (value: string) => void
  onSelectSugestao: (sugestao: IGDBJogoSugestao) => void
  consoleName: string
  setConsoleName: (value: string) => void
  igdbId: number | null
  consoleOptions: Array<{ value: string; label: string }>
  abandonadoEm: string
  setAbandonadoEm: (value: string) => void
  horas: string
  minutos: string
  segundos: string
  setHoras: (value: string) => void
  setMinutos: (value: string) => void
  setSegundos: (value: string) => void
  motivo: string
  setMotivo: (value: string) => void
  iniciadoEm?: string
}

export function AbandonarJogoCampos({
  origemFila,
  aviso,
  errors,
  nome,
  onChangeNome,
  onSelectSugestao,
  consoleName,
  setConsoleName,
  igdbId,
  consoleOptions,
  abandonadoEm,
  setAbandonadoEm,
  horas,
  minutos,
  segundos,
  setHoras,
  setMinutos,
  setSegundos,
  motivo,
  setMotivo,
  iniciadoEm,
}: AbandonarJogoCamposProps) {
  const erroTempo = errors.tempo_jogado_horas || errors.tempo_jogado_minutos || errors.tempo_jogado_segundos

  return (
    <div className="flex flex-col gap-4">
      {origemFila && (
        <div
          className={
            'rounded-lg border border-[var(--abandonado-banner-border)] ' +
            'bg-[var(--abandonado-banner-bg)] px-4 py-3 text-sm text-[var(--abandonado-banner-text)]'
          }
        >
          Vindo da fila {origemFila.listaNome}. Ao salvar, o jogo sai da fila.
        </div>
      )}
      {aviso && <div className="rounded-lg border border-[var(--abandonado-banner-border)] bg-[var(--abandonado-banner-bg)] px-4 py-3 text-sm text-[var(--abandonado-banner-text)]">{aviso}</div>}
      {iniciadoEm && <p className="text-xs text-[var(--text-muted)]">Começou em {iniciadoEm.split('-').reverse().join('/')}</p>}
      {errors.form && <div className="text-sm text-[var(--danger)]">{errors.form}</div>}
      <GameFormAutocomplete
        nome={nome}
        onChangeNome={onChangeNome}
        onSelectSugestao={onSelectSugestao}
        error={errors.nome}
      />
      <div className="flex flex-col gap-1.5">
        <Label htmlFor="console" className="text-sm text-[var(--text-secondary)]">Plataforma *</Label>
        {igdbId ? (
          <CustomSelect
            id="console"
            name="console"
            value={consoleName}
            onChange={(v) => setConsoleName(v.slice(0, 100))}
            options={consoleOptions}
            placeholder="Selecione a plataforma"
            error={!!errors.console}
            ariaLabel="Plataforma"
          />
        ) : (
          <Input
            id="console"
            name="console"
            maxLength={100}
            value={consoleName}
            onChange={(e) => setConsoleName(e.target.value)}
            placeholder="Ex: PC, PlayStation"
            aria-invalid={!!errors.console}
            className="bg-[var(--bg-surface-alt)] text-[var(--text-primary)]"
          />
        )}
        {errors.console && <span className="text-xs text-[var(--danger)]">{errors.console}</span>}
      </div>
      <div className="grid min-w-0 gap-4 sm:grid-cols-2">
        <DateInput
          id="abandonado_em"
          label="Abandonado em *"
          value={abandonadoEm}
          onChange={setAbandonadoEm}
          error={errors.abandonado_em}
        />
        <div className="min-w-0">
          <Label className="text-sm text-[var(--text-secondary)]">Tempo jogado</Label>
          <div className="mt-1">
            <TimeInput
              horas={horas}
              minutos={minutos}
              segundos={segundos}
              setHoras={setHoras}
              setMinutos={setMinutos}
              setSegundos={setSegundos}
            />
          </div>
          {erroTempo && <span className="text-xs text-[var(--danger)]">{erroTempo}</span>}
        </div>
      </div>
      <div>
        <div className="flex items-center justify-between">
          <Label htmlFor="motivo" className="text-sm text-[var(--text-secondary)]">Motivo</Label>
          <span className="text-xs text-[var(--text-muted)]">{motivo.length}/500</span>
        </div>
        <textarea
          id="motivo"
          name="motivo"
          maxLength={500}
          rows={3}
          value={motivo}
          onChange={(e) => setMotivo(e.target.value)}
          aria-invalid={!!errors.motivo}
          placeholder="Por que você desistiu ou parou de jogar?"
          className={
            'mt-1 w-full rounded-lg border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] ' +
            'px-3 py-2 text-sm text-[var(--text-primary)] outline-none ' +
            'focus-visible:ring-2 focus-visible:ring-[var(--accent)]'
          }
        />
        {errors.motivo && <span className="text-xs text-[var(--danger)]">{errors.motivo}</span>}
      </div>
    </div>
  )
}
