import { Gamepad2 } from 'lucide-react'
import type { JogoZeradoDTO } from '@/types/jogos'
import { formatarCapaIGDB, isoParaDataPt } from '@/lib/utils'

export interface JogoDetalheColunaEsquerdaProps {
  jogo: JogoZeradoDTO
  className?: string
}

export function JogoDetalheColunaEsquerda({ jogo, className = '' }: JogoDetalheColunaEsquerdaProps) {
  const capaUrl = formatarCapaIGDB(jogo.igdb_capa_url, 't_cover_big_2x')
  const numeroFormatado = `Registro #${jogo.numero ?? jogo.id}`

  const dataCriacao = isoParaDataPt(jogo.created_at)
  const dataAtualizacao = isoParaDataPt(jogo.updated_at)
  const datasIguais = !dataAtualizacao || dataCriacao === dataAtualizacao
  const datasTexto = datasIguais
    ? `Criado em ${dataCriacao}`
    : `Criado em ${dataCriacao} · atualizado em ${dataAtualizacao}`

  return (
    <aside className={`flex flex-col items-center sm:items-start gap-3.5 ${className}`.trim()}>
      <div className="relative aspect-[3/4] w-[200px] sm:w-[240px] lg:w-[320px] lg:h-[427px] overflow-hidden rounded-[14px] lg:rounded-[16px] border border-[var(--detalhe-capa-border)] bg-[var(--detalhe-capa-bg)]">
        {capaUrl ? (
          <img
            src={capaUrl}
            alt={jogo.nome}
            className="h-full w-full object-cover"
          />
        ) : (
          <div className="flex h-full w-full flex-col items-center justify-center p-6 text-[var(--detalhe-capa-icon-sem-arte)]">
            <Gamepad2 className="h-12 w-12" aria-hidden="true" />
          </div>
        )}
      </div>

      <div className="hidden sm:flex flex-col gap-0.5 text-[12px] tabular-nums text-[var(--detalhe-text-meta)]">
        <p className="font-medium">{numeroFormatado}</p>
        <p>{datasTexto}</p>
      </div>
    </aside>
  )
}
