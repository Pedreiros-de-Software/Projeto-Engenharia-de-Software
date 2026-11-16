package voting

import (
	"errors"
	"testing"
	"time"
)

var (
	inicioPadrao = time.Date(2026, 10, 10, 8, 0, 0, 0, time.UTC)
	fimPadrao    = time.Date(2026, 10, 10, 18, 0, 0, 0, time.UTC)
)

type casoPeriodo struct {
	nome           string
	inicio, fim    time.Time
	agora          time.Time
	estadoEsperado EstadoVotacao
	erroEsperado   error
}

func verificarPeriodo(t *testing.T, testes []casoPeriodo) {
	t.Helper()
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

// Classes de equivalência de StatusVotacao(inicio, fim, agora):
//
//	CE1 (válida)   agora antes do início -> VotacaoNaoIniciada
//	CE2 (válida)   agora entre o início (incluso) e o fim (excluso) -> VotacaoAberta
//	CE3 (válida)   agora no fim ou depois -> VotacaoEncerrada
//	CE4 (inválida) fim antes ou igual ao início -> ErrJanelaInvalida
//	CE5 (inválida) início, fim ou agora com data zerada -> ErrJanelaInvalida
//
// Um fim zerado com início válido também fica antes do início, então cai em CE4 e CE5.
func TestStatusVotacaoClassesEquivalencia(t *testing.T) {
	verificarPeriodo(t, []casoPeriodo{
		{nome: "CE1: uma hora antes do início", inicio: inicioPadrao, fim: fimPadrao, agora: inicioPadrao.Add(-time.Hour), estadoEsperado: VotacaoNaoIniciada},
		{nome: "CE2: durante a votação", inicio: inicioPadrao, fim: fimPadrao, agora: inicioPadrao.Add(5 * time.Hour), estadoEsperado: VotacaoAberta},
		{nome: "CE2: mesmo instante em outro fuso horário", inicio: inicioPadrao, fim: fimPadrao, agora: inicioPadrao.In(time.FixedZone("BRT", -3*60*60)), estadoEsperado: VotacaoAberta},
		{nome: "CE3: um dia depois do fim", inicio: inicioPadrao, fim: fimPadrao, agora: fimPadrao.Add(24 * time.Hour), estadoEsperado: VotacaoEncerrada},
		{nome: "CE4: fim antes do início", inicio: fimPadrao, fim: inicioPadrao, agora: inicioPadrao, estadoEsperado: VotacaoInvalida, erroEsperado: ErrJanelaInvalida},
		{nome: "CE5: início zerado", fim: fimPadrao, agora: inicioPadrao, estadoEsperado: VotacaoInvalida, erroEsperado: ErrJanelaInvalida},
		{nome: "CE5: fim zerado", inicio: inicioPadrao, agora: inicioPadrao, estadoEsperado: VotacaoInvalida, erroEsperado: ErrJanelaInvalida},
		{nome: "CE5: instante atual zerado", inicio: inicioPadrao, fim: fimPadrao, estadoEsperado: VotacaoInvalida, erroEsperado: ErrJanelaInvalida},
	})
}

// Análise de valor limite de StatusVotacao:
//
//	VL1 instante de abertura: agora = início (fronteira entre CE1 e CE2)
//	VL2 instante de encerramento: agora = fim (fronteira entre CE2 e CE3)
//	VL3 duração da janela: fim = início (fronteira entre CE4 e as classes válidas)
func TestStatusVotacaoValorLimite(t *testing.T) {
	verificarPeriodo(t, []casoPeriodo{
		{nome: "VL1: 1ns antes do início", inicio: inicioPadrao, fim: fimPadrao, agora: inicioPadrao.Add(-time.Nanosecond), estadoEsperado: VotacaoNaoIniciada},
		{nome: "VL1: exatamente no início", inicio: inicioPadrao, fim: fimPadrao, agora: inicioPadrao, estadoEsperado: VotacaoAberta},
		{nome: "VL2: 1ns antes do fim", inicio: inicioPadrao, fim: fimPadrao, agora: fimPadrao.Add(-time.Nanosecond), estadoEsperado: VotacaoAberta},
		{nome: "VL2: exatamente no fim", inicio: inicioPadrao, fim: fimPadrao, agora: fimPadrao, estadoEsperado: VotacaoEncerrada},
		{nome: "VL3: fim igual ao início", inicio: inicioPadrao, fim: inicioPadrao, agora: inicioPadrao, estadoEsperado: VotacaoInvalida, erroEsperado: ErrJanelaInvalida},
		{nome: "VL3: fim 1ns depois do início", inicio: inicioPadrao, fim: inicioPadrao.Add(time.Nanosecond), agora: inicioPadrao, estadoEsperado: VotacaoAberta},
	})
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
