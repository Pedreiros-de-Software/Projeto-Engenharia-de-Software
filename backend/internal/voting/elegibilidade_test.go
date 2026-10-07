package voting

import "testing"

func TestValidarElegibilidade(t *testing.T) {
	testes := []struct {
		nome         string
		condomino    Condomino
		esperaValido bool
		msgEsperada  string
	}{
		{
			nome: "Condomino regular e com fracao ideal",
			condomino: Condomino{
				ID:          "apto-101",
				Adimplente:  true,
				FracaoIdeal: 45.5,
			},
			esperaValido: true,
			msgEsperada:  "apto a votar",
		},
		{
			nome: "Condomino inadimplente bloqueado",
			condomino: Condomino{
				ID:          "apto-202",
				Adimplente:  false,
				FracaoIdeal: 60.0,
			},
			esperaValido: false,
			msgEsperada:  "condomino impedido de votar por inadimplencia",
		},
		{
			nome: "Condomino com fracao zerada",
			condomino: Condomino{
				ID:          "apto-303",
				Adimplente:  true,
				FracaoIdeal: 0.0,
			},
			esperaValido: false,
			msgEsperada:  "fracao ideal deve ser maior que zero",
		},
	}

	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			valido, msg := ValidarElegibilidade(tc.condomino)
			if valido != tc.esperaValido {
				t.Errorf("%s: esperava valido=%v, obteve %v", tc.nome, tc.esperaValido, valido)
			}
			if msg != tc.msgEsperada {
				t.Errorf("%s: esperava msg='%s', obteve '%s'", tc.nome, tc.msgEsperada, msg)
			}
		})
	}
}
