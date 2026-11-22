const moeda = new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" });
const numero = new Intl.NumberFormat("pt-BR", { maximumFractionDigits: 2 });
const percentual = new Intl.NumberFormat("pt-BR", {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
});
const data = new Intl.DateTimeFormat("pt-BR", { day: "2-digit", month: "2-digit", year: "numeric" });
const hora = new Intl.DateTimeFormat("pt-BR", { hour: "2-digit", minute: "2-digit" });

/** Valor em reais a partir de centavos, por exemplo 4500000 → "R$ 45.000,00". */
export function formatarMoeda(centavos: number): string {
  return moeda.format(centavos / 100);
}

/** Número com até duas casas decimais, por exemplo 333.3333 → "333,33". */
export function formatarNumero(valor: number): string {
  return numero.format(valor);
}

/** Percentual com duas casas, por exemplo 60 → "60,00%". */
export function formatarPercentual(valor: number): string {
  return `${percentual.format(valor)}%`;
}

/** Data e hora locais de uma data RFC 3339, por exemplo "18/11/2026 às 14:30". */
export function formatarDataHora(iso: string): string {
  const instante = new Date(iso);
  return `${data.format(instante)} às ${hora.format(instante)}`;
}

/** Valor no formato do campo datetime-local ("AAAA-MM-DDTHH:mm") no fuso do navegador. */
export function paraCampoDataHora(instante: Date): string {
  const doisDigitos = (n: number) => String(n).padStart(2, "0");
  const dia = `${instante.getFullYear()}-${doisDigitos(instante.getMonth() + 1)}-${doisDigitos(instante.getDate())}`;
  return `${dia}T${doisDigitos(instante.getHours())}:${doisDigitos(instante.getMinutes())}`;
}

/** Converte o valor de um campo datetime-local para RFC 3339 (UTC). */
export function campoDataHoraParaIso(valor: string): string {
  return new Date(valor).toISOString();
}

const REAIS_INTEIROS = /^\d+$/;
const REAIS_PONTO_DECIMAL = /^\d+\.\d{1,2}$/;
const REAIS_VIRGULA_DECIMAL = /^\d+,\d{1,2}$/;
const REAIS_PONTO_MILHAR = /^\d{1,3}(\.\d{3})+(,\d{1,2})?$/;

/** Reais digitados ("35000", "35.000,50", "12000,50", "12.5") em centavos; ponto + 3 dígitos é milhar; null se inválido. */
export function reaisParaCentavos(texto: string): number | null {
  const valor = texto.trim();
  let normalizado: string;
  if (REAIS_INTEIROS.test(valor) || REAIS_PONTO_DECIMAL.test(valor)) {
    normalizado = valor;
  } else if (REAIS_VIRGULA_DECIMAL.test(valor)) {
    normalizado = valor.replace(",", ".");
  } else if (REAIS_PONTO_MILHAR.test(valor)) {
    normalizado = valor.replace(/\./g, "").replace(",", ".");
  } else {
    return null;
  }
  const [inteiro, fracao = ""] = normalizado.split(".");
  const centavos = Number(inteiro) * 100 + Number(fracao.padEnd(2, "0"));
  return Number.isSafeInteger(centavos) ? centavos : null;
}
