package service

import (
	"slices"
	"strings"
	"testing"
	"time"
)

func TestCriarVotacaoValidacao(t *testing.T) {
	c := novoCenario(t)
	adminID := c.usuario(t, "admin", PapelAdmin)
	anaID := c.usuario(t, "ana", PapelEleitor)

	inicio, fim := instanteInicial, instanteInicial.Add(time.Hour)
	base := func(tipo string, opcoes []NovaOpcao) NovaVotacao {
		return NovaVotacao{Tipo: tipo, Titulo: "Título", Inicio: inicio, Fim: fim, Opcoes: opcoes}
	}
	comOrcamento := func(n NovaVotacao, orcamento int64) NovaVotacao {
		n.OrcamentoCentavos = &orcamento
		return n
	}
	comAptos := func(n NovaVotacao, aptos ...int64) NovaVotacao {
		n.Aptos = aptos
		return n
	}
	alterar := func(n NovaVotacao, mudar func(*NovaVotacao)) NovaVotacao {
		mudar(&n)
		return n
	}
	chapas := opcoesNomeadas("Chapa 1", "Chapa 2")

	testes := []struct {
		nome     string
		nova     NovaVotacao
		esperado error
	}{
		{"pauta válida", base(TipoPauta, nil), nil},
		{"eleição válida", base(TipoEleicao, chapas), nil},
		{"consulta válida com orçamento zero e custo zero", comOrcamento(base(TipoConsulta, projetos(0, 10)), 0), nil},
		{"eleição com 20 opções", base(TipoEleicao, opcoesNomeadas(strings.Split("a b c d e f g h i j k l m n o p q r s t", " ")...)), nil},
		{"título com 200 caracteres acentuados", alterar(base(TipoPauta, nil), func(n *NovaVotacao) { n.Titulo = strings.Repeat("ç", 200) }), nil},
		{"título só com espaços", alterar(base(TipoPauta, nil), func(n *NovaVotacao) { n.Titulo = "   " }), errValidacao},
		{"tipo desconhecido", base("referendo", nil), errValidacao},
		{"título com 201 caracteres", alterar(base(TipoPauta, nil), func(n *NovaVotacao) { n.Titulo = strings.Repeat("a", 201) }), errValidacao},
		{"descrição com 2001 caracteres", alterar(base(TipoPauta, nil), func(n *NovaVotacao) { n.Descricao = strings.Repeat("a", 2001) }), errValidacao},
		{"sem início", alterar(base(TipoPauta, nil), func(n *NovaVotacao) { n.Inicio = time.Time{} }), errValidacao},
		{"fim igual ao início", alterar(base(TipoPauta, nil), func(n *NovaVotacao) { n.Fim = n.Inicio }), errValidacao},
		{"fim antes do início", alterar(base(TipoPauta, nil), func(n *NovaVotacao) { n.Fim = n.Inicio.Add(-time.Second) }), errValidacao},
		{"pauta com opções", base(TipoPauta, chapas), errValidacao},
		{"pauta com orçamento", comOrcamento(base(TipoPauta, nil), 10), errValidacao},
		{"pauta com aptos", comAptos(base(TipoPauta, nil), anaID), errValidacao},
		{"eleição com uma chapa", base(TipoEleicao, opcoesNomeadas("Única")), errValidacao},
		{"eleição com 21 opções", base(TipoEleicao, opcoesNomeadas(strings.Split("a b c d e f g h i j k l m n o p q r s t u", " ")...)), errValidacao},
		{"opção sem nome", base(TipoEleicao, opcoesNomeadas("Chapa 1", " ")), errValidacao},
		{"opção com nome de 121 caracteres", base(TipoEleicao, opcoesNomeadas("Chapa 1", strings.Repeat("a", 121))), errValidacao},
		{"opção com descrição de 2001 caracteres", base(TipoEleicao, []NovaOpcao{{Nome: "A"}, {Nome: "B", Descricao: strings.Repeat("a", 2001)}}), errValidacao},
		{"nomes repetidos depois de tirar espaços", base(TipoEleicao, opcoesNomeadas("Chapa 1", " Chapa 1 ")), errValidacao},
		{"chapa com custo", base(TipoEleicao, projetos(10, 20)), errValidacao},
		{"eleição com orçamento", comOrcamento(base(TipoEleicao, chapas), 10), errValidacao},
		{"consulta sem orçamento", base(TipoConsulta, projetos(10, 20)), errValidacao},
		{"consulta com orçamento negativo", comOrcamento(base(TipoConsulta, projetos(10, 20)), -1), errValidacao},
		{"projeto sem custo", comOrcamento(base(TipoConsulta, opcoesNomeadas("A", "B")), 10), errValidacao},
		{"projeto com custo negativo", comOrcamento(base(TipoConsulta, projetos(10, -1)), 10), errValidacao},
		{"aptos com administrador", comAptos(base(TipoEleicao, chapas), anaID, adminID), errValidacao},
		{"aptos com usuário inexistente", comAptos(base(TipoEleicao, chapas), anaID+100), errValidacao},
		{"projetos e orçamento no limite de R$ 1 trilhão", comOrcamento(base(TipoConsulta, projetos(maxCentavos, maxCentavos)), maxCentavos), nil},
		{"projeto com custo acima de R$ 1 trilhão", comOrcamento(base(TipoConsulta, projetos(10, maxCentavos+1)), 10), errValidacao},
		{"orçamento acima de R$ 1 trilhão", comOrcamento(base(TipoConsulta, projetos(10, 20)), maxCentavos+1), errValidacao},
		{"título com caractere NUL", alterar(base(TipoPauta, nil), func(n *NovaVotacao) { n.Titulo = "a\x00b" }), errValidacao},
		{"descrição com caractere NUL", alterar(base(TipoPauta, nil), func(n *NovaVotacao) { n.Descricao = "\x00" }), errValidacao},
		{"opção com caractere NUL no nome", base(TipoEleicao, opcoesNomeadas("Chapa 1", "Chapa\x002")), errValidacao},
		{"janela de um microssegundo", alterar(base(TipoPauta, nil), func(n *NovaVotacao) { n.Fim = n.Inicio.Add(time.Microsecond) }), nil},
		{"janela menor que um microssegundo", alterar(base(TipoPauta, nil), func(n *NovaVotacao) {
			n.Inicio, n.Fim = instanteInicial.Add(100*time.Nanosecond), instanteInicial.Add(900*time.Nanosecond)
		}), errValidacao},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			_, err := c.svc.CriarVotacao(t.Context(), tc.nova)
			conferirErro(t, err, tc.esperado)
		})
	}
}

