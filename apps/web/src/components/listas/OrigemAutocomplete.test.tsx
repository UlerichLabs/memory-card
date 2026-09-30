import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { ApiError } from "@/lib/api";
import { listasService } from "@/lib/services/listasService";
import type { ListaOrigem } from "@/types/listas";
import { OrigemAutocomplete } from "./OrigemAutocomplete";

const origem: ListaOrigem = { tipo: "franquia", igdb_id: 596, nome: "Zelda" };

describe("OrigemAutocomplete", () => {
  it("exibe erro quando a busca remota falha", async () => {
    vi.spyOn(listasService, "buscarFranquias").mockRejectedValue(
      new ApiError("igdb.unavailable", ""),
    );
    const user = userEvent.setup();
    render(<OrigemAutocomplete tipo="franquia" valor={null} onChange={vi.fn()} />);

    await user.type(screen.getByRole("combobox"), "ab");
    expect(
      await screen.findByText("Não foi possível buscar agora. Tente de novo."),
    ).toBeInTheDocument();
  });

  it("mantém o texto digitado depois de editar uma origem escolhida", async () => {
    let valor: ListaOrigem | null = origem;
    const { rerender } = render(
      <OrigemAutocomplete
        tipo="franquia"
        valor={valor}
        onChange={(novoValor) => {
          valor = novoValor;
          rerender(
            <OrigemAutocomplete
              tipo="franquia"
              valor={valor}
              onChange={(proximoValor) => {
                valor = proximoValor;
              }}
            />,
          );
        }}
      />,
    );
    const user = userEvent.setup();
    const input = screen.getByRole("combobox");
    await user.clear(input);
    await user.type(input, "Novo nome");
    expect(input).toHaveValue("Novo nome");
  });

  it("exibe Gênero com acento no rótulo", () => {
    render(<OrigemAutocomplete tipo="genero" valor={null} onChange={vi.fn()} />);
    expect(screen.getByText("Gênero *")).toBeInTheDocument();
  });
});
