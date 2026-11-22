import { describe, expect, it } from "vitest";

import {
  campoDataHoraParaIso,
  formatarDataHora,
  formatarMoeda,
  formatarNumero,
  formatarPercentual,
  paraCampoDataHora,
  reaisParaCentavos,
} from "./formatacao";

const espaco = "\u00a0"; // Intl usa espaço não separável entre "R$" e o valor

describe("formatarMoeda", () => {
  it.each([
    { centavos: 0, esperado: `R$${espaco}0,00` },
    { centavos: 1, esperado: `R$${espaco}0,01` },
    { centavos: 4_500_000, esperado: `R$${espaco}45.000,00` },
    { centavos: 123_456_789, esperado: `R$${espaco}1.234.567,89` },
  ])("$centavos centavos → $esperado", ({ centavos, esperado }) => {
    expect(formatarMoeda(centavos)).toBe(esperado);
  });
});

describe("formatarNumero", () => {
  it.each([
    { valor: 500, esperado: "500" },
    { valor: 1000 / 3, esperado: "333,33" },
    { valor: 12.5, esperado: "12,5" },
    { valor: 1500, esperado: "1.500" },
  ])("$valor → $esperado", ({ valor, esperado }) => {
    expect(formatarNumero(valor)).toBe(esperado);
  });
});

describe("formatarPercentual", () => {
  it.each([
    { valor: 60, esperado: "60,00%" },
    { valor: 200 / 3, esperado: "66,67%" },
    { valor: 0, esperado: "0,00%" },
    { valor: 100, esperado: "100,00%" },
  ])("$valor → $esperado", ({ valor, esperado }) => {
    expect(formatarPercentual(valor)).toBe(esperado);
  });
});

describe("datas", () => {
  it("formata data RFC 3339 no fuso local (America/Sao_Paulo nos testes)", () => {
    expect(formatarDataHora("2026-11-18T13:05:00Z")).toBe("18/11/2026 às 10:05");
  });

  it("converte Date para o valor do campo datetime-local", () => {
    expect(paraCampoDataHora(new Date(2026, 0, 5, 9, 7))).toBe("2026-01-05T09:07");
  });

  it("converte o valor do campo datetime-local para RFC 3339 em UTC", () => {
    expect(campoDataHoraParaIso("2026-11-18T10:05")).toBe("2026-11-18T13:05:00.000Z");
  });
});

describe("reaisParaCentavos", () => {
  it.each([
    { texto: "35000", esperado: 3_500_000 },
    { texto: "12.5", esperado: 1250 },
    { texto: "0.1", esperado: 10 },
    { texto: "0", esperado: 0 },
    { texto: "35.000", esperado: 3_500_000 },
    { texto: "35.000,50", esperado: 3_500_050 },
    { texto: "1.234.567,89", esperado: 123_456_789 },
    { texto: "12000,50", esperado: 1_200_050 },
    { texto: "12000,5", esperado: 1_200_050 },
    { texto: "60.000,00", esperado: 6_000_000 },
    { texto: "1.500", esperado: 150_000 },
    { texto: " 9.000 ", esperado: 900_000 },
    { texto: "  ", esperado: null },
    { texto: "-1", esperado: null },
    { texto: "abc", esperado: null },
    { texto: "1.5000", esperado: null },
    { texto: "12.345", esperado: 1_234_500 },
    { texto: "12.3456", esperado: null },
    { texto: "1,234", esperado: null },
    { texto: "1.000.00", esperado: null },
    { texto: "35.000.50", esperado: null },
    { texto: "R$ 35.000", esperado: null },
    { texto: "1e5", esperado: null },
    { texto: "99999999999999999999", esperado: null },
  ])('"$texto" → $esperado', ({ texto, esperado }) => {
    expect(reaisParaCentavos(texto)).toBe(esperado);
  });
});
