package voting

import (
	"errors"
	"reflect"
	"testing"
)

func TestContarVotos(t *testing.T) {
	testes := []struct {
		nome         string
		opcoes       []string
		votos        []Voto
		ponderacao   Ponderacao
		resultado    map[string]float64
		erroEsperado error
	}{
		{
			nome:   "Peso unitario",
			opcoes: []string{"sim", "nao"},
			votos: []Voto{
				{EleitorID: "1", Opcao: "sim"},
				{EleitorID: "2", Opcao: "sim"},
				{EleitorID: "3", Opcao: "nao"},
			},
			ponderacao: PesoUnitario{},
			resultado:  map[string]float64{"sim": 2, "nao": 1},
		},
		{
			nome:   "Peso por unidade",
			opcoes: []string{"sim", "nao"},
			votos: []Voto{
				{EleitorID: "1", Opcao: "sim"},
				{EleitorID: "2", Opcao: "sim"},
				{EleitorID: "3", Opcao: "nao"},
			},
			ponderacao: NovoPesoPorUnidade([]Condomino{
				{ID: "1", FracaoIdeal: 0.25},
				{ID: "2", FracaoIdeal: 0.5},
				{ID: "3", FracaoIdeal: 0.75},
			}),
			resultado: map[string]float64{"sim": 0.75, "nao": 0.75},
		},
		{
			nome:       "Opcao sem votos permanece zerada",
			opcoes:     []string{"sim", "nao"},
			votos:      []Voto{{EleitorID: "1", Opcao: "sim"}},
			ponderacao: PesoUnitario{},
			resultado:  map[string]float64{"sim": 1, "nao": 0},
		},
		{
			nome:         "Opcao desconhecida",
			opcoes:       []string{"sim"},
			votos:        []Voto{{EleitorID: "1", Opcao: "abster"}},
			ponderacao:   PesoUnitario{},
			erroEsperado: ErrOpcaoDesconhecida,
		},
		{
			nome:         "Eleitor desconhecido",
			opcoes:       []string{"sim"},
			votos:        []Voto{{EleitorID: "desconhecido", Opcao: "sim"}},
			ponderacao:   NovoPesoPorUnidade(nil),
			erroEsperado: ErrEleitorDesconhecido,
		},
		{
			nome:       "Lista de votos vazia",
			opcoes:     []string{"sim", "nao"},
			ponderacao: PesoUnitario{},
			resultado:  map[string]float64{"sim": 0, "nao": 0},
		},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			resultado, err := ContarVotos(tt.opcoes, tt.votos, tt.ponderacao)
			if !errors.Is(err, tt.erroEsperado) {
				t.Fatalf("erro: esperado %v, obtido %v", tt.erroEsperado, err)
			}
			if tt.erroEsperado != nil {
				if resultado != nil {
					t.Fatalf("resultado: esperado nil em caso de erro, obtido %v", resultado)
				}
				return
			}
			if !reflect.DeepEqual(resultado, tt.resultado) {
				t.Errorf("resultado: esperado %v, obtido %v", tt.resultado, resultado)
			}
		})
	}
}

func TestContarVotosPreservaErroDaPonderacao(t *testing.T) {
	erroPonderacao := errors.New("falha da estrategia")
	ponderacao := ponderacaoComErro{err: erroPonderacao}

	resultado, err := ContarVotos(
		[]string{"sim"},
		[]Voto{{EleitorID: "1", Opcao: "sim"}},
		ponderacao,
	)
	if !errors.Is(err, erroPonderacao) {
		t.Fatalf("erro: esperado causa %v, obtido %v", erroPonderacao, err)
	}
	if resultado != nil {
		t.Fatalf("resultado: esperado nil em caso de erro, obtido %v", resultado)
	}
}

type ponderacaoComErro struct {
	err error
}

func (p ponderacaoComErro) Peso(string) (float64, error) {
	return 0, p.err
}
