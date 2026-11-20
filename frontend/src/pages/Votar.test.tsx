import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import * as api from "../api/client";
import { ApiErro } from "../api/erros";
import type { VotacaoDetalhe } from "../api/types";
import { CHAPAS, PROJETOS, votacaoDetalhe } from "../test/fixtures";
import { ADMIN, renderizar } from "../test/renderizar";
import Votar from "./Votar";

vi.mock("../api/client");

function renderizarVotar(votacao: VotacaoDetalhe) {
  vi.mocked(api.obterVotacao).mockResolvedValue(votacao);
  renderizar(<Votar />, { rota: `/votacoes/${votacao.id}`, caminho: "/votacoes/:id" });
}

async function confirmarNoDialogo(nomeBotao: string | RegExp) {
  const dialogo = await screen.findByRole("dialog");
  await userEvent.click(within(dialogo).getByRole("button", { name: nomeBotao }));
}

describe("Votar: estados da votação", () => {
  it("carrega a votação pelo id da rota e mostra título e status", async () => {
    renderizarVotar(votacaoDetalhe({ id: 9, titulo: "Pauta #02" }));

    expect(await screen.findByRole("heading", { level: 1, name: "Pauta #02" })).toBeInTheDocument();
    expect(api.obterVotacao).toHaveBeenCalledWith(9);
    expect(screen.getByText("Assembleia de condomínio")).toBeInTheDocument();
  });

  it("votação encerrada não permite votar e aponta para o resultado", async () => {
    renderizarVotar(votacaoDetalhe({ id: 3, status: "encerrada" }));

    expect(await screen.findByText("Votação encerrada")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Ver resultado" })).toHaveAttribute("href", "/votacoes/3/resultado");
    expect(screen.queryByRole("button", { name: "Aprovar" })).not.toBeInTheDocument();
  });

  it("votação não iniciada informa quando abre", async () => {
    renderizarVotar(votacaoDetalhe({ status: "nao_iniciada" }));

    expect(await screen.findByText("Votação não iniciada")).toBeInTheDocument();
    expect(screen.getByText(/A votação abre em 18\/11\/2026 às 10:00/)).toBeInTheDocument();
  });

  it("quem já votou vê o aviso em vez da cédula", async () => {
    renderizarVotar(votacaoDetalhe({ tipo: "eleicao", opcoes: CHAPAS, ja_votou: true }));

    expect(await screen.findByText("Você já votou nesta votação")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Chapa 1/ })).not.toBeInTheDocument();
  });

  it("eleitor fora da lista de aptos não recebe a cédula", async () => {
    renderizarVotar(votacaoDetalhe({ tipo: "consulta", opcoes: PROJETOS, apto: false }));

    expect(await screen.findByText("Você não está apto")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Enviar cédula" })).not.toBeInTheDocument();
  });

  it.each([
    { caso: "eleição aberta", votacao: votacaoDetalhe({ tipo: "eleicao", opcoes: CHAPAS, apto: false, minha_unidade: null }) },
    { caso: "pauta aberta", votacao: votacaoDetalhe({ minha_unidade: null }) },
  ])("admin numa $caso vê o aviso de acompanhamento em vez da cédula", async ({ votacao }) => {
    vi.mocked(api.obterVotacao).mockResolvedValue(votacao);
    renderizar(<Votar />, { usuario: ADMIN, rota: `/votacoes/${votacao.id}`, caminho: "/votacoes/:id" });

    const aviso = (await screen.findByText("Acompanhamento")).closest(".alerta") as HTMLElement;
    expect(aviso).toHaveRole("status");
    expect(aviso).toHaveTextContent("Administradores não votam");
    expect(within(aviso).getByRole("link", { name: "Painel" })).toHaveAttribute("href", "/");
    expect(screen.queryByText("Você não está apto")).not.toBeInTheDocument();
    expect(screen.queryByText("Sem unidade vinculada")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Chapa 1|Aprovar/ })).not.toBeInTheDocument();
  });

  it("admin numa votação encerrada continua vendo o link para o resultado", async () => {
    vi.mocked(api.obterVotacao).mockResolvedValue(votacaoDetalhe({ id: 3, status: "encerrada" }));
    renderizar(<Votar />, { usuario: ADMIN, rota: "/votacoes/3", caminho: "/votacoes/:id" });

    expect(await screen.findByText("Votação encerrada")).toBeInTheDocument();
    expect(screen.queryByText("Acompanhamento")).not.toBeInTheDocument();
  });

  it("mostra o erro quando a votação não existe", async () => {
    vi.mocked(api.obterVotacao).mockRejectedValue(new ApiErro(404, "nao_encontrado", "Votação não encontrada."));
    renderizar(<Votar />, { rota: "/votacoes/99", caminho: "/votacoes/:id" });

    expect(await screen.findByRole("alert")).toHaveTextContent("Votação não encontrada.");
  });
});

