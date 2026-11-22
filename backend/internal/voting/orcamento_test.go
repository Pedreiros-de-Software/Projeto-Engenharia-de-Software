package voting

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func projeto(opcao string, pontos, primeiras int) OpcaoPontuada {
	return OpcaoPontuada{Opcao: opcao, Pontos: pontos, PrimeirasPreferencias: primeiras}
}

func orcado(p OpcaoPontuada, custo, acumulado int64, dentro bool) ProjetoOrcado {
	return ProjetoOrcado{OpcaoPontuada: p, Custo: custo, CustoAcumulado: acumulado, DentroDoOrcamento: dentro}
}

func TestAplicarCorteOrcamentario(t *testing.T) {
	a, b, c := projeto("A", 30, 3), projeto("B", 20, 2), projeto("C", 10, 1)
	empatadoA := OpcaoPontuada{Opcao: "A", Pontos: 10, PrimeirasPreferencias: 1, Empatado: true}
	empatadoB := OpcaoPontuada{Opcao: "B", Pontos: 10, PrimeirasPreferencias: 1, Empatado: true}
	maximo := int64(math.MaxInt64)

	testes := []struct {
		nome         string
		ranking      []OpcaoPontuada
		custos       map[string]int64
		orcamento    int64
		esperado     ResultadoOrcamento
		erroEsperado error
	}{
		{
			nome:      "Todos os projetos cabem no orçamento",
			ranking:   []OpcaoPontuada{a, b},
			custos:    map[string]int64{"A": 10, "B": 20},
			orcamento: 100,
			esperado: ResultadoOrcamento{
				Projetos:         []ProjetoOrcado{orcado(a, 10, 10, true), orcado(b, 20, 30, true)},
				Orcamento:        100,
				TotalSelecionado: 30,
				Saldo:            70,
			},
		},
		{
			nome:      "Custo acumulado igual ao orçamento inclui o último projeto",
			ranking:   []OpcaoPontuada{a, b},
			custos:    map[string]int64{"A": 40, "B": 60},
			orcamento: 100,
			esperado: ResultadoOrcamento{
				Projetos:         []ProjetoOrcado{orcado(a, 40, 40, true), orcado(b, 60, 100, true)},
				Orcamento:        100,
				TotalSelecionado: 100,
			},
		},
		{
			nome:      "Um centavo a menos deixa o último projeto fora",
			ranking:   []OpcaoPontuada{a, b},
			custos:    map[string]int64{"A": 40, "B": 60},
			orcamento: 99,
			esperado: ResultadoOrcamento{
				Projetos:         []ProjetoOrcado{orcado(a, 40, 40, true), orcado(b, 60, 100, false)},
				Orcamento:        99,
				TotalSelecionado: 40,
				Saldo:            59,
			},
		},
		{
			nome:      "Corte estrito deixa fora um projeto mais barato abaixo da linha",
			ranking:   []OpcaoPontuada{a, b, c},
			custos:    map[string]int64{"A": 50, "B": 60, "C": 10},
			orcamento: 100,
			esperado: ResultadoOrcamento{
				Projetos: []ProjetoOrcado{
					orcado(a, 50, 50, true),
					orcado(b, 60, 110, false),
					orcado(c, 10, 120, false),
				},
				Orcamento:        100,
				TotalSelecionado: 50,
				Saldo:            50,
			},
		},
		{
			nome:      "Orçamento zero deixa fora projetos com custo",
			ranking:   []OpcaoPontuada{a},
			custos:    map[string]int64{"A": 10},
			orcamento: 0,
			esperado: ResultadoOrcamento{
				Projetos: []ProjetoOrcado{orcado(a, 10, 10, false)},
			},
		},
		{
			nome:      "Orçamento zero aceita projeto sem custo no topo do ranking",
			ranking:   []OpcaoPontuada{a, b},
			custos:    map[string]int64{"A": 0, "B": 10},
			orcamento: 0,
			esperado: ResultadoOrcamento{
				Projetos: []ProjetoOrcado{orcado(a, 0, 0, true), orcado(b, 10, 10, false)},
			},
		},
		{
			nome:      "Primeiro projeto não cabe e nenhum fica dentro",
			ranking:   []OpcaoPontuada{empatadoA, empatadoB},
			custos:    map[string]int64{"A": 10, "B": 1},
			orcamento: 5,
			esperado: ResultadoOrcamento{
				Projetos:  []ProjetoOrcado{orcado(empatadoA, 10, 10, false), orcado(empatadoB, 1, 11, false)},
				Orcamento: 5,
				Saldo:     5,
			},
		},
		{
			nome:      "Empate na linha de corte é sinalizado",
			ranking:   []OpcaoPontuada{empatadoA, empatadoB},
			custos:    map[string]int64{"A": 50, "B": 60},
			orcamento: 100,
			esperado: ResultadoOrcamento{
				Projetos:             []ProjetoOrcado{orcado(empatadoA, 50, 50, true), orcado(empatadoB, 60, 110, false)},
				Orcamento:            100,
				TotalSelecionado:     50,
				Saldo:                50,
				EmpateNaLinhaDeCorte: true,
			},
		},
		{
			nome:      "Empate de pontos desfeito por primeiras preferências não sinaliza a linha de corte",
			ranking:   []OpcaoPontuada{projeto("A", 10, 2), projeto("B", 10, 1)},
			custos:    map[string]int64{"A": 50, "B": 60},
			orcamento: 100,
			esperado: ResultadoOrcamento{
				Projetos: []ProjetoOrcado{
					orcado(projeto("A", 10, 2), 50, 50, true),
					orcado(projeto("B", 10, 1), 60, 110, false),
				},
				Orcamento:        100,
				TotalSelecionado: 50,
				Saldo:            50,
			},
		},
		{
			nome:      "Empate entre projetos dentro do orçamento não sinaliza a linha de corte",
			ranking:   []OpcaoPontuada{empatadoA, empatadoB, projeto("D", 5, 0)},
			custos:    map[string]int64{"A": 10, "B": 10, "D": 100},
			orcamento: 50,
			esperado: ResultadoOrcamento{
				Projetos: []ProjetoOrcado{
					orcado(empatadoA, 10, 10, true),
					orcado(empatadoB, 10, 20, true),
					orcado(projeto("D", 5, 0), 100, 120, false),
				},
				Orcamento:        50,
				TotalSelecionado: 20,
				Saldo:            30,
			},
		},
		{
			nome:      "Empate entre projetos fora do orçamento não sinaliza a linha de corte",
			ranking:   []OpcaoPontuada{projeto("L", 12, 2), empatadoA, empatadoB},
			custos:    map[string]int64{"L": 10, "A": 10, "B": 1},
			orcamento: 15,
			esperado: ResultadoOrcamento{
				Projetos: []ProjetoOrcado{
					orcado(projeto("L", 12, 2), 10, 10, true),
					orcado(empatadoA, 10, 20, false),
					orcado(empatadoB, 1, 21, false),
				},
				Orcamento:        15,
				TotalSelecionado: 10,
				Saldo:            5,
			},
		},
		{
			nome:      "Ranking vazio mantém todo o orçamento como saldo",
			ranking:   nil,
			custos:    map[string]int64{},
			orcamento: 100,
			esperado: ResultadoOrcamento{
				Projetos:  []ProjetoOrcado{},
				Orcamento: 100,
				Saldo:     100,
			},
		},
		{
			nome:      "Custo máximo cabe no orçamento máximo",
			ranking:   []OpcaoPontuada{a},
			custos:    map[string]int64{"A": maximo},
			orcamento: maximo,
			esperado: ResultadoOrcamento{
				Projetos:         []ProjetoOrcado{orcado(a, maximo, maximo, true)},
				Orcamento:        maximo,
				TotalSelecionado: maximo,
			},
		},
		{
			nome:         "Orçamento negativo",
			ranking:      []OpcaoPontuada{a},
			custos:       map[string]int64{"A": 10},
			orcamento:    -1,
			erroEsperado: ErrOrcamentoInvalido,
		},
		{
			nome:         "Custo ausente",
			ranking:      []OpcaoPontuada{a, b},
			custos:       map[string]int64{"A": 10},
			orcamento:    100,
			erroEsperado: ErrCustoInvalido,
		},
		{
			nome:         "Mapa de custos ausente",
			ranking:      []OpcaoPontuada{a},
			custos:       nil,
			orcamento:    100,
			erroEsperado: ErrCustoInvalido,
		},
		{
			nome:         "Custo negativo",
			ranking:      []OpcaoPontuada{a, b},
			custos:       map[string]int64{"A": 10, "B": -1},
			orcamento:    100,
			erroEsperado: ErrCustoInvalido,
		},
		{
			nome:         "Soma dos custos acima do limite de int64",
			ranking:      []OpcaoPontuada{a, b},
			custos:       map[string]int64{"A": maximo, "B": 1},
			orcamento:    maximo,
			erroEsperado: ErrCustoInvalido,
		},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			obtido, err := AplicarCorteOrcamentario(tt.ranking, tt.custos, tt.orcamento)
			if !errors.Is(err, tt.erroEsperado) {
				t.Fatalf("erro: esperado %v, obtido %v", tt.erroEsperado, err)
			}
			if !reflect.DeepEqual(obtido, tt.esperado) {
				t.Errorf("resultado: esperado %+v, obtido %+v", tt.esperado, obtido)
			}
		})
	}
}

