import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import type { ComponentProps } from "react";
import { describe, expect, it, vi } from "vitest";
import { ListasContext, type ListasStore } from "@/stores/listasStore";
import { EscolherJogosDialog } from "./EscolherJogosDialog";

const origem = { tipo: "franquia" as const, igdb_id: 596, nome: "Zelda" };
const jogo = {
  igdb_id: 1,
  nome: "Ocarina",
  igdb_capa_url: null,
  ano_lancamento: 1998,
  sugerido: true,
  ja_zerado: false,
  jogo_zerado_id: null,
};

vi.mock("./useEscolherJogos", () => ({
  useEscolherJogos: () => ({
    token: undefined,
    itens: [jogo],
    selecionados: new Map([[1, { igdb_id: 1, nome: "Ocarina", ano_lancamento: 1998 }]]),
    payload: [{ igdb_id: 1, nome: "Ocarina", ano_lancamento: 1998 }],
    idsExistentes: new Set<number>(),
    meta: { pagina: 1, por_pagina: 60, total: 1, total_sugeridos: 1, total_todos: 1 },
    carregando: false,
    erro: null,
    busca: "",
    generoId: undefined,
    plataformaId: undefined,
    ordenar: "populares",
    setBusca: vi.fn(),
    setGeneroId: vi.fn(),
    setPlataformaId: vi.fn(),
    setOrdenar: vi.fn(),
    alternar: vi.fn(),
    desmarcar: vi.fn(),
    carregarMais: vi.fn(),
    recarregar: vi.fn(),
    marcarSugeridos: vi.fn().mockResolvedValue(undefined),
    somenteSugeridos: true,
    setSomenteSugeridos: vi.fn(),
  }),
}));

function renderDialog(
  overrides: Partial<ListasStore> = {},
  props: Partial<ComponentProps<typeof EscolherJogosDialog>> = {},
) {
  const store = {
    criarLista: vi.fn().mockResolvedValue({ id: 9 }),
    adicionarItensLote: vi.fn().mockResolvedValue({ adicionados: 1 }),
    ...overrides,
  } as unknown as ListasStore;
  return render(
    <MemoryRouter>
      <ListasContext.Provider value={store}>
        <EscolherJogosDialog
          open
          modo="criar"
          config={{ nome: "Desafio", descricao: null, origem }}
          onClose={vi.fn()}
          {...props}
        />
      </ListasContext.Provider>
    </MemoryRouter>,
  );
}

describe("EscolherJogosDialog", () => {
  it("renderiza o modo criar com contador e meta", () => {
    renderDialog();
    expect(screen.getByText(/meta do desafio: zerar 1/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Criar desafio" })).toBeInTheDocument();
  });

  it("renderiza o modo adicionar com meta acumulada", () => {
    renderDialog(
      {},
      {
        modo: "adicionar",
        config: undefined,
        lista: { origem, nome: "Atual", itens: [], id: 1 } as never,
      },
    );
    expect(screen.getByText(/meta passa para 1/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Adicionar 1 jogos" })).toBeInTheDocument();
  });

  it("mostra sugeridos somente para franquia", async () => {
    renderDialog();
    expect(screen.getByRole("button", { name: "Marcar os 1 sugeridos" })).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /Mostrar outros jogos ligados/ }),
    ).not.toBeInTheDocument();
    const user = userEvent.setup();
    await user.click(screen.getByText("Ver e revisar seleção"));
    expect(screen.getByRole("button", { name: "Desmarcar Ocarina" })).toBeInTheDocument();
  });

  it("não renderiza catálogo com origem inválida", () => {
    renderDialog(
      {},
      { config: { nome: "Inválido", descricao: null, origem: { ...origem, igdb_id: null } } },
    );
    expect(screen.getByText("Origem inválida.")).toBeInTheDocument();
    expect(screen.queryByText("Marcar visíveis")).not.toBeInTheDocument();
  });
});
