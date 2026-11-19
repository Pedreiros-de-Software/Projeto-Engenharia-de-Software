import { Link, NavLink, Outlet, useNavigate } from "react-router-dom";

import { useAuth } from "../auth/useAuth";
import { NOME_PAPEL } from "../utils/rotulos";

export function Layout() {
  const { usuario, carregando, sair } = useAuth();
  const navigate = useNavigate();

  async function aoSair() {
    await sair();
    navigate("/login");
  }

  return (
    <div className="layout">
      <header className="cabecalho">
        <div className="cabecalho-conteudo">
          <Link to="/" className="marca">
            Consensu
          </Link>
          {usuario && (
            <nav className="navegacao" aria-label="Navegação principal">
              <NavLink to="/" end>
                Painel
              </NavLink>
              {usuario.papel === "admin" && (
                <>
                  <NavLink to="/admin/nova-votacao">Nova votação</NavLink>
                  <NavLink to="/admin/unidades">Unidades</NavLink>
                </>
              )}
            </nav>
          )}
          <div className="usuario-atual">
            {usuario && (
              <>
                <span className="usuario-nome">{usuario.nome}</span>
                <span className="usuario-papel">{NOME_PAPEL[usuario.papel]}</span>
                <button type="button" className="botao botao-fantasma" onClick={aoSair}>
                  Sair
                </button>
              </>
            )}
            {!usuario && !carregando && (
              <Link to="/login" className="botao">
                Entrar
              </Link>
            )}
          </div>
        </div>
      </header>
      <main className="conteudo">
        <Outlet />
      </main>
      <footer className="rodape">
        Consensu · votação para condomínios, centros acadêmicos e priorização de projetos
      </footer>
    </div>
  );
}
