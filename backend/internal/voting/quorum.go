package voting

import "math"

// AtingiuQuorum verifica se os votos favoráveis atingiram 2/3 da fração total.
// Totais não positivos, valores não finitos ou votos fora do total são inválidos.
func AtingiuQuorum(votosFavoraveis float64, totalFracoes float64) bool {
	if math.IsNaN(votosFavoraveis) || math.IsNaN(totalFracoes) ||
		math.IsInf(votosFavoraveis, 0) || math.IsInf(totalFracoes, 0) ||
		totalFracoes <= 0 || votosFavoraveis < 0 || votosFavoraveis > totalFracoes {
		return false
	}
	return (votosFavoraveis / totalFracoes) >= (2.0 / 3.0)
}
