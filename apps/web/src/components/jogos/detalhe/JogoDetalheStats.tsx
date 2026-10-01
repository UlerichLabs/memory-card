import type { JogoZeradoDTO } from '@/types/jogos'
import { isoParaDataPt } from '@/lib/utils'
import { obterRotuloNota } from '@/lib/rotulosNota'
import { NotaBadge } from '../NotaBadge'
import { DificuldadePill } from '../DificuldadePill'
import { DificuldadeBarra } from '../DificuldadeBarra'

export interface JogoDetalheStatsProps {
  jogo: JogoZeradoDTO
}

function formatarTempoExtenso(s: number): string {
  if (s <= 0) return '—'
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  return `${h}h ${m}m`
}

function formatarTempoHMS(s: number): string {
  const h = String(Math.floor(s / 3600)).padStart(2, '0')
  const m = String(Math.floor((s % 3600) / 60)).padStart(2, '0')
  const sec = String(s % 60).padStart(2, '0')
  return `${h}:${m}:${sec}`
}

function calcularDiferencaDias(inicioIso: string, fimIso: string): number {
  const dInicio = new Date(inicioIso.slice(0, 10))
  const dFim = new Date(fimIso.slice(0, 10))
  const diffTime = dFim.getTime() - dInicio.getTime()
  return Math.max(0, Math.round(diffTime / (1000 * 60 * 60 * 24)))
}

export function JogoDetalheStats({ jogo }: JogoDetalheStatsProps) {
  const rotuloNota = obterRotuloNota(jogo.nota)
  const dataFinal = isoParaDataPt(jogo.finalizado_em)

  let subtituloFinalizado: string | null = null
  if (jogo.iniciado_em) {
    const dataInicio = isoParaDataPt(jogo.iniciado_em)
    const dias = calcularDiferencaDias(jogo.iniciado_em, jogo.finalizado_em)
    subtituloFinalizado = `Iniciado em ${dataInicio} · ${dias} ${dias === 1 ? 'dia' : 'dias'}`
  }

  return (
    <div className="grid grid-cols-2 gap-3 lg:grid-cols-4 lg:gap-4">
      <div className="flex flex-col justify-between rounded-[14px] border border-[var(--detalhe-bloco-border)] bg-[var(--detalhe-bloco-bg)] p-3.5 sm:p-[18px]">
        <span className="text-[11px] sm:text-[12px] font-semibold uppercase tracking-[.04em] text-[var(--detalhe-bloco-rotulo)]">
          Nota
        </span>
        <div className="mt-3 flex items-center gap-3 sm:gap-3.5">
          <NotaBadge nota={jogo.nota} tamanho="md" pulsar={jogo.nota === 11} className="sm:hidden" />
          <NotaBadge nota={jogo.nota} tamanho="lg" pulsar={jogo.nota === 11} className="hidden sm:inline-flex" />
          <div className="flex flex-col">
            <span
              style={{ color: `var(--nota-${jogo.nota}-text)` }}
              className="text-[14px] sm:text-[16px] font-bold"
            >
              {rotuloNota}
            </span>
            <span className="text-[12px] text-[var(--detalhe-bloco-rotulo)]">de 11</span>
          </div>
        </div>
      </div>

      <div className="flex flex-col justify-between rounded-[14px] border border-[var(--detalhe-bloco-border)] bg-[var(--detalhe-bloco-bg)] p-3.5 sm:p-[18px]">
        <span className="text-[11px] sm:text-[12px] font-semibold uppercase tracking-[.04em] text-[var(--detalhe-bloco-rotulo)]">
          Dificuldade
        </span>
        <div className="mt-3 flex flex-col items-start gap-2">
          <DificuldadePill nivel={jogo.dificuldade} variante="solida" />
          <DificuldadeBarra nivel={jogo.dificuldade} className="hidden sm:flex" />
        </div>
      </div>

      <div className="flex flex-col justify-between rounded-[14px] border border-[var(--detalhe-bloco-border)] bg-[var(--detalhe-bloco-bg)] p-3.5 sm:p-[18px]">
        <span className="text-[11px] sm:text-[12px] font-semibold uppercase tracking-[.04em] text-[var(--detalhe-bloco-rotulo)]">
          Tempo jogado
        </span>
        <div className="mt-3 flex flex-col">
          <span className="text-[20px] sm:text-[26px] font-bold tabular-nums text-[var(--text-primary)]">
            {formatarTempoExtenso(jogo.tempo_jogado)}
          </span>
          {jogo.tempo_jogado > 0 && (
            <span className="text-[12px] text-[var(--detalhe-bloco-rotulo)]">
              {formatarTempoHMS(jogo.tempo_jogado)}
            </span>
          )}
        </div>
      </div>

      <div className="flex flex-col justify-between rounded-[14px] border border-[var(--detalhe-bloco-border)] bg-[var(--detalhe-bloco-bg)] p-3.5 sm:p-[18px]">
        <span className="text-[11px] sm:text-[12px] font-semibold uppercase tracking-[.04em] text-[var(--detalhe-bloco-rotulo)]">
          Finalizado em
        </span>
        <div className="mt-3 flex flex-col">
          <span className="text-[20px] sm:text-[26px] font-bold tabular-nums text-[var(--text-primary)]">
            {dataFinal}
          </span>
          {subtituloFinalizado && (
            <span className="text-[12px] text-[var(--detalhe-bloco-rotulo)]">
              {subtituloFinalizado}
            </span>
          )}
        </div>
      </div>
    </div>
  )
}
