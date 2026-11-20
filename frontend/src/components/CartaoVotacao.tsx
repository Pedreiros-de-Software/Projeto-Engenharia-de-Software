import { Link } from "react-router-dom";

import type { VotacaoResumo } from "../api/types";
import { formatarDataHora } from "../utils/formatacao";
import { BadgeStatus } from "./BadgeStatus";

type Props = {
  votacao: VotacaoResumo;
  admin: boolean;
  onEncerrar: (votacao: VotacaoResumo) => void;
};

export function CartaoVotacao({ votacao, admin, onEncerrar }: Props) {
  const podeVotar = votacao.status === "aberta" && votacao.apto && !votacao.ja_votou;

  return (
    <article className="cartao cartao-votacao" aria-label={votacao.titulo}>
      <div className="cartao-votacao-topo">
        <h3>{votacao.titulo}</h3>
        <BadgeStatus status={votacao.status} />
      </div>
      {votacao.descricao && <p className="cartao-votacao-descricao">{votacao.descricao}</p>}
      <p className="texto-secundario">
        {votacao.tipo === "eleicao" && <span className="etiqueta">{votacao.turno}º turno</span>}
        De {formatarDataHora(votacao.inicio)} até {formatarDataHora(votacao.fim)}
      </p>
      <div className="cartao-votacao-rodape">
        <div className="etiquetas">
          {votacao.ja_votou && <span className="etiqueta etiqueta-sucesso">Você já votou</span>}
          {!votacao.apto && !admin && <span className="etiqueta etiqueta-neutra">Não apto</span>}
        </div>
        <div className="acoes">
          {podeVotar && (
            <Link to={`/votacoes/${votacao.id}`} className="botao">
              Votar
            </Link>
          )}
          {votacao.status === "encerrada" && (
            <Link to={`/votacoes/${votacao.id}/resultado`} className="botao botao-secundario">
              Ver resultado
            </Link>
          )}
          {admin && votacao.status === "aberta" && (
            <button type="button" className="botao botao-perigo" onClick={() => onEncerrar(votacao)}>
              Encerrar agora
            </button>
          )}
        </div>
      </div>
    </article>
  );
}
