import { useEffect, useMemo, useRef, useState, type KeyboardEvent } from "react";
import { Search } from "lucide-react";
import { listasService } from "@/lib/services/listasService";
import { normalizarTexto } from "./listas.utils";
import type { ListaOrigem, OpcaoOrigem } from "@/types/listas";

const cache: Partial<Record<"plataforma" | "genero", OpcaoOrigem[]>> = {};

interface OrigemAutocompleteProps {
  tipo: "franquia" | "plataforma" | "genero";
  valor: ListaOrigem | null;
  token?: string;
  erro?: string;
  onChange: (origem: ListaOrigem | null) => void;
}

const placeholders = {
  franquia: "Ex: Castlevania, Final Fantasy",
  plataforma: "Ex: Super Nintendo, Mega Drive",
  genero: "Ex: RPG, Plataforma, Luta",
};

export function OrigemAutocomplete({
  tipo,
  valor,
  token,
  erro,
  onChange,
}: OrigemAutocompleteProps) {
  const [texto, setTexto] = useState(valor?.nome ?? "");
  const [opcoes, setOpcoes] = useState<OpcaoOrigem[]>([]);
  const [aberto, setAberto] = useState(false);
  const [carregando, setCarregando] = useState(false);
  const [indice, setIndice] = useState(0);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const controller = useRef<AbortController | null>(null);

  useEffect(() => {
    setTexto(valor?.nome ?? "");
  }, [valor]);
  useEffect(() => {
    if (tipo === "franquia") return;
    if (cache[tipo]) {
      setOpcoes(cache[tipo] ?? []);
      return;
    }
    setCarregando(true);
    const busca = tipo === "genero" ? listasService.listarGeneros : listasService.listarPlataformas;
    busca(token)
      .then((resultado) => {
        cache[tipo] = resultado;
        setOpcoes(resultado);
      })
      .catch(() => setOpcoes([]))
      .finally(() => setCarregando(false));
  }, [tipo, token]);

  const filtradas = useMemo(
    () =>
      opcoes
        .filter((item) => normalizarTexto(item.nome).includes(normalizarTexto(texto)))
        .slice(0, 20),
    [opcoes, texto],
  );
  const buscar = (novoTexto: string) => {
    setTexto(novoTexto);
    onChange(null);
    setAberto(true);
    setIndice(0);
    if (tipo !== "franquia") return;
    if (timer.current) clearTimeout(timer.current);
    controller.current?.abort();
    if (novoTexto.trim().length < 2) {
      setOpcoes([]);
      return;
    }
    const atual = new AbortController();
    controller.current = atual;
    timer.current = setTimeout(() => {
      setCarregando(true);
      listasService
        .buscarFranquias(novoTexto.trim(), token, atual.signal)
        .then((items) => setOpcoes(items.map((item) => ({ id: item.id, nome: item.name }))))
        .catch(() => setOpcoes([]))
        .finally(() => setCarregando(false));
    }, 300);
  };
  const escolher = (item: OpcaoOrigem) => {
    onChange({ tipo, igdb_id: item.id, nome: item.nome });
    setTexto(item.nome);
    setAberto(false);
  };
  const teclado = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "ArrowDown") {
      event.preventDefault();
      setIndice((valorAtual) => Math.min(valorAtual + 1, filtradas.length - 1));
    }
    if (event.key === "ArrowUp") {
      event.preventDefault();
      setIndice((valorAtual) => Math.max(valorAtual - 1, 0));
    }
    if (event.key === "Enter" && filtradas[indice]) {
      event.preventDefault();
      escolher(filtradas[indice]);
    }
    if (event.key === "Escape") setAberto(false);
  };

  return (
    <div className="flex min-w-0 flex-col gap-1.5">
      <label
        htmlFor="origem-valor"
        className="text-[13px] font-medium text-[var(--lista-text-light)]"
      >
        {tipo[0].toUpperCase() + tipo.slice(1)} *
      </label>
      <div className="relative flex min-w-0 items-center">
        <Search className="pointer-events-none absolute left-3 h-4 w-4 text-[var(--lista-text-muted)]" />
        <input
          id="origem-valor"
          role="combobox"
          aria-expanded={aberto}
          aria-controls="origem-opcoes"
          value={texto}
          placeholder={placeholders[tipo]}
          onChange={(event) => buscar(event.target.value)}
          onFocus={() => setAberto(true)}
          onKeyDown={teclado}
          autoComplete="off"
          className={["h-11 w-full min-w-0 rounded-[10px] border",
  "border-[var(--lista-input-border)] bg-[var(--lista-input-bg)]",
  "pl-9 pr-3 text-[14px] text-[var(--text-primary)] focus:outline-none"].join(" ")}
        />
        {carregando && (
          <span className="absolute right-3 text-xs text-[var(--lista-text-muted)]">
            Buscando...
          </span>
        )}
      </div>
      {aberto && (filtradas.length > 0 || carregando) && (
        <ul
          id="origem-opcoes"
          role="listbox"
          className={["custom-scrollbar max-h-48 overflow-y-auto rounded-lg border",
  "border-[var(--border)] bg-[var(--bg-surface)] p-1"].join(" ")}
        >
          {filtradas.map((item, itemIndex) => (
            <li
              key={item.id}
              role="option"
              aria-selected={itemIndex === indice}
            >
              <button
                type="button"
                onMouseDown={(event) => event.preventDefault()}
                onClick={() => escolher(item)}
                className={[
                  "w-full rounded px-3 py-2 text-left text-sm",
                  "text-[var(--text-primary)] hover:bg-[var(--bg-surface-alt)]",
                ].join(" ")}
              >
                {item.nome}
              </button>
            </li>
          ))}
        </ul>
      )}
      {erro && <span className="text-xs text-[var(--danger)]">{erro}</span>}
    </div>
  );
}
