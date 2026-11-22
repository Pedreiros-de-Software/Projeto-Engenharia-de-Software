package service

import (
	"context"
	"errors"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store"
	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/voting"
)

const tipoCedulaRanking = "ranking"

// Voto é o que o eleitor envia; cada tipo de votação usa só os seus campos.
type Voto struct {
	Escolha      string  // pauta: favor | contra | abstencao; eleição: chapa | branco | nulo
	OpcaoID      *int64  // eleição, quando Escolha é chapa
	Preferencias []int64 // consulta: ids de todas as opções, da preferida para a menos preferida
}

// Votar registra o voto do usuário numa votação aberta. A janela é conferida de novo dentro da
// transação do voto, com a votação travada contra o encerramento (ver travarAberta).
func (s *Service) Votar(ctx context.Context, usuarioID, votacaoID int64, voto Voto) error {
	v, err := buscarVotacao(ctx, s.repo, votacaoID)
	if err != nil {
		return err
	}
	status, err := statusEm(v, s.relogio())
	if err != nil {
		return err
	}
	if status != StatusAberta {
		return ErrVotacaoNaoAberta
	}

	switch v.Tipo {
	case TipoPauta:
		return s.votarPauta(ctx, usuarioID, v, voto)
	case TipoEleicao:
		return s.votarEleicao(ctx, usuarioID, v, voto)
	default:
		return s.votarConsulta(ctx, usuarioID, v, voto)
	}
}

// votarPauta grava o voto nominal da unidade do usuário; o peso vem da fração ideal na apuração.
func (s *Service) votarPauta(ctx context.Context, usuarioID int64, v store.Votacao, voto Voto) error {
	unidade, err := s.repo.Unidades.BuscarPorResponsavel(ctx, usuarioID)
	if errors.Is(err, store.ErrNaoEncontrado) {
		return ErrNaoApto
	}
	if err != nil {
		return err
	}
	if err := voting.ValidarElegibilidade(condominoDe(unidade)); err != nil {
		if errors.Is(err, voting.ErrInadimplente) {
			return ErrInadimplente
		}
		return err
	}

	escolhaValida := voto.Escolha == voting.OpcaoFavor || voto.Escolha == voting.OpcaoContra || voto.Escolha == voting.OpcaoAbstencao
	if !escolhaValida || voto.OpcaoID != nil || len(voto.Preferencias) > 0 {
		return invalido("Na pauta, envie apenas escolha: favor, contra ou abstencao.")
	}

	return store.EmTransacao(ctx, s.db, func(r store.Repositorios) error {
		if err := s.travarAberta(ctx, r, v.ID); err != nil {
			return err
		}
		err := r.Votos.InserirVotoPauta(ctx, v.ID, unidade.ID, voto.Escolha)
		if errors.Is(err, store.ErrDuplicado) {
			return ErrVotoDuplicado
		}
		return err
	})
}

// votarEleicao grava, na mesma transação, a cédula anônima e a participação do eleitor.
func (s *Service) votarEleicao(ctx context.Context, usuarioID int64, v store.Votacao, voto Voto) error {
	if err := s.exigirApto(ctx, v.ID, usuarioID); err != nil {
		return err
	}
	tipoCedula, err := s.validarVotoEleicao(ctx, v.ID, voto)
	if err != nil {
		return err
	}

	return store.EmTransacao(ctx, s.db, func(r store.Repositorios) error {
		if err := s.travarAberta(ctx, r, v.ID); err != nil {
			return err
		}
		if _, err := r.Votos.InserirCedula(ctx, v.ID, tipoCedula, voto.OpcaoID); err != nil {
			return err
		}
		return registrarParticipacao(ctx, r, v.ID, usuarioID)
	})
}

