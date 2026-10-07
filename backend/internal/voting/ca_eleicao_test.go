package voting

import "testing"

func TestVerificarMaioriaAbsoluta(t *testing.T) {
	testes := []struct {
		nome               string
		votosLider         int
		outrosVotosValidos int
		esperaVencedor     bool
		esperaSegundoTurno bool
	}{
		{
			nome:               "Vencedor no primeiro turno com 60%",
			votosLider:         60,
			outrosVotosValidos: 40,
			esperaVencedor:     true,
			esperaSegundoTurno: false,
		},
		{
			nome:               "Empate tecnico ou dispersao (40%), necessário segundo turno",
			votosLider:         40,
			outrosVotosValidos: 60,
			esperaVencedor:     false,
			esperaSegundoTurno: true,
		},
		{
			nome:               "Exatamente 50%, nao atinge mais de 50%, necessário segundo turno",
			votosLider:         50,
			outrosVotosValidos: 50,
			esperaVencedor:     false,
			esperaSegundoTurno: true,
		},
		{
			nome:               "Sem votos computados",
			votosLider:         0,
			outrosVotosValidos: 0,
			esperaVencedor:     false,
			esperaSegundoTurno: true,
		},
	}

	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			res := VerificarMaioriaAbsoluta(tc.votosLider, tc.outrosVotosValidos)
			if res.VencedorDefinido != tc.esperaVencedor {
				t.Errorf("%s: esperava VencedorDefinido=%v, obteve %v", tc.nome, tc.esperaVencedor, res.VencedorDefinido)
			}
			if res.PrecisaSegundoTurno != tc.esperaSegundoTurno {
				t.Errorf("%s: esperava PrecisaSegundoTurno=%v, obteve %v", tc.nome, tc.esperaSegundoTurno, res.PrecisaSegundoTurno)
			}
		})
	}
}
