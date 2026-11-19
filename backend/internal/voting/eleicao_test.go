package voting

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

// votosEm cria uma cédula anônima para cada opção informada.
func votosEm(opcoes ...string) []Voto {
	votos := make([]Voto, len(opcoes))
	for i, opcao := range opcoes {
		votos[i] = Voto{Opcao: opcao}
	}
	return votos
}

func TestApurarEleicao(t *testing.T) {
	const (
		a = "Chapa A"
		b = "Chapa B"
		c = "Chapa C"
		d = "Chapa D"
	)
	tresChapas := []string{a, b, c}

	testes := []struct {
		nome         string
		chapas       []string
		votos        []Voto
		esperado     ResultadoEleicao
		erroEsperado error
	}{
		{
			nome:   "Vencedora com um voto a mais que as demais somadas",
			chapas: tresChapas,
			votos:  votosEm(a, a, a, b, c, OpcaoBranco, OpcaoBranco, OpcaoNulo),
			esperado: ResultadoEleicao{
				Chapas:    []VotosOpcao{{a, 3}, {b, 1}, {c, 1}},
				Brancos:   2,
				Nulos:     1,
				Validos:   5,
				Maioria:   ResultadoSegundoTurno{VencedorDefinido: true, PercentualLider: 60},
				Vencedora: a,
			},
		},
		{
			nome:   "Brancos e nulos não entram na base da maioria absoluta",
			chapas: []string{a, b},
			votos:  votosEm(a, a, a, b, OpcaoBranco, OpcaoBranco, OpcaoNulo, OpcaoNulo),
			esperado: ResultadoEleicao{
				Chapas:    []VotosOpcao{{a, 3}, {b, 1}},
				Brancos:   2,
				Nulos:     2,
				Validos:   4,
				Maioria:   ResultadoSegundoTurno{VencedorDefinido: true, PercentualLider: 75},
				Vencedora: a,
			},
		},
		{
			nome:   "Metade dos válidos exige segundo turno entre as duas mais votadas",
			chapas: tresChapas,
			votos:  votosEm(a, a, a, b, b, c, OpcaoBranco),
			esperado: ResultadoEleicao{
				Chapas:       []VotosOpcao{{a, 3}, {b, 2}, {c, 1}},
				Brancos:      1,
				Validos:      6,
				Maioria:      ResultadoSegundoTurno{PrecisaSegundoTurno: true, PercentualLider: 50},
				SegundoTurno: []string{a, b},
			},
		},
		{
			nome:   "Chapas ordenadas por votos e não pela ordem de cadastro",
			chapas: tresChapas,
			votos:  votosEm(a, b, b, c, c, c),
			esperado: ResultadoEleicao{
				Chapas:       []VotosOpcao{{c, 3}, {b, 2}, {a, 1}},
				Validos:      6,
				Maioria:      ResultadoSegundoTurno{PrecisaSegundoTurno: true, PercentualLider: 50},
				SegundoTurno: []string{c, b},
			},
		},
		{
			nome:   "Empate pelo primeiro lugar leva as empatadas ao segundo turno na ordem de cadastro",
			chapas: tresChapas,
			votos:  votosEm(b, b, a, a, c),
			esperado: ResultadoEleicao{
				Chapas:       []VotosOpcao{{a, 2}, {b, 2}, {c, 1}},
				Validos:      5,
				Maioria:      ResultadoSegundoTurno{PrecisaSegundoTurno: true, PercentualLider: 40},
				SegundoTurno: []string{a, b},
			},
		},
		{
			nome:   "Duas chapas empatadas vão ao segundo turno",
			chapas: []string{a, b},
			votos:  votosEm(b, a),
			esperado: ResultadoEleicao{
				Chapas:       []VotosOpcao{{a, 1}, {b, 1}},
				Validos:      2,
				Maioria:      ResultadoSegundoTurno{PrecisaSegundoTurno: true, PercentualLider: 50},
				SegundoTurno: []string{a, b},
			},
		},
		{
			nome:   "Empate pelo segundo lugar impede definir o segundo turno",
			chapas: tresChapas,
			votos:  votosEm(a, a, b, c),
			esperado: ResultadoEleicao{
				Chapas:           []VotosOpcao{{a, 2}, {b, 1}, {c, 1}},
				Validos:          4,
				Maioria:          ResultadoSegundoTurno{PrecisaSegundoTurno: true, PercentualLider: 50},
				EmpateIndefinido: true,
			},
		},
		{
			nome:   "Empate entre todas as chapas impede definir o segundo turno",
			chapas: []string{a, b, c, d},
			votos:  votosEm(d, c, b, a),
			esperado: ResultadoEleicao{
				Chapas:           []VotosOpcao{{a, 1}, {b, 1}, {c, 1}, {d, 1}},
				Validos:          4,
				Maioria:          ResultadoSegundoTurno{PrecisaSegundoTurno: true, PercentualLider: 25},
				EmpateIndefinido: true,
			},
		},
		{
			nome:   "Sem votos não há vencedora nem segundo turno",
			chapas: []string{a, b},
			esperado: ResultadoEleicao{
				Chapas: []VotosOpcao{{a, 0}, {b, 0}},
			},
		},
		{
			nome:   "Somente brancos e nulos não há vencedora nem segundo turno",
			chapas: []string{a, b},
			votos:  votosEm(OpcaoBranco, OpcaoNulo, OpcaoBranco),
			esperado: ResultadoEleicao{
				Chapas:  []VotosOpcao{{a, 0}, {b, 0}},
				Brancos: 2,
				Nulos:   1,
			},
		},
		{
			nome:   "Chapa única vence com qualquer voto válido",
			chapas: []string{a},
			votos:  votosEm(OpcaoBranco, a, OpcaoBranco, OpcaoNulo),
			esperado: ResultadoEleicao{
				Chapas:    []VotosOpcao{{a, 1}},
				Brancos:   2,
				Nulos:     1,
				Validos:   1,
				Maioria:   ResultadoSegundoTurno{VencedorDefinido: true, PercentualLider: 100},
				Vencedora: a,
			},
		},
		{
			nome:   "Chapa única sem votos válidos não vence",
			chapas: []string{a},
			votos:  votosEm(OpcaoNulo),
			esperado: ResultadoEleicao{
				Chapas: []VotosOpcao{{a, 0}},
				Nulos:  1,
			},
		},
		{
			nome:         "Voto em opção desconhecida",
			chapas:       []string{a, b},
			votos:        votosEm(a, "Chapa Z"),
			erroEsperado: ErrOpcaoDesconhecida,
		},
		{
			nome:         "Sem chapas",
			chapas:       nil,
			votos:        votosEm(OpcaoBranco),
			erroEsperado: ErrOpcoesVazias,
		},
		{
			nome:         "Lista de chapas vazia",
			chapas:       []string{},
			erroEsperado: ErrOpcoesVazias,
		},
		{
			nome:         "Chapa repetida",
			chapas:       []string{a, b, a},
			erroEsperado: ErrOpcoesDuplicadas,
		},
		{
			nome:         "Chapa com o nome reservado do voto branco",
			chapas:       []string{a, OpcaoBranco},
			erroEsperado: ErrOpcoesDuplicadas,
		},
		{
			nome:         "Chapa com o nome reservado do voto nulo",
			chapas:       []string{OpcaoNulo, a},
			erroEsperado: ErrOpcoesDuplicadas,
		},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			obtido, err := ApurarEleicao(tt.chapas, tt.votos)
			if !errors.Is(err, tt.erroEsperado) {
				t.Fatalf("erro: esperado %v, obtido %v", tt.erroEsperado, err)
			}
			if !reflect.DeepEqual(obtido, tt.esperado) {
				t.Errorf("resultado: esperado %+v, obtido %+v", tt.esperado, obtido)
			}
		})
	}
}

