import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/store/authStore'
import { BottomNav } from '@/components/layout/BottomNav'

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

  return sessao ? <><Outlet /><BottomNav /></> : <Navigate to="/login" state={{ from: location }} replace />
}
