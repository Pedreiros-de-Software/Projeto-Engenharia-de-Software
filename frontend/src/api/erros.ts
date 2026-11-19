export class ApiErro extends Error {
  status: number;
  codigo: string;
  mensagem: string;

  constructor(status: number, codigo: string, mensagem: string) {
    super(mensagem);
    this.name = "ApiErro";
    this.status = status;
    this.codigo = codigo;
    this.mensagem = mensagem;
  }
}

export const MENSAGEM_ERRO_GENERICA = "Ocorreu um erro inesperado. Tente novamente.";

/** Texto para mostrar ao usuário a partir de qualquer erro capturado. */
export function mensagemDeErro(erro: unknown): string {
  if (erro instanceof ApiErro && erro.mensagem) {
    return erro.mensagem;
  }
  return MENSAGEM_ERRO_GENERICA;
}