func TestApurarEleicaoNaoAlteraAsChapasRecebidas(t *testing.T) {
	espaco := make([]string, 4)
	chapas := espaco[:2]
	chapas[0], chapas[1] = "Chapa A", "Chapa B"

	if _, err := ApurarEleicao(chapas, votosEm("Chapa A")); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if espaco[2] != "" || espaco[3] != "" {
		t.Errorf("a apuracao escreveu na lista de chapas do chamador: %q", espaco)
	}
}

func TestMontarBoletim(t *testing.T) {
	apurado := ResultadoEleicao{
		Chapas:  []VotosOpcao{{"Chapa A", 3}, {"Chapa B", 1}},
		Brancos: 3,
		Nulos:   1,
		Validos: 4,
	}

	testes := []struct {
		nome          string
		aptos         int
		participantes int
		resultado     ResultadoEleicao
		esperado      Boletim
		erroEsperado  error
	}{
		{
			nome:          "Comparecimento parcial com percentuais sobre os válidos",
			aptos:         10,
			participantes: 8,
			resultado:     apurado,
			esperado: Boletim{
				Aptos:                    10,
				Participantes:            8,
				Abstencoes:               2,
				PercentualComparecimento: 80,
				PercentuaisValidos:       map[string]float64{"Chapa A": 75, "Chapa B": 25},
				Resultado:                apurado,
			},
		},
		{
			nome:          "Todos os aptos participaram",
			aptos:         4,
			participantes: 4,
			resultado:     ResultadoEleicao{Chapas: []VotosOpcao{{"Chapa A", 4}, {"Chapa B", 0}}, Validos: 4},
			esperado: Boletim{
				Aptos:                    4,
				Participantes:            4,
				PercentualComparecimento: 100,
				PercentuaisValidos:       map[string]float64{"Chapa A": 100, "Chapa B": 0},
				Resultado:                ResultadoEleicao{Chapas: []VotosOpcao{{"Chapa A", 4}, {"Chapa B", 0}}, Validos: 4},
			},
		},
		{
			nome:          "Nenhum apto participou",
			aptos:         5,
			participantes: 0,
			resultado:     ResultadoEleicao{Chapas: []VotosOpcao{{"Chapa A", 0}, {"Chapa B", 0}}},
			esperado: Boletim{
				Aptos:              5,
				Abstencoes:         5,
				PercentuaisValidos: map[string]float64{"Chapa A": 0, "Chapa B": 0},
				Resultado:          ResultadoEleicao{Chapas: []VotosOpcao{{"Chapa A", 0}, {"Chapa B", 0}}},
			},
		},
		{
			nome:          "Votação sem aptos tem percentuais zerados",
			aptos:         0,
			participantes: 0,
			resultado:     ResultadoEleicao{Chapas: []VotosOpcao{{"Chapa A", 0}}},
			esperado: Boletim{
				PercentuaisValidos: map[string]float64{"Chapa A": 0},
				Resultado:          ResultadoEleicao{Chapas: []VotosOpcao{{"Chapa A", 0}}},
			},
		},
		{
			nome:          "Somente brancos e nulos zeram os percentuais das chapas",
			aptos:         3,
			participantes: 3,
			resultado:     ResultadoEleicao{Chapas: []VotosOpcao{{"Chapa A", 0}}, Brancos: 2, Nulos: 1},
			esperado: Boletim{
				Aptos:                    3,
				Participantes:            3,
				PercentualComparecimento: 100,
				PercentuaisValidos:       map[string]float64{"Chapa A": 0},
				Resultado:                ResultadoEleicao{Chapas: []VotosOpcao{{"Chapa A", 0}}, Brancos: 2, Nulos: 1},
			},
		},
		{
			nome:          "Aptos negativos",
			aptos:         -1,
			participantes: 0,
			erroEsperado:  ErrTotaisInconsistentes,
		},
		{
			nome:          "Participantes negativos",
			aptos:         5,
			participantes: -1,
			erroEsperado:  ErrTotaisInconsistentes,
		},
		{
			nome:          "Mais participantes que aptos",
			aptos:         7,
			participantes: 8,
			resultado:     apurado,
			erroEsperado:  ErrTotaisInconsistentes,
		},
		{
			nome:          "Participantes acima do total de cédulas",
			aptos:         10,
			participantes: 9,
			resultado:     apurado,
			erroEsperado:  ErrTotaisInconsistentes,
		},
		{
			nome:          "Participantes abaixo do total de cédulas",
			aptos:         10,
			participantes: 7,
			resultado:     apurado,
			erroEsperado:  ErrTotaisInconsistentes,
		},
		{
			nome:          "Brancos negativos",
			aptos:         3,
			participantes: 3,
			resultado:     ResultadoEleicao{Chapas: []VotosOpcao{{"Chapa A", 3}}, Brancos: -1, Nulos: 1, Validos: 3},
			erroEsperado:  ErrTotaisInconsistentes,
		},
		{
			nome:          "Nulos negativos",
			aptos:         3,
			participantes: 3,
			resultado:     ResultadoEleicao{Chapas: []VotosOpcao{{"Chapa A", 3}}, Brancos: 1, Nulos: -1, Validos: 3},
			erroEsperado:  ErrTotaisInconsistentes,
		},
		{
			nome:          "Chapa com votos negativos",
			aptos:         3,
			participantes: 3,
			resultado:     ResultadoEleicao{Chapas: []VotosOpcao{{"Chapa A", 4}, {"Chapa B", -1}}, Validos: 3},
			erroEsperado:  ErrTotaisInconsistentes,
		},
		{
			nome:          "Válidos diferentes da soma das chapas",
			aptos:         4,
			participantes: 4,
			resultado:     ResultadoEleicao{Chapas: []VotosOpcao{{"Chapa A", 3}}, Validos: 4},
			erroEsperado:  ErrTotaisInconsistentes,
		},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			obtido, err := MontarBoletim(tt.aptos, tt.participantes, tt.resultado)
			if !errors.Is(err, tt.erroEsperado) {
				t.Fatalf("erro: esperado %v, obtido %v", tt.erroEsperado, err)
			}
			if !reflect.DeepEqual(obtido, tt.esperado) {
				t.Errorf("boletim: esperado %+v, obtido %+v", tt.esperado, obtido)
			}
		})
	}
}

