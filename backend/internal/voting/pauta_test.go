package voting

import (
	"errors"
	"reflect"
	"testing"
)

func TestApurarPauta(t *testing.T) {
	tresUnidades := []Condomino{
		{ID: "x", Adimplente: true, FracaoIdeal: 1},
		{ID: "y", Adimplente: true, FracaoIdeal: 1},
		{ID: "z", Adimplente: true, FracaoIdeal: 1},
	}
	unidadeMaior := []Condomino{
		{ID: "x", Adimplente: true, FracaoIdeal: 2},
		{ID: "y", Adimplente: true, FracaoIdeal: 1},
		{ID: "z", Adimplente: true, FracaoIdeal: 1},
	}

	testes := []struct {
		nome         string
		condominos   []Condomino
		votos        []Voto
		resultado    ResultadoPauta
		erroEsperado error
	}{
		{
			nome:       "Aprovada com exatamente dois tercos",
			condominos: tresUnidades,
			votos: []Voto{
				{EleitorID: "x", Opcao: OpcaoFavor},
				{EleitorID: "y", Opcao: OpcaoFavor},
				{EleitorID: "z", Opcao: OpcaoContra},
			},
			resultado: ResultadoPauta{
				Contagem:  map[string]float64{OpcaoFavor: 2, OpcaoContra: 1, OpcaoAbstencao: 0},
				PesoTotal: 3,
				Aprovada:  true,
			},
		},
		{
			nome:       "Reprovada abaixo de dois tercos",
			condominos: tresUnidades,
			votos: []Voto{
				{EleitorID: "x", Opcao: OpcaoFavor},
				{EleitorID: "y", Opcao: OpcaoContra},
				{EleitorID: "z", Opcao: OpcaoAbstencao},
			},
			resultado: ResultadoPauta{
				Contagem:  map[string]float64{OpcaoFavor: 1, OpcaoContra: 1, OpcaoAbstencao: 1},
				PesoTotal: 3,
			},
		},
		{
			nome:       "Unidade ausente conta no peso total",
			condominos: unidadeMaior,
			votos: []Voto{
				{EleitorID: "x", Opcao: OpcaoFavor},
				{EleitorID: "y", Opcao: OpcaoContra},
			},
			resultado: ResultadoPauta{
				Contagem:  map[string]float64{OpcaoFavor: 2, OpcaoContra: 1, OpcaoAbstencao: 0},
				PesoTotal: 4,
			},
		},
		{
			nome:       "Abstencao nao conta como favoravel",
			condominos: unidadeMaior,
			votos: []Voto{
				{EleitorID: "x", Opcao: OpcaoFavor},
				{EleitorID: "y", Opcao: OpcaoAbstencao},
				{EleitorID: "z", Opcao: OpcaoContra},
			},
			resultado: ResultadoPauta{
				Contagem:  map[string]float64{OpcaoFavor: 2, OpcaoContra: 1, OpcaoAbstencao: 1},
				PesoTotal: 4,
			},
		},
		{
			nome:       "Sem votos",
			condominos: tresUnidades,
			resultado: ResultadoPauta{
				Contagem:  map[string]float64{OpcaoFavor: 0, OpcaoContra: 0, OpcaoAbstencao: 0},
				PesoTotal: 3,
			},
		},
		{
			nome:       "Voto duplicado da mesma unidade",
			condominos: tresUnidades,
			votos: []Voto{
				{EleitorID: "x", Opcao: OpcaoFavor},
				{EleitorID: "x", Opcao: OpcaoContra},
			},
			erroEsperado: ErrVotoDuplicado,
		},
		{
			nome:         "Eleitor desconhecido",
			condominos:   tresUnidades,
			votos:        []Voto{{EleitorID: "w", Opcao: OpcaoFavor}},
			erroEsperado: ErrEleitorDesconhecido,
		},
		{
			nome:         "Opcao desconhecida",
			condominos:   tresUnidades,
			votos:        []Voto{{EleitorID: "x", Opcao: "sim"}},
			erroEsperado: ErrOpcaoDesconhecida,
		},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			resultado, err := ApurarPauta(tt.condominos, tt.votos)
			if !errors.Is(err, tt.erroEsperado) {
				t.Fatalf("erro: esperado %v, obtido %v", tt.erroEsperado, err)
			}
			if !reflect.DeepEqual(resultado, tt.resultado) {
				t.Errorf("resultado: esperado %+v, obtido %+v", tt.resultado, resultado)
			}
		})
	}
}

func TestApurarPautaUnanimeComVotosForaDaOrdemDoCadastro(t *testing.T) {
	condominos := []Condomino{
		{ID: "a", Adimplente: true, FracaoIdeal: 0.1},
		{ID: "b", Adimplente: true, FracaoIdeal: 0.1},
		{ID: "c", Adimplente: true, FracaoIdeal: 0.1},
		{ID: "d", Adimplente: true, FracaoIdeal: 4.55},
		{ID: "e", Adimplente: true, FracaoIdeal: 0.1},
	}
	// Somados nesta ordem, os pesos dão 4.95; na ordem do cadastro, 4.949999999999999.
	votos := []Voto{
		{EleitorID: "a", Opcao: OpcaoFavor},
		{EleitorID: "b", Opcao: OpcaoFavor},
		{EleitorID: "c", Opcao: OpcaoFavor},
		{EleitorID: "e", Opcao: OpcaoFavor},
		{EleitorID: "d", Opcao: OpcaoFavor},
	}

	resultado, err := ApurarPauta(condominos, votos)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resultado.Contagem[OpcaoFavor] != resultado.PesoTotal {
		t.Errorf("peso favoravel %v diferente do total %v", resultado.Contagem[OpcaoFavor], resultado.PesoTotal)
	}
	if !resultado.Aprovada {
		t.Errorf("votacao unanime reprovada: %+v", resultado)
	}
}
