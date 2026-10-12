import { CustomSelect } from '@/components/ui/CustomSelect'
import { CONSOLES_PADRAO } from '@/components/jogos/GameForm/GameForm.constants'
import { PerfilJogoFavoritoSelector } from './PerfilJogoFavoritoSelector'
import type { JogoFavoritoDTO } from '@/types/perfil'

const anoAtual = new Date().getFullYear()

interface PerfilDestaquesCardProps {
  jogoFavorito: JogoFavoritoDTO | null
  consoleFavorito: string
  jogandoDesde: string
  jogandoDesdeEfetivo?: number | null
  erros: Record<string, string>
  onSelectJogo: (jogo: JogoFavoritoDTO | null) => void
  onChangeConsole: (val: string) => void
  onChangeJogandoDesde: (val: string) => void
}

export function PerfilDestaquesCard({
  jogoFavorito,
  consoleFavorito,
  jogandoDesde,
  jogandoDesdeEfetivo,
  erros,
  onSelectJogo,
  onChangeConsole,
  onChangeJogandoDesde,
}: PerfilDestaquesCardProps) {
  const opcoesConsole = [{ value: '', label: 'Nenhum selecionado' }, ...CONSOLES_PADRAO]

  return (
    <section className="space-y-4 rounded-[12px] border border-[var(--border)] bg-[var(--bg-surface)] p-5 sm:p-6">
      <header>
        <h3 className="text-sm font-bold text-[var(--text-primary)]">Em destaque no perfil</h3>
      </header>

      <div className="space-y-4">
        <PerfilJogoFavoritoSelector
          jogoFavorito={jogoFavorito}
          erro={erros.jogo_favorito_id}
          onSelect={onSelectJogo}
        />

        <div className="space-y-1.5">
          <label className="text-xs font-semibold text-[var(--text-secondary)]">Console favorito</label>
          <CustomSelect
            value={consoleFavorito}
            onChange={onChangeConsole}
            options={opcoesConsole}
            placeholder="Selecione um console"
            ariaLabel="Console favorito"
          />
          {erros.console_favorito && <p className="text-xs text-[var(--danger)]">{erros.console_favorito}</p>}
        </div>

        <div className="space-y-1.5">
          <label htmlFor="perfil-jogando-desde" className="text-xs font-semibold text-[var(--text-secondary)]">
            Jogando desde
          </label>
          <input
            id="perfil-jogando-desde"
            type="number"
            min={1970}
            max={anoAtual}
            value={jogandoDesde}
            onChange={(e) => onChangeJogandoDesde(e.target.value)}
            placeholder={jogandoDesdeEfetivo ? String(jogandoDesdeEfetivo) : 'Ex: 1998'}
            className="h-10 w-full rounded-[7px] border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-3 text-xs text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:border-[var(--accent)] focus:outline-none"
          />
          <p className="text-[11px] text-[var(--text-muted)]">
            Calculado pelo primeiro zerado. Edite se jogou antes.
          </p>
          {erros.jogando_desde && <p className="text-xs text-[var(--danger)]">{erros.jogando_desde}</p>}
        </div>
      </div>
    </section>
  )
}
