import { useContext, useEffect, useRef, useState, type KeyboardEvent } from "react";
import { Search, Gamepad2 } from "lucide-react";
import { AuthContext } from "@/store/authStore";
import { jogosService, type IGDBJogoSugestao } from "@/lib/services/jogosService";
import { formatarCapaIGDB } from "@/lib/utils";
import { ApiError } from "@/lib/api";
import { resolverMensagemErro } from "./listas.constants";
import { versaoLabelIGDB } from "./listas.utils";
import type { AdicionarItemPayload } from "@/types/listas";

export interface AdicionarJogoInputProps {
  onAdicionar: (payload: AdicionarItemPayload) => Promise<unknown>;
  placeholder?: string;
  ariaLabel?: string;
}

export function AdicionarJogoInput({
  onAdicionar,
  placeholder = "Adicionar jogo: busque pelo nome…",
  ariaLabel = "Adicionar jogo à fila",
}: AdicionarJogoInputProps) {
  const token = useContext(AuthContext)?.sessao?.access_token;
  const [busca, setBusca] = useState("");
  const [sugestoes, setSugestoes] = useState<IGDBJogoSugestao[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const [erro, setErro] = useState<string | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const controllerRef = useRef<AbortController | null>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node))
        setIsOpen(false);
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleInputChange = (val: string) => {
    setBusca(val);
    setErro(null);
    if (timerRef.current) clearTimeout(timerRef.current);
    controllerRef.current?.abort();
    const termo = val.trim();
    if (termo.length < 2) {
      setSugestoes([]);
      setIsOpen(false);
      return;
    }
    const controller = new AbortController();
    controllerRef.current = controller;
    timerRef.current = setTimeout(async () => {
      setIsSearching(true);
      try {
        const res = await jogosService.buscarIGDB(termo, token, controller.signal);
        if (!controller.signal.aborted) {
          setSugestoes(res || []);
          setIsOpen(true);
        }
      } catch {
        if (!controller.signal.aborted) setSugestoes([]);
      } finally {
        if (!controller.signal.aborted) setIsSearching(false);
      }
    }, 300);
  };

  const enviarPayload = async (payload: AdicionarItemPayload) => {
    setErro(null);
    try {
      await onAdicionar(payload);
      setBusca("");
      setSugestoes([]);
      setIsOpen(false);
    } catch (err: unknown) {
      setErro(
        err instanceof ApiError
          ? resolverMensagemErro(err.codigo)
          : "Não foi possível adicionar o jogo.",
      );
    }
  };

  const handleSelectSugestao = (sugestao: IGDBJogoSugestao) => {
    const ano = sugestao.first_release_date
      ? new Date(sugestao.first_release_date * 1000).getFullYear()
      : null;
    enviarPayload({
      igdb_id: sugestao.id,
      nome: sugestao.name,
      igdb_capa_url: sugestao.cover?.url ? formatarCapaIGDB(sugestao.cover.url) : null,
      ano_lancamento: ano,
      console: sugestao.platforms?.[0]?.name ?? null,
    });
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      e.preventDefault();
      const nomeTrim = busca.trim();
      if (!nomeTrim || nomeTrim.length > 200) return;
      enviarPayload({ nome: nomeTrim });
    }
  };

  return (
    <div
      ref={containerRef}
      className="relative flex flex-col gap-1.5 w-full"
    >
      <div className="relative flex items-center">
        <Search className="pointer-events-none absolute left-3.5 h-[18px] w-[18px] text-[var(--lista-text-muted)]" />
        <input
          value={busca}
          onChange={(e) => handleInputChange(e.target.value)}
          onKeyDown={handleKeyDown}
          onFocus={() => {
            if (sugestoes.length) setIsOpen(true);
          }}
          placeholder={placeholder}
          aria-label={ariaLabel}
          className={[
          "h-12 w-full rounded-[10px] border",
          "border-[var(--lista-input-border)] bg-[var(--lista-input-bg)] pl-[42px] pr-4",
          "text-[15px] text-[var(--text-primary)] placeholder:text-[var(--lista-text-muted)] focus:outline-none"
        ].join(" ")}
        />
        {isSearching && (
          <span className="absolute right-3.5 text-xs text-[var(--lista-text-muted)]">
            Buscando...
          </span>
        )}
      </div>

      {erro && <span className="text-xs text-[var(--danger)]">{erro}</span>}

      {isOpen && (
        <ul
          role="listbox"
          aria-label="Sugestões de jogos"
          className={[
          "custom-scrollbar absolute left-0 right-0",
          "top-full z-50 mt-1 max-h-60",
          "overflow-y-auto rounded-lg border border-[var(--border)]",
          "bg-[var(--bg-surface)] p-1 shadow-2xl"
        ].join(" ")}
        >
          {isSearching ? (
            <li className="py-3 text-center text-xs text-[var(--text-muted)]">Buscando...</li>
          ) : sugestoes.length === 0 ? (
            <li className="py-3 text-center text-xs text-[var(--text-muted)]">
              Nenhum jogo encontrado
            </li>
          ) : (
            sugestoes.map((sugestao) => {
              const capa = formatarCapaIGDB(sugestao.cover?.url);
              const ano = sugestao.first_release_date
                ? new Date(sugestao.first_release_date * 1000).getFullYear()
                : null;
              const metaStr = versaoLabelIGDB(ano, sugestao.platforms);
              return (
                <li
                  key={sugestao.id}
                  role="option"
                  aria-selected={false}
                  onClick={() => handleSelectSugestao(sugestao)}
                  className={[
          "flex cursor-pointer items-center gap-3",
          "rounded-md px-3 py-2 text-sm",
          "text-[var(--text-primary)] hover:bg-[var(--bg-surface-alt)]"
        ].join(" ")}
                >
                  {capa ? (
                    <img
                      src={capa}
                      alt=""
                      className="h-11 w-8 shrink-0 rounded border border-[var(--border)] object-cover"
                    />
                  ) : (
                    <div className={[
          "flex h-11 w-8 shrink-0",
          "items-center justify-center rounded border",
          "border-[var(--border)] text-[var(--text-muted)]"
        ].join(" ")}>
                      <Gamepad2 className="h-4 w-4" />
                    </div>
                  )}
                  <span className="min-w-0 flex-1">
                    <span className="block truncate font-semibold">{sugestao.name}</span>
                    {metaStr && (
                      <span className="block truncate text-xs text-[var(--text-muted)]">
                        {metaStr}
                      </span>
                    )}
                  </span>
                </li>
              );
            })
          )}
        </ul>
      )}
    </div>
  );
}
