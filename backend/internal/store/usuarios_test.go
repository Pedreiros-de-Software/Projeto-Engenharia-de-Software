package store_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store"
)

func inserirUsuario(t *testing.T, repo store.Repositorios, login, papel string) int64 {
	t.Helper()
	id, err := repo.Usuarios.Inserir(t.Context(), store.Usuario{Nome: "Nome " + login, Login: login, SenhaHash: "hash-" + login, Papel: papel})
	if err != nil {
		t.Fatalf("inserindo usuário %s: %v", login, err)
	}
	return id
}

func TestUsuarios(t *testing.T) {
	_, repo := novoBanco(t)
	ctx := t.Context()

	adminID := inserirUsuario(t, repo, "admin", "admin")
	mariaID := inserirUsuario(t, repo, "maria", "eleitor")

	t.Run("login repetido é ErrDuplicado", func(t *testing.T) {
		_, err := repo.Usuarios.Inserir(ctx, store.Usuario{Nome: "Outra", Login: "maria", SenhaHash: "h", Papel: "eleitor"})
		if !errors.Is(err, store.ErrDuplicado) {
			t.Errorf("esperava ErrDuplicado, obteve %v", err)
		}
	})

	t.Run("busca por login devolve todos os campos", func(t *testing.T) {
		u, err := repo.Usuarios.BuscarPorLogin(ctx, "maria")
		if err != nil {
			t.Fatal(err)
		}
		esperado := store.Usuario{ID: mariaID, Nome: "Nome maria", Login: "maria", SenhaHash: "hash-maria", Papel: "eleitor"}
		if u != esperado {
			t.Errorf("esperado %+v, obtido %+v", esperado, u)
		}
	})

	t.Run("login inexistente é ErrNaoEncontrado", func(t *testing.T) {
		if _, err := repo.Usuarios.BuscarPorLogin(ctx, "ninguem"); !errors.Is(err, store.ErrNaoEncontrado) {
			t.Errorf("esperava ErrNaoEncontrado, obteve %v", err)
		}
	})

	t.Run("listar segue a ordem de cadastro", func(t *testing.T) {
		usuarios, err := repo.Usuarios.Listar(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(usuarios) != 2 || usuarios[0].ID != adminID || usuarios[1].ID != mariaID {
			t.Errorf("listagem inesperada: %+v", usuarios)
		}
	})

	t.Run("contar", func(t *testing.T) {
		if total, err := repo.Usuarios.Contar(ctx); err != nil || total != 2 {
			t.Errorf("esperava 2 usuários, obteve %d (err %v)", total, err)
		}
	})
}

func TestSessoes(t *testing.T) {
	_, repo := novoBanco(t)
	ctx := t.Context()
	usuarioID := inserirUsuario(t, repo, "maria", "eleitor")

	if err := repo.Sessoes.Inserir(ctx, "valida", usuarioID, instante.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := repo.Sessoes.Inserir(ctx, "vencida", usuarioID, instante.Add(-time.Second)); err != nil {
		t.Fatal(err)
	}

	testes := []struct {
		nome       string
		tokenHash  string
		agora      time.Time
		encontrada bool
	}{
		{"sessão dentro da validade", "valida", instante, true},
		{"sessão no instante exato da expiração", "valida", instante.Add(time.Hour), false},
		{"sessão vencida", "vencida", instante, false},
		{"hash desconhecido", "outro", instante, false},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			u, err := repo.Sessoes.BuscarUsuario(ctx, tc.tokenHash, tc.agora)
			if tc.encontrada {
				if err != nil || u.ID != usuarioID {
					t.Errorf("esperava o usuário %d, obteve %+v (err %v)", usuarioID, u, err)
				}
				return
			}
			if !errors.Is(err, store.ErrNaoEncontrado) {
				t.Errorf("esperava ErrNaoEncontrado, obteve %v", err)
			}
		})
	}

	t.Run("apagar expiradas mantém as válidas", func(t *testing.T) {
		if err := repo.Sessoes.ApagarExpiradas(ctx, instante); err != nil {
			t.Fatal(err)
		}
		if _, err := repo.Sessoes.BuscarUsuario(ctx, "vencida", instante.Add(-time.Hour)); !errors.Is(err, store.ErrNaoEncontrado) {
			t.Errorf("sessão vencida deveria ter sido apagada, err %v", err)
		}
		if _, err := repo.Sessoes.BuscarUsuario(ctx, "valida", instante); err != nil {
			t.Errorf("sessão válida não deveria ter sido apagada: %v", err)
		}
	})

	t.Run("apagar encerra a sessão", func(t *testing.T) {
		if err := repo.Sessoes.Apagar(ctx, "valida"); err != nil {
			t.Fatal(err)
		}
		if _, err := repo.Sessoes.BuscarUsuario(ctx, "valida", instante); !errors.Is(err, store.ErrNaoEncontrado) {
			t.Errorf("esperava ErrNaoEncontrado, obteve %v", err)
		}
	})
}
