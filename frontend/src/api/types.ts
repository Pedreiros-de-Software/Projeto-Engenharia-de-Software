export type Papel = "admin" | "eleitor";

export type Usuario = {
  id: number;
  nome: string;
  login: string;
  papel: Papel;
};

export type Unidade = {
  id: number;
  identificador: string;
  fracao_ideal: number;
  percentual: number;
  adimplente: boolean;
  responsavel: { id: number; nome: string } | null;
};

export type Status = "nao_iniciada" | "aberta" | "encerrada";

export type Tipo = "pauta" | "eleicao" | "consulta";

export type VotacaoResumo = {
  id: number;
  tipo: Tipo;
  titulo: string;
  descricao: string;
  inicio: string;
  fim: string;
  turno: 1 | 2;
  status: Status;
  apto: boolean;
  ja_votou: boolean;
  votacao_anterior_id: number | null;
  segundo_turno_id: number | null;
};

export type Opcao = {
  id: number;
  nome: string;
  descricao: string;
  custo_centavos: number | null;
};

export type MinhaUnidade = {
  identificador: string;
  fracao_ideal: number;
  percentual: number;
  adimplente: boolean;
  elegivel: boolean;
  motivo: string;
};

export type VotacaoDetalhe = VotacaoResumo & {
  opcoes: Opcao[];
  orcamento_centavos: number | null;
  minha_unidade: MinhaUnidade | null;
};

export type ResultadoPauta = {
  tipo: "pauta";
  votacao_id: number;
  titulo: string;
  descricao: string;
  contagem: { favor: number; contra: number; abstencao: number };
  peso_total: number;
  limiar: number;
  percentual_favor: number;
  aprovada: boolean;
  unidades_total: number;
  unidades_votantes: number;
};

export type ChapaResumo = { opcao_id: number; nome: string };

export type VotosChapa = {
  opcao_id: number;
  nome: string;
  votos: number;
  percentual_validos: number;
};

export type ResultadoEleicao = {
  tipo: "eleicao";
  votacao_id: number;
  titulo: string;
  descricao: string;
  turno: 1 | 2;
  aptos: number;
  participantes: number;
  abstencoes: number;
  percentual_comparecimento: number;
  votos: VotosChapa[];
  brancos: number;
  nulos: number;
  validos: number;
  vencedora: ChapaResumo | null;
  segundo_turno: {
    necessario: boolean;
    empate_indefinido: boolean;
    chapas: ChapaResumo[];
    votacao_id: number | null;
  };
};

export type ItemRanking = {
  posicao: number;
  opcao_id: number;
  nome: string;
  pontos: number;
  primeiras_preferencias: number;
  custo_centavos: number;
  custo_acumulado_centavos: number;
  dentro_do_orcamento: boolean;
  empatado: boolean;
};

export type ResultadoConsulta = {
  tipo: "consulta";
  votacao_id: number;
  titulo: string;
  descricao: string;
  cedulas: number;
  orcamento_centavos: number;
  total_selecionado_centavos: number;
  saldo_centavos: number;
  empate_na_linha_de_corte: boolean;
  ranking: ItemRanking[];
};

export type Resultado = ResultadoPauta | ResultadoEleicao | ResultadoConsulta;

export type RespostaLogin = { token: string; usuario: Usuario };

export type RespostaMe = { usuario: Usuario; unidade: Unidade | null };

export type NovaOpcao = {
  nome: string;
  descricao: string;
  custo_centavos?: number;
};

export type NovaVotacao = {
  tipo: Tipo;
  titulo: string;
  descricao: string;
  inicio: string;
  fim: string;
  opcoes: NovaOpcao[];
  orcamento_centavos?: number;
  aptos?: number[];
};

export type VotoPauta = { escolha: "favor" | "contra" | "abstencao" };

export type VotoEleicao =
  | { escolha: "chapa"; opcao_id: number }
  | { escolha: "branco" | "nulo" };

export type VotoConsulta = { preferencias: number[] };

export type Voto = VotoPauta | VotoEleicao | VotoConsulta;

export type RespostaVoto = { mensagem: string };

export type CodigoErro =
  | "requisicao_invalida"
  | "cedula_invalida"
  | "opcao_desconhecida"
  | "nao_autenticado"
  | "credenciais_invalidas"
  | "acesso_negado"
  | "nao_apto"
  | "inadimplente"
  | "nao_encontrado"
  | "voto_duplicado"
  | "votacao_nao_aberta"
  | "votacao_nao_encerrada"
  | "segundo_turno_indisponivel"
  | "muitas_tentativas"
  | "erro_interno";

export type CorpoErro = { erro: string; codigo: CodigoErro };
