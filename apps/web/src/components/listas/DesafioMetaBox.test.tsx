import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { DesafioMetaBox } from './DesafioMetaBox'
import type { ListaDetalhada } from '@/types/listas'

const baseLista: ListaDetalhada = {
  id: 1,
  tipo: 'desafio',
  nome: 'Desafio SNES',
  descricao: null,
  regra: { tipo: 'plataforma', valor: 'SNES', igdb_id: null },
  meta: 5,
  total_itens: 3,
  itens_pendentes: 0,
  progresso: {
    feitos: 3,
    meta: 5,
    percentual: 60,
    concluido: false,
    concluido_em: null,
  },
  created_at: '2026-09-01T00:00:00Z',
  updated_at: '2026-09-01T00:00:00Z',
  itens: [],
}

describe('DesafioMetaBox', () => {
  it('renderiza em andamento com Faltam N e percentual', () => {
    render(<DesafioMetaBox lista={baseLista} />)
    expect(screen.getByText('3')).toBeInTheDocument()
    expect(screen.getByText('/ 5')).toBeInTheDocument()
    expect(screen.getByText('jogos zerados na plataforma')).toBeInTheDocument()
    expect(screen.getByText('60%')).toBeInTheDocument()
    expect(screen.getByText(/Faltam 2 · o progresso conta sozinho/i)).toBeInTheDocument()
  })

  it('renderiza singular Falta 1 quando falta 1 jogo', () => {
    const listaFalta1: ListaDetalhada = {
      ...baseLista,
      progresso: {
        feitos: 4,
        meta: 5,
        percentual: 80,
        concluido: false,
        concluido_em: null,
      },
    }
    render(<DesafioMetaBox lista={listaFalta1} />)
    expect(screen.getByText(/Falta 1 · o progresso conta sozinho/i)).toBeInTheDocument()
  })

  it('renderiza concluído com data formatada e estilo ouro', () => {
    const listaConcluida: ListaDetalhada = {
      ...baseLista,
      progresso: {
        feitos: 5,
        meta: 5,
        percentual: 100,
        concluido: true,
        concluido_em: '2026-09-15T12:00:00Z',
      },
    }
    render(<DesafioMetaBox lista={listaConcluida} />)
    expect(
      screen.getByText(/Desafio concluído em 15\/09\/2026 — conquista desbloqueada/i)
    ).toBeInTheDocument()
  })

  it('retorna null se não houver progresso', () => {
    const semProgresso: ListaDetalhada = { ...baseLista, progresso: null }
    const { container } = render(<DesafioMetaBox lista={semProgresso} />)
    expect(container.firstChild).toBeNull()
  })
})
