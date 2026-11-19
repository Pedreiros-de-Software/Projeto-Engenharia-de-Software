import { Fragment } from "react";

import type { ResultadoConsulta } from "../../api/types";
import { Alerta } from "../../components/Alerta";
import { Indicador } from "../../components/Indicador";
import { RolagemDaTabela } from "../../components/RolagemDaTabela";
import { formatarMoeda } from "../../utils/formatacao";

const COLUNAS = 7;

/** Ranking Borda dos projetos com a linha de corte do orçamento. */
export function RankingDaConsulta({ resultado }: { resultado: ResultadoConsulta }) {
  const { ranking } = resultado;
  const indiceCorte = ranking.findIndex((projeto) => !projeto.dentro_do_orcamento);
  const totalProjetos = ranking.length;

  return (
    <div className="resultado">
      <div className="indicadores">
        <Indicador rotulo="Orçamento" valor={formatarMoeda(resultado.orcamento_centavos)} />
        <Indicador
          rotulo="Total selecionado"
          valor={formatarMoeda(resultado.total_selecionado_centavos)}
        />
        <Indicador rotulo="Saldo" valor={formatarMoeda(resultado.saldo_centavos)} />
        <Indicador rotulo="Cédulas apuradas" valor={String(resultado.cedulas)} />
      </div>

      {resultado.empate_na_linha_de_corte && (
        <Alerta tipo="aviso" titulo="Empate na linha de corte">
          O último projeto dentro do orçamento e o primeiro fora dele estão empatados. A decisão
          fica com a comissão organizadora.
        </Alerta>
      )}

      <RolagemDaTabela rotulo="Ranking pela pontuação Borda">
        <table className="tabela tabela-ranking">
          <caption>Ranking pela pontuação Borda</caption>
          <thead>
            <tr>
              <th scope="col" className="numero">
                Posição
              </th>
              <th scope="col">Projeto</th>
              <th scope="col" className="numero">
                Pontos
              </th>
              <th scope="col" className="numero">
                1ªs preferências
              </th>
              <th scope="col" className="numero">
                Custo
              </th>
              <th scope="col" className="numero">
                Custo acumulado
              </th>
              <th scope="col">Situação</th>
            </tr>
          </thead>
          <tbody>
            {ranking.map((projeto, indice) => (
              <Fragment key={projeto.opcao_id}>
                {indice === indiceCorte && (
                  <tr className="linha-corte">
                    <td colSpan={COLUNAS}>Linha de corte do orçamento</td>
                  </tr>
                )}
                <tr className={projeto.dentro_do_orcamento ? "dentro" : "fora"}>
                  <td className="numero">{projeto.posicao}º</td>
                  <th scope="row">
                    {projeto.nome}
                    {projeto.empatado && <span className="etiqueta etiqueta-aviso">Empatado</span>}
                  </th>
                  <td className="numero">{projeto.pontos}</td>
                  <td className="numero">{projeto.primeiras_preferencias}</td>
                  <td className="numero">{formatarMoeda(projeto.custo_centavos)}</td>
                  <td className="numero">{formatarMoeda(projeto.custo_acumulado_centavos)}</td>
                  <td>
                    <span className={`badge ${projeto.dentro_do_orcamento ? "badge-dentro" : "badge-fora"}`}>
                      {projeto.dentro_do_orcamento ? "Dentro do orçamento" : "Fora do orçamento"}
                    </span>
                  </td>
                </tr>
              </Fragment>
            ))}
          </tbody>
        </table>
      </RolagemDaTabela>

      <p className="nota">
        Método Borda: com {totalProjetos} projetos, a 1ª posição de cada cédula vale {totalProjetos}{" "}
        pontos, a 2ª vale {totalProjetos - 1}, e assim por diante até 1 ponto para a última. Empate em pontos é desfeito pelo número de 1ªs preferências; se persistir, mantém-se
        a ordem de cadastro. Os projetos entram no orçamento na ordem do ranking até o primeiro que
        não cabe; ele e os seguintes ficam fora.
      </p>
    </div>
  );
}
