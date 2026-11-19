package store_test

import (
	"errors"
	"testing"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store"
)

func TestUnidades(t *testing.T) {
	_, repo := novoBanco(t)
	ctx := t.Context()

	mariaID := inserirUsuario(t, repo, "maria", "eleitor")
	aptoComDono, err := repo.Unidades.Inserir(ctx, store.Unidade{Identificador: "Apto 101", FracaoIdeal: 60, Adimplente: true, ResponsavelID: &mariaID})
	if err != nil {
		t.Fatal(err)
	}
	aptoVago, err := repo.Unidades.Inserir(ctx, store.Unidade{Identificador: "Apto 102", FracaoIdeal: 40, Adimplente: false})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("identificador repetido é ErrDuplicado", func(t *testing.T) {
		_, err := repo.Unidades.Inserir(ctx, store.Unidade{Identificador: "Apto 101", FracaoIdeal: 10, Adimplente: true})
		if !errors.Is(err, store.ErrDuplicado) {
			t.Errorf("esperava ErrDuplicado, obteve %v", err)
		}
	})

	t.Run("listar traz o nome do responsável e nil para unidade sem responsável", func(t *testing.T) {
		unidades, err := repo.Unidades.Listar(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(unidades) != 2 {
			t.Fatalf("esperava 2 unidades, obteve %+v", unidades)
		}
		comDono, vaga := unidades[0], unidades[1]
		if comDono.ID != aptoComDono || comDono.ResponsavelID == nil || *comDono.ResponsavelID != mariaID ||
			comDono.ResponsavelNome != "Nome maria" || comDono.FracaoIdeal != 60 || !comDono.Adimplente {
			t.Errorf("unidade com responsável inesperada: %+v", comDono)
		}
		if vaga.ID != aptoVago || vaga.ResponsavelID != nil || vaga.ResponsavelNome != "" || vaga.Adimplente {
			t.Errorf("unidade sem responsável inesperada: %+v", vaga)
		}
	})

	t.Run("buscar por responsável", func(t *testing.T) {
		u, err := repo.Unidades.BuscarPorResponsavel(ctx, mariaID)
		if err != nil || u.ID != aptoComDono {
			t.Errorf("esperava a unidade %d, obteve %+v (err %v)", aptoComDono, u, err)
		}
		if _, err := repo.Unidades.BuscarPorResponsavel(ctx, mariaID+100); !errors.Is(err, store.ErrNaoEncontrado) {
			t.Errorf("esperava ErrNaoEncontrado, obteve %v", err)
		}
	})

	t.Run("peso total soma apenas unidades com responsável", func(t *testing.T) {
		if total, err := repo.Unidades.PesoTotal(ctx); err != nil || total != 60 {
			t.Errorf("esperava 60, obteve %v (err %v)", total, err)
		}
	})

	t.Run("alterar adimplência", func(t *testing.T) {
		if err := repo.Unidades.AlterarAdimplencia(ctx, aptoVago, true); err != nil {
			t.Fatal(err)
		}
		u, err := repo.Unidades.Buscar(ctx, aptoVago)
		if err != nil || !u.Adimplente {
			t.Errorf("esperava unidade adimplente, obteve %+v (err %v)", u, err)
		}
	})

	t.Run("unidade inexistente é ErrNaoEncontrado", func(t *testing.T) {
		if err := repo.Unidades.AlterarAdimplencia(ctx, 9999, true); !errors.Is(err, store.ErrNaoEncontrado) {
			t.Errorf("alterar: esperava ErrNaoEncontrado, obteve %v", err)
		}
		if _, err := repo.Unidades.Buscar(ctx, 9999); !errors.Is(err, store.ErrNaoEncontrado) {
			t.Errorf("buscar: esperava ErrNaoEncontrado, obteve %v", err)
		}
	})
}

func TestPesoTotalSemUnidades(t *testing.T) {
	_, repo := novoBanco(t)
	if total, err := repo.Unidades.PesoTotal(t.Context()); err != nil || total != 0 {
		t.Errorf("esperava 0, obteve %v (err %v)", total, err)
	}
}
