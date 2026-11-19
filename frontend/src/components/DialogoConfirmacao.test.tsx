import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";

import { DialogoConfirmacao } from "./DialogoConfirmacao";

function renderizarDialogo() {
  const onConfirmar = vi.fn();
  const onCancelar = vi.fn();
  render(
    <DialogoConfirmacao titulo="Confirmar voto" textoConfirmar="Votar" onConfirmar={onConfirmar} onCancelar={onCancelar}>
      <p>Seu voto: Aprovar</p>
    </DialogoConfirmacao>,
  );
  return { onConfirmar, onCancelar };
}

function ComGatilho() {
  const [aberto, setAberto] = useState(false);
  return (
    <>
      <a href="/painel">Painel</a>
      <button type="button" onClick={() => setAberto(true)}>
        Encerrar agora
      </button>
      {aberto && (
        <DialogoConfirmacao titulo="Encerrar?" onConfirmar={() => setAberto(false)} onCancelar={() => setAberto(false)}>
          <p>Ninguém mais poderá votar.</p>
        </DialogoConfirmacao>
      )}
    </>
  );
}

describe("DialogoConfirmacao", () => {
  it("é um diálogo modal nomeado pelo título e descrito pelo conteúdo", () => {
    renderizarDialogo();

    const dialogo = screen.getByRole("dialog", { name: "Confirmar voto" });
    expect(dialogo).toHaveAttribute("aria-modal", "true");
    expect(dialogo).toHaveAccessibleDescription("Seu voto: Aprovar");
    expect(screen.getByRole("button", { name: "Votar" })).toHaveFocus();
  });

  it("confirma e cancela pelos botões", async () => {
    const { onConfirmar, onCancelar } = renderizarDialogo();

    await userEvent.click(screen.getByRole("button", { name: "Votar" }));
    await userEvent.click(screen.getByRole("button", { name: "Cancelar" }));

    expect(onConfirmar).toHaveBeenCalledOnce();
    expect(onCancelar).toHaveBeenCalledOnce();
  });

  it("Esc cancela; outras teclas não", async () => {
    const { onCancelar } = renderizarDialogo();

    await userEvent.keyboard("a");
    expect(onCancelar).not.toHaveBeenCalled();
    await userEvent.keyboard("{Escape}");
    expect(onCancelar).toHaveBeenCalledOnce();
  });

  it("Tab e Shift+Tab circulam só entre os botões do diálogo", async () => {
    render(<ComGatilho />);
    await userEvent.click(screen.getByRole("button", { name: "Encerrar agora" }));
    const confirmar = screen.getByRole("button", { name: "Confirmar" });
    const cancelar = screen.getByRole("button", { name: "Cancelar" });
    expect(confirmar).toHaveFocus();

    await userEvent.tab();
    expect(cancelar).toHaveFocus();
    await userEvent.tab();
    expect(confirmar).toHaveFocus();
    await userEvent.tab({ shift: true });
    expect(cancelar).toHaveFocus();
    await userEvent.tab({ shift: true });
    expect(confirmar).toHaveFocus();
  });

  it("foco fora do diálogo volta para dentro no próximo Tab", async () => {
    render(<ComGatilho />);
    await userEvent.click(screen.getByRole("button", { name: "Encerrar agora" }));
    screen.getByRole("link", { name: "Painel" }).focus();

    await userEvent.tab();
    expect(screen.getByRole("button", { name: "Cancelar" })).toHaveFocus();
    screen.getByRole("link", { name: "Painel" }).focus();
    await userEvent.tab({ shift: true });
    expect(screen.getByRole("button", { name: "Confirmar" })).toHaveFocus();
  });

  it.each([
    { caso: "Cancelar", fechar: () => userEvent.click(screen.getByRole("button", { name: "Cancelar" })) },
    { caso: "Esc", fechar: () => userEvent.keyboard("{Escape}") },
    { caso: "Confirmar", fechar: () => userEvent.click(screen.getByRole("button", { name: "Confirmar" })) },
  ])("ao fechar por $caso, o foco volta ao botão que abriu o diálogo", async ({ fechar }) => {
    render(<ComGatilho />);
    const gatilho = screen.getByRole("button", { name: "Encerrar agora" });
    await userEvent.click(gatilho);
    expect(gatilho).not.toHaveFocus();

    await fechar();

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(gatilho).toHaveFocus();
  });
});
