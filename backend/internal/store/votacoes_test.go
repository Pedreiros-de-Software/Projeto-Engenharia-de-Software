package store_test

import (
	"database/sql"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store"
)

func ptr[T any](v T) *T { return &v }

func inserirVotacao(t *testing.T, repo store.Repositorios, v store.Votacao) int64 {
	t.Helper()
	if v.Inicio.IsZero() {
		v.Inicio, v.Fim = instante, instante.Add(time.Hour)
	}
	if v.Turno == 0 {
		v.Turno = 1
	}
	id, err := repo.Votacoes.Inserir(t.Context(), v)
	if err != nil {
		t.Fatalf("inserindo votação: %v", err)
	}
	return id
}

func TestVotacoes(t *testing.T) {
	_, repo := novoBanco(t)
	ctx := t.Context()

	pautaID := inserirVotacao(t, repo, store.Votacao{Tipo: "pauta", Titulo: "Pauta", Descricao: "Desc"})
	consultaID := inserirVotacao(t, repo, store.Votacao{Tipo: "consulta", Titulo: "Consulta", OrcamentoCentavos: ptr[int64](500)})

	t.Run("buscar devolve todos os campos", func(t *testing.T) {
		v, err := repo.Votacoes.Buscar(ctx, consultaID)
		if err != nil {
			t.Fatal(err)
		}
		if v.Tipo != "consulta" || v.Titulo != "Consulta" || v.Turno != 1 || !v.Inicio.Equal(instante) ||
			!v.Fim.Equal(instante.Add(time.Hour)) || v.OrcamentoCentavos == nil || *v.OrcamentoCentavos != 500 ||
			v.VotacaoAnteriorID != nil || v.SegundoTurnoID != nil {
			t.Errorf("votação inesperada: %+v", v)
		}
	})

	t.Run("votação inexistente é ErrNaoEncontrado", func(t *testing.T) {
		if _, err := repo.Votacoes.Buscar(ctx, 9999); !errors.Is(err, store.ErrNaoEncontrado) {
			t.Errorf("buscar: esperava ErrNaoEncontrado, obteve %v", err)
		}
		if _, err := repo.Votacoes.BuscarComTrava(ctx, 9999); !errors.Is(err, store.ErrNaoEncontrado) {
			t.Errorf("buscar com trava: esperava ErrNaoEncontrado, obteve %v", err)
		}
		if encerrou, err := repo.Votacoes.EncerrarSeAberta(ctx, 9999, instante); err != nil || encerrou {
			t.Errorf("encerrar: esperava false sem erro, obteve %v (err %v)", encerrou, err)
		}
	})

	t.Run("listar começa pela mais recente", func(t *testing.T) {
		votacoes, err := repo.Votacoes.Listar(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(votacoes) != 2 || votacoes[0].ID != consultaID || votacoes[1].ID != pautaID {
			t.Errorf("ordem inesperada: %+v", votacoes)
		}
	})

	t.Run("contar", func(t *testing.T) {
		if total, err := repo.Votacoes.Contar(ctx); err != nil || total != 2 {
			t.Errorf("total: esperado 2, obtido %d (err %v)", total, err)
		}
	})

	t.Run("segundo turno fica ligado à votação anterior e é único", func(t *testing.T) {
		eleicaoID := inserirVotacao(t, repo, store.Votacao{Tipo: "eleicao", Titulo: "Eleição"})
		turno2 := inserirVotacao(t, repo, store.Votacao{Tipo: "eleicao", Titulo: "Eleição — 2º turno", Turno: 2, VotacaoAnteriorID: &eleicaoID})

		v, err := repo.Votacoes.Buscar(ctx, eleicaoID)
		if err != nil || v.SegundoTurnoID == nil || *v.SegundoTurnoID != turno2 {
			t.Errorf("segundo_turno_id: esperado %d, obtido %+v (err %v)", turno2, v, err)
		}

		_, err = repo.Votacoes.Inserir(ctx, store.Votacao{Tipo: "eleicao", Titulo: "Outro", Inicio: instante,
			Fim: instante.Add(time.Hour), Turno: 2, VotacaoAnteriorID: &eleicaoID})
		if !errors.Is(err, store.ErrDuplicado) {
			t.Errorf("esperava ErrDuplicado, obteve %v", err)
		}
	})
}

func TestOpcoesEAptos(t *testing.T) {
	_, repo := novoBanco(t)
	ctx := t.Context()

	origem := inserirVotacao(t, repo, store.Votacao{Tipo: "consulta", Titulo: "Consulta", OrcamentoCentavos: ptr[int64](0)})
	destino := inserirVotacao(t, repo, store.Votacao{Tipo: "consulta", Titulo: "Outra", OrcamentoCentavos: ptr[int64](0)})
	anaID := inserirUsuario(t, repo, "ana", "eleitor")
	brunoID := inserirUsuario(t, repo, "bruno", "eleitor")

	segunda, err := repo.Votacoes.InserirOpcao(ctx, store.Opcao{VotacaoID: origem, Nome: "B", Posicao: 2, CustoCentavos: ptr[int64](20)})
	if err != nil {
		t.Fatal(err)
	}
	primeira, err := repo.Votacoes.InserirOpcao(ctx, store.Opcao{VotacaoID: origem, Nome: "A", Descricao: "da", Posicao: 1})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("opção com nome repetido é ErrDuplicado", func(t *testing.T) {
		_, err := repo.Votacoes.InserirOpcao(ctx, store.Opcao{VotacaoID: origem, Nome: "A", Posicao: 3})
		if !errors.Is(err, store.ErrDuplicado) {
			t.Errorf("esperava ErrDuplicado, obteve %v", err)
		}
	})

	t.Run("opções saem na ordem de posição", func(t *testing.T) {
		opcoes, err := repo.Votacoes.ListarOpcoes(ctx, origem)
		if err != nil {
			t.Fatal(err)
		}
		if len(opcoes) != 2 || opcoes[0].ID != primeira || opcoes[1].ID != segunda {
			t.Fatalf("ordem inesperada: %+v", opcoes)
		}
		if opcoes[0].Descricao != "da" || opcoes[0].CustoCentavos != nil || *opcoes[1].CustoCentavos != 20 {
			t.Errorf("campos inesperados: %+v", opcoes)
		}
	})

	if err := repo.Votacoes.InserirApto(ctx, origem, anaID); err != nil {
		t.Fatal(err)
	}
	if err := repo.Votacoes.InserirApto(ctx, origem, brunoID); err != nil {
		t.Fatal(err)
	}

	t.Run("apto repetido é ErrDuplicado", func(t *testing.T) {
		if err := repo.Votacoes.InserirApto(ctx, origem, anaID); !errors.Is(err, store.ErrDuplicado) {
			t.Errorf("esperava ErrDuplicado, obteve %v", err)
		}
	})

	t.Run("copiar aptos", func(t *testing.T) {
		if err := repo.Votacoes.CopiarAptos(ctx, origem, destino); err != nil {
			t.Fatal(err)
		}
		if total, err := repo.Votacoes.ContarAptos(ctx, destino); err != nil || total != 2 {
			t.Errorf("aptos copiados: esperado 2, obtido %d (err %v)", total, err)
		}
	})

	t.Run("está apto e votações aptas do usuário", func(t *testing.T) {
		if apto, err := repo.Votacoes.EstaApto(ctx, origem, anaID); err != nil || !apto {
			t.Errorf("ana deveria estar apta (err %v)", err)
		}
		if apto, err := repo.Votacoes.EstaApto(ctx, origem, anaID+100); err != nil || apto {
			t.Errorf("usuário desconhecido não deveria estar apto (err %v)", err)
		}
		ids, err := repo.Votacoes.IDsAptas(ctx, anaID)
		slices.Sort(ids)
		if err != nil || !slices.Equal(ids, []int64{origem, destino}) {
			t.Errorf("votações aptas: esperado %v, obtido %v (err %v)", []int64{origem, destino}, ids, err)
		}
	})
}

func TestEncerrarSeAberta(t *testing.T) {
	_, repo := novoBanco(t)
	fimOriginal := instante.Add(time.Hour)

	testes := []struct {
		nome        string
		agora       time.Time
		encerrou    bool
		fimEsperado time.Time
	}{
		{"aberta: o fim passa a ser agora", instante.Add(30 * time.Minute), true, instante.Add(30 * time.Minute)},
		{"antes do início", instante.Add(-time.Minute), false, fimOriginal},
		{"no instante do início a janela ficaria vazia", instante, false, fimOriginal},
		{"no instante do fim já está encerrada", fimOriginal, false, fimOriginal},
		{"depois do fim", fimOriginal.Add(time.Minute), false, fimOriginal},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			id := inserirVotacao(t, repo, store.Votacao{Tipo: "pauta", Titulo: "Pauta"})
			encerrou, err := repo.Votacoes.EncerrarSeAberta(t.Context(), id, tc.agora)
			if err != nil || encerrou != tc.encerrou {
				t.Fatalf("encerrou: esperado %v, obtido %v (err %v)", tc.encerrou, encerrou, err)
			}
			if v, _ := repo.Votacoes.Buscar(t.Context(), id); !v.Fim.Equal(tc.fimEsperado) {
				t.Errorf("fim: esperado %v, obtido %v", tc.fimEsperado, v.Fim)
			}
		})
	}
}