describe("Votar: pauta de condomínio", () => {
  const unidadeInadimplente = {
    identificador: "Apto 402",
    fracao_ideal: 120,
    percentual: 24,
    adimplente: false,
    elegivel: false,
    motivo: "Unidade inadimplente: regularize os pagamentos para votar.",
  };

  it("mostra unidade, fração ideal e peso antes de confirmar e envia a escolha", async () => {
    vi.mocked(api.votar).mockResolvedValue({ mensagem: "Voto registrado com sucesso." });
    renderizarVotar(votacaoDetalhe({ id: 2 }));

    const unidade = await screen.findByRole("region", { name: "Sua unidade" });
    expect(within(unidade).getByText("Apto 101")).toBeInTheDocument();
    expect(within(unidade).getByText("60")).toBeInTheDocument();
    expect(within(unidade).getByText("12,00% do total")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Aprovar" }));
    const dialogo = screen.getByRole("dialog", { name: "Confirmar voto" });
    expect(dialogo).toHaveTextContent(
      "Seu voto: APROVAR pela unidade Apto 101, com peso de 12,00% do total (fração ideal 60).",
    );
    expect(api.votar).not.toHaveBeenCalled();

    await confirmarNoDialogo("Confirmar: Aprovar");

    expect(api.votar).toHaveBeenCalledWith(2, { escolha: "favor" });
    expect(await screen.findByRole("heading", { name: "Voto registrado" })).toBeInTheDocument();
    expect(screen.getByText("Voto registrado com sucesso.")).toBeInTheDocument();
  });

  it.each([
    { botao: "Rejeitar", escolha: "contra" },
    { botao: "Abster-se", escolha: "abstencao" },
  ])("$botao envia escolha $escolha", async ({ botao, escolha }) => {
    vi.mocked(api.votar).mockResolvedValue({ mensagem: "Voto registrado com sucesso." });
    renderizarVotar(votacaoDetalhe({ id: 2 }));

    await userEvent.click(await screen.findByRole("button", { name: botao }));
    await confirmarNoDialogo(`Confirmar: ${botao}`);

    expect(api.votar).toHaveBeenCalledWith(2, { escolha });
  });

  it("cancelar a confirmação não envia o voto", async () => {
    renderizarVotar(votacaoDetalhe());

    await userEvent.click(await screen.findByRole("button", { name: "Rejeitar" }));
    await confirmarNoDialogo("Cancelar");

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(api.votar).not.toHaveBeenCalled();
  });

  it("bloqueia o inadimplente: mostra o peso, o motivo e nenhum botão de voto", async () => {
    renderizarVotar(votacaoDetalhe({ minha_unidade: unidadeInadimplente }));

    expect(await screen.findByText("Apto 402")).toBeInTheDocument();
    expect(screen.getByText("24,00% do total")).toBeInTheDocument();
    expect(screen.getByText("Inadimplente")).toBeInTheDocument();
    const alerta = screen.getByRole("alert");
    expect(alerta).toHaveTextContent("Voto bloqueado");
    expect(alerta).toHaveTextContent("Unidade inadimplente: regularize os pagamentos para votar.");
    for (const botao of ["Aprovar", "Rejeitar", "Abster-se"]) {
      expect(screen.queryByRole("button", { name: botao })).not.toBeInTheDocument();
    }
  });

  it("usa um motivo padrão quando o servidor não envia o texto do bloqueio", async () => {
    renderizarVotar(votacaoDetalhe({ minha_unidade: { ...unidadeInadimplente, motivo: "" } }));

    expect(await screen.findByRole("alert")).toHaveTextContent("A unidade não está apta a votar.");
  });

  it("usuário sem unidade é informado de que não pode votar na pauta", async () => {
    renderizarVotar(votacaoDetalhe({ minha_unidade: null, apto: false }));

    expect(await screen.findByText("Sem unidade vinculada")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Aprovar" })).not.toBeInTheDocument();
  });

  it("mostra o erro do servidor (ex.: voto duplicado) e mantém a cédula", async () => {
    vi.mocked(api.votar).mockRejectedValue(new ApiErro(409, "voto_duplicado", "Você já votou nesta votação."));
    renderizarVotar(votacaoDetalhe());

    await userEvent.click(await screen.findByRole("button", { name: "Aprovar" }));
    await confirmarNoDialogo("Confirmar: Aprovar");

    expect(await screen.findByRole("alert")).toHaveTextContent("Você já votou nesta votação.");
    expect(screen.getByRole("button", { name: "Aprovar" })).toBeEnabled();
  });
});

describe("Votar: eleição do CA", () => {
  const eleicao = votacaoDetalhe({
    id: 5,
    tipo: "eleicao",
    titulo: "Eleição Diretoria DCE 2026",
    opcoes: CHAPAS,
    minha_unidade: null,
  });

  it("mostra o selo de voto secreto, as chapas com descrição, branco e nulo", async () => {
    renderizarVotar(eleicao);

    expect(await screen.findByText("Voto secreto")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Chapa 1 — Movimenta Ana / Victor / Amanda" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Chapa 3 — Integração/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Votar em branco" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Votar nulo" })).toBeInTheDocument();
  });

  it("confirma a chapa, envia o voto e a tela final não mostra a escolha", async () => {
    vi.mocked(api.votar).mockResolvedValue({ mensagem: "Voto registrado com sucesso." });
    renderizarVotar(eleicao);

    await userEvent.click(await screen.findByRole("button", { name: /Chapa 2 — Renovação/ }));
    expect(screen.getByRole("dialog")).toHaveTextContent("Seu voto: Chapa 2 — Renovação");
    await confirmarNoDialogo("Confirmar voto");

    expect(api.votar).toHaveBeenCalledWith(5, { escolha: "chapa", opcao_id: 12 });
    expect(await screen.findByRole("heading", { name: "Voto registrado" })).toBeInTheDocument();
    expect(screen.getByText(/a escolha não é exibida e não fica associada ao seu nome/)).toBeInTheDocument();
    for (const chapa of CHAPAS) {
      expect(screen.queryByText(chapa.nome)).not.toBeInTheDocument();
    }
    expect(document.body).not.toHaveTextContent("Renovação");
  });

  it.each([
    { botao: "Votar em branco", voto: { escolha: "branco" }, rotulo: "Em branco" },
    { botao: "Votar nulo", voto: { escolha: "nulo" }, rotulo: "Nulo" },
  ])("$botao envia $voto.escolha sem opção", async ({ botao, voto, rotulo }) => {
    vi.mocked(api.votar).mockResolvedValue({ mensagem: "Voto registrado com sucesso." });
    renderizarVotar(eleicao);

    await userEvent.click(await screen.findByRole("button", { name: botao }));
    expect(screen.getByRole("dialog")).toHaveTextContent(`Seu voto: ${rotulo}`);
    await confirmarNoDialogo("Confirmar voto");

    expect(api.votar).toHaveBeenCalledWith(5, voto);
    expect(await screen.findByRole("heading", { name: "Voto registrado" })).toBeInTheDocument();
  });

  it("cancelar a confirmação mantém a cédula sem enviar", async () => {
    renderizarVotar(eleicao);

    await userEvent.click(await screen.findByRole("button", { name: /Chapa 1/ }));
    await confirmarNoDialogo("Cancelar");

    expect(api.votar).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: /Chapa 1/ })).toBeInTheDocument();
  });

  it("mostra a mensagem do servidor quando o eleitor não está apto", async () => {
    vi.mocked(api.votar).mockRejectedValue(new ApiErro(403, "nao_apto", "Você não está apto a votar nesta votação."));
    renderizarVotar(eleicao);

    await userEvent.click(await screen.findByRole("button", { name: "Votar nulo" }));
    await confirmarNoDialogo("Confirmar voto");

    expect(await screen.findByRole("alert")).toHaveTextContent("Você não está apto a votar nesta votação.");
  });
});

