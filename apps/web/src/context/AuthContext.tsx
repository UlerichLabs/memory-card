import { createContext, useCallback, useEffect, useMemo, useState, type ReactNode } from "react";
import type { Usuario } from "@memory-card/types";
import { loginRequest, meRequest } from "@/lib/auth.service";

type AuthContextValue = {
  user: Usuario | null;
  accessToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, senha: string) => Promise<void>;
  logout: () => void;
};

export const AuthContext = createContext<AuthContextValue | null>(null);

type AuthProviderProps = {
  children: ReactNode;
};

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<Usuario | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(() => localStorage.getItem("mc_token"));
  const [isLoading, setIsLoading] = useState(Boolean(accessToken));

  const logout = useCallback(() => {
    localStorage.removeItem("mc_token");
    setAccessToken(null);
    setUser(null);
    window.location.assign("/login");
  }, []);

  useEffect(() => {
    if (!accessToken) {
      setIsLoading(false);
      return;
    }

    const token = accessToken;

    async function loadUser() {
      try {
        const usuario = await meRequest(token);
        setUser(usuario);
      } catch {
        logout();
      } finally {
        setIsLoading(false);
      }
    }

    void loadUser();
  }, [accessToken, logout]);

  async function login(email: string, senha: string) {
    const data = await loginRequest(email, senha);
    localStorage.setItem("mc_token", data.accessToken);
    setAccessToken(data.accessToken);
    setUser(data.usuario);
  }

  const value = useMemo(
    () => ({
      user,
      accessToken,
      isAuthenticated: Boolean(user && accessToken),
      isLoading,
      login,
      logout
    }),
    [accessToken, isLoading, logout, user]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
