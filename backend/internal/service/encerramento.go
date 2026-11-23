package service

import (
	"context"
	"errors"
	"time"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store"
)

// Encerrar fecha agora uma votação aberta, movendo o fim da janela para o instante atual.
// Votos em andamento confirmam antes; os seguintes já encontram a votação encerrada.
func (s *Service) Encerrar(ctx context.Context, votacaoID int64) error {
	v, err := buscarVotacao(ctx, s.repo, votacaoID)
	if err != nil {
		return err
	}
	agora := s.relogio()
	status, err := statusEm(v, agora)
	if err != nil {
		return err
	}
	if status != StatusAberta {
		return ErrVotacaoNaoAberta
	}
	encerrou, err := s.repo.Votacoes.EncerrarSeAberta(ctx, v.ID, agora)
	if err != nil {
		return err
	}
	if !encerrou {
		return ErrVotacaoNaoAberta
	}
	return nil
}

// AbrirSegundoTurno cria o 2º turno de uma eleição de 1º turno encerrada sem maioria absoluta:
// as duas chapas mais votadas, os mesmos aptos e a janela de agora até fim (padrão: 24 horas).
func (s *Service) AbrirSegundoTurno(ctx context.Context, votacaoID int64, fim *time.Time) (int64, error) {
	v, err := buscarVotacao(ctx, s.repo, votacaoID)
	if err != nil {
		return 0, err
	}
	agora := s.relogio()
	status, err := statusEm(v, agora)
	if err != nil {
		return 0, err
	}
	if v.Tipo != TipoEleicao || v.Turno != 1 || status != StatusEncerrada || v.SegundoTurnoID != nil {
		return 0, ErrSegundoTurnoIndisponivel
	}

	apuracao, err := apurarEleicao(ctx, s.repo, v)
	if err != nil {
		return 0, err
	}
	if len(apuracao.resultado.SegundoTurno) != 2 {
		return 0, ErrSegundoTurnoIndisponivel
	}

	fimDoTurno := agora.Add(duracaoPadraoSegundoTurno)
	if fim != nil {
		fimDoTurno = *fim
	}
	if !fimDoTurno.After(agora) {
		return 0, invalido("O término do 2º turno precisa ser no futuro.")
	}

	var id int64
	err = store.EmTransacao(ctx, s.db, func(r store.Repositorios) error {
		var err error
		id, err = r.Votacoes.Inserir(ctx, store.Votacao{
			Tipo:              TipoEleicao,
			Titulo:            v.Titulo + " — 2º turno",
			Descricao:         v.Descricao,
			Inicio:            agora,
			Fim:               fimDoTurno,
			Turno:             2,
			VotacaoAnteriorID: &v.ID,
		})
		if errors.Is(err, store.ErrDuplicado) {
			return ErrSegundoTurnoIndisponivel // outro pedido abriu o 2º turno ao mesmo tempo
		}
		if err != nil {
			return err
		}
		for i, chave := range apuracao.resultado.SegundoTurno {
			chapa := apuracao.chapas[chave]
			opcao := store.Opcao{VotacaoID: id, Nome: chapa.Nome, Descricao: chapa.Descricao, Posicao: i + 1}
			if _, err := r.Votacoes.InserirOpcao(ctx, opcao); err != nil {
				return err
			}
		}
		return r.Votacoes.CopiarAptos(ctx, v.ID, id)
	})
	return id, err
}
