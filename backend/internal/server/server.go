package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/service"
)

// api reúne os handlers HTTP; cada um traduz a requisição para um caso de uso do service.
type api struct {
	svc *service.Service
}

// NewRouter monta as rotas da API com seus middlewares; db é usado apenas pelo health check.
func NewRouter(db *sql.DB, svc *service.Service) http.Handler {
	a := &api{svc: svc}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler(db))
	// Rota desconhecida ou método errado sob /api respondem no formato de erro do contrato.
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) { escreverErro(w, r, errNaoEncontrado) })

	mux.HandleFunc("POST /api/login", a.login)
	mux.Handle("POST /api/logout", a.autenticado(a.logout))
	mux.Handle("GET /api/me", a.autenticado(a.me))

	mux.Handle("GET /api/votacoes", a.autenticado(a.listarVotacoes))
	mux.Handle("POST /api/votacoes", a.somenteAdmin(a.criarVotacao))
	mux.Handle("GET /api/votacoes/{id}", a.autenticado(a.detalharVotacao))
	mux.Handle("POST /api/votacoes/{id}/encerrar", a.somenteAdmin(a.encerrarVotacao))
	mux.Handle("POST /api/votacoes/{id}/votos", a.autenticado(a.votar))
	mux.HandleFunc("GET /api/votacoes/{id}/resultado", a.resultado)
	mux.Handle("POST /api/votacoes/{id}/segundo-turno", a.somenteAdmin(a.abrirSegundoTurno))

	mux.Handle("GET /api/unidades", a.somenteAdmin(a.listarUnidades))
	mux.Handle("PATCH /api/unidades/{id}", a.somenteAdmin(a.alterarUnidade))
	mux.Handle("GET /api/usuarios", a.somenteAdmin(a.listarUsuarios))

	return registrarAcesso(recuperarPanico(limitarCorpo(mux)))
}

// pinger é a parte do banco usada pelo health check; *sql.DB a satisfaz.
type pinger interface {
	Ping() error
}

func healthHandler(db pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		status := "ok"
		code := http.StatusOK

		if err := db.Ping(); err != nil {
			status = "database unavailable"
			code = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": status}); err != nil {
			log.Printf("encoding health response: %v", err)
		}
	}
}
