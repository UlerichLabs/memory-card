import { afterEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes, useLocation } from "react-router-dom";
import { AuthProvider } from "@/store/authStore";
import { JogosProvider } from "@/stores/jogosStore";
import { GameFormDialog } from "@/components/jogos/GameForm/GameFormDialog";
import { JogoDetalhePage } from "./JogoDetalhePage";
import { jogosService, JogosApiError } from "@/lib/services/jogosService";
import type { JogoZeradoDTO } from "@/types/jogos";

const jogoCompletoMock: JogoZeradoDTO = {
  id: 42,
  numero: 3,
  usuario_id: 10,
  nome: "The Legend of Zelda: Ocarina of Time",
  console: "Nintendo 64",
  genero: "Aventura, RPG, Ação",
  tipo: "Campanha",
  iniciado_em: "2026-01-01T00:00:00Z",
  finalizado_em: "2026-01-11T00:00:00Z",
  tempo_jogado: 128700,
  nota: 11,
  dificuldade: "AA",
  review: "Um dos melhores jogos da minha vida.",
  destaque: true,
  igdb_capa_url: "https://images.igdb.com/cover.jpg",
  igdb_descricao: "Link viaja no tempo para deter Ganondorf.",
  created_at: "2026-01-11T12:00:00Z",
  updated_at: "2026-01-12T15:00:00Z",
};

function BibliotecaMock() {
  const loc = useLocation();
  return (
    <div data-testid="pagina-biblioteca">
      {loc.pathname}
      {loc.search}
    </div>
  );
}

function renderDetalhe(initialRoute = "/biblioteca/42", state?: { from?: string }) {
  return render(
    <MemoryRouter initialEntries={[{ pathname: initialRoute, state }]}>
      <AuthProvider>
        <JogosProvider>
          <GameFormDialog />
          <Routes>
            <Route
              path="/biblioteca"
              element={<BibliotecaMock />}
            />
            <Route
              path="/biblioteca/:id"
              element={<JogoDetalhePage />}
            />
          </Routes>
        </JogosProvider>
      </AuthProvider>
    </MemoryRouter>,
  );
}

