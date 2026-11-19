import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import * as api from "../../api/client";
import { ApiErro } from "../../api/erros";
import type { Usuario } from "../../api/types";
import { votacaoDetalhe } from "../../test/fixtures";
import { ADMIN, renderizar } from "../../test/renderizar";
import NovaVotacao from "./NovaVotacao";

vi.mock("../../api/client");

const USUARIOS: Usuario[] = [
  ADMIN,
  { id: 10, nome: "Ana Beatriz", login: "ana", papel: "eleitor" },
  { id: 11, nome: "Bruno Costa", login: "bruno", papel: "eleitor" },
  { id: 12, nome: "Maria Souza", login: "maria", papel: "eleitor" },
];

function renderizarFormulario() {
  vi.mocked(api.listarUsuarios).mockResolvedValue(USUARIOS);
  renderizar(<NovaVotacao />, { usuario: ADMIN, rota: "/admin/nova-votacao", caminho: "/admin/nova-votacao" });
}

async function preencherBasico(tipo: string, titulo: string) {
  await userEvent.selectOptions(screen.getByLabelText("Tipo de votação"), tipo);
  await userEvent.type(screen.getByLabelText("Título"), titulo);
  await definirData("Início", "2026-12-01T09:00");
  await definirData("Fim", "2026-12-08T18:00");
}

async function definirData(rotulo: string, valor: string) {
  const campo = screen.getByLabelText(rotulo);
  await userEvent.clear(campo);
  await userEvent.type(campo, valor);
}

async function preencherOpcao(grupo: string, nome: string, descricao: string, custo?: string) {
  const fieldset = screen.getByRole("group", { name: grupo });
  await userEvent.type(within(fieldset).getByLabelText("Nome"), nome);
  if (descricao) {
    await userEvent.type(within(fieldset).getByLabelText("Descrição"), descricao);
  }
  if (custo !== undefined) {
    await userEvent.type(within(fieldset).getByLabelText("Custo (R$)"), custo);
  }
}

