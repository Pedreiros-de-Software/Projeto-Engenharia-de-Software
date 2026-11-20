type Props = { rotulo: string; valor: string; detalhe?: string };

/** Número de destaque com rótulo, usado nos resumos de resultado. */
export function Indicador({ rotulo, valor, detalhe }: Props) {
  return (
    <div className="indicador">
      <span className="indicador-rotulo">{rotulo}</span>
      <strong className="indicador-valor">{valor}</strong>
      {detalhe && <span className="indicador-detalhe">{detalhe}</span>}
    </div>
  );
}