// Consulta do seed: 5 cédulas sobre 4 projetos e orçamento de R$ 60.000,00.
func TestConsultaDoSeed(t *testing.T) {
	const (
		impressoras = "Impressoras 3D e CNC"
		arduino     = "Kit Arduino e Robótica"
		gpu         = "Servidor GPU para IA"
		mobiliario  = "Mobiliário colaborativo"
	)
	projetos := []string{impressoras, arduino, gpu, mobiliario}
	cedulas := []VotoBorda{
		{EleitorID: "ana", Preferencias: []string{impressoras, arduino, gpu, mobiliario}},
		{EleitorID: "bruno", Preferencias: []string{gpu, impressoras, arduino, mobiliario}},
		{EleitorID: "carla", Preferencias: []string{arduino, impressoras, mobiliario, gpu}},
		{EleitorID: "maria", Preferencias: []string{impressoras, mobiliario, arduino, gpu}},
		{EleitorID: "joao", Preferencias: []string{gpu, arduino, impressoras, mobiliario}},
	}
	custos := map[string]int64{impressoras: 35000_00, arduino: 12000_00, gpu: 48000_00, mobiliario: 9000_00}

	ranking, err := CalcularApuracaoBorda(projetos, cedulas)
	if err != nil {
		t.Fatalf("erro inesperado na apuracao Borda: %v", err)
	}
	rankingEsperado := []OpcaoPontuada{
		projeto(impressoras, 16, 2),
		projeto(arduino, 14, 1),
		projeto(gpu, 12, 2),
		projeto(mobiliario, 8, 0),
	}
	if !reflect.DeepEqual(ranking, rankingEsperado) {
		t.Fatalf("ranking: esperado %+v, obtido %+v", rankingEsperado, ranking)
	}

	obtido, err := AplicarCorteOrcamentario(ranking, custos, 60000_00)
	if err != nil {
		t.Fatalf("erro inesperado no corte orcamentario: %v", err)
	}
	esperado := ResultadoOrcamento{
		Projetos: []ProjetoOrcado{
			orcado(rankingEsperado[0], 35000_00, 35000_00, true),
			orcado(rankingEsperado[1], 12000_00, 47000_00, true),
			orcado(rankingEsperado[2], 48000_00, 95000_00, false),
			orcado(rankingEsperado[3], 9000_00, 104000_00, false),
		},
		Orcamento:        60000_00,
		TotalSelecionado: 47000_00,
		Saldo:            13000_00,
	}
	if !reflect.DeepEqual(obtido, esperado) {
		t.Errorf("corte orcamentario: esperado %+v, obtido %+v", esperado, obtido)
	}
}
