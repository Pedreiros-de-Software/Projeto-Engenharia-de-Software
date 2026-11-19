package server

import (
	"net/http"
	"strconv"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/service"
)

const mensagemVotoRegistrado = "Voto registrado com sucesso."

// idDaRota lê o {id} do caminho; id que não é inteiro positivo não existe.
func idDaRota(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errNaoEncontrado
	}
	return id, nil
}

func (a *api) listarVotacoes(w http.ResponseWriter, r *http.Request) {
	votacoes, err := a.svc.ListarVotacoes(r.Context(), usuarioDe(r).ID)
	if err != nil {
		escreverErro(w, r, err)
		return
	}
	dtos := make([]votacaoResumoDTO, len(votacoes))
	for i, v := range votacoes {
		dtos[i] = paraResumoDTO(v)
	}
	escreverJSON(w, http.StatusOK, dtos)
}

func (a *api) detalharVotacao(w http.ResponseWriter, r *http.Request) {
	id, err := idDaRota(r)
	if err != nil {
		escreverErro(w, r, err)
		return
	}
	a.responderDetalhe(w, r, http.StatusOK, id)
}

// responderDetalhe responde com a votação vista pelo usuário da requisição.
func (a *api) responderDetalhe(w http.ResponseWriter, r *http.Request, status int, votacaoID int64) {
	detalhe, err := a.svc.Detalhar(r.Context(), usuarioDe(r).ID, votacaoID)
	if err != nil {
		escreverErro(w, r, err)
		return
	}
	escreverJSON(w, status, paraDetalheDTO(detalhe))
}

func (a *api) criarVotacao(w http.ResponseWriter, r *http.Request) {
	var pedido pedidoNovaVotacao
	if err := lerJSON(r, &pedido); err != nil {
		escreverErro(w, r, err)
		return
	}

	nova := service.NovaVotacao{
		Tipo:              pedido.Tipo,
		Titulo:            pedido.Titulo,
		Descricao:         pedido.Descricao,
		Inicio:            pedido.Inicio,
		Fim:               pedido.Fim,
		Opcoes:            make([]service.NovaOpcao, len(pedido.Opcoes)),
		OrcamentoCentavos: pedido.OrcamentoCentavos,
		Aptos:             pedido.Aptos,
	}
	for i, o := range pedido.Opcoes {
		nova.Opcoes[i] = service.NovaOpcao{Nome: o.Nome, Descricao: o.Descricao, CustoCentavos: o.CustoCentavos}
	}

	id, err := a.svc.CriarVotacao(r.Context(), nova)
	if err != nil {
		escreverErro(w, r, err)
		return
	}
	a.responderDetalhe(w, r, http.StatusCreated, id)
}

func (a *api) encerrarVotacao(w http.ResponseWriter, r *http.Request) {
	id, err := idDaRota(r)
	if err != nil {
		escreverErro(w, r, err)
		return
	}
	if err := a.svc.Encerrar(r.Context(), id); err != nil {
		escreverErro(w, r, err)
		return
	}
	detalhe, err := a.svc.Detalhar(r.Context(), usuarioDe(r).ID, id)
	if err != nil {
		escreverErro(w, r, err)
		return
	}
	escreverJSON(w, http.StatusOK, paraResumoDTO(detalhe.VotacaoResumo))
}

// votar responde só com uma mensagem fixa: a escolha nunca é ecoada.
func (a *api) votar(w http.ResponseWriter, r *http.Request) {
	id, err := idDaRota(r)
	if err != nil {
		escreverErro(w, r, err)
		return
	}
	var pedido pedidoVoto
	if err := lerJSON(r, &pedido); err != nil {
		escreverErro(w, r, err)
		return
	}

	voto := service.Voto{Escolha: pedido.Escolha, OpcaoID: pedido.OpcaoID, Preferencias: pedido.Preferencias}
	if err := a.svc.Votar(r.Context(), usuarioDe(r).ID, id, voto); err != nil {
		escreverErro(w, r, err)
		return
	}
	escreverJSON(w, http.StatusCreated, respostaVoto{Mensagem: mensagemVotoRegistrado})
}

// resultado é público: o boletim de uma votação encerrada não depende de login.
func (a *api) resultado(w http.ResponseWriter, r *http.Request) {
	id, err := idDaRota(r)
	if err != nil {
		escreverErro(w, r, err)
		return
	}
	resultado, err := a.svc.Resultado(r.Context(), id)
	if err != nil {
		escreverErro(w, r, err)
		return
	}
	escreverJSON(w, http.StatusOK, paraResultadoDTO(resultado))
}

func (a *api) abrirSegundoTurno(w http.ResponseWriter, r *http.Request) {
	id, err := idDaRota(r)
	if err != nil {
		escreverErro(w, r, err)
		return
	}
	var pedido pedidoSegundoTurno
	if err := lerJSONOpcional(r, &pedido); err != nil {
		escreverErro(w, r, err)
		return
	}

	novoID, err := a.svc.AbrirSegundoTurno(r.Context(), id, pedido.Fim)
	if err != nil {
		escreverErro(w, r, err)
		return
	}
	a.responderDetalhe(w, r, http.StatusCreated, novoID)
}
