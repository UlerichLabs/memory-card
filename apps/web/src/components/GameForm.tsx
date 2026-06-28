import { useEffect, useMemo, useRef, useState } from "react";
import type { InputHTMLAttributes, ReactNode } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, useForm } from "react-hook-form";
import { CalendarDays, Check, ImageOff } from "lucide-react";
import { Link } from "react-router-dom";
import { z } from "zod";
import {
  CONSOLES,
  DIFICULDADE_VALORES,
  DIFICULDADES,
  GENEROS,
  type CriarJogoDto,
  type Dificuldade,
  type IgdbJogo,
  type JogoZerado
} from "@memory-card/types";

import { difficultyConfig } from "@/components/DifficultyBadge";
import { notaLabels } from "@/components/NotaBadge";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { ApiRequestError, atualizarJogo, buscarJogoIgdbPorId, buscarJogosIgdb, criarJogo } from "@/lib/api";
import { dateInputToDisplay, dateInputToIso, dateInputToLocalDate, localDateToDateInput, toDateInput } from "@/lib/formatters";
import { cn } from "@/lib/utils";

const formSchema = z.object({
  igdbId: z.number().int().optional().nullable(),
  nome: z.string().min(1, "Informe o nome").max(300),
  console: z.string().min(1, "Selecione a plataforma").max(100),
  genero: z.string().min(1, "Selecione o gênero").max(100),
  tipo: z.string().min(1, "Selecione o tipo").max(100),
  iniciadoEm: z.string().optional().nullable(),
  finalizadoEm: z.string().min(1, "Informe a data de conclusão"),
  horas: z.coerce.number().int().min(0),
  minutos: z.coerce.number().int().min(0).max(59),
  segundos: z.coerce.number().int().min(0).max(59),
  nota: z.number().int().min(1).max(11),
  dificuldade: z.enum(DIFICULDADE_VALORES),
  review: z.string().max(1000, "Máximo de 1000 caracteres").optional().nullable(),
  destaque: z.boolean(),
  igdbCapaUrl: z.string().url().optional().nullable(),
  igdbDescricao: z.string().optional().nullable()
});

type GameFormFields = z.infer<typeof formSchema>;
type ConsoleOption = (typeof CONSOLES)[number];

type GameFormProps = {
  jogoInicial?: JogoZerado;
  onSuccess: (jogo: JogoZerado) => void;
  formId?: string;
  hideFooter?: boolean;
  compact?: boolean;
  jogoSelecionadoInicial?: IgdbJogo | null;
  ocultarNome?: boolean;
  desativarAutocomplete?: boolean;
};

const notaVisual: Record<number, { color: string; label: string; value: string }> = {
  1: { color: "var(--red)", label: notaLabels[1], value: "💀" },
  2: { color: "var(--red)", label: notaLabels[2], value: "2" },
  3: { color: "var(--red)", label: notaLabels[3], value: "3" },
  4: { color: "var(--red)", label: notaLabels[4], value: "4" },
  5: { color: "var(--muted)", label: notaLabels[5], value: "5" },
  6: { color: "var(--accent)", label: notaLabels[6], value: "6" },
  7: { color: "var(--accent)", label: notaLabels[7], value: "7" },
  8: { color: "var(--green)", label: notaLabels[8], value: "8" },
  9: { color: "var(--green)", label: notaLabels[9], value: "9" },
  10: { color: "var(--green)", label: notaLabels[10], value: "10" },
  11: { color: "var(--gold)", label: notaLabels[11], value: "⭐" }
};

function secondsToParts(total: number | null | undefined) {
  const value = total ?? 0;

  return {
    horas: Math.floor(value / 3600),
    minutos: Math.floor((value % 3600) / 60),
    segundos: value % 60
  };
}

function defaultValues(jogo?: JogoZerado): GameFormFields {
  const tempo = secondsToParts(jogo?.tempoJogado);

  return {
    igdbId: jogo?.igdbId ?? null,
    nome: jogo?.nome ?? "",
    console: jogo?.console ?? "",
    genero: jogo?.genero ?? "",
    tipo: jogo?.tipo ?? "",
    iniciadoEm: toDateInput(jogo?.iniciadoEm),
    finalizadoEm: toDateInput(jogo?.finalizadoEm) || toDateInput(new Date()),
    ...tempo,
    nota: jogo?.nota ?? 8,
    dificuldade: jogo?.dificuldade ?? "B",
    review: jogo?.review ?? "",
    destaque: jogo?.destaque ?? false,
    igdbCapaUrl: jogo?.igdbCapaUrl ?? null,
    igdbDescricao: jogo?.igdbDescricao ?? null
  };
}

