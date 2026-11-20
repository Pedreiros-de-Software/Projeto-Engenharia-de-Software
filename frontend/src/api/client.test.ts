import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import * as api from "./client";
import { ApiErro } from "./erros";

const fetchFalso = vi.fn();

function resposta(status: number, corpo?: unknown): Response {
  if (corpo === undefined) {
    return new Response(null, { status });
  }
  return new Response(JSON.stringify(corpo), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function ultimaChamada() {
  const [url, init] = fetchFalso.mock.calls.at(-1) as [string, RequestInit];
  return { url, init, headers: init.headers as Record<string, string> };
}

beforeEach(() => {
  vi.stubGlobal("fetch", fetchFalso);
});

afterEach(() => {
  vi.unstubAllGlobals();
  api.definirAoPerderSessao(null);
});

describe("client: token e cabeçalhos", () => {
  it("envia o token salvo no header Authorization Bearer", async () => {
    api.salvarToken("tok-123");
    fetchFalso.mockResolvedValue(resposta(200, []));

    await api.listarVotacoes();

    const { url, headers } = ultimaChamada();
    expect(url).toBe("/api/votacoes");
    expect(headers.Authorization).toBe("Bearer tok-123");
    expect(localStorage.getItem(api.CHAVE_TOKEN)).toBe("tok-123");
  });

  it("não envia Authorization quando não há token", async () => {
    fetchFalso.mockResolvedValue(resposta(200, []));

    await api.listarVotacoes();

    expect(ultimaChamada().headers.Authorization).toBeUndefined();
  });

  it("login guarda o token recebido e devolve o usuário", async () => {
    const usuario = { id: 2, nome: "Maria Souza", login: "maria", papel: "eleitor" };
    fetchFalso.mockResolvedValue(resposta(200, { token: "novo-token", usuario }));

    const resultado = await api.login("maria", "consensu");

    expect(resultado.usuario).toEqual(usuario);
    expect(api.obterToken()).toBe("novo-token");
    const { init, headers } = ultimaChamada();
    expect(init.method).toBe("POST");
    expect(headers["Content-Type"]).toBe("application/json");
    expect(JSON.parse(init.body as string)).toEqual({ login: "maria", senha: "consensu" });
  });

  it("logout apaga o token mesmo quando o servidor falha", async () => {
    api.salvarToken("tok");
    fetchFalso.mockResolvedValue(resposta(500, { erro: "Erro interno.", codigo: "erro_interno" }));

    await expect(api.logout()).rejects.toBeInstanceOf(ApiErro);

    expect(api.obterToken()).toBeNull();
  });

  it("logout com 204 resolve sem corpo e apaga o token", async () => {
    api.salvarToken("tok");
    fetchFalso.mockResolvedValue(resposta(204));

    await expect(api.logout()).resolves.toBeUndefined();

    expect(api.obterToken()).toBeNull();
  });

  it("resposta 401 apaga o token e avisa que a sessão foi perdida", async () => {
    const aoPerder = vi.fn();
    api.definirAoPerderSessao(aoPerder);
    api.salvarToken("expirado");
    fetchFalso.mockResolvedValue(
      resposta(401, { erro: "Sessão expirada.", codigo: "nao_autenticado" }),
    );

    await expect(api.obterSessao()).rejects.toMatchObject({ status: 401, codigo: "nao_autenticado" });

    expect(api.obterToken()).toBeNull();
    expect(aoPerder).toHaveBeenCalledOnce();
  });
});

describe("client: erros", () => {
  it("converte o corpo de erro da API em ApiErro com status, código e mensagem", async () => {
    fetchFalso.mockResolvedValue(
      resposta(409, { erro: "Você já votou nesta votação.", codigo: "voto_duplicado" }),
    );

    const erro = await api.votar(1, { escolha: "favor" }).catch((e: unknown) => e);

    expect(erro).toBeInstanceOf(ApiErro);
    expect(erro).toMatchObject({
      status: 409,
      codigo: "voto_duplicado",
      mensagem: "Você já votou nesta votação.",
    });
  });

  it("usa mensagem genérica quando o erro não tem corpo JSON", async () => {
    fetchFalso.mockResolvedValue(new Response("Bad Gateway", { status: 502 }));

    await expect(api.listarVotacoes()).rejects.toMatchObject({
      status: 502,
      codigo: "erro_interno",
      mensagem: "Ocorreu um erro inesperado. Tente novamente.",
    });
  });

  it("falha de rede vira ApiErro com status 0", async () => {
    fetchFalso.mockRejectedValue(new TypeError("Failed to fetch"));

    await expect(api.listarVotacoes()).rejects.toMatchObject({
      status: 0,
      codigo: "sem_conexao",
      mensagem: "Não foi possível conectar ao servidor.",
    });
  });
});

describe("client: rotas do contrato", () => {
  const casos: { nome: string; chamar: () => Promise<unknown>; metodo: string; url: string; corpo?: unknown }[] = [
    { nome: "obterSessao", chamar: () => api.obterSessao(), metodo: "GET", url: "/api/me" },
    { nome: "listarVotacoes", chamar: () => api.listarVotacoes(), metodo: "GET", url: "/api/votacoes" },
    { nome: "obterVotacao", chamar: () => api.obterVotacao(7), metodo: "GET", url: "/api/votacoes/7" },
    {
      nome: "criarVotacao",
      chamar: () =>
        api.criarVotacao({ tipo: "pauta", titulo: "T", descricao: "", inicio: "a", fim: "b", opcoes: [] }),
      metodo: "POST",
      url: "/api/votacoes",
      corpo: { tipo: "pauta", titulo: "T", descricao: "", inicio: "a", fim: "b", opcoes: [] },
    },
    { nome: "encerrarVotacao", chamar: () => api.encerrarVotacao(7), metodo: "POST", url: "/api/votacoes/7/encerrar" },
    {
      nome: "votar (consulta)",
      chamar: () => api.votar(7, { preferencias: [3, 1, 2] }),
      metodo: "POST",
      url: "/api/votacoes/7/votos",
      corpo: { preferencias: [3, 1, 2] },
    },
    { nome: "obterResultado", chamar: () => api.obterResultado(7), metodo: "GET", url: "/api/votacoes/7/resultado" },
    {
      nome: "abrirSegundoTurno com fim",
      chamar: () => api.abrirSegundoTurno(7, "2026-11-20T12:00:00.000Z"),
      metodo: "POST",
      url: "/api/votacoes/7/segundo-turno",
      corpo: { fim: "2026-11-20T12:00:00.000Z" },
    },
    {
      nome: "abrirSegundoTurno sem fim",
      chamar: () => api.abrirSegundoTurno(7),
      metodo: "POST",
      url: "/api/votacoes/7/segundo-turno",
      corpo: {},
    },
    { nome: "listarUnidades", chamar: () => api.listarUnidades(), metodo: "GET", url: "/api/unidades" },
    {
      nome: "alterarAdimplencia",
      chamar: () => api.alterarAdimplencia(5, false),
      metodo: "PATCH",
      url: "/api/unidades/5",
      corpo: { adimplente: false },
    },
    { nome: "listarUsuarios", chamar: () => api.listarUsuarios(), metodo: "GET", url: "/api/usuarios" },
  ];

  it.each(casos)("$nome chama $metodo $url", async ({ chamar, metodo, url, corpo }) => {
    fetchFalso.mockResolvedValue(resposta(200, { ok: true }));

    await expect(chamar()).resolves.toEqual({ ok: true });

    const chamada = ultimaChamada();
    expect(chamada.url).toBe(url);
    expect(chamada.init.method).toBe(metodo);
    if (corpo === undefined) {
      expect(chamada.init.body).toBeUndefined();
    } else {
      expect(JSON.parse(chamada.init.body as string)).toEqual(corpo);
    }
  });
});
