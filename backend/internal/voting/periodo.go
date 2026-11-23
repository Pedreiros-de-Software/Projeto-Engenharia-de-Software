package voting

import (
	"errors"
	"time"
)

// EstadoVotacao indica a situação de uma votação em relação à sua janela.
type EstadoVotacao int

// VotacaoInvalida é o valor zero, devolvido junto com erro para janelas
// inválidas, de modo que um erro ignorado não seja lido como estado válido.
const (
	VotacaoInvalida EstadoVotacao = iota
	VotacaoNaoIniciada
	VotacaoAberta
	VotacaoEncerrada
)

func (e EstadoVotacao) String() string {
	switch e {
	case VotacaoInvalida:
		return "invalida"
	case VotacaoNaoIniciada:
		return "nao iniciada"
	case VotacaoAberta:
		return "aberta"
	case VotacaoEncerrada:
		return "encerrada"
	}
	return "desconhecido"
}

var ErrJanelaInvalida = errors.New("janela de votacao invalida")

// StatusVotacao classifica o instante agora em relação à janela [inicio, fim):
// a votação abre no início e encerra exatamente no fim. Datas zeradas ou fim
// anterior ou igual ao início são rejeitados.
func StatusVotacao(inicio, fim, agora time.Time) (EstadoVotacao, error) {
	if inicio.IsZero() || fim.IsZero() || agora.IsZero() || !fim.After(inicio) {
		return VotacaoInvalida, ErrJanelaInvalida
	}
	if agora.Before(inicio) {
		return VotacaoNaoIniciada, nil
	}
	if agora.Before(fim) {
		return VotacaoAberta, nil
	}
	return VotacaoEncerrada, nil
}
