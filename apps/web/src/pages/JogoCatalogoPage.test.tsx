import { afterEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { AuthProvider } from "@/store/authStore";
import { JogandoProvider } from "@/stores/jogandoStore";
import { JogoCatalogoPage } from "./JogoCatalogoPage";
import { jogosService, JogosApiError } from "@/lib/services/jogosService";
import { jogandoService } from "@/lib/services/jogandoService";
import { ApiError } from "@/lib/api";
import type { IGDBJogoSugestao } from "@/types/jogos";

const jogoMock: IGDBJogoSugestao = {
  id: 1029,
  name: "Chrono Trigger",
  cover: { id: 1, url: "//images.igdb.com/igdb/image/upload/t_thumb/co123.jpg" },
  first_release_date: 794966400,
  summary: "A classic RPG where heroes travel through time to save the world.",
  platforms: [
    { id: 19, name: "SNES" },
    { id: 24, name: "PlayStation" },
  ],
  genres: [
    { id: 12, name: "RPG" },
    { id: 31, name: "Aventura" },
  ],
};

function renderCatalogo(
  initialRoute = "/catalogo/1029",
  state?: { voltarPara?: string },
) {
  return render(
    <MemoryRouter initialEntries={[{ pathname: initialRoute, state }]}>
      <AuthProvider>
        <JogandoProvider>
          <Routes>
            <Route path="/listas" element={<div data-testid="pagina-listas">Listas</div>} />
            <Route path="/desafios/ds" element={<div data-testid="pagina-desafio">Desafio DS</div>} />
            <Route path="/catalogo/:igdbId" element={<JogoCatalogoPage />} />
          </Routes>
        </JogandoProvider>
      </AuthProvider>
    </MemoryRouter>,
  );
}

describe("JogoCatalogoPage", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renderiza skeleton enquanto está carregando", () => {
    vi.spyOn(jogosService, "obterDetalhesIGDB").mockReturnValue(new Promise(() => {}));
    const { container } = renderCatalogo();

    expect(container.querySelector(".animate-pulse")).toBeInTheDocument();
  });

  it("renderiza detalhes do jogo, capas, plataformas, gêneros, ano e Sobre o jogo", async () => {
    vi.spyOn(jogosService, "obterDetalhesIGDB").mockResolvedValue(jogoMock);
    renderCatalogo();

    expect(await screen.findByRole("heading", { name: "Chrono Trigger" })).toBeInTheDocument();
    expect(document.title).toBe("Chrono Trigger · Memory Card");
    expect(screen.getByText("SNES")).toBeInTheDocument();
    expect(screen.getByText("PlayStation")).toBeInTheDocument();
    expect(screen.getByText("RPG")).toBeInTheDocument();
    expect(screen.getByText("Aventura")).toBeInTheDocument();
    expect(screen.getByText("1995")).toBeInTheDocument();
    expect(screen.getByText("Sobre o jogo")).toBeInTheDocument();
    expect(
      screen.getByText("A classic RPG where heroes travel through time to save the world."),
    ).toBeInTheDocument();
  });

  it("exibe estado 404 quando o jogo não for encontrado", async () => {
    vi.spyOn(jogosService, "obterDetalhesIGDB").mockRejectedValue(
      new JogosApiError("jogos.not_found", "Não encontrado", 404),
    );
    renderCatalogo("/catalogo/999999");

    expect(await screen.findByText("Jogo não encontrado")).toBeInTheDocument();
  });

  it("exibe erro de rede e recupera após clicar em Tentar novamente", async () => {
    const user = userEvent.setup();
    const spy = vi
      .spyOn(jogosService, "obterDetalhesIGDB")
      .mockRejectedValueOnce(new Error("Falha na rede"))
      .mockResolvedValueOnce(jogoMock);

    renderCatalogo();

    expect(await screen.findByText("Erro ao carregar registro")).toBeInTheDocument();
    const btnTentar = screen.getByRole("button", { name: "Tentar novamente" });
    await user.click(btnTentar);

    expect(await screen.findByRole("heading", { name: "Chrono Trigger" })).toBeInTheDocument();
    expect(spy).toHaveBeenCalledTimes(2);
  });

  it("chama criar com os dados corretos ao clicar em Jogar agora e atualiza estado", async () => {
    const user = userEvent.setup();
    vi.spyOn(jogosService, "obterDetalhesIGDB").mockResolvedValue(jogoMock);
    const spyCriar = vi.spyOn(jogandoService, "criar").mockResolvedValue({
      id: 50,
      nome: "Chrono Trigger",
      igdb_id: 1029,
      igdb_capa_url: "https://images.igdb.com/igdb/image/upload/t_cover_big/co123.jpg",
      iniciado_em: "2026-10-24",
    });

    renderCatalogo();

    expect(await screen.findByRole("heading", { name: "Chrono Trigger" })).toBeInTheDocument();
    const botoesJogar = screen.getAllByRole("button", { name: /Jogar agora/ });
    await user.click(botoesJogar[0]);

    expect(spyCriar).toHaveBeenCalledWith(
      expect.objectContaining({
        nome: "Chrono Trigger",
        igdb_id: 1029,
        igdb_capa_url: "https://images.igdb.com/igdb/image/upload/t_cover_big/co123.jpg",
      }),
      undefined,
    );
    expect(await screen.findByText("Adicionado aos jogos em andamento!")).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: /Em andamento/ })[0]).toBeDisabled();
  });

  it("exibe loading no botão enquanto a requisição de iniciar está em andamento", async () => {
    const user = userEvent.setup();
    vi.spyOn(jogosService, "obterDetalhesIGDB").mockResolvedValue(jogoMock);
    vi.spyOn(jogandoService, "criar").mockReturnValue(new Promise(() => {}));

    renderCatalogo();

    expect(await screen.findByRole("heading", { name: "Chrono Trigger" })).toBeInTheDocument();
    const botoesJogar = screen.getAllByRole("button", { name: /Jogar agora/ });
    await user.click(botoesJogar[0]);

    const botoesIniciando = screen.getAllByRole("button", { name: /Iniciando\.\.\./ });
    expect(botoesIniciando[0]).toBeDisabled();
  });

  it("exibe mensagem de erro resolvida pelo código e não pela mensagem literal", async () => {
    const user = userEvent.setup();
    vi.spyOn(jogosService, "obterDetalhesIGDB").mockResolvedValue(jogoMock);
    vi.spyOn(jogandoService, "criar").mockRejectedValue(
      new ApiError("jogando.item_duplicado", "Mensagem backend ignorada", 409),
    );

    renderCatalogo();

    expect(await screen.findByRole("heading", { name: "Chrono Trigger" })).toBeInTheDocument();
    const botoesJogar = screen.getAllByRole("button", { name: /Jogar agora/ });
    await user.click(botoesJogar[0]);

    expect(await screen.findByText("Este jogo já está em andamento.")).toBeInTheDocument();
    expect(screen.queryByText("Mensagem backend ignorada")).not.toBeInTheDocument();
  });

  it("navega para state.voltarPara ao clicar em voltar quando informado", async () => {
    const user = userEvent.setup();
    vi.spyOn(jogosService, "obterDetalhesIGDB").mockResolvedValue(jogoMock);
    renderCatalogo("/catalogo/1029", { voltarPara: "/desafios/ds" });

    expect(await screen.findByRole("heading", { name: "Chrono Trigger" })).toBeInTheDocument();
    const btnVoltar = screen.getAllByRole("button", { name: "Voltar para o desafio" })[0];
    await user.click(btnVoltar);

    expect(await screen.findByTestId("pagina-desafio")).toBeInTheDocument();
  });

  it("navega para fallback /listas ao clicar em voltar quando state não é fornecido", async () => {
    const user = userEvent.setup();
    vi.spyOn(jogosService, "obterDetalhesIGDB").mockResolvedValue(jogoMock);
    renderCatalogo();

    expect(await screen.findByRole("heading", { name: "Chrono Trigger" })).toBeInTheDocument();
    const btnVoltar = screen.getAllByRole("button", { name: "Voltar para o desafio" })[0];
    await user.click(btnVoltar);

    expect(await screen.findByTestId("pagina-listas")).toBeInTheDocument();
  });
});