// esperaPouco é o tempo dado a uma operação que deveria estar bloqueada para, por engano, terminar.
const esperaPouco = 200 * time.Millisecond

// segurarTransacao abre uma transação, roda passo nela e a mantém aberta até liberar ser fechado.
func segurarTransacao(t *testing.T, db *sql.DB, passo func(store.Repositorios) error) (liberar chan struct{}, fim chan error) {
	t.Helper()
	pronto, liberar, fim := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	go func() {
		fim <- store.EmTransacao(t.Context(), db, func(r store.Repositorios) error {
			err := passo(r)
			close(pronto)
			<-liberar
			return err
		})
	}()
	<-pronto
	return liberar, fim
}

func TestTravaDoVotoContraEncerramento(t *testing.T) {
	t.Run("encerrar espera a transação de voto em andamento", func(t *testing.T) {
		db, repo := novoBanco(t)
		id := inserirVotacao(t, repo, store.Votacao{Tipo: "pauta", Titulo: "Pauta"})

		liberar, voto := segurarTransacao(t, db, func(r store.Repositorios) error {
			_, err := r.Votacoes.BuscarComTrava(t.Context(), id)
			return err
		})

		outroVoto := store.EmTransacao(t.Context(), db, func(r store.Repositorios) error {
			_, err := r.Votacoes.BuscarComTrava(t.Context(), id)
			return err
		})
		if outroVoto != nil {
			t.Errorf("votos não deveriam bloquear uns aos outros: %v", outroVoto)
		}

		encerramento := make(chan bool, 1)
		go func() {
			encerrou, err := repo.Votacoes.EncerrarSeAberta(t.Context(), id, instante.Add(30*time.Minute))
			if err != nil {
				t.Errorf("encerrar: %v", err)
			}
			encerramento <- encerrou
		}()
		select {
		case <-encerramento:
			t.Fatal("o encerramento terminou antes de a transação do voto confirmar")
		case <-time.After(esperaPouco):
		}

		close(liberar)
		if err := <-voto; err != nil {
			t.Fatal(err)
		}
		if !<-encerramento {
			t.Error("depois do voto confirmar, o encerramento deveria acontecer")
		}
	})

	t.Run("voto que espera o encerramento enxerga o fim novo", func(t *testing.T) {
		db, repo := novoBanco(t)
		id := inserirVotacao(t, repo, store.Votacao{Tipo: "pauta", Titulo: "Pauta"})
		novoFim := instante.Add(30 * time.Minute)

		liberar, encerramento := segurarTransacao(t, db, func(r store.Repositorios) error {
			encerrou, err := r.Votacoes.EncerrarSeAberta(t.Context(), id, novoFim)
			if !encerrou {
				t.Error("a votação deveria ter sido encerrada")
			}
			return err
		})

		fimVisto, voto := make(chan time.Time, 1), make(chan error, 1)
		go func() {
			voto <- store.EmTransacao(t.Context(), db, func(r store.Repositorios) error {
				v, err := r.Votacoes.BuscarComTrava(t.Context(), id)
				fimVisto <- v.Fim
				return err
			})
		}()
		select {
		case <-fimVisto:
			t.Fatal("o voto leu a votação antes de o encerramento confirmar")
		case <-time.After(esperaPouco):
		}

		close(liberar)
		if err := <-encerramento; err != nil {
			t.Fatal(err)
		}
		if fim := <-fimVisto; !fim.Equal(novoFim) {
			t.Errorf("fim visto pelo voto: esperado %v, obtido %v", novoFim, fim)
		}
		if err := <-voto; err != nil {
			t.Error(err)
		}
	})
}
