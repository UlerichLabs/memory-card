interface PerfilInformacoesCardProps {
  nome: string
  username: string
  bio: string
  erros: Record<string, string>
  onChangeNome: (val: string) => void
  onChangeUsername: (val: string) => void
  onChangeBio: (val: string) => void
}

export function PerfilInformacoesCard({
  nome,
  username,
  bio,
  erros,
  onChangeNome,
  onChangeUsername,
  onChangeBio,
}: PerfilInformacoesCardProps) {
  return (
    <section className="space-y-4 rounded-[12px] border border-[var(--border)] bg-[var(--bg-surface)] p-5 sm:p-6">
      <header>
        <h3 className="text-sm font-bold text-[var(--text-primary)]">Informações</h3>
      </header>

      <div className="space-y-4">
        <div className="space-y-1.5">
          <label htmlFor="perfil-nome" className="text-xs font-semibold text-[var(--text-secondary)]">
            Nome de exibição *
          </label>
          <input
            id="perfil-nome"
            type="text"
            maxLength={100}
            value={nome}
            onChange={(e) => onChangeNome(e.target.value)}
            placeholder="Seu nome"
            aria-invalid={Boolean(erros.nome)}
            className="h-10 w-full rounded-[7px] border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] px-3 text-xs text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:border-[var(--accent)] focus:outline-none"
          />
          {erros.nome && <p className="text-xs text-[var(--danger)]">{erros.nome}</p>}
        </div>

        <div className="space-y-1.5">
          <label htmlFor="perfil-username" className="text-xs font-semibold text-[var(--text-secondary)]">
            Nome de usuário
          </label>
          <div className="relative flex items-center">
            <span className="pointer-events-none absolute left-3 text-xs font-semibold text-[var(--text-muted)]" aria-hidden="true">
              @
            </span>
            <input
              id="perfil-username"
              type="text"
              maxLength={30}
              value={username}
              onChange={(e) => onChangeUsername(e.target.value.replace(/^@/, ''))}
              placeholder="seu_usuario"
              aria-invalid={Boolean(erros.username)}
              className="h-10 w-full rounded-[7px] border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] pl-7 pr-3 text-xs text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:border-[var(--accent)] focus:outline-none"
            />
          </div>
          <p className="text-[11px] text-[var(--text-muted)]">
            Será o endereço do seu perfil. Letras, números e _.
          </p>
          {erros.username && <p className="text-xs text-[var(--danger)]">{erros.username}</p>}
        </div>

        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <label htmlFor="perfil-bio" className="text-xs font-semibold text-[var(--text-secondary)]">
              Sobre você
            </label>
            <span className="text-[11px] text-[var(--text-muted)]">
              {bio.length}/280
            </span>
          </div>
          <textarea
            id="perfil-bio"
            rows={3}
            maxLength={280}
            value={bio}
            onChange={(e) => onChangeBio(e.target.value)}
            placeholder="Conte um pouco sobre seu estilo como jogador…"
            aria-invalid={Boolean(erros.bio)}
            className="w-full rounded-[7px] border border-[var(--border-subtle)] bg-[var(--bg-surface-alt)] p-3 text-xs text-[var(--text-primary)] placeholder:text-[var(--text-muted)] focus:border-[var(--accent)] focus:outline-none resize-none"
          />
          <p className="text-[11px] text-[var(--text-muted)]">
            Aparece no topo do seu perfil.
          </p>
          {erros.bio && <p className="text-xs text-[var(--danger)]">{erros.bio}</p>}
        </div>
      </div>
    </section>
  )
}
