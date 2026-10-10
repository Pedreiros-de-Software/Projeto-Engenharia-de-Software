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
			estadoEsperado: VotacaoNaoIniciada,
		},
		{
			nome:           "Exatamente no inicio",
			inicio:         inicio,
			fim:            fim,
			agora:          inicio,
			estadoEsperado: VotacaoAberta,
		},
		{
			nome:           "Durante a votacao",
			inicio:         inicio,
			fim:            fim,
			agora:          inicio.Add(5 * time.Hour),
			estadoEsperado: VotacaoAberta,
		},
		{
			nome:           "Um instante antes do fim",
			inicio:         inicio,
			fim:            fim,
			agora:          fim.Add(-time.Nanosecond),
			estadoEsperado: VotacaoAberta,
		},
		{
			nome:           "Exatamente no fim",
			inicio:         inicio,
			fim:            fim,
			agora:          fim,
			estadoEsperado: VotacaoEncerrada,
		},
		{
			nome:           "Depois do fim",
			inicio:         inicio,
			fim:            fim,
			agora:          fim.Add(24 * time.Hour),
			estadoEsperado: VotacaoEncerrada,
		},
		{
			nome:           "Fuso horario diferente no mesmo instante",
			inicio:         inicio,
			fim:            fim,
			agora:          inicio.In(time.FixedZone("BRT", -3*60*60)),
			estadoEsperado: VotacaoAberta,
		},
		{
			nome:           "Fim antes do inicio",
			inicio:         fim,
			fim:            inicio,
			agora:          inicio,
			erroEsperado:   ErrJanelaInvalida,
			estadoEsperado: VotacaoInvalida,
		},
		{
			nome:           "Fim igual ao inicio",
			inicio:         inicio,
			fim:            inicio,
			agora:          inicio,
			erroEsperado:   ErrJanelaInvalida,
			estadoEsperado: VotacaoInvalida,
		},
		{
			nome:           "Inicio zerado",
			fim:            fim,
			agora:          inicio,
			erroEsperado:   ErrJanelaInvalida,
			estadoEsperado: VotacaoInvalida,
		},
		{
			nome:           "Fim zerado",
			inicio:         inicio,
			agora:          inicio,
			erroEsperado:   ErrJanelaInvalida,
			estadoEsperado: VotacaoInvalida,
		},
		{
			nome:           "Instante atual zerado",
			inicio:         inicio,
			fim:            fim,
			erroEsperado:   ErrJanelaInvalida,
			estadoEsperado: VotacaoInvalida,
		},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			estado, err := StatusVotacao(tt.inicio, tt.fim, tt.agora)
			if !errors.Is(err, tt.erroEsperado) {
				t.Fatalf("erro: esperado %v, obtido %v", tt.erroEsperado, err)
			}
			if estado != tt.estadoEsperado {
				t.Errorf("estado: esperado %s, obtido %s", tt.estadoEsperado, estado)
			}
		})
	}
}

func TestEstadoVotacaoString(t *testing.T) {
	casos := map[EstadoVotacao]string{
		VotacaoInvalida:    "invalida",
		VotacaoNaoIniciada: "nao iniciada",
		VotacaoAberta:      "aberta",
		VotacaoEncerrada:   "encerrada",
		EstadoVotacao(9):   "desconhecido",
	}
	for estado, esperado := range casos {
		if obtido := estado.String(); obtido != esperado {
			t.Errorf("String(%d): esperado %q, obtido %q", int(estado), esperado, obtido)
		}
	}
}
