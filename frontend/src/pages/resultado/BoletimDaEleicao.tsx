import { useState, type FormEvent } from "react";
import { Link } from "react-router-dom";

import { abrirSegundoTurno } from "../../api/client";
import { mensagemDeErro } from "../../api/erros";
import type { ResultadoEleicao, VotacaoDetalhe } from "../../api/types";
import { Alerta } from "../../components/Alerta";
import { Indicador } from "../../components/Indicador";
import { RolagemDaTabela } from "../../components/RolagemDaTabela";
import {
  campoDataHoraParaIso,
  formatarDataHora,
  formatarPercentual,
  paraCampoDataHora,
} from "../../utils/formatacao";

type Props = {
  resultado: ResultadoEleicao;
  admin: boolean;
  logado: boolean;
};

/** Boletim público da eleição: comparecimento, votos por chapa e desfecho (vencedora ou 2º turno). */
export function BoletimDaEleicao({ resultado, admin, logado }: Props) {
  return (
    <div className="resultado">
      <Desfecho resultado={resultado} admin={admin} logado={logado} />

      <div className="indicadores">
        <Indicador rotulo="Eleitores aptos" valor={String(resultado.aptos)} />
        <Indicador
          rotulo="Participantes"
          valor={String(resultado.participantes)}
          detalhe={`Comparecimento de ${formatarPercentual(resultado.percentual_comparecimento)}`}
        />
        <Indicador rotulo="Abstenções" valor={String(resultado.abstencoes)} />
        <Indicador rotulo="Votos válidos" valor={String(resultado.validos)} />
      </div>

      <RolagemDaTabela rotulo={`Boletim de apuração · ${resultado.turno}º turno`}>
        <table className="tabela">
          <caption>Boletim de apuração · {resultado.turno}º turno</caption>
          <thead>
            <tr>
              <th scope="col">Opção</th>
              <th scope="col" className="numero">
                Votos
              </th>
              <th scope="col" className="numero">
                % dos válidos
              </th>
            </tr>
          </thead>
          <tbody>
            {resultado.votos.map((chapa) => (
              <tr key={chapa.opcao_id}>
                <th scope="row">{chapa.nome}</th>
                <td className="numero">{chapa.votos}</td>
                <td className="numero">{formatarPercentual(chapa.percentual_validos)}</td>
              </tr>
            ))}
            <tr className="linha-secundaria">
              <th scope="row">Brancos</th>
              <td className="numero">{resultado.brancos}</td>
              <td className="numero">—</td>
            </tr>
            <tr className="linha-secundaria">
              <th scope="row">Nulos</th>
              <td className="numero">{resultado.nulos}</td>
              <td className="numero">—</td>
            </tr>
          </tbody>
        </table>
      </RolagemDaTabela>

      <p className="nota">
        Percentuais calculados sobre {resultado.validos} votos válidos (brancos e nulos excluídos).{" "}
        {resultado.turno === 1
          ? "Vence no 1º turno a chapa com mais da metade dos votos válidos."
          : "No 2º turno vence a chapa mais votada."}{" "}
        O boletim não traz horários nem dados individuais.
      </p>
    </div>
  );
}

function Desfecho({ resultado, admin, logado }: Props) {
  const { vencedora, segundo_turno: segundoTurno } = resultado;

  if (vencedora) {
    return (
      <div className="veredito veredito-aprovada">
        <span className="veredito-rotulo">Chapa vencedora</span>
        <strong>{vencedora.nome}</strong>
      </div>
    );
  }
  if (segundoTurno.empate_indefinido) {
    return (
      <div className="veredito veredito-neutro">
        <span className="veredito-rotulo">Sem maioria absoluta</span>
        <strong>Empate impede definir o 2º turno: decisão da comissão</strong>
      </div>
    );
  }
  if (resultado.turno === 2 && resultado.validos > 0) {
    return (
      <div className="veredito veredito-neutro">
        <span className="veredito-rotulo">Sem maioria no 2º turno</span>
        <strong>Empate no 2º turno: decisão da comissão</strong>
      </div>
    );
  }
  if (segundoTurno.necessario && segundoTurno.chapas.length === 2) {
    const [primeira, segunda] = segundoTurno.chapas;
    return (
      <div className="veredito veredito-neutro">
        <span className="veredito-rotulo">Sem maioria absoluta</span>
        <strong>
          2º turno entre {primeira.nome} e {segunda.nome}
        </strong>
        {segundoTurno.votacao_id !== null ? (
          // Sem login só o boletim é acessível; logado, a página da votação permite votar ou ver o resultado.
          <Link
            to={logado ? `/votacoes/${segundoTurno.votacao_id}` : `/votacoes/${segundoTurno.votacao_id}/resultado`}
            className="botao"
          >
            Ir para o 2º turno
          </Link>
        ) : (
          admin && <FormSegundoTurno votacaoId={resultado.votacao_id} />
        )}
      </div>
    );
  }
  return (
    <div className="veredito veredito-reprovada">
      <span className="veredito-rotulo">Resultado</span>
      <strong>
        {resultado.validos === 0 ? "Sem votos válidos: sem vencedora" : "Sem vencedora definida"}
      </strong>
    </div>
  );
}

function FormSegundoTurno({ votacaoId }: { votacaoId: number }) {
  const [fim, setFim] = useState(() => paraCampoDataHora(new Date(Date.now() + 24 * 60 * 60 * 1000)));
  const [enviando, setEnviando] = useState(false);
  const [erro, setErro] = useState("");
  const [aberto, setAberto] = useState<VotacaoDetalhe | null>(null);

  async function aoEnviar(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault();
    setEnviando(true);
    setErro("");
    try {
      setAberto(await abrirSegundoTurno(votacaoId, campoDataHoraParaIso(fim)));
    } catch (e) {
      setErro(mensagemDeErro(e));
    } finally {
      setEnviando(false);
    }
  }

  if (aberto) {
    return (
      <Alerta tipo="sucesso" titulo="2º turno aberto">
        Votação aberta até {formatarDataHora(aberto.fim)}.{" "}
        <Link to={`/votacoes/${aberto.id}`}>Ir para o 2º turno</Link>
      </Alerta>
    );
  }

  return (
    <form className="form-segundo-turno" onSubmit={aoEnviar}>
      <label htmlFor="fim-segundo-turno">Término do 2º turno</label>
      <input
        id="fim-segundo-turno"
        className="campo"
        type="datetime-local"
        value={fim}
        onChange={(e) => setFim(e.target.value)}
        required
      />
      <button type="submit" className="botao" disabled={enviando}>
        {enviando ? "Abrindo…" : "Abrir 2º turno"}
      </button>
      {erro && <Alerta tipo="erro">{erro}</Alerta>}
    </form>
  );
}
