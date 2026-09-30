import { afterEach, describe, expect, it, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import { listasService } from "@/lib/services/listasService";
import { ListasProvider, useListasStore } from "./listasStore";
import { ApiError } from "@/lib/api";
import type { ListaResumo, ListaDetalhada, ListaItem } from "@/types/listas";

const mockResumo: ListaResumo = {
  id: 1,
  tipo: "fila",
  nome: "Fila 1",
  descricao: null,
  origem: null,
  total_itens: 2,
  itens_pendentes: 2,
  progresso: null,
  created_at: "2026-09-01T00:00:00Z",
  updated_at: "2026-09-01T00:00:00Z",
};

const mockDetalhe: ListaDetalhada = {
  ...mockResumo,
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
  ],
};

describe("listasStore", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("inicia com estado padrão", () => {
    const { result } = renderHook(() => useListasStore(), {
      wrapper: ListasProvider,
    });
    expect(result.current.listas).toEqual([]);
    expect(result.current.listaAberta).toBeNull();
    expect(result.current.isLoading).toBe(false);
    expect(result.current.isLoadingDetalhe).toBe(false);
    expect(result.current.error).toBeNull();
    expect(result.current.filtroAba).toBe("todos");
  });

  it("lança erro se usado fora do ListasProvider", () => {
    expect(() => renderHook(() => useListasStore())).toThrow(
      "useListasStore deve ser utilizado dentro de um ListasProvider",
    );
  });

  it("carregarListas armazena listas ordenadas e reseta erro", async () => {
    const desafio: ListaResumo = {
      ...mockResumo,
      id: 2,
      tipo: "desafio",
      created_at: "2026-09-02T00:00:00Z",
    };
    vi.spyOn(listasService, "listar").mockResolvedValue([mockResumo, desafio]);

    const { result } = renderHook(() => useListasStore(), {
      wrapper: ListasProvider,
    });

    await act(async () => {
      await result.current.carregarListas();
    });

    expect(result.current.listas[0].id).toBe(2);
    expect(result.current.listas[1].id).toBe(1);
    expect(result.current.isLoading).toBe(false);
  });

  it("abrirLista carrega detalhe da lista e reseta filtro de aba", async () => {
    vi.spyOn(listasService, "obterPorId").mockResolvedValue(mockDetalhe);

    const { result } = renderHook(() => useListasStore(), {
      wrapper: ListasProvider,
    });
    act(() => result.current.setFiltroAba("pendentes"));

    await act(async () => {
      await result.current.abrirLista(1);
    });

    expect(result.current.listaAberta).toEqual(mockDetalhe);
    expect(result.current.filtroAba).toBe("todos");
  });

  it("criarLista chama service, recarrega listas e define como aberta", async () => {
    vi.spyOn(listasService, "criar").mockResolvedValue(mockDetalhe);
    vi.spyOn(listasService, "listar").mockResolvedValue([mockResumo]);

    const { result } = renderHook(() => useListasStore(), {
      wrapper: ListasProvider,
    });

    await act(async () => {
      await result.current.criarLista({ tipo: "fila", nome: "Fila 1" });
    });

    expect(result.current.listaAberta).toEqual(mockDetalhe);
    expect(result.current.listas).toEqual([mockResumo]);
  });

  it("atualizarLista chama service e recarrega listas", async () => {
    const atualizada = { ...mockDetalhe, nome: "Fila Atualizada" };
    vi.spyOn(listasService, "atualizar").mockResolvedValue(atualizada);
    vi.spyOn(listasService, "listar").mockResolvedValue([
      { ...mockResumo, nome: "Fila Atualizada" },
    ]);

    const { result } = renderHook(() => useListasStore(), {
      wrapper: ListasProvider,
    });

    await act(async () => {
      await result.current.atualizarLista(1, { nome: "Fila Atualizada" });
    });

    expect(result.current.listaAberta?.nome).toBe("Fila Atualizada");
  });

  it("excluirLista chama service, limpa aberta se for a mesma e recarrega listas", async () => {
    vi.spyOn(listasService, "excluir").mockResolvedValue(undefined);
    vi.spyOn(listasService, "listar").mockResolvedValue([]);

    const { result } = renderHook(() => useListasStore(), {
      wrapper: ({ children }) =>
        ListasProvider({ children, initialListaAberta: mockDetalhe }),
    });

    await act(async () => {
      await result.current.excluirLista(1);
    });

    expect(result.current.listaAberta).toBeNull();
    expect(result.current.listas).toEqual([]);
  });

  it("adicionarItem e removerItem chamam service e recarregam lista e resumo", async () => {
    const novoItem: ListaItem = {
      id: 12,
      igdb_id: 102,
      nome: "Super Mario World",
      console: "SNES",
      igdb_capa_url: null,
      ano_lancamento: 1990,
      posicao: 3,
      origem: "item",
      zerado: false,
      jogo_zerado: null,
    };
    vi.spyOn(listasService, "adicionarItem").mockResolvedValue(novoItem);
    vi.spyOn(listasService, "removerItem").mockResolvedValue(undefined);
    vi.spyOn(listasService, "obterPorId").mockResolvedValue({
      ...mockDetalhe,
      itens: [...mockDetalhe.itens, novoItem],
    });
    vi.spyOn(listasService, "listar").mockResolvedValue([mockResumo]);

    const { result } = renderHook(() => useListasStore(), {
      wrapper: ({ children }) =>
        ListasProvider({ children, initialListaAberta: mockDetalhe }),
    });

    await act(async () => {
      await result.current.adicionarItem({ nome: "Super Mario World" });
    });

    expect(result.current.listaAberta?.itens).toHaveLength(3);

    await act(async () => {
      await result.current.removerItem(12);
    });

    expect(listasService.removerItem).toHaveBeenCalledWith(1, 12, undefined);
  });

  it("reordenarItens aplica otimista e faz rollback em erro com mensagem", async () => {
    vi.spyOn(listasService, "reordenarItens").mockRejectedValue(
      new ApiError("listas.ordem_invalida", "Ordem inválida", 400),
    );

    const { result } = renderHook(() => useListasStore(), {
      wrapper: ({ children }) =>
        ListasProvider({ children, initialListaAberta: mockDetalhe }),
    });

    await act(async () => {
      try {
        await result.current.reordenarItens([11, 10]);
      } catch {}
    });

    expect(result.current.listaAberta?.itens[0].id).toBe(10);
    expect(result.current.error).toBe(
      "Não foi possível salvar a nova ordem dos jogos.",
    );

    act(() => result.current.limparErro());
    expect(result.current.error).toBeNull();
  });

  it("associarZeramento chama service e recarrega dados", async () => {
    vi.spyOn(listasService, "associarZeramento").mockResolvedValue(
      mockDetalhe.itens[0],
    );
    vi.spyOn(listasService, "obterPorId").mockResolvedValue(mockDetalhe);
    vi.spyOn(listasService, "listar").mockResolvedValue([mockResumo]);

    const { result } = renderHook(() => useListasStore(), {
      wrapper: ({ children }) =>
        ListasProvider({ children, initialListaAberta: mockDetalhe }),
    });

    await act(async () => {
      await result.current.associarZeramento(10, 50);
    });

    expect(listasService.associarZeramento).toHaveBeenCalledWith(
      1,
      10,
      50,
      undefined,
    );
  });

  it("controla abertura e fechamento de modais criar, editar e excluir", () => {
    const { result } = renderHook(() => useListasStore(), {
      wrapper: ListasProvider,
    });

    act(() => result.current.abrirModalCriar());
    expect(result.current.isNovaListaOpen).toBe(true);
    expect(result.current.listaEmEdicao).toBeNull();

    act(() => result.current.abrirModalEditar(mockResumo));
    expect(result.current.isNovaListaOpen).toBe(true);
    expect(result.current.listaEmEdicao).toEqual(mockResumo);

    act(() => result.current.fecharModalNovaLista());
    expect(result.current.isNovaListaOpen).toBe(false);
    expect(result.current.listaEmEdicao).toBeNull();

    act(() => result.current.abrirModalExcluir(mockResumo));
    expect(result.current.isExcluirListaOpen).toBe(true);
    expect(result.current.listaParaExcluir).toEqual(mockResumo);

    act(() => result.current.fecharModalExcluir());
    expect(result.current.isExcluirListaOpen).toBe(false);
    expect(result.current.listaParaExcluir).toBeNull();
  });
});
