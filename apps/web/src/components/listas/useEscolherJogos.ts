import { useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react'
import { AuthContext } from '@/store/authStore'
import { listasService } from '@/lib/services/listasService'
import { resolverMensagemErro } from './listas.constants'
import type { CatalogoItem, FiltroCatalogo, ListaOrigem, ListaItem, OrdenarCatalogo, CriarListaItemPayload } from '@/types/listas'

export interface EscolherJogosConfig { nome: string; descricao: string | null; origem: ListaOrigem }
export interface UseEscolherJogosOptions { origem: ListaOrigem; existentes?: ListaItem[]; config?: EscolherJogosConfig; modo: 'criar' | 'adicionar' }

export function useEscolherJogos({ origem, existentes = [], modo }: UseEscolherJogosOptions) {
  const token = useContext(AuthContext)?.sessao?.access_token
  const [itens, setItens] = useState<CatalogoItem[]>([]); const [selecionados, setSelecionados] = useState<Map<number, CriarListaItemPayload>>(new Map())
  const [pagina, setPagina] = useState(1); const [meta, setMeta] = useState({ pagina: 1, por_pagina: 60, total: 0, total_sugeridos: null as number | null }); const [carregando, setCarregando] = useState(true); const [erro, setErro] = useState<string | null>(null)
  const [busca, setBusca] = useState(''); const [generoId, setGeneroId] = useState<number | undefined>(); const [plataformaId, setPlataformaId] = useState<number | undefined>(); const [ordenar, setOrdenar] = useState<OrdenarCatalogo>('populares'); const controller = useRef<AbortController | null>(null)
  const idsExistentes = useMemo(() => new Set(existentes.map((item) => item.igdb_id).filter((id): id is number => id !== null)), [existentes])
  const carregar = useCallback(async (page: number, substituir: boolean): Promise<CatalogoItem[]> => { controller.current?.abort(); const atual = new AbortController(); controller.current = atual; setCarregando(true); setErro(null)
    try { const filtro: FiltroCatalogo = { origem: origem.tipo, id: origem.igdb_id ?? 0, busca, ordenar, pagina: page, por_pagina: 60 }; if (generoId) filtro.genero_id = generoId; if (plataformaId) filtro.plataforma_id = plataformaId; const resposta = await listasService.buscarCatalogo(filtro, token, atual.signal); if (!atual.signal.aborted) { setItens((anteriores) => substituir ? resposta.itens : [...anteriores, ...resposta.itens.filter((item) => !anteriores.some((anterior) => anterior.igdb_id === item.igdb_id))]); setMeta(resposta.meta); setPagina(page) } return resposta.itens } catch (caught: unknown) { if (!atual.signal.aborted) setErro(caught instanceof Error && 'codigo' in caught ? resolverMensagemErro(String(caught.codigo)) : 'Não foi possível carregar os jogos agora. Tente de novo.'); return [] } finally { if (!atual.signal.aborted) setCarregando(false) }
  }, [busca, generoId, origem.igdb_id, origem.tipo, plataformaId, ordenar, token])
  useEffect(() => { void carregar(1, true); return () => controller.current?.abort() }, [carregar])
  const atualizarFiltro = (atualizador: () => void) => { atualizador(); setPagina(1) }
  const alternar = (item: CatalogoItem) => { if (modo === 'adicionar' && idsExistentes.has(item.igdb_id)) return; setSelecionados((anteriores) => { const novo = new Map(anteriores); if (novo.has(item.igdb_id)) novo.delete(item.igdb_id); else novo.set(item.igdb_id, { igdb_id: item.igdb_id, nome: item.nome, igdb_capa_url: item.igdb_capa_url, ano_lancamento: item.ano_lancamento }); return novo }) }
  const marcarSugeridos = async () => { const todos = [...itens]; const totalPaginas = Math.ceil(meta.total / meta.por_pagina); for (let proxima = pagina + 1; proxima <= totalPaginas; proxima += 1) todos.push(...await carregar(proxima, false)); setSelecionados((anteriores) => { const novo = new Map(anteriores); todos.forEach((item) => { if (item.sugerido && !idsExistentes.has(item.igdb_id)) novo.set(item.igdb_id, { igdb_id: item.igdb_id, nome: item.nome, igdb_capa_url: item.igdb_capa_url, ano_lancamento: item.ano_lancamento }) }); return novo }) }
  const payload = [...selecionados.values()].filter((item) => !idsExistentes.has(item.igdb_id));
  return { token, itens, selecionados, payload, idsExistentes, meta, pagina, carregando, erro, busca, generoId, plataformaId, ordenar, setBusca: (valor: string) => atualizarFiltro(() => setBusca(valor)), setGeneroId: (valor?: number) => atualizarFiltro(() => setGeneroId(valor)), setPlataformaId: (valor?: number) => atualizarFiltro(() => setPlataformaId(valor)), setOrdenar: (valor: OrdenarCatalogo) => atualizarFiltro(() => setOrdenar(valor)), alternar, carregarMais: () => carregar(pagina + 1, false), recarregar: () => carregar(1, true), marcarSugeridos }
}
