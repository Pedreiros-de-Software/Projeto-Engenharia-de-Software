package service

import (
	"context"
	"errors"
	"time"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store"
)

// ListarVotacoes devolve todas as votações, das mais recentes para as mais antigas,
// com apto e ja_votou calculados para o usuário.
func (s *Service) ListarVotacoes(ctx context.Context, usuarioID int64) ([]VotacaoResumo, error) {
	votacoes, err := s.repo.Votacoes.Listar(ctx)
	if err != nil {
		return nil, err
	}
	situacao, err := s.carregarSituacao(ctx, usuarioID)
	if err != nil {
		return nil, err
	}

	agora := s.relogio()
	resumos := make([]VotacaoResumo, 0, len(votacoes))
	for _, v := range votacoes {
		resumo, err := resumir(v, situacao, agora)
		if err != nil {
			return nil, err
		}
		resumos = append(resumos, resumo)
	}
	return resumos, nil
}

// Detalhar devolve a votação com suas opções e, na pauta, a unidade do usuário.
func (s *Service) Detalhar(ctx context.Context, usuarioID, votacaoID int64) (VotacaoDetalhe, error) {
	v, err := buscarVotacao(ctx, s.repo, votacaoID)
	if err != nil {
		return VotacaoDetalhe{}, err
	}
	situacao, err := s.carregarSituacao(ctx, usuarioID)
	if err != nil {
		return VotacaoDetalhe{}, err
	}
	resumo, err := resumir(v, situacao, s.relogio())
	if err != nil {
		return VotacaoDetalhe{}, err
	}
	opcoes, err := s.repo.Votacoes.ListarOpcoes(ctx, v.ID)
	if err != nil {
		return VotacaoDetalhe{}, err
	}

	detalhe := VotacaoDetalhe{
		VotacaoResumo:     resumo,
		Opcoes:            make([]Opcao, len(opcoes)),
		OrcamentoCentavos: v.OrcamentoCentavos,
	}
	for i, o := range opcoes {
		detalhe.Opcoes[i] = paraOpcao(o)
	}
	if v.Tipo == TipoPauta && situacao.unidade != nil {
		detalhe.MinhaUnidade, err = s.minhaUnidade(ctx, *situacao.unidade)
	}
	return detalhe, err
}

func (s *Service) minhaUnidade(ctx context.Context, u store.Unidade) (*MinhaUnidade, error) {
	pesoTotal, err := s.repo.Unidades.PesoTotal(ctx)
	if err != nil {
		return nil, err
	}
	motivo := motivoDoBloqueio(u)
	return &MinhaUnidade{
		Identificador: u.Identificador,
		FracaoIdeal:   u.FracaoIdeal,
		Percentual:    percentualDoPeso(u, pesoTotal),
		Adimplente:    u.Adimplente,
		Elegivel:      motivo == "",
		Motivo:        motivo,
	}, nil
}

// situacaoDoUsuario reúne o que define apto e ja_votou de um usuário em qualquer votação.
type situacaoDoUsuario struct {
	unidade *store.Unidade // unidade sob responsabilidade do usuário, se houver
	aptas   map[int64]bool // eleições e consultas em que o usuário é apto
	votadas map[int64]bool // votações em que o usuário (ou sua unidade, na pauta) já votou
}

func (s *Service) carregarSituacao(ctx context.Context, usuarioID int64) (situacaoDoUsuario, error) {
	situacao := situacaoDoUsuario{aptas: map[int64]bool{}, votadas: map[int64]bool{}}

	unidade, err := s.repo.Unidades.BuscarPorResponsavel(ctx, usuarioID)
	switch {
	case err == nil:
		situacao.unidade = &unidade
		pautasVotadas, err := s.repo.Votos.IDsVotadasPelaUnidade(ctx, unidade.ID)
		if err != nil {
			return situacao, err
		}
		marcar(situacao.votadas, pautasVotadas)
	case !errors.Is(err, store.ErrNaoEncontrado):
		return situacao, err
	}

	aptas, err := s.repo.Votacoes.IDsAptas(ctx, usuarioID)
	if err != nil {
		return situacao, err
	}
	marcar(situacao.aptas, aptas)

	participadas, err := s.repo.Votos.IDsParticipadas(ctx, usuarioID)
	if err != nil {
		return situacao, err
	}
	marcar(situacao.votadas, participadas)
	return situacao, nil
}

// apto: na pauta, votam os responsáveis por unidade; nas demais, quem está na lista de aptos.
func (sit situacaoDoUsuario) apto(v store.Votacao) bool {
	if v.Tipo == TipoPauta {
		return sit.unidade != nil
	}
	return sit.aptas[v.ID]
}

func marcar(conjunto map[int64]bool, ids []int64) {
	for _, id := range ids {
		conjunto[id] = true
	}
}

func resumir(v store.Votacao, situacao situacaoDoUsuario, agora time.Time) (VotacaoResumo, error) {
	status, err := statusEm(v, agora)
	if err != nil {
		return VotacaoResumo{}, err
	}
	return VotacaoResumo{
		ID:                v.ID,
		Tipo:              v.Tipo,
		Titulo:            v.Titulo,
		Descricao:         v.Descricao,
		Inicio:            v.Inicio,
		Fim:               v.Fim,
		Turno:             v.Turno,
		Status:            status,
		Apto:              situacao.apto(v),
		JaVotou:           situacao.votadas[v.ID],
		VotacaoAnteriorID: v.VotacaoAnteriorID,
		SegundoTurnoID:    v.SegundoTurnoID,
	}, nil
}
