package service

import (
	"errors"
	"fmt"
	"maps"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestHashDeSenha(t *testing.T) {
	hash, err := gerarHashSenha("segredo123", iteracoesDeTeste)
	if err != nil {
		t.Fatal(err)
	}
	partes := strings.Split(hash, "$")
	if len(partes) != 4 || partes[0] != "pbkdf2-sha256" || partes[1] != "1000" {
		t.Fatalf("formato inesperado: %q", hash)
	}
	if outro, _ := gerarHashSenha("segredo123", iteracoesDeTeste); outro == hash {
		t.Error("dois hashes da mesma senha deveriam ter sais diferentes")
	}

	testes := []struct {
		nome    string
		hash    string
		senha   string
		confere bool
	}{
		{"senha correta", hash, "segredo123", true},
		{"senha errada", hash, "segredo124", false},
		{"senha vazia", hash, "", false},
		{"prefixo desconhecido", strings.Replace(hash, "pbkdf2-sha256", "md5", 1), "segredo123", false},
		{"partes faltando", "pbkdf2-sha256$1000$abc", "segredo123", false},
		{"iterações não numéricas", "pbkdf2-sha256$mil$" + partes[2] + "$" + partes[3], "segredo123", false},
		{"iterações zero", "pbkdf2-sha256$0$" + partes[2] + "$" + partes[3], "segredo123", false},
		{"sal fora de base64", "pbkdf2-sha256$1000$***$" + partes[3], "segredo123", false},
		{"hash fora de base64", "pbkdf2-sha256$1000$" + partes[2] + "$***", "segredo123", false},
		{"hash vazio", "pbkdf2-sha256$1000$" + partes[2] + "$", "segredo123", false},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			if obtido := senhaConfere(tc.hash, tc.senha); obtido != tc.confere {
				t.Errorf("senhaConfere: esperado %v, obtido %v", tc.confere, obtido)
			}
		})
	}
}

func TestTokenDeSessao(t *testing.T) {
	token := novoToken()
	if len(token) != 43 || strings.ContainsAny(token, "+/=") {
		t.Errorf("token deveria ter 32 bytes em base64url sem padding, obteve %q", token)
	}
	if novoToken() == token {
		t.Error("tokens deveriam ser aleatórios")
	}
	if h := hashToken(token); len(h) != 64 || h == token {
		t.Errorf("hash do token deveria ser sha256 em hex, obteve %q", h)
	}
}

func TestLimitador(t *testing.T) {
	l := novoLimitador(2, time.Minute)

	passos := []struct {
		nome      string
		chave     string
		decorrido time.Duration
		permitido bool
	}{
		{"1ª tentativa", "ip-a", 0, true},
		{"2ª tentativa", "ip-a", time.Second, true},
		{"3ª tentativa na mesma janela é bloqueada", "ip-a", 2 * time.Second, false},
		{"outro IP tem sua própria janela", "ip-b", 2 * time.Second, true},
		{"ainda bloqueado pouco antes de a janela acabar", "ip-a", time.Minute - time.Nanosecond, false},
		{"janela nova depois de um minuto", "ip-a", time.Minute, true},
		{"dois minutos depois as duas janelas antigas venceram", "ip-b", 2 * time.Minute, true},
	}
	for _, p := range passos {
		t.Run(p.nome, func(t *testing.T) {
			if obtido := l.permitir(p.chave, instanteInicial.Add(p.decorrido)); obtido != p.permitido {
				t.Errorf("permitir: esperado %v, obtido %v", p.permitido, obtido)
			}
		})
	}

	if len(l.janelas) != 1 {
		t.Errorf("janelas vencidas deveriam ser descartadas, restaram %v", l.janelas)
	}
}

func TestLimitadorDevolver(t *testing.T) {
	l := novoLimitador(2, time.Minute)
	agora := instanteInicial

	l.permitir("ip-a", agora)
	l.permitir("ip-a", agora)
	l.devolver("ip-a")
	if !l.permitir("ip-a", agora) {
		t.Error("a tentativa devolvida deveria liberar uma vaga na janela")
	}
	if l.permitir("ip-a", agora) {
		t.Error("sem devolução, a 3ª tentativa continua bloqueada")
	}

	l.permitir("ip-b", agora)
	l.devolver("ip-b")
	l.devolver("ip-desconhecido")
	if _, existe := l.janelas["ip-b"]; existe || len(l.janelas) != 1 {
		t.Errorf("janelas zeradas ou inexistentes não deveriam ficar na memória: %v", l.janelas)
	}
}

