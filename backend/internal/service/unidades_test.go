package service

import (
	"math"
	"testing"
)

func TestCriarUnidade(t *testing.T) {
	c := novoCenario(t)
	mariaID := c.usuario(t, "maria", PapelEleitor)
	joaoID := c.usuario(t, "joao", PapelEleitor)
	c.unidade(t, "Apto 101", 60, true, mariaID)

	testes := []struct {
		nome          string
		identificador string
		fracao        float64
		responsavel   *int64
		esperado      error
	}{
		{"unidade válida", "Apto 102", 60, &joaoID, nil},
		{"unidade sem responsável", "Loja 1", 10, nil, nil},
		{"identificador vazio", " ", 10, nil, errValidacao},
		{"fração zero", "Apto 201", 0, nil, errValidacao},
		{"fração negativa", "Apto 202", -1, nil, errValidacao},
		{"fração NaN", "Apto 203", math.NaN(), nil, errValidacao},
		{"fração infinita", "Apto 204", math.Inf(1), nil, errValidacao},
		{"identificador repetido", "Apto 101", 10, nil, errValidacao},
		{"responsável que já tem unidade", "Apto 301", 10, &mariaID, errValidacao},
		{"identificador com caractere NUL", "Apto\x00401", 10, nil, errValidacao},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			_, err := c.svc.CriarUnidade(t.Context(), tc.identificador, tc.fracao, true, tc.responsavel)
			conferirErro(t, err, tc.esperado)
		})
	}
}

func TestUnidades(t *testing.T) {
	c := novoCenario(t)
	mariaID := c.usuario(t, "maria", PapelEleitor)
	sofiaID := c.usuario(t, "sofia", PapelEleitor)
	anaID := c.usuario(t, "ana", PapelEleitor)
	apto101 := c.unidade(t, "Apto 101", 60, true, mariaID)
	apto402 := c.unidade(t, "Apto 402", 140, false, sofiaID)
	loja, err := c.svc.CriarUnidade(t.Context(), "Loja 1", 50, true, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("listar calcula o percentual sobre as unidades com responsável", func(t *testing.T) {
		unidades, err := c.svc.ListarUnidades(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		esperado := []Unidade{
			{ID: apto101, Identificador: "Apto 101", FracaoIdeal: 60, Percentual: 30, Adimplente: true, Responsavel: &Responsavel{ID: mariaID, Nome: "Nome maria"}},
			{ID: apto402, Identificador: "Apto 402", FracaoIdeal: 140, Percentual: 70, Responsavel: &Responsavel{ID: sofiaID, Nome: "Nome sofia"}},
			{ID: loja, Identificador: "Loja 1", FracaoIdeal: 50, Percentual: 0, Adimplente: true},
		}
		if len(unidades) != len(esperado) {
			t.Fatalf("esperava %d unidades, obteve %+v", len(esperado), unidades)
		}
		for i := range esperado {
			if !mesmaUnidade(unidades[i], esperado[i]) {
				t.Errorf("unidade %d: esperado %+v, obtido %+v", i, esperado[i], unidades[i])
			}
		}
	})

	t.Run("marcar adimplente libera o voto", func(t *testing.T) {
		u, err := c.svc.AlterarAdimplencia(t.Context(), apto402, true)
		if err != nil || !u.Adimplente || u.ID != apto402 || u.Percentual != 70 {
			t.Errorf("unidade inesperada: %+v (err %v)", u, err)
		}
		pauta := c.votacao(t, NovaVotacao{Tipo: TipoPauta})
		if err := c.svc.Votar(t.Context(), sofiaID, pauta, Voto{Escolha: "favor"}); err != nil {
			t.Errorf("sofia deveria poder votar depois de regularizada: %v", err)
		}
	})

	t.Run("unidade inexistente", func(t *testing.T) {
		_, err := c.svc.AlterarAdimplencia(t.Context(), 9999, true)
		conferirErro(t, err, ErrNaoEncontrado)
	})

	t.Run("unidade do usuário", func(t *testing.T) {
		u, err := c.svc.UnidadeDoUsuario(t.Context(), mariaID)
		if err != nil || u == nil || u.ID != apto101 || u.Responsavel == nil || u.Responsavel.ID != mariaID {
			t.Errorf("esperava a unidade %d, obteve %+v (err %v)", apto101, u, err)
		}
		u, err = c.svc.UnidadeDoUsuario(t.Context(), anaID)
		if err != nil || u != nil {
			t.Errorf("usuário sem unidade: esperava nil, obteve %+v (err %v)", u, err)
		}
	})
}

func mesmaUnidade(a, b Unidade) bool {
	mesmoResponsavel := (a.Responsavel == nil && b.Responsavel == nil) ||
		(a.Responsavel != nil && b.Responsavel != nil && *a.Responsavel == *b.Responsavel)
	a.Responsavel, b.Responsavel = nil, nil
	return mesmoResponsavel && a == b
}
