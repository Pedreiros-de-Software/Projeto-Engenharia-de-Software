import type { ReactNode } from "react";

/** Contêiner com rolagem horizontal (também pelo teclado) para tabelas mais largas que a tela. */
export function RolagemDaTabela({ rotulo, children }: { rotulo: string; children: ReactNode }) {
  return (
    <div className="tabela-rolagem" role="region" aria-label={rotulo} tabIndex={0}>
      {children}
    </div>
  );
}
