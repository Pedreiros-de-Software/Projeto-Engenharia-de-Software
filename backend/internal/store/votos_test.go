package store_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store"
)

func TestVotosPauta(t *testing.T) {
	_, repo := novoBanco(t)
	ctx := t.Context()

	pautaID := inserirVotacao(t, repo, store.Votacao{Tipo: "pauta", Titulo: "Pauta"})
	outraPauta := inserirVotacao(t, repo, store.Votacao{Tipo: "pauta", Titulo: "Outra"})
	unidade1, _ := repo.Unidades.Inserir(ctx, store.Unidade{Identificador: "101", FracaoIdeal: 1, Adimplente: true})
	unidade2, _ := repo.Unidades.Inserir(ctx, store.Unidade{Identificador: "102", FracaoIdeal: 1, Adimplente: true})

	for _, voto := range []struct {
		votacao, unidade int64
		escolha          string
	}{{pautaID, unidade2, "contra"}, {pautaID, unidade1, "favor"}, {outraPauta, unidade1, "abstencao"}} {
		if err := repo.Votos.InserirVotoPauta(ctx, voto.votacao, voto.unidade, voto.escolha); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("segundo voto da unidade é ErrDuplicado", func(t *testing.T) {
		if err := repo.Votos.InserirVotoPauta(ctx, pautaID, unidade1, "contra"); !errors.Is(err, store.ErrDuplicado) {
			t.Errorf("esperava ErrDuplicado, obteve %v", err)
		}
	})

	t.Run("listar votos ordenados por unidade", func(t *testing.T) {
		votos, err := repo.Votos.ListarVotosPauta(ctx, pautaID)
		esperado := []store.VotoPauta{{UnidadeID: unidade1, Escolha: "favor"}, {UnidadeID: unidade2, Escolha: "contra"}}
		if err != nil || !slices.Equal(votos, esperado) {
			t.Errorf("esperado %+v, obtido %+v (err %v)", esperado, votos, err)
		}
	})

	t.Run("pautas votadas pela unidade", func(t *testing.T) {
		ids, err := repo.Votos.IDsVotadasPelaUnidade(ctx, unidade1)
		slices.Sort(ids)
		if err != nil || !slices.Equal(ids, []int64{pautaID, outraPauta}) {
			t.Errorf("esperado %v, obtido %v (err %v)", []int64{pautaID, outraPauta}, ids, err)
		}
	})
}

func TestParticipacoesECedulas(t *testing.T) {
	_, repo := novoBanco(t)
	ctx := t.Context()

	eleicaoID := inserirVotacao(t, repo, store.Votacao{Tipo: "eleicao", Titulo: "Eleição"})
	consultaID := inserirVotacao(t, repo, store.Votacao{Tipo: "consulta", Titulo: "Consulta", OrcamentoCentavos: ptr[int64](0)})
	chapaID, _ := repo.Votacoes.InserirOpcao(ctx, store.Opcao{VotacaoID: eleicaoID, Nome: "Chapa", Posicao: 1})
	projetoA, _ := repo.Votacoes.InserirOpcao(ctx, store.Opcao{VotacaoID: consultaID, Nome: "A", Posicao: 1, CustoCentavos: ptr[int64](1)})
	projetoB, _ := repo.Votacoes.InserirOpcao(ctx, store.Opcao{VotacaoID: consultaID, Nome: "B", Posicao: 2, CustoCentavos: ptr[int64](1)})
	anaID := inserirUsuario(t, repo, "ana", "eleitor")

	t.Run("participação repetida é ErrDuplicado", func(t *testing.T) {
		if err := repo.Votos.InserirParticipacao(ctx, eleicaoID, anaID); err != nil {
			t.Fatal(err)
		}
		if err := repo.Votos.InserirParticipacao(ctx, eleicaoID, anaID); !errors.Is(err, store.ErrDuplicado) {
			t.Errorf("esperava ErrDuplicado, obteve %v", err)
		}
		if total, err := repo.Votos.ContarParticipacoes(ctx, eleicaoID); err != nil || total != 1 {
			t.Errorf("participações: esperado 1, obtido %d (err %v)", total, err)
		}
		if ids, err := repo.Votos.IDsParticipadas(ctx, anaID); err != nil || !slices.Equal(ids, []int64{eleicaoID}) {
			t.Errorf("votações participadas: esperado [%d], obtido %v (err %v)", eleicaoID, ids, err)
		}
	})

	t.Run("cédulas de eleição", func(t *testing.T) {
		for _, c := range []store.CedulaEleicao{{Tipo: "chapa", OpcaoID: &chapaID}, {Tipo: "branco"}, {Tipo: "nulo"}} {
			id, err := repo.Votos.InserirCedula(ctx, eleicaoID, c.Tipo, c.OpcaoID)
			if err != nil || len(id) != 36 {
				t.Fatalf("inserindo cédula %s: id %q, err %v", c.Tipo, id, err)
			}
		}

		cedulas, err := repo.Votos.ListarCedulasEleicao(ctx, eleicaoID)
		if err != nil || len(cedulas) != 3 {
			t.Fatalf("esperava 3 cédulas, obteve %+v (err %v)", cedulas, err)
		}
		porTipo := map[string]int{}
		for _, c := range cedulas {
			porTipo[c.Tipo]++
			if (c.Tipo == "chapa") != (c.OpcaoID != nil) || (c.OpcaoID != nil && *c.OpcaoID != chapaID) {
				t.Errorf("cédula inconsistente: %+v", c)
			}
		}
		if porTipo["chapa"] != 1 || porTipo["branco"] != 1 || porTipo["nulo"] != 1 {
			t.Errorf("tipos inesperados: %v", porTipo)
		}
	})

	t.Run("cédula de chapa sem opção viola a restrição do banco", func(t *testing.T) {
		if _, err := repo.Votos.InserirCedula(ctx, eleicaoID, "chapa", nil); err == nil {
			t.Error("esperava erro de CHECK")
		}
	})

	t.Run("cédulas de consulta agrupam as preferências em ordem", func(t *testing.T) {
		ordens := [][]int64{{projetoB, projetoA}, {projetoA, projetoB}, {projetoA, projetoB}}
		for _, ordem := range ordens {
			cedulaID, err := repo.Votos.InserirCedula(ctx, consultaID, "ranking", nil)
			if err != nil {
				t.Fatal(err)
			}
			for posicao, opcaoID := range ordem {
				if err := repo.Votos.InserirPreferencia(ctx, cedulaID, posicao+1, opcaoID); err != nil {
					t.Fatal(err)
				}
			}
		}

		cedulas, err := repo.Votos.ListarCedulasRanking(ctx, consultaID)
		if err != nil {
			t.Fatal(err)
		}
		contagem := map[int64]int{}
		for _, c := range cedulas {
			if len(c) != 2 {
				t.Fatalf("cédula incompleta: %v", c)
			}
			contagem[c[0]]++
		}
		if len(cedulas) != 3 || contagem[projetoA] != 2 || contagem[projetoB] != 1 {
			t.Errorf("cédulas inesperadas: %v", cedulas)
		}
	})

	t.Run("consulta sem cédulas devolve lista vazia", func(t *testing.T) {
		cedulas, err := repo.Votos.ListarCedulasRanking(ctx, eleicaoID)
		if err != nil || cedulas == nil || len(cedulas) != 0 {
			t.Errorf("esperava lista vazia, obteve %v (err %v)", cedulas, err)
		}
	})
}
