package voting

import (
	"errors"
	"math"
	"testing"
)

func TestValidarElegibilidade(t *testing.T) {
	testes := []struct {
		nome        string
		condomino   Condomino
		errEsperado error
	}{
		{
			nome: "Condomino regular e com fracao ideal",
			condomino: Condomino{
				ID:          "apto-101",
				Adimplente:  true,
				FracaoIdeal: 45.5,
			},
			errEsperado: nil,
		},
		{
			nome: "Condomino inadimplente bloqueado",
			condomino: Condomino{
				ID:          "apto-202",
				Adimplente:  false,
				FracaoIdeal: 60.0,
			},
			errEsperado: ErrInadimplente,
		},
		{
			nome: "Condomino com fracao zerada",
			condomino: Condomino{
				ID:          "apto-303",
				Adimplente:  true,
				FracaoIdeal: 0.0,
			},
			errEsperado: ErrFracaoIdealInvalida,
		},
		{
			nome: "ID vazio",
			condomino: Condomino{
				ID:          "",
				Adimplente:  true,
				FracaoIdeal: 10,
			},
			errEsperado: ErrIdentificadorInvalido,
		},
		{
			nome: "ID apenas espaços",
			condomino: Condomino{
				ID:          " ",
				Adimplente:  true,
				FracaoIdeal: 10,
			},
			errEsperado: ErrIdentificadorInvalido,
		},
		{
			nome: "Fração negativa",
			condomino: Condomino{
				ID:          "apto-1",
				Adimplente:  true,
				FracaoIdeal: -1,
			},
			errEsperado: ErrFracaoIdealInvalida,
		},
		{
			nome: "Fração NaN",
			condomino: Condomino{
				ID:          "apto-1",
				Adimplente:  true,
				FracaoIdeal: math.NaN(),
			},
			errEsperado: ErrFracaoIdealInvalida,
		},
		{
			nome: "Fração infinita",
			condomino: Condomino{
				ID:          "apto-1",
				Adimplente:  true,
				FracaoIdeal: math.Inf(1),
			},
			errEsperado: ErrFracaoIdealInvalida,
		},
		{
			nome: "Fração infinita negativa",
			condomino: Condomino{
				ID:          "apto-1",
				Adimplente:  true,
				FracaoIdeal: math.Inf(-1),
			},
			errEsperado: ErrFracaoIdealInvalida,
		},
	}

	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			err := ValidarElegibilidade(tc.condomino)
			if !errors.Is(err, tc.errEsperado) {
				t.Errorf("%s: esperava erro '%v', obteve '%v'", tc.nome, tc.errEsperado, err)
			}
		})
	}
}
