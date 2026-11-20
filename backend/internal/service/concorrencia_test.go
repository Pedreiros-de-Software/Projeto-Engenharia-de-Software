package service

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/voting"
)

const (
	eleitoresConcorrentes = 20
	rodadasConcorrentes   = 20
)

func eleitoresDeTeste(t *testing.T, c *cenario) []int64 {
	t.Helper()
	ids := make([]int64, eleitoresConcorrentes)
	for i := range ids {
		ids[i] = c.usuario(t, fmt.Sprintf("eleitor%02d", i), PapelEleitor)
	}
	return ids
}

// votarTodos dispara o voto em branco de cada eleitor ao mesmo tempo; o canal fecha quando todos terminam.
func votarTodos(t *testing.T, c *cenario, eleicaoID int64, eleitores []int64) <-chan struct{} {
	t.Helper()
	largada, fim := make(chan struct{}), make(chan struct{})
	var wg sync.WaitGroup
	for _, id := range eleitores {
		wg.Go(func() {
			<-largada
			err := c.svc.Votar(t.Context(), id, eleicaoID, Voto{Escolha: voting.OpcaoBranco})
			if err != nil && !errors.Is(err, ErrVotacaoNaoAberta) {
				t.Errorf("voto do eleitor %d: %v", id, err)
			}
		})
	}
	close(largada)
	go func() {
		wg.Wait()
		close(fim)
	}()
	return fim
}

// esperarAlgunsVotos espera um quarto dos votos confirmar (ou todos terminarem).
func esperarAlgunsVotos(t *testing.T, c *cenario, eleicaoID int64, votos <-chan struct{}) {
	t.Helper()
	for c.contar(t, "SELECT count(*) FROM participacoes WHERE votacao_id = $1", eleicaoID) < eleitoresConcorrentes/4 {
		select {
		case <-votos:
			return
		default:
		}
	}
}

func TestVotosConcorrentesComEncerramento(t *testing.T) {
	c := novoCenario(t)
	eleitores := eleitoresDeTeste(t, c)

	for rodada := range rodadasConcorrentes {
		eleicaoID := c.votacao(t, NovaVotacao{Tipo: TipoEleicao, Opcoes: opcoesNomeadas("C1", "C2"), Aptos: eleitores})
		votos := votarTodos(t, c, eleicaoID, eleitores)

		c.encerrar(t, eleicaoID)
		publicado, err := c.svc.Resultado(t.Context(), eleicaoID)
		if err != nil {
			t.Fatalf("rodada %d: resultado logo depois de encerrar: %v", rodada, err)
		}
		<-votos

		final, err := c.svc.Resultado(t.Context(), eleicaoID)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(publicado, final) {
			t.Errorf("rodada %d: o boletim mudou depois do encerramento:\n publicado %+v\n final     %+v",
				rodada, *publicado.Eleicao, *final.Eleicao)
		}
	}
}

func TestResultadoConsistenteComVotosNoFimDoPrazo(t *testing.T) {
	c := novoCenario(t)
	eleitores := eleitoresDeTeste(t, c)

	for rodada := range rodadasConcorrentes {
		eleicaoID := c.votacao(t, NovaVotacao{Tipo: TipoEleicao, Opcoes: opcoesNomeadas("C1", "C2"), Aptos: eleitores})
		votos := votarTodos(t, c, eleicaoID, eleitores)
		esperarAlgunsVotos(t, c, eleicaoID, votos)
		c.relogio.Avancar(2 * time.Hour) // o prazo acaba com votos ainda em andamento

		for terminou := false; !terminou; {
			select {
			case <-votos:
				terminou = true
			default:
			}
			if _, err := c.svc.Resultado(t.Context(), eleicaoID); err != nil {
				t.Fatalf("rodada %d: boletim durante votos em andamento: %v", rodada, err)
			}
		}
	}
}
