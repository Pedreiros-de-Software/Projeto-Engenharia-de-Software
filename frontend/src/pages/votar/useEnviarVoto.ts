import { useState } from "react";

import { votar } from "../../api/client";
import { mensagemDeErro } from "../../api/erros";
import type { Voto } from "../../api/types";

/** Envia o voto e guarda o estado do envio; em sucesso repassa a mensagem do servidor. */
export function useEnviarVoto(votacaoId: number, onVotou: (mensagem: string) => void) {
  const [enviando, setEnviando] = useState(false);
  const [erro, setErro] = useState("");

  async function enviar(voto: Voto) {
    setEnviando(true);
    setErro("");
    try {
      const resposta = await votar(votacaoId, voto);
      onVotou(resposta.mensagem);
    } catch (e) {
      setErro(mensagemDeErro(e));
      setEnviando(false);
    }
  }

  return { enviando, erro, enviar };
}