// Eleição do seed: 8 aptos; C1 = 3, C2 = 2, C3 = 1, 1 branco e 1 abstenção.
func TestBoletimDaEleicaoDoSeed(t *testing.T) {
	const (
		c1 = "Chapa 1 — Movimenta"
		c2 = "Chapa 2 — Renovação"
		c3 = "Chapa 3 — Integração"
	)

	resultado, err := ApurarEleicao([]string{c1, c2, c3}, votosEm(c1, c1, c1, c2, c2, c3, OpcaoBranco))
	if err != nil {
		t.Fatalf("erro inesperado na apuracao: %v", err)
	}
	if resultado.Vencedora != "" || !reflect.DeepEqual(resultado.SegundoTurno, []string{c1, c2}) {
		t.Errorf("esperado 2o turno entre %q e %q sem vencedora, obtido %+v", c1, c2, resultado)
	}

	boletim, err := MontarBoletim(8, 7, resultado)
	if err != nil {
		t.Fatalf("erro inesperado no boletim: %v", err)
	}
	if boletim.Abstencoes != 1 || boletim.PercentualComparecimento != 87.5 {
		t.Errorf("esperado 1 abstencao e 87,5%% de comparecimento, obtido %+v", boletim)
	}

	percentuaisEsperados := map[string]float64{c1: 50, c2: 100.0 / 3, c3: 100.0 / 6}
	for chapa, esperado := range percentuaisEsperados {
		if obtido := boletim.PercentuaisValidos[chapa]; math.Abs(obtido-esperado) > 1e-9 {
			t.Errorf("percentual de %q: esperado %v, obtido %v", chapa, esperado, obtido)
		}
	}
}
