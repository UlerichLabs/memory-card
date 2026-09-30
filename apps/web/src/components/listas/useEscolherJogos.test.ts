import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ApiError } from "@/lib/api";
import { listasService } from "@/lib/services/listasService";
import type { CatalogoResposta, ListaOrigem } from "@/types/listas";
import { useEscolherJogos } from "./useEscolherJogos";

const franquia: ListaOrigem = { tipo: "franquia", igdb_id: 596, nome: "Zelda" };
const plataforma: ListaOrigem = { tipo: "plataforma", igdb_id: 130, nome: "SNES" };
const genero: ListaOrigem = { tipo: "genero", igdb_id: 12, nome: "RPG" };
const item = (id: number, sugerido = false) => ({
  igdb_id: id,
  nome: `Jogo ${id}`,
  igdb_capa_url: null,
  ano_lancamento: 1990 + id,
  sugerido,
  ja_zerado: false,
  jogo_zerado_id: null,
});
const resposta = (itens: ReturnType<typeof item>[], total = itens.length): CatalogoResposta => ({
  itens,
  meta: {
    pagina: 1,
    por_pagina: 60,
    total,
    total_sugeridos: itens.filter((jogo) => jogo.sugerido).length,
    total_todos: total,
  },
});

describe("useEscolherJogos", () => {
  it.each([
    [franquia, { origem: "franquia", id: 596 }],
    [plataforma, { origem: "plataforma", id: 130 }],
    [genero, { origem: "genero", id: 12 }],
  ])("carrega a página inicial para %s", async (origem, esperado) => {
    const buscar = vi.spyOn(listasService, "buscarCatalogo").mockResolvedValue(resposta([item(1)]));
    renderHook(() => useEscolherJogos({ origem, modo: "criar" }));
    await waitFor(() => expect(buscar).toHaveBeenCalled());
    expect(buscar.mock.calls[0][0]).toMatchObject(esperado);
    if (origem.tipo === "franquia") {
      expect(buscar.mock.calls[0][0]).toMatchObject({ somente_sugeridos: true });
    }
  });

  it("envia o filtro cruzado de plataforma e gênero", async () => {
    const buscar = vi.spyOn(listasService, "buscarCatalogo").mockResolvedValue(resposta([]));
    const { result } = renderHook(() => useEscolherJogos({ origem: plataforma, modo: "criar" }));
    await waitFor(() => expect(buscar).toHaveBeenCalled());
    act(() => result.current.setGeneroId(12));
    await waitFor(() => expect(buscar.mock.calls.at(-1)?.[0]).toMatchObject({ genero_id: 12 }));
  });

  it("concatena páginas sem duplicar", async () => {
    vi
      .spyOn(listasService, "buscarCatalogo")
      .mockResolvedValueOnce(resposta([item(1)], 2))
      .mockResolvedValueOnce({
        ...resposta([item(2)], 2),
        meta: { ...resposta([item(2)], 2).meta, pagina: 2 },
      });
    const { result } = renderHook(() => useEscolherJogos({ origem: franquia, modo: "criar" }));
    await waitFor(() => expect(result.current.itens).toHaveLength(1));
    await act(async () => {
      await result.current.carregarMais();
    });
    expect(result.current.itens.map((jogo) => jogo.igdb_id)).toEqual([1, 2]);
  });

  it("desativa somente sugeridos, volta à página 1 e mantém a seleção", async () => {
    const buscar = vi
      .spyOn(listasService, "buscarCatalogo")
      .mockResolvedValueOnce(resposta([item(1, true)], 2))
      .mockResolvedValueOnce({
        ...resposta([item(2, false)], 2),
        meta: { ...resposta([item(2, false)], 2).meta, pagina: 1, total_todos: 2 },
      });
    const { result } = renderHook(() => useEscolherJogos({ origem: franquia, modo: "criar" }));
    await waitFor(() => expect(result.current.itens).toHaveLength(1));
    act(() => result.current.alternar(item(1, true)));
    await act(async () => {
      result.current.setSomenteSugeridos(false);
    });
    await waitFor(() => expect(buscar).toHaveBeenCalledTimes(2));
    expect(buscar.mock.calls[1][0]).toMatchObject({ pagina: 1, somente_sugeridos: false });
    expect(result.current.payload.map((jogo) => jogo.igdb_id)).toEqual([1]);
  });

  it("mantém a seleção ao trocar filtro", async () => {
    vi.spyOn(listasService, "buscarCatalogo").mockResolvedValue(resposta([item(1)]));
    const { result } = renderHook(() => useEscolherJogos({ origem: franquia, modo: "criar" }));
    await waitFor(() => expect(result.current.itens).toHaveLength(1));
    act(() => {
      result.current.alternar(item(1));
      result.current.setBusca("zelda");
    });
    expect(result.current.payload[0].igdb_id).toBe(1);
  });

  it("alterna e desmarca pelo id mesmo fora da página", async () => {
    vi.spyOn(listasService, "buscarCatalogo").mockResolvedValue(resposta([item(1)]));
    const { result } = renderHook(() => useEscolherJogos({ origem: franquia, modo: "criar" }));
    await waitFor(() => expect(result.current.itens).toHaveLength(1));
    act(() => result.current.alternar(item(1)));
    act(() => result.current.desmarcar(1));
    expect(result.current.payload).toHaveLength(0);
  });

  it("marca somente sugeridos de todas as páginas", async () => {
    const buscar = vi
      .spyOn(listasService, "buscarCatalogo")
      .mockResolvedValueOnce(resposta([item(1, true)], 61))
      .mockResolvedValueOnce({
        ...resposta([item(2, false)], 61),
        meta: { ...resposta([item(2, false)], 61).meta, pagina: 2, total_sugeridos: 1 },
      });
    const { result } = renderHook(() => useEscolherJogos({ origem: franquia, modo: "criar" }));
    await waitFor(() => expect(result.current.itens).toHaveLength(1));
    await act(async () => {
      await result.current.marcarSugeridos();
    });
    expect(result.current.payload.map((jogo) => jogo.igdb_id)).toEqual([1]);
    expect(buscar).toHaveBeenCalledTimes(2);
  });

  it("não alterna nem envia existentes no modo adicionar", async () => {
    vi.spyOn(listasService, "buscarCatalogo").mockResolvedValue(resposta([item(1)]));
    const { result } = renderHook(() =>
      useEscolherJogos({
        origem: franquia,
        modo: "adicionar",
        existentes: [{ igdb_id: 1 } as never],
      }),
    );
    await waitFor(() => expect(result.current.itens).toHaveLength(1));
    act(() => result.current.alternar(item(1)));
    expect(result.current.payload).toHaveLength(0);
  });

  it("mapeia indisponibilidade para mensagem neutra", async () => {
    vi.spyOn(listasService, "buscarCatalogo").mockRejectedValue(
      new ApiError("igdb.unavailable", ""),
    );
    const { result } = renderHook(() => useEscolherJogos({ origem: franquia, modo: "criar" }));
    await waitFor(() =>
      expect(result.current.erro).toBe("Não foi possível carregar os jogos agora. Tente de novo."),
    );
  });
});
