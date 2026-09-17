import { useEffect, useState } from 'react'
import { Button } from '@/components/ui/button'
import { getHealth } from '@/services/api'

export function HomePage() {
  const [attempt, setAttempt] = useState(0)
  const [health, setHealth] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const controller = new AbortController()
    const timeout = window.setTimeout(() => controller.abort(), 10_000)
    let active = true

    getHealth(controller.signal)
      .then((result) => {
        if (active) setHealth(result.status)
      })
      .catch((cause: unknown) => {
        if (active) {
          setError(controller.signal.aborted
            ? 'A API demorou para responder. Tente novamente.'
            : cause instanceof Error ? cause.message : 'Não foi possível consultar a API.')
        }
      })
      .finally(() => window.clearTimeout(timeout))

    return () => {
      active = false
      window.clearTimeout(timeout)
      controller.abort()
    }
  }, [attempt])

  const loading = health === null && error === null

  return (
    <main className="flex min-h-svh items-center justify-center bg-background p-6 text-foreground">
      <section className="w-full max-w-lg space-y-6 rounded-xl border p-8 shadow-sm">
        <h1 className="text-2xl font-semibold tracking-tight">Memory Card — em construção</h1>
        <p className="text-muted-foreground">Em breve, um espaço para suas memórias de jogos.</p>
        <p role="status" aria-live="polite">
          {loading ? 'Verificando conexão…' : error ? `Conexão indisponível: ${error}` : `Status da API: ${health}`}
        </p>
        <Button
          disabled={loading}
          onClick={() => {
            setHealth(null)
            setError(null)
            setAttempt((value) => value + 1)
          }}
        >
          Verificar novamente
        </Button>
      </section>
    </main>
  )
}
