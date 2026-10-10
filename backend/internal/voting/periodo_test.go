package voting

import (
	"errors"
	"testing"
	"time"
)

func TestStatusVotacao(t *testing.T) {
	inicio := time.Date(2026, 10, 10, 8, 0, 0, 0, time.UTC)
	fim := time.Date(2026, 10, 10, 18, 0, 0, 0, time.UTC)

	testes := []struct {
		nome           string
		inicio         time.Time
		fim            time.Time
		agora          time.Time
		estadoEsperado EstadoVotacao
		erroEsperado   error
	}{
		{
			nome:           "Antes do inicio",
			inicio:         inicio,
			fim:            fim,
			agora:          inicio.Add(-time.Minute),
			estadoEsperado: NaoIniciada,
		},
		{
			nome:           "Exatamente no inicio",
			inicio:         inicio,
			fim:            fim,
			agora:          inicio,
			estadoEsperado: Aberta,
		},
		{
			nome:           "Durante a votacao",
			inicio:         inicio,
			fim:            fim,
			agora:          inicio.Add(5 * time.Hour),
			estadoEsperado: Aberta,
		},
		{
			nome:           "Um instante antes do fim",
			inicio:         inicio,
			fim:            fim,
			agora:          fim.Add(-time.Nanosecond),
			estadoEsperado: Aberta,
		},
		{
			nome:           "Exatamente no fim",
			inicio:         inicio,
			fim:            fim,
			agora:          fim,
			estadoEsperado: Encerrada,
		},
		{
			nome:           "Depois do fim",
			inicio:         inicio,
			fim:            fim,
			agora:          fim.Add(24 * time.Hour),
			estadoEsperado: Encerrada,
		},
		{
			nome:           "Fuso horario diferente no mesmo instante",
			inicio:         inicio,
			fim:            fim,
			agora:          inicio.In(time.FixedZone("BRT", -3*60*60)),
			estadoEsperado: Aberta,
		},
		{
			nome:         "Fim antes do inicio",
			inicio:       fim,
			fim:          inicio,
			agora:        inicio,
			erroEsperado: ErrJanelaInvalida,
		},
		{
			nome:         "Fim igual ao inicio",
			inicio:       inicio,
			fim:          inicio,
			agora:        inicio,
			erroEsperado: ErrJanelaInvalida,
		},
		{
			nome:         "Inicio zerado",
			fim:          fim,
			agora:        inicio,
			erroEsperado: ErrJanelaInvalida,
		},
		{
			nome:         "Fim zerado",
			inicio:       inicio,
			agora:        inicio,
			erroEsperado: ErrJanelaInvalida,
		},
		{
			nome:         "Instante atual zerado",
			inicio:       inicio,
			fim:          fim,
			erroEsperado: ErrJanelaInvalida,
		},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			estado, err := StatusVotacao(tt.inicio, tt.fim, tt.agora)
			if !errors.Is(err, tt.erroEsperado) {
				t.Fatalf("erro: esperado %v, obtido %v", tt.erroEsperado, err)
			}
			if tt.erroEsperado == nil && estado != tt.estadoEsperado {
				t.Errorf("estado: esperado %s, obtido %s", tt.estadoEsperado, estado)
			}
		})
	}
}

func TestEstadoVotacaoString(t *testing.T) {
	casos := map[EstadoVotacao]string{
		NaoIniciada:      "nao iniciada",
		Aberta:           "aberta",
		Encerrada:        "encerrada",
		EstadoVotacao(9): "desconhecido",
	}
	for estado, esperado := range casos {
		if obtido := estado.String(); obtido != esperado {
			t.Errorf("String(%d): esperado %q, obtido %q", int(estado), esperado, obtido)
		}
	}
}