function filtrarPlataformas(plataformasRawg: string[]) {
  return CONSOLES.filter((console) =>
    plataformasRawg.some((plataforma) =>
      plataforma.toLowerCase().includes(console.toLowerCase()) ||
      console.toLowerCase().includes(plataforma.toLowerCase())
    )
  );
}

function encontrarGenero(generosRawg: string[]) {
  const aliases: Record<string, string[]> = {
    "Ação": ["action"],
    RPG: ["rpg", "role-playing"],
    Plataforma: ["platformer", "platform"],
    Shooter: ["shooter"],
    Estratégia: ["strategy"],
    Corrida: ["racing"],
    Simulação: ["simulation"],
    Esporte: ["sports"],
    Puzzle: ["puzzle"],
    Luta: ["fighting"]
  };

  return Object.keys(GENEROS).find((genero) =>
    generosRawg.some((generoRawg) => {
      const rawg = generoRawg.toLowerCase();
      const local = genero.toLowerCase();

      return rawg.includes(local) || local.includes(rawg) || aliases[genero]?.some((alias) => rawg.includes(alias));
    })
  );
}

function carregarImagem(url: string | null) {
  if (!url) {
    return Promise.resolve(false);
  }

  return new Promise<boolean>((resolve) => {
    const imagem = new Image();
    imagem.onload = () => resolve(true);
    imagem.onerror = () => resolve(false);
    imagem.src = url;
  });
}

