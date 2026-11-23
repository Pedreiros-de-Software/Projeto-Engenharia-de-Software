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
		msgEsperada string
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
			msgEsperada: "condomino impedido de votar por inadimplencia",
		},
		{
			nome: "Condomino com fracao zerada",
			condomino: Condomino{
				ID:          "apto-303",
				Adimplente:  true,
				FracaoIdeal: 0.0,
			},
			errEsperado: ErrFracaoIdealInvalida,
			msgEsperada: "fracao ideal deve ser maior que zero",
		},
		{
			nome: "ID vazio",
			condomino: Condomino{
				ID:          "",
				Adimplente:  true,
				FracaoIdeal: 10,
			},
			errEsperado: ErrIdentificadorInvalido,
			msgEsperada: "identificador do condomino invalido",
		},
		{
			nome: "ID apenas espaços",
			condomino: Condomino{
				ID:          " ",
				Adimplente:  true,
				FracaoIdeal: 10,
			},
			errEsperado: ErrIdentificadorInvalido,
			msgEsperada: "identificador do condomino invalido",
		},
		{
			nome: "Fração negativa",
			condomino: Condomino{
				ID:          "apto-1",
				Adimplente:  true,
				FracaoIdeal: -1,
			},
			errEsperado: ErrFracaoIdealInvalida,
			msgEsperada: "fracao ideal deve ser maior que zero",
		},
		{
			nome: "Fração NaN",
			condomino: Condomino{
				ID:          "apto-1",
				Adimplente:  true,
				FracaoIdeal: math.NaN(),
			},
			errEsperado: ErrFracaoIdealInvalida,
			msgEsperada: "fracao ideal deve ser maior que zero",
		},
		{
			nome: "Fração infinita",
			condomino: Condomino{
				ID:          "apto-1",
				Adimplente:  true,
				FracaoIdeal: math.Inf(1),
			},
			errEsperado: ErrFracaoIdealInvalida,
			msgEsperada: "fracao ideal deve ser maior que zero",
		},
		{
			nome: "Fração infinita negativa",
			condomino: Condomino{
				ID:          "apto-1",
				Adimplente:  true,
				FracaoIdeal: math.Inf(-1),
			},
			errEsperado: ErrFracaoIdealInvalida,
			msgEsperada: "fracao ideal deve ser maior que zero",
		},
		{
			nome: "ID vazio e inadimplente: identificador e verificado primeiro",
			condomino: Condomino{
				ID:          "",
				Adimplente:  false,
				FracaoIdeal: 10,
			},
			errEsperado: ErrIdentificadorInvalido,
			msgEsperada: "identificador do condomino invalido",
		},
		{
			nome: "Inadimplente e fracao zerada: inadimplencia e verificada antes da fracao",
			condomino: Condomino{
				ID:          "apto-404",
				Adimplente:  false,
				FracaoIdeal: 0,
			},
			errEsperado: ErrInadimplente,
			msgEsperada: "condomino impedido de votar por inadimplencia",
		},
	}

	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			err := ValidarElegibilidade(tc.condomino)
			if !errors.Is(err, tc.errEsperado) {
				t.Fatalf("esperava erro '%v', obteve '%v'", tc.errEsperado, err)
			}
			if err != nil && err.Error() != tc.msgEsperada {
				t.Errorf("mensagem: esperava %q, obteve %q", tc.msgEsperada, err.Error())
			}
		})
	}
}
