import { useContext, useEffect, useRef, useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useListasStore } from '@/stores/listasStore'
import { listasService } from '@/lib/services/listasService'
import { jogosService } from '@/lib/services/jogosService'
import { AuthContext } from '@/store/authStore'
import { ApiError } from '@/lib/api'
import { mapearErroApiParaCampo } from './listas.utils'
import { novaListaSchema } from './novaLista.schema'
import type { ListaTipo, RegraTipo, IGDBFranquiaSugestao } from '@/types/listas'

export function useNovaListaForm(onClose: () => void) {
  const navigate = useNavigate()
  const token = useContext(AuthContext)?.sessao?.access_token
  const { listaEmEdicao, criarLista, atualizarLista } = useListasStore()
  const [tipo, setTipo] = useState<ListaTipo>(() => listaEmEdicao?.tipo ?? 'fila')
  const [nome, setNome] = useState(() => listaEmEdicao?.nome ?? '')
  const [descricao, setDescricao] = useState(() => listaEmEdicao?.descricao ?? '')
  const [regraTipo, setRegraTipo] = useState<RegraTipo>(() => listaEmEdicao?.regra?.tipo ?? 'franquia')
  const [regraValor, setRegraValor] = useState(() => listaEmEdicao?.regra?.valor ?? '')
  const [igdbId, setIgdbId] = useState<number | null>(() => listaEmEdicao?.regra?.igdb_id ?? null)
  const [meta, setMeta] = useState(() => listaEmEdicao?.meta ? String(listaEmEdicao.meta) : '')
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [franquias, setFranquias] = useState<IGDBFranquiaSugestao[]>([])
  const [isSearchingFranquias, setIsSearchingFranquias] = useState(false)
  const [filtros, setFiltros] = useState<{ p: string[]; g: string[] }>({ p: [], g: [] })
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    jogosService.obterFiltros(token).then((res) => setFiltros({ p: res.consoles, g: res.generos })).catch(() => undefined)
  }, [token])

  const buscarFranquias = (termo: string) => {
    setRegraValor(termo); setIgdbId(null)
    if (timerRef.current) clearTimeout(timerRef.current)
    if (termo.trim().length < 2) { setFranquias([]); return }
    timerRef.current = setTimeout(async () => {
      setIsSearchingFranquias(true)
      try { setFranquias(await listasService.buscarFranquiasIGDB(termo.trim(), token)) }
      catch { setFranquias([]) }
      finally { setIsSearchingFranquias(false) }
    }, 300)
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault(); setErrors({})
    const metaNum = meta ? parseInt(meta, 10) : undefined
    const parsed = novaListaSchema.safeParse({ tipo, nome, descricao: descricao || undefined, regraTipo: tipo === 'desafio' ? regraTipo : undefined, regraValor: tipo === 'desafio' ? regraValor : undefined, igdbId: tipo === 'desafio' ? igdbId : undefined, meta: metaNum })
    if (!parsed.success) {
      const errMap: Record<string, string> = {}
      parsed.error.issues.forEach((it) => { errMap[it.path[0] as string] = it.message })
      setErrors(errMap); return
    }
    setIsSubmitting(true)
    try {
      if (listaEmEdicao) {
        await atualizarLista(listaEmEdicao.id, { nome: parsed.data.nome, descricao: parsed.data.descricao ?? null, meta: parsed.data.meta ?? null })
        onClose()
      } else {
        const regra = tipo === 'desafio' ? { tipo: regraTipo, valor: regraTipo === 'manual' ? null : regraValor.trim(), igdb_id: regraTipo === 'franquia' ? igdbId : null } : null
        const criada = await criarLista({ tipo, nome: parsed.data.nome, descricao: parsed.data.descricao ?? null, regra, meta: parsed.data.meta ?? null })
        onClose(); navigate(`/listas/${criada.id}`)
      }
    } catch (err: unknown) {
      const { campo, mensagem } = err instanceof ApiError ? mapearErroApiParaCampo(err.codigo) : { campo: 'form', mensagem: 'Erro ao salvar a lista.' }
      setErrors((p) => ({ ...p, [campo]: mensagem }))
    } finally {
      setIsSubmitting(false)
    }
  }

  return {
    tipo, setTipo, nome, setNome, descricao, setDescricao, regraTipo, setRegraTipo, regraValor, setRegraValor,
    igdbId, setIgdbId, meta, setMeta, errors, isSubmitting, franquias, setFranquias, isSearchingFranquias,
    plataformas: filtros.p, generos: filtros.g, buscarFranquias, handleSubmit, isEdicao: !!listaEmEdicao,
  }
}
