import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { DesafioMetaBox } from "./DesafioMetaBox";
import type { ListaDetalhada } from "@/types/listas";

const lista: ListaDetalhada = {
  id: 1,
  tipo: "desafio",
  nome: "Desafio",
  descricao: null,
  origem: null,
  total_itens: 5,
  itens_pendentes: 2,
  progresso: {
    feitos: 3,
    meta: 5,
    percentual: 60,
    concluido: false,
    concluido_em: null,
  },
  created_at: "",
  updated_at: "",
  itens: [],
};

describe("DesafioMetaBox", () => {
  it("exibe jogos zerados e a distância para concluir", () => {
    render(<DesafioMetaBox lista={lista} />);
    expect(screen.getByText("jogos zerados")).toBeInTheDocument();
    expect(screen.getByText(/Faltam 2 · zere e registre/)).toBeInTheDocument();
  });
  it("exibe a conquista quando concluído", () => {
    render(
      <DesafioMetaBox
        lista={{
          ...lista,
          progresso: {
            feitos: 5,
            meta: 5,
            percentual: 100,
            concluido: true,
            concluido_em: "2026-09-15T12:00:00Z",
          },
        }}
      />,
    );
    expect(
      screen.getByText(/Desafio concluído em 15\/09\/2026/),
    ).toBeInTheDocument();
  });
  it("não renderiza sem progresso", () => {
    const { container } = render(<DesafioMetaBox lista={{ ...lista, progresso: null }} />);
    expect(container.firstChild).toBeNull();
  });
});
