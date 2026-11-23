package voting

import (
	"math"
	"testing"
)

type casoMaioria struct {
	nome                   string
	lider, outros          int
	vencedor, segundoTurno bool
	percentual             float64
}

func verificarMaioria(t *testing.T, testes []casoMaioria) {
	t.Helper()
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			res := VerificarMaioriaAbsoluta(tc.lider, tc.outros)
			if res.VencedorDefinido != tc.vencedor || res.PrecisaSegundoTurno != tc.segundoTurno || math.Abs(res.PercentualLider-tc.percentual) > 1e-9 {
				t.Errorf("resultado = %+v; esperado vencedor=%v, segundo turno=%v, percentual=%v", res, tc.vencedor, tc.segundoTurno, tc.percentual)
			}
		})
	}
}

// Classes de equivalência de VerificarMaioriaAbsoluta(votosLider, outrosVotosValidos):
//
//	CE1 (válida)   líder com mais votos que os demais somados -> vencedor definido
//	CE2 (válida)   líder com votos menores ou iguais aos demais, total > 0 -> segundo turno
//	CE3 (válida)   nenhum voto válido (0 e 0) -> sem vencedor e sem segundo turno
//	CE4 (inválida) votos do líder negativos -> resultado vazio
//	CE5 (inválida) outros votos negativos -> resultado vazio
func TestVerificarMaioriaAbsolutaClassesEquivalencia(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	verificarMaioria(t, []casoMaioria{
		{"CE1: maioria de 60%", 60, 40, true, false, 60},
		{"CE1: total ímpar, dois de três", 2, 1, true, false, 200.0 / 3},
		{"CE1: unanimidade", 10, 0, true, false, 100},
		{"CE2: sem maioria (40%)", 40, 60, false, true, 40},
		{"CE2: líder sem votos", 0, 10, false, true, 0},
		{"CE2: contagens grandes sem overflow", maxInt, maxInt, false, true, 50},
		{"CE3: sem votos válidos", 0, 0, false, false, 0},
		{"CE4: votos do líder negativos", -5, 10, false, false, 0},
		{"CE5: outros votos negativos", 10, -5, false, false, 0},
		{"CE4 e CE5: ambas as contagens negativas", -1, -1, false, false, 0},
	})
}

// Análise de valor limite de VerificarMaioriaAbsoluta:
//
//	VL1 empate entre o líder e os demais (fronteira entre CE2 e CE1)
//	VL2 zero votos (fronteira entre as classes válidas e as negativas)
func TestVerificarMaioriaAbsolutaValorLimite(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	verificarMaioria(t, []casoMaioria{
		{"VL1: um voto a menos que os demais", 49, 50, false, true, 4900.0 / 99},
		{"VL1: empate exato (50%)", 50, 50, false, true, 50},
		{"VL1: um voto a mais que os demais", 51, 50, true, false, 5100.0 / 101},
		{"VL1: um voto a mais com contagens máximas", maxInt, maxInt - 1, true, false, 50},
		{"VL2: líder com -1 voto", -1, 10, false, false, 0},
		{"VL2: líder com 0 votos", 0, 10, false, true, 0},
		{"VL2: outros com -1 voto", 10, -1, false, false, 0},
		{"VL2: outros com 0 votos", 10, 0, true, false, 100},
	})
}
