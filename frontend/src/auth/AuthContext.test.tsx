import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import * as api from "../api/client";
import { ApiErro } from "../api/erros";
import { ELEITORA } from "../test/renderizar";
import { AuthProvider } from "./AuthContext";
import { useAuth } from "./useAuth";

vi.mock("../api/client");

function Consumidor() {
  const { usuario, carregando, entrar, sair } = useAuth();
  if (carregando) return <p>carregando</p>;
  return (
    <div>
      <p>{usuario ? `logado: ${usuario.nome}` : "anônimo"}</p>
      <button onClick={() => entrar("maria", "consensu").catch(() => undefined)}>entrar</button>
      <button onClick={() => sair()}>sair</button>
    </div>
  );
}

function renderizarProvider() {
  return render(
    <AuthProvider>
      <Consumidor />
    </AuthProvider>,
  );
}

describe("AuthProvider", () => {
  it("sem token salvo, não consulta /api/me e fica anônimo", () => {
    renderizarProvider();

    expect(screen.getByText("anônimo")).toBeInTheDocument();
    expect(api.obterSessao).not.toHaveBeenCalled();
  });

  it("com token salvo, restaura o usuário por /api/me", async () => {
    vi.mocked(api.obterToken).mockReturnValue("tok");
    vi.mocked(api.obterSessao).mockResolvedValue({ usuario: ELEITORA, unidade: null });

    renderizarProvider();

    expect(screen.getByText("carregando")).toBeInTheDocument();
    expect(await screen.findByText("logado: Maria Souza")).toBeInTheDocument();
  });

  it("token inválido no /api/me deixa o usuário anônimo", async () => {
    vi.mocked(api.obterToken).mockReturnValue("tok");
    vi.mocked(api.obterSessao).mockRejectedValue(new ApiErro(401, "nao_autenticado", "Sessão expirada."));

    renderizarProvider();

    expect(await screen.findByText("anônimo")).toBeInTheDocument();
  });

  it("entrar faz login e guarda o usuário; sair faz logout", async () => {
    vi.mocked(api.login).mockResolvedValue({ token: "t", usuario: ELEITORA });
    vi.mocked(api.logout).mockResolvedValue();
    renderizarProvider();

    await userEvent.click(screen.getByRole("button", { name: "entrar" }));
    expect(await screen.findByText("logado: Maria Souza")).toBeInTheDocument();
    expect(api.login).toHaveBeenCalledWith("maria", "consensu");

    await userEvent.click(screen.getByRole("button", { name: "sair" }));
    expect(await screen.findByText("anônimo")).toBeInTheDocument();
    expect(api.logout).toHaveBeenCalledOnce();
  });

  it("sair descarta a sessão local mesmo se o servidor falhar", async () => {
    vi.mocked(api.login).mockResolvedValue({ token: "t", usuario: ELEITORA });
    vi.mocked(api.logout).mockRejectedValue(new ApiErro(0, "sem_conexao", "Sem conexão."));
    renderizarProvider();

    await userEvent.click(screen.getByRole("button", { name: "entrar" }));
    await screen.findByText("logado: Maria Souza");
    await userEvent.click(screen.getByRole("button", { name: "sair" }));

    expect(await screen.findByText("anônimo")).toBeInTheDocument();
  });

  it("registra o aviso de sessão perdida, que volta o usuário para anônimo", async () => {
    vi.mocked(api.login).mockResolvedValue({ token: "t", usuario: ELEITORA });
    renderizarProvider();
    await userEvent.click(screen.getByRole("button", { name: "entrar" }));
    await screen.findByText("logado: Maria Souza");

    const aoPerderSessao = vi.mocked(api.definirAoPerderSessao).mock.calls[0][0];
    expect(aoPerderSessao).toBeTypeOf("function");
    await vi.waitFor(() => {
      aoPerderSessao?.();
      expect(screen.getByText("anônimo")).toBeInTheDocument();
    });
  });

  it("useAuth fora do provider lança erro explicativo", () => {
    vi.spyOn(console, "error").mockImplementation(() => undefined);
    expect(() => render(<Consumidor />)).toThrow("useAuth precisa estar dentro de <AuthProvider>");
  });
});
