import type { ReactNode } from "react";

type TipoAlerta = "sucesso" | "erro" | "aviso" | "info";

type Props = {
  tipo: TipoAlerta;
  titulo?: string;
  children: ReactNode;
};

/** Caixa de mensagem: erros usam role="alert"; os demais, role="status". */
export function Alerta({ tipo, titulo, children }: Props) {
  return (
    <div className={`alerta alerta-${tipo}`} role={tipo === "erro" ? "alert" : "status"}>
      {titulo && <strong className="alerta-titulo">{titulo}</strong>}
      <div>{children}</div>
    </div>
  );
}

export function Carregando({ texto = "Carregando…" }: { texto?: string }) {
  return (
    <p className="carregando" role="status">
      {texto}
    </p>
  );
}