func TestLogin(t *testing.T) {
	c := novoCenario(t)
	mariaID := c.usuario(t, "maria", PapelEleitor)

	testes := []struct {
		nome     string
		login    string
		senha    string
		esperado error
	}{
		{"credenciais corretas", "maria", senhaDeTeste, nil},
		{"senha errada", "maria", "outra-senha", ErrCredenciaisInvalidas},
		{"usuário inexistente", "ninguem", senhaDeTeste, ErrCredenciaisInvalidas},
		{"login com caixa diferente", "Maria", senhaDeTeste, ErrCredenciaisInvalidas},
		{"login com caractere NUL", "maria\x00", senhaDeTeste, ErrCredenciaisInvalidas},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			sessao, err := c.svc.Login(t.Context(), tc.login, tc.senha, "10.0.0.1")
			conferirErro(t, err, tc.esperado)
			if tc.esperado != nil {
				return
			}
			if sessao.Token == "" || sessao.Usuario != (Usuario{ID: mariaID, Nome: "Nome maria", Login: "maria", Papel: PapelEleitor}) {
				t.Errorf("sessão inesperada: %+v", sessao)
			}
			if c.contar(t, "SELECT count(*) FROM sessoes WHERE token_hash = $1", sessao.Token) != 0 {
				t.Error("o token em si não pode ser gravado no banco")
			}
			if c.contar(t, "SELECT count(*) FROM sessoes WHERE token_hash = $1", hashToken(sessao.Token)) != 1 {
				t.Error("o hash do token deveria estar gravado")
			}
		})
	}
}

func TestLoginLimitaTentativasPorIP(t *testing.T) {
	c := novoCenario(t)
	c.usuario(t, "maria", PapelEleitor)

	for i := 1; i <= limiteTentativasLogin; i++ {
		if _, err := c.svc.Login(t.Context(), "maria", "errada", "10.0.0.1"); !errors.Is(err, ErrCredenciaisInvalidas) {
			t.Fatalf("tentativa %d: esperava ErrCredenciaisInvalidas, obteve %v", i, err)
		}
	}
	if _, err := c.svc.Login(t.Context(), "maria", senhaDeTeste, "10.0.0.1"); !errors.Is(err, ErrMuitasTentativas) {
		t.Errorf("11ª tentativa, mesmo com a senha certa: esperava ErrMuitasTentativas, obteve %v", err)
	}
	if _, err := c.svc.Login(t.Context(), "maria", senhaDeTeste, "10.0.0.2"); err != nil {
		t.Errorf("outro IP não deveria ser bloqueado: %v", err)
	}

	c.relogio.Avancar(time.Minute)
	if _, err := c.svc.Login(t.Context(), "maria", senhaDeTeste, "10.0.0.1"); err != nil {
		t.Errorf("depois de um minuto o IP volta a poder tentar: %v", err)
	}
}

func TestLoginSoContaCredenciaisInvalidas(t *testing.T) {
	c := novoCenario(t)
	c.usuario(t, "maria", PapelEleitor)
	c.usuario(t, "joao", PapelEleitor)
	const proxy = "172.18.0.3" // vários usuários atrás do mesmo proxy reverso

	t.Run("logins certos do mesmo IP não contam no limite", func(t *testing.T) {
		for i := 1; i <= limiteTentativasPorIP+1; i++ {
			login := []string{"maria", "joao"}[i%2]
			if _, err := c.svc.Login(t.Context(), login, senhaDeTeste, proxy); err != nil {
				t.Fatalf("login %d (%s): %v", i, login, err)
			}
		}
	})

	t.Run("falhas de um login não bloqueiam outro login do mesmo IP", func(t *testing.T) {
		for i := 1; i <= limiteTentativasLogin; i++ {
			_, err := c.svc.Login(t.Context(), "maria", "errada", proxy)
			conferirErro(t, err, ErrCredenciaisInvalidas)
		}
		_, err := c.svc.Login(t.Context(), "maria", senhaDeTeste, proxy)
		conferirErro(t, err, ErrMuitasTentativas)
		if _, err := c.svc.Login(t.Context(), "joao", senhaDeTeste, proxy); err != nil {
			t.Errorf("joao não deveria ser bloqueado pelas falhas de maria: %v", err)
		}
	})

	c.relogio.Avancar(time.Minute)

	t.Run("falhas em logins diferentes somam no limite por IP", func(t *testing.T) {
		for i := 1; i <= limiteTentativasPorIP; i++ {
			_, err := c.svc.Login(t.Context(), fmt.Sprintf("inexistente-%d", i), "errada", proxy)
			conferirErro(t, err, ErrCredenciaisInvalidas)
		}
		_, err := c.svc.Login(t.Context(), "joao", senhaDeTeste, proxy)
		conferirErro(t, err, ErrMuitasTentativas)
		if _, existe := c.svc.tentativasPorConta.janelas[proxy+"\x00joao"]; existe {
			t.Error("a tentativa barrada pelo limite do IP não deveria gastar o limite da conta")
		}
		if _, err := c.svc.Login(t.Context(), "joao", senhaDeTeste, "10.0.0.9"); err != nil {
			t.Errorf("outro IP não deveria ser bloqueado: %v", err)
		}
	})
}

func TestLoginLimitaTentativasSimultaneas(t *testing.T) {
	c := novoCenario(t)
	c.usuario(t, "maria", PapelEleitor)

	const pedidos = 3 * limiteTentativasLogin
	erros := make(chan error, pedidos)
	var wg sync.WaitGroup
	for range pedidos {
		wg.Go(func() {
			_, err := c.svc.Login(t.Context(), "maria", "errada", "10.0.0.1")
			erros <- err
		})
	}
	wg.Wait()
	close(erros)

	contagem := map[error]int{}
	for err := range erros {
		contagem[err]++
	}
	esperado := map[error]int{ErrCredenciaisInvalidas: limiteTentativasLogin, ErrMuitasTentativas: pedidos - limiteTentativasLogin}
	if !maps.Equal(contagem, esperado) {
		t.Errorf("resultados: esperado %v, obtido %v", esperado, contagem)
	}
}

