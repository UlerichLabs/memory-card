import { Link } from 'react-router-dom'
import { CadastroForm } from '@/components/auth/CadastroForm'

export function CadastroPage() {
  return (
    <main className="flex min-h-svh items-center justify-center bg-background p-6 text-foreground">
      <section className="w-full max-w-md space-y-6 rounded-xl border p-8 shadow-sm">
        <header className="space-y-1 text-center">
          <h1 className="text-2xl font-bold tracking-tight">Criar conta</h1>
          <p className="text-sm text-muted-foreground">
            Cadastre-se para começar a registrar sua biblioteca
          </p>
        </header>

        <CadastroForm />

        <footer className="text-center text-sm text-muted-foreground">
          Já tem uma conta?{' '}
          <Link to="/login" className="font-medium text-primary underline-offset-4 hover:underline">
            Entrar
          </Link>
        </footer>
      </section>
    </main>
  )
}
