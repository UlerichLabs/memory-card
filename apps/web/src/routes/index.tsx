import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider } from '@/store/authStore'
import { JogosProvider } from '@/stores/jogosStore'
import { AbandonadosProvider } from '@/stores/abandonadosStore'
import { PrivateRoute } from '@/components/auth/PrivateRoute'
import { LoginPage } from '@/pages/auth/LoginPage'
import { DashboardPage } from '@/pages/DashboardPage'
import { CadastroPage } from '@/pages/auth/CadastroPage'
import { EsqueciSenhaPage } from '@/pages/auth/EsqueciSenhaPage'
import { RedefinirSenhaPage } from '@/pages/auth/RedefinirSenhaPage'
import { ContaTrocarSenhaPage } from '@/pages/ContaTrocarSenhaPage'
import { BibliotecaPage } from '@/pages/BibliotecaPage'
import { JogoDetalhePage } from '@/pages/JogoDetalhePage'
import { HallDaFamaPage } from '@/pages/HallDaFamaPage'
import { ListasPage } from '@/pages/ListasPage'
import { AbandonadosPage } from '@/pages/AbandonadosPage'
import { EmConstrucaoPage } from '@/pages/EmConstrucaoPage'
import { GameFormDialog } from '@/components/jogos/GameForm/GameFormDialog'
import { AbandonarJogoDialog } from '@/components/abandonados/AbandonarJogoDialog'

export function AppRoutes() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <JogosProvider>
          <AbandonadosProvider>
            <GameFormDialog />
            <AbandonarJogoDialog />
            <Routes>
              <Route element={<PrivateRoute />}>
                <Route path="/" element={<DashboardPage />} />
                <Route path="/conta/trocar-senha" element={<ContaTrocarSenhaPage />} />
                <Route path="/biblioteca" element={<BibliotecaPage />} />
                <Route path="/biblioteca/:id" element={<JogoDetalhePage />} />
                <Route path="/hall-da-fama" element={<HallDaFamaPage />} />
                <Route path="/jogos/novo" element={<Navigate to="/biblioteca" replace />} />
                <Route path="/jogos/:id/editar" element={<Navigate to="/biblioteca" replace />} />
                <Route path="/abandonados" element={<AbandonadosPage />} />
                <Route path="/desafios" element={<Navigate to="/listas" replace />} />
                <Route path="/listas" element={<ListasPage />} />
                <Route path="/listas/:id" element={<ListasPage />} />
                <Route path="/explorador" element={<EmConstrucaoPage modulo="Explorador" />} />
              </Route>
              <Route path="/login" element={<LoginPage />} />
              <Route path="/cadastro" element={<CadastroPage />} />
              <Route path="/esqueci-senha" element={<EsqueciSenhaPage />} />
              <Route path="/redefinir-senha" element={<RedefinirSenhaPage />} />
              <Route path="*" element={<main className="p-8">Página não encontrada.</main>} />
            </Routes>
          </AbandonadosProvider>
        </JogosProvider>
      </AuthProvider>
    </BrowserRouter>
  )
}
