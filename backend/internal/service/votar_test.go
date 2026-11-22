package service

import (
	"slices"
	"testing"
	"time"
)

func TestVotarExigeVotacaoAberta(t *testing.T) {
	c := novoCenario(t)
	mariaID := c.usuario(t, "maria", PapelEleitor)
	c.unidade(t, "Apto 101", 60, true, mariaID)

	agora := c.relogio.Agora()
	naoIniciada := c.votacao(t, NovaVotacao{Tipo: TipoPauta, Inicio: agora.Add(time.Minute), Fim: agora.Add(time.Hour)})
	encerrada := c.votacao(t, NovaVotacao{Tipo: TipoPauta, Inicio: agora.Add(-time.Hour), Fim: agora})
	abrindoAgora := c.votacao(t, NovaVotacao{Tipo: TipoPauta, Inicio: agora, Fim: agora.Add(time.Hour)})

	testes := []struct {
		nome      string
		votacaoID int64
		esperado  error
	}{
		{"votação ainda não iniciada", naoIniciada, ErrVotacaoNaoAberta},
		{"votação no instante exato do fim já está encerrada", encerrada, ErrVotacaoNaoAberta},
		{"votação no instante exato do início já está aberta", abrindoAgora, nil},
		{"votação inexistente", 9999, ErrNaoEncontrado},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			err := c.svc.Votar(t.Context(), mariaID, tc.votacaoID, Voto{Escolha: "favor"})
			conferirErro(t, err, tc.esperado)
		})
	}
}

func TestVotarPauta(t *testing.T) {
	c := novoCenario(t)
	mariaID := c.usuario(t, "maria", PapelEleitor)
	joaoID := c.usuario(t, "joao", PapelEleitor)
	sofiaID := c.usuario(t, "sofia", PapelEleitor)
	anaID := c.usuario(t, "ana", PapelEleitor)
	c.unidade(t, "Apto 101", 60, true, mariaID)
	c.unidade(t, "Apto 102", 60, true, joaoID)
	c.unidade(t, "Apto 402", 120, false, sofiaID)
	pautaID := c.votacao(t, NovaVotacao{Tipo: TipoPauta})

	// Os passos rodam em ordem: o voto repetido depende do primeiro voto da maria.
	passos := []struct {
		nome      string
		usuarioID int64
		voto      Voto
		esperado  error
	}{
		{"responsável adimplente vota a favor", mariaID, Voto{Escolha: "favor"}, nil},
		{"segundo voto da mesma unidade", mariaID, Voto{Escolha: "contra"}, ErrVotoDuplicado},
		{"unidade inadimplente é bloqueada", sofiaID, Voto{Escolha: "favor"}, ErrInadimplente},
		{"usuário sem unidade não é apto", anaID, Voto{Escolha: "favor"}, ErrNaoApto},
		{"escolha fora de favor/contra/abstencao", joaoID, Voto{Escolha: "sim"}, errValidacao},
		{"escolha vazia", joaoID, Voto{}, errValidacao},
		{"pauta não aceita opcao_id", joaoID, Voto{Escolha: "favor", OpcaoID: ptr[int64](1)}, errValidacao},
		{"pauta não aceita preferencias", joaoID, Voto{Escolha: "favor", Preferencias: []int64{1}}, errValidacao},
		{"abstenção é um voto válido", joaoID, Voto{Escolha: "abstencao"}, nil},
	}
	for _, p := range passos {
		t.Run(p.nome, func(t *testing.T) {
			conferirErro(t, c.svc.Votar(t.Context(), p.usuarioID, pautaID, p.voto), p.esperado)
		})
	}

	if total := c.contar(t, "SELECT count(*) FROM votos_pauta WHERE votacao_id = $1", pautaID); total != 2 {
		t.Errorf("votos gravados: esperado 2, obtido %d", total)
	}
}

