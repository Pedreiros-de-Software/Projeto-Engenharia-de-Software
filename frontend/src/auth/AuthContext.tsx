import { useEffect, useState, type ReactNode } from "react";

import { definirAoPerderSessao, login, logout, obterSessao, obterToken } from "../api/client";
import type { Usuario } from "../api/types";
import { AuthContext, type Sessao } from "./useAuth";

/** Guarda o usuário logado: restaura a sessão pelo token salvo (/api/me), entra e sai. */
export function AuthProvider({ children }: { children: ReactNode }) {
  const [usuario, setUsuario] = useState<Usuario | null>(null);
  const [carregando, setCarregando] = useState(() => Boolean(obterToken()));

  useEffect(() => {
    definirAoPerderSessao(() => setUsuario(null));
    return () => definirAoPerderSessao(null);
  }, []);

  useEffect(() => {
    if (!obterToken()) {
      return;
    }
    obterSessao()
      .then((resposta) => setUsuario(resposta.usuario))
      .catch(() => setUsuario(null))
      .finally(() => setCarregando(false));
  }, []);

  const sessao: Sessao = {
    usuario,
    carregando,
    entrar: async (nomeDeUsuario, senha) => {
      const resposta = await login(nomeDeUsuario, senha);
      setUsuario(resposta.usuario);
    },
    sair: async () => {
      try {
        await logout();
      } catch {
        // a sessão local é descartada mesmo se o servidor não responder
      }
      setUsuario(null);
    },
  };

  return <AuthContext.Provider value={sessao}>{children}</AuthContext.Provider>;
}
