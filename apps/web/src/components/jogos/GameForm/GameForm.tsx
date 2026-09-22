import type { Dificuldade, JogoZeradoDTO, SalvarJogoPayload } from '@/lib/services/jogosService'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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
  const {
    nome, setNome, consoleName, setConsoleName, genero, setGenero, tipo, setTipo,
    iniciadoEm, setIniciadoEm, finalizadoEm, setFinalizadoEm, horas, setHoras,
    minutos, setMinutos, segundos, setSegundos, nota, setNota, dificuldade,
    setDificuldade, condicao, setCondicao, destaque, setDestaque, setIgdbId,
    igdbCapaUrl, setIgdbCapaUrl, setIgdbDescricao, errors, destaqueError,
    isSubmitting, handleSubmit,
  } = useGameForm({ initialData, onSubmit })

  return (
    <form noValidate onSubmit={handleSubmit} className="flex flex-col gap-6 bg-surface p-6 rounded-xl border border-border">
      {errors.form && <div className="text-sm text-[#E05A4E]">{errors.form}</div>}
      <GameFormAutocomplete
        nome={nome}
        onChangeNome={setNome}
        onSelectSugestao={(sugestao) => {
          setNome(sugestao.name)
          setIgdbId(sugestao.id)
          const capa = sugestao.cover?.url
            ? sugestao.cover.url.startsWith('//')
              ? `https:${sugestao.cover.url}`
              : sugestao.cover.url
            : ''
          setIgdbCapaUrl(capa)
          setIgdbDescricao(sugestao.summary || '')
        }}
        igdbCapaUrl={igdbCapaUrl}
        error={errors.nome}
      />
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="console" className="text-secondary text-sm">Console *</Label>
          <Input id="console" name="console" value={consoleName} onChange={(e) => setConsoleName(e.target.value)} placeholder="Ex: Super Nintendo" className="bg-surface-alt border-border-subtle text-primary" aria-invalid={!!errors.console} />
          {errors.console && <span className="text-xs text-[#E05A4E]">{errors.console}</span>}
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="genero" className="text-secondary text-sm">Gênero</Label>
          <Input id="genero" name="genero" value={genero} onChange={(e) => setGenero(e.target.value)} placeholder="Ex: JRPG" className="bg-surface-alt border-border-subtle text-primary" />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="tipo" className="text-secondary text-sm">Tipo</Label>
          <Input id="tipo" name="tipo" value={tipo} onChange={(e) => setTipo(e.target.value)} placeholder="Ex: Principal, DLC" className="bg-surface-alt border-border-subtle text-primary" />
        </div>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="iniciado_em" className="text-secondary text-sm">Iniciado em</Label>
          <Input id="iniciado_em" name="iniciado_em" type="date" value={iniciadoEm} onChange={(e) => setIniciadoEm(e.target.value)} className="bg-surface-alt border-border-subtle text-primary" />
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="finalizado_em" className="text-secondary text-sm">Finalizado em *</Label>
          <Input id="finalizado_em" name="finalizado_em" type="date" value={finalizadoEm} onChange={(e) => setFinalizadoEm(e.target.value)} className="bg-surface-alt border-border-subtle text-primary" aria-invalid={!!errors.finalizado_em} />
          {errors.finalizado_em && <span className="text-xs text-[#E05A4E]">{errors.finalizado_em}</span>}
        </div>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="flex flex-col gap-1.5">
          <Label className="text-secondary text-sm">Tempo jogado</Label>
          <div className="flex gap-2 items-center">
            <Input id="tempo_jogado_horas" name="tempo_jogado_horas" type="number" min={0} value={horas} onChange={(e) => setHoras(Number(e.target.value))} placeholder="h" className="bg-surface-alt border-border-subtle text-primary text-center" aria-label="Horas" />
            <span className="text-muted">:</span>
            <Input id="tempo_jogado_minutos" name="tempo_jogado_minutos" type="number" min={0} max={59} value={minutos} onChange={(e) => setMinutos(Number(e.target.value))} placeholder="m" className="bg-surface-alt border-border-subtle text-primary text-center" aria-label="Minutos" />
            <span className="text-muted">:</span>
            <Input id="tempo_jogado_segundos" name="tempo_jogado_segundos" type="number" min={0} max={59} value={segundos} onChange={(e) => setSegundos(Number(e.target.value))} placeholder="s" className="bg-surface-alt border-border-subtle text-primary text-center" aria-label="Segundos" />
          </div>
          {(errors.tempo_jogado_minutos || errors.tempo_jogado_segundos) && (
            <span className="text-xs text-[#E05A4E]">{errors.tempo_jogado_minutos || errors.tempo_jogado_segundos}</span>
          )}
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="nota" className="text-secondary text-sm">Nota (1 a 11) *</Label>
          <Input id="nota" name="nota" type="number" min={1} max={11} value={nota} onChange={(e) => setNota(Number(e.target.value))} className="bg-surface-alt border-border-subtle text-primary" aria-invalid={!!errors.nota} />
          {errors.nota && <span className="text-xs text-[#E05A4E]">{errors.nota}</span>}
        </div>
        <div className="flex flex-col gap-1.5">
          <Label htmlFor="dificuldade" className="text-secondary text-sm">Dificuldade *</Label>
          <select id="dificuldade" name="dificuldade" value={dificuldade} onChange={(e) => setDificuldade(e.target.value as Dificuldade)} className="h-8 rounded-lg border border-border-subtle bg-surface-alt px-2.5 py-1 text-sm text-primary outline-none focus-visible:ring-2 focus-visible:ring-accent">
            {DIFICULDADES.map((d) => <option key={d} value={d}>{d}</option>)}
          </select>
        </div>
      </div>
      <div className="flex flex-col gap-1.5">
        <div className="flex justify-between items-center">
          <Label htmlFor="condicao_zeramento" className="text-secondary text-sm">Condição de zeramento</Label>
          <span className="text-xs text-muted">{condicao.length}/500</span>
        </div>
        <textarea id="condicao_zeramento" name="condicao_zeramento" maxLength={500} rows={3} value={condicao} onChange={(e) => setCondicao(e.target.value)} placeholder="Ex: 100% de conquistas, final secreto, zerado no modo difícil" className="w-full rounded-lg border border-border-subtle bg-surface-alt px-3 py-2 text-sm text-primary outline-none focus-visible:ring-2 focus-visible:ring-accent" />
        {errors.condicao_zeramento && <span className="text-xs text-[#E05A4E]">{errors.condicao_zeramento}</span>}
      </div>
      <div className="flex flex-col gap-1.5">
        <label className="flex items-center gap-2 cursor-pointer select-none">
          <input id="destaque" name="destaque" type="checkbox" checked={destaque} onChange={(e) => setDestaque(e.target.checked)} className="h-4 w-4 rounded border-border-subtle accent-[#4F7CFF]" />
          <span className="text-sm font-medium text-primary">Marcar como jogo destaque do ano</span>
        </label>
        {destaqueError && <span className="text-xs text-[#E05A4E]">{destaqueError}</span>}
      </div>
      <div className="flex gap-3 justify-end pt-2">
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting} className="border-border text-primary hover:bg-surface-alt">Cancelar</Button>
        )}
        <Button type="submit" disabled={isSubmitting} className="bg-[#4F7CFF] text-[#0E0F12] font-bold hover:bg-[#4F7CFF]/90">
          {isSubmitting ? 'Salvando...' : isEditing ? 'Atualizar registro' : 'Salvar registro'}
        </Button>
      </div>
    </form>
  )
}
