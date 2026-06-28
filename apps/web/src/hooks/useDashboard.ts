import { useQuery } from "@tanstack/react-query";

import {
  buscarDashboardPorAno,
  buscarDashboardPorPlataforma,
  buscarDashboardRecordes,
  buscarDashboardTotais,
  listarJogos
} from "@/lib/api";

export function useDashboard() {
  const totais = useQuery({
    queryKey: ["dashboard", "totais"],
    queryFn: buscarDashboardTotais
  });
  const porAno = useQuery({
    queryKey: ["dashboard", "por-ano"],
    queryFn: buscarDashboardPorAno
  });
  const porPlataforma = useQuery({
    queryKey: ["dashboard", "por-plataforma"],
    queryFn: buscarDashboardPorPlataforma
  });
  const recordes = useQuery({
    queryKey: ["dashboard", "recordes"],
    queryFn: buscarDashboardRecordes
  });
  const ultimos = useQuery({
    queryKey: ["dashboard", "ultimos"],
    queryFn: () => listarJogos({ page: 1, limit: 4 })
  });

  return {
    totais,
    porAno,
    porPlataforma,
    recordes,
    ultimos,
    isLoading: totais.isLoading || porAno.isLoading || porPlataforma.isLoading || recordes.isLoading || ultimos.isLoading
  };
}
