package borda

import (
	"errors"
	"sort"
)

// OpcaoPontuada representa o resultado de cada projeto/opção na apuração.
type OpcaoPontuada struct {
	Opcao  string `json:"opcao"`
	Pontos int    `json:"pontos"`
}

// VotoBorda representa a cédula de um eleitor com sua lista ordenada de preferências.
type VotoBorda struct {
	EleitorID    string   `json:"eleitor_id"`
	Preferencias []string `json:"preferencias"`
}

var (
	ErrOpcoesVazias    = errors.New("a lista de opções da votação não pode ser vazia")
	ErrSemVotosValidos = errors.New("nenhum voto válido foi computado na apuração")
	ErrVotoInvalido    = errors.New("o voto contém opções inválidas ou duplicadas")
)

// CalcularApuracaoBorda executa a contagem pelo Método de Borda.
// Para N opções válidas, a 1ª opção da cédula recebe N pontos, a 2ª recebe N-1, até a N-ésima que recebe 1 ponto.
func CalcularApuracaoBorda(opcoesValidas []string, votos []VotoBorda) ([]OpcaoPontuada, error) {
	if len(opcoesValidas) == 0 {
		return nil, ErrOpcoesVazias
	}

	numOpcoes := len(opcoesValidas)
	mapaOpcoesValidas := make(map[string]bool, numOpcoes)
	pontuacao := make(map[string]int, numOpcoes)

	for _, op := range opcoesValidas {
		mapaOpcoesValidas[op] = true
		pontuacao[op] = 0
	}

	votosComputados := 0

	for _, voto := range votos {
		if err := validarVoto(voto.Preferencias, mapaOpcoesValidas, numOpcoes); err != nil {
			// Voto inválido é ignorado na contagem oficial
			continue
		}

		votosComputados++
		for pos, opcao := range voto.Preferencias {
			pontos := numOpcoes - pos
			pontuacao[opcao] += pontos
		}
	}

	if votosComputados == 0 && len(votos) > 0 {
		return nil, ErrSemVotosValidos
	}

	// Converte o mapa para slice para possibilitar ordenação
	resultado := make([]OpcaoPontuada, 0, numOpcoes)
	for op, pts := range pontuacao {
		resultado = append(resultado, OpcaoPontuada{
			Opcao:  op,
			Pontos: pts,
		})
	}

	// Ordena decrescente por pontuação acumulada (em caso de empate, por ordem alfabética)
	sort.Slice(resultado, func(i, j int) bool {
		if resultado[i].Pontos == resultado[j].Pontos {
			return resultado[i].Opcao < resultado[j].Opcao
		}
		return resultado[i].Pontos > resultado[j].Pontos
	})

	return resultado, nil
}

func validarVoto(preferencias []string, mapaOpcoesValidas map[string]bool, numOpcoes int) error {
	if len(preferencias) == 0 || len(preferencias) > numOpcoes {
		return ErrVotoInvalido
	}

	visitados := make(map[string]bool, len(preferencias))
	for _, op := range preferencias {
		if !mapaOpcoesValidas[op] || visitados[op] {
			return ErrVotoInvalido
		}
		visitados[op] = true
	}

	return nil
}
