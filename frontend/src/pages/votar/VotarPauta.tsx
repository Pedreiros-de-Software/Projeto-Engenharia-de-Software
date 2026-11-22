import { useState } from "react";

import type { VotacaoDetalhe, VotoPauta } from "../../api/types";
import { Alerta, Carregando } from "../../components/Alerta";
import { DialogoConfirmacao } from "../../components/DialogoConfirmacao";
import { formatarNumero, formatarPercentual } from "../../utils/formatacao";
import { useEnviarVoto } from "./useEnviarVoto";

type OpcaoPauta = { escolha: VotoPauta["escolha"]; rotulo: string; classe: string };

const OPCOES: OpcaoPauta[] = [
  { escolha: "favor", rotulo: "Aprovar", classe: "botao-sucesso" },
  { escolha: "contra", rotulo: "Rejeitar", classe: "botao-perigo" },
  { escolha: "abstencao", rotulo: "Abster-se", classe: "botao-secundario" },
];

type Props = {
  votacao: VotacaoDetalhe;
  onVotou: (mensagem: string) => void;
};

/** Voto da unidade na pauta: mostra o peso antes de confirmar e bloqueia unidade inelegível. */
export function VotarPauta({ votacao, onVotou }: Props) {
  const unidade = votacao.minha_unidade;
  const [escolhida, setEscolhida] = useState<OpcaoPauta | null>(null);
  const { enviando, erro, enviar } = useEnviarVoto(votacao.id, onVotou);

  if (!unidade) {
    return (
      <Alerta tipo="aviso" titulo="Sem unidade vinculada">
        Somente responsáveis por uma unidade do condomínio podem votar nesta pauta.
      </Alerta>
    );
  }

  function confirmar(opcao: OpcaoPauta) {
    setEscolhida(null);
    void enviar({ escolha: opcao.escolha });
  }

  return (
    <section className="cartao" aria-labelledby="titulo-unidade">
      <h2 id="titulo-unidade">Sua unidade</h2>
      <dl className="dados">
        <div>
          <dt>Unidade</dt>
          <dd>{unidade.identificador}</dd>
        </div>
        <div>
          <dt>Fração ideal</dt>
          <dd>{formatarNumero(unidade.fracao_ideal)}</dd>
        </div>
        <div>
          <dt>Peso do seu voto</dt>
          <dd>{formatarPercentual(unidade.percentual)} do total</dd>
        </div>
        <div>
          <dt>Situação</dt>
          <dd>{unidade.adimplente ? "Adimplente" : "Inadimplente"}</dd>
        </div>
      </dl>

      {unidade.elegivel ? (
        <>
          <p className="texto-secundario">
            O voto é contado com o peso da fração ideal da unidade. A pauta é aprovada com pelo
            menos 2/3 do peso total cadastrado.
          </p>
          <div className="acoes-voto">
            {OPCOES.map((opcao) => (
              <button
                key={opcao.escolha}
                type="button"
                className={`botao botao-grande ${opcao.classe}`}
                onClick={() => setEscolhida(opcao)}
                disabled={enviando}
              >
                {opcao.rotulo}
              </button>
            ))}
          </div>
        </>
      ) : (
        <Alerta tipo="erro" titulo="Voto bloqueado">
          {unidade.motivo || "A unidade não está apta a votar."}
        </Alerta>
      )}

      {enviando && <Carregando texto="Registrando voto…" />}
      {erro && <Alerta tipo="erro">{erro}</Alerta>}

      {escolhida && (
        <DialogoConfirmacao
          titulo="Confirmar voto"
          textoConfirmar={`Confirmar: ${escolhida.rotulo}`}
          onConfirmar={() => confirmar(escolhida)}
          onCancelar={() => setEscolhida(null)}
        >
          <p>
            Seu voto: <strong>{escolhida.rotulo.toUpperCase()}</strong> pela unidade{" "}
            {unidade.identificador}, com peso de {formatarPercentual(unidade.percentual)} do total
            (fração ideal {formatarNumero(unidade.fracao_ideal)}).
          </p>
          <p>Depois de confirmado, o voto não pode ser alterado.</p>
        </DialogoConfirmacao>
      )}
    </section>
  );
}