// validarVotoEleicao devolve o tipo da cédula (chapa, branco ou nulo).
func (s *Service) validarVotoEleicao(ctx context.Context, votacaoID int64, voto Voto) (string, error) {
	if len(voto.Preferencias) > 0 {
		return "", invalido("Na eleição, envie escolha (chapa, branco ou nulo) e, para chapa, opcao_id.")
	}

	switch voto.Escolha {
	case voting.OpcaoBranco, voting.OpcaoNulo:
		if voto.OpcaoID != nil {
			return "", invalido("Voto em branco ou nulo não leva opcao_id.")
		}
		return voto.Escolha, nil
	case EscolhaChapa:
		if voto.OpcaoID == nil {
			return "", invalido("Informe opcao_id da chapa escolhida.")
		}
		opcoes, err := s.repo.Votacoes.ListarOpcoes(ctx, votacaoID)
		if err != nil {
			return "", err
		}
		for _, o := range opcoes {
			if o.ID == *voto.OpcaoID {
				return EscolhaChapa, nil
			}
		}
		return "", ErrOpcaoDesconhecida
	default:
		return "", invalido("Na eleição, escolha chapa, branco ou nulo.")
	}
}

// votarConsulta valida a ordenação completa (Borda) e grava cédula, preferências e participação
// na mesma transação.
func (s *Service) votarConsulta(ctx context.Context, usuarioID int64, v store.Votacao, voto Voto) error {
	if err := s.exigirApto(ctx, v.ID, usuarioID); err != nil {
		return err
	}
	if voto.Escolha != "" || voto.OpcaoID != nil {
		return invalido("Na priorização, envie apenas preferencias.")
	}

	opcoes, err := s.repo.Votacoes.ListarOpcoes(ctx, v.ID)
	if err != nil {
		return err
	}
	idsOpcoes := make([]int64, len(opcoes))
	for i, o := range opcoes {
		idsOpcoes[i] = o.ID
	}
	err = voting.ValidarCedulaBorda(idsTexto(idsOpcoes), idsTexto(voto.Preferencias))
	if errors.Is(err, voting.ErrVotoInvalido) {
		return ErrCedulaInvalida
	}
	if err != nil {
		return err
	}

	return store.EmTransacao(ctx, s.db, func(r store.Repositorios) error {
		if err := s.travarAberta(ctx, r, v.ID); err != nil {
			return err
		}
		cedulaID, err := r.Votos.InserirCedula(ctx, v.ID, tipoCedulaRanking, nil)
		if err != nil {
			return err
		}
		for posicao, opcaoID := range voto.Preferencias {
			if err := r.Votos.InserirPreferencia(ctx, cedulaID, posicao+1, opcaoID); err != nil {
				return err
			}
		}
		return registrarParticipacao(ctx, r, v.ID, usuarioID)
	})
}

// travarAberta é o primeiro passo da transação do voto: trava a votação, de modo que Encerrar
// espera este voto confirmar, e só então lê o relógio e confere que a votação continua aberta.
// Um voto que esperou o encerramento lê o fim novo e é recusado.
func (s *Service) travarAberta(ctx context.Context, r store.Repositorios, votacaoID int64) error {
	v, err := r.Votacoes.BuscarComTrava(ctx, votacaoID)
	if err != nil {
		return err
	}
	status, err := statusEm(v, s.relogio())
	if err != nil {
		return err
	}
	if status != StatusAberta {
		return ErrVotacaoNaoAberta
	}
	return nil
}

func (s *Service) exigirApto(ctx context.Context, votacaoID, usuarioID int64) error {
	apto, err := s.repo.Votacoes.EstaApto(ctx, votacaoID, usuarioID)
	if err != nil {
		return err
	}
	if !apto {
		return ErrNaoApto
	}
	return nil
}

// registrarParticipacao é gravada depois da cédula: se o eleitor já votou, a falha aqui
// desfaz a transação inteira e nenhuma cédula extra fica no banco.
func registrarParticipacao(ctx context.Context, r store.Repositorios, votacaoID, usuarioID int64) error {
	err := r.Votos.InserirParticipacao(ctx, votacaoID, usuarioID)
	if errors.Is(err, store.ErrDuplicado) {
		return ErrVotoDuplicado
	}
	return err
}
