import { Link, useLocation } from 'react-router-dom'
import { LoginForm } from '@/components/auth/LoginForm'
import { authTheme } from './authTheme'

export function LoginPage() {
  const location = useLocation()
  const mensagem = (location.state as { mensagem?: string } | null)?.mensagem

  return (
    <main style={authTheme} className="flex min-h-svh items-center justify-center bg-background p-6 text-foreground">
      <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap" />
      <section className="w-full max-w-md space-y-6 rounded-xl border bg-card p-5">
        <header className="space-y-2">
          <p className="text-xs font-medium text-muted-foreground">Memory Card</p>
          <h1 className="text-[22px] font-bold">Entrar na sua conta</h1>
          <p className="text-[13px] text-muted-foreground">Acesse suas memórias de jogos.</p>
        </header>
        <LoginForm />
        {mensagem && <p role="status" className="text-xs text-primary">{mensagem}</p>}
        <footer className="text-center text-[13px] text-muted-foreground">
          <Link to="/esqueci-senha" className="font-medium text-primary underline-offset-4 hover:underline">
            Esqueci minha senha
          </Link>
          <br />
          Ainda não tem uma conta?{' '}
          <Link to="/cadastro" className="font-medium text-primary underline-offset-4 hover:underline">
            Criar conta
          </Link>
        </footer>
      </section>
    </main>
  )
}
