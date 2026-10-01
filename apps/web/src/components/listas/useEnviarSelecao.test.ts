import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { CriarListaItemPayload } from "@/types/listas";
import { useEnviarSelecao } from "./useEnviarSelecao";

const origem = { tipo: "franquia" as const, igdb_id: 596, nome: "Zelda" };
const itens: CriarListaItemPayload[] = [{ igdb_id: 1, nome: "Ocarina" }];

describe("useEnviarSelecao", () => {
  it("ignora o segundo clique enquanto o primeiro está enviando", async () => {
    let resolver: (valor: { id: number }) => void = () => undefined;
    const criarLista = vi.fn(
      () =>
        new Promise<{ id: number }>((resolve) => {
          resolver = resolve;
        }),
    );
    const store = {
      criarLista,
      carregarListas: vi.fn().mockResolvedValue([]),
      adicionarItensLote: vi.fn(),
    };
    const onClose = vi.fn();
    const { result } = renderHook(() =>
      useEnviarSelecao({
        modo: "criar",
        config: { nome: "Desafio", descricao: null, origem },
        itens,
        store,
        navigate: vi.fn(),
        onClose,
      }),
    );

    act(() => {
      void result.current.enviar();
      void result.current.enviar();
    });
    expect(criarLista).toHaveBeenCalledTimes(1);

    await act(async () => {
      resolver({ id: 9 });
    });
    await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    expect(result.current.enviando).toBe(true);
  });
});
