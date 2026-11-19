package server

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/service"
)

func capturarLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var saida bytes.Buffer
	log.SetOutput(&saida)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	return &saida
}

func TestRespostaDeErro(t *testing.T) {
	testes := []struct {
		nome   string
		erro   error
		status int
		codigo string
	}{
		{"validação do service", &service.ErroValidacao{Mensagem: "Título inválido."}, 400, "requisicao_invalida"},
		{"cédula inválida", service.ErrCedulaInvalida, 400, "cedula_invalida"},
		{"opção desconhecida", service.ErrOpcaoDesconhecida, 400, "opcao_desconhecida"},
		{"sessão inválida", service.ErrNaoAutenticado, 401, "nao_autenticado"},
		{"sem header de autenticação", errNaoAutenticado, 401, "nao_autenticado"},
		{"credenciais inválidas", service.ErrCredenciaisInvalidas, 401, "credenciais_invalidas"},
		{"acesso negado", errAcessoNegado, 403, "acesso_negado"},
		{"não apto", service.ErrNaoApto, 403, "nao_apto"},
		{"inadimplente", service.ErrInadimplente, 403, "inadimplente"},
		{"não encontrado", service.ErrNaoEncontrado, 404, "nao_encontrado"},
		{"voto duplicado", service.ErrVotoDuplicado, 409, "voto_duplicado"},
		{"votação não aberta", service.ErrVotacaoNaoAberta, 409, "votacao_nao_aberta"},
		{"votação não encerrada", service.ErrVotacaoNaoEncerrada, 409, "votacao_nao_encerrada"},
		{"segundo turno indisponível", service.ErrSegundoTurnoIndisponivel, 409, "segundo_turno_indisponivel"},
		{"muitas tentativas", service.ErrMuitasTentativas, 429, "muitas_tentativas"},
		{"erro embrulhado mantém o código", fmt.Errorf("contexto: %w", service.ErrVotoDuplicado), 409, "voto_duplicado"},
		{"erro desconhecido vira 500", errors.New("conexão recusada"), 500, "erro_interno"},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			resposta := respostaDeErro(tc.erro)
			if resposta.status != tc.status || resposta.codigo != tc.codigo || resposta.mensagem == "" {
				t.Errorf("esperado %d %s, obtido %+v", tc.status, tc.codigo, resposta)
			}
		})
	}

	if m := respostaDeErro(&service.ErroValidacao{Mensagem: "Título inválido."}).mensagem; m != "Título inválido." {
		t.Errorf("a mensagem de validação deveria ir para o usuário, obteve %q", m)
	}
}

func TestErroInternoNaoVazaDetalhes(t *testing.T) {
	saida := capturarLog(t)
	rec := httptest.NewRecorder()
	escreverErro(rec, httptest.NewRequest(http.MethodGet, "/api/x", nil), errors.New("pq: senha do banco errada"))

	if rec.Code != 500 || strings.Contains(rec.Body.String(), "senha do banco") {
		t.Errorf("resposta deveria ser 500 genérica, obteve %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(saida.String(), "senha do banco") {
		t.Errorf("o detalhe deveria ir para o log, log: %q", saida.String())
	}
}

func TestLerJSON(t *testing.T) {
	type destino struct {
		Nome string `json:"nome"`
	}
	testes := []struct {
		nome     string
		corpo    string
		esperado error
	}{
		{"objeto válido", `{"nome":"ana"}`, nil},
		{"objeto com espaços no fim", "{\"nome\":\"ana\"}\n ", nil},
		{"corpo vazio", ``, errCorpoVazio},
		{"JSON malformado", `{"nome":`, errCorpoInvalido},
		{"campo desconhecido", `{"nome":"ana","papel":"admin"}`, errCorpoInvalido},
		{"tipo errado", `{"nome":1}`, errCorpoInvalido},
		{"dois objetos", `{"nome":"a"}{"nome":"b"}`, errCorpoInvalido},
		{"lixo depois do objeto", `{"nome":"a"} x`, errCorpoInvalido},
		{"corpo acima de 64 KiB", `{"nome":"` + strings.Repeat("a", limiteCorpo) + `"}`, errCorpoGrande},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.corpo))
			req.Body = http.MaxBytesReader(rec, req.Body, limiteCorpo)

			var d destino
			if err := lerJSON(req, &d); !errors.Is(err, tc.esperado) {
				t.Errorf("esperado %v, obtido %v", tc.esperado, err)
			}
		})
	}

	t.Run("corpo opcional aceita vazio", func(t *testing.T) {
		var d destino
		if err := lerJSONOpcional(httptest.NewRequest(http.MethodPost, "/", nil), &d); err != nil {
			t.Errorf("esperado nil, obtido %v", err)
		}
		if err := lerJSONOpcional(httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{")), &d); !errors.Is(err, errCorpoInvalido) {
			t.Errorf("JSON inválido continua sendo erro, obtido %v", err)
		}
	})
}

