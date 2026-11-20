import type { NovaOpcao, NovaVotacao, Tipo } from "../../api/types";
import { reaisParaCentavos } from "../../utils/formatacao";

export const MAXIMO_OPCOES = 20;

export type OpcaoFormulario = { nome: string; descricao: string; custo: string };

export type FormularioVotacao = {
  tipo: Tipo;
  titulo: string;
  descricao: string;
  inicio: string;
  fim: string;
  opcoes: OpcaoFormulario[];
  orcamento: string;
  aptos: number[];
};

type Montagem = { ok: true; dados: NovaVotacao } | { ok: false; erro: string };

/** Valida o formulário de nova votação e monta o corpo do POST /api/votacoes. */
export function montarNovaVotacao(form: FormularioVotacao): Montagem {
  const titulo = form.titulo.trim();
  if (titulo === "" || titulo.length > 200) {
    return falha("Informe um título com até 200 caracteres.");
  }
  if (form.descricao.length > 2000) {
    return falha("A descrição pode ter no máximo 2000 caracteres.");
  }
  const inicio = new Date(form.inicio);
  const fim = new Date(form.fim);
  if (Number.isNaN(inicio.getTime()) || Number.isNaN(fim.getTime())) {
    return falha("Informe o início e o fim da votação.");
  }
  if (fim <= inicio) {
    return falha("O fim da votação precisa ser depois do início.");
  }

  const base = {
    tipo: form.tipo,
    titulo,
    descricao: form.descricao.trim(),
    inicio: inicio.toISOString(),
    fim: fim.toISOString(),
  };
  if (form.tipo === "pauta") {
    return { ok: true, dados: { ...base, opcoes: [] } };
  }

  const erroOpcoes = validarOpcoes(form.opcoes);
  if (erroOpcoes) {
    return falha(erroOpcoes);
  }
  if (form.aptos.length === 0) {
    return falha("Selecione pelo menos um eleitor apto.");
  }

  if (form.tipo === "eleicao") {
    const opcoes = form.opcoes.map((o) => ({ nome: o.nome.trim(), descricao: o.descricao.trim() }));
    return { ok: true, dados: { ...base, opcoes, aptos: form.aptos } };
  }

  const orcamento = reaisParaCentavos(form.orcamento);
  if (orcamento === null) {
    return falha("Informe o orçamento disponível em reais, por exemplo 60.000 ou 60.000,00.");
  }
  const opcoes: NovaOpcao[] = [];
  for (const opcao of form.opcoes) {
    const custo = reaisParaCentavos(opcao.custo);
    if (custo === null) {
      return falha(`Informe o custo em reais do projeto "${opcao.nome.trim()}", por exemplo 35.000 ou 35.000,50.`);
    }
    opcoes.push({ nome: opcao.nome.trim(), descricao: opcao.descricao.trim(), custo_centavos: custo });
  }
  return { ok: true, dados: { ...base, opcoes, orcamento_centavos: orcamento, aptos: form.aptos } };
}

function validarOpcoes(opcoes: OpcaoFormulario[]): string | null {
  const nomes = opcoes.map((o) => o.nome.trim());
  if (nomes.length < 2) {
    return "Cadastre pelo menos 2 opções.";
  }
  if (nomes.length > MAXIMO_OPCOES) {
    return `Cadastre no máximo ${MAXIMO_OPCOES} opções.`;
  }
  if (nomes.some((nome) => nome === "" || nome.length > 120)) {
    return "Cada opção precisa de um nome com até 120 caracteres.";
  }
  if (new Set(nomes).size !== nomes.length) {
    return "Os nomes das opções não podem se repetir.";
  }
  return null;
}

function falha(erro: string): Montagem {
  return { ok: false, erro };
}
