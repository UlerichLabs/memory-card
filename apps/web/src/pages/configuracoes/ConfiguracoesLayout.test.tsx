import { render, screen, within } from '@testing-library/react'
import { MemoryRouter, Navigate, Route, Routes, useLocation } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { AuthProvider } from '@/store/authStore'
import { JogosProvider } from '@/stores/jogosStore'
import { JogandoProvider } from '@/stores/jogandoStore'
import { ConfiguracoesLayout } from './ConfiguracoesLayout'
import { PerfilSecao } from './PerfilSecao'
import { ContaSecao } from './ContaSecao'
import { AparenciaSecao } from './AparenciaSecao'
import { DadosSecao } from './DadosSecao'
import { SobreSecao } from './SobreSecao'
import { ContaTrocarSenhaPage } from '@/pages/ContaTrocarSenhaPage'

function LocationEcho() {
  const location = useLocation()
  return <div data-testid="location">{location.pathname}</div>
}

function renderConfiguracoes(initialPath: string) {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <AuthProvider>
        <JogosProvider>
          <JogandoProvider>
            <LocationEcho />
            <Routes>
              <Route path="/conta/trocar-senha" element={<ContaTrocarSenhaPage />} />
              <Route path="/configuracoes" element={<ConfiguracoesLayout />}>
                <Route index element={<Navigate to="/configuracoes/perfil" replace />} />
                <Route path="perfil" element={<PerfilSecao />} />
                <Route path="conta" element={<ContaSecao />} />
                <Route path="aparencia" element={<AparenciaSecao />} />
                <Route path="dados" element={<DadosSecao />} />
                <Route path="sobre" element={<SobreSecao />} />
                <Route path="*" element={<Navigate to="/configuracoes/perfil" replace />} />
              </Route>
            </Routes>
          </JogandoProvider>
        </JogosProvider>
      </AuthProvider>
    </MemoryRouter>
  )
}

describe('ConfiguracoesLayout e seções', () => {
  it('redireciona /configuracoes para /configuracoes/perfil', () => {
    renderConfiguracoes('/configuracoes')
    expect(screen.getByTestId('location')).toHaveTextContent('/configuracoes/perfil')
    expect(screen.getByRole('heading', { level: 2, name: 'Perfil' })).toBeInTheDocument()
  })

  it('redireciona rota desconhecida dentro de /configuracoes para /configuracoes/perfil', () => {
    renderConfiguracoes('/configuracoes/rota-inexistente')
    expect(screen.getByTestId('location')).toHaveTextContent('/configuracoes/perfil')
    expect(screen.getByRole('heading', { level: 2, name: 'Perfil' })).toBeInTheDocument()
  })

  it('redireciona /conta/trocar-senha para /configuracoes/conta', () => {
    renderConfiguracoes('/conta/trocar-senha')
    expect(screen.getByTestId('location')).toHaveTextContent('/configuracoes/conta')
    expect(screen.getByRole('heading', { level: 2, name: 'Trocar senha' })).toBeInTheDocument()
  })

  it('renderiza título e estado da seção Perfil', () => {
    renderConfiguracoes('/configuracoes/perfil')
    const secao = screen.getByRole('heading', { level: 2, name: 'Perfil' }).closest('section')!
    expect(secao).toBeInTheDocument()
    expect(within(secao).getByText('Em construção')).toBeInTheDocument()
  })

  it('renderiza título e TrocarSenhaForm na seção Conta', () => {
    renderConfiguracoes('/configuracoes/conta')
    const secao = screen.getByRole('heading', { level: 2, name: 'Trocar senha' }).closest('section')!
    expect(secao).toBeInTheDocument()
    expect(within(secao).getByLabelText(/Senha atual/i)).toBeInTheDocument()
    expect(within(secao).getByLabelText(/^Nova senha/i)).toBeInTheDocument()
    expect(within(secao).getByRole('button', { name: /Alterar senha/i })).toBeInTheDocument()
  })

  it('renderiza título e estado da seção Aparência', () => {
    renderConfiguracoes('/configuracoes/aparencia')
    const secao = screen.getByRole('heading', { level: 2, name: 'Aparência' }).closest('section')!
    expect(secao).toBeInTheDocument()
    expect(within(secao).getByText('Em construção')).toBeInTheDocument()
  })

  it('renderiza título e estado da seção Dados', () => {
    renderConfiguracoes('/configuracoes/dados')
    const secao = screen.getByRole('heading', { level: 2, name: 'Dados' }).closest('section')!
    expect(secao).toBeInTheDocument()
    expect(within(secao).getByText('Em breve')).toBeInTheDocument()
  })

  it('renderiza título, nome e aviso legal na seção Sobre o app', () => {
    renderConfiguracoes('/configuracoes/sobre')
    const secao = screen.getByRole('heading', { level: 2, name: 'Sobre o app' }).closest('section')!
    expect(secao).toBeInTheDocument()
    expect(within(secao).getByRole('heading', { level: 3, name: 'Memory Card' })).toBeInTheDocument()
    expect(
      within(secao).getByText(/Nomes e logotipos de consoles e plataformas pertencem aos seus respectivos donos/i)
    ).toBeInTheDocument()
  })
})
