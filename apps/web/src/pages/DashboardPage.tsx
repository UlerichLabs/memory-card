import { Bar, BarChart, Cell, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import type { ReactNode } from "react";
import { Gamepad2, Star, Target, Trophy } from "lucide-react";
import { Link } from "react-router-dom";

import { DifficultyBadge } from "@/components/DifficultyBadge";
import { NotaBadge } from "@/components/NotaBadge";
import { Button } from "@/components/ui/button";
import { formatDate } from "@/lib/formatters";
import { useDashboard } from "@/hooks/useDashboard";

const currentYear = new Date().getFullYear();
const statTone = {
  accent: { line: "bg-accent", text: "text-accent" },
  gold: { line: "bg-gold", text: "text-gold" },
  green: { line: "bg-green", text: "text-green" },
  purple: { line: "bg-purple", text: "text-purple" }
};

export default function DashboardPage() {
  const { totais, porAno, porPlataforma, ultimos, isLoading } = useDashboard();
  const total = totais.data;
  const anos = porAno.data ?? [];
  const plataformas = porPlataforma.data ?? [];
  const recentes = ultimos.data?.data ?? [];
  const goty = anos.find((item) => item.ano === currentYear)?.destaque ?? null;
  const liderPlataforma = Math.max(...plataformas.map((item) => item.total), 1);

  if (isLoading) {
    return (
      <main className="min-h-screen bg-background p-4 text-foreground md:p-6">
        <div className="mx-auto max-w-7xl space-y-6">
          <div className="h-24 animate-pulse rounded-lg bg-surface" />
          <div className="grid gap-4 md:grid-cols-4">
            {Array.from({ length: 4 }, (_, index) => <div key={index} className="h-32 animate-pulse rounded-lg bg-surface" />)}
          </div>
          <div className="grid gap-4 lg:grid-cols-3">
            <div className="h-80 animate-pulse rounded-lg bg-surface lg:col-span-2" />
            <div className="h-80 animate-pulse rounded-lg bg-surface" />
          </div>
        </div>
      </main>
    );
  }

  if ((total?.totalJogos ?? 0) === 0) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-background p-6 text-foreground">
        <section className="max-w-sm space-y-4 text-center">
          <Gamepad2 className="mx-auto h-12 w-12 text-muted" />
          <h1 className="text-2xl font-bold">Nenhum zeramento ainda</h1>
          <Button asChild>
            <Link to="/biblioteca/novo">Registrar primeiro jogo</Link>
          </Button>
        </section>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-background p-4 text-foreground md:p-6">
      <div className="mx-auto max-w-7xl space-y-6">
        <header className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold">Sua jornada</h1>
            <p className="mt-2 text-sm text-muted">Resumo dos zeramentos, plataformas e destaques.</p>
          </div>
          <span className="rounded-md border border-border bg-surface px-3 py-1 font-mono text-sm text-accent">{currentYear}</span>
        </header>
        <section className="grid gap-4 md:grid-cols-4">
          <StatCard label="Zeramentos" value={String(total?.totalJogos ?? 0)} tone="accent" />
          <StatCard label="Horas jogadas" value={String(Math.round((total?.totalSegundos ?? 0) / 3600))} tone="gold" />
          <StatCard label="Nota média" value={(total?.notaMedia ?? 0).toFixed(1)} tone="green" />
          <StatCard label="Plataformas" value={String(total?.totalConsoles ?? 0)} tone="purple" />
        </section>
        <section className="grid gap-4 lg:grid-cols-3">
          <Panel className="lg:col-span-2" title="Zeramentos por ano">
            <div className="h-72">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={anos} barCategoryGap="40%">
                  <XAxis dataKey="ano" axisLine={false} tickLine={false} tick={{ fill: "var(--muted)", fontSize: 12 }} />
                  <YAxis hide />
                  <Tooltip cursor={false} contentStyle={{ background: "var(--surface-raised)", border: "1px solid var(--border)", color: "var(--foreground)" }} />
                  <Bar dataKey="totalJogos" radius={[6, 6, 0, 0]} maxBarSize={48}>
                    {anos.map((item) => <Cell key={item.ano} fill="var(--accent)" fillOpacity={item.ano === currentYear ? 1 : 0.15} />)}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>
          </Panel>
          <div className="space-y-4">
            <Panel title="Jogo do Ano" className="border-gold-dim bg-gold-dim">
              {goty ? (
                <div className="space-y-3">
                  <Star className="h-6 w-6 text-gold" />
                  <div>
                    <p className="font-medium text-foreground">{goty.nome}</p>
                    <p className="text-sm text-muted">{goty.console}</p>
                  </div>
                  <NotaBadge nota={goty.nota} small />
                </div>
              ) : (
                <p className="text-sm text-muted">Nenhum destaque marcado neste ano.</p>
              )}
            </Panel>
            <Panel title="Missão ativa">
              <div className="space-y-3">
                <div className="flex items-center gap-2 text-sm text-muted">
                  <Target className="h-4 w-4 text-accent" />
                  Criar a próxima missão de zeramentos
                </div>
                <div className="h-2 rounded-full bg-surface-raised">
                  <div className="h-full w-0 rounded-full bg-accent" />
                </div>
                <div className="flex items-center justify-between font-mono text-xs text-muted">
                  <span>0/0</span>
                  <span>0%</span>
                </div>
                <Button asChild variant="secondary" size="sm">
                  <Link to="/desafios">Criar missão</Link>
                </Button>
              </div>
            </Panel>
          </div>
        </section>
        <section className="grid gap-4 lg:grid-cols-2">
          <Panel title="Últimos zeramentos">
            <div className="space-y-3">
              {recentes.map((jogo) => (
                <div key={jogo.id} className="flex items-center gap-3 rounded-md bg-surface-raised p-3">
                  {jogo.igdbCapaUrl ? <img src={jogo.igdbCapaUrl} alt="" className="h-14 w-10 rounded object-cover" /> : <div className="flex h-14 w-10 items-center justify-center rounded bg-surface"><Gamepad2 className="h-5 w-5 text-muted" /></div>}
                  <div className="min-w-0 flex-1">
                    <p className="truncate font-medium">{jogo.nome}</p>
                    <p className="text-xs text-muted">{jogo.console} · {formatDate(jogo.finalizadoEm)}</p>
                  </div>
                  <DifficultyBadge dificuldade={jogo.dificuldade} small />
                  <NotaBadge nota={jogo.nota} small />
                </div>
              ))}
            </div>
          </Panel>
          <Panel title="Top plataformas">
            <div className="space-y-4">
              {plataformas.map((item) => (
                <div key={item.console} className="space-y-1.5">
                  <div className="flex items-center justify-between text-sm">
                    <span>{item.console}</span>
                    <span className="font-mono text-muted">{item.total}</span>
                  </div>
                  <div className="h-2 rounded-full bg-surface-raised">
                    <div className="h-full rounded-full bg-purple" style={{ width: `${Math.round((item.total / liderPlataforma) * 100)}%` }} />
                  </div>
                </div>
              ))}
            </div>
          </Panel>
        </section>
      </div>
    </main>
  );
}

function StatCard({ label, value, tone }: { label: string; value: string; tone: "accent" | "gold" | "green" | "purple" }) {
  const config = statTone[tone];

  return (
    <div className="overflow-hidden rounded-lg border border-border bg-surface">
      <div className={`h-1 ${config.line}`} />
      <div className="p-4">
        <p className="text-xs uppercase tracking-wide text-muted">{label}</p>
        <p className={`mt-2 font-mono text-3xl ${config.text}`}>{value}</p>
      </div>
    </div>
  );
}

function Panel({ title, className = "", children }: { title: string; className?: string; children: ReactNode }) {
  return (
    <section className={`rounded-lg border border-border bg-surface p-4 ${className}`}>
      <div className="mb-4 flex items-center gap-2">
        <Trophy className="h-4 w-4 text-accent" />
        <h2 className="font-medium">{title}</h2>
      </div>
      {children}
    </section>
  );
}
