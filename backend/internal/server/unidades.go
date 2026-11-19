package server

import "net/http"

func (a *api) listarUnidades(w http.ResponseWriter, r *http.Request) {
	unidades, err := a.svc.ListarUnidades(r.Context())
	if err != nil {
		escreverErro(w, r, err)
		return
	}
	dtos := make([]unidadeDTO, len(unidades))
	for i, u := range unidades {
		dtos[i] = paraUnidadeDTO(u)
	}
	escreverJSON(w, http.StatusOK, dtos)
}

func (a *api) alterarUnidade(w http.ResponseWriter, r *http.Request) {
	id, err := idDaRota(r)
	if err != nil {
		escreverErro(w, r, err)
		return
	}
	var pedido pedidoAlterarUnidade
	if err := lerJSON(r, &pedido); err != nil {
		escreverErro(w, r, err)
		return
	}
	if pedido.Adimplente == nil {
		escreverErro(w, r, requisicaoInvalida("Informe adimplente: true ou false."))
		return
	}

	unidade, err := a.svc.AlterarAdimplencia(r.Context(), id, *pedido.Adimplente)
	if err != nil {
		escreverErro(w, r, err)
		return
	}
	escreverJSON(w, http.StatusOK, paraUnidadeDTO(unidade))
}
