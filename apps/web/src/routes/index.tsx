import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { AuthProvider } from '@/store/authStore'
import { PrivateRoute } from '@/components/auth/PrivateRoute'
import { LoginPage } from '@/pages/auth/LoginPage'
import { DashboardPage } from '@/pages/DashboardPage'
import { CadastroPage } from '@/pages/auth/CadastroPage'
import { EsqueciSenhaPage } from '@/pages/auth/EsqueciSenhaPage'
import { RedefinirSenhaPage } from '@/pages/auth/RedefinirSenhaPage'

export function AppRoutes() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route element={<PrivateRoute />}>
            <Route path="/" element={<DashboardPage />} />
          </Route>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/cadastro" element={<CadastroPage />} />
          <Route path="/esqueci-senha" element={<EsqueciSenhaPage />} />
          <Route path="/redefinir-senha" element={<RedefinirSenhaPage />} />
          <Route path="*" element={<main className="p-8">Página não encontrada.</main>} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  )
}
