import { create } from "zustand";

interface GameFormStore {
  isOpen: boolean;
  modo: "criar" | "editar";
  jogoId: number | null;
  abrirCriar: () => void;
  abrirEditar: (jogoId: number) => void;
  fechar: () => void;
}

export const useGameFormStore = create<GameFormStore>((set) => ({
  isOpen: false,
  modo: "criar",
  jogoId: null,
  abrirCriar: () => set({ isOpen: true, modo: "criar", jogoId: null }),
  abrirEditar: (jogoId) => set({ isOpen: true, modo: "editar", jogoId }),
  fechar: () => set({ isOpen: false, modo: "criar", jogoId: null })
}));
