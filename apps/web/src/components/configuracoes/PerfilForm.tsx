import { useState } from 'react'
import { AlertCircle } from 'lucide-react'
import { usePerfilStore } from '@/stores/perfilStore'
import { PerfilInformacoesCard } from './PerfilInformacoesCard'
import { PerfilDestaquesCard } from './PerfilDestaquesCard'
import { PerfilBarraAcoes } from './PerfilBarraAcoes'
import { perfilSchema } from '@/lib/perfilSchema'
import { PERFIL_BACKEND_ERROS, ERRO_GENERICO_PERFIL } from '@/lib/perfilErros'
import { ApiError } from '@/lib/api'
import type { JogoFavoritoDTO } from '@/types/perfil'

export function PerfilForm() {
  const { perfil, isSalvando, salvar } = usePerfilStore()

  const [nome, setNome] = useState(perfil?.nome ?? '')
  const [username, setUsername] = useState(perfil?.username ?? '')
  const [bio, setBio] = useState(perfil?.bio ?? '')
  const [jogoFavorito, setJogoFavorito] = useState<JogoFavoritoDTO | null>(perfil?.jogo_favorito ?? null)
  const [consoleFavorito, setConsoleFavorito] = useState(perfil?.console_favorito ?? '')
  const [jogandoDesde, setJogandoDesde] = useState(perfil?.jogando_desde ? String(perfil.jogando_desde) : '')
  const [erros, setErros] = useState<Record<string, string>>({})
  const [erroGeral, setErroGeral] = useState<string | null>(null)

  function resetarValores() {
    setNome(perfil?.nome ?? '')
    setUsername(perfil?.username ?? '')
    setBio(perfil?.bio ?? '')
    setJogoFavorito(perfil?.jogo_favorito ?? null)
    setConsoleFavorito(perfil?.console_favorito ?? '')
    setJogandoDesde(perfil?.jogando_desde ? String(perfil.jogando_desde) : '')
    setErros({})
    setErroGeral(null)
  }


  const initialNome = perfil?.nome ?? ''
  const initialUsername = perfil?.username ?? ''
  const initialBio = perfil?.bio ?? ''
  const initialJogoId = perfil?.jogo_favorito?.id ?? null
  const initialConsole = perfil?.console_favorito ?? ''
  const initialAno = perfil?.jogando_desde_ano ? String(perfil.jogando_desde_ano) : ''
  const currentJogoId = jogoFavorito?.id ?? null

  const isDirty =
    nome !== initialNome ||
    username !== initialUsername ||
    bio !== initialBio ||
    currentJogoId !== initialJogoId ||
    consoleFavorito !== initialConsole ||
    jogandoDesde !== initialAno

  async function handleSalvar() {
    setErros({})
    setErroGeral(null)

    const payload = {
      nome: nome.trim(),
      username: username.trim() || null,
      bio: bio.trim() || null,
      jogo_favorito_id: currentJogoId,
      console_favorito: consoleFavorito.trim() || null,
      jogando_desde: jogandoDesde.trim() || null,
    }

    const validacao = perfilSchema.safeParse(payload)
    if (!validacao.success) {
      const novosErros: Record<string, string> = {}
      for (const issue of validacao.error.issues) {
        const campo = issue.path[0] as string
        if (!novosErros[campo]) novosErros[campo] = issue.message
      }
      setErros(novosErros)
      return
    }

    try {
      await salvar({
        nome: validacao.data.nome,
        username: validacao.data.username ?? null,
        bio: validacao.data.bio ?? null,
        jogo_favorito_id: validacao.data.jogo_favorito_id ?? null,
        console_favorito: validacao.data.console_favorito ?? null,
        jogando_desde: validacao.data.jogando_desde ?? null,
      })
    } catch (err) {
      if (err instanceof ApiError && err.codigo && PERFIL_BACKEND_ERROS[err.codigo]) {
        const mapeado = PERFIL_BACKEND_ERROS[err.codigo]
        if (mapeado.campo && mapeado.campo !== 'geral') {
          setErros({ [mapeado.campo]: mapeado.mensagem })
        } else {
          setErroGeral(mapeado.mensagem)
        }
      } else {
        setErroGeral(ERRO_GENERICO_PERFIL)
      }
    }
  }

  return (
    <div className="space-y-6">
      {erroGeral && (
        <div role="alert" className="flex items-center gap-2 rounded-[8px] border border-[var(--danger)]/30 bg-[var(--danger)]/10 px-3.5 py-2.5 text-xs text-[var(--danger)]">
          <AlertCircle className="h-4 w-4 shrink-0" aria-hidden="true" />
          <span>{erroGeral}</span>
        </div>
      )}

      <PerfilInformacoesCard
        nome={nome}
        username={username}
        bio={bio}
        erros={erros}
        onChangeNome={setNome}
        onChangeUsername={setUsername}
        onChangeBio={setBio}
      />

      <PerfilDestaquesCard
        jogoFavorito={jogoFavorito}
        consoleFavorito={consoleFavorito}
        jogandoDesde={jogandoDesde}
        jogandoDesdeEfetivo={perfil?.jogando_desde_efetivo}
        erros={erros}
        onSelectJogo={setJogoFavorito}
        onChangeConsole={setConsoleFavorito}
        onChangeJogandoDesde={setJogandoDesde}
      />

      <PerfilBarraAcoes
        isDirty={isDirty}
        isSalvando={isSalvando}
        onDescartar={resetarValores}
        onSalvar={handleSalvar}
      />
    </div>
  )
}
