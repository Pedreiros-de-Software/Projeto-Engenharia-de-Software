import type {
  Opcao,
  ResultadoConsulta,
  ResultadoEleicao,
  ResultadoPauta,
  VotacaoDetalhe,
  VotacaoResumo,
} from "../api/types";

export function votacaoResumo(parcial: Partial<VotacaoResumo> = {}): VotacaoResumo {
  return {
    id: 1,
    tipo: "pauta",
    titulo: "Pauta #02 — Reforma da portaria",
    descricao: "",
    inicio: "2026-11-18T13:00:00Z",
    fim: "2026-11-25T13:00:00Z",
    turno: 1,
    status: "aberta",
    apto: true,
    ja_votou: false,
    votacao_anterior_id: null,
    segundo_turno_id: null,
    ...parcial,
  };
}

export function votacaoDetalhe(parcial: Partial<VotacaoDetalhe> = {}): VotacaoDetalhe {
  return {
    ...votacaoResumo(),
    opcoes: [],
    orcamento_centavos: null,
    minha_unidade: {
      identificador: "Apto 101",
      fracao_ideal: 60,
      percentual: 12,
      adimplente: true,
      elegivel: true,
      motivo: "",
    },
    ...parcial,
  };
}

export const CHAPAS: Opcao[] = [
  { id: 11, nome: "Chapa 1 — Movimenta", descricao: "Ana / Victor / Amanda", custo_centavos: null },
  { id: 12, nome: "Chapa 2 — Renovação", descricao: "João / Ana / Marcela", custo_centavos: null },
  { id: 13, nome: "Chapa 3 — Integração", descricao: "Paulo / Rita / Caio", custo_centavos: null },
];

export const PROJETOS: Opcao[] = [
  { id: 21, nome: "Impressoras 3D e CNC", descricao: "", custo_centavos: 3_500_000 },
  { id: 22, nome: "Kit Arduino e Robótica", descricao: "", custo_centavos: 1_200_000 },
  { id: 23, nome: "Servidor GPU para IA", descricao: "", custo_centavos: 4_800_000 },
];

export function resultadoPauta(parcial: Partial<ResultadoPauta> = {}): ResultadoPauta {
  return {
    tipo: "pauta",
    votacao_id: 1,
    titulo: "Pauta #01 — Troca do sistema de câmeras",
    descricao: "",
    contagem: { favor: 300, contra: 80, abstencao: 0 },
    peso_total: 500,
    limiar: 1000 / 3,
    percentual_favor: 60,
    aprovada: false,
    unidades_total: 6,
    unidades_votantes: 5,
    ...parcial,
  };
}

export function resultadoEleicao(parcial: Partial<ResultadoEleicao> = {}): ResultadoEleicao {
  return {
    tipo: "eleicao",
    votacao_id: 3,
    titulo: "Eleição Diretoria DCE 2026",
    descricao: "",
    turno: 1,
    aptos: 8,
    participantes: 7,
    abstencoes: 1,
    percentual_comparecimento: 87.5,
    votos: [
      { opcao_id: 11, nome: "Chapa 1 — Movimenta", votos: 3, percentual_validos: 50 },
      { opcao_id: 12, nome: "Chapa 2 — Renovação", votos: 2, percentual_validos: 100 / 3 },
      { opcao_id: 13, nome: "Chapa 3 — Integração", votos: 1, percentual_validos: 100 / 6 },
    ],
    brancos: 1,
    nulos: 0,
    validos: 6,
    vencedora: null,
    segundo_turno: {
      necessario: true,
      empate_indefinido: false,
      chapas: [
        { opcao_id: 11, nome: "Chapa 1 — Movimenta" },
        { opcao_id: 12, nome: "Chapa 2 — Renovação" },
      ],
      votacao_id: null,
    },
    ...parcial,
  };
}

export function resultadoConsulta(parcial: Partial<ResultadoConsulta> = {}): ResultadoConsulta {
  return {
    tipo: "consulta",
    votacao_id: 4,
    titulo: "Orçamento participativo — Laboratório de Inovação",
    descricao: "",
    cedulas: 5,
    orcamento_centavos: 6_000_000,
    total_selecionado_centavos: 4_700_000,
    saldo_centavos: 1_300_000,
    empate_na_linha_de_corte: false,
    ranking: [
      itemRanking(1, 21, "Impressoras 3D e CNC", 3_500_000, 3_500_000, true),
      itemRanking(2, 22, "Kit Arduino e Robótica", 1_200_000, 4_700_000, true),
      itemRanking(3, 23, "Servidor GPU para IA", 4_800_000, 9_500_000, false),
      itemRanking(4, 24, "Mobiliário colaborativo", 900_000, 10_400_000, false),
    ],
    ...parcial,
  };
}

function itemRanking(
  posicao: number,
  id: number,
  nome: string,
  custo: number,
  acumulado: number,
  dentro: boolean,
) {
  return {
    posicao,
    opcao_id: id,
    nome,
    pontos: 20 - posicao * 3,
    primeiras_preferencias: 4 - posicao,
    custo_centavos: custo,
    custo_acumulado_centavos: acumulado,
    dentro_do_orcamento: dentro,
    empatado: false,
  };
}
