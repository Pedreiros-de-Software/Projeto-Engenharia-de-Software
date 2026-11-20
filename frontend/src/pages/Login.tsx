import { useState, type FormEvent } from "react";
import { Navigate, useLocation, useNavigate } from "react-router-dom";

import { mensagemDeErro } from "../api/erros";
import { useAuth } from "../auth/useAuth";
import { Alerta } from "../components/Alerta";
import "../styles/styles-pages/Login.css";

function Login() {
  const { usuario, entrar } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const destino = (location.state as { de?: string } | null)?.de ?? "/";
  const [nomeDeUsuario, setNomeDeUsuario] = useState("");
  const [senha, setSenha] = useState("");
  const [erro, setErro] = useState("");
  const [enviando, setEnviando] = useState(false);

  if (usuario) {
    return <Navigate to={destino} replace />;
  }

  async function aoEnviar(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault();
    setErro("");
    setEnviando(true);
    try {
      await entrar(nomeDeUsuario.trim(), senha);
      navigate(destino, { replace: true });
    } catch (e) {
      setErro(mensagemDeErro(e));
      setEnviando(false);
    }
  }

  return (
    <main className="login">
      <form className="login-cartao" onSubmit={aoEnviar}>
        <h1>Consensu</h1>
        <p className="login-subtitulo">Entre para votar e acompanhar as votações.</p>

        <label htmlFor="login-usuario">Usuário</label>
        <input
          id="login-usuario"
          className="campo"
          autoComplete="username"
          value={nomeDeUsuario}
          onChange={(e) => setNomeDeUsuario(e.target.value)}
          required
        />

        <label htmlFor="login-senha">Senha</label>
        <input
          id="login-senha"
          className="campo"
          type="password"
          autoComplete="current-password"
          value={senha}
          onChange={(e) => setSenha(e.target.value)}
          required
        />

        {erro && <Alerta tipo="erro">{erro}</Alerta>}

        <button type="submit" className="botao botao-largo" disabled={enviando}>
          {enviando ? "Entrando…" : "Entrar"}
        </button>
      </form>
    </main>
  );
}

export default Login;
