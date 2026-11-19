import type { ResultadoPauta } from "../../api/types";
import { Indicador } from "../../components/Indicador";
import { RolagemDaTabela } from "../../components/RolagemDaTabela";
import { formatarNumero, formatarPercentual } from "../../utils/formatacao";

/** Apuração ponderada da pauta com quórum qualificado de 2/3 do peso total. */
export function ResultadoDaPauta({ resultado }: { resultado: ResultadoPauta }) {
  const { contagem, peso_total: pesoTotal } = resultado;
  const naoVotaram = pesoTotal - contagem.favor - contagem.contra - contagem.abstencao;
  const linhas = [
    { rotulo: "A favor", peso: contagem.favor },
    { rotulo: "Contra", peso: contagem.contra },
    { rotulo: "Abstenção", peso: contagem.abstencao },
    { rotulo: "Não votaram", peso: naoVotaram },
  ];
  const percentualDoTotal = (peso: number) => (pesoTotal > 0 ? (peso / pesoTotal) * 100 : 0);

  return (
    <div className="resultado">
      <div className={`veredito ${resultado.aprovada ? "veredito-aprovada" : "veredito-reprovada"}`}>
        <span className="veredito-rotulo">Resultado</span>
        <strong>{resultado.aprovada ? "APROVADA" : "REPROVADA"}</strong>
        <span>
          {formatarPercentual(resultado.percentual_favor)} do peso total a favor (mínimo para
          aprovar: {formatarPercentual(percentualDoTotal(resultado.limiar))})
        </span>
      </div>

      <div className="indicadores">
        <Indicador rotulo="Peso total cadastrado" valor={formatarNumero(pesoTotal)} />
        <Indicador rotulo="Limiar de aprovação (2/3)" valor={formatarNumero(resultado.limiar)} />
        <Indicador rotulo="Peso a favor" valor={formatarNumero(contagem.favor)} />
        <Indicador
          rotulo="Unidades que votaram"
          valor={`${resultado.unidades_votantes} de ${resultado.unidades_total}`}
        />
      </div>

      <RolagemDaTabela rotulo="Apuração ponderada pela fração ideal">
        <table className="tabela">
          <caption>Apuração ponderada pela fração ideal</caption>
          <thead>
            <tr>
              <th scope="col">Escolha</th>
              <th scope="col" className="numero">
                Peso
              </th>
              <th scope="col" className="numero">
                % do peso total
              </th>
            </tr>
          </thead>
          <tbody>
            {linhas.map((linha) => (
              <tr key={linha.rotulo}>
                <th scope="row">{linha.rotulo}</th>
                <td className="numero">{formatarNumero(linha.peso)}</td>
                <td className="numero">{formatarPercentual(percentualDoTotal(linha.peso))}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </RolagemDaTabela>

      <p className="nota">
        A aprovação exige ≥ 2/3 do peso total cadastrado, incluindo unidades que não votaram.
        Abstenções e ausências não contam a favor.
      </p>
    </div>
  );
}
