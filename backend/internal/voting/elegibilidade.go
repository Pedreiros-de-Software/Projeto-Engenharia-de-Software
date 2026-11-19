package voting

import (
	"errors"
	"math"
	"strings"
)

// Erros sentinela exportados para validação de elegibilidade
var (
	ErrIdentificadorInvalido = errors.New("identificador do condomino invalido")
	ErrInadimplente          = errors.New("condomino impedido de votar por inadimplencia")
	ErrFracaoIdealInvalida   = errors.New("fracao ideal deve ser maior que zero")
)

type Condomino struct {
	ID          string
	Adimplente  bool
	FracaoIdeal float64
}

// ValidarElegibilidade checa se o condômino tem direito a voto e se sua fração é positiva.
// Retorna nil quando o condômino está apto a votar ou um erro sentinela específico caso contrário.
func ValidarElegibilidade(c Condomino) error {
	if strings.TrimSpace(c.ID) == "" {
		return ErrIdentificadorInvalido
	}

	if !c.Adimplente {
		return ErrInadimplente
	}

	if c.FracaoIdeal <= 0 || math.IsNaN(c.FracaoIdeal) || math.IsInf(c.FracaoIdeal, 0) {
		return ErrFracaoIdealInvalida
	}

	return nil
}
