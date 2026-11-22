import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState, type ReactNode } from "react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";

import * as api from "./api/client";
import type { Usuario } from "./api/types";
import App, { Rotas } from "./App";
import { AuthContext, type Sessao } from "./auth/useAuth";
import { resultadoPauta } from "./test/fixtures";
import { ADMIN, criarSessao, ELEITORA } from "./test/renderizar";

vi.mock("./api/client");

function renderizarApp(rota: string, sessao: Sessao) {
  return render(
    <AuthContext.Provider value={sessao}>
      <MemoryRouter initialEntries={[rota]}>
        <Rotas />
      </MemoryRouter>
    </AuthContext.Provider>,
  );
}

function sessaoDe(usuario: Usuario | null) {
  return criarSessao(usuario);
}

function SessaoComEstado({ children }: { children: ReactNode }) {
  const [usuario, setUsuario] = useState<Usuario | null>(ELEITORA);
  const sessao = criarSessao(usuario, { sair: async () => setUsuario(null) });
  return <AuthContext.Provider value={sessao}>{children}</AuthContext.Provider>;
}

describe("rotas", () => {
  it("App monta o roteador e o provedor de sessão: sem token, abre o login", async () => {
    render(<App />);

    expect(await screen.findByRole("button", { name: "Entrar" })).toBeInTheDocument();
    expect(screen.getByLabelText("Usuário")).toBeInTheDocument();
  });

  it("sem login, o painel redireciona para a tela de login", async () => {
    renderizarApp("/", sessaoDe(null));

    expect(await screen.findByLabelText("Usuário")).toBeInTheDocument();
    expect(screen.getByLabelText("Senha")).toBeInTheDocument();
  });

  it("enquanto a sessão é verificada, mostra aviso em vez de redirecionar", () => {
    renderizarApp("/", criarSessao(null, { carregando: true }));

    expect(screen.getByText("Verificando sua sessão…")).toBeInTheDocument();
  });

  it("o resultado é público: funciona sem login e oferece o botão Entrar", async () => {
    vi.mocked(api.obterResultado).mockResolvedValue(resultadoPauta());

    renderizarApp("/votacoes/1/resultado", sessaoDe(null));

    expect(await screen.findByText("REPROVADA")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Entrar" })).toHaveAttribute("href", "/login");
    expect(api.obterVotacao).not.toHaveBeenCalled();
  });

  it("eleitor não vê os links de administração e é barrado nas páginas de admin", () => {
    vi.mocked(api.listarUnidades).mockReturnValue(new Promise(() => undefined));

    renderizarApp("/admin/unidades", sessaoDe(ELEITORA));

    expect(screen.getByText("Maria Souza")).toBeInTheDocument();
    expect(screen.getByText("Eleitor")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Painel" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Nova votação" })).not.toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Unidades" })).not.toBeInTheDocument();
    expect(screen.getByRole("alert")).toHaveTextContent("Acesso restrito");
    expect(api.listarUnidades).not.toHaveBeenCalled();
  });

  it("admin vê os links de administração e acessa a página de unidades", async () => {
    vi.mocked(api.listarUnidades).mockResolvedValue([]);

    renderizarApp("/admin/unidades", sessaoDe(ADMIN));

    expect(screen.getByRole("link", { name: "Nova votação" })).toHaveAttribute("href", "/admin/nova-votacao");
    expect(screen.getByRole("link", { name: "Unidades" })).toHaveAttribute("href", "/admin/unidades");
    expect(await screen.findByRole("heading", { level: 1, name: "Unidades" })).toBeInTheDocument();
  });

  it("Sair encerra a sessão e leva para o login", async () => {
    vi.mocked(api.listarVotacoes).mockResolvedValue([]);
    render(
      <SessaoComEstado>
        <MemoryRouter initialEntries={["/"]}>
          <Rotas />
        </MemoryRouter>
      </SessaoComEstado>,
    );

    await userEvent.click(screen.getByRole("button", { name: "Sair" }));

    expect(await screen.findByLabelText("Usuário")).toBeInTheDocument();
    expect(screen.queryByText("Maria Souza")).not.toBeInTheDocument();
  });

  it("endereço desconhecido mostra página não encontrada", () => {
    renderizarApp("/nao-existe", sessaoDe(ELEITORA));

    expect(screen.getByRole("heading", { name: "Página não encontrada" })).toBeInTheDocument();
  });
});
