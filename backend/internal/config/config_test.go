package config

import "testing"

func TestLoad(t *testing.T) {
	const url = "postgres://usuario:senha@localhost:5432/banco"

	testes := []struct {
		nome          string
		port, dbURL   string
		seed          string
		portaEsperada string
		seedEsperado  bool
		esperaErro    bool
	}{
		{"PORT vazia usa a porta padrão 8080", "", url, "", "8080", false, false},
		{"PORT definida é respeitada", "9090", url, "", "9090", false, false},
		{"DATABASE_URL vazia é erro", "9090", "", "", "", false, true},
		{"SEED_DEMO=true liga o seed", "", url, "true", "8080", true, false},
		{"SEED_DEMO=false desliga o seed", "", url, "false", "8080", false, false},
		{"SEED_DEMO inválido é erro", "", url, "talvez", "", false, true},
	}

	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			t.Setenv("PORT", tc.port)
			t.Setenv("DATABASE_URL", tc.dbURL)
			t.Setenv("SEED_DEMO", tc.seed)

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
			if cfg.Port != tc.portaEsperada || cfg.DatabaseURL != tc.dbURL || cfg.SeedDemo != tc.seedEsperado {
				t.Errorf("config: esperado porta %q, url %q e seed %v, obtido %+v", tc.portaEsperada, tc.dbURL, tc.seedEsperado, cfg)
			}
		})
	}
}
