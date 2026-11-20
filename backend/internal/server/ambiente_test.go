package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/seed"
	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/service"
	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store/storetest"
)

type relogioFalso struct {
	mu    sync.Mutex
	agora time.Time
}

func (r *relogioFalso) Agora() time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.agora
}

func (r *relogioFalso) Avancar(d time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agora = r.agora.Add(d)
}

// ambiente é a API real (rotas + service + banco de teste) já com os dados do seed de demonstração.
type ambiente struct {
	t       *testing.T
	handler http.Handler
	db      *sql.DB
	relogio *relogioFalso
	tokens  map[string]string // token por login
}

func novoAmbiente(t *testing.T) *ambiente {
	t.Helper()
	db := storetest.NovoBanco(t)
	relogio := &relogioFalso{agora: time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)}
	svc := service.Novo(db, service.Config{Relogio: relogio.Agora, IteracoesSenha: 1000})
	if _, err := seed.Executar(t.Context(), svc); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return &ambiente{t: t, handler: NewRouter(db, svc), db: db, relogio: relogio, tokens: map[string]string{}}
}

// requisitar envia a requisição ao router; corpo string vai como está, outro valor vira JSON.
func (a *ambiente) requisitar(metodo, rota, token string, corpo any) *httptest.ResponseRecorder {
	a.t.Helper()
	var leitor io.Reader
	switch c := corpo.(type) {
	case nil:
	case string:
		leitor = strings.NewReader(c)
	default:
		b, err := json.Marshal(c)
		if err != nil {
			a.t.Fatal(err)
		}
		leitor = bytes.NewReader(b)
	}
	req := httptest.NewRequest(metodo, rota, leitor)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	a.handler.ServeHTTP(rec, req)
	return rec
}

// token faz login (uma vez por usuário, cada um de um IP diferente) e devolve o token.
func (a *ambiente) token(login string) string {
	a.t.Helper()
	if token, ok := a.tokens[login]; ok {
		return token
	}
	senha := "consensu"
	if login == "admin" {
		senha = "admin123"
	}
	corpo := fmt.Sprintf(`{"login":%q,"senha":%q}`, login, senha)
	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(corpo))
	req.RemoteAddr = fmt.Sprintf("10.0.0.%d:5000", len(a.tokens)+1)
	rec := httptest.NewRecorder()
	a.handler.ServeHTTP(rec, req)

	resposta := decodificar[respostaLogin](a.t, rec, http.StatusOK)
	a.tokens[login] = resposta.Token
	return resposta.Token
}

// votacaoPorTitulo busca o id de uma votação do seed pelo título.
func (a *ambiente) votacaoPorTitulo(titulo string) int64 {
	a.t.Helper()
	lista := decodificar[[]votacaoResumoDTO](a.t, a.requisitar(http.MethodGet, "/api/votacoes", a.token("admin"), nil), http.StatusOK)
	for _, v := range lista {
		if v.Titulo == titulo {
			return v.ID
		}
	}
	a.t.Fatalf("votação %q não encontrada", titulo)
	return 0
}

func rota(formato string, id int64) string {
	return fmt.Sprintf(formato, id)
}

// decodificar confere o status e decodifica o corpo JSON.
func decodificar[T any](t *testing.T, rec *httptest.ResponseRecorder, status int) T {
	t.Helper()
	var valor T
	if rec.Code != status {
		t.Fatalf("status: esperado %d, obtido %d, corpo %s", status, rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type: esperado application/json, obtido %q", ct)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &valor); err != nil {
		t.Fatalf("decodificando %s: %v", rec.Body.String(), err)
	}
	return valor
}

// conferirErro confere status e código do corpo de erro {erro, codigo}.
func conferirErro(t *testing.T, rec *httptest.ResponseRecorder, status int, codigo string) {
	t.Helper()
	corpo := decodificar[corpoErro](t, rec, status)
	if corpo.Codigo != codigo || corpo.Erro == "" {
		t.Errorf("erro: esperado código %q com mensagem, obtido %+v", codigo, corpo)
	}
}

// chaves devolve os campos de um objeto JSON em ordem alfabética.
func chaves(t *testing.T, corpo []byte) []string {
	t.Helper()
	var objeto map[string]json.RawMessage
	if err := json.Unmarshal(corpo, &objeto); err != nil {
		t.Fatalf("esperava um objeto JSON: %v (%s)", err, corpo)
	}
	nomes := make([]string, 0, len(objeto))
	for nome := range objeto {
		nomes = append(nomes, nome)
	}
	slices.Sort(nomes)
	return nomes
}

func conferirChaves(t *testing.T, corpo []byte, esperadas ...string) {
	t.Helper()
	slices.Sort(esperadas)
	if obtidas := chaves(t, corpo); !slices.Equal(obtidas, esperadas) {
		t.Errorf("campos do JSON:\n esperado %v\n obtido   %v", esperadas, obtidas)
	}
}
