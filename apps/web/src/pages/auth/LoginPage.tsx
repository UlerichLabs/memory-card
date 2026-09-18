import type { CSSProperties } from 'react'
import { Link } from 'react-router-dom'
import { LoginForm } from '@/components/auth/LoginForm'

const loginTheme = {
  '--background': '#15161A', '--foreground': '#EDEDED', '--card': '#1A1B20',
  '--secondary': '#1D1F25', '--border': '#24262C', '--input': '#2A2C33',
  '--muted-foreground': '#9A9CA5', '--primary': '#4F7CFF', '--primary-foreground': '#0E0F12',
  '--destructive': '#E05A4E', '--ring': '#4F7CFF', fontFamily: 'Inter, sans-serif',
} as CSSProperties

export function LoginPage() {
  return (
    <main style={loginTheme} className="flex min-h-svh items-center justify-center bg-background p-6 text-foreground">
      <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap" />
      <section className="w-full max-w-md space-y-6 rounded-xl border bg-card p-5">
        <header className="space-y-2">
          <p className="text-xs font-medium text-muted-foreground">Memory Card</p>
          <h1 className="text-[22px] font-bold">Entrar na sua conta</h1>
          <p className="text-[13px] text-muted-foreground">Acesse suas memórias de jogos.</p>
        </header>
        <LoginForm />
        <footer className="text-center text-[13px] text-muted-foreground">
          Ainda não tem uma conta?{' '}
          <Link to="/cadastro" className="font-medium text-primary underline-offset-4 hover:underline">
            Criar conta
          </Link>
        </footer>
      </section>
    </main>
  )
}
