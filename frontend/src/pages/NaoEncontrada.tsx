import { Link } from "react-router-dom";

function NaoEncontrada() {
  return (
    <div className="pagina">
      <h1>Página não encontrada</h1>
      <p>
        O endereço acessado não existe. <Link to="/">Voltar ao painel</Link>
      </p>
    </div>
  );
}

export default NaoEncontrada;
