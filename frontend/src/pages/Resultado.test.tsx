import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import * as api from "../api/client";
import { ApiErro } from "../api/erros";
import type { Resultado as ResultadoApi, Usuario } from "../api/types";
import { resultadoConsulta, resultadoEleicao, resultadoPauta, votacaoDetalhe } from "../test/fixtures";
import { ADMIN, ELEITORA, renderizar } from "../test/renderizar";
import Resultado from "./Resultado";

vi.mock("../api/client");

function renderizarResultado(resultado: ResultadoApi, usuario: Usuario | null = null) {
  vi.mocked(api.obterResultado).mockResolvedValue(resultado);
  renderizar(<Resultado />, {
    usuario,
    rota: `/votacoes/${resultado.votacao_id}/resultado`,
    caminho: "/votacoes/:id/resultado",
  });
}

function linhaDaTabela(nome: string) {
  return screen.getByRole("rowheader", { name: new RegExp(`^${nome}`) }).closest("tr") as HTMLElement;
}

function celulas(linha: HTMLElement) {
  return within(linha)
    .getAllByRole("cell")
    .map((c) => c.textContent?.replace(/\s/g, " "));
}

describe("Resultado: carregamento", () => {
  it("sem login mostra o título e a descrição que vêm no próprio resultado", async () => {
    renderizarResultado(
      resultadoPauta({ votacao_id: 1, titulo: "Pauta #01 — Troca do sistema de câmeras", descricao: "Assembleia de março." }),
    );

    expect(
      await screen.findByRole("heading", { level: 1, name: "Pauta #01 — Troca do sistema de câmeras" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Assembleia de março.")).toBeInTheDocument();
    expect(screen.getByText("REPROVADA")).toBeInTheDocument();
    expect(api.obterResultado).toHaveBeenCalledWith(1);
    expect(api.obterVotacao).not.toHaveBeenCalled();
    expect(screen.queryByRole("link", { name: "Voltar ao painel" })).not.toBeInTheDocument();
  });

  it("título vazio no resultado usa o número da votação", async () => {
    renderizarResultado(resultadoConsulta({ votacao_id: 4, titulo: "", descricao: "" }));

    expect(await screen.findByRole("heading", { level: 1, name: "Votação nº 4" })).toBeInTheDocument();
  });

  it("com login mostra o mesmo título e o link para o painel", async () => {
    renderizarResultado(resultadoEleicao({ titulo: "Eleição Diretoria DCE 2026" }), ELEITORA);

    expect(await screen.findByRole("heading", { level: 1, name: "Eleição Diretoria DCE 2026" })).toBeInTheDocument();
    expect(screen.getByText("Resultado · Eleição do CA")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Voltar ao painel" })).toBeInTheDocument();
    expect(api.obterVotacao).not.toHaveBeenCalled();
  });

  it("votação ainda aberta mostra a mensagem do servidor", async () => {
    vi.mocked(api.obterResultado).mockRejectedValue(
      new ApiErro(409, "votacao_nao_encerrada", "O resultado só fica disponível após o encerramento."),
    );
    renderizar(<Resultado />, { usuario: null, rota: "/votacoes/2/resultado", caminho: "/votacoes/:id/resultado" });

    expect(await screen.findByRole("alert")).toHaveTextContent("O resultado só fica disponível após o encerramento.");
  });
});

describe("Resultado: pauta (quórum de 2/3)", () => {
  it("reprovada com 60% a favor: mostra pesos, % do total, limiar e unidades", async () => {
    renderizarResultado(resultadoPauta());

    expect(await screen.findByText("REPROVADA")).toBeInTheDocument();
    expect(screen.getByText(/60,00% do peso total a favor \(mínimo para aprovar: 66,67%\)/)).toBeInTheDocument();
    expect(celulas(linhaDaTabela("A favor"))).toEqual(["300", "60,00%"]);
    expect(celulas(linhaDaTabela("Contra"))).toEqual(["80", "16,00%"]);
    expect(celulas(linhaDaTabela("Abstenção"))).toEqual(["0", "0,00%"]);
    expect(celulas(linhaDaTabela("Não votaram"))).toEqual(["120", "24,00%"]);
    expect(screen.getByText("Limiar de aprovação (2/3)").nextSibling).toHaveTextContent("333,33");
    expect(screen.getByText("Peso total cadastrado").nextSibling).toHaveTextContent("500");
    expect(screen.getByText("Unidades que votaram").nextSibling).toHaveTextContent("5 de 6");
    expect(
      screen.getByText(/aprovação exige ≥ 2\/3 do peso total cadastrado, incluindo unidades que não votaram/),
    ).toBeInTheDocument();
  });

  it("aprovada quando o peso a favor atinge o limiar", async () => {
    renderizarResultado(
      resultadoPauta({ contagem: { favor: 380, contra: 0, abstencao: 0 }, percentual_favor: 76, aprovada: true }),
    );

    expect(await screen.findByText("APROVADA")).toBeInTheDocument();
    expect(screen.queryByText("REPROVADA")).not.toBeInTheDocument();
  });

  it("peso total zero não gera divisão inválida", async () => {
    renderizarResultado(
      resultadoPauta({ contagem: { favor: 0, contra: 0, abstencao: 0 }, peso_total: 0, limiar: 0, percentual_favor: 0 }),
    );

    expect(await screen.findByText("REPROVADA")).toBeInTheDocument();
    expect(celulas(linhaDaTabela("A favor"))).toEqual(["0", "0,00%"]);
  });
});

describe("Resultado: boletim da eleição", () => {
  it("mostra comparecimento, votos por chapa, brancos, nulos e a base dos percentuais", async () => {
    renderizarResultado(resultadoEleicao());

    expect(await screen.findByText("Eleitores aptos")).toBeInTheDocument();
    expect(screen.getByText("Participantes").nextSibling).toHaveTextContent("7");
    expect(screen.getByText("Comparecimento de 87,50%")).toBeInTheDocument();
    expect(screen.getByText("Abstenções").nextSibling).toHaveTextContent("1");
    expect(celulas(linhaDaTabela("Chapa 1 — Movimenta"))).toEqual(["3", "50,00%"]);
    expect(celulas(linhaDaTabela("Chapa 2 — Renovação"))).toEqual(["2", "33,33%"]);
    expect(celulas(linhaDaTabela("Brancos"))).toEqual(["1", "—"]);
    expect(celulas(linhaDaTabela("Nulos"))).toEqual(["0", "—"]);
    expect(
      screen.getByText(/Percentuais calculados sobre 6 votos válidos \(brancos e nulos excluídos\)/),
    ).toBeInTheDocument();
  });

  it("sem maioria absoluta anuncia o 2º turno entre as duas mais votadas", async () => {
    renderizarResultado(resultadoEleicao(), ELEITORA);

    expect(
      await screen.findByText("2º turno entre Chapa 1 — Movimenta e Chapa 2 — Renovação"),
    ).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Abrir 2º turno" })).not.toBeInTheDocument();
  });

  it("admin abre o 2º turno com a data de término escolhida", async () => {
    vi.mocked(api.abrirSegundoTurno).mockResolvedValue(
      votacaoDetalhe({ id: 30, tipo: "eleicao", turno: 2, fim: "2026-11-20T21:00:00Z" }),
    );
    renderizarResultado(resultadoEleicao(), ADMIN);

    const campo = await screen.findByLabelText("Término do 2º turno");
    await userEvent.clear(campo);
    await userEvent.type(campo, "2026-11-20T18:00");
    await userEvent.click(screen.getByRole("button", { name: "Abrir 2º turno" }));

    expect(api.abrirSegundoTurno).toHaveBeenCalledWith(3, "2026-11-20T21:00:00.000Z");
    expect(await screen.findByText("2º turno aberto")).toBeInTheDocument();
    expect(screen.getByText(/Votação aberta até 20\/11\/2026 às 18:00/)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Ir para o 2º turno" })).toHaveAttribute("href", "/votacoes/30");
  });

  it("mostra o erro do servidor ao abrir o 2º turno", async () => {
    vi.mocked(api.abrirSegundoTurno).mockRejectedValue(
      new ApiErro(409, "segundo_turno_indisponivel", "O 2º turno já foi aberto."),
    );
    renderizarResultado(resultadoEleicao(), ADMIN);

    await userEvent.click(await screen.findByRole("button", { name: "Abrir 2º turno" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("O 2º turno já foi aberto.");
    expect(screen.getByRole("button", { name: "Abrir 2º turno" })).toBeEnabled();
  });

  it("quando o 2º turno já existe, mostra o link para ele em vez do botão", async () => {
    renderizarResultado(
      resultadoEleicao({ segundo_turno: { ...resultadoEleicao().segundo_turno, votacao_id: 30 } }),
      ADMIN,
    );

    expect(await screen.findByRole("link", { name: "Ir para o 2º turno" })).toHaveAttribute("href", "/votacoes/30");
    expect(screen.queryByRole("button", { name: "Abrir 2º turno" })).not.toBeInTheDocument();
  });

  it("sem login, o link do 2º turno leva ao boletim dele (a página de votar exige login)", async () => {
    renderizarResultado(resultadoEleicao({ segundo_turno: { ...resultadoEleicao().segundo_turno, votacao_id: 30 } }));

    expect(await screen.findByRole("link", { name: "Ir para o 2º turno" })).toHaveAttribute(
      "href",
      "/votacoes/30/resultado",
    );
    expect(screen.getByText(/Vence no 1º turno a chapa com mais da metade dos votos válidos/)).toBeInTheDocument();
  });

  it("anuncia a chapa vencedora com maioria absoluta", async () => {
    renderizarResultado(
      resultadoEleicao({
        turno: 2,
        vencedora: { opcao_id: 11, nome: "Chapa 1 — Movimenta" },
        segundo_turno: { necessario: false, empate_indefinido: false, chapas: [], votacao_id: null },
      }),
    );

    expect(await screen.findByText("Chapa vencedora")).toBeInTheDocument();
    expect(screen.getByText("Chapa vencedora").nextSibling).toHaveTextContent("Chapa 1 — Movimenta");
    expect(screen.getByText("Boletim de apuração · 2º turno")).toBeInTheDocument();
    expect(screen.getByText(/No 2º turno vence a chapa mais votada/)).toBeInTheDocument();
    expect(screen.queryByText(/Vence no 1º turno/)).not.toBeInTheDocument();
  });

  it("empate pelo 2º lugar: o sistema não escolhe e indica decisão da comissão", async () => {
    renderizarResultado(
      resultadoEleicao({
        segundo_turno: { necessario: true, empate_indefinido: true, chapas: [], votacao_id: null },
      }),
      ADMIN,
    );

    expect(await screen.findByText("Empate impede definir o 2º turno: decisão da comissão")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Abrir 2º turno" })).not.toBeInTheDocument();
  });

  it("sem votos válidos não há vencedora", async () => {
    renderizarResultado(
      resultadoEleicao({
        votos: [],
        brancos: 2,
        validos: 0,
        segundo_turno: { necessario: false, empate_indefinido: false, chapas: [], votacao_id: null },
      }),
    );

    expect(await screen.findByText("Sem votos válidos: sem vencedora")).toBeInTheDocument();
  });

  it("empate no 2º turno indica decisão da comissão", async () => {
    renderizarResultado(
      resultadoEleicao({
        turno: 2,
        segundo_turno: { necessario: true, empate_indefinido: false, chapas: [], votacao_id: null },
      }),
    );

    expect(await screen.findByText("Empate no 2º turno: decisão da comissão")).toBeInTheDocument();
  });

  it("sem desfecho definido pelo servidor mostra sem vencedora", async () => {
    renderizarResultado(
      resultadoEleicao({
        segundo_turno: { necessario: false, empate_indefinido: false, chapas: [], votacao_id: null },
      }),
    );

    expect(await screen.findByText("Sem vencedora definida")).toBeInTheDocument();
  });
});

describe("Resultado: ranking da consulta", () => {
  function linhasDoRanking() {
    const tabela = screen.getByRole("table", { name: "Ranking pela pontuação Borda" });
    return within(tabela).getAllByRole("row").slice(1);
  }

  it("mostra a linha de corte entre o último dentro e o primeiro fora do orçamento", async () => {
    renderizarResultado(resultadoConsulta());
    await screen.findByText("Total selecionado");

    const textos = linhasDoRanking().map((linha) => linha.textContent ?? "");
    expect(textos).toHaveLength(5);
    expect(textos[0]).toContain("Impressoras 3D e CNC");
    expect(textos[0]).toContain("Dentro do orçamento");
    expect(textos[1]).toContain("Kit Arduino e Robótica");
    expect(textos[2]).toBe("Linha de corte do orçamento");
    expect(textos[3]).toContain("Servidor GPU para IA");
    expect(textos[3]).toContain("Fora do orçamento");
    expect(textos[4]).toContain("Fora do orçamento");
  });

  it.each([
    { caso: "ranking da consulta", resultado: resultadoConsulta(), rotulo: "Ranking pela pontuação Borda" },
    { caso: "pauta", resultado: resultadoPauta(), rotulo: "Apuração ponderada pela fração ideal" },
    { caso: "boletim", resultado: resultadoEleicao(), rotulo: "Boletim de apuração · 1º turno" },
  ])("tabela do $caso rola na horizontal em telas estreitas, também pelo teclado", async ({ resultado, rotulo }) => {
    renderizarResultado(resultado);

    const rolagem = await screen.findByRole("region", { name: rotulo });
    expect(rolagem).toHaveClass("tabela-rolagem");
    expect(rolagem).toHaveAttribute("tabindex", "0");
    expect(within(rolagem).getByRole("table", { name: rotulo })).toBeInTheDocument();
  });

  it("mostra pontos, 1ªs preferências, custo e custo acumulado de cada projeto", async () => {
    renderizarResultado(resultadoConsulta());
    await screen.findByText("Total selecionado");

    expect(celulas(linhaDaTabela("Kit Arduino e Robótica"))).toEqual([
      "2º",
      "14",
      "2",
      "R$ 12.000,00",
      "R$ 47.000,00",
      "Dentro do orçamento",
    ]);
  });

  it("mostra orçamento, total selecionado, saldo e a legenda Borda", async () => {
    renderizarResultado(resultadoConsulta());

    expect(await screen.findByText("Orçamento")).toBeInTheDocument();
    expect(screen.getByText("Orçamento").nextSibling).toHaveTextContent("R$ 60.000,00");
    expect(screen.getByText("Total selecionado").nextSibling).toHaveTextContent("R$ 47.000,00");
    expect(screen.getByText("Saldo").nextSibling).toHaveTextContent("R$ 13.000,00");
    expect(screen.getByText("Cédulas apuradas").nextSibling).toHaveTextContent("5");
    expect(screen.getByText(/com 4 projetos, a 1ª posição de cada cédula vale 4 pontos, a 2ª vale 3/)).toBeInTheDocument();
  });

  it("linha de corte fica no topo quando nenhum projeto cabe", async () => {
    const base = resultadoConsulta();
    renderizarResultado({ ...base, ranking: base.ranking.map((p) => ({ ...p, dentro_do_orcamento: false })) });
    await screen.findByText("Total selecionado");
    expect(linhasDoRanking()[0]).toHaveTextContent("Linha de corte do orçamento");
  });

  it("todos dentro do orçamento: não desenha a linha de corte", async () => {
    const base = resultadoConsulta();
    renderizarResultado({ ...base, ranking: base.ranking.map((p) => ({ ...p, dentro_do_orcamento: true })) });
    await screen.findByText("Total selecionado");

    expect(screen.queryByText("Linha de corte do orçamento")).not.toBeInTheDocument();
  });

  it("marca projetos empatados e avisa empate na linha de corte", async () => {
    const base = resultadoConsulta();
    renderizarResultado({
      ...base,
      empate_na_linha_de_corte: true,
      ranking: base.ranking.map((p, i) => ({ ...p, empatado: i === 1 || i === 2 })),
    });

    expect(await screen.findByText("Empate na linha de corte")).toBeInTheDocument();
    expect(within(linhaDaTabela("Kit Arduino e Robótica")).getByText("Empatado")).toBeInTheDocument();
    expect(within(linhaDaTabela("Servidor GPU para IA")).getByText("Empatado")).toBeInTheDocument();
    expect(within(linhaDaTabela("Impressoras 3D e CNC")).queryByText("Empatado")).not.toBeInTheDocument();
  });
});
