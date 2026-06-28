import { lazy, Suspense } from "react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { GameFormModal } from "@/components/GameFormModal";
import { PrivateRoute } from "@/components/layout/PrivateRoute";
import { AuthProvider } from "@/context/AuthContext";

const Login = lazy(() => import("@/pages/login"));
const Cadastro = lazy(() => import("@/pages/cadastro"));
const Biblioteca = lazy(() => import("@/pages/biblioteca/BibliotecaPage"));
const NovoJogo = lazy(() => import("@/pages/biblioteca/NovoJogoPage"));
const DetalheJogo = lazy(() => import("@/pages/biblioteca/DetalheJogoPage"));
const EditarJogo = lazy(() => import("@/pages/biblioteca/EditarJogoPage"));
const Dashboard = lazy(() => import("@/pages/DashboardPage"));
const Desafios = lazy(() => import("@/pages/desafios"));

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Suspense fallback={null}>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/cadastro" element={<Cadastro />} />
            <Route element={<PrivateRoute />}>
              <Route path="/" element={<Navigate to="/biblioteca" replace />} />
              <Route path="/biblioteca" element={<Biblioteca />} />
              <Route path="/biblioteca/novo" element={<NovoJogo />} />
              <Route path="/biblioteca/:id" element={<DetalheJogo />} />
              <Route path="/biblioteca/:id/editar" element={<EditarJogo />} />
              <Route path="/dashboard" element={<Dashboard />} />
              <Route path="/desafios" element={<Desafios />} />
            </Route>
          </Routes>
          <GameFormModal />
        </Suspense>
      </AuthProvider>
    </BrowserRouter>
  );
}
