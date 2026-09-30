import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { FilaLista } from "./FilaLista";
import { ListasContext, type ListasStore } from "@/stores/listasStore";
import { JogosContext, type JogosStore } from "@/stores/jogosStore";
import { ApiError } from "@/lib/api";
import type { ListaDetalhada } from "@/types/listas";
import type { JogoZeradoDTO } from "@/lib/services/jogosService";

const mockFilaLista: ListaDetalhada = {
  id: 1,
  tipo: "fila",
  nome: "Minha Fila",
  descricao: null,
  origem: null,
  total_itens: 3,
  itens_pendentes: 2,
  progresso: null,
  created_at: "2026-09-01T00:00:00Z",
  updated_at: "2026-09-01T00:00:00Z",
  itens: [
    {
      id: 10,
      igdb_id: 100,
      nome: "Chrono Trigger",
      console: "SNES",
      igdb_capa_url: null,
      ano_lancamento: 1995,
      posicao: 1,
      origem: "item",
      zerado: false,
      jogo_zerado: null,
    },
    {
      id: 11,
      igdb_id: 101,
      nome: "Final Fantasy VI",
      console: "SNES",
      igdb_capa_url: null,
      ano_lancamento: 1994,
      posicao: 2,
      origem: "item",
      zerado: false,
      jogo_zerado: null,
    },
    {
      id: 12,
      igdb_id: 102,
      nome: "Super Mario World",
      console: "SNES",
      igdb_capa_url: null,
      ano_lancamento: 1990,
      posicao: 3,
      origem: "item",
      zerado: true,
      jogo_zerado: { id: 80, nota: 10, finalizado_em: "2026-01-01T00:00:00Z" },
    },
  ],
};

function renderFilaLista(
  storeOverrides: Partial<ListasStore> = {},
  jogosOverrides: Partial<JogosStore> = {},
) {
  const store: ListasStore = {
    listas: [],
    listaAberta: mockFilaLista,
    isLoading: false,
    isLoadingDetalhe: false,
    error: null,
    filtroAba: "todos",
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
    adicionarItensLote: vi.fn(),
    removerItem: vi.fn(),
    reordenarItens: vi.fn(),
    associarZeramento: vi.fn(),
    ...storeOverrides,
  };

  const jogosStore: JogosStore = {
    jogos: [],
    meta: { pagina: 1, por_pagina: 24, total: 0, total_paginas: 0 },
    filtros: { consoles: [], generos: [], tipos: [], anos: [] },
    isLoading: false,
    error: null,
    isModalOpen: false,
    jogoEmEdicao: null,
    modalRegistroOpcoes: null,
    abrirModalRegistro: vi.fn(),
    abrirModalEdicao: vi.fn(),
    fecharModal: vi.fn(),
    carregarJogos: vi.fn(),
    carregarFiltros: vi.fn(),
    limparBiblioteca: vi.fn(),
    criarJogo: vi.fn(),
    atualizarJogo: vi.fn(),
    excluirJogo: vi.fn(),
    buscarIGDB: vi.fn(),
    obterDetalhesIGDB: vi.fn(),
    obterJogoPorId: vi.fn(),
    setJogos: vi.fn(),
    limparErro: vi.fn(),
    ...jogosOverrides,
  };

  return {
    ...render(
      <ListasContext.Provider value={store}>
        <JogosContext.Provider value={jogosStore}>
          <FilaLista lista={mockFilaLista} />
        </JogosContext.Provider>
      </ListasContext.Provider>,
    ),
    store,
    jogosStore,
  };
}

describe("FilaLista", () => {
  it("renderiza apenas itens pendentes em A jogar e marca o primeiro como Próximo", () => {
    renderFilaLista();
    expect(screen.getByText("Chrono Trigger")).toBeInTheDocument();
    expect(screen.getByText("Final Fantasy VI")).toBeInTheDocument();
    expect(screen.queryByText("Super Mario World")).not.toBeInTheDocument();
    expect(screen.getByText("Próximo")).toBeInTheDocument();
    expect(screen.getByText("2 jogos")).toBeInTheDocument();
  });

  it("clicar em Zerei! abre modal de registro com valores iniciais e vincula no onSalvo", async () => {
    const user = userEvent.setup();
    const { store, jogosStore } = renderFilaLista();

    const botoesZerei = screen.getAllByRole("button", { name: "Zerei!" });
    await user.click(botoesZerei[0]);

    expect(jogosStore.abrirModalRegistro).toHaveBeenCalledTimes(1);
    const callArg = vi.mocked(jogosStore.abrirModalRegistro).mock.calls[0][0];
    expect(callArg?.valoresIniciais).toEqual({
      nome: "Chrono Trigger",
      igdb_id: 100,
      igdb_capa_url: undefined,
      console: "SNES",
    });

    const jogoSalvo: JogoZeradoDTO = {
      id: 99,
      usuario_id: 1,
      nome: "Chrono Trigger",
      console: "SNES",
      genero: "RPG",
      finalizado_em: "2026-09-29",
      tempo_jogado: 3600,
      nota: 10,
      dificuldade: "A",
      destaque: false,
    };

    if (callArg?.onSalvo) {
      await callArg.onSalvo(jogoSalvo);
    }

    expect(store.associarZeramento).toHaveBeenCalledWith(10, 99);
  });

  it("exibe erro 409 inline abaixo do input de adicionar", async () => {
    const user = userEvent.setup();
    const adicionarItemMock = vi
      .fn()
      .mockRejectedValue(
        new ApiError("listas.item_duplicado", "Jogo duplicado", 409),
      );
    renderFilaLista({ adicionarItem: adicionarItemMock });

    const input = screen.getByRole("textbox", {
      name: /Adicionar jogo à fila/i,
    });
    await user.type(input, "Chrono Trigger{Enter}");

    expect(
      await screen.findByText("Esse jogo já está na lista."),
    ).toBeInTheDocument();
  });
});