describe("JogoDetalhePage", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renderiza todos os campos, Registro #N, chips de gênero quebrados e selo de destaque", async () => {
    vi.spyOn(jogosService, "obterPorId").mockResolvedValue(jogoCompletoMock);
    renderDetalhe();

    expect(
      await screen.findByRole("heading", {
        name: "The Legend of Zelda: Ocarina of Time",
      }),
    ).toBeInTheDocument();
    expect(screen.getAllByText("Registro #3").length).toBeGreaterThanOrEqual(1);
    expect(
      screen.getAllByText(/Criado em 11\/01\/2026 · atualizado em 12\/01\/2026/).length,
    ).toBeGreaterThanOrEqual(1);
    expect(screen.getByText("Jogo do ano 2026")).toBeInTheDocument();
    expect(screen.getByText("Nintendo 64")).toBeInTheDocument();
    expect(screen.getByText("Aventura")).toBeInTheDocument();
    expect(screen.getByText("RPG")).toBeInTheDocument();
    expect(screen.getByText("Ação")).toBeInTheDocument();
    expect(screen.getByText("Campanha")).toBeInTheDocument();
    expect(screen.getByText("Jogo da Vida")).toBeInTheDocument();
    expect(screen.getByText("35h 45m")).toBeInTheDocument();
    expect(screen.getByText("35:45:00")).toBeInTheDocument();
    expect(screen.getByText("11/01/2026")).toBeInTheDocument();
    expect(screen.getByText(/Iniciado em 01\/01\/2026 · 10 dias/)).toBeInTheDocument();
    expect(screen.getByText("Um dos melhores jogos da minha vida.")).toBeInTheDocument();
    expect(screen.getByText("Link viaja no tempo para deter Ganondorf.")).toBeInTheDocument();
    expect(screen.getByText("Fonte: catálogo de jogos")).toBeInTheDocument();
  });

  it(
    "lida com campos opcionais ausentes (sem iniciado_em, sem tipo, sem review, sem descricao, sem destaque)",
    async () => {
    const jogoMinimoMock: JogoZeradoDTO = {
      id: 99,
      usuario_id: 10,
      nome: "Tetris",
      console: "Game Boy",
      finalizado_em: "2026-02-01T00:00:00Z",
      tempo_jogado: 0,
      nota: 8,
      dificuldade: "B",
      destaque: false,
      created_at: "2026-02-01T00:00:00Z",
    };

    vi.spyOn(jogosService, "obterPorId").mockResolvedValue(jogoMinimoMock);
    renderDetalhe("/biblioteca/99");

    expect(await screen.findByRole("heading", { name: "Tetris" })).toBeInTheDocument();
    expect(screen.queryByText(/Jogo do ano/)).not.toBeInTheDocument();
    expect(screen.queryByText(/Iniciado em/)).not.toBeInTheDocument();
    expect(screen.getByText("Sem review.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Adicionar review" })).toBeInTheDocument();
    expect(screen.queryByText("Sobre o jogo")).not.toBeInTheDocument();
    expect(screen.getByText("—")).toBeInTheDocument();
  });

  it("exibe estado de não encontrado quando API retorna 404 ou 400", async () => {
    vi.spyOn(jogosService, "obterPorId").mockRejectedValue(
      new JogosApiError("jogos.not_found", "Não encontrado", 404),
    );
    renderDetalhe("/biblioteca/999999");

    expect(await screen.findByText("Jogo não encontrado")).toBeInTheDocument();
    expect(
      screen.getAllByRole("button", { name: "Voltar para a Biblioteca" }).length,
    ).toBeGreaterThanOrEqual(1);
  });

  it("exibe mensagem genérica e botão tentar novamente em caso de erro de rede", async () => {
    const user = userEvent.setup();
    const spy = vi
      .spyOn(jogosService, "obterPorId")
      .mockRejectedValueOnce(new Error("Network failure"))
      .mockResolvedValueOnce(jogoCompletoMock);

    renderDetalhe();

    expect(await screen.findByText("Erro ao carregar registro")).toBeInTheDocument();
    const btnTentar = screen.getByRole("button", { name: "Tentar novamente" });
    await user.click(btnTentar);

    expect(
      await screen.findByRole("heading", {
        name: "The Legend of Zelda: Ocarina of Time",
      }),
    ).toBeInTheDocument();
    expect(spy).toHaveBeenCalledTimes(2);
  });

  it("botão Voltar restaura os filtros anteriores passados pelo state", async () => {
    const user = userEvent.setup();
    vi.spyOn(jogosService, "obterPorId").mockResolvedValue(jogoCompletoMock);
    renderDetalhe("/biblioteca/42", { from: "?busca=Zelda&modo=list" });

    expect(
      await screen.findByRole("heading", {
        name: "The Legend of Zelda: Ocarina of Time",
      }),
    ).toBeInTheDocument();
    const btnVoltar = screen.getAllByRole("button", {
      name: "Voltar para a Biblioteca",
    })[0];
    await user.click(btnVoltar);

    expect(await screen.findByTestId("pagina-biblioteca")).toHaveTextContent(
      "/biblioteca?busca=Zelda&modo=list",
    );
  });

  it("abre o modal de edição ao clicar em Editar e modal de exclusão ao clicar em Excluir", async () => {
    const user = userEvent.setup();
    vi.spyOn(jogosService, "obterPorId").mockResolvedValue(jogoCompletoMock);
    renderDetalhe();

    expect(
      await screen.findByRole("heading", {
        name: "The Legend of Zelda: Ocarina of Time",
      }),
    ).toBeInTheDocument();
    const btnEditar = screen.getAllByRole("button", { name: /Editar/ })[0];
    await user.click(btnEditar);

    expect(await screen.findByRole("heading", { name: "Editar registro" })).toBeVisible();

    const btnCancelar = screen.getByRole("button", { name: "Cancelar" });
    await user.click(btnCancelar);

    const btnExcluir = screen.getAllByRole("button", { name: /Excluir/ })[0];
    await user.click(btnExcluir);

    expect(await screen.findByRole("heading", { name: "Excluir registro" })).toBeVisible();
  });
});
