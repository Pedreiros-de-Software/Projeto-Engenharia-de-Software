package voting

import (
	"math"
	"testing"
)

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
		{"ID vazio", Condomino{"", true, 10}, false, "identificador do condomino invalido"},
		{"ID apenas espaços", Condomino{"   ", true, 10}, false, "identificador do condomino invalido"},
		{"Fração negativa", Condomino{"apto-1", true, -1}, false, "fracao ideal deve ser maior que zero"},
		{"Fração NaN", Condomino{"apto-1", true, math.NaN()}, false, "fracao ideal deve ser maior que zero"},
		{"Fração infinita", Condomino{"apto-1", true, math.Inf(1)}, false, "fracao ideal deve ser maior que zero"},
		{"Fração infinita negativa", Condomino{"apto-1", true, math.Inf(-1)}, false, "fracao ideal deve ser maior que zero"},
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
