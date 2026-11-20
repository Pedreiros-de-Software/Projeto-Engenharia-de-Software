package service

import (
	"fmt"
	"testing"
	"time"
)

func TestEncerrar(t *testing.T) {
	c := novoCenario(t)
	agora := c.relogio.Agora()
	aberta := c.votacao(t, NovaVotacao{Tipo: TipoPauta})
	naoIniciada := c.votacao(t, NovaVotacao{Tipo: TipoPauta, Inicio: agora.Add(time.Hour), Fim: agora.Add(2 * time.Hour)})
	jaEncerrada := c.votacao(t, NovaVotacao{Tipo: TipoPauta, Inicio: agora.Add(-2 * time.Hour), Fim: agora.Add(-time.Hour)})
	abertaAgora := c.votacao(t, NovaVotacao{Tipo: TipoPauta, Inicio: agora, Fim: agora.Add(time.Hour)})

	testes := []struct {
		nome      string
		votacaoID int64
		esperado  error
	}{
		{"votação aberta é encerrada", aberta, nil},
		{"encerrar de novo", aberta, ErrVotacaoNaoAberta},
		{"votação não iniciada", naoIniciada, ErrVotacaoNaoAberta},
		{"votação já encerrada pelo prazo", jaEncerrada, ErrVotacaoNaoAberta},
		{"votação aberta neste instante ficaria com janela vazia", abertaAgora, ErrVotacaoNaoAberta},
		{"votação inexistente", 9999, ErrNaoEncontrado},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			conferirErro(t, c.svc.Encerrar(t.Context(), tc.votacaoID), tc.esperado)
		})
	}

	detalhe, err := c.svc.Detalhar(t.Context(), 0, aberta)
	if err != nil {
		t.Fatal(err)
	}
	if detalhe.Status != StatusEncerrada || !detalhe.Fim.Equal(agora) {
		t.Errorf("esperava encerrada com fim = agora (%v), obteve status %s e fim %v", agora, detalhe.Status, detalhe.Fim)
	}
}

