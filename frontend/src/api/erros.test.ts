import { describe, expect, it } from "vitest";

import { ApiErro, MENSAGEM_ERRO_GENERICA, mensagemDeErro } from "./erros";

describe("mensagemDeErro", () => {
  it.each([
    { caso: "ApiErro com mensagem", erro: new ApiErro(403, "inadimplente", "Unidade inadimplente."), esperado: "Unidade inadimplente." },
    { caso: "ApiErro sem mensagem", erro: new ApiErro(500, "erro_interno", ""), esperado: MENSAGEM_ERRO_GENERICA },
    { caso: "Error comum", erro: new Error("detalhe interno"), esperado: MENSAGEM_ERRO_GENERICA },
    { caso: "valor qualquer", erro: "texto", esperado: MENSAGEM_ERRO_GENERICA },
  ])("$caso", ({ erro, esperado }) => {
    expect(mensagemDeErro(erro)).toBe(esperado);
  });
});
