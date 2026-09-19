import { Link } from 'react-router-dom'
import { SolicitarResetForm } from '@/components/auth/SolicitarResetForm'
import { authTheme } from './authTheme'

export function EsqueciSenhaPage() {
  return (
    <main style={authTheme} className="flex min-h-svh items-center justify-center bg-background p-6 text-foreground">
      <section className="w-full max-w-md space-y-6 rounded-xl border bg-card p-5">
        <header className="space-y-2">
          <p className="text-xs font-medium text-muted-foreground">Memory Card</p>
          <h1 className="text-[22px] font-bold">Recuperar senha</h1>
          <p className="text-[13px] text-muted-foreground">Informe seu email para receber as instruções.</p>
        </header>
        <SolicitarResetForm />
        <footer className="text-center text-[13px] text-muted-foreground">
          <Link to="/login" className="font-medium text-primary underline-offset-4 hover:underline">Voltar para entrar</Link>
        </footer>
      </section>
    </main>
  )
}