export function GameForm({
  jogoInicial,
  onSuccess,
  formId,
  hideFooter,
  compact,
  jogoSelecionadoInicial,
  ocultarNome,
  desativarAutocomplete
}: GameFormProps) {
  const [plataformasDisponiveis, setPlataformasDisponiveis] = useState<ConsoleOption[]>([]);
  const [resultados, setResultados] = useState<IgdbJogo[]>([]);
  const [jogoSelecionado, setJogoSelecionado] = useState<IgdbJogo | null>(null);
  const [autocompleteOpen, setAutocompleteOpen] = useState(false);
  const [buscando, setBuscando] = useState(false);
  const autocompleteRef = useRef<HTMLDivElement>(null);
  const jogoSelecionadoInicialRef = useRef<number | null>(null);
  const form = useForm<GameFormFields>({
    resolver: zodResolver(formSchema),
    defaultValues: defaultValues(jogoInicial)
  });
  const genero = form.watch("genero");
  const finalizadoEm = form.watch("finalizadoEm");
  const nome = form.watch("nome");
  const igdbId = form.watch("igdbId");
  const review = form.watch("review") ?? "";
  const destaque = form.watch("destaque");
  const capaUrl = form.watch("igdbCapaUrl");
  const tipos = useMemo(() => (genero ? GENEROS[genero] ?? [] : []), [genero]);
  const opcoesConsole = plataformasDisponiveis.length > 0 ? plataformasDisponiveis : [...CONSOLES];
  const ano = finalizadoEm ? new Date(`${finalizadoEm}T00:00:00.000Z`).getUTCFullYear() : new Date().getUTCFullYear();

  useEffect(() => {
    if (desativarAutocomplete || igdbId || nome.trim().length < 3) {
      setResultados([]);
      setAutocompleteOpen(false);
      return;
    }

    let active = true;
    const timeout = window.setTimeout(async () => {
      setBuscando(true);

      try {
        const jogos = await buscarJogosIgdb(nome);

        if (active) {
          setResultados(jogos);
          setAutocompleteOpen(true);
        }
      } finally {
        if (active) {
          setBuscando(false);
        }
      }
    }, 400);

    return () => {
      active = false;
      window.clearTimeout(timeout);
    };
  }, [desativarAutocomplete, igdbId, nome]);

  useEffect(() => {
    if (!jogoSelecionadoInicial || jogoSelecionadoInicialRef.current === jogoSelecionadoInicial.igdbId) {
      return;
    }

    jogoSelecionadoInicialRef.current = jogoSelecionadoInicial.igdbId;
    void selecionarJogo(jogoSelecionadoInicial);
  }, [jogoSelecionadoInicial]);

  useEffect(() => {
    function handleClick(event: MouseEvent) {
      if (!autocompleteRef.current?.contains(event.target as Node)) {
        setAutocompleteOpen(false);
      }
    }

    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  async function selecionarJogo(jogo: IgdbJogo) {
    const plataformasDoJogo = jogo.plataformas ?? [];
    const generosDoJogo = jogo.generos ?? [];
    const plataformasFiltradas = filtrarPlataformas(plataformasDoJogo);
    const plataformas = plataformasFiltradas.length > 0 ? plataformasFiltradas : [...CONSOLES];
    const generoEncontradoBusca = encontrarGenero(generosDoJogo);

    console.info("[GameForm] capaUrl da busca", jogo.capaUrl);
    setJogoSelecionado(jogo);
    form.setValue("igdbId", jogo.igdbId);
    form.setValue("nome", jogo.nome, { shouldValidate: true });
    form.setValue("igdbCapaUrl", jogo.capaUrl, { shouldValidate: true });
    setPlataformasDisponiveis(plataformas);

    if (plataformas.length === 1) {
      form.setValue("console", plataformas[0], { shouldValidate: true });
    } else if (!plataformas.includes(form.getValues("console") as ConsoleOption)) {
      form.setValue("console", "", { shouldValidate: true });
    }

    if (generoEncontradoBusca) {
      form.setValue("genero", generoEncontradoBusca, { shouldValidate: true });
      form.setValue("tipo", GENEROS[generoEncontradoBusca][0], { shouldValidate: true });
    }

    setAutocompleteOpen(false);

    try {
      const detalhe = await buscarJogoIgdbPorId(jogo.igdbId);
      const generoEncontrado = encontrarGenero(detalhe.generos);

      form.setValue("igdbId", detalhe.igdbId);
      form.setValue("nome", detalhe.nome, { shouldValidate: true });
      form.setValue("igdbCapaUrl", (await carregarImagem(detalhe.capaUrl)) ? detalhe.capaUrl : jogo.capaUrl, { shouldValidate: true });
      form.setValue("igdbDescricao", detalhe.descricao);

      if (generoEncontrado) {
        form.setValue("genero", generoEncontrado, { shouldValidate: true });
        form.setValue("tipo", GENEROS[generoEncontrado][0], { shouldValidate: true });
      }
    } catch {
      form.setValue("igdbDescricao", null);
    }
  }

  async function onSubmit(values: GameFormFields) {
    const payload: CriarJogoDto = {
      igdbId: values.igdbId,
      nome: values.nome,
      console: values.console,
      genero: values.genero,
      tipo: values.tipo,
      iniciadoEm: dateInputToIso(values.iniciadoEm),
      finalizadoEm: dateInputToIso(values.finalizadoEm) ?? new Date().toISOString(),
      tempoJogado: values.horas * 3600 + values.minutos * 60 + values.segundos,
      nota: values.nota,
      dificuldade: values.dificuldade,
      review: values.review || null,
      destaque: values.destaque,
      igdbCapaUrl: values.igdbCapaUrl,
      igdbDescricao: values.igdbDescricao
    };

    try {
      const jogo = jogoInicial ? await atualizarJogo(jogoInicial.id, payload) : await criarJogo(payload);
      onSuccess(jogo);
    } catch (error) {
      if (error instanceof ApiRequestError && error.status === 409) {
        form.setError("destaque", { message: error.message });
        return;
      }

      throw error;
    }
  }

  return (
    <form id={formId} noValidate onSubmit={form.handleSubmit(onSubmit)} className={cn("overflow-hidden rounded-lg border border-border bg-background", compact && "rounded-none border-0 bg-surface")}>
      <div className={cn("grid max-h-[calc(100vh-180px)] grid-cols-1 md:grid-cols-[280px_1fr]", compact && "max-h-none md:grid-cols-[220px_1fr]")}>
        <aside className={cn("space-y-4 border-b border-border bg-background p-4 md:border-b-0 md:border-r", compact && "space-y-3 bg-surface p-4")}>
          <div className="aspect-[2/3] overflow-hidden rounded-lg border border-border bg-surface">
            {capaUrl ? (
              <img
                src={capaUrl}
                alt=""
                onError={() => form.setValue("igdbCapaUrl", jogoSelecionado?.capaUrl ?? null, { shouldValidate: true })}
                className="h-full w-full object-cover object-[center_top]"
              />
            ) : (
              <div className="flex h-full flex-col items-center justify-center gap-3 p-4 text-center text-xs text-muted">
                <ImageOff className="h-7 w-7" />
                <span>Capa carregada automaticamente</span>
              </div>
            )}
          </div>
          {!ocultarNome && (
            <Field label="Nome do jogo" error={form.formState.errors.nome?.message}>
              <div ref={autocompleteRef} className="relative">
                <Controller
                  control={form.control}
                  name="nome"
                  render={({ field }) => (
                    <Input
                      className="bg-surface-raised"
                      value={field.value}
                      onChange={(event) => {
                        field.onChange(event.target.value);
                        form.setValue("igdbId", null);
                        setPlataformasDisponiveis([]);
                      }}
                    />
                  )}
                />
                {autocompleteOpen && (
                  <div className="absolute z-40 mt-2 max-h-72 w-full overflow-auto rounded-md border border-border bg-surface shadow-xl">
                    {buscando ? (
                      <p className="p-3 text-xs text-muted">Buscando...</p>
                    ) : resultados.length === 0 ? (
                      <p className="p-3 text-xs text-muted">Nenhum resultado</p>
                    ) : (
                      resultados.map((jogo) => (
                        <button
                          type="button"
                          key={jogo.igdbId}
                          onClick={() => void selecionarJogo(jogo)}
                          className="flex w-full items-center gap-2 p-2 text-left hover:bg-surface-raised"
                        >
                          {jogo.capaUrl ? <img src={jogo.capaUrl} alt="" className="h-8 w-8 rounded-sm object-cover" /> : <span className="flex h-8 w-8 items-center justify-center rounded-sm bg-surface-raised"><ImageOff className="h-4 w-4 text-muted" /></span>}
                          <span className="min-w-0">
                            <span className="block truncate text-xs text-foreground">{jogo.nome}</span>
                            <span className="font-mono text-[10px] text-muted">{jogo.ano ?? "—"}</span>
                          </span>
                        </button>
                      ))
                    )}
                  </div>
                )}
              </div>
            </Field>
          )}
          <SelectField
            label="Plataforma"
            value={form.watch("console")}
            error={form.formState.errors.console?.message}
            options={opcoesConsole}
            onChange={(value) => form.setValue("console", value, { shouldValidate: true })}
          />
        </aside>
        <section className={cn("min-h-0 space-y-8 overflow-y-auto p-4 md:p-6", compact && "max-h-[calc(92vh-129px)] space-y-3 overflow-y-auto p-4 min-[1200px]:max-h-none min-[1200px]:overflow-visible")}>
          <FormSection title="Zeramento">
            <div className="grid gap-4 md:grid-cols-2">
              <SelectField label="Gênero" value={genero} error={form.formState.errors.genero?.message} options={Object.keys(GENEROS)} onChange={(value) => {
                form.setValue("genero", value, { shouldValidate: true });
                form.setValue("tipo", GENEROS[value]?.[0] ?? "", { shouldValidate: true });
              }} />
              <SelectField label="Tipo" value={form.watch("tipo")} error={form.formState.errors.tipo?.message} options={tipos} onChange={(value) => form.setValue("tipo", value, { shouldValidate: true })} />
            </div>
            <div className="grid gap-4 md:grid-cols-2">
              <Controller
                control={form.control}
                name="iniciadoEm"
                render={({ field }) => (
                  <DatePickerField
                    label="Iniciado em"
                    value={field.value ?? ""}
                    error={form.formState.errors.iniciadoEm?.message}
                    optional
                    onChange={field.onChange}
                  />
                )}
              />
              <Controller
                control={form.control}
                name="finalizadoEm"
                render={({ field }) => (
                  <DatePickerField
                    label="Concluído em"
                    value={field.value ?? ""}
                    error={form.formState.errors.finalizadoEm?.message}
                    onChange={field.onChange}
                  />
                )}
              />
            </div>
            <div className="grid grid-cols-3 gap-2">
              <Controller
                control={form.control}
                name="horas"
                render={({ field }) => (
                  <TimeField
                    label="horas"
                    value={field.value}
                    max={9999}
                    error={form.formState.errors.horas?.message}
                    onChange={field.onChange}
                  />
                )}
              />
              <Controller
                control={form.control}
                name="minutos"
                render={({ field }) => (
                  <TimeField
                    label="minutos"
                    value={field.value}
                    max={59}
                    clampMax
                    error={form.formState.errors.minutos?.message}
                    onChange={field.onChange}
                  />
                )}
              />
              <Controller
                control={form.control}
                name="segundos"
                render={({ field }) => (
                  <TimeField
                    label="segundos"
                    value={field.value}
                    max={59}
                    clampMax
                    error={form.formState.errors.segundos?.message}
                    onChange={field.onChange}
                  />
                )}
              />
            </div>
          </FormSection>
          <FormSection title="Avaliação">
            <Controller
              control={form.control}
              name="nota"
              render={({ field }) => (
                <div className="grid grid-cols-6 gap-2 min-[1200px]:grid-cols-11">
                  {Array.from({ length: 11 }, (_, index) => index + 1).map((nota) => (
                    <button
                      type="button"
                      key={nota}
                      onClick={() => field.onChange(nota)}
                      className={cn("flex items-center justify-center rounded-md border text-center text-xs", compact ? "h-11 w-11 p-0" : "p-2")}
                      style={{
                        background: field.value === nota ? `color-mix(in srgb, ${notaVisual[nota].color} 12%, transparent)` : "var(--surface)",
                        borderColor: field.value === nota ? `color-mix(in srgb, ${notaVisual[nota].color} 40%, transparent)` : "var(--border)",
                        color: field.value === nota ? notaVisual[nota].color : `color-mix(in srgb, ${notaVisual[nota].color} 72%, var(--muted))`
                      }}
                    >
                      <span className="block font-mono text-lg">{notaVisual[nota].value}</span>
                    </button>
                  ))}
                </div>
              )}
            />
            <Controller
              control={form.control}
              name="dificuldade"
              render={({ field }) => (
                <div className="grid grid-cols-5 gap-2">
                  {DIFICULDADES.map((dificuldade) => (
                    <button type="button" key={dificuldade.valor} onClick={() => field.onChange(dificuldade.valor)} className={cn("rounded-md border p-3 text-center", compact && "p-2", field.value === dificuldade.valor ? difficultyConfig[dificuldade.valor as Dificuldade].className : "border-border bg-surface")}>
                      <span className="block font-mono text-[13px] font-semibold">{dificuldade.valor}</span>
                      <span className="block truncate text-[9px]">{dificuldade.label}</span>
                    </button>
                  ))}
                </div>
              )}
            />
          </FormSection>
          <FormSection title="Review">
            <Field label="Sua análise" error={form.formState.errors.review?.message}>
              <div className="relative">
                <Textarea
                  maxLength={1000}
                  placeholder="O que você achou do jogo? Pontos altos, baixos, recomendaria?"
                  className={cn("min-h-[100px] resize-y bg-surface-raised pb-7", compact && "min-h-[60px] pb-5")}
                  {...form.register("review")}
                />
                <span className="absolute bottom-2 right-3 font-mono text-[10px] text-muted">{1000 - review.length}</span>
              </div>
            </Field>
            <Field error={form.formState.errors.destaque?.message}>
              <button
                type="button"
                onClick={() => form.setValue("destaque", !destaque, { shouldDirty: true, shouldValidate: true })}
                className="flex w-full cursor-pointer items-center gap-3 rounded-md border border-border bg-surface px-3 py-2.5 text-left text-sm"
              >
                <span className={cn("flex h-4 w-4 items-center justify-center rounded-sm border", destaque ? "border-accent bg-accent" : "border-border bg-surface")}>
                  {destaque && <Check className="h-2.5 w-2.5 text-white" />}
                </span>
                <span>Marcar como destaque de {ano}</span>
              </button>
            </Field>
          </FormSection>
        </section>
      </div>
      {!hideFooter && (
        <footer className="flex items-center justify-end gap-3 border-t border-border bg-background p-4">
          <Button type="button" variant="secondary" asChild>
            <Link to="/biblioteca">Cancelar</Link>
          </Button>
          <Button type="submit" disabled={form.formState.isSubmitting}>
            {form.formState.isSubmitting ? "Salvando..." : "Salvar zeramento"}
          </Button>
        </footer>
      )}
    </form>
  );
}

function FormSection({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="space-y-4">
      <div className="flex items-center gap-3">
        <h2 className="text-[10px] font-semibold uppercase tracking-[0.6px] text-muted">{title}</h2>
        <div className="h-px flex-1 bg-border" />
      </div>
      {children}
    </section>
  );
}

function Field({ label, error, children }: { label?: string; error?: string; children: ReactNode }) {
  return (
    <div className="space-y-1.5">
      {label && <Label className="text-[10px] uppercase tracking-[0.6px] text-muted">{label}</Label>}
      {children}
      {error && <p className="text-xs text-red">{error}</p>}
    </div>
  );
}

function TextField({ label, error, ...props }: InputHTMLAttributes<HTMLInputElement> & { label: string; error?: string }) {
  return (
    <Field label={label} error={error}>
      <Input className={cn("bg-surface-raised", error && "border-red")} {...props} />
    </Field>
  );
}

function normalizeTimeValue(rawValue: string, max: number, clampMax?: boolean) {
  if (rawValue === "") {
    return 0;
  }

  const value = Math.max(0, Math.trunc(Number(rawValue)));

  return clampMax && value > max ? max : value;
}

function TimeField({ label, value, max, clampMax, error, onChange }: { label: string; value: number; max: number; clampMax?: boolean; error?: string; onChange: (value: number) => void }) {
  return (
    <Field error={error}>
      <Input
        type="number"
        min={0}
        max={max}
        placeholder="0"
        value={value > 0 ? value : ""}
        onChange={(event) => onChange(normalizeTimeValue(event.target.value, max, clampMax))}
        className={cn(
          "h-12 appearance-none border-border bg-surface text-center font-mono text-lg text-foreground placeholder:text-muted [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none",
          error && "border-red"
        )}
      />
      <p className="text-center text-[10px] uppercase text-muted">{label}</p>
    </Field>
  );
}

function DatePickerField({ label, value, error, optional, onChange }: { label: string; value: string; error?: string; optional?: boolean; onChange: (value: string) => void }) {
  const [open, setOpen] = useState(false);

  return (
    <Field label={label} error={error}>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <button
            type="button"
            className={cn(
              "flex h-9 w-full items-center justify-between rounded-md border border-input bg-surface-raised px-3 py-2 text-left text-sm text-foreground shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring",
              !value && "text-muted",
              error && "border-red"
            )}
          >
            <span>{dateInputToDisplay(value) || label}</span>
            <CalendarDays className="h-4 w-4 text-muted" />
          </button>
        </PopoverTrigger>
        <PopoverContent align="start" className="w-auto p-0">
          <Calendar
            mode="single"
            selected={dateInputToLocalDate(value)}
            defaultMonth={dateInputToLocalDate(value)}
            onSelect={(date) => {
              if (date) {
                onChange(localDateToDateInput(date));
                setOpen(false);
              }
            }}
          />
          {optional && (
            <div className="border-t border-border p-2">
              <Button
                type="button"
                variant="secondary"
                size="sm"
                className="w-full"
                onClick={() => {
                  onChange("");
                  setOpen(false);
                }}
              >
                Limpar
              </Button>
            </div>
          )}
        </PopoverContent>
      </Popover>
    </Field>
  );
}

function SelectField({ label, value, options, error, onChange }: { label: string; value: string; options: string[]; error?: string; onChange: (value: string) => void }) {
  return (
    <Field label={label} error={error}>
      <Select value={value} onValueChange={onChange}>
        <SelectTrigger className={cn("bg-surface-raised", error && "border-red")}>
          <SelectValue placeholder={label} />
        </SelectTrigger>
        <SelectContent>
          {options.map((option) => (
            <SelectItem key={option} value={option}>
              {option}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </Field>
  );
}
