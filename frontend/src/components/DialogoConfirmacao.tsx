import { useEffect, useId, useLayoutEffect, useRef, type ReactNode } from "react";

type Props = {
  titulo: string;
  children: ReactNode;
  textoConfirmar?: string;
  onConfirmar: () => void;
  onCancelar: () => void;
};

const FOCAVEIS =
  'button:not([disabled]), a[href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

/** Janela modal de confirmação; Esc ou "Cancelar" fecham sem executar a ação e o foco fica preso nela. */
export function DialogoConfirmacao({
  titulo,
  children,
  textoConfirmar = "Confirmar",
  onConfirmar,
  onCancelar,
}: Props) {
  const idTitulo = useId();
  const idDescricao = useId();
  const refDialogo = useRef<HTMLDivElement>(null);
  const refConfirmar = useRef<HTMLButtonElement>(null);

  useLayoutEffect(() => {
    const anterior = document.activeElement;
    refConfirmar.current?.focus();
    return () => {
      if (anterior instanceof HTMLElement) anterior.focus();
    };
  }, []);

  useEffect(() => {
    function aoTeclar(evento: KeyboardEvent) {
      if (evento.key === "Escape") {
        onCancelar();
        return;
      }
      if (evento.key !== "Tab" || !refDialogo.current) return;
      const focaveis = refDialogo.current.querySelectorAll<HTMLElement>(FOCAVEIS);
      const primeiro = focaveis[0];
      const ultimo = focaveis[focaveis.length - 1];
      const atual = document.activeElement;
      const fora = !refDialogo.current.contains(atual);
      if (evento.shiftKey && (fora || atual === primeiro)) {
        evento.preventDefault();
        ultimo.focus();
      } else if (!evento.shiftKey && (fora || atual === ultimo)) {
        evento.preventDefault();
        primeiro.focus();
      }
    }
    window.addEventListener("keydown", aoTeclar);
    return () => window.removeEventListener("keydown", aoTeclar);
  }, [onCancelar]);

  return (
    <div className="dialogo-fundo">
      <div
        ref={refDialogo}
        className="dialogo"
        role="dialog"
        aria-modal="true"
        aria-labelledby={idTitulo}
        aria-describedby={idDescricao}
      >
        <h2 id={idTitulo}>{titulo}</h2>
        <div id={idDescricao} className="dialogo-corpo">
          {children}
        </div>
        <div className="dialogo-acoes">
          <button type="button" className="botao botao-secundario" onClick={onCancelar}>
            Cancelar
          </button>
          <button ref={refConfirmar} type="button" className="botao" onClick={onConfirmar}>
            {textoConfirmar}
          </button>
        </div>
      </div>
    </div>
  );
}
