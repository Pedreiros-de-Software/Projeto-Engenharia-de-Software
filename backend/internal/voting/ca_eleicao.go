package voting

// armazena o resultado da verificação de maioria
type ResultadoSegundoTurno struct {
	VencedorDefinido    bool
	PrecisaSegundoTurno bool
	PercentualLider     float64
}

// calcula se a chapa líder obteve mais de 50% dos votos válidos
// considera apenas votos válidos (brancos e nulos descartados)
func VerificarMaioriaAbsoluta(votosLider int, outrosVotosValidos int) ResultadoSegundoTurno {
	totalValidos := votosLider + outrosVotosValidos

	if totalValidos <= 0 || votosLider <= 0 {
		return ResultadoSegundoTurno{
			VencedorDefinido:    false,
			PrecisaSegundoTurno: true,
			PercentualLider:     0.0,
		}
	}

	percentual := (float64(votosLider) / float64(totalValidos)) * 100.0

	if percentual > 50.0 {
		return ResultadoSegundoTurno{
			VencedorDefinido:    true,
			PrecisaSegundoTurno: false,
			PercentualLider:     percentual,
		}
	}

	return ResultadoSegundoTurno{
		VencedorDefinido:    false,
		PrecisaSegundoTurno: true,
		PercentualLider:     percentual,
	}
}
