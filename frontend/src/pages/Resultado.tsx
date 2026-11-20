import { useCallback } from "react";
import { Link, useParams } from "react-router-dom";

import { obterResultado } from "../api/client";
import type { Resultado as ResultadoApi } from "../api/types";
import { useAuth } from "../auth/useAuth";
import { Alerta, Carregando } from "../components/Alerta";
import { NOME_CONTEXTO } from "../utils/rotulos";
import { useCarregar } from "../utils/useCarregar";
import { BoletimDaEleicao } from "./resultado/BoletimDaEleicao";
import { RankingDaConsulta } from "./resultado/RankingDaConsulta";
import { ResultadoDaPauta } from "./resultado/ResultadoDaPauta";
import "../styles/styles-pages/Resultado.css";

/** Resultado público de uma votação encerrada (funciona sem login). */
function Resultado() {
  const id = Number(useParams().id);
  const { usuario } = useAuth();
  const carregarResultado = useCallback(() => obterResultado(id), [id]);
  const { dados: resultado, erro, carregando } = useCarregar(carregarResultado);
  const logado = usuario !== null;

  return (
    <div className="pagina pagina-resultado">
      <header className="pagina-cabecalho">
        <p className="contexto">{resultado ? `Resultado · ${NOME_CONTEXTO[resultado.tipo]}` : "Resultado"}</p>
        <h1>{resultado?.titulo || `Votação nº ${id}`}</h1>
        {resultado?.descricao && <p>{resultado.descricao}</p>}
      </header>

      {carregando && <Carregando texto="Carregando resultado…" />}
      {erro && <Alerta tipo="erro">{erro}</Alerta>}
      {resultado && (
        <ConteudoResultado resultado={resultado} admin={usuario?.papel === "admin"} logado={logado} />
      )}

      {logado && (
        <p>
          <Link to="/">Voltar ao painel</Link>
        </p>
      )}
    </div>
  );
}

type PropsConteudo = { resultado: ResultadoApi; admin: boolean; logado: boolean };

function ConteudoResultado({ resultado, admin, logado }: PropsConteudo) {
  switch (resultado.tipo) {
    case "pauta":
      return <ResultadoDaPauta resultado={resultado} />;
    case "eleicao":
      return <BoletimDaEleicao resultado={resultado} admin={admin} logado={logado} />;
    case "consulta":
      return <RankingDaConsulta resultado={resultado} />;
  }
}

export default Resultado;
