import { BookOpen, Gauge, ListChecks, LogOut, Plus, Target, User } from "lucide-react";
import { NavLink } from "react-router-dom";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from "@/components/ui/dropdown-menu";
import { useAuth } from "@/hooks/useAuth";
import { cn } from "@/lib/utils";
import { useGameFormStore } from "@/store/gameFormStore";

const navItems = [
  { label: "Dashboard", to: "/dashboard", icon: Gauge },
  { label: "Biblioteca", to: "/biblioteca", icon: BookOpen },
  { label: "Desafios", to: "/desafios", icon: Target },
  { label: "Missões", to: "/desafios", icon: ListChecks }
];

export function Navbar() {
  const { user, logout } = useAuth();
  const abrirCriar = useGameFormStore((state) => state.abrirCriar);
  const initials = user?.nome?.slice(0, 1).toUpperCase() ?? "M";
  const userName = user?.nome ?? "Memory Card";
  const userEmail = user?.email ?? "";

  return (
    <header className="sticky top-0 z-50 border-b border-border bg-background">
      <div className="mx-auto flex h-16 max-w-7xl items-center gap-4 px-4 md:px-6">
        <div className="flex items-center gap-2 font-sans font-bold text-foreground">
          <span className="h-2.5 w-2.5 rounded-full bg-accent" />
          <span>Memory Card</span>
        </div>
        <nav className="flex flex-1 items-center gap-1 overflow-x-auto">
          {navItems.map((item) => (
            <NavLink
              key={`${item.label}-${item.to}`}
              to={item.to}
              className={({ isActive }) =>
                cn(
                  "inline-flex items-center gap-2 rounded-md px-3 py-2 text-sm text-muted transition hover:bg-surface-raised hover:text-foreground",
                  isActive && "bg-accent-dim text-accent"
                )
              }
            >
              <item.icon className="h-4 w-4" />
              {item.label}
            </NavLink>
          ))}
        </nav>
        <button
          type="button"
          onClick={abrirCriar}
          className="inline-flex items-center gap-2 rounded-md bg-accent px-3 py-2 text-sm font-medium text-background"
        >
          <Plus className="h-4 w-4" />
          Registrar
        </button>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button
              type="button"
              className="flex h-9 w-9 items-center justify-center rounded-full border border-border bg-surface-raised font-mono text-sm text-foreground outline-none transition hover:bg-surface focus-visible:ring-1 focus-visible:ring-ring"
              aria-label="Abrir menu do usuário"
            >
              {initials}
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <div className="px-2 py-1.5">
              <p className="truncate text-sm font-medium text-foreground">{userName}</p>
              <p className="truncate text-xs text-muted">{userEmail}</p>
            </div>
            <DropdownMenuSeparator />
            <DropdownMenuItem disabled onSelect={(event) => event.preventDefault()} className="text-muted">
              <User className="h-4 w-4" />
              <span>Perfil</span>
            </DropdownMenuItem>
            <DropdownMenuItem
              onSelect={logout}
              className="focus:bg-[color-mix(in_srgb,var(--red)_8%,transparent)] focus:text-red"
            >
              <LogOut className="h-4 w-4" />
              <span>Sair</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
