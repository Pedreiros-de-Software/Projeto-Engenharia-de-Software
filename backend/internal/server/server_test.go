package server

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type bancoFalso struct{ err error }

func (b bancoFalso) Ping() error { return b.err }

func TestHealthHandler(t *testing.T) {
	testes := []struct {
		nome           string
		metodo         string
		banco          pinger
		codigoEsperado int
		corpoEsperado  string
	}{
		{"GET com banco disponível", http.MethodGet, bancoFalso{}, http.StatusOK, `{"status":"ok"}`},
		{"GET com banco indisponível", http.MethodGet, bancoFalso{errors.New("sem conexao")}, http.StatusServiceUnavailable, `{"status":"database unavailable"}`},
		{"POST não é permitido", http.MethodPost, bancoFalso{}, http.StatusMethodNotAllowed, "method not allowed"},
		{"HEAD não é permitido", http.MethodHead, bancoFalso{}, http.StatusMethodNotAllowed, "method not allowed"},
	}

	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			rec := httptest.NewRecorder()
			healthHandler(tc.banco)(rec, httptest.NewRequest(tc.metodo, "/health", nil))

			if rec.Code != tc.codigoEsperado {
				t.Errorf("status: esperado %d, obtido %d", tc.codigoEsperado, rec.Code)
			}
			if corpo := strings.TrimSpace(rec.Body.String()); corpo != tc.corpoEsperado {
				t.Errorf("corpo: esperado %q, obtido %q", tc.corpoEsperado, corpo)
			}
			if tc.codigoEsperado == http.StatusMethodNotAllowed {
				if allow := rec.Header().Get("Allow"); allow != http.MethodGet {
					t.Errorf("header Allow: esperado %q, obtido %q", http.MethodGet, allow)
				}
			} else if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type: esperado application/json, obtido %q", ct)
			}
		})
	}
}

// escritorComFalha simula um cliente que desconectou: toda escrita falha.
type escritorComFalha struct {
	header http.Header
	codigo int
}

func (e *escritorComFalha) Header() http.Header       { return e.header }
func (e *escritorComFalha) WriteHeader(codigo int)    { e.codigo = codigo }
func (e *escritorComFalha) Write([]byte) (int, error) { return 0, errors.New("conexao fechada") }

func TestHealthHandlerFalhaAoEscreverResposta(t *testing.T) {
	var saidaLog bytes.Buffer
	log.SetOutput(&saidaLog)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	w := &escritorComFalha{header: http.Header{}}
	healthHandler(bancoFalso{})(w, httptest.NewRequest(http.MethodGet, "/health", nil))

	if w.codigo != http.StatusOK {
		t.Errorf("status: esperado %d, obtido %d", http.StatusOK, w.codigo)
	}
	if !strings.Contains(saidaLog.String(), "encoding health response") {
		t.Errorf("esperava o erro de escrita no log, obteve %q", saidaLog.String())
	}
}

func TestNewRouter(t *testing.T) {
	// Nenhuma das requisições chega ao banco, então ele pode ser nil.
	router := NewRouter(nil, nil)

	testes := []struct {
		nome           string
		metodo, rota   string
		codigoEsperado int
	}{
		{"rota /health registrada", http.MethodPost, "/health", http.StatusMethodNotAllowed},
		{"rota desconhecida", http.MethodGet, "/nao-existe", http.StatusNotFound},
	}

	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(tc.metodo, tc.rota, nil))
			if rec.Code != tc.codigoEsperado {
				t.Errorf("status: esperado %d, obtido %d", tc.codigoEsperado, rec.Code)
			}
		})
	}
}

func TestRotasDesconhecidasDaAPIRespondemJSON(t *testing.T) {
	router := NewRouter(nil, nil)

	testes := []struct {
		nome         string
		metodo, rota string
	}{
		{"rota inexistente", http.MethodGet, "/api/nao-existe"},
		{"raiz da API", http.MethodGet, "/api/"},
		{"método errado em rota de ação", http.MethodGet, "/api/votacoes/1/encerrar"},
		{"método sem rota", http.MethodDelete, "/api/votacoes/1"},
		{"login só aceita POST", http.MethodGet, "/api/login"},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(tc.metodo, tc.rota, nil))
			conferirErro(t, rec, http.StatusNotFound, "nao_encontrado")
		})
	}
}
