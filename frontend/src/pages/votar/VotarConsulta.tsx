import { useState } from "react";

import type { Opcao, VotacaoDetalhe } from "../../api/types";
import { Alerta, Carregando } from "../../components/Alerta";
import { DialogoConfirmacao } from "../../components/DialogoConfirmacao";
import { formatarMoeda } from "../../utils/formatacao";
import { useEnviarVoto } from "./useEnviarVoto";

type Props = {
  votacao: VotacaoDetalhe;
  onVotou: (mensagem: string) => void;
};

/** Cédula Borda: o eleitor ordena todos os projetos, do mais ao menos importante. */
export function VotarConsulta({ votacao, onVotou }: Props) {
  const projetos = votacao.opcoes;
  const [ordem, setOrdem] = useState<Opcao[]>([]);
  const [confirmando, setConfirmando] = useState(false);
  const { enviando, erro, enviar } = useEnviarVoto(votacao.id, onVotou);

  const restantes = projetos.filter((p) => !ordem.includes(p));
  const completa = projetos.length > 0 && restantes.length === 0;

  function mover(indice: number, destino: number) {
    const nova = [...ordem];
    [nova[indice], nova[destino]] = [nova[destino], nova[indice]];
    setOrdem(nova);
  }

  function confirmar() {
    setConfirmando(false);
    void enviar({ preferencias: ordem.map((p) => p.id) });
  }

  return (
    <section className="cartao" aria-labelledby="titulo-cedula">
      <h2 id="titulo-cedula">Cédula de priorização</h2>
      <p className="texto-secundario">
        Clique nos projetos na ordem da sua preferência: o primeiro é o mais importante. Com{" "}
        {projetos.length} projetos, a 1ª posição vale {projetos.length} pontos e a última vale 1
        ponto (método Borda).
        {votacao.orcamento_centavos !== null &&
          ` Orçamento disponível: ${formatarMoeda(votacao.orcamento_centavos)}.`}
      </p>

      <div className="ordenacao">
        <section className="ordenacao-coluna" aria-labelledby="titulo-restantes">
          <h3 id="titulo-restantes">Projetos a ordenar</h3>
          {restantes.length === 0 ? (
            <p className="texto-secundario">Todos os projetos foram ordenados.</p>
          ) : (
            <ul className="lista-projetos">
              {restantes.map((projeto) => (
                <li key={projeto.id}>
                  <button
                    type="button"
                    className="item-projeto"
                    onClick={() => setOrdem([...ordem, projeto])}
                    disabled={enviando}
                  >
                    <span className="item-projeto-nome">{projeto.nome}</span>{" "}
                    {projeto.custo_centavos !== null && (
                      <span className="item-projeto-custo">{formatarMoeda(projeto.custo_centavos)}</span>
                    )}{" "}
                    {projeto.descricao && (
                      <span className="item-projeto-descricao">{projeto.descricao}</span>
                    )}
                  </button>
                </li>
              ))}
            </ul>
          )}
        </section>

        <section className="ordenacao-coluna" aria-labelledby="titulo-ordem">
          <h3 id="titulo-ordem">Sua ordem</h3>
          {ordem.length === 0 ? (
            <p className="texto-secundario">Nenhum projeto escolhido ainda.</p>
          ) : (
            <ol className="lista-ordem">
              {ordem.map((projeto, indice) => (
                <li key={projeto.id} className="item-ordem">
                  <span className="item-ordem-posicao">{indice + 1}º</span>
                  <span className="item-ordem-nome">
                    {projeto.nome}
                    <small>{projetos.length - indice === 1 ? "1 ponto" : `${projetos.length - indice} pontos`}</small>
                  </span>
                  <span className="item-ordem-acoes">
                    <button
                      type="button"
                      className="botao botao-pequeno botao-secundario"
                      aria-label={`Subir ${projeto.nome}`}
                      onClick={() => mover(indice, indice - 1)}
                      disabled={indice === 0 || enviando}
                    >
                      Subir
                    </button>
                    <button
                      type="button"
                      className="botao botao-pequeno botao-secundario"
                      aria-label={`Descer ${projeto.nome}`}
                      onClick={() => mover(indice, indice + 1)}
                      disabled={indice === ordem.length - 1 || enviando}
                    >
                      Descer
                    </button>
                    <button
                      type="button"
                      className="botao botao-pequeno botao-fantasma-escuro"
                      aria-label={`Remover ${projeto.nome}`}
                      onClick={() => setOrdem(ordem.filter((p) => p !== projeto))}
                      disabled={enviando}
                    >
                      Remover
                    </button>
                  </span>
                </li>
              ))}
            </ol>
          )}
        </section>
      </div>

      <div className="cedula-envio">
        <span className="texto-secundario">
          {ordem.length} de {projetos.length} projetos ordenados
        </span>
        <button
          type="button"
          className="botao botao-grande"
          disabled={!completa || enviando}
          onClick={() => setConfirmando(true)}
        >
          Enviar cédula
        </button>
      </div>

      {enviando && <Carregando texto="Registrando cédula…" />}
      {erro && <Alerta tipo="erro">{erro}</Alerta>}

      {confirmando && (
        <DialogoConfirmacao
          titulo="Confirmar cédula"
          textoConfirmar="Enviar cédula"
          onConfirmar={confirmar}
          onCancelar={() => setConfirmando(false)}
        >
          <p>Sua ordem de preferência:</p>
          <ol className="lista-confirmacao">
            {ordem.map((projeto) => (
              <li key={projeto.id}>{projeto.nome}</li>
            ))}
          </ol>
          <p>Depois de enviada, a cédula não pode ser alterada.</p>
        </DialogoConfirmacao>
      )}
    </section>
  );
}
