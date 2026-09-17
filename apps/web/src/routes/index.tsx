import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { AuthProvider } from '@/store/authStore'
import { PrivateRoute } from '@/components/auth/PrivateRoute'
import { LoginPage } from '@/pages/auth/LoginPage'
import { HomePage } from '@/pages/HomePage'
import { CadastroPage } from '@/pages/auth/CadastroPage'

export function AppRoutes() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route element={<PrivateRoute />}>
            <Route path="/" element={<HomePage />} />
          </Route>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/cadastro" element={<CadastroPage />} />
          <Route path="*" element={<main className="p-8">Página não encontrada.</main>} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  )
}
