import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { ApiErro } from "../api/erros";
import { useCarregar } from "./useCarregar";

function promessaControlada<T>() {
  let resolver!: (valor: T) => void;
  let rejeitar!: (erro: unknown) => void;
  const promessa = new Promise<T>((res, rej) => {
    resolver = res;
    rejeitar = rej;
  });
  return { promessa, resolver, rejeitar };
}

describe("useCarregar", () => {
  it("começa carregando e expõe os dados quando a promessa resolve", async () => {
    const { result } = renderHook(() => useCarregar(() => Promise.resolve(["a"])));

    expect(result.current.carregando).toBe(true);
    await waitFor(() => expect(result.current.carregando).toBe(false));
    expect(result.current.dados).toEqual(["a"]);
    expect(result.current.erro).toBe("");
  });

  it("expõe a mensagem do erro quando a promessa falha", async () => {
    const carregar = () => Promise.reject(new ApiErro(404, "nao_encontrado", "Votação não encontrada."));
    const { result } = renderHook(() => useCarregar(carregar));

    await waitFor(() => expect(result.current.erro).toBe("Votação não encontrada."));
    expect(result.current.dados).toBeNull();
  });

  it("ignora a resposta de uma carga antiga quando a função muda", async () => {
    const antiga = promessaControlada<string>();
    const nova = promessaControlada<string>();
    const { result, rerender } = renderHook(({ carregar }) => useCarregar(carregar), {
      initialProps: { carregar: () => antiga.promessa },
    });

    rerender({ carregar: () => nova.promessa });
    await act(async () => nova.resolver("nova"));
    await act(async () => antiga.resolver("antiga"));

    expect(result.current.dados).toBe("nova");
  });

  it("ignora falha que chega depois de desmontar", async () => {
    const pendente = promessaControlada<string>();
    const { result, unmount } = renderHook(() => useCarregar(() => pendente.promessa));

    unmount();
    await act(async () => pendente.rejeitar(new Error("tarde demais")));

    expect(result.current.erro).toBe("");
  });

  it("atualizarDados substitui os dados após uma ação", async () => {
    const { result } = renderHook(() => useCarregar(() => Promise.resolve(1)));
    await waitFor(() => expect(result.current.dados).toBe(1));

    act(() => result.current.atualizarDados(2));

    expect(result.current.dados).toBe(2);
  });
});
