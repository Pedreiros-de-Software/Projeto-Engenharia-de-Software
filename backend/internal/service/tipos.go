package service

import (
	"time"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store"
)

// Usuario é a conta exposta pela aplicação (sem o hash da senha).
type Usuario struct {
	ID    int64
	Nome  string
	Login string
	Papel string
}

type Responsavel struct {
	ID   int64
	Nome string
}

type Unidade struct {
	ID            int64
	Identificador string
	FracaoIdeal   float64
	Percentual    float64 // fração sobre o peso total das unidades com responsável (0..100)
	Adimplente    bool
	Responsavel   *Responsavel // nil quando a unidade não tem responsável
}

type VotacaoResumo struct {
	ID                int64
	Tipo              string
	Titulo            string
	Descricao         string
	Inicio            time.Time
	Fim               time.Time
	Turno             int
	Status            string
	Apto              bool // o usuário pode votar nesta votação
	JaVotou           bool // o usuário (ou sua unidade, na pauta) já votou
	VotacaoAnteriorID *int64
	SegundoTurnoID    *int64
}

type Opcao struct {
	ID            int64
	Nome          string
	Descricao     string
	CustoCentavos *int64
}

// MinhaUnidade é a unidade do usuário vista de dentro de uma pauta: peso e se pode votar.
type MinhaUnidade struct {
	Identificador string
	FracaoIdeal   float64
	Percentual    float64
	Adimplente    bool
	Elegivel      bool
	Motivo        string // vazio quando elegível
}

type VotacaoDetalhe struct {
	VotacaoResumo
	Opcoes            []Opcao
	OrcamentoCentavos *int64
	MinhaUnidade      *MinhaUnidade // só na pauta, quando o usuário é responsável por uma unidade
}

func paraUsuario(u store.Usuario) Usuario {
	return Usuario{ID: u.ID, Nome: u.Nome, Login: u.Login, Papel: u.Papel}
}

func paraOpcao(o store.Opcao) Opcao {
	return Opcao{ID: o.ID, Nome: o.Nome, Descricao: o.Descricao, CustoCentavos: o.CustoCentavos}
}
