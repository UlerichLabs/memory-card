import { render, screen, within } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'
import { ConfiguracoesNav } from './ConfiguracoesNav'

describe('ConfiguracoesNav', () => {
  it('renderiza todos os itens no desktop com a ordem e links esperados', () => {
    render(
      <MemoryRouter initialEntries={['/configuracoes/perfil']}>
        <ConfiguracoesNav />
      </MemoryRouter>
    )

    const navDesktop = screen.getByRole('navigation', { name: 'Menu lateral de configurações' })
    expect(navDesktop).toHaveClass('hidden', 'md:flex')

    expect(within(navDesktop).getByRole('link', { name: /^Perfil$/i })).toHaveAttribute('href', '/configuracoes/perfil')
    expect(within(navDesktop).getByRole('link', { name: /Conta e segurança/i })).toHaveAttribute('href', '/configuracoes/conta')
    expect(within(navDesktop).getByRole('link', { name: /Aparência/i })).toHaveAttribute('href', '/configuracoes/aparencia')
    expect(within(navDesktop).getByRole('link', { name: /Dados/i })).toHaveAttribute('href', '/configuracoes/dados')
    expect(within(navDesktop).getByRole('link', { name: /Sobre o app/i })).toHaveAttribute('href', '/configuracoes/sobre')
  })

  it('mantém Preferências desabilitado e sem link', () => {
    render(
      <MemoryRouter initialEntries={['/configuracoes/perfil']}>
        <ConfiguracoesNav />
      </MemoryRouter>
    )

    expect(screen.queryByRole('link', { name: /Preferências/i })).not.toBeInTheDocument()
    const botaoPrefs = screen.getByRole('button', { name: /Preferências/i })
    expect(botaoPrefs).toBeDisabled()
    expect(botaoPrefs).toHaveTextContent('Em breve')
  })

  it('destaca o item ativo conforme a rota atual', () => {
    render(
      <MemoryRouter initialEntries={['/configuracoes/conta']}>
        <ConfiguracoesNav />
      </MemoryRouter>
    )

    const navDesktop = screen.getByRole('navigation', { name: 'Menu lateral de configurações' })
    const linkConta = within(navDesktop).getByRole('link', { name: /Conta e segurança/i })
    expect(linkConta).toHaveClass('text-[var(--accent)]')

    const linkPerfil = within(navDesktop).getByRole('link', { name: /^Perfil$/i })
    expect(linkPerfil).not.toHaveClass('text-[var(--accent)]')
  })

  it('renderiza chips mobile sem overflow-x-auto, com flex-wrap e com os 5 links', () => {
    render(
      <MemoryRouter initialEntries={['/configuracoes/dados']}>
        <ConfiguracoesNav />
      </MemoryRouter>
    )

    const navMobile = screen.getByRole('navigation', { name: 'Navegação mobile de configurações' })
    expect(navMobile).toHaveClass('md:hidden', 'flex', 'flex-wrap')
    expect(navMobile).not.toHaveClass('overflow-x-auto')

    const links = within(navMobile).getAllByRole('link')
    expect(links).toHaveLength(5)
    expect(within(navMobile).getByRole('link', { name: 'Perfil' })).toHaveAttribute('href', '/configuracoes/perfil')
    expect(within(navMobile).getByRole('link', { name: 'Conta' })).toHaveAttribute('href', '/configuracoes/conta')
    expect(within(navMobile).getByRole('link', { name: 'Aparência' })).toHaveAttribute('href', '/configuracoes/aparencia')
    expect(within(navMobile).getByRole('link', { name: 'Dados' })).toHaveAttribute('href', '/configuracoes/dados')
    expect(within(navMobile).getByRole('link', { name: 'Sobre' })).toHaveAttribute('href', '/configuracoes/sobre')

    const chipDados = within(navMobile).getByRole('link', { name: 'Dados' })
    expect(chipDados).toHaveClass('text-[var(--accent)]')
    expect(chipDados).toHaveClass('min-h-11')
  })
})
