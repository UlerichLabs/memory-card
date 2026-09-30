interface Props {
  carregados: number;
  total: number;
  totalSugeridos: number | null;
  totalTodos: number | null;
  mostrarSugeridos: boolean;
  somenteSugeridos: boolean | undefined;
  marcando: boolean;
  onMarcarSugeridos: () => void;
  onMarcarVisiveis: () => void;
  onAlternarSugeridos: () => void;
}

export function EscolherJogosBarra({
  carregados,
  total,
  totalSugeridos,
  totalTodos,
  mostrarSugeridos,
  somenteSugeridos,
  marcando,
  onMarcarSugeridos,
  onMarcarVisiveis,
  onAlternarSugeridos,
}: Props) {
  const mostrarAlternativa =
    mostrarSugeridos &&
    totalTodos !== null &&
    totalSugeridos !== null &&
    totalTodos > totalSugeridos;
  return (
    <div
      className="flex shrink-0 items-center justify-between gap-3 p-[12px_28px]
        text-[13px] text-[var(--lista-text-muted)]"
    >
      <span className="tabular-nums">
        Mostrando {carregados} de {total} jogos
      </span>
      <div className="flex flex-wrap justify-end gap-2">
        {mostrarAlternativa && (
          <button
            type="button"
            onClick={onAlternarSugeridos}
            className="h-8 rounded-lg border border-[var(--lista-btn-icon-border)] px-3
              text-[12px] text-[var(--lista-text-light)]"
          >
            {somenteSugeridos
              ? `Mostrar outros jogos ligados à franquia (${totalTodos - totalSugeridos})`
              : "Mostrar só os da franquia"}
          </button>
        )}
        {mostrarSugeridos && totalSugeridos && totalSugeridos > 0 && (
          <button
            type="button"
            onClick={onMarcarSugeridos}
            disabled={marcando}
            className="h-8 rounded-lg border border-[var(--lista-pill-desafio-border)]
              bg-[var(--lista-pill-desafio-bg)] px-3 text-[12px] font-semibold
              text-[var(--hall-ouro)] disabled:opacity-50"
          >
            {marcando ? "Marcando..." : `Marcar os ${totalSugeridos} sugeridos`}
          </button>
        )}
        <button
          type="button"
          onClick={onMarcarVisiveis}
          className="h-8 rounded-lg border border-[var(--lista-btn-icon-border)] px-3
            text-[12px] text-[var(--lista-text-light)]"
        >
          Marcar visíveis
        </button>
      </div>
    </div>
  );
}
