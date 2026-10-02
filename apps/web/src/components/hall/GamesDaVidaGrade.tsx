import { Link } from 'react-router-dom'
import { Crown, Gamepad2 } from 'lucide-react'
import type { JogoZeradoDTO } from '@/types/jogos'
import { formatarCapaIGDB } from '@/lib/utils'

export interface GamesDaVidaGradeProps {
  jogos: JogoZeradoDTO[]
}

export function GamesDaVidaGrade({ jogos }: GamesDaVidaGradeProps) {
  const contagemTexto = jogos.length === 1 ? '1 jogo' : `${jogos.length} jogos`

  return (
    <section className="flex flex-col gap-[18px]">
      <div className="flex flex-col justify-between gap-3 md:flex-row md:items-center">
        <div className="flex flex-col gap-1.5">
          <div className="flex items-center gap-[10px]">
            <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-[6px] bg-[var(--hall-ouro)] text-[12px] font-extrabold text-[var(--hall-dark-icon)]">
              11
            </div>
            <h2 className="text-[20px] font-bold text-[var(--text-primary)]">
              Games da Vida
            </h2>
            <span className="rounded-full border border-[var(--hall-pill-contagem-border)] bg-[var(--hall-pill-contagem-bg)] px-[9px] py-[3px] text-[12px] font-semibold text-[var(--hall-pill-contagem-text)]">
              {contagemTexto}
            </span>
          </div>
          <p className="text-[13px] text-[var(--hall-muted)]">
            Todos os jogos que receberam nota 11, do zeramento mais recente ao mais antigo.
          </p>
        </div>

        <Link
          to="/biblioteca?nota_min=11"
          className="self-start text-[14px] font-medium text-[var(--hall-link-biblioteca)] transition-opacity hover:opacity-80 md:self-auto"
        >
          Ver na Biblioteca →
        </Link>
      </div>

      {jogos.length === 0 ? (
        <div className="flex min-h-[160px] flex-col items-center justify-center gap-2 rounded-[16px] border border-dashed border-[var(--hall-card-vazio-border)] p-6 text-center">
          <span className="text-[15px] font-semibold text-[var(--text-primary)]">
            Nenhum Game da Vida ainda
          </span>
          <p className="text-[13px] text-[var(--hall-muted)]">
            Quando um jogo marcar a sua vida, dê nota 11 a ele.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-x-3 gap-y-5 sm:grid-cols-4 lg:grid-cols-6 xl:grid-cols-8">
          {jogos.map((jogo) => {
            const capaUrl = formatarCapaIGDB(jogo.igdb_capa_url, 't_cover_big')
            const anoZerado = jogo.finalizado_em ? new Date(jogo.finalizado_em).getUTCFullYear() : null
            const metaTexto = anoZerado ? `${jogo.console} · ${anoZerado}` : jogo.console

            return (
              <Link
                key={jogo.id}
                to={`/biblioteca/${jogo.id}`}
                className="group flex flex-col overflow-hidden transition-transform duration-150 hover:-translate-y-0.5"
              >
                <div className="relative aspect-[3/4] w-full overflow-hidden rounded-[8px] border border-[var(--hall-ouro)] bg-[var(--biblioteca-card-cover-bg)] shadow-[0_0_10px_1px_var(--hall-ouro)]">
                  {capaUrl ? (
                    <img
                      src={capaUrl}
                      alt={jogo.nome}
                      className="h-full w-full object-cover transition duration-150 group-hover:opacity-90"
                      loading="lazy"
                    />
                  ) : (
                    <div className="flex h-full w-full flex-col items-center justify-center p-3 text-center text-[var(--biblioteca-control-placeholder)]">
                      <Gamepad2 className="h-8 w-8 opacity-40" aria-hidden="true" />
                      <span className="mt-2 line-clamp-2 text-xs font-semibold text-[var(--biblioteca-text-secondary)]">
                        {jogo.nome}
                      </span>
                    </div>
                  )}

                  <div className="absolute right-1.5 top-1.5 z-10">
                    <span className="flex h-6 min-w-6 items-center justify-center rounded-full bg-[var(--hall-ouro)] px-1.5 text-[10px] font-bold text-[var(--hall-dark-icon)]">11</span>
                  </div>

                  {jogo.destaque && (
                    <div className="absolute bottom-1.5 left-1.5 rounded-[4px] bg-[var(--ouro-jogo-ano)] px-1.5 py-0.5 text-[9.5px] font-bold text-[var(--ouro-jogo-ano-text)]">
                      <Crown className="h-3 w-3" aria-hidden="true" />
                      <span>{anoZerado ? `GOTY ${anoZerado}` : 'GOTY'}</span>
                      {anoZerado && <span className="sr-only">Jogo do ano {anoZerado}</span>}
                    </div>
                  )}
                </div>

                <div className="flex flex-col gap-0.5 pt-2.5">
                  <h3
                    className="line-clamp-2 min-h-[32px] text-[12px] font-semibold leading-[1.35] text-[var(--biblioteca-text-primary)] transition-colors group-hover:text-[var(--hall-ano-ouro)]"
                    title={jogo.nome}
                  >
                    {jogo.nome}
                  </h3>
                  <p className="truncate text-[10px] text-[var(--hall-plataforma)]">
                    {metaTexto}
                  </p>
                </div>
              </Link>
            )
          })}
        </div>
      )}
    </section>
  )
}