func TestVotarEleicao(t *testing.T) {
	c := novoCenario(t)
	anaID := c.usuario(t, "ana", PapelEleitor)
	brunoID := c.usuario(t, "bruno", PapelEleitor)
	carlaID := c.usuario(t, "carla", PapelEleitor)
	diegoID := c.usuario(t, "diego", PapelEleitor)
	eleicaoID := c.votacao(t, NovaVotacao{Tipo: TipoEleicao, Opcoes: opcoesNomeadas("C1", "C2"), Aptos: []int64{anaID, brunoID, diegoID}})
	outraEleicao := c.votacao(t, NovaVotacao{Tipo: TipoEleicao, Opcoes: opcoesNomeadas("X", "Y")})
	chapa1 := c.opcoes(t, eleicaoID)[0]
	chapaDeOutraEleicao := c.opcoes(t, outraEleicao)[0]

	passos := []struct {
		nome      string
		usuarioID int64
		voto      Voto
		esperado  error
	}{
		{"voto em chapa", anaID, Voto{Escolha: "chapa", OpcaoID: &chapa1}, nil},
		{"segundo voto do mesmo eleitor", anaID, Voto{Escolha: "branco"}, ErrVotoDuplicado},
		{"eleitor fora da lista de aptos", carlaID, Voto{Escolha: "branco"}, ErrNaoApto},
		{"chapa de outra votação", brunoID, Voto{Escolha: "chapa", OpcaoID: &chapaDeOutraEleicao}, ErrOpcaoDesconhecida},
		{"chapa sem opcao_id", brunoID, Voto{Escolha: "chapa"}, errValidacao},
		{"branco com opcao_id", brunoID, Voto{Escolha: "branco", OpcaoID: &chapa1}, errValidacao},
		{"escolha de pauta numa eleição", brunoID, Voto{Escolha: "favor"}, errValidacao},
		{"eleição não aceita preferencias", brunoID, Voto{Escolha: "branco", Preferencias: []int64{chapa1}}, errValidacao},
		{"voto em branco", brunoID, Voto{Escolha: "branco"}, nil},
		{"voto nulo", diegoID, Voto{Escolha: "nulo"}, nil},
	}
	for _, p := range passos {
		t.Run(p.nome, func(t *testing.T) {
			conferirErro(t, c.svc.Votar(t.Context(), p.usuarioID, eleicaoID, p.voto), p.esperado)
		})
	}

	t.Run("cédulas anônimas e participações gravadas em separado", func(t *testing.T) {
		for _, tipo := range []string{"chapa", "branco", "nulo"} {
			if n := c.contar(t, "SELECT count(*) FROM cedulas WHERE votacao_id = $1 AND tipo = $2", eleicaoID, tipo); n != 1 {
				t.Errorf("cédulas %s: esperado 1, obtido %d (o voto repetido não pode deixar cédula)", tipo, n)
			}
		}
		if n := c.contar(t, "SELECT count(*) FROM participacoes WHERE votacao_id = $1", eleicaoID); n != 3 {
			t.Errorf("participações: esperado 3, obtido %d", n)
		}
	})
}

func TestVotarConsulta(t *testing.T) {
	c := novoCenario(t)
	anaID := c.usuario(t, "ana", PapelEleitor)
	brunoID := c.usuario(t, "bruno", PapelEleitor)
	carlaID := c.usuario(t, "carla", PapelEleitor)
	consultaID := c.votacao(t, NovaVotacao{Tipo: TipoConsulta, Opcoes: projetos(10, 20, 30), OrcamentoCentavos: ptr[int64](100), Aptos: []int64{anaID, brunoID}})
	ids := c.opcoes(t, consultaID)
	a, b, cc := ids[0], ids[1], ids[2]

	passos := []struct {
		nome      string
		usuarioID int64
		voto      Voto
		esperado  error
	}{
		{"cédula vazia", anaID, Voto{}, ErrCedulaInvalida},
		{"cédula incompleta", anaID, Voto{Preferencias: []int64{a, b}}, ErrCedulaInvalida},
		{"cédula com projeto repetido", anaID, Voto{Preferencias: []int64{a, a, b}}, ErrCedulaInvalida},
		{"cédula com projeto desconhecido", anaID, Voto{Preferencias: []int64{a, b, 9999}}, ErrCedulaInvalida},
		{"cédula com projeto a mais", anaID, Voto{Preferencias: []int64{a, b, cc, a}}, ErrCedulaInvalida},
		{"consulta não aceita escolha", anaID, Voto{Escolha: "chapa", Preferencias: []int64{a, b, cc}}, errValidacao},
		{"eleitor fora da lista de aptos", carlaID, Voto{Preferencias: []int64{a, b, cc}}, ErrNaoApto},
		{"cédula completa", anaID, Voto{Preferencias: []int64{cc, a, b}}, nil},
		{"segundo voto do mesmo eleitor", anaID, Voto{Preferencias: []int64{a, b, cc}}, ErrVotoDuplicado},
	}
	for _, p := range passos {
		t.Run(p.nome, func(t *testing.T) {
			conferirErro(t, c.svc.Votar(t.Context(), p.usuarioID, consultaID, p.voto), p.esperado)
		})
	}

	t.Run("voto repetido não deixa cédula nem preferências extras", func(t *testing.T) {
		if n := c.contar(t, "SELECT count(*) FROM cedulas WHERE votacao_id = $1", consultaID); n != 1 {
			t.Errorf("cédulas: esperado 1, obtido %d", n)
		}
		if n := c.contar(t, "SELECT count(*) FROM cedula_preferencias"); n != 3 {
			t.Errorf("preferências: esperado 3, obtido %d", n)
		}
	})

	t.Run("preferências gravadas na ordem enviada", func(t *testing.T) {
		rows, err := c.db.Query("SELECT opcao_id FROM cedula_preferencias ORDER BY posicao")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = rows.Close() }()
		var ordem []int64
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				t.Fatal(err)
			}
			ordem = append(ordem, id)
		}
		if !slices.Equal(ordem, []int64{cc, a, b}) {
			t.Errorf("ordem: esperado %v, obtido %v", []int64{cc, a, b}, ordem)
		}
	})
}
