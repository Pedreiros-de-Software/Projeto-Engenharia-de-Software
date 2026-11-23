package store_test

import (
	"database/sql"
	"errors"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store"
	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store/storetest"
)

var instante = time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)

func novoBanco(t *testing.T) (*sql.DB, store.Repositorios) {
	t.Helper()
	db := storetest.NovoBanco(t)
	return db, store.Novo(db)
}

func TestMigrarEIdempotenteEConcorrente(t *testing.T) {
	db := storetest.NovoBancoVazio(t)

	const instancias = 6
	var wg sync.WaitGroup
	erros := make(chan error, instancias)
	for range instancias {
		wg.Go(func() { erros <- store.Migrar(t.Context(), db) })
	}
	wg.Wait()
	close(erros)
	for err := range erros {
		if err != nil {
			t.Fatalf("migrações concorrentes num schema vazio: %v", err)
		}
	}
	if err := store.Migrar(t.Context(), db); err != nil {
		t.Fatalf("migrar de novo: %v", err)
	}

	arquivos, err := filepath.Glob("migrations/*.sql")
	if err != nil || len(arquivos) == 0 {
		t.Fatalf("listando migrações: %v %v", arquivos, err)
	}
	esperadas := make([]string, len(arquivos))
	for i, arquivo := range arquivos {
		esperadas[i] = filepath.Base(arquivo)
	}

	var versoes []string
	rows, err := db.Query("SELECT versao FROM schema_migrations ORDER BY versao")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		versoes = append(versoes, v)
	}
	if !slices.Equal(versoes, esperadas) {
		t.Errorf("versões aplicadas: esperado uma linha por arquivo %v, obtido %v", esperadas, versoes)
	}
}

func TestEmTransacao(t *testing.T) {
	falha := errors.New("falha no meio da transação")

	testes := []struct {
		nome          string
		retorno       error
		totalEsperado int
	}{
		{"erro desfaz o que foi gravado", falha, 0},
		{"sucesso confirma o que foi gravado", nil, 1},
	}

	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			db, repo := novoBanco(t)

			err := store.EmTransacao(t.Context(), db, func(r store.Repositorios) error {
				if _, err := r.Usuarios.Inserir(t.Context(), store.Usuario{Nome: "A", Login: "a", SenhaHash: "h", Papel: "eleitor"}); err != nil {
					return err
				}
				return tc.retorno
			})
			if !errors.Is(err, tc.retorno) {
				t.Fatalf("erro: esperado %v, obtido %v", tc.retorno, err)
			}

			total, err := repo.Usuarios.Contar(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			if total != tc.totalEsperado {
				t.Errorf("usuários gravados: esperado %d, obtido %d", tc.totalEsperado, total)
			}
		})
	}
}

func TestEmLeituraVeUmRetratoFixoESoLe(t *testing.T) {
	db, repo := novoBanco(t)
	inserirUsuario(t, repo, "ana", "eleitor")

	err := store.EmLeitura(t.Context(), db, func(r store.Repositorios) error {
		antes, err := r.Usuarios.Contar(t.Context())
		if err != nil {
			return err
		}
		inserirUsuario(t, repo, "bruno", "eleitor") // confirmado fora da transação de leitura
		depois, err := r.Usuarios.Contar(t.Context())
		if err != nil {
			return err
		}
		if antes != 1 || depois != 1 {
			t.Errorf("a leitura deveria ver sempre 1 usuário, viu %d e depois %d", antes, depois)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = store.EmLeitura(t.Context(), db, func(r store.Repositorios) error {
		_, err := r.Usuarios.Inserir(t.Context(), store.Usuario{Nome: "C", Login: "c", SenhaHash: "h", Papel: "eleitor"})
		return err
	})
	if err == nil {
		t.Error("a transação de leitura não deveria aceitar escrita")
	}

	falha := errors.New("falha na leitura")
	if err := store.EmLeitura(t.Context(), db, func(store.Repositorios) error { return falha }); !errors.Is(err, falha) {
		t.Errorf("erro de fn: esperado %v, obtido %v", falha, err)
	}
}

// O sigilo do voto depende de a cédula não ter coluna que identifique o eleitor ou o horário,
// e de a participação não guardar a escolha.
func TestColunasDasTabelasDeVotoSecreto(t *testing.T) {
	db, _ := novoBanco(t)

	testes := []struct {
		tabela  string
		colunas []string
	}{
		{"cedulas", []string{"id", "votacao_id", "tipo", "opcao_id"}},
		{"cedula_preferencias", []string{"cedula_id", "posicao", "opcao_id"}},
		{"participacoes", []string{"votacao_id", "usuario_id"}},
	}

	for _, tc := range testes {
		t.Run(tc.tabela, func(t *testing.T) {
			rows, err := db.Query(`SELECT column_name FROM information_schema.columns
				WHERE table_schema = current_schema() AND table_name = $1 ORDER BY ordinal_position`, tc.tabela)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = rows.Close() }()

			var colunas []string
			for rows.Next() {
				var c string
				if err := rows.Scan(&c); err != nil {
					t.Fatal(err)
				}
				colunas = append(colunas, c)
			}
			if !slices.Equal(colunas, tc.colunas) {
				t.Errorf("colunas: esperado %v, obtido %v", tc.colunas, colunas)
			}
		})
	}
}
