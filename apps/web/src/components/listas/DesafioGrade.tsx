import { useState, useEffect } from 'react'
import { RefreshCw } from 'lucide-react'
import { useListasStore } from '@/stores/listasStore'
import { DesafioCard } from './DesafioCard'
import { AdicionarJogoInput } from './AdicionarJogoInput'
import type { ListaDetalhada, FiltroAba } from '@/types/listas'

export interface DesafioGradeProps {
  lista: ListaDetalhada
}

export function DesafioGrade({ lista }: DesafioGradeProps) {
  const { filtroAba, setFiltroAba, sincronizarFranquia, adicionarItem, removerItem } = useListasStore()
  const [isSincronizando, setIsSincronizando] = useState(false)
  const [msgSinc, setMsgSinc] = useState<string | null>(null)

  const regraTipo = lista.regra?.tipo ?? 'manual'
  const isContagem = regraTipo === 'plataforma' || regraTipo === 'genero'
  const isFranquia = regraTipo === 'franquia'
  const isManual = regraTipo === 'manual'

  useEffect(() => {
    if (!msgSinc) return
    const timer = setTimeout(() => setMsgSinc(null), 4000)
    return () => clearTimeout(timer)
  }, [msgSinc])

  const handleSincronizar = async () => {
    setIsSincronizando(true)
    setMsgSinc(null)
    try {
      const res = await sincronizarFranquia()
      setMsgSinc(res.adicionados > 0 ? `${res.adicionados} jogos novos adicionados` : 'Nenhum jogo novo')
    } catch {
      setMsgSinc('Erro ao sincronizar do IGDB')
    } finally {
      setIsSincronizando(false)
    }
  }

  const itens = lista.itens ?? []
  const totalZerados = itens.filter((it) => it.zerado).length
  const totalPendentes = itens.filter((it) => !it.zerado).length

  const itensFiltrados = itens.filter((it) => {
    if (filtroAba === 'zerados') return it.zerado
    if (filtroAba === 'pendentes') return !it.zerado
    return true
  })

  const faltamContagem = isContagem && lista.progresso
    ? Math.max(0, lista.progresso.meta - lista.progresso.feitos)
    : 0

  return (
    <div className="flex flex-col gap-5">
      {!isContagem && (
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div
            role="tablist"
            aria-label="Filtrar jogos do desafio"
            className="flex items-center gap-1 rounded-[10px] border border-[var(--lista-tabs-border)] bg-[var(--lista-tabs-bg)] p-1"
          >
            {(['todos', 'zerados', 'pendentes'] as FiltroAba[]).map((tab) => {
              const ativo = filtroAba === tab
              const count = tab === 'todos' ? itens.length : tab === 'zerados' ? totalZerados : totalPendentes
              const label = tab === 'todos' ? 'Todos' : tab === 'zerados' ? 'Zerados' : 'Pendentes'
              return (
                <button
                  key={tab}
                  type="button"
                  role="tab"
                  aria-selected={ativo}
                  onClick={() => setFiltroAba(tab)}
                  className={`flex h-8 items-center rounded-[7px] px-3 text-[13px] font-semibold transition-colors ${
                    ativo
                      ? 'bg-[var(--lista-tab-active-bg)] text-[var(--lista-tab-active-text)]'
                      : 'bg-transparent text-[var(--lista-text-muted)] hover:text-[var(--text-primary)]'
                  }`}
                >
                  {label} · {count}
                </button>
              )
            })}
          </div>

          {isFranquia && (
            <div className="flex items-center gap-2.5">
              {msgSinc && <span className="text-[12px] text-[var(--lista-text-light)] animate-fade-in">{msgSinc}</span>}
              <button
                type="button"
                disabled={isSincronizando}
                onClick={handleSincronizar}
                className="flex h-9 items-center gap-2 rounded-lg border border-[var(--lista-btn-icon-border)] px-3 text-[13px] font-medium text-[var(--lista-btn-icon-text)] hover:bg-[var(--bg-surface-alt)] disabled:opacity-50"
              >
                <RefreshCw className={`h-3.5 w-3.5 ${isSincronizando ? 'animate-spin' : ''}`} />
                <span>Atualizar do IGDB</span>
              </button>
            </div>
          )}
        </div>
      )}

      {isManual && (
        <AdicionarJogoInput
          onAdicionar={adicionarItem}
          placeholder="Adicionar jogo ao desafio: busque no IGDB ou digite o nome…"
          ariaLabel="Adicionar jogo ao desafio"
        />
      )}

      {itensFiltrados.length === 0 && !isContagem && (
        <div className="py-12 text-center text-[14px] text-[var(--lista-text-muted)]">
          Nenhum jogo aqui ainda.
        </div>
      )}

      <div className="grid grid-cols-3 gap-x-4 gap-y-5 md:grid-cols-4 xl:grid-cols-6">
        {itensFiltrados.map((it) => (
          <DesafioCard
            key={it.id}
            item={it}
            isManual={isManual}
            onRemover={isManual ? removerItem : undefined}
          />
        ))}

        {isContagem && faltamContagem > 0 && (
          <div className="flex aspect-[3/4] flex-col items-center justify-center gap-1 rounded-[10px] border border-dashed border-[var(--lista-card-pendente-border)] p-4 text-center">
            <span className="text-[26px] font-extrabold text-[var(--lista-text-dimmer)]">
              +{faltamContagem}
            </span>
            <span className="text-[11px] text-[var(--lista-text-dim)]">
              para completar
            </span>
          </div>
        )}
      </div>
    </div>
  )
}
