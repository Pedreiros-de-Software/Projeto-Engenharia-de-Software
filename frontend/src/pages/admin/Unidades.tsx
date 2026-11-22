import { useState } from "react";

import { alterarAdimplencia, listarUnidades } from "../../api/client";
import { mensagemDeErro } from "../../api/erros";
import type { Unidade } from "../../api/types";
import { Alerta, Carregando } from "../../components/Alerta";
import { RolagemDaTabela } from "../../components/RolagemDaTabela";
import { formatarNumero, formatarPercentual } from "../../utils/formatacao";
import { useCarregar } from "../../utils/useCarregar";
import "../../styles/styles-pages/Admin.css";

type Aviso = { tipo: "sucesso" | "erro"; texto: string };

/** Unidades do condomínio com a situação de adimplência, alterável pelo administrador. */
function Unidades() {
  const { dados: unidades, erro, carregando, atualizarDados } = useCarregar(listarUnidades);
  const [aviso, setAviso] = useState<Aviso | null>(null);
  const [alterando, setAlterando] = useState<number | null>(null);

  async function alternar(unidade: Unidade) {
    setAlterando(unidade.id);
    try {
      const atualizada = await alterarAdimplencia(unidade.id, !unidade.adimplente);
      atualizarDados((unidades ?? []).map((u) => (u.id === atualizada.id ? atualizada : u)));
      const situacao = atualizada.adimplente ? "adimplente" : "inadimplente";
      setAviso({ tipo: "sucesso", texto: `Unidade ${atualizada.identificador} marcada como ${situacao}.` });
    } catch (e) {
      setAviso({ tipo: "erro", texto: mensagemDeErro(e) });
    } finally {
      setAlterando(null);
    }
  }

  return (
    <div className="pagina">
      <header className="pagina-cabecalho">
        <p className="contexto">Administração · Assembleia de condomínio</p>
        <h1>Unidades</h1>
        <p className="texto-secundario">
          O peso de cada unidade nas pautas é a sua fração ideal. Unidades inadimplentes não podem
          votar.
        </p>
      </header>

      {aviso && <Alerta tipo={aviso.tipo}>{aviso.texto}</Alerta>}
      {carregando && <Carregando texto="Carregando unidades…" />}
      {erro && <Alerta tipo="erro">{erro}</Alerta>}

      {unidades && (
        <div className="cartao">
          <RolagemDaTabela rotulo="Unidades cadastradas">
            <table className="tabela">
              <caption>Unidades cadastradas</caption>
              <thead>
                <tr>
                  <th scope="col">Unidade</th>
                  <th scope="col">Responsável</th>
                  <th scope="col" className="numero">
                    Fração ideal
                  </th>
                  <th scope="col" className="numero">
                    % do total
                  </th>
                  <th scope="col">Situação</th>
                  <th scope="col">Ação</th>
                </tr>
              </thead>
              <tbody>
                {unidades.map((unidade) => (
                  <tr key={unidade.id}>
                    <th scope="row">{unidade.identificador}</th>
                    <td>{unidade.responsavel?.nome ?? "Sem responsável"}</td>
                    <td className="numero">{formatarNumero(unidade.fracao_ideal)}</td>
                    <td className="numero">{formatarPercentual(unidade.percentual)}</td>
                    <td>
                      <span className={`badge ${unidade.adimplente ? "badge-dentro" : "badge-fora"}`}>
                        {unidade.adimplente ? "Adimplente" : "Inadimplente"}
                      </span>
                    </td>
                    <td>
                      <button
                        type="button"
                        className="botao botao-pequeno botao-secundario"
                        onClick={() => alternar(unidade)}
                        disabled={alterando !== null}
                        aria-label={`Marcar ${unidade.identificador} como ${unidade.adimplente ? "inadimplente" : "adimplente"}`}
                      >
                        {unidade.adimplente ? "Marcar inadimplente" : "Marcar adimplente"}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </RolagemDaTabela>
        </div>
      )}
    </div>
  );
}

export default Unidades;
