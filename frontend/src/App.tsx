import { BrowserRouter, Route, Routes } from "react-router-dom";

import { AuthProvider } from "./auth/AuthContext";
import { Layout } from "./components/Layout";
import { RotaAdmin, RotaAutenticada } from "./components/RotasProtegidas";
import NovaVotacao from "./pages/admin/NovaVotacao";
import Unidades from "./pages/admin/Unidades";
import Home from "./pages/Home";
import Login from "./pages/Login";
import NaoEncontrada from "./pages/NaoEncontrada";
import Resultado from "./pages/Resultado";
import Votar from "./pages/Votar";

/** Mapa de rotas da aplicação; separado de `App` para os testes usarem um MemoryRouter. */
export function Rotas() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route element={<Layout />}>
        <Route path="/votacoes/:id/resultado" element={<Resultado />} />
        <Route element={<RotaAutenticada />}>
          <Route path="/" element={<Home />} />
          <Route path="/votacoes/:id" element={<Votar />} />
          <Route element={<RotaAdmin />}>
            <Route path="/admin/nova-votacao" element={<NovaVotacao />} />
            <Route path="/admin/unidades" element={<Unidades />} />
          </Route>
        </Route>
        <Route path="*" element={<NaoEncontrada />} />
      </Route>
    </Routes>
  );
}

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Rotas />
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