describe("Votar: consulta de priorização (Borda)", () => {
  const consulta = votacaoDetalhe({
    id: 7,
    tipo: "consulta",
    titulo: "Orçamento participativo",
    opcoes: PROJETOS,
    orcamento_centavos: 6_000_000,
    minha_unidade: null,
  });

  function listaAOrdenar() {
    return screen.getByRole("region", { name: "Projetos a ordenar" });
  }

  function suaOrdem() {
    return screen.getByRole("region", { name: "Sua ordem" });
  }

  async function adicionar(nome: string) {
    await userEvent.click(within(listaAOrdenar()).getByRole("button", { name: new RegExp(nome) }));
  }

  function nomesNaOrdem() {
    return within(suaOrdem())
      .getAllByRole("listitem")
      .map((item) => item.querySelector(".item-ordem-nome")?.firstChild?.textContent);
  }

  it("mostra a regra Borda, o orçamento e os custos dos projetos", async () => {
    renderizarVotar(consulta);

    expect(await screen.findByText(/a 1ª posição vale 3 pontos e a última vale 1/)).toBeInTheDocument();
    expect(screen.getByText(/Orçamento disponível: R\$ 60\.000,00/)).toBeInTheDocument();
    expect(within(listaAOrdenar()).getByRole("button", { name: /^Impressoras 3D e CNC R\$\s35\.000,00$/ })).toBeInTheDocument();
  });

  it("habilita o envio só quando todos os projetos estão ordenados", async () => {
    renderizarVotar(consulta);
    const enviar = await screen.findByRole("button", { name: "Enviar cédula" });

    expect(enviar).toBeDisabled();
    await adicionar("Servidor GPU");
    await adicionar("Impressoras");
    expect(enviar).toBeDisabled();
    expect(screen.getByText("2 de 3 projetos ordenados")).toBeInTheDocument();

    await adicionar("Kit Arduino");
    expect(enviar).toBeEnabled();
    expect(within(listaAOrdenar()).getByText("Todos os projetos foram ordenados.")).toBeInTheDocument();
  });

  it("subir, descer e remover reorganizam a ordem antes do envio", async () => {
    renderizarVotar(consulta);
    await screen.findByRole("button", { name: "Enviar cédula" });
    await adicionar("Impressoras");
    await adicionar("Kit Arduino");
    await adicionar("Servidor GPU");
    expect(nomesNaOrdem()).toEqual(["Impressoras 3D e CNC", "Kit Arduino e Robótica", "Servidor GPU para IA"]);

    expect(screen.getByRole("button", { name: "Subir Impressoras 3D e CNC" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Descer Servidor GPU para IA" })).toBeDisabled();

    expect(within(suaOrdem()).getAllByRole("listitem").map((item) => item.querySelector("small")?.textContent)).toEqual([
      "3 pontos",
      "2 pontos",
      "1 ponto",
    ]);

    await userEvent.click(screen.getByRole("button", { name: "Subir Servidor GPU para IA" }));
    expect(nomesNaOrdem()).toEqual(["Impressoras 3D e CNC", "Servidor GPU para IA", "Kit Arduino e Robótica"]);

    await userEvent.click(screen.getByRole("button", { name: "Descer Impressoras 3D e CNC" }));
    expect(nomesNaOrdem()).toEqual(["Servidor GPU para IA", "Impressoras 3D e CNC", "Kit Arduino e Robótica"]);

    await userEvent.click(screen.getByRole("button", { name: "Remover Impressoras 3D e CNC" }));
    expect(nomesNaOrdem()).toEqual(["Servidor GPU para IA", "Kit Arduino e Robótica"]);
    expect(screen.getByRole("button", { name: "Enviar cédula" })).toBeDisabled();
    expect(within(listaAOrdenar()).getByRole("button", { name: /Impressoras/ })).toBeInTheDocument();
  });

  it("envia os ids na ordem escolhida após confirmar", async () => {
    vi.mocked(api.votar).mockResolvedValue({ mensagem: "Voto registrado com sucesso." });
    renderizarVotar(consulta);
    await screen.findByRole("button", { name: "Enviar cédula" });

    await adicionar("Kit Arduino");
    await adicionar("Servidor GPU");
    await adicionar("Impressoras");
    await userEvent.click(screen.getByRole("button", { name: "Enviar cédula" }));

    const dialogo = screen.getByRole("dialog", { name: "Confirmar cédula" });
    expect(within(dialogo).getAllByRole("listitem").map((li) => li.textContent)).toEqual([
      "Kit Arduino e Robótica",
      "Servidor GPU para IA",
      "Impressoras 3D e CNC",
    ]);
    await userEvent.click(within(dialogo).getByRole("button", { name: "Enviar cédula" }));

    expect(api.votar).toHaveBeenCalledWith(7, { preferencias: [22, 23, 21] });
    expect(await screen.findByRole("heading", { name: "Voto registrado" })).toBeInTheDocument();
  });

  it("cancelar a confirmação não envia a cédula", async () => {
    renderizarVotar(consulta);
    await screen.findByRole("button", { name: "Enviar cédula" });
    for (const nome of ["Kit Arduino", "Servidor GPU", "Impressoras"]) await adicionar(nome);

    await userEvent.click(screen.getByRole("button", { name: "Enviar cédula" }));
    await confirmarNoDialogo("Cancelar");

    expect(api.votar).not.toHaveBeenCalled();
  });

  it("mostra o erro cedula_invalida vindo do servidor", async () => {
    vi.mocked(api.votar).mockRejectedValue(
      new ApiErro(400, "cedula_invalida", "A cédula precisa ordenar todos os projetos, sem repetição."),
    );
    renderizarVotar(consulta);
    await screen.findByRole("button", { name: "Enviar cédula" });
    for (const nome of ["Kit Arduino", "Servidor GPU", "Impressoras"]) await adicionar(nome);

    await userEvent.click(screen.getByRole("button", { name: "Enviar cédula" }));
    await confirmarNoDialogo("Enviar cédula");

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "A cédula precisa ordenar todos os projetos, sem repetição.",
    );
    expect(screen.getByRole("button", { name: "Enviar cédula" })).toBeEnabled();
  });
});
