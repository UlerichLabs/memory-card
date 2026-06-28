export const CONSOLES = [
  "Master System",
  "Game Gear",
  "NES",
  "SNES",
  "Mega Drive",
  "Sega CD",
  "Sega Saturn",
  "Dreamcast",
  "PlayStation 1",
  "Nintendo 64",
  "PlayStation 2",
  "GameCube",
  "Xbox",
  "Game Boy",
  "Game Boy Color",
  "Game Boy Advance",
  "Nintendo DS",
  "PSP",
  "Wii",
  "PlayStation 3",
  "Xbox 360",
  "Nintendo 3DS",
  "PS Vita",
  "Wii U",
  "PlayStation 4",
  "Xbox One",
  "Switch",
  "PlayStation 5",
  "Xbox Series X",
  "Switch 2",
  "PC",
  "Arcade",
  "Mobile",
  "Outro"
] as const;

export const GENEROS: Record<string, string[]> = {
  RPG: [
    "Action RPG",
    "RPG de Turno",
    "RPG Estratégico",
    "RPG Plataforma",
    "Estilo Ocidental",
    "MMORPG",
    "Roguelike"
  ],
  Plataforma: [
    "2D Clássico",
    "Explorativo/Metroidvania",
    "Ambiente 3D",
    "Run & Gun",
    "Collectathon",
    "Estilo Kaizo"
  ],
  Shooter: [
    "Direção Vertical",
    "Direção Horizontal",
    "Bullet Hell",
    "Rail Shooter",
    "FPS"
  ],
  Ação: [
    "Hack & Slash",
    "Furtivo",
    "Survival Horror",
    "Mundo Aberto",
    "Exploração 3D"
  ],
  Luta: ["Luta 2D", "Luta 3D", "Luta em Área"],
  "Briga de Rua": ["Isométrico", "2D Clássico"],
  Estratégia: ["Turno/Unidades", "Turno/Isométrico", "Gerenciamento", "Tower Defense"],
  Corrida: ["Estilo Kart", "Contra o Tempo", "Visão Superior", "Simulação"],
  Simulação: ["Fazendinha", "Vida Real", "Pescaria", "Veículos"],
  Esporte: ["Futebol", "Tênis", "Golfe", "Outro"],
  Mesa: ["Tabuleiro", "Card Game"],
  Puzzle: ["Lógico", "Combinativo"],
  Outro: ["Adventure", "Ritmo", "Visual Novel", "Game Show"]
};

export const NOTAS: Record<number, { label: string; especial?: boolean }> = {
  11: { label: "⭐ Jogo da Vida", especial: true },
  10: { label: "Obra-Prima" },
  9: { label: "Excepcional" },
  8: { label: "Muito Bom" },
  7: { label: "Bom" },
  6: { label: "Razoável" },
  5: { label: "Mediano" },
  4: { label: "Fraco" },
  3: { label: "Ruim" },
  2: { label: "Horrível" },
  1: { label: "💀 Tragédia", especial: true }
};

export const DIFICULDADES = [
  { valor: "C", label: "Fácil", cor: "#6B7280" },
  { valor: "B", label: "Normal", cor: "#3B82F6" },
  { valor: "A", label: "Difícil", cor: "#EAB308" },
  { valor: "AA", label: "Muito Difícil", cor: "#F97316" },
  { valor: "AAA", label: "Platina", cor: "#EF4444" }
] as const;

export type DificuldadeValor = (typeof DIFICULDADES)[number]["valor"];

export const DIFICULDADE_VALORES = ["C", "B", "A", "AA", "AAA"] as const;
