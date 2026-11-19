package seed

import (
	"database/sql"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/service"
	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store/storetest"
)

const (
	tituloPautaEncerrada = "Pauta #01 — Troca do sistema de câmeras"
	tituloPautaAberta    = "Pauta #02 — Aprovação do orçamento para reforma da portaria (R$ 45.000,00)"
	tituloEleicao        = "Eleição Diretoria DCE 2026"
	tituloConsulta       = "Orçamento participativo — Laboratório de Inovação"
)

// relogio permite avançar o tempo para encerrar as votações abertas do seed.
type relogio struct{ agora time.Time }

func (r *relogio) Agora() time.Time { return r.agora }

func novoServico(t *testing.T) (*service.Service, *relogio, *sql.DB) {
	t.Helper()
	db := storetest.NovoBanco(t)
	r := &relogio{agora: time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)}
	return service.Novo(db, service.Config{Relogio: r.Agora, IteracoesSenha: 1000}), r, db
}

func contar(t *testing.T, db *sql.DB, consulta string) int {
	t.Helper()
	var total int
	if err := db.QueryRow(consulta).Scan(&total); err != nil {
		t.Fatal(err)
	}
	return total
}

func TestExecutarCriaOCenarioDaDemonstracao(t *testing.T) {
	svc, relogio, db := novoServico(t)
	ctx := t.Context()

	criado, err := Executar(ctx, svc)
	if err != nil || !criado {
		t.Fatalf("Executar: criado=%v err=%v", criado, err)
	}

	t.Run("usuários com as senhas de demonstração", func(t *testing.T) {
		if n := contar(t, db, "SELECT count(*) FROM usuarios"); n != 15 {
			t.Errorf("usuários: esperado 15, obtido %d", n)
		}
		logins := []struct{ login, senha, papel string }{
			{"admin", "admin123", service.PapelAdmin},
			{"maria", "consensu", service.PapelEleitor},
			{"hugo", "consensu", service.PapelEleitor},
		}
		for i, l := range logins {
			sessao, err := svc.Login(ctx, l.login, l.senha, "seed-teste")
			if err != nil || sessao.Usuario.Papel != l.papel {
				t.Errorf("login %s: %+v (err %v)", l.login, sessao.Usuario, err)
			}
			if i == 0 && sessao.Usuario.Nome != "Administração" {
				t.Errorf("nome do admin: %q", sessao.Usuario.Nome)
			}
		}
	})

	t.Run("unidades do Edifício Solar somam 500 com o Apto 402 inadimplente", func(t *testing.T) {
		unidades, err := svc.ListarUnidades(ctx)
		if err != nil || len(unidades) != 6 {
			t.Fatalf("unidades: %+v (err %v)", unidades, err)
		}
		total := 0.0
		for _, u := range unidades {
			total += u.FracaoIdeal
			if inadimplente := u.Identificador == "Apto 402"; u.Adimplente == inadimplente {
				t.Errorf("%s: adimplente=%v", u.Identificador, u.Adimplente)
			}
		}
		if total != 500 || unidades[5].Responsavel == nil || unidades[5].Responsavel.Nome != "Sofia Prado" {
			t.Errorf("total %v, última unidade %+v", total, unidades[5])
		}
	})

	adminID := idDoLogin(t, db, "admin")
	votacoes := votacoesPorTitulo(t, svc, adminID)

	t.Run("quatro votações com os status esperados", func(t *testing.T) {
		esperado := map[string]string{
			tituloPautaEncerrada: service.StatusEncerrada,
			tituloPautaAberta:    service.StatusAberta,
			tituloEleicao:        service.StatusEncerrada,
			tituloConsulta:       service.StatusAberta,
		}
		if len(votacoes) != len(esperado) {
			t.Fatalf("esperava %d votações, obteve %d", len(esperado), len(votacoes))
		}
		for titulo, status := range esperado {
			if votacoes[titulo].Status != status {
				t.Errorf("%s: esperado %s, obtido %s", titulo, status, votacoes[titulo].Status)
			}
		}
	})

	t.Run("pauta #01 reprovada com 300 de 500", func(t *testing.T) {
		resultado, err := svc.Resultado(ctx, votacoes[tituloPautaEncerrada].ID)
		if err != nil {
			t.Fatal(err)
		}
		p := resultado.Pauta
		if p.Favor != 300 || p.Contra != 80 || p.Abstencao != 0 || p.PesoTotal != 500 || p.PercentualFavor != 60 ||
			p.Aprovada || p.UnidadesTotal != 6 || p.UnidadesVotantes != 5 {
			t.Errorf("resultado inesperado: %+v", p)
		}
	})

	t.Run("eleição de 1º turno pede 2º turno entre Chapa 1 e Chapa 2", func(t *testing.T) {
		resultado, err := svc.Resultado(ctx, votacoes[tituloEleicao].ID)
		if err != nil {
			t.Fatal(err)
		}
		e := resultado.Eleicao
		if e.Turno != 1 || e.Aptos != 8 || e.Participantes != 7 || e.Abstencoes != 1 || e.PercentualComparecimento != 87.5 ||
			e.Brancos != 1 || e.Nulos != 0 || e.Validos != 6 || e.Vencedora != nil {
			t.Errorf("boletim inesperado: %+v", e)
		}
		var votos []int
		for _, v := range e.Votos {
			votos = append(votos, v.Votos)
		}
		if !slices.Equal(votos, []int{3, 2, 1}) || e.Votos[0].Nome != "Chapa 1 — Movimenta" || e.Votos[0].PercentualValidos != 50 {
			t.Errorf("votos por chapa inesperados: %+v", e.Votos)
		}
		st := e.SegundoTurno
		if !st.Necessario || st.EmpateIndefinido || st.VotacaoID != nil || len(st.Chapas) != 2 ||
			st.Chapas[0].Nome != "Chapa 1 — Movimenta" || st.Chapas[1].Nome != "Chapa 2 — Renovação" {
			t.Errorf("segundo turno inesperado: %+v", st)
		}
	})

	t.Run("votações abertas já têm os votos do seed", func(t *testing.T) {
		mariaID, rafaelID, hugoID := idDoLogin(t, db, "maria"), idDoLogin(t, db, "rafael"), idDoLogin(t, db, "hugo")
		situacoes := []struct {
			usuarioID     int64
			titulo        string
			apto, jaVotou bool
		}{
			{mariaID, tituloPautaAberta, true, false},
			{rafaelID, tituloPautaAberta, true, true},
			{mariaID, tituloConsulta, true, true},
			{hugoID, tituloConsulta, true, false},
		}
		for _, s := range situacoes {
			v := votacoesPorTitulo(t, svc, s.usuarioID)[s.titulo]
			if v.Apto != s.apto || v.JaVotou != s.jaVotou {
				t.Errorf("usuário %d em %q: apto=%v ja_votou=%v", s.usuarioID, s.titulo, v.Apto, v.JaVotou)
			}
		}
		if n := contar(t, db, "SELECT count(*) FROM aptos a JOIN votacoes v ON v.id = a.votacao_id WHERE v.tipo = 'consulta'"); n != 14 {
			t.Errorf("aptos da consulta: esperado 14 (todos os eleitores), obtido %d", n)
		}
	})

	relogio.agora = relogio.agora.Add(8 * 24 * time.Hour)

	t.Run("pauta #02 tem 180 a favor antes da demonstração", func(t *testing.T) {
		resultado, err := svc.Resultado(ctx, votacoes[tituloPautaAberta].ID)
		if err != nil {
			t.Fatal(err)
		}
		if p := resultado.Pauta; p.Favor != 180 || p.UnidadesVotantes != 2 || p.Aprovada {
			t.Errorf("resultado inesperado: %+v", p)
		}
	})

	t.Run("consulta com 5 cédulas: Impressoras e Arduino dentro do orçamento", func(t *testing.T) {
		resultado, err := svc.Resultado(ctx, votacoes[tituloConsulta].ID)
		if err != nil {
			t.Fatal(err)
		}
		c := resultado.Consulta
		var nomes []string
		var pontos []int
		for _, item := range c.Ranking {
			nomes = append(nomes, item.Nome)
			pontos = append(pontos, item.Pontos)
		}
		if c.Cedulas != 5 || c.OrcamentoCentavos != 60000_00 || c.TotalSelecionadoCentavos != 47000_00 || c.SaldoCentavos != 13000_00 ||
			!slices.Equal(pontos, []int{16, 14, 12, 8}) ||
			!slices.Equal(nomes, []string{"Impressoras 3D e CNC", "Kit Arduino e Robótica", "Servidor GPU para IA", "Mobiliário colaborativo"}) ||
			!c.Ranking[1].DentroDoOrcamento || c.Ranking[2].DentroDoOrcamento || c.Ranking[3].DentroDoOrcamento {
			t.Errorf("resultado inesperado: %+v", c)
		}
	})
}

