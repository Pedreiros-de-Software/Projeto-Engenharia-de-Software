package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	tituloPautaEncerrada = "Pauta #01 — Troca do sistema de câmeras"
	tituloPautaAberta    = "Pauta #02 — Aprovação do orçamento para reforma da portaria (R$ 45.000,00)"
	tituloEleicao        = "Eleição Diretoria DCE 2026"
	tituloConsulta       = "Orçamento participativo — Laboratório de Inovação"
)

// rotasProtegidas são todas as rotas que exigem login; as de admin também exigem o papel.
var rotasProtegidas = []struct {
	metodo, caminho string
	somenteAdmin    bool
}{
	{http.MethodPost, "/api/logout", false},
	{http.MethodGet, "/api/me", false},
	{http.MethodGet, "/api/votacoes", false},
	{http.MethodGet, "/api/votacoes/1", false},
	{http.MethodPost, "/api/votacoes/1/votos", false},
	{http.MethodPost, "/api/votacoes", true},
	{http.MethodPost, "/api/votacoes/1/encerrar", true},
	{http.MethodPost, "/api/votacoes/1/segundo-turno", true},
	{http.MethodGet, "/api/unidades", true},
	{http.MethodPatch, "/api/unidades/1", true},
	{http.MethodGet, "/api/usuarios", true},
}

func TestRotasExigemLogin(t *testing.T) {
	a := novoAmbiente(t)
	for _, r := range rotasProtegidas {
		for _, token := range []string{"", "token-inventado"} {
			t.Run(fmt.Sprintf("%s %s com token %q", r.metodo, r.caminho, token), func(t *testing.T) {
				conferirErro(t, a.requisitar(r.metodo, r.caminho, token, `{}`), http.StatusUnauthorized, "nao_autenticado")
			})
		}
	}
}

func TestRotasDeAdminRecusamEleitor(t *testing.T) {
	a := novoAmbiente(t)
	for _, r := range rotasProtegidas {
		if !r.somenteAdmin {
			continue
		}
		t.Run(r.metodo+" "+r.caminho, func(t *testing.T) {
			conferirErro(t, a.requisitar(r.metodo, r.caminho, a.token("maria"), `{}`), http.StatusForbidden, "acesso_negado")
		})
	}
}

func TestLogin(t *testing.T) {
	a := novoAmbiente(t)

	t.Run("credenciais corretas devolvem token e usuário", func(t *testing.T) {
		rec := a.requisitar(http.MethodPost, "/api/login", "", `{"login":"maria","senha":"consensu"}`)
		resposta := decodificar[respostaLogin](t, rec, http.StatusOK)
		conferirChaves(t, rec.Body.Bytes(), "token", "usuario")
		if resposta.Token == "" || resposta.Usuario.Login != "maria" || resposta.Usuario.Nome != "Maria Souza" || resposta.Usuario.Papel != "eleitor" {
			t.Errorf("resposta inesperada: %+v", resposta)
		}
		if rec.Header().Get("Cache-Control") != "no-store" {
			t.Error("respostas da API não devem ser guardadas em cache")
		}
	})

	testes := []struct {
		nome   string
		corpo  string
		status int
		codigo string
	}{
		{"senha errada", `{"login":"maria","senha":"errada"}`, http.StatusUnauthorized, "credenciais_invalidas"},
		{"login com caractere NUL", `{"login":"ma\u0000ria","senha":"consensu"}`, http.StatusUnauthorized, "credenciais_invalidas"},
		{"senha com caractere NUL", `{"login":"maria","senha":"con\u0000sensu"}`, http.StatusUnauthorized, "credenciais_invalidas"},
		{"usuário inexistente", `{"login":"ninguem","senha":"consensu"}`, http.StatusUnauthorized, "credenciais_invalidas"},
		{"campo desconhecido", `{"login":"maria","senha":"consensu","papel":"admin"}`, http.StatusBadRequest, "requisicao_invalida"},
		{"corpo vazio", ``, http.StatusBadRequest, "requisicao_invalida"},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			conferirErro(t, a.requisitar(http.MethodPost, "/api/login", "", tc.corpo), tc.status, tc.codigo)
		})
	}
}

func TestLoginLimitaTentativas(t *testing.T) {
	a := novoAmbiente(t)
	tentar := func(senha string) *httptest.ResponseRecorder {
		return a.requisitar(http.MethodPost, "/api/login", "", fmt.Sprintf(`{"login":"maria","senha":%q}`, senha))
	}
	for i := 0; i < 10; i++ {
		conferirErro(t, tentar("errada"), http.StatusUnauthorized, "credenciais_invalidas")
	}
	conferirErro(t, tentar("consensu"), http.StatusTooManyRequests, "muitas_tentativas")

	a.relogio.Avancar(time.Minute)
	decodificar[respostaLogin](t, tentar("consensu"), http.StatusOK)
}

