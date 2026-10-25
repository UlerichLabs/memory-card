import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, useLocation } from "react-router-dom";
import { DesafioCard } from "./DesafioCard";
import type { ListaItem } from "@/types/listas";

const itemZerado: ListaItem = {
  id: 1,
  igdb_id: 100,
  nome: "Super Mario World",
  console: "SNES",
  igdb_capa_url: "https://images.igdb.com/cover.jpg",
  ano_lancamento: 1990,
  posicao: 1,
  origem: "item",
  zerado: true,
  jogo_zerado: {
    id: 42,
    finalizado_em: "2026-01-01",
    nota: 10,
  },
};

const itemPendenteComIgdb: ListaItem = {
  id: 2,
  igdb_id: 200,
  nome: "Chrono Trigger",
  console: "SNES",
  igdb_capa_url: "https://images.igdb.com/chrono.jpg",
  ano_lancamento: 1995,
  posicao: 2,
  origem: "item",
  zerado: false,
  jogo_zerado: null,
};

const itemPendenteSemIgdb: ListaItem = {
  id: 3,
  igdb_id: null,
  nome: "Jogo Desconhecido",
  console: null,
  igdb_capa_url: null,
  ano_lancamento: null,
  posicao: 3,
  origem: "item",
  zerado: false,
  jogo_zerado: null,
};

function LocationObserver() {
  const loc = useLocation();
  return (
    <div data-testid="location-state">
      {JSON.stringify(loc.state)}
    </div>
  );
}

function renderCard(item: ListaItem, initialRoute = "/listas/desafio-1?filtro=todos") {
  const onRemover = vi.fn();
  const onZerei = vi.fn();
  const utils = render(
    <MemoryRouter initialEntries={[initialRoute]}>
      <DesafioCard item={item} onRemover={onRemover} onZerei={onZerei} />
      <LocationObserver />
    </MemoryRouter>,
  );
  return { ...utils, onRemover, onZerei };
}

describe("DesafioCard", () => {
  it("redireciona para /biblioteca/:id com state.voltarPara quando o item está zerado", async () => {
    const user = userEvent.setup();
    renderCard(itemZerado, "/listas/5?aba=todos");
    const link = screen.getByRole("link");
    expect(link).toHaveAttribute("href", "/biblioteca/42");
    await user.click(link);
    expect(screen.getByTestId("location-state")).toHaveTextContent(
      JSON.stringify({ voltarPara: "/listas/5?aba=todos" }),
    );
  });

  it("renderiza links para /catalogo/:igdbId na capa e no nome quando pendente com igdb_id", async () => {
    const user = userEvent.setup();
    renderCard(itemPendenteComIgdb, "/listas/5?aba=pendentes");
    const links = screen.getAllByRole("link");
    expect(links).toHaveLength(2);
    expect(links[0]).toHaveAttribute("href", "/catalogo/200");
    expect(links[1]).toHaveAttribute("href", "/catalogo/200");
    await user.click(links[0]);
    expect(screen.getByTestId("location-state")).toHaveTextContent(
      JSON.stringify({ voltarPara: "/listas/5?aba=pendentes" }),
    );
  });

  it("não renderiza links quando o item é pendente sem igdb_id", () => {
    renderCard(itemPendenteSemIgdb);
    expect(screen.queryByRole("link")).not.toBeInTheDocument();
  });

  it("aciona onZerei e onRemover sem disparar navegação nos botões de ação", async () => {
    const user = userEvent.setup();
    const { onRemover, onZerei } = renderCard(itemPendenteComIgdb, "/listas/5");

    const botoesZerei = screen.getAllByRole("button", { name: "Zerei!" });
    expect(botoesZerei).toHaveLength(2);
    await user.click(botoesZerei[0]);
    expect(onZerei).toHaveBeenCalledWith(itemPendenteComIgdb);

    await user.click(botoesZerei[1]);
    expect(onZerei).toHaveBeenCalledTimes(2);

    const btnRemover = screen.getByRole("button", { name: "Tirar do desafio" });
    await user.click(btnRemover);
    expect(onRemover).toHaveBeenCalledWith(2);

    expect(screen.getByTestId("location-state")).toHaveTextContent("null");
  });
});
