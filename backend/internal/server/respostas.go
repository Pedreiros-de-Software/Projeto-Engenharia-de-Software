package server

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/service"
)

// limiteCorpo é o tamanho máximo aceito no corpo de uma requisição (64 KiB).
const limiteCorpo = 64 << 10

// erroAPI é uma resposta de erro pronta: status HTTP, código estável e mensagem para o usuário.
type erroAPI struct {
	status   int
	codigo   string
	mensagem string
}

func (e erroAPI) Error() string { return e.codigo }

type corpoErro struct {
	Erro   string `json:"erro"`
	Codigo string `json:"codigo"`
}

var (
	errNaoAutenticado = erroAPI{http.StatusUnauthorized, "nao_autenticado", "Faça login para continuar."}
	errAcessoNegado   = erroAPI{http.StatusForbidden, "acesso_negado", "Acesso restrito à administração."}
	errNaoEncontrado  = erroAPI{http.StatusNotFound, "nao_encontrado", "Recurso não encontrado."}
	errInterno        = erroAPI{http.StatusInternalServerError, "erro_interno", "Erro interno. Tente novamente mais tarde."}
	errCorpoVazio     = requisicaoInvalida("O corpo da requisição é obrigatório.")
	errCorpoGrande    = requisicaoInvalida("O corpo da requisição excede 64 KiB.")
	errCorpoInvalido  = requisicaoInvalida("Corpo da requisição inválido: envie um único objeto JSON com os campos e tipos esperados.")
)

func requisicaoInvalida(mensagem string) erroAPI {
	return erroAPI{http.StatusBadRequest, "requisicao_invalida", mensagem}
}

// errosDoServico associa cada erro sentinela do service à resposta do contrato da API.
var errosDoServico = []struct {
	erro     error
	resposta erroAPI
}{
	{service.ErrNaoAutenticado, erroAPI{http.StatusUnauthorized, "nao_autenticado", "Sessão inválida ou expirada. Faça login novamente."}},
	{service.ErrCredenciaisInvalidas, erroAPI{http.StatusUnauthorized, "credenciais_invalidas", "Usuário ou senha inválidos."}},
	{service.ErrMuitasTentativas, erroAPI{http.StatusTooManyRequests, "muitas_tentativas", "Muitas tentativas de login. Aguarde um minuto e tente de novo."}},
	{service.ErrNaoEncontrado, errNaoEncontrado},
	{service.ErrNaoApto, erroAPI{http.StatusForbidden, "nao_apto", "Você não está habilitado a votar nesta votação."}},
	{service.ErrInadimplente, erroAPI{http.StatusForbidden, "inadimplente", service.MotivoInadimplente}},
	{service.ErrVotoDuplicado, erroAPI{http.StatusConflict, "voto_duplicado", "Você já votou nesta votação."}},
	{service.ErrVotacaoNaoAberta, erroAPI{http.StatusConflict, "votacao_nao_aberta", "Esta votação não está aberta."}},
	{service.ErrVotacaoNaoEncerrada, erroAPI{http.StatusConflict, "votacao_nao_encerrada", "O resultado só é divulgado depois do encerramento."}},
	{service.ErrSegundoTurnoIndisponivel, erroAPI{http.StatusConflict, "segundo_turno_indisponivel",
		"O 2º turno só pode ser aberto uma vez, para eleição de 1º turno encerrada sem maioria absoluta e com as duas chapas definidas."}},
	{service.ErrCedulaInvalida, erroAPI{http.StatusBadRequest, "cedula_invalida", "A cédula deve ordenar todos os projetos, cada um uma única vez."}},
	{service.ErrOpcaoDesconhecida, erroAPI{http.StatusBadRequest, "opcao_desconhecida", "Opção desconhecida para esta votação."}},
}

// respostaDeErro escolhe a resposta para um erro; o que não for reconhecido vira 500 genérico.
func respostaDeErro(err error) erroAPI {
	var pronta erroAPI
	if errors.As(err, &pronta) {
		return pronta
	}
	var validacao *service.ErroValidacao
	if errors.As(err, &validacao) {
		return requisicaoInvalida(validacao.Mensagem)
	}
	for _, e := range errosDoServico {
		if errors.Is(err, e.erro) {
			return e.resposta
		}
	}
	return errInterno
}

// escreverErro responde no formato {erro, codigo}; o detalhe de erros internos vai só para o log.
func escreverErro(w http.ResponseWriter, r *http.Request, err error) {
	resposta := respostaDeErro(err)
	if resposta.status == http.StatusInternalServerError {
		log.Printf("erro interno em %s %s: %v", r.Method, r.URL.Path, err)
	}
	escreverJSON(w, resposta.status, corpoErro{Erro: resposta.mensagem, Codigo: resposta.codigo})
}

func escreverJSON(w http.ResponseWriter, status int, corpo any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(corpo); err != nil {
		log.Printf("escrevendo resposta JSON: %v", err)
	}
}

// lerJSON decodifica o corpo exigindo um único objeto JSON, sem campos desconhecidos.
func lerJSON(r *http.Request, destino any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destino); err != nil {
		var grande *http.MaxBytesError
		switch {
		case errors.Is(err, io.EOF):
			return errCorpoVazio
		case errors.As(err, &grande):
			return errCorpoGrande
		default:
			return errCorpoInvalido
		}
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return errCorpoInvalido // há algo depois do objeto
	}
	return nil
}

// lerJSONOpcional é como lerJSON, mas aceita corpo vazio (destino fica com os valores zero).
func lerJSONOpcional(r *http.Request, destino any) error {
	if err := lerJSON(r, destino); !errors.Is(err, errCorpoVazio) {
		return err
	}
	return nil
}
