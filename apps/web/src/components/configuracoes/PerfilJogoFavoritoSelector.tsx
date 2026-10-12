import { useEffect, useState } from 'react'
import { Search, X } from 'lucide-react'
import { jogosService, type JogoZeradoDTO } from '@/lib/services/jogosService'
import type { JogoFavoritoDTO } from '@/types/perfil'

interface PerfilJogoFavoritoSelectorProps {
  jogoFavorito: JogoFavoritoDTO | null
  erro?: string
  onSelect: (jogo: JogoFavoritoDTO | null) => void
}

export function PerfilJogoFavoritoSelector({ jogoFavorito, erro, onSelect }: PerfilJogoFavoritoSelectorProps) {
  const [busca, setBusca] = useState('')
  const [resultados, setResultados] = useState<JogoZeradoDTO[]>([])
  const [buscando, setBuscando] = useState(false)
  const [aberto, setAberto] = useState(false)

  useEffect(() => {
    let ativo = true
    const timer = setTimeout(async () => {
      if (!busca.trim()) {
        if (ativo) setResultados([])
        return
      }
      setBuscando(true)
      try {
        const resp = await jogosService.listar({ busca: busca.trim(), por_pagina: 8 })
        if (ativo) setResultados(resp.data)
      } catch {
        if (ativo) setResultados([])
      } finally {
        if (ativo) setBuscando(false)
      }
    }, 300)

    return () => {
      ativo = false
      clearTimeout(timer)
    }
  }, [busca])


  return (
    <div className="space-y-1.5">
      <label className="text-xs font-semibold text-[var(--text-secondary)]">Jogo favorito</label>
      {jogoFavorito ? (
        <div className="flex min-h-11 items-center justify-between rounded-[7px] border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-3 py-2">
          <div className="flex items-center gap-2.5 min-w-0">
            {jogoFavorito.capa_url && (
              <img src={jogoFavorito.capa_url} alt="" className="h-7 w-7 rounded-[4px] object-cover shrink-0" />
            )}
            <span className="truncate text-xs font-medium text-[var(--text-primary)]">{jogoFavorito.nome}</span>
          </div>
          <button
            type="button"
            onClick={() => onSelect(null)}
            aria-label="Remover jogo favorito"
            className="flex h-8 w-8 items-center justify-center rounded-[4px] text-[var(--text-secondary)] hover:text-[var(--danger)]"
          >
            <X className="h-4 w-4" aria-hidden="true" />
          </button>
        </div>
      ) : (
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--text-faint)]" aria-hidden="true" />
          <input
            type="text"
            value={busca}
            onChange={(e) => { setBusca(e.target.value); setAberto(true) }}
            onFocus={() => setAberto(true)}
            placeholder="Buscar entre seus jogos zerados…"
            className="h-10 w-full rounded-[7px] border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] pl-9 pr-3 text-xs text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:border-[var(--accent)] focus:outline-none"
          />
          {aberto && busca.trim() && (
            <ul className="absolute top-full left-0 right-0 z-20 mt-1 max-h-48 overflow-y-auto rounded-[8px] border border-[var(--border)] bg-[var(--bg-surface)] p-1 shadow-xl">
              {buscando ? (
                <li className="p-3 text-center text-xs text-[var(--text-muted)]">Buscando…</li>
              ) : resultados.length === 0 ? (
                <li className="p-3 text-center text-xs text-[var(--text-muted)]">Nenhum jogo encontrado</li>
              ) : (
                resultados.map((j) => (
                  <li key={j.id}>
                    <button
                      type="button"
                      onClick={() => {
                        onSelect({ id: j.id, nome: j.nome, capa_url: j.igdb_capa_url ?? null })
                        setBusca('')
                        setAberto(false)
                      }}
                      className="flex min-h-11 w-full items-center gap-2.5 rounded-[6px] px-2.5 py-2 text-left text-xs hover:bg-[var(--bg-surface-alt)]"
                    >
                      {j.igdb_capa_url && (
                        <img src={j.igdb_capa_url} alt="" className="h-6 w-6 rounded-[3px] object-cover shrink-0" />
                      )}
                      <span className="truncate text-[var(--text-primary)]">{j.nome}</span>
                    </button>
                  </li>
                ))
              )}
            </ul>
          )}
        </div>
      )}
      <p className="text-[11px] text-[var(--text-muted)]">Entre os jogos que você zerou.</p>
      {erro && <p className="text-xs text-[var(--danger)]">{erro}</p>}
    </div>
  )
}
