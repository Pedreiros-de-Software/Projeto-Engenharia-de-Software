import type { Papel, Status, Tipo } from "../api/types";

export const ORDEM_CONTEXTOS: Tipo[] = ["pauta", "eleicao", "consulta"];

export const NOME_CONTEXTO: Record<Tipo, string> = {
  pauta: "Assembleia de condomínio",
  eleicao: "Eleição do CA",
  consulta: "Priorização de projetos",
};

export const NOME_STATUS: Record<Status, string> = {
  nao_iniciada: "Não iniciada",
  aberta: "Aberta",
  encerrada: "Encerrada",
};

export const NOME_PAPEL: Record<Papel, string> = {
  admin: "Administrador",
  eleitor: "Eleitor",
};