func TestTokenBearer(t *testing.T) {
	testes := []struct {
		nome      string
		cabecalho string
		token     string
		ok        bool
	}{
		{"bearer válido", "Bearer abc", "abc", true},
		{"prefixo em minúsculas", "bearer abc", "abc", true},
		{"sem header", "", "", false},
		{"só o prefixo", "Bearer ", "", false},
		{"prefixo com espaços e sem token", "Bearer    ", "", false},
		{"outro esquema", "Basic abc", "", false},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.cabecalho != "" {
				req.Header.Set("Authorization", tc.cabecalho)
			}
			token, ok := tokenBearer(req)
			if token != tc.token || ok != tc.ok {
				t.Errorf("esperado (%q, %v), obtido (%q, %v)", tc.token, tc.ok, token, ok)
			}
		})
	}
}

func TestRecuperarPanico(t *testing.T) {
	saida := capturarLog(t)
	handler := recuperarPanico(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("falha inesperada com segredo")
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/x", nil))

	if rec.Code != 500 || !strings.Contains(rec.Body.String(), `"codigo":"erro_interno"`) || strings.Contains(rec.Body.String(), "segredo") {
		t.Errorf("esperava 500 genérico, obteve %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(saida.String(), "falha inesperada com segredo") {
		t.Errorf("o panic deveria ir para o log: %q", saida.String())
	}
}

func TestRecuperarPanicoRepassaAbortHandler(t *testing.T) {
	handler := recuperarPanico(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))
	defer func() {
		if p := recover(); p != http.ErrAbortHandler {
			t.Errorf("esperava repassar http.ErrAbortHandler, obteve %v", p)
		}
	}()
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
}

func TestRegistrarAcessoNaoLogaCorpoNemHeaders(t *testing.T) {
	saida := capturarLog(t)
	handler := registrarAcesso(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.WriteHeader(http.StatusOK) // segunda chamada é ignorada pelo net/http e pelo log
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/login?x=1", strings.NewReader(`{"senha":"segredo-123"}`))
	req.Header.Set("Authorization", "Bearer token-secreto")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	linha := saida.String()
	if !strings.Contains(linha, "POST /api/login 418") {
		t.Errorf("log deveria ter método, caminho e status: %q", linha)
	}
	for _, segredo := range []string{"segredo-123", "token-secreto", "x=1"} {
		if strings.Contains(linha, segredo) {
			t.Errorf("log não pode conter %q: %q", segredo, linha)
		}
	}
}

func TestRegistrarAcessoStatusPadrao(t *testing.T) {
	saida := capturarLog(t)
	handler := registrarAcesso(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))
	if !strings.Contains(saida.String(), "GET /health 200") {
		t.Errorf("sem WriteHeader o status é 200: %q", saida.String())
	}
}

func TestIDDaRota(t *testing.T) {
	testes := []struct {
		valor string
		id    int64
		erro  error
	}{
		{"42", 42, nil},
		{"0", 0, errNaoEncontrado},
		{"-1", 0, errNaoEncontrado},
		{"abc", 0, errNaoEncontrado},
		{"99999999999999999999", 0, errNaoEncontrado},
	}
	for _, tc := range testes {
		t.Run(tc.valor, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.SetPathValue("id", tc.valor)
			id, err := idDaRota(req)
			if id != tc.id || !errors.Is(err, tc.erro) {
				t.Errorf("esperado (%d, %v), obtido (%d, %v)", tc.id, tc.erro, id, err)
			}
		})
	}
}

func TestIPDoCliente(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.7:51234"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	if ip := ipDoCliente(req); ip != "203.0.113.7" {
		t.Errorf("esperado o IP da conexão, obtido %q", ip)
	}
	req.RemoteAddr = "sem-porta"
	if ip := ipDoCliente(req); ip != "sem-porta" {
		t.Errorf("endereço sem porta deveria ser usado inteiro, obtido %q", ip)
	}
}
