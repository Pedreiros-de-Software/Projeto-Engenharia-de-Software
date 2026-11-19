package voting

import (
	"errors"
	"sort"
)

// OpcaoPontuada representa o resultado de cada projeto/opção na apuração.
type OpcaoPontuada struct {
	Opcao                 string
	Pontos                int
	PrimeirasPreferencias int
}

// VotoBorda representa a cédula de um eleitor com sua lista ordenada de preferências.
type VotoBorda struct {
	EleitorID    string
	Preferencias []string
}

var (
	ErrOpcoesVazias = errors.New("a lista de opções da votação não pode ser vazia")
	ErrVotoInvalido = errors.New("o voto contém opções inválidas, duplicadas ou lista incompleta")
)

// CalcularApuracaoBorda executa a contagem pelo Método de Borda.
// Rejeita votos inválidos ou incompletos e desempata por primeiras preferências.
func CalcularApuracaoBorda(opcoesValidas []string, votos []VotoBorda) ([]OpcaoPontuada, error) {
	if len(opcoesValidas) == 0 {
		return nil, ErrOpcoesVazias
	}

	numOpcoes := len(opcoesValidas)
	mapaOpcoesValidas := make(map[string]bool, numOpcoes)
	pontuacao := make(map[string]int, numOpcoes)
	primeirasPref := make(map[string]int, numOpcoes)

	for _, op := range opcoesValidas {
		mapaOpcoesValidas[op] = true
		pontuacao[op] = 0
		primeirasPref[op] = 0
	}

	for _, voto := range votos {
		if err := validarVoto(voto.Preferencias, mapaOpcoesValidas, numOpcoes); err != nil {
			return nil, err // Retorna erro em vez de ignorar silenciosamente
		}

		for pos, opcao := range voto.Preferencias {
			pontos := numOpcoes - pos
			pontuacao[opcao] += pontos

			if pos == 0 {
				primeirasPref[opcao]++ // Contagem das 1ªs preferências (Issue #24)
			}
		}
	}

	resultado := make([]OpcaoPontuada, 0, numOpcoes)
	for op, pts := range pontuacao {
		resultado = append(resultado, OpcaoPontuada{
			Opcao:                 op,
			Pontos:                pts,
			PrimeirasPreferencias: primeirasPref[op],
		})
	}

	// Ordenação: 1º Pontuação total (decrescente); 2º Primeiras preferências (decrescente)
	sort.SliceStable(resultado, func(i, j int) bool {
		if resultado[i].Pontos != resultado[j].Pontos {
			return resultado[i].Pontos > resultado[j].Pontos
		}
		return resultado[i].PrimeirasPreferencias > resultado[j].PrimeirasPreferencias
	})

	return resultado, nil
}

func validarVoto(preferencias []string, mapaOpcoesValidas map[string]bool, numOpcoes int) error {
	// Rejeita listas incompletas ou com tamanho incorreto (Issue #23)
	if len(preferencias) != numOpcoes {
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
