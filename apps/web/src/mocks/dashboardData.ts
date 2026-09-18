export interface EstatisticasGerais {
  totalJogosZerados: number
  totalHorasJogadas: number
  totalAbandonados: number
  notaMedia: number
  mediaHorasPorJogo?: number
  jogosZeradosAnoAtual?: number
}

export interface JogoResumo {
  id: string
  titulo: string
  capaUrl: string
  anoLancamento: number
  plataforma: string
  horasJogadas: number
  nota: number
}

export interface JogoDoAno {
  ano: number
  jogo: JogoResumo
  comentario: string
}

export interface JogoDaVida {
  posicao: number
  jogo: JogoResumo
}

export interface DesafioAtivo {
  id: string
  titulo: string
  descricao: string
  progressoAtual: number
  progressoMeta: number
  unidade: string
  prazo: string
}

export interface PontoAtividade {
  data: string
  horas: number
  intensidade: number
}

export interface AtividadeAno {
  ano: number
  totalHoras: number
  diasAtivos: number
  pontos: PontoAtividade[]
}

export interface JogoZeradoRecente {
  id: string
  titulo: string
  capaUrl: string
  plataforma: string
  horasJogadas: number
  nota: number
  dataFinalizacao: string
  dificuldade?: string
}

export interface PlataformaEstatistica {
  plataforma: string
  totalJogos: number
  horasJogadas: number
  percentual: number
}

export interface GeneroEstatistica {
  genero: string
  totalJogos: number
  horasJogadas: number
  percentual: number
}

export interface DashboardData {
  estatisticas: EstatisticasGerais
  jogoDoAno: JogoDoAno
  jogosDaVida: JogoDaVida[]
  desafiosAtivos: DesafioAtivo[]
  atividadeAno: AtividadeAno
  zeradosRecentemente: JogoZeradoRecente[]
  distribuicaoPlataformas: PlataformaEstatistica[]
  principaisGeneros: GeneroEstatistica[]
}

export function calcularIntensidade(horas: number): number {
  if (horas <= 0) return 0
  if (horas <= 2) return 1
  if (horas <= 4) return 2
  if (horas <= 6) return 3
  return 4
}

function gerarPontosAtividade(ano: number): PontoAtividade[] {
  const pontos: PontoAtividade[] = []
  const totalDias = 365
  const dataBase = new Date(ano, 0, 1)

  for (let i = 0; i < totalDias; i++) {
    const dataAtual = new Date(dataBase)
    dataAtual.setDate(dataBase.getDate() + i)
    const isoData = dataAtual.toISOString().split('T')[0]
    const fator = (i * 17 + 23) % 10
    const horas = fator > 6 ? (fator % 4) * 2 + 1 : fator > 3 ? 2 : 0
    const intensidade = calcularIntensidade(horas)

    pontos.push({
      data: isoData,
      horas,
      intensidade,
    })
  }

  return pontos
}

