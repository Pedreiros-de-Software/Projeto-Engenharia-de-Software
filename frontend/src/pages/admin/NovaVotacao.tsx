import { useRef, useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";

import { criarVotacao, listarUsuarios } from "../../api/client";
import { mensagemDeErro } from "../../api/erros";
import type { Tipo } from "../../api/types";
import { Alerta } from "../../components/Alerta";
import { formatarMoeda, paraCampoDataHora, reaisParaCentavos } from "../../utils/formatacao";
import { useCarregar } from "../../utils/useCarregar";
import { MAXIMO_OPCOES, montarNovaVotacao, type OpcaoFormulario } from "./montarNovaVotacao";
import "../../styles/styles-pages/Admin.css";

type OpcaoComChave = OpcaoFormulario & { chave: number };

const UM_DIA_MS = 24 * 60 * 60 * 1000;

const ROTULO_TIPO: Record<Tipo, string> = {
  pauta: "Assembleia de condomínio (pauta)",
  eleicao: "Eleição do CA (chapas)",
  consulta: "Priorização de projetos (consulta Borda)",
};

function opcaoVazia(chave: number): OpcaoComChave {
  return { chave, nome: "", descricao: "", custo: "" };
}

/** Cadastro de votação pelo administrador (pauta, eleição ou consulta). */
function NovaVotacao() {
  const navigate = useNavigate();
  const { dados: usuarios, erro: erroUsuarios } = useCarregar(listarUsuarios);
  const proximaChave = useRef(3);

  const [tipo, setTipo] = useState<Tipo>("pauta");
  const [titulo, setTitulo] = useState("");
  const [descricao, setDescricao] = useState("");
  const [inicio, setInicio] = useState(() => paraCampoDataHora(new Date()));
  const [fim, setFim] = useState(() => paraCampoDataHora(new Date(Date.now() + 7 * UM_DIA_MS)));
  const [opcoes, setOpcoes] = useState<OpcaoComChave[]>([opcaoVazia(1), opcaoVazia(2)]);
  const [orcamento, setOrcamento] = useState("");
  const [desmarcados, setDesmarcados] = useState<Set<number>>(new Set());
  const [erro, setErro] = useState("");
  const [enviando, setEnviando] = useState(false);

  const eleitores = (usuarios ?? []).filter((u) => u.papel === "eleitor");
  const aptos = eleitores.filter((u) => !desmarcados.has(u.id)).map((u) => u.id);
  const nomeOpcao = tipo === "eleicao" ? "Chapa" : "Projeto";

  function alterarOpcao(chave: number, campo: keyof OpcaoFormulario, valor: string) {
    setOpcoes(opcoes.map((o) => (o.chave === chave ? { ...o, [campo]: valor } : o)));
  }

  function adicionarOpcao() {
    setOpcoes([...opcoes, opcaoVazia(proximaChave.current)]);
    proximaChave.current += 1;
  }

  function alternarApto(id: number) {
    const novos = new Set(desmarcados);
    if (novos.has(id)) novos.delete(id);
    else novos.add(id);
    setDesmarcados(novos);
  }

  async function aoEnviar(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault();
    const montagem = montarNovaVotacao({ tipo, titulo, descricao, inicio, fim, opcoes, orcamento, aptos });
    if (!montagem.ok) {
      setErro(montagem.erro);
      return;
    }
    setErro("");
    setEnviando(true);
    try {
      const criada = await criarVotacao(montagem.dados);
      navigate("/", { state: { mensagem: `Votação "${criada.titulo}" criada com sucesso.` } });
    } catch (e) {
      setErro(mensagemDeErro(e));
      setEnviando(false);
    }
  }

  return (
    <div className="pagina">
      <header className="pagina-cabecalho">
        <p className="contexto">Administração</p>
        <h1>Nova votação</h1>
      </header>

      <form className="cartao formulario" onSubmit={aoEnviar} noValidate>
        <div className="campo-grupo">
          <label htmlFor="tipo">Tipo de votação</label>
          <select id="tipo" className="campo" value={tipo} onChange={(e) => setTipo(e.target.value as Tipo)}>
            {(Object.keys(ROTULO_TIPO) as Tipo[]).map((t) => (
              <option key={t} value={t}>
                {ROTULO_TIPO[t]}
              </option>
            ))}
          </select>
        </div>

        <div className="campo-grupo">
          <label htmlFor="titulo">Título</label>
          <input
            id="titulo"
            className="campo"
            value={titulo}
            maxLength={200}
            onChange={(e) => setTitulo(e.target.value)}
            required
          />
        </div>

        <div className="campo-grupo">
          <label htmlFor="descricao">Descrição</label>
          <textarea
            id="descricao"
            className="campo"
            rows={3}
            maxLength={2000}
            value={descricao}
            onChange={(e) => setDescricao(e.target.value)}
          />
        </div>

        <div className="campos-lado-a-lado">
          <div className="campo-grupo">
            <label htmlFor="inicio">Início</label>
            <input
              id="inicio"
              className="campo"
              type="datetime-local"
              value={inicio}
              onChange={(e) => setInicio(e.target.value)}
              required
            />
          </div>
          <div className="campo-grupo">
            <label htmlFor="fim">Fim</label>
            <input
              id="fim"
              className="campo"
              type="datetime-local"
              value={fim}
              onChange={(e) => setFim(e.target.value)}
              required
            />
          </div>
        </div>

        {tipo === "pauta" ? (
          <p className="texto-secundario">
            Votam os responsáveis pelas unidades, com peso pela fração ideal. Aprovação com 2/3 do
            peso total cadastrado.
          </p>
        ) : (
          <>
            <section className="secao-formulario" aria-labelledby="titulo-opcoes">
              <h2 id="titulo-opcoes">{tipo === "eleicao" ? "Chapas" : "Projetos"}</h2>
              {tipo === "eleicao" && (
                <p className="texto-secundario">Branco e nulo são incluídos automaticamente na cédula.</p>
              )}
              {opcoes.map((opcao, indice) => (
                <fieldset key={opcao.chave} className="opcao-formulario">
                  <legend>
                    {nomeOpcao} {indice + 1}
                  </legend>
                  <div className="campo-grupo">
                    <label htmlFor={`opcao-${opcao.chave}-nome`}>Nome</label>
                    <input
                      id={`opcao-${opcao.chave}-nome`}
                      className="campo"
                      maxLength={120}
                      value={opcao.nome}
                      onChange={(e) => alterarOpcao(opcao.chave, "nome", e.target.value)}
                    />
                  </div>
                  <div className="campo-grupo">
                    <label htmlFor={`opcao-${opcao.chave}-descricao`}>Descrição</label>
                    <input
                      id={`opcao-${opcao.chave}-descricao`}
                      className="campo"
                      value={opcao.descricao}
                      onChange={(e) => alterarOpcao(opcao.chave, "descricao", e.target.value)}
                    />
                  </div>
                  {tipo === "consulta" && (
                    <div className="campo-grupo campo-curto">
                      <label htmlFor={`opcao-${opcao.chave}-custo`}>Custo (R$)</label>
                      <input
                        id={`opcao-${opcao.chave}-custo`}
                        className="campo"
                        type="text"
                        inputMode="decimal"
                        autoComplete="off"
                        placeholder="ex.: 35.000,00"
                        aria-describedby={`opcao-${opcao.chave}-custo-valor`}
                        value={opcao.custo}
                        onChange={(e) => alterarOpcao(opcao.chave, "custo", e.target.value)}
                      />
                    </div>
                  )}
                  <button
                    type="button"
                    className="botao botao-pequeno botao-fantasma-escuro"
                    aria-label={`Remover ${nomeOpcao} ${indice + 1}`}
                    onClick={() => setOpcoes(opcoes.filter((o) => o.chave !== opcao.chave))}
                    disabled={opcoes.length <= 2}
                  >
                    Remover
                  </button>
                  {tipo === "consulta" && (
                    <ValorEmReais id={`opcao-${opcao.chave}-custo-valor`} texto={opcao.custo} rotulo="Custo" />
                  )}
                </fieldset>
              ))}
              <button
                type="button"
                className="botao botao-secundario"
                onClick={adicionarOpcao}
                disabled={opcoes.length >= MAXIMO_OPCOES}
              >
                Adicionar {nomeOpcao.toLowerCase()}
              </button>
            </section>

            {tipo === "consulta" && (
              <div className="campo-grupo campo-curto">
                <label htmlFor="orcamento">Orçamento disponível (R$)</label>
                <input
                  id="orcamento"
                  className="campo"
                  type="text"
                  inputMode="decimal"
                  autoComplete="off"
                  placeholder="ex.: 60.000,00"
                  aria-describedby="orcamento-valor"
                  value={orcamento}
                  onChange={(e) => setOrcamento(e.target.value)}
                />
                <ValorEmReais id="orcamento-valor" texto={orcamento} rotulo="Orçamento" />
              </div>
            )}

            <fieldset className="secao-formulario">
              <legend>Eleitores aptos</legend>
              {erroUsuarios && <Alerta tipo="erro">{erroUsuarios}</Alerta>}
              <div className="aptos-acoes">
                <span className="texto-secundario">
                  {aptos.length} de {eleitores.length} selecionados
                </span>
                <button type="button" className="botao botao-pequeno botao-secundario" onClick={() => setDesmarcados(new Set())}>
                  Marcar todos
                </button>
                <button
                  type="button"
                  className="botao botao-pequeno botao-secundario"
                  onClick={() => setDesmarcados(new Set(eleitores.map((u) => u.id)))}
                >
                  Desmarcar todos
                </button>
              </div>
              <ul className="lista-aptos">
                {eleitores.map((eleitor) => (
                  <li key={eleitor.id}>
                    <label>
                      <input
                        type="checkbox"
                        checked={!desmarcados.has(eleitor.id)}
                        onChange={() => alternarApto(eleitor.id)}
                      />
                      {eleitor.nome} <span className="texto-secundario">({eleitor.login})</span>
                    </label>
                  </li>
                ))}
              </ul>
            </fieldset>
          </>
        )}

        {erro && <Alerta tipo="erro">{erro}</Alerta>}

        <div className="formulario-acoes">
          <button type="button" className="botao botao-secundario" onClick={() => navigate("/")}>
            Cancelar
          </button>
          <button type="submit" className="botao" disabled={enviando}>
            {enviando ? "Criando…" : "Criar votação"}
          </button>
        </div>
      </form>
    </div>
  );
}

type PropsValor = { id: string; texto: string; rotulo: string };

/** Mostra como o valor digitado em reais será enviado, ou que o formato não foi reconhecido. */
function ValorEmReais({ id, texto, rotulo }: PropsValor) {
  const centavos = reaisParaCentavos(texto);
  let mensagem = "";
  if (texto.trim() !== "") {
    mensagem =
      centavos === null
        ? `${rotulo}: formato não reconhecido (use, por exemplo, 35000, 35.000 ou 35.000,50).`
        : `${rotulo}: ${formatarMoeda(centavos)}`;
  }
  return (
    <small id={id} className={`valor-em-reais${centavos === null ? " valor-invalido" : ""}`} aria-live="polite">
      {mensagem}
    </small>
  );
}

export default NovaVotacao;
