package config

import "testing"

func TestLoad(t *testing.T) {
	const url = "postgres://usuario:senha@localhost:5432/banco"

	testes := []struct {
		nome          string
		port, dbURL   string
		portaEsperada string
		esperaErro    bool
	}{
		{"PORT vazia usa a porta padrão 8080", "", url, "8080", false},
		{"PORT definida é respeitada", "9090", url, "9090", false},
		{"DATABASE_URL vazia é erro", "9090", "", "", true},
	}

	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			t.Setenv("PORT", tc.port)
			t.Setenv("DATABASE_URL", tc.dbURL)

			cfg, err := Load()

			if tc.esperaErro {
				if err == nil || cfg != nil {
					t.Fatalf("esperava erro e config nil, obteve cfg=%+v err=%v", cfg, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if cfg.Port != tc.portaEsperada || cfg.DatabaseURL != tc.dbURL {
				t.Errorf("config: esperado porta %q e url %q, obtido %+v", tc.portaEsperada, tc.dbURL, cfg)
			}
		})
	}
}
