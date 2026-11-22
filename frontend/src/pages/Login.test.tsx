import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { ApiErro } from "../api/erros";
import { criarSessao, ELEITORA, renderizar } from "../test/renderizar";
import Login from "./Login";

vi.mock("../api/client");

function renderizarLogin(sessao = criarSessao(null)) {
  renderizar(<Login />, { sessao, rota: "/login", caminho: "/login" });
  return sessao;
}

describe("Login", () => {
  it("envia usuário e senha e vai para o painel", async () => {
    const sessao = renderizarLogin();
    vi.mocked(sessao.entrar).mockResolvedValue();

    await userEvent.type(screen.getByLabelText("Usuário"), " maria ");
    await userEvent.type(screen.getByLabelText("Senha"), "consensu");
    await userEvent.click(screen.getByRole("button", { name: "Entrar" }));

    expect(sessao.entrar).toHaveBeenCalledWith("maria", "consensu");
    expect(await screen.findByTestId("endereco")).toHaveTextContent("/");
  });

  it("mostra a mensagem do servidor quando as credenciais são inválidas", async () => {
    const sessao = renderizarLogin();
    vi.mocked(sessao.entrar).mockRejectedValue(
      new ApiErro(401, "credenciais_invalidas", "Usuário ou senha inválidos."),
    );

    await userEvent.type(screen.getByLabelText("Usuário"), "maria");
    await userEvent.type(screen.getByLabelText("Senha"), "errada");
    await userEvent.click(screen.getByRole("button", { name: "Entrar" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Usuário ou senha inválidos.");
    expect(screen.getByRole("button", { name: "Entrar" })).toBeEnabled();
  });

  it("usuário já logado é levado direto ao painel", () => {
    renderizarLogin(criarSessao(ELEITORA));

    expect(screen.getByTestId("endereco")).toHaveTextContent("/");
    expect(screen.queryByLabelText("Senha")).not.toBeInTheDocument();
  });

  it("a senha usa campo do tipo password", () => {
    renderizarLogin();

    expect(screen.getByLabelText("Senha")).toHaveAttribute("type", "password");
  });
});
