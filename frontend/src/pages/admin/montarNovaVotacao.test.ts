import { describe, expect, it } from "vitest";

import { montarNovaVotacao, type FormularioVotacao } from "./montarNovaVotacao";

function formulario(parcial: Partial<FormularioVotacao> = {}): FormularioVotacao {
  return {
    tipo: "consulta",
    titulo: "Orçamento participativo",
    descricao: "",
    inicio: "2026-12-01T09:00",
    fim: "2026-12-08T18:00",
    opcoes: [
      { nome: "Projeto A", descricao: "", custo: "100" },
      { nome: "Projeto B", descricao: "", custo: "200" },
    ],
    orcamento: "250",
    aptos: [10],
    ...parcial,
  };
}

const opcao = (nome: string, custo = "1") => ({ nome, descricao: "", custo });

describe("montarNovaVotacao", () => {
  it("monta o corpo da consulta com valores em centavos e datas em UTC", () => {
    expect(montarNovaVotacao(formulario())).toEqual({
      ok: true,
      dados: {
        tipo: "consulta",
        titulo: "Orçamento participativo",
        descricao: "",
        inicio: "2026-12-01T12:00:00.000Z",
        fim: "2026-12-08T21:00:00.000Z",
        opcoes: [
          { nome: "Projeto A", descricao: "", custo_centavos: 10_000 },
          { nome: "Projeto B", descricao: "", custo_centavos: 20_000 },
        ],
        orcamento_centavos: 25_000,
        aptos: [10],
      },
    });
  });

  it("entende valores no formato brasileiro, com ponto de milhar e vírgula decimal", () => {
    const resultado = montarNovaVotacao(
      formulario({
        opcoes: [
          { nome: "Impressoras", descricao: "", custo: "35.000" },
          { nome: "Arduino", descricao: "", custo: "12.000,50" },
        ],
        orcamento: "60.000,00",
      }),
    );
    expect(resultado).toMatchObject({
      ok: true,
      dados: {
        opcoes: [
          { nome: "Impressoras", custo_centavos: 3_500_000 },
          { nome: "Arduino", custo_centavos: 1_200_050 },
        ],
        orcamento_centavos: 6_000_000,
      },
    });
  });

  it.each([
    { caso: "título vazio", parcial: { titulo: "   " }, erro: "Informe um título com até 200 caracteres." },
    { caso: "título longo", parcial: { titulo: "x".repeat(201) }, erro: "Informe um título com até 200 caracteres." },
    { caso: "descrição longa", parcial: { descricao: "x".repeat(2001) }, erro: "A descrição pode ter no máximo 2000 caracteres." },
    { caso: "início vazio", parcial: { inicio: "" }, erro: "Informe o início e o fim da votação." },
    { caso: "fim igual ao início", parcial: { fim: "2026-12-01T09:00" }, erro: "O fim da votação precisa ser depois do início." },
    { caso: "uma opção só", parcial: { opcoes: [opcao("A")] }, erro: "Cadastre pelo menos 2 opções." },
    {
      caso: "mais de 20 opções",
      parcial: { opcoes: Array.from({ length: 21 }, (_, i) => opcao(`P${i}`)) },
      erro: "Cadastre no máximo 20 opções.",
    },
    { caso: "opção sem nome", parcial: { opcoes: [opcao("A"), opcao(" ")] }, erro: "Cada opção precisa de um nome com até 120 caracteres." },
    { caso: "nomes repetidos", parcial: { opcoes: [opcao("A"), opcao(" A ")] }, erro: "Os nomes das opções não podem se repetir." },
    { caso: "sem aptos", parcial: { aptos: [] }, erro: "Selecione pelo menos um eleitor apto." },
    { caso: "orçamento vazio", parcial: { orcamento: "" }, erro: "Informe o orçamento disponível em reais, por exemplo 60.000 ou 60.000,00." },
    { caso: "orçamento em formato ambíguo", parcial: { orcamento: "1.5000" }, erro: "Informe o orçamento disponível em reais, por exemplo 60.000 ou 60.000,00." },
    {
      caso: "custo negativo",
      parcial: { opcoes: [opcao("A"), opcao("B", "-5")] },
      erro: 'Informe o custo em reais do projeto "B", por exemplo 35.000 ou 35.000,50.',
    },
  ])("rejeita $caso", ({ parcial, erro }) => {
    expect(montarNovaVotacao(formulario(parcial))).toEqual({ ok: false, erro });
  });

  it("pauta ignora opções, orçamento e aptos", () => {
    const resultado = montarNovaVotacao(formulario({ tipo: "pauta", opcoes: [], aptos: [], orcamento: "" }));
    expect(resultado).toMatchObject({ ok: true, dados: { tipo: "pauta", opcoes: [] } });
    expect(resultado.ok && "aptos" in resultado.dados).toBe(false);
  });

  it("eleição não envia custo nem orçamento", () => {
    const resultado = montarNovaVotacao(formulario({ tipo: "eleicao", orcamento: "" }));
    expect(resultado).toMatchObject({
      ok: true,
      dados: { opcoes: [{ nome: "Projeto A", descricao: "" }, { nome: "Projeto B", descricao: "" }], aptos: [10] },
    });
    expect(resultado.ok && "orcamento_centavos" in resultado.dados).toBe(false);
  });
});
