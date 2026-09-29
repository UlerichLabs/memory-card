import { describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { DesafioGrade } from './DesafioGrade'
import { ListasContext, type ListasStore } from '@/stores/listasStore'
import type { ListaDetalhada } from '@/types/listas'

const mockFranquiaLista: ListaDetalhada = {
  id: 1,
  tipo: 'desafio',
  nome: 'Franquia Zelda',
  descricao: null,
  regra: { tipo: 'franquia', valor: 'The Legend of Zelda', igdb_id: 10 },
  meta: 3,
  total_itens: 3,
  itens_pendentes: 1,
  progresso: { feitos: 2, meta: 3, percentual: 66, concluido: false, concluido_em: null },
  created_at: '2026-09-01T00:00:00Z',
  updated_at: '2026-09-01T00:00:00Z',
  itens: [
    {
      id: 101,
      igdb_id: 1022,
      nome: 'The Legend of Zelda',
      console: 'NES',
      igdb_capa_url: '//images.igdb.com/capa1.jpg',
      ano_lancamento: 1986,
      posicao: 1,
      origem: 'item',
      zerado: true,
      jogo_zerado: { id: 50, nota: 10, finalizado_em: '2026-01-01T00:00:00Z' },
    },
    {
      id: 102,
      igdb_id: 1023,
      nome: 'Zelda II: The Adventure of Link',
      console: 'NES',
      igdb_capa_url: null,
      ano_lancamento: 1987,
      posicao: 2,
      origem: 'item',
      zerado: false,
      jogo_zerado: null,
    },
  ],
}

function renderDesafioGrade(lista: ListaDetalhada, storeOverrides: Partial<ListasStore> = {}) {
  const store: ListasStore = {
    listas: [],
    listaAberta: lista,
    isLoading: false,
    isLoadingDetalhe: false,
    error: null,
    filtroAba: 'todos',
    isNovaListaOpen: false,
    listaEmEdicao: null,
    isExcluirListaOpen: false,
    listaParaExcluir: null,
    setFiltroAba: vi.fn(),
    limparErro: vi.fn(),
    abrirModalCriar: vi.fn(),
    abrirModalEditar: vi.fn(),
    fecharModalNovaLista: vi.fn(),
    abrirModalExcluir: vi.fn(),
    fecharModalExcluir: vi.fn(),
    carregarListas: vi.fn(),
    abrirLista: vi.fn(),
    criarLista: vi.fn(),
    atualizarLista: vi.fn(),
    excluirLista: vi.fn(),
    adicionarItem: vi.fn(),
    removerItem: vi.fn(),
    reordenarItens: vi.fn(),
    associarZeramento: vi.fn(),
    sincronizarFranquia: vi.fn().mockResolvedValue({ ...lista, adicionados: 1 }),
    ...storeOverrides,
  }

  return {
    ...render(
      <MemoryRouter>
        <ListasContext.Provider value={store}>
          <DesafioGrade lista={lista} />
        </ListasContext.Provider>
      </MemoryRouter>
    ),
    store,
  }
}

describe('DesafioGrade', () => {
  it('renderiza abas com contadores e botão Atualizar do IGDB em franquia', () => {
    renderDesafioGrade(mockFranquiaLista)
    expect(screen.getByRole('tab', { name: /Todos · 2/i })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: /Zerados · 1/i })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: /Pendentes · 1/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Atualizar do IGDB/i })).toBeInTheDocument()
  })

  it('card zerado tem NotaBadge e link para /biblioteca/:id', () => {
    renderDesafioGrade(mockFranquiaLista)
    expect(screen.getByText('The Legend of Zelda')).toBeInTheDocument()
    expect(screen.getByLabelText('Nota 10')).toBeInTheDocument()

    const link = screen.getByRole('link')
    expect(link).toHaveAttribute('href', '/biblioteca/50')
  })

  it('card pendente exibe badge Pendente e iniciais quando sem capa', () => {
    renderDesafioGrade(mockFranquiaLista)
    expect(screen.getByText('Zelda II: The Adventure of Link')).toBeInTheDocument()
    expect(screen.getByText('Pendente')).toBeInTheDocument()
    expect(screen.getByText('ZI')).toBeInTheDocument()
  })

  it('contagem não exibe barra de filtros e exibe card +N para completar', () => {
    const contagemLista: ListaDetalhada = {
      ...mockFranquiaLista,
      regra: { tipo: 'plataforma', valor: 'SNES', igdb_id: null },
      meta: 5,
      progresso: { feitos: 2, meta: 5, percentual: 40, concluido: false, concluido_em: null },
    }
    renderDesafioGrade(contagemLista)
    expect(screen.queryByRole('tablist')).not.toBeInTheDocument()
    expect(screen.getByText('+3')).toBeInTheDocument()
    expect(screen.getByText('para completar')).toBeInTheDocument()
  })

  it('chama sincronizarFranquia ao clicar em Atualizar do IGDB', async () => {
    const user = userEvent.setup()
    const { store } = renderDesafioGrade(mockFranquiaLista)

    const btn = screen.getByRole('button', { name: /Atualizar do IGDB/i })
    await user.click(btn)

    expect(store.sincronizarFranquia).toHaveBeenCalledTimes(1)
    expect(await screen.findByText('1 jogos novos adicionados')).toBeInTheDocument()
  })
})
