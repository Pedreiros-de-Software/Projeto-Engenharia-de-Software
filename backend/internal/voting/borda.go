package voting

import (
	"errors"
	"sort"
)

// OpcaoPontuada representa o resultado de cada projeto/opção na apuração Borda.
// Empatado indica que a opção divide pontos e primeiras preferências com outra.
type OpcaoPontuada struct {
	Opcao                 string
	Pontos                int
	PrimeirasPreferencias int
	Empatado              bool
}

// VotoBorda representa a cédula de um eleitor com sua lista ordenada de preferências.
type VotoBorda struct {
	EleitorID    string
	Preferencias []string
}

var (
	ErrOpcoesVazias     = errors.New("a lista de opcoes da votacao nao pode ser vazia")
	ErrOpcoesDuplicadas = errors.New("a lista de opcoes contem itens repetidos")
	ErrVotoInvalido     = errors.New("o voto contem opcoes invalidas, duplicadas ou lista incompleta")
)

// ValidarCedulaBorda exige que a cédula ordene cada opção válida exatamente uma vez.
func ValidarCedulaBorda(opcoesValidas []string, preferencias []string) error {
	indice, err := indexarOpcoes(opcoesValidas)
	if err != nil {
		return err
	}
	return validarPreferencias(preferencias, indice)
}

// CalcularApuracaoBorda executa a contagem pelo Método de Borda: com N opções, a 1ª posição
// vale N pontos e a última vale 1. O ranking ordena por pontos e depois por primeiras
// preferências; empate persistente mantém a ordem de cadastro e marca Empatado.
func CalcularApuracaoBorda(opcoesValidas []string, votos []VotoBorda) ([]OpcaoPontuada, error) {
	indice, err := indexarOpcoes(opcoesValidas)
	if err != nil {
		return nil, err
	}

	for _, voto := range votos {
		if err := validarPreferencias(voto.Preferencias, indice); err != nil {
			return nil, err
		}
	}

	ranking := pontuarOpcoes(opcoesValidas, indice, votos)
	ordenarRanking(ranking)
	marcarEmpates(ranking)
	return ranking, nil
}

// indexarOpcoes devolve a posição de cadastro de cada opção, rejeitando lista vazia ou repetida.
func indexarOpcoes(opcoes []string) (map[string]int, error) {
	if len(opcoes) == 0 {
		return nil, ErrOpcoesVazias
	}

	indice := make(map[string]int, len(opcoes))
	for posicao, opcao := range opcoes {
		if _, repetida := indice[opcao]; repetida {
			return nil, ErrOpcoesDuplicadas
		}
		indice[opcao] = posicao
	}
	return indice, nil
}

func validarPreferencias(preferencias []string, indice map[string]int) error {
	if len(preferencias) != len(indice) {
		return ErrVotoInvalido
	}

	vistas := make(map[string]bool, len(preferencias))
	for _, opcao := range preferencias {
		if _, valida := indice[opcao]; !valida || vistas[opcao] {
			return ErrVotoInvalido
		}
		vistas[opcao] = true
	}
	return nil
}

// pontuarOpcoes soma os pontos de cada opção mantendo a ordem de cadastro.
func pontuarOpcoes(opcoes []string, indice map[string]int, votos []VotoBorda) []OpcaoPontuada {
	ranking := make([]OpcaoPontuada, len(opcoes))
	for posicao, opcao := range opcoes {
		ranking[posicao].Opcao = opcao
	}

	for _, voto := range votos {
		for posicao, opcao := range voto.Preferencias {
			linha := &ranking[indice[opcao]]
			linha.Pontos += len(opcoes) - posicao
			if posicao == 0 {
				linha.PrimeirasPreferencias++
			}
		}
	}
	return ranking
}

// ordenarRanking coloca a opção a antes de b se ela tem mais pontos ou, com os mesmos pontos,
// mais primeiras preferências; a ordenação estável preserva a ordem de cadastro nos empates.
func ordenarRanking(ranking []OpcaoPontuada) {
	sort.SliceStable(ranking, func(i, j int) bool {
		a, b := ranking[i], ranking[j]
		return a.Pontos > b.Pontos ||
			(a.Pontos == b.Pontos && a.PrimeirasPreferencias > b.PrimeirasPreferencias)
	})
}

// marcarEmpates marca as opções vizinhas no ranking ordenado que nenhum critério separa.
func marcarEmpates(ranking []OpcaoPontuada) {
	for i := 1; i < len(ranking); i++ {
		if empatadas(ranking[i-1], ranking[i]) {
			ranking[i-1].Empatado = true
			ranking[i].Empatado = true
		}
	}
}

func empatadas(a, b OpcaoPontuada) bool {
	return a.Pontos == b.Pontos && a.PrimeirasPreferencias == b.PrimeirasPreferencias
}
