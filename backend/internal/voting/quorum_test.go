package voting

import "testing"

func TestAtingiuQuorum(t *testing.T) {
	testes := []struct {
		nome            string
		votosFavoraveis float64
		totalFracoes    float64
		esperaAprovacao bool
	}{
		{
			nome:            "Aprovado com exatamente dois tercos",
			votosFavoraveis: 200.0,
			totalFracoes:    300.0,
			esperaAprovacao: true,
		},
		{
			nome:            "Aprovado com ampla maioria (80%)",
			votosFavoraveis: 80.0,
			totalFracoes:    100.0,
			esperaAprovacao: true,
		},
		{
			nome:            "Reprovado com apenas metade dos votos (50%)",
			votosFavoraveis: 50.0,
			totalFracoes:    100.0,
			esperaAprovacao: false,
		},
		{
			nome:            "Reprovado com total de fracoes zerado ou invalido",
			votosFavoraveis: 10.0,
			totalFracoes:    0.0,
			esperaAprovacao: false,
		},
	}

	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			resultado := AtingiuQuorum(tc.votosFavoraveis, tc.totalFracoes)
			if resultado != tc.esperaAprovacao {
				t.Errorf("para %s: esperava %v, mas obteve %v", tc.nome, tc.esperaAprovacao, resultado)
			}
		})
	}
}
