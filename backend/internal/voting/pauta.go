package voting

import (
	"errors"
	"fmt"
)

const (
	OpcaoFavor     = "favor"
	OpcaoContra    = "contra"
	OpcaoAbstencao = "abstencao"
)

var ErrVotoDuplicado = errors.New("unidade com mais de um voto na pauta")

// ResultadoPauta reúne o resultado da pauta e os valores usados na apuração.
type ResultadoPauta struct {
	Contagem  map[string]float64
	PesoTotal float64
	Aprovada  bool
}

// ApurarPauta apura uma pauta de condomínio pelo quórum qualificado de 2/3 do
// peso total cadastrado, que inclui as unidades que não votaram. Votos e peso
// total são somados na ordem do cadastro, de modo que o arredondamento nunca
// torne o peso favorável maior que o total.
func ApurarPauta(condominos []Condomino, votos []Voto) (ResultadoPauta, error) {
	votoPorUnidade := make(map[string]Voto, len(votos))
	for _, v := range votos {
		if _, repetido := votoPorUnidade[v.EleitorID]; repetido {
			return ResultadoPauta{}, fmt.Errorf("unidade %q: %w", v.EleitorID, ErrVotoDuplicado)
		}
		votoPorUnidade[v.EleitorID] = v
	}

	ordenados := make([]Voto, 0, len(votos))
	var pesoTotal float64
	for _, c := range condominos {
		pesoTotal += c.FracaoIdeal
		if v, votou := votoPorUnidade[c.ID]; votou {
			ordenados = append(ordenados, v)
			delete(votoPorUnidade, c.ID)
		}
	}
	for _, v := range votos {
		if _, desconhecido := votoPorUnidade[v.EleitorID]; desconhecido {
			return ResultadoPauta{}, fmt.Errorf("obtendo peso do eleitor %q: %w", v.EleitorID, ErrEleitorDesconhecido)
		}
	}

	opcoes := []string{OpcaoFavor, OpcaoContra, OpcaoAbstencao}
	contagem, err := ContarVotos(opcoes, ordenados, NovoPesoPorUnidade(condominos))
	if err != nil {
		return ResultadoPauta{}, err
	}

	return ResultadoPauta{
		Contagem:  contagem,
		PesoTotal: pesoTotal,
		Aprovada:  AtingiuQuorum(contagem[OpcaoFavor], pesoTotal),
	}, nil
}