describe("Admin: nova votação", () => {
  it("cria uma pauta sem opções nem aptos e volta ao painel com mensagem", async () => {
    vi.mocked(api.criarVotacao).mockResolvedValue(votacaoDetalhe({ titulo: "Pauta #03" }));
    renderizarFormulario();

    await preencherBasico("pauta", "Pauta #03");
    await userEvent.type(screen.getByLabelText("Descrição"), "Pintura da fachada");
    expect(screen.queryByRole("group", { name: "Eleitores aptos" })).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Criar votação" }));

    expect(api.criarVotacao).toHaveBeenCalledWith({
      tipo: "pauta",
      titulo: "Pauta #03",
      descricao: "Pintura da fachada",
      inicio: "2026-12-01T12:00:00.000Z",
      fim: "2026-12-08T21:00:00.000Z",
      opcoes: [],
    });
    expect(await screen.findByTestId("endereco")).toHaveTextContent("/");
  });

  it("cria uma eleição com chapas e só os eleitores marcados como aptos", async () => {
    vi.mocked(api.criarVotacao).mockResolvedValue(votacaoDetalhe({ tipo: "eleicao" }));
    renderizarFormulario();

    await preencherBasico("eleicao", "Eleição DCE");
    await preencherOpcao("Chapa 1", "Chapa 1 — Movimenta", "Ana / Victor");
    await preencherOpcao("Chapa 2", "Chapa 2 — Renovação", "João / Ana");
    const aptos = screen.getByRole("group", { name: "Eleitores aptos" });
    expect(await within(aptos).findByText("3 de 3 selecionados")).toBeInTheDocument();
    expect(within(aptos).queryByLabelText(/Administração/)).not.toBeInTheDocument();
    await userEvent.click(within(aptos).getByLabelText(/Maria Souza/));
    expect(within(aptos).getByText("2 de 3 selecionados")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Criar votação" }));

    expect(api.criarVotacao).toHaveBeenCalledWith(
      expect.objectContaining({
        tipo: "eleicao",
        opcoes: [
          { nome: "Chapa 1 — Movimenta", descricao: "Ana / Victor" },
          { nome: "Chapa 2 — Renovação", descricao: "João / Ana" },
        ],
        aptos: [10, 11],
      }),
    );
  });

  it("cria uma consulta convertendo custos e orçamento de reais para centavos", async () => {
    vi.mocked(api.criarVotacao).mockResolvedValue(votacaoDetalhe({ tipo: "consulta" }));
    renderizarFormulario();

    await preencherBasico("consulta", "Orçamento participativo");
    await preencherOpcao("Projeto 1", "Impressoras 3D", "Laboratório", "35000");
    await preencherOpcao("Projeto 2", "Kit Arduino", "Robótica", "12000.50");
    await userEvent.click(screen.getByRole("button", { name: "Adicionar projeto" }));
    await preencherOpcao("Projeto 3", "Mobiliário", "", "9000");
    await userEvent.type(screen.getByLabelText("Orçamento disponível (R$)"), "60000");
    await userEvent.click(screen.getByRole("button", { name: "Criar votação" }));

    expect(api.criarVotacao).toHaveBeenCalledWith(
      expect.objectContaining({
        tipo: "consulta",
        opcoes: [
          { nome: "Impressoras 3D", descricao: "Laboratório", custo_centavos: 3_500_000 },
          { nome: "Kit Arduino", descricao: "Robótica", custo_centavos: 1_200_050 },
          { nome: "Mobiliário", descricao: "", custo_centavos: 900_000 },
        ],
        orcamento_centavos: 6_000_000,
        aptos: [10, 11, 12],
      }),
    );
  });

  it("aceita reais no formato brasileiro e mostra o valor entendido antes de enviar", async () => {
    vi.mocked(api.criarVotacao).mockResolvedValue(votacaoDetalhe({ tipo: "consulta" }));
    renderizarFormulario();

    await preencherBasico("consulta", "Orçamento participativo");
    await preencherOpcao("Projeto 1", "Impressoras 3D", "", "35.000");
    await preencherOpcao("Projeto 2", "Kit Arduino", "", "12000,50");
    await userEvent.type(screen.getByLabelText("Orçamento disponível (R$)"), "60.000,00");

    const projeto1 = screen.getByRole("group", { name: "Projeto 1" });
    expect(within(projeto1).getByText("Custo: R$ 35.000,00")).toBeInTheDocument();
    expect(within(projeto1).getByLabelText("Custo (R$)")).toHaveAccessibleDescription(/^Custo: R\$\s35\.000,00$/);
    expect(within(screen.getByRole("group", { name: "Projeto 2" })).getByText("Custo: R$ 12.000,50")).toBeInTheDocument();
    expect(screen.getByText("Orçamento: R$ 60.000,00")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Criar votação" }));

    expect(api.criarVotacao).toHaveBeenCalledWith(
      expect.objectContaining({
        opcoes: [
          { nome: "Impressoras 3D", descricao: "", custo_centavos: 3_500_000 },
          { nome: "Kit Arduino", descricao: "", custo_centavos: 1_200_050 },
        ],
        orcamento_centavos: 6_000_000,
      }),
    );
  });

  it("valor em reais fora do formato é sinalizado no campo e bloqueia o envio", async () => {
    renderizarFormulario();

    await preencherBasico("consulta", "Orçamento participativo");
    await preencherOpcao("Projeto 1", "Impressoras 3D", "", "1.5000");
    await preencherOpcao("Projeto 2", "Kit Arduino", "", "12000");
    await userEvent.type(screen.getByLabelText("Orçamento disponível (R$)"), "60000");

    expect(
      within(screen.getByRole("group", { name: "Projeto 1" })).getByText(/^Custo: formato não reconhecido/),
    ).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Criar votação" }));

    expect(screen.getByRole("alert")).toHaveTextContent('Informe o custo em reais do projeto "Impressoras 3D"');
    expect(api.criarVotacao).not.toHaveBeenCalled();
  });

  it("adiciona e remove opções, mantendo no mínimo duas", async () => {
    renderizarFormulario();
    await userEvent.selectOptions(screen.getByLabelText("Tipo de votação"), "eleicao");

    expect(screen.getByRole("button", { name: "Remover Chapa 1" })).toBeDisabled();
    await userEvent.click(screen.getByRole("button", { name: "Adicionar chapa" }));
    expect(screen.getByRole("group", { name: "Chapa 3" })).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Remover Chapa 1" }));
    expect(screen.queryByRole("group", { name: "Chapa 3" })).not.toBeInTheDocument();
    expect(screen.getAllByRole("group", { name: /^Chapa/ })).toHaveLength(2);
  });

  it("marcar e desmarcar todos os eleitores aptos", async () => {
    renderizarFormulario();
    await userEvent.selectOptions(screen.getByLabelText("Tipo de votação"), "eleicao");
    const aptos = screen.getByRole("group", { name: "Eleitores aptos" });
    await within(aptos).findByText("3 de 3 selecionados");

    await userEvent.click(within(aptos).getByRole("button", { name: "Desmarcar todos" }));
    expect(within(aptos).getByText("0 de 3 selecionados")).toBeInTheDocument();
    expect(within(aptos).getByLabelText(/Ana Beatriz/)).not.toBeChecked();

    await userEvent.click(within(aptos).getByRole("button", { name: "Marcar todos" }));
    expect(within(aptos).getByText("3 de 3 selecionados")).toBeInTheDocument();
  });

  it("valida no navegador antes de enviar (fim antes do início)", async () => {
    renderizarFormulario();

    await userEvent.type(screen.getByLabelText("Título"), "Pauta");
    await definirData("Início", "2026-12-08T09:00");
    await definirData("Fim", "2026-12-01T09:00");
    await userEvent.click(screen.getByRole("button", { name: "Criar votação" }));

    expect(screen.getByRole("alert")).toHaveTextContent("O fim da votação precisa ser depois do início.");
    expect(api.criarVotacao).not.toHaveBeenCalled();
  });

  it("mostra o erro do servidor e permite tentar de novo", async () => {
    vi.mocked(api.criarVotacao).mockRejectedValue(
      new ApiErro(400, "requisicao_invalida", "Título obrigatório."),
    );
    renderizarFormulario();

    await preencherBasico("pauta", "Pauta");
    await userEvent.click(screen.getByRole("button", { name: "Criar votação" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Título obrigatório.");
    expect(screen.getByRole("button", { name: "Criar votação" })).toBeEnabled();
  });

  it("mostra o erro quando a lista de eleitores não carrega", async () => {
    vi.mocked(api.listarUsuarios).mockRejectedValue(new ApiErro(403, "acesso_negado", "Acesso negado."));
    renderizar(<NovaVotacao />, { usuario: ADMIN, rota: "/admin/nova-votacao", caminho: "/admin/nova-votacao" });
    await userEvent.selectOptions(screen.getByLabelText("Tipo de votação"), "consulta");

    expect(await screen.findByRole("alert")).toHaveTextContent("Acesso negado.");
  });

  it("Cancelar volta ao painel sem criar", async () => {
    renderizarFormulario();

    await userEvent.click(screen.getByRole("button", { name: "Cancelar" }));

    expect(screen.getByTestId("endereco")).toHaveTextContent("/");
    expect(api.criarVotacao).not.toHaveBeenCalled();
  });
});