export const mockDashboardData: DashboardData = {
  estatisticas: {
    totalJogosZerados: 128,
    totalHorasJogadas: 2450,
    totalAbandonados: 14,
    notaMedia: 8.2,
    mediaHorasPorJogo: 19.1,
    jogosZeradosAnoAtual: 24,
  },
  jogoDoAno: {
    ano: 2025,
    jogo: {
      id: 'goty-2025',
      titulo: "Baldur's Gate 3",
      capaUrl: 'https://images.unsplash.com/photo-1542751371-adc38448a05e?w=600&auto=format&fit=crop&q=80',
      anoLancamento: 2023,
      plataforma: 'PC',
      horasJogadas: 148,
      nota: 10,
    },
    comentario: 'Uma obra-prima absoluta de liberdade narrativa, profundidade de sistemas e companheiros inesquecíveis.',
  },
  jogosDaVida: [
    {
      posicao: 1,
      jogo: {
        id: 'jdv-1',
        titulo: 'Chrono Trigger',
        capaUrl: 'https://images.unsplash.com/photo-1550745165-9bc0b252726f?w=600&auto=format&fit=crop&q=80',
        anoLancamento: 1995,
        plataforma: 'SNES',
        horasJogadas: 35,
        nota: 10,
      },
    },
    {
      posicao: 2,
      jogo: {
        id: 'jdv-2',
        titulo: 'Elden Ring',
        capaUrl: 'https://images.unsplash.com/photo-1538481199705-c710c4e965fc?w=600&auto=format&fit=crop&q=80',
        anoLancamento: 2022,
        plataforma: 'PC',
        horasJogadas: 160,
        nota: 10,
      },
    },
    {
      posicao: 3,
      jogo: {
        id: 'jdv-3',
        titulo: 'Hollow Knight',
        capaUrl: 'https://images.unsplash.com/photo-1511512578047-dfb367046420?w=600&auto=format&fit=crop&q=80',
        anoLancamento: 2017,
        plataforma: 'PC',
        horasJogadas: 55,
        nota: 10,
      },
    },
    {
      posicao: 4,
      jogo: {
        id: 'jdv-4',
        titulo: 'The Witcher 3: Wild Hunt',
        capaUrl: 'https://images.unsplash.com/photo-1552824722-ddab1374e622?w=600&auto=format&fit=crop&q=80',
        anoLancamento: 2015,
        plataforma: 'PC',
        horasJogadas: 120,
        nota: 10,
      },
    },
    {
      posicao: 5,
      jogo: {
        id: 'jdv-5',
        titulo: 'Bloodborne',
        capaUrl: 'https://images.unsplash.com/photo-1518709268805-4e9042af9f23?w=600&auto=format&fit=crop&q=80',
        anoLancamento: 2015,
        plataforma: 'PlayStation 4',
        horasJogadas: 65,
        nota: 10,
      },
    },
  ],
  desafiosAtivos: [
    {
      id: 'desafio-1',
      titulo: 'Desafio Anual 2026',
      descricao: 'Meta pessoal de jogos zerados durante o ano.',
      progressoAtual: 18,
      progressoMeta: 30,
      unidade: 'jogos',
      prazo: '31/12/2026',
    },
    {
      id: 'desafio-2',
      titulo: 'Maratona Soulsborne',
      descricao: 'Finalizar todos os títulos da FromSoftware.',
      progressoAtual: 5,
      progressoMeta: 7,
      unidade: 'jogos',
      prazo: '30/11/2026',
    },
    {
      id: 'desafio-3',
      titulo: 'Backlog Retrô 16-bits',
      descricao: 'Revisitar clássicos esquecidos do SNES e Mega Drive.',
      progressoAtual: 4,
      progressoMeta: 10,
      unidade: 'jogos',
      prazo: '15/10/2026',
    },
    {
      id: 'desafio-4',
      titulo: 'RPGs com mais de 50h',
      descricao: 'Grandes jornadas finalizadas neste semestre.',
      progressoAtual: 2,
      progressoMeta: 5,
      unidade: 'jogos',
      prazo: '31/12/2026',
    },
  ],
  atividadeAno: {
    ano: 2026,
    totalHoras: 342,
    diasAtivos: 184,
    pontos: gerarPontosAtividade(2026),
  },
  zeradosRecentemente: [
    {
      id: 'zr-1',
      titulo: 'Metroid Prime Remastered',
      capaUrl: 'https://images.unsplash.com/photo-1579373903781-fd5c0c30c4cd?w=600&auto=format&fit=crop&q=80',
      plataforma: 'Nintendo Switch',
      horasJogadas: 14,
      nota: 9.5,
      dataFinalizacao: '14/09/2026',
      dificuldade: 'A',
    },
    {
      id: 'zr-2',
      titulo: 'Cyberpunk 2077: Phantom Liberty',
      capaUrl: 'https://images.unsplash.com/photo-1551103782-8ab07afd45c1?w=600&auto=format&fit=crop&q=80',
      plataforma: 'PC',
      horasJogadas: 38,
      nota: 9.0,
      dataFinalizacao: '02/09/2026',
      dificuldade: 'AA',
    },
    {
      id: 'zr-3',
      titulo: 'Resident Evil 4 Remake',
      capaUrl: 'https://images.unsplash.com/photo-1509198397868-475647b2a1e5?w=600&auto=format&fit=crop&q=80',
      plataforma: 'PlayStation 5',
      horasJogadas: 22,
      nota: 9.5,
      dataFinalizacao: '20/08/2026',
      dificuldade: 'A',
    },
    {
      id: 'zr-4',
      titulo: 'Sea of Stars',
      capaUrl: 'https://images.unsplash.com/photo-1563089145-599997674d42?w=600&auto=format&fit=crop&q=80',
      plataforma: 'PC',
      horasJogadas: 32,
      nota: 8.5,
      dataFinalizacao: '05/08/2026',
      dificuldade: 'B',
    },
  ],
  distribuicaoPlataformas: [
    {
      plataforma: 'PC',
      totalJogos: 62,
      horasJogadas: 1320,
      percentual: 48,
    },
    {
      plataforma: 'PlayStation 5',
      totalJogos: 34,
      horasJogadas: 640,
      percentual: 27,
    },
    {
      plataforma: 'Nintendo Switch',
      totalJogos: 22,
      horasJogadas: 350,
      percentual: 17,
    },
    {
      plataforma: 'Xbox Series X',
      totalJogos: 10,
      horasJogadas: 140,
      percentual: 8,
    },
  ],
  principaisGeneros: [
    {
      genero: 'RPG',
      totalJogos: 44,
      horasJogadas: 1120,
      percentual: 34,
    },
    {
      genero: 'Ação / Aventura',
      totalJogos: 36,
      horasJogadas: 620,
      percentual: 28,
    },
    {
      genero: 'Soulslike',
      totalJogos: 20,
      horasJogadas: 390,
      percentual: 16,
    },
    {
      genero: 'Metroidvania',
      totalJogos: 16,
      horasJogadas: 180,
      percentual: 12,
    },
    {
      genero: 'Estratégia',
      totalJogos: 12,
      horasJogadas: 140,
      percentual: 10,
    },
  ],
}
