package voting

// verifica se os votos favoráveis atingiram 2/3 da fração total do condomínio
func AtingiuQuorum(votosFavoraveis float64, totalFracoes float64) bool {
	if totalFracoes <= 0 {
		return false
	}
	// 2.0 / 3.0 representa o quorum qualificado de dois terços
	return (votosFavoraveis / totalFracoes) >= (2.0 / 3.0)
}
