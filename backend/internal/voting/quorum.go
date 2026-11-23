package voting

import "math"

// AtingiuQuorum verifica se os votos favoráveis atingiram 2/3 da fração total.
// Totais não positivos, valores não finitos ou votos fora do total são inválidos.
func AtingiuQuorum(votosFavoraveis float64, totalFracoes float64) bool {
	if !finito(votosFavoraveis) || !finito(totalFracoes) ||
		totalFracoes <= 0 || votosFavoraveis < 0 || votosFavoraveis > totalFracoes {
		return false
	}
	return (votosFavoraveis / totalFracoes) >= (2.0 / 3.0)
}

// finito indica se x não é NaN nem infinito.
func finito(x float64) bool {
	return !math.IsNaN(x) && !math.IsInf(x, 0)
}
