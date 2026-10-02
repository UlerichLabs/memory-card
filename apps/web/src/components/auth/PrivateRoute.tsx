import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/store/authStore'
import { BottomNav } from '@/components/layout/BottomNav'
import { Footer } from '@/components/layout/Footer'

export function PrivateRoute() {
  const { sessao, isCarregandoSessao } = useAuthStore()
  const location = useLocation()

  if (isCarregandoSessao) {
    return (
      <div
        role="status"
        aria-label="Carregando sessão"
        className="flex min-h-svh items-center justify-center bg-[var(--bg-primary)] text-[var(--text-secondary)]"
      >
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-[var(--border)] border-t-[var(--accent)]" />
      </div>
    )
  }

  return sessao ? <div className="flex min-h-svh flex-col bg-[var(--bg-primary)]"><div className="min-w-0 flex-1"><Outlet /></div><Footer /><BottomNav /></div> : <Navigate to="/login" state={{ from: location }} replace />
}
