import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { vi } from "vitest";

import type { Usuario } from "@memory-card/types";

import { AuthProvider } from "@/context/AuthContext";
import LoginPage from "@/pages/login";

const authServiceMock = vi.hoisted(() => ({
  loginRequest: vi.fn(),
  meRequest: vi.fn(),
  registerRequest: vi.fn()
}));

vi.mock("@/lib/auth.service", () => authServiceMock);

type LoginResponse = {
  usuario: Usuario;
  accessToken: string;
};

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

function renderLoginPage() {
  return render(
    <MemoryRouter initialEntries={["/login"]}>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/dashboard" element={<div>Dashboard</div>} />
        </Routes>
      </AuthProvider>
    </MemoryRouter>
  );
}

async function fillLoginForm() {
  const user = userEvent.setup();

  await user.type(screen.getByLabelText("Email"), "teste@test.com");
  await user.type(screen.getByLabelText("Senha"), "senha1234");

  return user;
}

describe("LoginPage", () => {
  beforeEach(() => {
    localStorage.clear();
    authServiceMock.loginRequest.mockReset();
    authServiceMock.meRequest.mockReset();
    authServiceMock.registerRequest.mockReset();
  });

  it("renderiza campos de email e senha", () => {
    renderLoginPage();

    expect(screen.getByLabelText("Email")).toBeInTheDocument();
    expect(screen.getByLabelText("Senha")).toBeInTheDocument();
  });

  it("renderiza botão de entrar", () => {
    renderLoginPage();

    expect(screen.getByRole("button", { name: "Entrar" })).toBeInTheDocument();
  });

  it("exibe erro de validação se campos vazios", async () => {
    renderLoginPage();

    await userEvent.click(screen.getByRole("button", { name: "Entrar" }));

    expect(screen.getByLabelText("Email")).toBeInvalid();
    expect(screen.getByLabelText("Senha")).toBeInvalid();
    expect(authServiceMock.loginRequest).not.toHaveBeenCalled();
  });

  it("chama authService.login com email e senha corretos", async () => {
    authServiceMock.loginRequest.mockResolvedValue({
      usuario: usuario(),
      accessToken: "token"
    });
    renderLoginPage();

    const user = await fillLoginForm();
    await user.click(screen.getByRole("button", { name: "Entrar" }));

    expect(authServiceMock.loginRequest).toHaveBeenCalledWith("teste@test.com", "senha1234");
  });

  it("exibe mensagem de erro para credenciais inválidas (401)", async () => {
    authServiceMock.loginRequest.mockRejectedValue(new Error("Email ou senha inválidos"));
    renderLoginPage();

    const user = await fillLoginForm();
    await user.click(screen.getByRole("button", { name: "Entrar" }));

    expect(await screen.findByText("Email ou senha inválidos")).toBeInTheDocument();
  });

  it("redireciona para /dashboard após login bem-sucedido", async () => {
    authServiceMock.loginRequest.mockResolvedValue({
      usuario: usuario(),
      accessToken: "token"
    });
    renderLoginPage();

    const user = await fillLoginForm();
    await user.click(screen.getByRole("button", { name: "Entrar" }));

    expect(await screen.findByText("Dashboard")).toBeInTheDocument();
  });

  it("botão fica desabilitado durante a requisição", async () => {
    let resolveLogin: (value: LoginResponse) => void = () => undefined;
    const pendingLogin = new Promise<LoginResponse>((resolve) => {
      resolveLogin = resolve;
    });
    authServiceMock.loginRequest.mockReturnValue(pendingLogin);
    renderLoginPage();

    const user = await fillLoginForm();
    await user.click(screen.getByRole("button", { name: "Entrar" }));

    expect(screen.getByRole("button", { name: "Entrando..." })).toBeDisabled();

    resolveLogin({
      usuario: usuario(),
      accessToken: "token"
    });
    expect(await screen.findByText("Dashboard")).toBeInTheDocument();
  });
});
