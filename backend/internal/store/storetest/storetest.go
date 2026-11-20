// Package storetest cria bancos isolados para os testes de integração com Postgres.
package storetest

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/url"
	"os"
	"testing"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/database"
	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store"
)

// NovoBanco é NovoBancoVazio com as migrações já aplicadas.
func NovoBanco(t testing.TB) *sql.DB {
	t.Helper()
	db := NovoBancoVazio(t)
	if err := store.Migrar(t.Context(), db); err != nil {
		t.Fatalf("migrando schema de teste: %v", err)
	}
	return db
}

// NovoBancoVazio cria um schema exclusivo do teste, sem tabelas, conecta com search_path apontando
// para ele e apaga o schema ao final. Pula o teste se TEST_DATABASE_URL não existir.
func NovoBancoVazio(t testing.TB) *sql.DB {
	t.Helper()

	endereco := os.Getenv("TEST_DATABASE_URL")
	if endereco == "" {
		t.Skip("TEST_DATABASE_URL não definida: teste de integração com Postgres ignorado")
	}

	// O nome do schema é gerado aqui (hex aleatório), nunca vem de entrada externa.
	schema := "teste_" + sufixoAleatorio()

	admin := conectar(t, endereco)
	if _, err := admin.Exec("CREATE SCHEMA " + schema); err != nil {
		t.Fatalf("criando schema de teste: %v", err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Errorf("apagando schema de teste: %v", err)
		}
		_ = admin.Close()
	})

	db := conectar(t, comSearchPath(t, endereco, schema))
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func conectar(t testing.TB, endereco string) *sql.DB {
	t.Helper()
	db, err := database.Connect(endereco)
	if err != nil {
		t.Fatalf("conectando ao Postgres de teste: %v", err)
	}
	return db
}

// comSearchPath acrescenta search_path à URL; o pgx o envia como parâmetro da sessão.
func comSearchPath(t testing.TB, endereco, schema string) string {
	t.Helper()
	u, err := url.Parse(endereco)
	if err != nil {
		t.Fatalf("TEST_DATABASE_URL precisa estar no formato URL: %v", err)
	}
	consulta := u.Query()
	consulta.Set("search_path", schema)
	u.RawQuery = consulta.Encode()
	return u.String()
}

func sufixoAleatorio() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