func TestSessao(t *testing.T) {
	c := novoCenario(t)
	mariaID := c.usuario(t, "maria", PapelEleitor)
	sessao, err := c.svc.Login(t.Context(), "maria", senhaDeTeste, "10.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("token válido identifica o usuário", func(t *testing.T) {
		u, err := c.svc.UsuarioDaSessao(t.Context(), sessao.Token)
		if err != nil || u.ID != mariaID {
			t.Errorf("esperava o usuário %d, obteve %+v (err %v)", mariaID, u, err)
		}
	})

	t.Run("token desconhecido", func(t *testing.T) {
		_, err := c.svc.UsuarioDaSessao(t.Context(), "token-inventado")
		conferirErro(t, err, ErrNaoAutenticado)
	})

	t.Run("sessão expira em 12 horas", func(t *testing.T) {
		c.relogio.Avancar(validadeSessao - time.Second)
		if _, err := c.svc.UsuarioDaSessao(t.Context(), sessao.Token); err != nil {
			t.Fatalf("ainda dentro da validade: %v", err)
		}
		c.relogio.Avancar(time.Second)
		_, err := c.svc.UsuarioDaSessao(t.Context(), sessao.Token)
		conferirErro(t, err, ErrNaoAutenticado)
	})

	t.Run("novo login apaga sessões vencidas", func(t *testing.T) {
		novo, err := c.svc.Login(t.Context(), "maria", senhaDeTeste, "10.0.0.1")
		if err != nil {
			t.Fatal(err)
		}
		if total := c.contar(t, "SELECT count(*) FROM sessoes"); total != 1 {
			t.Errorf("esperava só a sessão nova, há %d", total)
		}

		if err := c.svc.Logout(t.Context(), novo.Token); err != nil {
			t.Fatal(err)
		}
		_, err = c.svc.UsuarioDaSessao(t.Context(), novo.Token)
		conferirErro(t, err, ErrNaoAutenticado)
	})
}

func TestCriarUsuario(t *testing.T) {
	c := novoCenario(t)
	c.usuario(t, "maria", PapelEleitor)

	testes := []struct {
		nome, nomeCompleto, login, senha, papel string
		esperado                                error
	}{
		{"usuário válido", "Ana", "ana", "senha-boa", PapelEleitor, nil},
		{"administrador válido", "Admin", "chefe", "senha-boa", PapelAdmin, nil},
		{"nome vazio", "  ", "x1", "senha-boa", PapelEleitor, errValidacao},
		{"nome com 121 caracteres", strings.Repeat("a", 121), "x2", "senha-boa", PapelEleitor, errValidacao},
		{"login vazio", "X", "", "senha-boa", PapelEleitor, errValidacao},
		{"login com espaço", "X", "x y", "senha-boa", PapelEleitor, errValidacao},
		{"senha com 7 caracteres", "X", "x3", "1234567", PapelEleitor, errValidacao},
		{"papel desconhecido", "X", "x4", "senha-boa", "auditor", errValidacao},
		{"login repetido", "Outra Maria", "maria", "senha-boa", PapelEleitor, errValidacao},
		{"login com caractere NUL", "X", "x\x005", "senha-boa", PapelEleitor, errValidacao},
		{"nome com caractere NUL", "X\x00Y", "x6", "senha-boa", PapelEleitor, errValidacao},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			_, err := c.svc.CriarUsuario(t.Context(), tc.nomeCompleto, tc.login, tc.senha, tc.papel)
			conferirErro(t, err, tc.esperado)
		})
	}

	usuarios, err := c.svc.ListarUsuarios(t.Context())
	if err != nil || len(usuarios) != 3 || usuarios[0].Login != "maria" {
		t.Errorf("listagem inesperada: %+v (err %v)", usuarios, err)
	}
	if existem, err := c.svc.ExistemUsuarios(t.Context()); err != nil || !existem {
		t.Errorf("ExistemUsuarios: esperado true, obtido %v (err %v)", existem, err)
	}
}

func TestExistemUsuariosEmBancoVazio(t *testing.T) {
	c := novoCenario(t)
	if existem, err := c.svc.ExistemUsuarios(t.Context()); err != nil || existem {
		t.Errorf("esperado false, obtido %v (err %v)", existem, err)
	}
}

func TestNovoUsaPadroesDeProducao(t *testing.T) {
	s := Novo(nil, Config{})
	if s.iteracoes != 210_000 {
		t.Errorf("iterações: esperado 210000, obtido %d", s.iteracoes)
	}
	if agora := s.Agora(); time.Since(agora) > time.Minute {
		t.Errorf("relógio padrão deveria ser time.Now, obteve %v", agora)
	}
}
