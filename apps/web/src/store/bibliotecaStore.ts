import { create } from "zustand";

interface BibliotecaStore {
  busca: string;
  setBusca: (texto: string) => void;
}

export const useBibliotecaStore = create<BibliotecaStore>((set) => ({
  busca: "",
  setBusca: (texto) => set({ busca: texto })
}));
