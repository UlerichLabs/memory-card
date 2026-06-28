import { FormEvent, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";
import { registerRequest } from "@/lib/auth.service";

export default function CadastroPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [nome, setNome] = useState("");
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [senha, setSenha] = useState("");
  const [confirmarSenha, setConfirmarSenha] = useState("");
  const [erro, setErro] = useState<string | null>(null);
  const [enviando, setEnviando] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErro(null);

    if (senha.length < 8) {
      setErro("A senha deve ter pelo menos 8 caracteres");
      return;
    }

    if (senha !== confirmarSenha) {
      setErro("As senhas não coincidem");
      return;
    }

    setEnviando(true);

    try {
      await registerRequest({ nome, username, email, senha });
      await login(email, senha);
      navigate("/dashboard", { replace: true });
    } catch (error) {
      setErro(error instanceof Error ? error.message : "Não foi possível criar a conta");
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
          <AuthInput label="Nome" value={nome} onChange={setNome} />
          <AuthInput label="Username" value={username} onChange={setUsername} />
          <AuthInput label="Email" type="email" value={email} onChange={setEmail} />
          <AuthInput label="Senha" type="password" value={senha} onChange={setSenha} />
          <AuthInput
            label="Confirmar senha"
            type="password"
            value={confirmarSenha}
            onChange={setConfirmarSenha}
          />
        </div>

        <button
          type="submit"
          disabled={enviando}
          className="mt-6 w-full rounded-lg bg-accent px-4 py-2 font-medium text-background disabled:cursor-not-allowed disabled:opacity-70"
        >
          {enviando ? "Criando..." : "Criar conta"}
        </button>

        {erro && <p className="mt-3 text-center text-sm text-red">{erro}</p>}

        <p className="mt-6 text-center text-sm text-muted">
          Já tem conta?{" "}
          <Link to="/login" className="font-medium text-accent">
            Entrar
          </Link>
        </p>
      </form>
    </main>
  );
}

type AuthInputProps = {
  label: string;
  value: string;
  onChange: (value: string) => void;
  type?: string;
};

function AuthInput({ label, value, onChange, type = "text" }: AuthInputProps) {
  return (
    <label className="flex flex-col gap-1.5">
      <span className="text-sm text-muted">{label}</span>
      <input
        type={type}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className="rounded-lg border border-border bg-surface-raised px-3 py-2 text-foreground outline-none focus:border-accent"
        required
      />
    </label>
  );
}
