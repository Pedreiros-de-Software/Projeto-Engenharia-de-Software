import { useState } from "react";

import type { VotacaoDetalhe, VotoEleicao } from "../../api/types";
import { Alerta, Carregando } from "../../components/Alerta";
import { DialogoConfirmacao } from "../../components/DialogoConfirmacao";
import { useEnviarVoto } from "./useEnviarVoto";

type Escolha = { voto: VotoEleicao; rotulo: string };

type Props = {
  votacao: VotacaoDetalhe;
  onVotou: (mensagem: string) => void;
};

/** Cédula secreta da eleição: uma chapa, branco ou nulo. */
export function VotarEleicao({ votacao, onVotou }: Props) {
  const [escolha, setEscolha] = useState<Escolha | null>(null);
  const { enviando, erro, enviar } = useEnviarVoto(votacao.id, onVotou);

  function confirmar(voto: VotoEleicao) {
    setEscolha(null);
    void enviar(voto);
  }

  return (
    <section className="cartao" aria-labelledby="titulo-cedula">
      <div className="cedula-topo">
        <h2 id="titulo-cedula">Cédula</h2>
        <span className="badge badge-secreto">Voto secreto</span>
      </div>
      <p className="texto-secundario">
        Sua escolha é gravada sem nenhum vínculo com seu nome e não é exibida depois de
        confirmada.
      </p>

      <ul className="grade-chapas">
        {votacao.opcoes.map((chapa) => (
          <li key={chapa.id}>
            <button
              type="button"
              className="cartao-chapa"
              disabled={enviando}
              onClick={() =>
                setEscolha({
                  voto: { escolha: "chapa", opcao_id: chapa.id },
                  rotulo: chapa.nome,
                })
              }
            >
              <span className="cartao-chapa-nome">{chapa.nome}</span>{" "}
              {chapa.descricao && <span className="cartao-chapa-descricao">{chapa.descricao}</span>}
            </button>
          </li>
        ))}
      </ul>

      <div className="acoes-voto">
        <button
          type="button"
          className="botao botao-secundario"
          disabled={enviando}
          onClick={() => setEscolha({ voto: { escolha: "branco" }, rotulo: "Em branco" })}
        >
          Votar em branco
        </button>
        <button
          type="button"
          className="botao botao-secundario"
          disabled={enviando}
          onClick={() => setEscolha({ voto: { escolha: "nulo" }, rotulo: "Nulo" })}
        >
          Votar nulo
        </button>
      </div>

      {enviando && <Carregando texto="Registrando voto…" />}
      {erro && <Alerta tipo="erro">{erro}</Alerta>}

      {escolha && (
        <DialogoConfirmacao
          titulo="Confirmar voto"
          textoConfirmar="Confirmar voto"
          onConfirmar={() => confirmar(escolha.voto)}
          onCancelar={() => setEscolha(null)}
        >
          <p>
            Seu voto: <strong>{escolha.rotulo}</strong>
          </p>
          <p>O voto é secreto: depois de confirmado, não pode ser consultado nem alterado.</p>
        </DialogoConfirmacao>
      )}
    </section>
  );
}
