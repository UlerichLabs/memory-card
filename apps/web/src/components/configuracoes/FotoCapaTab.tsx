import { useEffect, useState } from 'react'
import { Search, AlertCircle, Check, ImageIcon } from 'lucide-react'
import { jogosService, type JogoZeradoDTO } from '@/lib/services/jogosService'

interface FotoCapaTabProps {
  isLoading: boolean
  erroBackend: string | null
  onSubmit: (jogoId: number) => Promise<void>
  onCancel: () => void
}

export function FotoCapaTab({ isLoading, erroBackend, onSubmit, onCancel }: FotoCapaTabProps) {
  const [busca, setBusca] = useState('')
  const [jogos, setJogos] = useState<JogoZeradoDTO[]>([])
  const [isBuscando, setIsBuscando] = useState(false)
  const [jogoSelecionado, setJogoSelecionado] = useState<JogoZeradoDTO | null>(null)

  useEffect(() => {
    let ativo = true
    const timer = setTimeout(async () => {
      setIsBuscando(true)
      try {
        const resp = await jogosService.listar({ busca: busca.trim() || undefined, por_pagina: 12 })
        if (ativo) setJogos(resp.data)
      } catch {
        if (ativo) setJogos([])
      } finally {
        if (ativo) setIsBuscando(false)
      }
    }, 300)

    return () => {
      ativo = false
      clearTimeout(timer)
    }
  }, [busca])

  return (
    <div className="space-y-4">
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--text-faint)]" aria-hidden="true" />
        <input
          type="search"
          value={busca}
          onChange={(e) => setBusca(e.target.value)}
          placeholder="Buscar nos meus jogos zerados…"
          aria-label="Buscar nos meus jogos zerados"
          className="h-10 w-full rounded-[7px] border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] pl-9 pr-3 text-xs text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:border-[var(--accent)] focus:outline-none"
        />
      </div>

      <div className="max-h-56 overflow-y-auto rounded-[8px] border border-[var(--border)] bg-[var(--bg-surface-alt)]/40 p-1 divide-y divide-[var(--border)]">
        {isBuscando ? (
          <p className="p-4 text-center text-xs text-[var(--text-muted)]">Buscando jogos…</p>
        ) : jogos.length === 0 ? (
          <p className="p-4 text-center text-xs text-[var(--text-muted)]">Nenhum jogo encontrado.</p>
        ) : (
          jogos.map((jogo) => {
            const temCapa = Boolean(jogo.igdb_capa_url)
            const isSelected = jogoSelecionado?.id === jogo.id
            return (
              <button
                key={jogo.id}
                type="button"
                disabled={!temCapa}
                onClick={() => setJogoSelecionado(jogo)}
                className={`flex w-full min-h-11 items-center gap-3 p-2 text-left transition-colors ${
                  isSelected ? 'bg-[var(--bg-surface-alt)] font-semibold text-[var(--accent)]' : 'hover:bg-[var(--bg-surface-alt)]'
                } ${!temCapa ? 'opacity-40 cursor-not-allowed' : 'cursor-pointer'}`}
              >
                {jogo.igdb_capa_url ? (
                  <img src={jogo.igdb_capa_url} alt="" className="h-8 w-8 rounded-full object-cover shrink-0 border border-[var(--border-subtle)]" />
                ) : (
                  <div className="flex h-8 w-8 items-center justify-center rounded-full bg-[var(--bg-surface)] text-[var(--text-muted)] border border-[var(--border-subtle)] shrink-0">
                    <ImageIcon className="h-4 w-4" aria-hidden="true" />
                  </div>
                )}
                <div className="min-w-0 flex-1">
                  <p className="truncate text-xs text-[var(--text-primary)]">{jogo.nome}</p>
                  {!temCapa && <span className="text-[10px] text-[var(--text-muted)]">Sem capa disponível</span>}
                </div>
                {isSelected && <Check className="h-4 w-4 shrink-0 text-[var(--accent)]" aria-hidden="true" />}
              </button>
            )
          })
        )}
      </div>

      {jogoSelecionado && (
        <div className="flex items-center gap-3 rounded-[8px] border border-[var(--border)] bg-[var(--bg-surface-alt)] p-2.5">
          <img src={jogoSelecionado.igdb_capa_url} alt="Capa selecionada" className="h-10 w-10 rounded-full object-cover border border-[var(--border)] shrink-0" />
          <div className="min-w-0 flex-1">
            <span className="text-[10px] uppercase font-bold tracking-wider text-[var(--accent)]">Selecionado</span>
            <p className="truncate text-xs font-semibold text-[var(--text-primary)]">{jogoSelecionado.nome}</p>
          </div>
        </div>
      )}

      {erroBackend && (
        <div role="alert" className="flex items-center gap-2 rounded-[8px] border border-[var(--danger)]/30 bg-[var(--danger)]/10 px-3 py-2 text-xs text-[var(--danger)]">
          <AlertCircle className="h-4 w-4 shrink-0" aria-hidden="true" />
          <span>{erroBackend}</span>
        </div>
      )}

      <div className="flex items-center justify-end gap-2 pt-2">
        <button
          type="button"
          onClick={onCancel}
          disabled={isLoading}
          className="min-h-11 rounded-[7px] border border-[var(--border-subtle)] px-4 py-2 text-xs font-semibold text-[var(--text-secondary)] transition-colors hover:text-[var(--text-primary)] disabled:opacity-50"
        >
          Cancelar
        </button>
        <button
          type="button"
          onClick={() => { if (jogoSelecionado) void onSubmit(jogoSelecionado.id) }}
          disabled={!jogoSelecionado || isLoading}
          className="min-h-11 rounded-[7px] bg-[var(--accent)] px-5 py-2 text-xs font-bold text-[#0E0F12] transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {isLoading ? 'Salvando…' : 'Usar como foto'}
        </button>
      </div>
    </div>
  )
}
