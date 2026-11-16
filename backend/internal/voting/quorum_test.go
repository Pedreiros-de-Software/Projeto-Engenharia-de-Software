package voting

import (
	"math"
	"testing"
)

type casoQuorum struct {
	nome            string
	votosFavoraveis float64
	totalFracoes    float64
	esperaAprovacao bool
}

func verificarQuorum(t *testing.T, testes []casoQuorum) {
	t.Helper()
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			resultado := AtingiuQuorum(tc.votosFavoraveis, tc.totalFracoes)
			if resultado != tc.esperaAprovacao {
				t.Errorf("AtingiuQuorum(%v, %v): esperava %v, obteve %v", tc.votosFavoraveis, tc.totalFracoes, tc.esperaAprovacao, resultado)
			}
		})
	}
}

// Classes de equivalência de AtingiuQuorum(votosFavoraveis, totalFracoes):
//
//	CE1 (válida)   0 <= votos <= total, total > 0 e votos/total >= 2/3 -> aprovado
//	CE2 (válida)   0 <= votos <= total, total > 0 e votos/total < 2/3  -> reprovado
//	CE3 (inválida) total <= 0
//	CE4 (inválida) votos < 0
//	CE5 (inválida) votos > total
//	CE6 (inválida) votos ou total NaN
//	CE7 (inválida) votos ou total infinito
//
// Toda classe inválida resulta em reprovação (false). As classes inválidas podem
// se sobrepor (total negativo, por exemplo, também deixa votos > total); cada caso
// é rotulado pela violação que ele representa.
func TestAtingiuQuorumClassesEquivalencia(t *testing.T) {
	verificarQuorum(t, []casoQuorum{
		{"CE1: ampla maioria (80%)", 80, 100, true},
		{"CE2: metade dos votos (50%)", 50, 100, false},
		{"CE2: sem votos favoráveis", 0, 100, false},
		{"CE2: percentual truncado 66,66% não equivale a 2/3", 66.66, 100, false},
		{"CE3: total zerado", 10, 0, false},
		{"CE3: total negativo", 1, -100, false},
		{"CE4: votos negativos", -1, 100, false},
		{"CE5: votos excedem o total", 101, 100, false},
		{"CE6: votos NaN", math.NaN(), 100, false},
		{"CE6: total NaN", 100, math.NaN(), false},
		{"CE7: votos infinitos", math.Inf(1), 100, false},
		{"CE7: votos infinito negativo", math.Inf(-1), 100, false},
		{"CE7: total infinito", 100, math.Inf(1), false},
		{"CE7: total infinito negativo", 100, math.Inf(-1), false},
	})
}

// Análise de valor limite de AtingiuQuorum:
//
//	VL1 limite de aprovação: votos/total = 2/3 (fronteira entre CE2 e CE1)
//	VL2 limite superior dos votos: votos = total (fronteira entre CE1 e CE5)
//	VL3 limite inferior do total: total = 0 (fronteira entre CE3 e as classes válidas)
//
// "Logo abaixo" e "logo acima" usam o float64 vizinho (math.Nextafter).
func TestAtingiuQuorumValorLimite(t *testing.T) {
	verificarQuorum(t, []casoQuorum{
		{"VL1: logo abaixo de 2/3", math.Nextafter(200, 0), 300, false},
		{"VL1: exatamente 2/3", 200, 300, true},
		{"VL1: logo acima de 2/3", math.Nextafter(200, 300), 300, true},
		{"VL2: votação unânime (votos iguais ao total)", 100, 100, true},
		{"VL2: logo acima do total", math.Nextafter(100, 200), 100, false},
		{"VL3: total logo abaixo de zero", 0, math.Nextafter(0, -1), false},
		{"VL3: total zero", 0, 0, false},
		{"VL3: menor total positivo com unanimidade", math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64, true},
	})
}
