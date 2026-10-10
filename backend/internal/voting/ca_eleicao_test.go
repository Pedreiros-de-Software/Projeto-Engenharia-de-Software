package voting

import (
	"math"
	"testing"
)

func TestVerificarMaioriaAbsoluta(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	testes := []struct {
		nome                   string
		lider, outros          int
		vencedor, segundoTurno bool
		percentual             float64
	}{
		{"Maioria 60%", 60, 40, true, false, 60},
		{"Sem maioria", 40, 60, false, true, 40},
		{"Exatamente metade", 50, 50, false, true, 50},
		{"Total ímpar: dois de três", 2, 1, true, false, 200.0 / 3},
		{"Unanimidade", 10, 0, true, false, 100},
		{"Líder sem votos", 0, 10, false, true, 0},
		{"Sem votos válidos", 0, 0, false, false, 0},
		{"Votos do líder negativos", -1, 10, false, false, 0},
		{"Outros votos negativos", 10, -1, false, false, 0},
		{"Ambas contagens negativas", -1, -1, false, false, 0},
		{"Contagens grandes sem overflow", maxInt, maxInt, false, true, 50},
		{"Maioria inteira não perdida no arredondamento", maxInt, maxInt - 1, true, false, 50},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			res := VerificarMaioriaAbsoluta(tc.lider, tc.outros)
			if res.VencedorDefinido != tc.vencedor || res.PrecisaSegundoTurno != tc.segundoTurno || math.Abs(res.PercentualLider-tc.percentual) > 1e-9 {
				t.Errorf("resultado = %+v; esperado vencedor=%v, segundo turno=%v, percentual=%v", res, tc.vencedor, tc.segundoTurno, tc.percentual)
			}
		})
	}
}
