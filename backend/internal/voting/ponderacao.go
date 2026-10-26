package voting

import (
	"errors"
	"fmt"
)

type Voto struct {
	EleitorID string
	Opcao     string
}

type Ponderacao interface {
	Peso(eleitorID string) (float64, error)
}

type PesoUnitario struct{}

func (PesoUnitario) Peso(string) (float64, error) {
	return 1, nil
}

type PesoPorUnidade struct {
	fracoesPorEleitor map[string]float64
}

func NovoPesoPorUnidade(condominos []Condomino) *PesoPorUnidade {
	fracoesPorEleitor := make(map[string]float64, len(condominos))
	for _, condomino := range condominos {
		fracoesPorEleitor[condomino.ID] = condomino.FracaoIdeal
	}
	return &PesoPorUnidade{fracoesPorEleitor: fracoesPorEleitor}
}

func (p *PesoPorUnidade) Peso(eleitorID string) (float64, error) {
	peso, encontrado := p.fracoesPorEleitor[eleitorID]
	if !encontrado {
		return 0, ErrEleitorDesconhecido
	}
	return peso, nil
}

var (
	ErrEleitorDesconhecido = errors.New("eleitor desconhecido")
	ErrOpcaoDesconhecida   = errors.New("opcao desconhecida")
)

func ContarVotos(opcoes []string, votos []Voto, p Ponderacao) (map[string]float64, error) {
	resultados := make(map[string]float64, len(opcoes))
	for _, opcao := range opcoes {
		resultados[opcao] = 0
	}

	for _, voto := range votos {
		if _, encontrada := resultados[voto.Opcao]; !encontrada {
			return nil, ErrOpcaoDesconhecida
		}

		peso, err := p.Peso(voto.EleitorID)
		if err != nil {
			return nil, fmt.Errorf("obtendo peso do eleitor %q: %w", voto.EleitorID, err)
		}
		resultados[voto.Opcao] += peso
	}

	return resultados, nil
}
