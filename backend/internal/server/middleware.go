package server

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/service"
)

// registrarAcesso loga método, caminho, status e duração. Nunca loga corpo, query nem headers,
// para que senha, token e conteúdo de voto não apareçam no log.
func registrarAcesso(proximo http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		gravador := &gravadorDeStatus{ResponseWriter: w, status: http.StatusOK}
		proximo.ServeHTTP(gravador, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, gravador.status, time.Since(inicio).Round(time.Microsecond))
	})
}

// gravadorDeStatus guarda o status enviado ao cliente para o log de acesso.
type gravadorDeStatus struct {
	http.ResponseWriter
	status        int
	statusEnviado bool
}

func (g *gravadorDeStatus) WriteHeader(status int) {
	if !g.statusEnviado {
		g.status = status
		g.statusEnviado = true
	}
	g.ResponseWriter.WriteHeader(status)
}

func (g *gravadorDeStatus) Unwrap() http.ResponseWriter { return g.ResponseWriter }

// recuperarPanico transforma um panic em 500 genérico; o detalhe fica só no log.
func recuperarPanico(proximo http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			p := recover()
			if p == nil {
				return
			}
			if p == http.ErrAbortHandler {
				panic(p) // sinal do net/http para abortar a resposta; não é uma falha
			}
			log.Printf("panic em %s %s: %v\n%s", r.Method, r.URL.Path, p, debug.Stack())
			escreverJSON(w, errInterno.status, corpoErro{Erro: errInterno.mensagem, Codigo: errInterno.codigo})
		}()
		proximo.ServeHTTP(w, r)
	})
}

// limitarCorpo impede que um cliente envie corpos maiores que limiteCorpo.
func limitarCorpo(proximo http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, limiteCorpo)
		proximo.ServeHTTP(w, r)
	})
}

type chaveUsuario struct{}

// autenticado exige "Authorization: Bearer <token>" de uma sessão válida e coloca o usuário no contexto.
func (a *api) autenticado(proximo http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := tokenBearer(r)
		if !ok {
			escreverErro(w, r, errNaoAutenticado)
			return
		}
		usuario, err := a.svc.UsuarioDaSessao(r.Context(), token)
		if err != nil {
			escreverErro(w, r, err)
			return
		}
		proximo(w, r.WithContext(context.WithValue(r.Context(), chaveUsuario{}, usuario)))
	})
}

// somenteAdmin exige uma sessão válida de um usuário com papel admin.
func (a *api) somenteAdmin(proximo http.HandlerFunc) http.Handler {
	return a.autenticado(func(w http.ResponseWriter, r *http.Request) {
		if usuarioDe(r).Papel != service.PapelAdmin {
			escreverErro(w, r, errAcessoNegado)
			return
		}
		proximo(w, r)
	})
}

// usuarioDe devolve o usuário colocado no contexto por autenticado.
func usuarioDe(r *http.Request) service.Usuario {
	usuario, _ := r.Context().Value(chaveUsuario{}).(service.Usuario)
	return usuario
}

func tokenBearer(r *http.Request) (string, bool) {
	const prefixo = "Bearer "
	cabecalho := r.Header.Get("Authorization")
	if len(cabecalho) <= len(prefixo) || !strings.EqualFold(cabecalho[:len(prefixo)], prefixo) {
		return "", false
	}
	token := strings.TrimSpace(cabecalho[len(prefixo):])
	return token, token != ""
}
