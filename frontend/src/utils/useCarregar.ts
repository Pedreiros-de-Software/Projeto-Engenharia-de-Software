import { useEffect, useState } from "react";

import { mensagemDeErro } from "../api/erros";

type Carregamento<T> = {
  dados: T | null;
  erro: string;
  carregando: boolean;
};

/** Executa `carregar` (estável, via `useCallback`) ao montar e quando ela muda; expõe dados, erro e carregamento. */
export function useCarregar<T>(carregar: () => Promise<T>) {
  const [estado, setEstado] = useState<Carregamento<T>>({
    dados: null,
    erro: "",
    carregando: true,
  });

  useEffect(() => {
    let ativo = true;
    carregar().then(
      (dados) => {
        if (ativo) setEstado({ dados, erro: "", carregando: false });
      },
      (erro: unknown) => {
        if (ativo) setEstado({ dados: null, erro: mensagemDeErro(erro), carregando: false });
      },
    );
    return () => {
      ativo = false;
    };
  }, [carregar]);

  function atualizarDados(dados: T) {
    setEstado({ dados, erro: "", carregando: false });
  }

  return { ...estado, atualizarDados };
}