func TestExecutarNaoRepeteQuandoJaHaUsuarios(t *testing.T) {
	svc, _, db := novoServico(t)
	if criado, err := Executar(t.Context(), svc); err != nil || !criado {
		t.Fatalf("primeira execução: criado=%v err=%v", criado, err)
	}
	criado, err := Executar(t.Context(), svc)
	if err != nil || criado {
		t.Errorf("segunda execução: esperava criado=false sem erro, obteve criado=%v err=%v", criado, err)
	}
	if n := contar(t, db, "SELECT count(*) FROM votacoes"); n != 4 {
		t.Errorf("votações: esperado 4, obtido %d", n)
	}
}

func TestExecutarRecusaSeedIncompleto(t *testing.T) {
	svc, _, db := novoServico(t)
	if _, err := svc.CriarUsuario(t.Context(), "Administração", "admin", "admin123", service.PapelAdmin); err != nil {
		t.Fatal(err)
	}

	criado, err := Executar(t.Context(), svc)
	if criado || !errors.Is(err, ErrSeedIncompleto) {
		t.Errorf("esperava criado=false e ErrSeedIncompleto, obteve criado=%v err=%v", criado, err)
	}
	if n := contar(t, db, "SELECT count(*) FROM usuarios"); n != 1 {
		t.Errorf("o seed incompleto não deveria ser alterado: %d usuários", n)
	}
}

func idDoLogin(t *testing.T, db *sql.DB, login string) int64 {
	t.Helper()
	var id int64
	if err := db.QueryRow("SELECT id FROM usuarios WHERE login = $1", login).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func votacoesPorTitulo(t *testing.T, svc *service.Service, usuarioID int64) map[string]service.VotacaoResumo {
	t.Helper()
	lista, err := svc.ListarVotacoes(t.Context(), usuarioID)
	if err != nil {
		t.Fatal(err)
	}
	porTitulo := map[string]service.VotacaoResumo{}
	for _, v := range lista {
		porTitulo[v.Titulo] = v
	}
	return porTitulo
}
