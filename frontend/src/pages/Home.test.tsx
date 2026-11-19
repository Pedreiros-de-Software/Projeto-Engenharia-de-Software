import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import * as api from "../api/client";
import { ApiErro } from "../api/erros";
import { votacaoResumo } from "../test/fixtures";
import { ADMIN, renderizar } from "../test/renderizar";
import Home from "./Home";

vi.mock("../api/client");

const VOTACOES = [
  votacaoResumo({ id: 1, tipo: "pauta", titulo: "Pauta aberta", status: "aberta" }),
  votacaoResumo({ id: 2, tipo: "pauta", titulo: "Pauta encerrada", status: "encerrada", ja_votou: true }),
  votacaoResumo({ id: 3, tipo: "eleicao", titulo: "Eleição DCE", status: "encerrada", apto: false }),
  votacaoResumo({ id: 4, tipo: "eleicao", titulo: "Eleição DCE — 2º turno", status: "nao_iniciada", turno: 2 }),
  votacaoResumo({ id: 5, tipo: "eleicao", titulo: "Eleição já votada", status: "aberta", ja_votou: true }),
];

function cartao(titulo: string) {
  return screen.getByRole("article", { name: titulo });
}

describe("Home (painel)", () => {
  it("identifica o produto para quem acessa a página inicial", () => {
    vi.mocked(api.listarVotacoes).mockResolvedValue([]);
    renderizar(<Home />);
    expect(screen.getByRole("heading", { level: 1, name: "Consensu" })).toBeVisible();
  });

  it("informa os três contextos de decisão atendidos pelo produto", () => {
    vi.mocked(api.listarVotacoes).mockResolvedValue([]);
    renderizar(<Home />);
    const descricao = screen.getByRole("heading", { level: 3, name: /Sistema de votação/ });
    expect(descricao).toHaveTextContent("Eleições de CA");
    expect(descricao).toHaveTextContent("Condomínio");
    expect(descricao).toHaveTextContent("Priorização de projetos");
    expect(descricao).not.toHaveTextContent("Assembléia");
  });

  it("agrupa as votações por contexto", async () => {
    vi.mocked(api.listarVotacoes).mockResolvedValue(VOTACOES);
    renderizar(<Home />);

    const condominio = await screen.findByRole("region", { name: "Assembleia de condomínio" });
    const eleicoes = screen.getByRole("region", { name: "Eleição do CA" });
    const projetos = screen.getByRole("region", { name: "Priorização de projetos" });

    expect(within(condominio).getAllByRole("article").map((a) => a.getAttribute("aria-label"))).toEqual([
      "Pauta aberta",
      "Pauta encerrada",
    ]);
    expect(within(eleicoes).getAllByRole("article")).toHaveLength(3);
    expect(within(projetos).getByText("Nenhuma votação neste contexto.")).toBeInTheDocument();
  });

  it("mostra status, turno e os botões permitidos para cada votação", async () => {
    vi.mocked(api.listarVotacoes).mockResolvedValue(VOTACOES);
    renderizar(<Home />);
    await screen.findByRole("region", { name: "Eleição do CA" });

    const aberta = cartao("Pauta aberta");
    expect(within(aberta).getByText("Aberta")).toBeInTheDocument();
    expect(within(aberta).getByRole("link", { name: "Votar" })).toHaveAttribute("href", "/votacoes/1");
    expect(within(aberta).queryByRole("link", { name: "Ver resultado" })).not.toBeInTheDocument();

    const encerrada = cartao("Pauta encerrada");
    expect(within(encerrada).getByText("Encerrada")).toBeInTheDocument();
    expect(within(encerrada).getByText("Você já votou")).toBeInTheDocument();
    expect(within(encerrada).getByRole("link", { name: "Ver resultado" })).toHaveAttribute(
      "href",
      "/votacoes/2/resultado",
    );
    expect(within(encerrada).queryByRole("link", { name: "Votar" })).not.toBeInTheDocument();

    expect(within(cartao("Eleição DCE")).getByText("Não apto")).toBeInTheDocument();
    expect(within(cartao("Eleição DCE")).getByText("1º turno")).toBeInTheDocument();

    const naoIniciada = cartao("Eleição DCE — 2º turno");
    expect(within(naoIniciada).getByText("Não iniciada")).toBeInTheDocument();
    expect(within(naoIniciada).getByText("2º turno")).toBeInTheDocument();
    expect(within(naoIniciada).queryByRole("link")).not.toBeInTheDocument();

    const jaVotada = cartao("Eleição já votada");
    expect(within(jaVotada).getByText("Você já votou")).toBeInTheDocument();
    expect(within(jaVotada).queryByRole("link", { name: "Votar" })).not.toBeInTheDocument();

    expect(screen.queryByRole("button", { name: "Encerrar agora" })).not.toBeInTheDocument();
  });

  it("mostra a janela de votação em data e hora locais", async () => {
    vi.mocked(api.listarVotacoes).mockResolvedValue([VOTACOES[0]]);
    renderizar(<Home />);

    expect(await screen.findByText(/De 18\/11\/2026 às 10:00 até 25\/11\/2026 às 10:00/)).toBeInTheDocument();
  });

  it("admin encerra uma votação aberta depois de confirmar", async () => {
    vi.mocked(api.listarVotacoes).mockResolvedValue(VOTACOES);
    vi.mocked(api.encerrarVotacao).mockResolvedValue({ ...VOTACOES[0], status: "encerrada" });
    renderizar(<Home />, { usuario: ADMIN });
    await screen.findByRole("region", { name: "Eleição do CA" });

    expect(within(cartao("Pauta encerrada")).queryByRole("button", { name: "Encerrar agora" })).not.toBeInTheDocument();
    await userEvent.click(within(cartao("Pauta aberta")).getByRole("button", { name: "Encerrar agora" }));

    const dialogo = screen.getByRole("dialog", { name: "Encerrar votação agora?" });
    expect(dialogo).toHaveTextContent("Pauta aberta");
    expect(api.encerrarVotacao).not.toHaveBeenCalled();
    await userEvent.click(within(dialogo).getByRole("button", { name: "Encerrar votação" }));

    expect(api.encerrarVotacao).toHaveBeenCalledWith(1);
    expect(await screen.findByRole("status")).toHaveTextContent('Votação "Pauta aberta" encerrada.');
    expect(within(cartao("Pauta aberta")).getByText("Encerrada")).toBeInTheDocument();
    expect(within(cartao("Pauta aberta")).getByRole("link", { name: "Ver resultado" })).toBeInTheDocument();
  });

  it("cancelar o diálogo não encerra a votação", async () => {
    vi.mocked(api.listarVotacoes).mockResolvedValue(VOTACOES);
    renderizar(<Home />, { usuario: ADMIN });
    await screen.findByRole("region", { name: "Eleição do CA" });

    await userEvent.click(within(cartao("Pauta aberta")).getByRole("button", { name: "Encerrar agora" }));
    await userEvent.click(screen.getByRole("button", { name: "Cancelar" }));

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(api.encerrarVotacao).not.toHaveBeenCalled();
  });

  it("mostra o erro do servidor ao encerrar", async () => {
    vi.mocked(api.listarVotacoes).mockResolvedValue(VOTACOES);
    vi.mocked(api.encerrarVotacao).mockRejectedValue(
      new ApiErro(409, "votacao_nao_aberta", "A votação não está aberta."),
    );
    renderizar(<Home />, { usuario: ADMIN });
    await screen.findByRole("region", { name: "Eleição do CA" });

    await userEvent.click(within(cartao("Pauta aberta")).getByRole("button", { name: "Encerrar agora" }));
    await userEvent.click(screen.getByRole("button", { name: "Encerrar votação" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("A votação não está aberta.");
  });

  it("mostra o erro quando a lista não carrega", async () => {
    vi.mocked(api.listarVotacoes).mockRejectedValue(new ApiErro(500, "erro_interno", "Erro interno do servidor."));
    renderizar(<Home />);

    expect(await screen.findByRole("alert")).toHaveTextContent("Erro interno do servidor.");
  });

  it("mostra a mensagem recebida de outra página (ex.: votação criada)", async () => {
    vi.mocked(api.listarVotacoes).mockResolvedValue([]);
    renderizar(<Home />, { estado: { mensagem: 'Votação "Nova pauta" criada com sucesso.' } });

    expect(await screen.findByRole("status")).toHaveTextContent('Votação "Nova pauta" criada com sucesso.');
  });
});
