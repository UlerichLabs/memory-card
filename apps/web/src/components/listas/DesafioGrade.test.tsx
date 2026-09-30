import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { DesafioGrade } from "./DesafioGrade";
import { ListasContext, type ListasStore } from "@/stores/listasStore";
import { JogosContext, type JogosStore } from "@/stores/jogosStore";
import type { ListaDetalhada } from "@/types/listas";
import type { JogoZeradoDTO } from "@/lib/services/jogosService";

const lista: ListaDetalhada = {
  id: 1,
  tipo: "desafio",
  nome: "Zelda",
  descricao: null,
  origem: { tipo: "franquia", igdb_id: 596, nome: "Zelda" },
  total_itens: 2,
  itens_pendentes: 1,
  progresso: {
    feitos: 1,
    meta: 2,
    percentual: 50,
    concluido: false,
    concluido_em: null,
  },
  created_at: "",
  updated_at: "",
  itens: [
    {
      id: 1,
      igdb_id: 10,
      nome: "Zelda",
      console: null,
      igdb_capa_url: null,
      ano_lancamento: 1986,
      posicao: 1,
      origem: "item",
      zerado: false,
      jogo_zerado: null,
    },
    {
      id: 2,
      igdb_id: 11,
      nome: "Zelda II",
      console: null,
      igdb_capa_url: null,
      ano_lancamento: 1987,
      posicao: 2,
      origem: "item",
      zerado: true,
      jogo_zerado: null,
    },
  ],
};

function renderGrade(overrides: Partial<ListasStore> = {}, listaAtual = lista) {
  const store = {
    listas: [],
    listaAberta: listaAtual,
    isLoading: false,
    isLoadingDetalhe: false,
    error: null,
    filtroAba: "todos" as const,
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
    ...overrides,
  } as ListasStore;
  const jogosStore = {
    abrirModalRegistro: vi.fn(),
  } as unknown as JogosStore;
  return {
    store,
    jogosStore,
    ...render(
      <MemoryRouter>
        <ListasContext.Provider value={store}>
          <JogosContext.Provider value={jogosStore}>
            <DesafioGrade lista={listaAtual} />
          </JogosContext.Provider>
        </ListasContext.Provider>
      </MemoryRouter>,
    ),
  };
}

describe("DesafioGrade", () => {
  it("renderiza abas e o botão de adicionar jogos quando existe origem", () => {
    renderGrade();
    expect(screen.getByRole("tab", { name: /Todos · 2/ })).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /Adicionar jogos/ }),
    ).toBeInTheDocument();
    expect(screen.queryByText(/Fora do desafio/)).not.toBeInTheDocument();
  });
  it("permite remover qualquer card", async () => {
    const { store } = renderGrade();
    const botoes = screen.getAllByRole("button", { name: "Tirar do desafio" });
    expect(botoes).toHaveLength(2);
    await botoes[0].click();
    expect(store.removerItem).toHaveBeenCalledWith(1);
  });
  it("exibe contadores, nota e iniciais nos cards", () => {
    renderGrade();
    expect(screen.getByRole("tab", { name: /Zerados · 1/ })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /Pendentes · 1/ })).toBeInTheDocument();
    expect(screen.getByText("ZI")).toBeInTheDocument();
  });
  it("não exibe adicionar jogos sem origem", () => {
    renderGrade({}, { ...lista, origem: null });
    expect(screen.queryByRole("button", { name: /Adicionar jogos/ })).not.toBeInTheDocument();
  });

  it("abre o registro e associa o zeramento no card pendente", async () => {
    const { jogosStore, store } = renderGrade();
    const botoesZerei = screen.getAllByRole("button", { name: "Zerei!" });
    await botoesZerei[0].click();

    expect(jogosStore.abrirModalRegistro).toHaveBeenCalledWith(
      expect.objectContaining({
        valoresIniciais: {
          nome: "Zelda",
          igdb_id: 10,
          igdb_capa_url: undefined,
          console: undefined,
        },
      }),
    );

    const opcoes = vi.mocked(jogosStore.abrirModalRegistro).mock.calls[0][0];
    const jogo: JogoZeradoDTO = {
      id: 99,
      usuario_id: 1,
      nome: "Zelda",
      console: "SNES",
      genero: "RPG",
      finalizado_em: "2026-09-30",
      tempo_jogado: 3600,
      nota: 10,
      dificuldade: "A",
      destaque: false,
    };

    await opcoes?.onSalvo?.(jogo);
    expect(store.associarZeramento).toHaveBeenCalledWith(1, 99);
    expect(screen.queryAllByRole("button", { name: "Zerei!" })).toHaveLength(2);
  });

  it("mostra erro quando não consegue ligar o zeramento ao desafio", async () => {
    const associarZeramento = vi.fn().mockRejectedValue(new Error("falha"));
    const { jogosStore } = renderGrade({ associarZeramento });
    await screen.getAllByRole("button", { name: "Zerei!" })[0].click();

    const opcoes = vi.mocked(jogosStore.abrirModalRegistro).mock.calls[0][0];
    await opcoes?.onSalvo?.({ id: 99 } as JogoZeradoDTO);

    expect(
      await screen.findByText(
        "O jogo foi registrado, mas não foi possível ligar ao desafio. Tente de novo.",
      ),
    ).toBeInTheDocument();
  });
});