func TestCriarVotacaoGravaOpcoesEAptos(t *testing.T) {
	c := novoCenario(t)
	c.usuario(t, "admin", PapelAdmin)
	anaID := c.usuario(t, "ana", PapelEleitor)
	brunoID := c.usuario(t, "bruno", PapelEleitor)
	carlaID := c.usuario(t, "carla", PapelEleitor)

	t.Run("consulta grava projetos na ordem, com custo, e orçamento", func(t *testing.T) {
		id := c.votacao(t, NovaVotacao{
			Tipo: TipoConsulta, Titulo: "  Orçamento  ", Descricao: " Desc ",
			Opcoes:            []NovaOpcao{{Nome: " B ", Descricao: " d ", CustoCentavos: ptr[int64](200)}, {Nome: "A", CustoCentavos: ptr[int64](100)}},
			OrcamentoCentavos: ptr[int64](300),
		})
		detalhe, err := c.svc.Detalhar(t.Context(), anaID, id)
		if err != nil {
			t.Fatal(err)
		}
		if detalhe.Titulo != "Orçamento" || detalhe.Descricao != "Desc" || detalhe.Turno != 1 ||
			detalhe.OrcamentoCentavos == nil || *detalhe.OrcamentoCentavos != 300 {
			t.Errorf("votação inesperada: %+v", detalhe)
		}
		if len(detalhe.Opcoes) != 2 || detalhe.Opcoes[0].Nome != "B" || detalhe.Opcoes[0].Descricao != "d" ||
			*detalhe.Opcoes[0].CustoCentavos != 200 || detalhe.Opcoes[1].Nome != "A" {
			t.Errorf("opções inesperadas: %+v", detalhe.Opcoes)
		}
	})

	aptosDe := func(t *testing.T, votacaoID int64) []int64 {
		t.Helper()
		rows, err := c.db.Query("SELECT usuario_id FROM aptos WHERE votacao_id = $1 ORDER BY usuario_id", votacaoID)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = rows.Close() }()
		ids := []int64{}
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				t.Fatal(err)
			}
			ids = append(ids, id)
		}
		return ids
	}

	testes := []struct {
		nome     string
		aptos    []int64
		esperado []int64
	}{
		{"aptos omitidos habilitam todos os eleitores", nil, []int64{anaID, brunoID, carlaID}},
		{"aptos informados, com repetição, são gravados uma vez", []int64{carlaID, anaID, carlaID}, []int64{anaID, carlaID}},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			id := c.votacao(t, NovaVotacao{Tipo: TipoEleicao, Opcoes: opcoesNomeadas("C1", "C2"), Aptos: tc.aptos})
			if obtido := aptosDe(t, id); !slices.Equal(obtido, tc.esperado) {
				t.Errorf("aptos: esperado %v, obtido %v", tc.esperado, obtido)
			}
		})
	}

	t.Run("pauta não grava opções nem aptos", func(t *testing.T) {
		id := c.votacao(t, NovaVotacao{Tipo: TipoPauta})
		if n := c.contar(t, "SELECT count(*) FROM opcoes WHERE votacao_id = $1", id); n != 0 {
			t.Errorf("opções: esperado 0, obtido %d", n)
		}
		if aptos := aptosDe(t, id); len(aptos) != 0 {
			t.Errorf("aptos: esperado nenhum, obtido %v", aptos)
		}
	})
}
