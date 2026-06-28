import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { vi } from "vitest";

import type { Usuario } from "@memory-card/types";

import { AuthProvider } from "@/context/AuthContext";
import CadastroPage from "@/pages/cadastro";

const authServiceMock = vi.hoisted(() => ({
  loginRequest: vi.fn(),
  meRequest: vi.fn(),
  registerRequest: vi.fn()
}));

vi.mock("@/lib/auth.service", () => authServiceMock);

function usuario(): Usuario {
  return {
    id: 1,
    nome: "Teste",
    username: "teste",
    email: "teste@test.com",
    avatar: null,
    bio: null,
    createdAt: new Date(),
    updatedAt: new Date()
  };
}

function renderRegisterPage() {
  return render(
    <MemoryRouter initialEntries={["/cadastro"]}>
      <AuthProvider>
        <Routes>
          <Route path="/cadastro" element={<CadastroPage />} />
          <Route path="/dashboard" element={<div>Dashboard</div>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>
  );
}

async function fillRegisterForm() {
  const user = userEvent.setup();

  await user.type(screen.getByLabelText("Nome"), "Teste");
  await user.type(screen.getByLabelText("Username"), "teste");
  await user.type(screen.getByLabelText("Email"), "teste@test.com");
  await user.type(screen.getByLabelText("Senha"), "senha1234");
  await user.type(screen.getByLabelText("Confirmar senha"), "senha1234");

  return user;
}

describe("RegisterPage", () => {
  beforeEach(() => {
    localStorage.clear();
    authServiceMock.loginRequest.mockReset();
    authServiceMock.meRequest.mockReset();
    authServiceMock.registerRequest.mockReset();
  });

  it("renderiza campos de nome, email, username e senha", () => {
    renderRegisterPage();

    expect(screen.getByLabelText("Nome")).toBeInTheDocument();
    expect(screen.getByLabelText("Email")).toBeInTheDocument();
    expect(screen.getByLabelText("Username")).toBeInTheDocument();
    expect(screen.getByLabelText("Senha")).toBeInTheDocument();
  });

  it("exibe erro se senha menor que 8 caracteres", async () => {
    renderRegisterPage();
    const user = userEvent.setup();

    await user.type(screen.getByLabelText("Nome"), "Teste");
    await user.type(screen.getByLabelText("Username"), "teste");
    await user.type(screen.getByLabelText("Email"), "teste@test.com");
    await user.type(screen.getByLabelText("Senha"), "1234567");
    await user.type(screen.getByLabelText("Confirmar senha"), "1234567");
    await user.click(screen.getByRole("button", { name: "Criar conta" }));

    expect(screen.getByText("A senha deve ter pelo menos 8 caracteres")).toBeInTheDocument();
  });

  it("exibe erro se campos obrigatórios vazios", async () => {
    renderRegisterPage();

    await userEvent.click(screen.getByRole("button", { name: "Criar conta" }));

    expect(screen.getByLabelText("Nome")).toBeInvalid();
    expect(screen.getByLabelText("Email")).toBeInvalid();
    expect(authServiceMock.registerRequest).not.toHaveBeenCalled();
  });

  it("faz login automático após cadastro bem-sucedido", async () => {
    authServiceMock.registerRequest.mockResolvedValue(usuario());
    authServiceMock.loginRequest.mockResolvedValue({
      usuario: usuario(),
      accessToken: "token"
    });
    renderRegisterPage();

    const user = await fillRegisterForm();
    await user.click(screen.getByRole("button", { name: "Criar conta" }));

    expect(authServiceMock.registerRequest).toHaveBeenCalledWith({
      nome: "Teste",
      username: "teste",
      email: "teste@test.com",
      senha: "senha1234"
    });
    expect(authServiceMock.loginRequest).toHaveBeenCalledWith("teste@test.com", "senha1234");
  });

  it("redireciona para /dashboard após cadastro", async () => {
    authServiceMock.registerRequest.mockResolvedValue(usuario());
    authServiceMock.loginRequest.mockResolvedValue({
      usuario: usuario(),
      accessToken: "token"
    });
    renderRegisterPage();

    const user = await fillRegisterForm();
    await user.click(screen.getByRole("button", { name: "Criar conta" }));

    expect(await screen.findByText("Dashboard")).toBeInTheDocument();
  });

  it("exibe erro se email já cadastrado (409)", async () => {
    authServiceMock.registerRequest.mockRejectedValue(new Error("Email já cadastrado"));
    renderRegisterPage();

    const user = await fillRegisterForm();
    await user.click(screen.getByRole("button", { name: "Criar conta" }));

    expect(await screen.findByText("Email já cadastrado")).toBeInTheDocument();
  });
});
