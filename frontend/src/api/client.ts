import { ApiErro, MENSAGEM_ERRO_GENERICA } from "./erros";
import type {
  CorpoErro,
  NovaVotacao,
  RespostaLogin,
  RespostaMe,
  RespostaVoto,
  Resultado,
  Unidade,
  Usuario,
  VotacaoDetalhe,
  VotacaoResumo,
  Voto,
} from "./types";

export const CHAVE_TOKEN = "consensu.token";
const BASE = "/api";

let aoPerderSessao: (() => void) | null = null;

export function obterToken(): string | null {
  return localStorage.getItem(CHAVE_TOKEN);
}

export function salvarToken(token: string): void {
  localStorage.setItem(CHAVE_TOKEN, token);
}

export function limparToken(): void {
  localStorage.removeItem(CHAVE_TOKEN);
}

/** Registra a função chamada quando o servidor responde 401 (sessão inválida ou expirada). */
export function definirAoPerderSessao(funcao: (() => void) | null): void {
  aoPerderSessao = funcao;
}

async function requisitar<T>(metodo: string, caminho: string, corpo?: unknown): Promise<T> {
  const headers: Record<string, string> = { Accept: "application/json" };
  const token = obterToken();
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  if (corpo !== undefined) {
    headers["Content-Type"] = "application/json";
  }

  let resposta: Response;
  try {
    resposta = await fetch(BASE + caminho, {
      method: metodo,
      headers,
      body: corpo === undefined ? undefined : JSON.stringify(corpo),
    });
  } catch {
    throw new ApiErro(0, "sem_conexao", "Não foi possível conectar ao servidor.");
  }

  if (resposta.status === 401) {
    limparToken();
    aoPerderSessao?.();
  }
  if (!resposta.ok) {
    throw await lerErro(resposta);
  }
  if (resposta.status === 204) {
    return undefined as T;
  }
  return (await resposta.json()) as T;
}

async function lerErro(resposta: Response): Promise<ApiErro> {
  try {
    const corpo = (await resposta.json()) as Partial<CorpoErro>;
    if (corpo.erro && corpo.codigo) {
      return new ApiErro(resposta.status, corpo.codigo, corpo.erro);
    }
  } catch {
    // corpo ausente ou não JSON: usa a mensagem genérica
  }
  return new ApiErro(resposta.status, "erro_interno", MENSAGEM_ERRO_GENERICA);
}

export async function login(nomeDeUsuario: string, senha: string): Promise<RespostaLogin> {
  const resposta = await requisitar<RespostaLogin>("POST", "/login", { login: nomeDeUsuario, senha });
  salvarToken(resposta.token);
  return resposta;
}

export async function logout(): Promise<void> {
  try {
    await requisitar<void>("POST", "/logout");
  } finally {
    limparToken();
  }
}

export function obterSessao(): Promise<RespostaMe> {
  return requisitar("GET", "/me");
}

export function listarVotacoes(): Promise<VotacaoResumo[]> {
  return requisitar("GET", "/votacoes");
}

export function obterVotacao(id: number): Promise<VotacaoDetalhe> {
  return requisitar("GET", `/votacoes/${id}`);
}

export function criarVotacao(dados: NovaVotacao): Promise<VotacaoDetalhe> {
  return requisitar("POST", "/votacoes", dados);
}

export function encerrarVotacao(id: number): Promise<VotacaoResumo> {
  return requisitar("POST", `/votacoes/${id}/encerrar`);
}

export function votar(id: number, voto: Voto): Promise<RespostaVoto> {
  return requisitar("POST", `/votacoes/${id}/votos`, voto);
}

export function obterResultado(id: number): Promise<Resultado> {
  return requisitar("GET", `/votacoes/${id}/resultado`);
}

export function abrirSegundoTurno(id: number, fim?: string): Promise<VotacaoDetalhe> {
  return requisitar("POST", `/votacoes/${id}/segundo-turno`, fim ? { fim } : {});
}

export function listarUnidades(): Promise<Unidade[]> {
  return requisitar("GET", "/unidades");
}

export function alterarAdimplencia(id: number, adimplente: boolean): Promise<Unidade> {
  return requisitar("PATCH", `/unidades/${id}`, { adimplente });
}

export function listarUsuarios(): Promise<Usuario[]> {
  return requisitar("GET", "/usuarios");
}
