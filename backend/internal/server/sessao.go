package server

import (
	"net"
	"net/http"
)

func (a *api) login(w http.ResponseWriter, r *http.Request) {
	var pedido pedidoLogin
	if err := lerJSON(r, &pedido); err != nil {
		escreverErro(w, r, err)
		return
	}
	sessao, err := a.svc.Login(r.Context(), pedido.Login, pedido.Senha, ipDoCliente(r))
	if err != nil {
		escreverErro(w, r, err)
		return
	}
	escreverJSON(w, http.StatusOK, respostaLogin{Token: sessao.Token, Usuario: paraUsuarioDTO(sessao.Usuario)})
}

// ipDoCliente usa o endereço da conexão TCP; headers como X-Forwarded-For podem ser forjados.
func ipDoCliente(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (a *api) logout(w http.ResponseWriter, r *http.Request) {
	token, _ := tokenBearer(r)
	if err := a.svc.Logout(r.Context(), token); err != nil {
		escreverErro(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) me(w http.ResponseWriter, r *http.Request) {
	usuario := usuarioDe(r)
	unidade, err := a.svc.UnidadeDoUsuario(r.Context(), usuario.ID)
	if err != nil {
		escreverErro(w, r, err)
		return
	}
	resposta := respostaMe{Usuario: paraUsuarioDTO(usuario)}
	if unidade != nil {
		dto := paraUnidadeDTO(*unidade)
		resposta.Unidade = &dto
	}
	escreverJSON(w, http.StatusOK, resposta)
}

func (a *api) listarUsuarios(w http.ResponseWriter, r *http.Request) {
	usuarios, err := a.svc.ListarUsuarios(r.Context())
	if err != nil {
		escreverErro(w, r, err)
		return
	}
	dtos := make([]usuarioDTO, len(usuarios))
	for i, u := range usuarios {
		dtos[i] = paraUsuarioDTO(u)
	}
	escreverJSON(w, http.StatusOK, dtos)
}
