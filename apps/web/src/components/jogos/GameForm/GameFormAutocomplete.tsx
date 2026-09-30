import { useContext, useEffect, useRef, useState } from "react";
import { Gamepad2 } from "lucide-react";
import { jogosService, JogosApiError, type IGDBJogoSugestao } from "@/lib/services/jogosService";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { AuthContext } from "@/store/authStore";
import { formatarCapaIGDB } from "@/lib/utils";

export interface GameFormAutocompleteProps {
  nome: string;
  onChangeNome: (nome: string) => void;
  onSelectSugestao: (sugestao: IGDBJogoSugestao) => void;
  error?: string;
}

function versaoLabel(sugestao: IGDBJogoSugestao) {
  const ano = sugestao.first_release_date
    ? new Date(sugestao.first_release_date * 1000).getFullYear()
    : null;
  const plataformas = sugestao.platforms?.map((item) => item.name).join(", ");
  return [ano, plataformas].filter(Boolean).join(" · ");
}

export function GameFormAutocomplete({
  nome,
  onChangeNome,
  onSelectSugestao,
  error,
}: GameFormAutocompleteProps) {
  const auth = useContext(AuthContext);
  const token = auth?.sessao?.access_token;
  const [sugestoes, setSugestoes] = useState<IGDBJogoSugestao[]>([]);
  const [selecionado, setSelecionado] = useState<IGDBJogoSugestao | null>(null);
  const [isOpen, setIsOpen] = useState(false);
  const [isSearching, setIsSearching] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const ignorarBusca = useRef(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const controllerRef = useRef<AbortController | null>(null);

  function handleChange(valor: string) {
    ignorarBusca.current = false;
    setSelecionado(null);
    onChangeNome(valor);
    if (valor.trim().length < 2) {
      setSugestoes([]);
      setIsOpen(false);
    }
  }

  useEffect(() => {
    if (ignorarBusca.current) {
      ignorarBusca.current = false;
      return;
    }
    const termo = nome.trim();
    if (termo.length < 2) return;
    const controller = new AbortController();
    controllerRef.current = controller;
    const timer = setTimeout(async () => {
      setIsSearching(true);
      try {
        const resultado = await jogosService.buscarIGDB(termo, token, controller.signal);
        if (!controller.signal.aborted) {
          setSugestoes(resultado || []);
          setIsOpen(true);
        }
      } catch (error: unknown) {
        if (error instanceof JogosApiError && error.status === 401)
          window.location.assign("/login");
        if (!controller.signal.aborted) setSugestoes([]);
      } finally {
        if (!controller.signal.aborted) setIsSearching(false);
      }
    }, 350);
    timerRef.current = timer;
    return () => {
      clearTimeout(timer);
      controller.abort();
      if (timerRef.current === timer) timerRef.current = null;
    };
  }, [nome, token]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node))
        setIsOpen(false);
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  return (
    <div
      ref={containerRef}
      className="relative flex flex-col gap-2"
    >
      <Label
        htmlFor="nome"
        className="text-sm font-medium text-[var(--text-secondary)]"
      >
        Nome do jogo *
      </Label>
      <div className="relative">
        <Input
          id="nome"
          name="nome"
          maxLength={200}
          value={nome}
          onChange={(event) => handleChange(event.target.value)}
          onFocus={() => {
            if (sugestoes.length) setIsOpen(true);
          }}
          placeholder="Ex: God of War, Chrono Trigger"
          aria-invalid={!!error}
          autoComplete="off"
          className={`h-12 bg-[var(--bg-surface-alt)] text-[var(--text-primary)] ${selecionado ? "pr-48" : "pr-3"}`}
        />
        {isSearching && (
          <span className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-[var(--text-muted)]">
            Buscando...
          </span>
        )}
        {selecionado && (
          <span className={[
          "pointer-events-none absolute right-3 top-1/2",
          "max-w-[calc(100%-1.5rem)] -translate-y-1/2 truncate rounded-full",
          "border border-[var(--chip-border)] bg-[var(--chip-bg)] px-3",
          "py-1 text-xs text-[var(--chip-text)]"
        ].join(" ")}>
            {versaoLabel(selecionado) || selecionado.name}
          </span>
        )}
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
                return (
                  <li
                    key={sugestao.id}
                    role="option"
                    aria-selected={false}
                    onClick={() => {
                      ignorarBusca.current = true;
                      if (timerRef.current) clearTimeout(timerRef.current);
                      controllerRef.current?.abort();
                      setSelecionado(sugestao);
                      setIsOpen(false);
                      onSelectSugestao(sugestao);
                    }}
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
                        <Gamepad2 className="size-4" />
                      </div>
                    )}
                    <span className="min-w-0 flex-1">
                      <span className="block truncate font-semibold">{sugestao.name}</span>
                      {versaoLabel(sugestao) && (
                        <span className="block truncate text-xs text-[var(--text-muted)]">
                          {versaoLabel(sugestao)}
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
      {error && <span className="text-xs text-[var(--danger)]">{error}</span>}
    </div>
  );
}
