package voting

import (
	"errors"
	"time"
)

// EstadoVotacao indica a situação de uma votação em relação à sua janela.
type EstadoVotacao int

const (
	NaoIniciada EstadoVotacao = iota
	Aberta
	Encerrada
)

func (e EstadoVotacao) String() string {
	switch e {
	case NaoIniciada:
		return "nao iniciada"
	case Aberta:
		return "aberta"
	case Encerrada:
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
		return NaoIniciada, ErrJanelaInvalida
	}
	if agora.Before(inicio) {
		return NaoIniciada, nil
	}
	if agora.Before(fim) {
		return Aberta, nil
	}
	return Encerrada, nil
}