// Atrás do nginx do docker compose, todo navegador chega à API com o IP do proxy.
func TestLoginAtrasDeProxyReverso(t *testing.T) {
	a := novoAmbiente(t)
	logar := func(login, senha string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(fmt.Sprintf(`{"login":%q,"senha":%q}`, login, senha)))
		req.RemoteAddr = "172.18.0.3:40000"
		rec := httptest.NewRecorder()
		a.handler.ServeHTTP(rec, req)
		return rec
	}

	t.Run("todos os usuários do seed entram pelo mesmo IP", func(t *testing.T) {
		logins := []string{"admin", "maria", "joao", "lucia", "pedro", "rafael", "sofia",
			"ana", "bruno", "carla", "diego", "elisa", "fabio", "gabi", "hugo"}
		for _, login := range logins {
			senha := "consensu"
			if login == "admin" {
				senha = "admin123"
			}
			decodificar[respostaLogin](t, logar(login, senha), http.StatusOK)
		}
	})

	t.Run("falhas de um login não bloqueiam os demais", func(t *testing.T) {
		for range 10 {
			conferirErro(t, logar("naoexiste", "x"), http.StatusUnauthorized, "credenciais_invalidas")
		}
		conferirErro(t, logar("naoexiste", "x"), http.StatusTooManyRequests, "muitas_tentativas")
		decodificar[respostaLogin](t, logar("admin", "admin123"), http.StatusOK)
	})
}

func TestMeELogout(t *testing.T) {
	a := novoAmbiente(t)

	t.Run("condômino vê sua unidade", func(t *testing.T) {
		rec := a.requisitar(http.MethodGet, "/api/me", a.token("maria"), nil)
		me := decodificar[respostaMe](t, rec, http.StatusOK)
		conferirChaves(t, rec.Body.Bytes(), "usuario", "unidade")
		if me.Usuario.Login != "maria" || me.Unidade == nil || me.Unidade.Identificador != "Apto 101" ||
			me.Unidade.FracaoIdeal != 60 || me.Unidade.Percentual != 12 || !me.Unidade.Adimplente ||
			me.Unidade.Responsavel == nil || me.Unidade.Responsavel.Nome != "Maria Souza" {
			t.Errorf("resposta inesperada: %+v %+v", me, me.Unidade)
		}
	})

	t.Run("estudante não tem unidade", func(t *testing.T) {
		rec := a.requisitar(http.MethodGet, "/api/me", a.token("ana"), nil)
		if !strings.Contains(rec.Body.String(), `"unidade":null`) {
			t.Errorf("esperava unidade null: %s", rec.Body.String())
		}
	})

	t.Run("logout encerra a sessão", func(t *testing.T) {
		token := a.token("joao")
		if rec := a.requisitar(http.MethodPost, "/api/logout", token, nil); rec.Code != http.StatusNoContent || rec.Body.Len() != 0 {
			t.Fatalf("logout: esperado 204 sem corpo, obtido %d %s", rec.Code, rec.Body.String())
		}
		conferirErro(t, a.requisitar(http.MethodGet, "/api/me", token, nil), http.StatusUnauthorized, "nao_autenticado")
	})

	t.Run("sessão expira em 12 horas", func(t *testing.T) {
		token := a.token("lucia")
		a.relogio.Avancar(12 * time.Hour)
		conferirErro(t, a.requisitar(http.MethodGet, "/api/me", token, nil), http.StatusUnauthorized, "nao_autenticado")
	})
}

func TestUsuariosEUnidades(t *testing.T) {
	a := novoAmbiente(t)
	admin := a.token("admin")

	t.Run("lista de usuários", func(t *testing.T) {
		rec := a.requisitar(http.MethodGet, "/api/usuarios", admin, nil)
		usuarios := decodificar[[]usuarioDTO](t, rec, http.StatusOK)
		if len(usuarios) != 15 || usuarios[0].Login != "admin" || usuarios[0].Papel != "admin" {
			t.Errorf("lista inesperada: %+v", usuarios)
		}
		if strings.Contains(rec.Body.String(), "senha") || strings.Contains(rec.Body.String(), "pbkdf2") {
			t.Error("a lista de usuários não pode expor senha nem hash")
		}
	})

	unidades := decodificar[[]unidadeDTO](t, a.requisitar(http.MethodGet, "/api/unidades", admin, nil), http.StatusOK)
	if len(unidades) != 6 || unidades[5].Identificador != "Apto 402" || unidades[5].Adimplente || unidades[5].Percentual != 24 {
		t.Fatalf("unidades inesperadas: %+v", unidades)
	}
	apto402 := unidades[5].ID

	t.Run("marcar adimplente", func(t *testing.T) {
		rec := a.requisitar(http.MethodPatch, rota("/api/unidades/%d", apto402), admin, `{"adimplente":true}`)
		u := decodificar[unidadeDTO](t, rec, http.StatusOK)
		conferirChaves(t, rec.Body.Bytes(), "id", "identificador", "fracao_ideal", "percentual", "adimplente", "responsavel")
		if !u.Adimplente || u.ID != apto402 || u.Responsavel == nil || u.Responsavel.Nome != "Sofia Prado" {
			t.Errorf("unidade inesperada: %+v", u)
		}
	})

	t.Run("adimplente ausente", func(t *testing.T) {
		conferirErro(t, a.requisitar(http.MethodPatch, rota("/api/unidades/%d", apto402), admin, `{}`), http.StatusBadRequest, "requisicao_invalida")
	})

	t.Run("unidade inexistente", func(t *testing.T) {
		conferirErro(t, a.requisitar(http.MethodPatch, "/api/unidades/9999", admin, `{"adimplente":true}`), http.StatusNotFound, "nao_encontrado")
	})
}

func TestHealthNaoExigeLogin(t *testing.T) {
	a := novoAmbiente(t)
	rec := a.requisitar(http.MethodGet, "/health", "", nil)
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"status":"ok"}` {
		t.Errorf("health: %d %s", rec.Code, rec.Body.String())
	}
}
