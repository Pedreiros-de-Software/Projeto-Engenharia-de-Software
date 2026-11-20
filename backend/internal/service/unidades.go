package service

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store"
	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/voting"
)

// MotivoInadimplente é o texto mostrado ao responsável por uma unidade bloqueada.
const MotivoInadimplente = "Unidade inadimplente: o voto fica bloqueado até a regularização com a administração."

// CriarUnidade cadastra uma unidade; responsavelID pode ser nil (unidade sem responsável).
func (s *Service) CriarUnidade(ctx context.Context, identificador string, fracaoIdeal float64, adimplente bool, responsavelID *int64) (int64, error) {
	identificador = strings.TrimSpace(identificador)
	if !tamanhoEntre(identificador, 1, 60) {
		return 0, invalido("O identificador da unidade deve ter entre 1 e 60 caracteres.")
	}
	if !(fracaoIdeal > 0) || math.IsInf(fracaoIdeal, 0) {
		return 0, invalido("A fração ideal deve ser um número maior que zero.")
	}

	id, err := s.repo.Unidades.Inserir(ctx, store.Unidade{
		Identificador: identificador, FracaoIdeal: fracaoIdeal, Adimplente: adimplente, ResponsavelID: responsavelID,
	})
	if errors.Is(err, store.ErrDuplicado) {
		return 0, invalido("Identificador já cadastrado ou usuário já responsável por outra unidade.")
	}
	return id, err
}

func (s *Service) ListarUnidades(ctx context.Context) ([]Unidade, error) {
	unidades, err := s.repo.Unidades.Listar(ctx)
	if err != nil {
		return nil, err
	}
	pesoTotal, err := s.repo.Unidades.PesoTotal(ctx)
	if err != nil {
		return nil, err
	}

	resultado := make([]Unidade, len(unidades))
	for i, u := range unidades {
		resultado[i] = paraUnidade(u, pesoTotal)
	}
	return resultado, nil
}

// AlterarAdimplencia marca a unidade como adimplente ou inadimplente e devolve o novo estado.
func (s *Service) AlterarAdimplencia(ctx context.Context, unidadeID int64, adimplente bool) (Unidade, error) {
	err := s.repo.Unidades.AlterarAdimplencia(ctx, unidadeID, adimplente)
	if errors.Is(err, store.ErrNaoEncontrado) {
		return Unidade{}, ErrNaoEncontrado
	}
	if err != nil {
		return Unidade{}, err
	}
	return s.unidade(ctx, unidadeID)
}

// UnidadeDoUsuario devolve a unidade sob responsabilidade do usuário, ou nil se não houver.
func (s *Service) UnidadeDoUsuario(ctx context.Context, usuarioID int64) (*Unidade, error) {
	u, err := s.repo.Unidades.BuscarPorResponsavel(ctx, usuarioID)
	if errors.Is(err, store.ErrNaoEncontrado) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	unidade, err := s.unidade(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	return &unidade, nil
}

func (s *Service) unidade(ctx context.Context, id int64) (Unidade, error) {
	u, err := s.repo.Unidades.Buscar(ctx, id)
	if err != nil {
		return Unidade{}, err
	}
	pesoTotal, err := s.repo.Unidades.PesoTotal(ctx)
	if err != nil {
		return Unidade{}, err
	}
	return paraUnidade(u, pesoTotal), nil
}

func paraUnidade(u store.Unidade, pesoTotal float64) Unidade {
	unidade := Unidade{
		ID:            u.ID,
		Identificador: u.Identificador,
		FracaoIdeal:   u.FracaoIdeal,
		Percentual:    percentualDoPeso(u, pesoTotal),
		Adimplente:    u.Adimplente,
	}
	if u.ResponsavelID != nil {
		unidade.Responsavel = &Responsavel{ID: *u.ResponsavelID, Nome: u.ResponsavelNome}
	}
	return unidade
}

// percentualDoPeso é a participação da unidade no peso total; só unidades com responsável votam.
func percentualDoPeso(u store.Unidade, pesoTotal float64) float64 {
	if u.ResponsavelID == nil {
		return 0
	}
	return percentual(u.FracaoIdeal, pesoTotal)
}

// condominoDe converte a unidade no condômino do domínio, identificado pelo id da unidade.
func condominoDe(u store.Unidade) voting.Condomino {
	return voting.Condomino{ID: idTexto(u.ID), Adimplente: u.Adimplente, FracaoIdeal: u.FracaoIdeal}
}

// motivoDoBloqueio explica por que a unidade não pode votar, ou devolve "" se ela pode.
func motivoDoBloqueio(u store.Unidade) string {
	err := voting.ValidarElegibilidade(condominoDe(u))
	switch {
	case err == nil:
		return ""
	case errors.Is(err, voting.ErrInadimplente):
		return MotivoInadimplente
	default:
		return "Unidade com cadastro inválido: procure a administração."
	}
}
