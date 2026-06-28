import { FormEvent, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";

export default function LoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [email, setEmail] = useState("");
  const [senha, setSenha] = useState("");
  const [erro, setErro] = useState<string | null>(null);
  const [enviando, setEnviando] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErro(null);
    setEnviando(true);

    try {
      await login(email, senha);
      navigate("/dashboard", { replace: true });
    } catch (error) {
      setErro(error instanceof Error ? error.message : "Não foi possível entrar");
    } finally {
      setEnviando(false);
    }
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-background px-6 text-foreground">
      <form
        onSubmit={handleSubmit}
        className="w-full max-w-sm rounded-xl border border-border bg-surface p-8 shadow-sm"
      >
        <div className="mb-8 text-center">
          <h1 className="font-sans text-2xl font-bold text-foreground">Memory Card</h1>
          <p className="mt-2 text-sm text-muted">Seu diário de zeramentos</p>
        </div>

        <div className="space-y-4">
          <label className="flex flex-col gap-1.5">
            <span className="text-sm text-muted">Email</span>
            <input
              type="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              className="rounded-lg border border-border bg-surface-raised px-3 py-2 text-foreground outline-none focus:border-accent"
              required
            />
          </label>

          <label className="flex flex-col gap-1.5">
            <span className="text-sm text-muted">Senha</span>
            <input
              type="password"
              value={senha}
              onChange={(event) => setSenha(event.target.value)}
              className="rounded-lg border border-border bg-surface-raised px-3 py-2 text-foreground outline-none focus:border-accent"
              required
            />
          </label>
        </div>

        <button
          type="submit"
          disabled={enviando}
          className="mt-6 w-full rounded-lg bg-accent px-4 py-2 font-medium text-background disabled:cursor-not-allowed disabled:opacity-70"
        >
          {enviando ? "Entrando..." : "Entrar"}
        </button>

        {erro && <p className="mt-3 text-center text-sm text-red">{erro}</p>}

        <p className="mt-6 text-center text-sm text-muted">
          Não tem conta?{" "}
          <Link to="/cadastro" className="font-medium text-accent">
            Criar conta
          </Link>
        </p>
      </form>
    </main>
  );
}
