package service

import (
	"sync"
	"time"
)

// limitador conta tentativas por chave em janelas fixas, em memória.
type limitador struct {
	mu      sync.Mutex
	limite  int
	janela  time.Duration
	janelas map[string]janelaDeTentativas
}

type janelaDeTentativas struct {
	inicio     time.Time
	tentativas int
}

func novoLimitador(limite int, janela time.Duration) *limitador {
	return &limitador{limite: limite, janela: janela, janelas: map[string]janelaDeTentativas{}}
}

// permitir registra uma tentativa e informa se ela ainda cabe no limite da janela atual.
func (l *limitador) permitir(chave string, agora time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.descartarVencidas(agora)
	j := l.janelas[chave]
	if j.tentativas == 0 {
		j.inicio = agora
	}
	j.tentativas++
	l.janelas[chave] = j
	return j.tentativas <= l.limite
}

// devolver desconta uma tentativa registrada por permitir que não deve contar no limite.
func (l *limitador) devolver(chave string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	j, ok := l.janelas[chave]
	if !ok {
		return
	}
	j.tentativas--
	if j.tentativas <= 0 {
		delete(l.janelas, chave)
		return
	}
	l.janelas[chave] = j
}

// descartarVencidas remove as janelas já terminadas, para a memória não crescer sem limite.
func (l *limitador) descartarVencidas(agora time.Time) {
	for chave, j := range l.janelas {
		if agora.Sub(j.inicio) >= l.janela {
			delete(l.janelas, chave)
		}
	}
}
