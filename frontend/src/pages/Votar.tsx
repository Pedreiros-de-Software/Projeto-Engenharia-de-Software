import { useCallback, useState } from "react";
import { Link, useParams } from "react-router-dom";

import { obterVotacao } from "../api/client";
import type { VotacaoDetalhe } from "../api/types";
import { useAuth } from "../auth/useAuth";
import { Alerta, Carregando } from "../components/Alerta";
import { BadgeStatus } from "../components/BadgeStatus";
import { formatarDataHora } from "../utils/formatacao";
import { NOME_CONTEXTO } from "../utils/rotulos";
import { useCarregar } from "../utils/useCarregar";
import { VotarConsulta } from "./votar/VotarConsulta";
import { VotarEleicao } from "./votar/VotarEleicao";
import { VotarPauta } from "./votar/VotarPauta";
import "../styles/styles-pages/Votar.css";

function Votar() {
  const id = Number(useParams().id);
  const { usuario } = useAuth();
  const carregar = useCallback(() => obterVotacao(id), [id]);
  const { dados: votacao, erro, carregando } = useCarregar(carregar);
  const [mensagemSucesso, setMensagemSucesso] = useState("");

  if (carregando) {
    return <Carregando texto="Carregando votação…" />;
  }
  if (!votacao) {
    return (
      <div className="pagina">
        <Alerta tipo="erro">{erro}</Alerta>
        <Link to="/">Voltar ao painel</Link>
      </div>
    );
  }

  return (
    <div className="pagina pagina-votar">
      <header className="pagina-cabecalho">
        <p className="contexto">
          {NOME_CONTEXTO[votacao.tipo]}
          {votacao.tipo === "eleicao" && ` · ${votacao.turno}º turno`}
        </p>
        <div className="titulo-com-badge">
          <h1>{votacao.titulo}</h1>
          <BadgeStatus status={votacao.status} />
        </div>
        {votacao.descricao && <p>{votacao.descricao}</p>}
        <p className="texto-secundario">
          De {formatarDataHora(votacao.inicio)} até {formatarDataHora(votacao.fim)}
        </p>
      </header>

      {mensagemSucesso ? (
        <VotoRegistrado mensagem={mensagemSucesso} secreto={votacao.tipo !== "pauta"} />
      ) : (
        <AreaDeVoto votacao={votacao} admin={usuario?.papel === "admin"} onVotou={setMensagemSucesso} />
      )}
    </div>
  );
}

type AreaProps = {
  votacao: VotacaoDetalhe;
  admin: boolean;
  onVotou: (mensagem: string) => void;
};

function AreaDeVoto({ votacao, admin, onVotou }: AreaProps) {
  if (votacao.status === "encerrada") {
    return (
      <Alerta tipo="info" titulo="Votação encerrada">
        Não é mais possível votar.{" "}
        <Link to={`/votacoes/${votacao.id}/resultado`}>Ver resultado</Link>
      </Alerta>
    );
  }
  if (votacao.status === "nao_iniciada") {
    return (
      <Alerta tipo="info" titulo="Votação não iniciada">
        A votação abre em {formatarDataHora(votacao.inicio)}.
      </Alerta>
    );
  }
  if (admin) {
    return (
      <Alerta tipo="info" titulo="Acompanhamento">
        Administradores não votam; acompanhe e encerre a votação pelo <Link to="/">Painel</Link>.
      </Alerta>
    );
  }
  if (votacao.ja_votou) {
    return (
      <Alerta tipo="sucesso" titulo="Você já votou nesta votação">
        Cada eleitor vota uma única vez. <Link to="/">Voltar ao painel</Link>
      </Alerta>
    );
  }
  if (votacao.tipo === "pauta") {
    return <VotarPauta votacao={votacao} onVotou={onVotou} />;
  }
  if (!votacao.apto) {
    return (
      <Alerta tipo="aviso" titulo="Você não está apto">
        Seu usuário não está na lista de eleitores habilitados desta votação.
      </Alerta>
    );
  }
  if (votacao.tipo === "eleicao") {
    return <VotarEleicao votacao={votacao} onVotou={onVotou} />;
  }
  return <VotarConsulta votacao={votacao} onVotou={onVotou} />;
}

function VotoRegistrado({ mensagem, secreto }: { mensagem: string; secreto: boolean }) {
  return (
    <section className="cartao voto-registrado" aria-labelledby="titulo-voto-registrado">
      <h2 id="titulo-voto-registrado">Voto registrado</h2>
      <Alerta tipo="sucesso">{mensagem}</Alerta>
      <p className="texto-secundario">
        {secreto
          ? "Seu voto é secreto: a escolha não é exibida e não fica associada ao seu nome."
          : "O voto da sua unidade foi computado com o peso da fração ideal."}
      </p>
      <Link to="/" className="botao">
        Voltar ao painel
      </Link>
    </section>
  );
}

export default Votar;
