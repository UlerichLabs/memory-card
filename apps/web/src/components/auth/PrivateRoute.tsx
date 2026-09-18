import { Navigate, Outlet } from 'react-router-dom'
import { useAuthStore } from '@/store/authStore'

export function PrivateRoute() {
  const { sessao } = useAuthStore()
  return sessao ? <Outlet /> : <Navigate to="/login" replace />
}
