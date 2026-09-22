import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { AuthProvider } from '@/store/authStore'
import { JogosProvider } from '@/stores/jogosStore'
import { PrivateRoute } from '@/components/auth/PrivateRoute'
import { LoginPage } from '@/pages/auth/LoginPage'
import { DashboardPage } from '@/pages/DashboardPage'
import { CadastroPage } from '@/pages/auth/CadastroPage'
import { EsqueciSenhaPage } from '@/pages/auth/EsqueciSenhaPage'
import { RedefinirSenhaPage } from '@/pages/auth/RedefinirSenhaPage'
import { ContaTrocarSenhaPage } from '@/pages/ContaTrocarSenhaPage'
import { BibliotecaPage } from '@/pages/BibliotecaPage'
import { NovoJogoPage } from '@/pages/NovoJogoPage'
import { EditarJogoPage } from '@/pages/EditarJogoPage'
import { EmConstrucaoPage } from '@/pages/EmConstrucaoPage'

export function AppRoutes() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <JogosProvider>
          <Routes>
            <Route element={<PrivateRoute />}>
              <Route path="/" element={<DashboardPage />} />
              <Route path="/conta/trocar-senha" element={<ContaTrocarSenhaPage />} />
              <Route path="/biblioteca" element={<BibliotecaPage />} />
              <Route path="/jogos/novo" element={<NovoJogoPage />} />
              <Route path="/jogos/:id/editar" element={<EditarJogoPage />} />
              <Route path="/abandonados" element={<EmConstrucaoPage modulo="Abandonados" />} />
              <Route path="/desafios" element={<EmConstrucaoPage modulo="Desafios" />} />
              <Route path="/listas" element={<EmConstrucaoPage modulo="Listas" />} />
              <Route path="/explorador" element={<EmConstrucaoPage modulo="Explorador" />} />
            </Route>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/cadastro" element={<CadastroPage />} />
            <Route path="/esqueci-senha" element={<EsqueciSenhaPage />} />
            <Route path="/redefinir-senha" element={<RedefinirSenhaPage />} />
            <Route path="*" element={<main className="p-8">Página não encontrada.</main>} />
          </Routes>
        </JogosProvider>
      </AuthProvider>
    </BrowserRouter>
  )
}