func TestAbrirSegundoTurno(t *testing.T) {
	c := novoCenario(t)
	id, chapa := eleicaoEncerrada(t, c, []string{"C1", "C2", "C3"}, 8, []string{"C1", "C1", "C1", "C2", "C2", "C3", "branco"})
	agora := c.relogio.Agora()

	t.Run("fim no passado é rejeitado", func(t *testing.T) {
		_, err := c.svc.AbrirSegundoTurno(t.Context(), id, ptr(agora))
		conferirErro(t, err, errValidacao)
	})

	turno2, err := c.svc.AbrirSegundoTurno(t.Context(), id, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("cria o 2º turno com as duas chapas, os mesmos aptos e 24 horas de janela", func(t *testing.T) {
		detalhe, err := c.svc.Detalhar(t.Context(), 0, turno2)
		if err != nil {
			t.Fatal(err)
		}
		if detalhe.Titulo != "Votação de teste — 2º turno" || detalhe.Turno != 2 || detalhe.Status != StatusAberta ||
			!detalhe.Inicio.Equal(agora) || !detalhe.Fim.Equal(agora.Add(24*time.Hour)) ||
			detalhe.VotacaoAnteriorID == nil || *detalhe.VotacaoAnteriorID != id {
			t.Errorf("2º turno inesperado: %+v", detalhe.VotacaoResumo)
		}
		if len(detalhe.Opcoes) != 2 || detalhe.Opcoes[0].Nome != "C1" || detalhe.Opcoes[0].Descricao != "Descrição de C1" ||
			detalhe.Opcoes[1].Nome != "C2" || detalhe.Opcoes[0].ID == chapa["C1"] {
			t.Errorf("chapas inesperadas (devem ser cópias novas de C1 e C2): %+v", detalhe.Opcoes)
		}
		if n := c.contar(t, "SELECT count(*) FROM aptos WHERE votacao_id = $1", turno2); n != 8 {
			t.Errorf("aptos copiados: esperado 8, obtido %d", n)
		}
	})

	t.Run("o resultado do 1º turno aponta para o 2º turno", func(t *testing.T) {
		resultado, err := c.svc.Resultado(t.Context(), id)
		if err != nil {
			t.Fatal(err)
		}
		if st := resultado.Eleicao.SegundoTurno; st.VotacaoID == nil || *st.VotacaoID != turno2 {
			t.Errorf("segundo_turno.votacao_id: esperado %d, obtido %+v", turno2, st)
		}
	})

	t.Run("não abre um segundo 2º turno", func(t *testing.T) {
		_, err := c.svc.AbrirSegundoTurno(t.Context(), id, nil)
		conferirErro(t, err, ErrSegundoTurnoIndisponivel)
	})

	t.Run("vencedora definida no 2º turno", func(t *testing.T) {
		chapas := c.opcoes(t, turno2)
		c.relogio.Avancar(time.Minute)
		c.votar(t, idDoEstudante(t, c, 1), turno2, Voto{Escolha: EscolhaChapa, OpcaoID: &chapas[1]})
		c.votar(t, idDoEstudante(t, c, 2), turno2, Voto{Escolha: EscolhaChapa, OpcaoID: &chapas[1]})
		c.votar(t, idDoEstudante(t, c, 3), turno2, Voto{Escolha: EscolhaChapa, OpcaoID: &chapas[0]})
		c.encerrar(t, turno2)

		resultado, err := c.svc.Resultado(t.Context(), turno2)
		if err != nil {
			t.Fatal(err)
		}
		r := resultado.Eleicao
		if r.Turno != 2 || r.Vencedora == nil || r.Vencedora.Nome != "C2" || r.SegundoTurno.Necessario {
			t.Errorf("resultado do 2º turno inesperado: %+v", r)
		}
	})

	t.Run("2º turno não abre outro 2º turno", func(t *testing.T) {
		_, err := c.svc.AbrirSegundoTurno(t.Context(), turno2, nil)
		conferirErro(t, err, ErrSegundoTurnoIndisponivel)
	})
}

func idDoEstudante(t *testing.T, c *cenario, n int) int64 {
	t.Helper()
	var id int64
	if err := c.db.QueryRow("SELECT id FROM usuarios WHERE login = $1", fmt.Sprintf("estudante%d", n)).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestAbrirSegundoTurnoComFimInformado(t *testing.T) {
	c := novoCenario(t)
	id, _ := eleicaoEncerrada(t, c, []string{"C1", "C2", "C3"}, 5, []string{"C1", "C2", "C1", "C2", "C3"})
	fim := c.relogio.Agora().Add(3 * time.Hour)

	turno2, err := c.svc.AbrirSegundoTurno(t.Context(), id, &fim)
	if err != nil {
		t.Fatal(err)
	}
	detalhe, err := c.svc.Detalhar(t.Context(), 0, turno2)
	if err != nil || !detalhe.Fim.Equal(fim) {
		t.Errorf("fim: esperado %v, obtido %v (err %v)", fim, detalhe.Fim, err)
	}
}

func TestAbrirSegundoTurnoIndisponivel(t *testing.T) {
	testes := []struct {
		nome     string
		preparar func(t *testing.T, c *cenario) int64
		esperado error
	}{
		{"eleição com vencedora no 1º turno", func(t *testing.T, c *cenario) int64 {
			id, _ := eleicaoEncerrada(t, c, []string{"C1", "C2"}, 3, []string{"C1", "C1", "C2"})
			return id
		}, ErrSegundoTurnoIndisponivel},
		{"empate pelo 2º lugar", func(t *testing.T, c *cenario) int64 {
			id, _ := eleicaoEncerrada(t, c, []string{"C1", "C2", "C3"}, 4, []string{"C1", "C1", "C2", "C3"})
			return id
		}, ErrSegundoTurnoIndisponivel},
		{"eleição sem votos válidos", func(t *testing.T, c *cenario) int64 {
			id, _ := eleicaoEncerrada(t, c, []string{"C1", "C2"}, 2, []string{"branco"})
			return id
		}, ErrSegundoTurnoIndisponivel},
		{"eleição ainda aberta", func(t *testing.T, c *cenario) int64 {
			return c.votacao(t, NovaVotacao{Tipo: TipoEleicao, Opcoes: opcoesNomeadas("C1", "C2")})
		}, ErrSegundoTurnoIndisponivel},
		{"pauta encerrada", func(t *testing.T, c *cenario) int64 {
			id := c.votacao(t, NovaVotacao{Tipo: TipoPauta})
			c.encerrar(t, id)
			return id
		}, ErrSegundoTurnoIndisponivel},
		{"votação inexistente", func(*testing.T, *cenario) int64 { return 9999 }, ErrNaoEncontrado},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			c := novoCenario(t)
			_, err := c.svc.AbrirSegundoTurno(t.Context(), tc.preparar(t, c), nil)
			conferirErro(t, err, tc.esperado)
		})
	}
}

// No 2º turno, um empate não leva a outro turno: o resultado não pede 2º turno nem marca
// empate indefinido, e a decisão fica com a comissão.
func TestEmpateNoSegundoTurno(t *testing.T) {
	c := novoCenario(t)
	id, _ := eleicaoEncerrada(t, c, []string{"C1", "C2", "C3"}, 5, []string{"C1", "C2", "C1", "C2", "C3"})
	turno2, err := c.svc.AbrirSegundoTurno(t.Context(), id, nil)
	if err != nil {
		t.Fatal(err)
	}
	chapas := c.opcoes(t, turno2)
	c.relogio.Avancar(time.Minute)
	c.votar(t, idDoEstudante(t, c, 1), turno2, Voto{Escolha: EscolhaChapa, OpcaoID: &chapas[0]})
	c.votar(t, idDoEstudante(t, c, 2), turno2, Voto{Escolha: EscolhaChapa, OpcaoID: &chapas[1]})
	c.encerrar(t, turno2)

	resultado, err := c.svc.Resultado(t.Context(), turno2)
	if err != nil {
		t.Fatal(err)
	}
	r := resultado.Eleicao
	if r.Vencedora != nil || r.SegundoTurno.Necessario || r.SegundoTurno.EmpateIndefinido || len(r.SegundoTurno.Chapas) != 0 {
		t.Errorf("resultado inesperado: %+v", r)
	}
}
