import { useState } from "react";
import { useLocation } from "react-router-dom";

import { encerrarVotacao, listarVotacoes } from "../api/client";
import { mensagemDeErro } from "../api/erros";
import type { Tipo, VotacaoResumo } from "../api/types";
import { useAuth } from "../auth/useAuth";
import { Alerta, Carregando } from "../components/Alerta";
import { CartaoVotacao } from "../components/CartaoVotacao";
import { DialogoConfirmacao } from "../components/DialogoConfirmacao";
import { NOME_CONTEXTO, ORDEM_CONTEXTOS } from "../utils/rotulos";
import { useCarregar } from "../utils/useCarregar";
import "../styles/styles-pages/Home.css";

type Aviso = { tipo: "sucesso" | "erro"; texto: string };

/** Painel: lista as votações agrupadas por contexto, com as ações permitidas a cada uma. */
function Home() {
  const { usuario } = useAuth();
  const location = useLocation();
  const mensagemRecebida = (location.state as { mensagem?: string } | null)?.mensagem;
  const { dados: votacoes, erro, carregando, atualizarDados } = useCarregar(listarVotacoes);
  const [aEncerrar, setAEncerrar] = useState<VotacaoResumo | null>(null);
  const [aviso, setAviso] = useState<Aviso | null>(
    mensagemRecebida ? { tipo: "sucesso", texto: mensagemRecebida } : null,
  );
  const admin = usuario?.papel === "admin";

  async function confirmarEncerramento(alvo: VotacaoResumo) {
    setAEncerrar(null);
    try {
      const atualizada = await encerrarVotacao(alvo.id);
      atualizarDados((votacoes ?? []).map((v) => (v.id === atualizada.id ? atualizada : v)));
      setAviso({ tipo: "sucesso", texto: `Votação "${atualizada.titulo}" encerrada.` });
    } catch (e) {
      setAviso({ tipo: "erro", texto: mensagemDeErro(e) });
    }
  }

  return (
    <div className="painel">
      <section className="painel-apresentacao">
        <h1>Consensu</h1>
        <h3>
          Sistema de votação para Eleições de CA, Assembleia de Condomínio e Priorização de projetos
        </h3>
      </section>

      {aviso && <Alerta tipo={aviso.tipo}>{aviso.texto}</Alerta>}
      {carregando && <Carregando texto="Carregando votações…" />}
      {erro && <Alerta tipo="erro">{erro}</Alerta>}

      {votacoes &&
        ORDEM_CONTEXTOS.map((tipo) => (
          <GrupoVotacoes
            key={tipo}
            tipo={tipo}
            votacoes={votacoes.filter((v) => v.tipo === tipo)}
            admin={admin}
            onEncerrar={setAEncerrar}
          />
        ))}

      {aEncerrar && (
        <DialogoConfirmacao
          titulo="Encerrar votação agora?"
          textoConfirmar="Encerrar votação"
          onConfirmar={() => confirmarEncerramento(aEncerrar)}
          onCancelar={() => setAEncerrar(null)}
        >
          <p>
            <strong>{aEncerrar.titulo}</strong> será encerrada imediatamente: ninguém mais poderá
            votar e o resultado ficará disponível.
          </p>
        </DialogoConfirmacao>
      )}
    </div>
  );
}

type GrupoProps = {
  tipo: Tipo;
  votacoes: VotacaoResumo[];
  admin: boolean;
  onEncerrar: (votacao: VotacaoResumo) => void;
};

function GrupoVotacoes({ tipo, votacoes, admin, onEncerrar }: GrupoProps) {
  const idTitulo = `grupo-${tipo}`;
  return (
    <section className="painel-grupo" aria-labelledby={idTitulo}>
      <h2 id={idTitulo}>{NOME_CONTEXTO[tipo]}</h2>
      {votacoes.length === 0 ? (
        <p className="painel-vazio">Nenhuma votação neste contexto.</p>
      ) : (
        <div className="painel-lista">
          {votacoes.map((votacao) => (
            <CartaoVotacao key={votacao.id} votacao={votacao} admin={admin} onEncerrar={onEncerrar} />
          ))}
        </div>
      )}
    </section>
  );
}

export default Home;
