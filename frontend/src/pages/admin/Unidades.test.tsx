import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import * as api from "../../api/client";
import { ApiErro } from "../../api/erros";
import type { Unidade } from "../../api/types";
import { ADMIN, renderizar } from "../../test/renderizar";
import Unidades from "./Unidades";

vi.mock("../../api/client");

const UNIDADES: Unidade[] = [
  { id: 1, identificador: "Apto 101", fracao_ideal: 60, percentual: 12, adimplente: true, responsavel: { id: 2, nome: "Maria Souza" } },
  { id: 6, identificador: "Apto 402", fracao_ideal: 120, percentual: 24, adimplente: false, responsavel: { id: 7, nome: "Sofia Prado" } },
  { id: 7, identificador: "Loja 1", fracao_ideal: 40, percentual: 8, adimplente: true, responsavel: null },
];

function renderizarUnidades() {
  vi.mocked(api.listarUnidades).mockResolvedValue(UNIDADES);
  renderizar(<Unidades />, { usuario: ADMIN, rota: "/admin/unidades", caminho: "/admin/unidades" });
}

function linha(identificador: string) {
  return screen.getByRole("rowheader", { name: identificador }).closest("tr") as HTMLElement;
}

describe("Admin: unidades", () => {
  it("lista identificador, responsável, fração, percentual e situação", async () => {
    renderizarUnidades();

    const tabela = await screen.findByRole("table", { name: "Unidades cadastradas" });
    const rolagem = screen.getByRole("region", { name: "Unidades cadastradas" });
    expect(rolagem).toContainElement(tabela);
    expect(rolagem).toHaveAttribute("tabindex", "0");
    const apto402 = within(linha("Apto 402"));
    expect(apto402.getByText("Sofia Prado")).toBeInTheDocument();
    expect(apto402.getByText("120")).toBeInTheDocument();
    expect(apto402.getByText("24,00%")).toBeInTheDocument();
    expect(apto402.getByText("Inadimplente")).toBeInTheDocument();
    expect(within(linha("Loja 1")).getByText("Sem responsável")).toBeInTheDocument();
  });

  it("marca a unidade inadimplente como adimplente e atualiza a linha", async () => {
    vi.mocked(api.alterarAdimplencia).mockResolvedValue({ ...UNIDADES[1], adimplente: true });
    renderizarUnidades();

    await userEvent.click(await screen.findByRole("button", { name: "Marcar Apto 402 como adimplente" }));

    expect(api.alterarAdimplencia).toHaveBeenCalledWith(6, true);
    expect(await screen.findByRole("status")).toHaveTextContent("Unidade Apto 402 marcada como adimplente.");
    expect(within(linha("Apto 402")).getByText("Adimplente")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Marcar Apto 402 como inadimplente" })).toBeEnabled();
  });

  it("marca uma unidade adimplente como inadimplente", async () => {
    vi.mocked(api.alterarAdimplencia).mockResolvedValue({ ...UNIDADES[0], adimplente: false });
    renderizarUnidades();

    await userEvent.click(await screen.findByRole("button", { name: "Marcar Apto 101 como inadimplente" }));

    expect(api.alterarAdimplencia).toHaveBeenCalledWith(1, false);
    expect(await screen.findByRole("status")).toHaveTextContent("Unidade Apto 101 marcada como inadimplente.");
  });

  it("mostra o erro do servidor ao alterar", async () => {
    vi.mocked(api.alterarAdimplencia).mockRejectedValue(new ApiErro(404, "nao_encontrado", "Unidade não encontrada."));
    renderizarUnidades();

    await userEvent.click(await screen.findByRole("button", { name: "Marcar Apto 101 como inadimplente" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Unidade não encontrada.");
  });

  it("mostra o erro quando a lista não carrega", async () => {
    vi.mocked(api.listarUnidades).mockRejectedValue(new ApiErro(403, "acesso_negado", "Acesso negado."));
    renderizar(<Unidades />, { usuario: ADMIN, rota: "/admin/unidades", caminho: "/admin/unidades" });

    expect(await screen.findByRole("alert")).toHaveTextContent("Acesso negado.");
  });
});
