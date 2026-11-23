import { render } from "@testing-library/react";
import type { ReactElement } from "react";
import { MemoryRouter, Route, Routes, useLocation } from "react-router-dom";
import { vi } from "vitest";

import type { Usuario } from "../api/types";
import { AuthContext, type Sessao } from "../auth/useAuth";

export const ADMIN: Usuario = { id: 1, nome: "Administração", login: "admin", papel: "admin" };
export const ELEITORA: Usuario = { id: 2, nome: "Maria Souza", login: "maria", papel: "eleitor" };

export function criarSessao(usuario: Usuario | null, extra: Partial<Sessao> = {}): Sessao {
  return { usuario, carregando: false, entrar: vi.fn(), sair: vi.fn(), ...extra };
}

type Opcoes = {
  usuario?: Usuario | null;
  sessao?: Sessao;
  rota?: string;
  caminho?: string;
  estado?: unknown;
};

/** Renderiza `elemento` em `caminho` com uma sessão falsa; outras rotas mostram o endereço atual. */
export function renderizar(
  elemento: ReactElement,
  { usuario = ELEITORA, sessao, rota = "/", caminho = "/", estado }: Opcoes = {},
) {
  return render(
    <AuthContext.Provider value={sessao ?? criarSessao(usuario)}>
      <MemoryRouter initialEntries={[{ pathname: rota, state: estado }]}>
        <Routes>
          <Route path={caminho} element={elemento} />
          <Route path="*" element={<EnderecoAtual />} />
        </Routes>
      </MemoryRouter>
    </AuthContext.Provider>,
  );
}

function EnderecoAtual() {
  return <p data-testid="endereco">{useLocation().pathname}</p>;
}
