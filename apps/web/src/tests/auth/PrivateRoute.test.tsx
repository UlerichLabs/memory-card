import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { vi } from "vitest";

import { PrivateRoute } from "@/components/layout/PrivateRoute";

type AuthState = {
  isAuthenticated: boolean;
  isLoading: boolean;
};

const authMock = vi.hoisted(() => ({
  state: {
    isAuthenticated: false,
    isLoading: false
  } as AuthState
}));

vi.mock("@/hooks/useAuth", () => ({
  useAuth: () => authMock.state
}));

vi.mock("@/components/layout/Navbar", () => ({
  Navbar: () => <nav>Navbar</nav>
}));

function renderPrivateRoute() {
  return render(
    <MemoryRouter initialEntries={["/dashboard"]}>
      <Routes>
        <Route element={<PrivateRoute />}>
          <Route path="/dashboard" element={<div>Conteúdo protegido</div>} />
        </Route>
        <Route path="/login" element={<div>Login</div>} />
      </Routes>
    </MemoryRouter>
  );
}

describe("PrivateRoute", () => {
  it("renderiza children quando usuário autenticado", () => {
    authMock.state = {
      isAuthenticated: true,
      isLoading: false
    };

    renderPrivateRoute();

    expect(screen.getByText("Conteúdo protegido")).toBeInTheDocument();
  });

  it("redireciona para /login quando não autenticado", () => {
    authMock.state = {
      isAuthenticated: false,
      isLoading: false
    };

    renderPrivateRoute();

    expect(screen.getByText("Login")).toBeInTheDocument();
  });

  it("redireciona para /login quando token expirado", () => {
    authMock.state = {
      isAuthenticated: false,
      isLoading: false
    };

    renderPrivateRoute();

    expect(screen.getByText("Login")).toBeInTheDocument();
  });
});
