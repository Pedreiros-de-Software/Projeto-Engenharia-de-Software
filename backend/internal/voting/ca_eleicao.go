package voting

// ResultadoSegundoTurno armazena o resultado da verificação de maioria.
type ResultadoSegundoTurno struct {
	VencedorDefinido    bool
	PrecisaSegundoTurno bool
	PercentualLider     float64
}

// VerificarMaioriaAbsoluta recebe contagens já limitadas aos votos válidos.
// Brancos e nulos devem ser excluídos pelo chamador. Sem votos válidos ou com
// contagens negativas, retorna resultado vazio: não há vencedor nem segundo turno.
// A função sinaliza a necessidade de segundo turno; não cria um novo pleito.
func VerificarMaioriaAbsoluta(votosLider int, outrosVotosValidos int) ResultadoSegundoTurno {
	if votosLider < 0 || outrosVotosValidos < 0 || (votosLider == 0 && outrosVotosValidos == 0) {
		return ResultadoSegundoTurno{}
	}
	// Converte antes de somar para evitar overflow na soma de inteiros.
	totalValidos := float64(votosLider) + float64(outrosVotosValidos)
	venceu := votosLider > outrosVotosValidos
	return ResultadoSegundoTurno{
		VencedorDefinido:    venceu,
		PrecisaSegundoTurno: !venceu,
		PercentualLider:     float64(votosLider) / totalValidos * 100,
	}
}
