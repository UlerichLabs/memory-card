import { useEffect, useState } from 'react'
import { Crown } from 'lucide-react'
import { Topbar } from '@/components/layout/Topbar'
import { HallDaFamaProvider, useHallDaFamaStore } from '@/stores/hallDaFamaStore'
import { useJogosStore } from '@/stores/jogosStore'
import type { JogoZeradoDTO } from '@/types/jogos'
import { GameDoAnoCard } from '@/components/hall/GameDoAnoCard'
import { GameDoAnoVazio } from '@/components/hall/GameDoAnoVazio'
import { GamesDaVidaGrade } from '@/components/hall/GamesDaVidaGrade'
import { EscolherGameDoAnoDialog } from '@/components/hall/EscolherGameDoAnoDialog'

function HallDaFamaSkeletons() {
  return (
    <div className="flex flex-col gap-10">
      <section className="flex flex-col gap-[18px]">
        <div className="h-6 w-48 animate-pulse rounded-[6px] bg-[var(--bg-surface-alt)]" />
        <div className="grid grid-cols-1 gap-5 md:grid-cols-2 xl:grid-cols-4">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="h-[230px] animate-pulse rounded-[16px] bg-[var(--bg-surface)]" />
          ))}
        </div>
      </section>
      <section className="flex flex-col gap-[18px]">
        <div className="h-6 w-48 animate-pulse rounded-[6px] bg-[var(--bg-surface-alt)]" />
        <div className="grid grid-cols-2 gap-x-[14px] gap-y-[20px] sm:grid-cols-4 xl:grid-cols-6 xl:gap-x-5 xl:gap-y-6">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} className="aspect-[3/4] animate-pulse rounded-[12px] bg-[var(--bg-surface)]" />
          ))}
        </div>
      </section>
    </div>
  )
}

function HallDaFamaConteudo() {
  const { resumo, gamesDaVida, isLoading, error, carregarHallDaFama } = useHallDaFamaStore()
  const { abrirModalRegistro } = useJogosStore()

  const [dialogAberto, setDialogAberto] = useState(false)
  const [dialogAno, setDialogAno] = useState<number>(new Date().getFullYear())
  const [dialogJogoAtual, setDialogJogoAtual] = useState<JogoZeradoDTO | null>(null)

  useEffect(() => {
    document.title = 'Hall da Fama'
    const ctrl = new AbortController()
    carregarHallDaFama(ctrl.signal).catch(() => {})
    return () => {
      ctrl.abort()
      document.title = 'Memory Card'
    }
  }, [carregarHallDaFama])

  const handleAbrirEscolher = (ano: number) => {
    setDialogAno(ano)
    setDialogJogoAtual(null)
    setDialogAberto(true)
  }

  const handleAbrirTrocar = (ano: number, jogo: JogoZeradoDTO) => {
    setDialogAno(ano)
    setDialogJogoAtual(jogo)
    setDialogAberto(true)
  }

  const hallVazio = !isLoading && !error && resumo.length === 0

  return (
    <div className="min-h-full w-full overflow-x-hidden bg-[var(--bg-primary)]">
      <Topbar />

        <main className="mx-auto flex max-w-7xl flex-col gap-10 p-[20px_16px_32px] sm:p-[32px_40px_48px]">
        <header className="flex flex-col gap-1.5">
          <h1 className="text-[24px] font-bold tracking-[-0.01em] text-[var(--text-primary)] sm:text-[28px]">
            Hall da Fama
          </h1>
          <p className="text-[14px] text-[var(--hall-subtitulo)]">
            O Game do Ano de cada ano e os jogos que marcaram a sua vida.
          </p>
        </header>

        {isLoading ? (
          <HallDaFamaSkeletons />
        ) : error ? (
          <div className="flex min-h-[300px] flex-col items-center justify-center gap-4 text-center">
            <h2 className="text-xl font-bold text-[var(--text-primary)]">Erro ao carregar o Hall da Fama</h2>
            <p className="text-sm text-[var(--text-secondary)]">Não foi possível carregar os dados. Verifique sua conexão.</p>
            <button
              type="button"
              onClick={() => carregarHallDaFama()}
              className="btn-primario px-4 py-2 text-sm"
            >
              Tentar novamente
            </button>
          </div>
        ) : hallVazio ? (
          <div className="flex min-h-[360px] flex-col items-center justify-center gap-3 rounded-[16px] border border-dashed border-[var(--hall-card-vazio-border)] p-8 text-center sm:p-12">
            <h2 className="text-[20px] font-bold text-[var(--text-primary)]">
              Seu Hall da Fama ainda está vazio
            </h2>
            <p className="text-[14px] text-[var(--hall-subtitulo)]">
              Registre jogos zerados para escolher seus Games do Ano.
            </p>
            <button
              type="button"
              onClick={() => abrirModalRegistro()}
              className="btn-primario mt-2 inline-flex items-center justify-center px-4 py-2 text-[13px]"
            >
              + Registrar jogo
            </button>
          </div>
        ) : (
          <div className="flex flex-col gap-10">
            <section className="flex flex-col gap-[18px]">
              <div className="flex items-center gap-[10px]">
                <Crown className="h-5 w-5 fill-current text-[var(--hall-ouro)]" aria-hidden="true" />
                <h2 className="text-[20px] font-bold text-[var(--text-primary)]">
                  Game do Ano
                </h2>
                <span className="text-[13px] text-[var(--hall-muted)]">
                  Um por ano
                </span>
              </div>

              <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6">
                {resumo.map((item) =>
                  item.game_do_ano ? (
                    <GameDoAnoCard
                      key={item.ano}
                      ano={item.ano}
                      totalJogos={item.total_jogos}
                      jogo={item.game_do_ano}
                      onTrocar={handleAbrirTrocar}
                    />
                  ) : (
                    <GameDoAnoVazio
                      key={item.ano}
                      ano={item.ano}
                      totalJogos={item.total_jogos}
                      onEscolher={handleAbrirEscolher}
                    />
                  )
                )}
              </div>
            </section>

            <GamesDaVidaGrade jogos={gamesDaVida} />
          </div>
        )}
      </main>

      <EscolherGameDoAnoDialog
        open={dialogAberto}
        onOpenChange={setDialogAberto}
        ano={dialogAno}
        jogoAtual={dialogJogoAtual}
      />
    </div>
  )
}

export function HallDaFamaPage() {
  return (
    <HallDaFamaProvider>
      <HallDaFamaConteudo />
    </HallDaFamaProvider>
  )
}
