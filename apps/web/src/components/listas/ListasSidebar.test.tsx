import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { ListasSidebar } from "./ListasSidebar";
import type { ListaResumo } from "@/types/listas";

const listas: ListaResumo[] = [
  {
    id: 1,
    tipo: "desafio",
    nome: "Zelda",
    descricao: null,
    origem: { tipo: "franquia", igdb_id: 596, nome: "Zelda" },
    total_itens: 3,
    itens_pendentes: 2,
    progresso: { feitos: 1, meta: 3, percentual: 33, concluido: false, concluido_em: null },
    created_at: "",
    updated_at: "",
  },
  {
    id: 2,
    tipo: "fila",
    nome: "Próximos",
    descricao: null,
    origem: null,
    total_itens: 2,
    itens_pendentes: 1,
    progresso: null,
    created_at: "",
    updated_at: "",
  },
];

describe("ListasSidebar", () => {
  it("agrupa desafio e fila e exibe o subtítulo da origem", () => {
    render(<ListasSidebar listas={listas} selectedId={1} onSelect={vi.fn()} onNovaLista={vi.fn()} />);
    expect(screen.getByText("Desafios")).toBeInTheDocument();
    expect(screen.getByText("Filas")).toBeInTheDocument();
    expect(screen.getByText("3 jogos · Franquia")).toBeInTheDocument();
    expect(screen.getByText("1 jogo na fila")).toBeInTheDocument();
  });

  it("mantém três itens visíveis e permite expandir cada grupo", async () => {
    const user = userEvent.setup();
    const listasGrandes = Array.from({ length: 4 }, (_, index) => ({
      ...listas[0],
      id: index + 10,
      nome: `Desafio ${index + 1}`,
    })).concat(Array.from({ length: 4 }, (_, index) => ({
      ...listas[1],
      id: index + 20,
      nome: `Fila ${index + 1}`,
    })));

    render(<ListasSidebar listas={listasGrandes} selectedId={10} onSelect={vi.fn()} onNovaLista={vi.fn()} />);

    expect(screen.getByText("Desafio 1")).toBeInTheDocument();
    expect(screen.queryByText("Desafio 4")).not.toBeInTheDocument();
    expect(screen.getByText("Fila 1")).toBeInTheDocument();
    expect(screen.queryByText("Fila 4")).not.toBeInTheDocument();

    await user.click(screen.getAllByRole("button", { name: /Mostrar mais \(1\)/ })[0]);
    expect(screen.getByText("Desafio 4")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /Mostrar mais \(1\)/ }));
    expect(screen.getByText("Fila 4")).toBeInTheDocument();
  });
});
