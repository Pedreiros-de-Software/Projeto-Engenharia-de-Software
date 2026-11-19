import { Link, Navigate, Outlet, useLocation } from "react-router-dom";

import { useAuth } from "../auth/useAuth";
import { Alerta, Carregando } from "./Alerta";

/** Só renderiza as rotas filhas para usuário logado; senão redireciona para /login. */
export function RotaAutenticada() {
  const { usuario, carregando } = useAuth();
  const location = useLocation();

  if (carregando) {
    return <Carregando texto="Verificando sua sessão…" />;
  }
  if (!usuario) {
    return <Navigate to="/login" replace state={{ de: location.pathname }} />;
  }
  return <Outlet />;
}

/** Só renderiza as rotas filhas para administradores (o servidor também bloqueia). */
export function RotaAdmin() {
  const { usuario } = useAuth();

  if (usuario?.papel !== "admin") {
    return (
      <Alerta tipo="erro" titulo="Acesso restrito">
        Esta página é exclusiva da administração. <Link to="/">Voltar ao painel</Link>
      </Alerta>
    );
  }
  return <Outlet />;
}
