import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { NovaListaDialog } from "./NovaListaDialog";
import { ListasContext, type ListasStore } from "@/stores/listasStore";
import type { ListaResumo } from "@/types/listas";

function renderDialog(overrides: Partial<ListasStore> = {}) {
  const store = {
    listas: [],
    listaAberta: null,
    isLoading: false,
    isLoadingDetalhe: false,
    error: null,
    filtroAba: "todos" as const,
    isNovaListaOpen: true,
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
    criarLista: vi.fn().mockResolvedValue({ id: 5 }),
    atualizarLista: vi.fn(),
    excluirLista: vi.fn(),
    adicionarItem: vi.fn(),
    adicionarItensLote: vi.fn(),
    removerItem: vi.fn(),
    reordenarItens: vi.fn(),
    associarZeramento: vi.fn(),
    ...overrides,
  } as ListasStore;
  return {
    store,
    ...render(
      <MemoryRouter>
        <ListasContext.Provider value={store}>
          <NovaListaDialog />
        </ListasContext.Provider>
      </MemoryRouter>,
    ),
  };
}

describe("NovaListaDialog", () => {
  it("exibe fila por padrão e alterna para desafio", async () => {
    const user = userEvent.setup();
    renderDialog();
    expect(
      screen.getByRole("button", { name: "Criar fila" }),
    ).toBeInTheDocument();
    expect(screen.getByRole("dialog")).toHaveClass("sm:max-w-none");
    await user.click(screen.getByRole("radio", { name: /Desafio/ }));
    expect(screen.getByText("De onde vêm os jogos?")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Ver jogos" }),
    ).toBeInTheDocument();
  });
  it("exibe as três origens e a ajuda correspondente", async () => {
    const user = userEvent.setup();
    renderDialog();
    await user.click(screen.getByRole("radio", { name: /Desafio/ }));
    expect(screen.getAllByRole("radio")).toHaveLength(5);
    expect(screen.getByText(/todos os jogos da franquia/)).toBeInTheDocument();
    await user.click(screen.getByRole("radio", { name: "Plataforma" }));
    expect(screen.getByText(/jogos da plataforma/)).toBeInTheDocument();
    await user.click(screen.getByRole("radio", { name: "Genero" }));
    expect(screen.getByText(/jogos do gênero/)).toBeInTheDocument();
  });
  it("mostra erro de nome obrigatório por campo", async () => {
    const user = userEvent.setup();
    renderDialog();
    await user.click(screen.getByRole("button", { name: "Criar fila" }));
    expect(await screen.findByText("O nome da lista é obrigatório.")).toBeInTheDocument();
  });
  it("valida origem escolhida antes de abrir o catálogo", async () => {
    const user = userEvent.setup();
    renderDialog();
    await user.click(screen.getByRole("radio", { name: /Desafio/ }));
    await user.type(screen.getByLabelText("Nome *"), "Desafio");
    await user.click(screen.getByRole("button", { name: "Ver jogos" }));
    expect(
      await screen.findByText("Escolha uma opção da lista."),
    ).toBeInTheDocument();
  });
  it("cria fila sem meta ou regra", async () => {
    const user = userEvent.setup();
    const { store } = renderDialog();
    await user.type(screen.getByLabelText("Nome *"), "Minha fila");
    await user.click(screen.getByRole("button", { name: "Criar fila" }));
    expect(store.criarLista).toHaveBeenCalledWith({
      tipo: "fila",
      nome: "Minha fila",
      descricao: null,
    });
  });
  it("modo edição exibe apenas nome e descrição", () => {
    const lista: ListaResumo = {
      id: 1,
      tipo: "desafio",
      nome: "Desafio",
      descricao: null,
      origem: { tipo: "franquia", igdb_id: 1, nome: "Zelda" },
      total_itens: 1,
      itens_pendentes: 1,
      progresso: null,
      created_at: "",
      updated_at: "",
    };
    renderDialog({ listaEmEdicao: lista });
    expect(
      screen.getByRole("heading", { name: "Editar desafio" }),
    ).toBeInTheDocument();
    expect(screen.queryByText("De onde vêm os jogos?")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Salvar" })).toBeInTheDocument();
  });
});
