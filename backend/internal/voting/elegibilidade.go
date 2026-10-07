package voting

type Condomino struct {
	ID          string
	Adimplente  bool
	FracaoIdeal float64
}

// ValidarElegibilidade checa se o condômino tem direito a voto e se sua fração é positiva.
func ValidarElegibilidade(c Condomino) (bool, string) {
	if c.ID == "" {
		return false, "identificador do condomino invalido"
	}
	if !c.Adimplente {
		return false, "condomino impedido de votar por inadimplencia"
	}
	if c.FracaoIdeal <= 0 {
		return false, "fracao ideal deve ser maior que zero"
	}

	return true, "apto a votar"
}
