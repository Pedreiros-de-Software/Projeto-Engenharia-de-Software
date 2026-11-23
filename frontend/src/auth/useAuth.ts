import { createContext, useContext } from "react";

import type { Usuario } from "../api/types";

export type Sessao = {
  usuario: Usuario | null;
  carregando: boolean;
  entrar: (login: string, senha: string) => Promise<void>;
  sair: () => Promise<void>;
};

export const AuthContext = createContext<Sessao | null>(null);

export function useAuth(): Sessao {
  const sessao = useContext(AuthContext);
  if (!sessao) {
    throw new Error("useAuth precisa estar dentro de <AuthProvider>");
  }
  return sessao;
}
