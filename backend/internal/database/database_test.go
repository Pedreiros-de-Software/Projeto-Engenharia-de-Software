package database

import (
	"os"
	"strings"
	"testing"
)

func TestConnectFalhas(t *testing.T) {
	testes := []struct {
		nome       string
		url        string
		erroContem string
	}{
		{"URL mal formada", "postgres://%zz", "cannot parse"},
		{"banco inacessível", "postgres://usuario:senha@127.0.0.1:1/banco?sslmode=disable&connect_timeout=2", "pinging database"},
	}

	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			db, err := Connect(tc.url)
			if db != nil {
				t.Errorf("esperava conexão nil, obteve %v", db)
			}
			if err == nil || !strings.Contains(err.Error(), tc.erroContem) {
				t.Errorf("esperava erro contendo %q, obteve %v", tc.erroContem, err)
			}
		})
	}
}

func TestConnectComBancoDisponivel(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL não definida: teste de integração com Postgres ignorado")
	}

	db, err := Connect(url)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Ping(); err != nil {
		t.Errorf("conexão devolvida deveria responder ao ping: %v", err)
	}
}
