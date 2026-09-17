import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { HomePage } from '@/pages/HomePage'
import { CadastroPage } from '@/pages/auth/CadastroPage'

export function AppRoutes() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/cadastro" element={<CadastroPage />} />
        <Route path="*" element={<main className="p-8">Página não encontrada.</main>} />
      </Routes>
    </BrowserRouter>
  )
}
