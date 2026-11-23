package server

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

var camposResumo = []string{"id", "tipo", "titulo", "descricao", "inicio", "fim", "turno", "status", "apto", "ja_votou",
	"votacao_anterior_id", "segundo_turno_id"}

var camposDetalhe = append(append([]string{}, camposResumo...), "opcoes", "orcamento_centavos", "minha_unidade")

func TestListarEDetalharVotacoes(t *testing.T) {
	a := novoAmbiente(t)

	rec := a.requisitar(http.MethodGet, "/api/votacoes", a.token("maria"), nil)
	lista := decodificar[[]votacaoResumoDTO](t, rec, http.StatusOK)
	titulos := []string{tituloConsulta, tituloEleicao, tituloPautaAberta, tituloPautaEncerrada}
	if len(lista) != len(titulos) {
		t.Fatalf("esperava %d votações, obteve %+v", len(titulos), lista)
	}
	for i, titulo := range titulos {
		if lista[i].Titulo != titulo {
			t.Errorf("posição %d: esperado %q (mais recentes primeiro), obtido %q", i, titulo, lista[i].Titulo)
		}
	}
	pautaAberta := lista[2]
	if pautaAberta.Status != "aberta" || !pautaAberta.Apto || pautaAberta.JaVotou || pautaAberta.Turno != 1 ||
		pautaAberta.Inicio != "2026-11-18T11:00:00Z" || pautaAberta.Fim != "2026-11-25T12:00:00Z" {
		t.Errorf("resumo inesperado: %+v", pautaAberta)
	}

	t.Run("detalhe da pauta mostra a unidade e o peso antes do voto", func(t *testing.T) {
		rec := a.requisitar(http.MethodGet, rota("/api/votacoes/%d", pautaAberta.ID), a.token("maria"), nil)
		d := decodificar[votacaoDetalheDTO](t, rec, http.StatusOK)
		conferirChaves(t, rec.Body.Bytes(), camposDetalhe...)
		if !strings.Contains(rec.Body.String(), `"opcoes":[]`) || d.MinhaUnidade == nil ||
			*d.MinhaUnidade != (minhaUnidadeDTO{Identificador: "Apto 101", FracaoIdeal: 60, Percentual: 12, Adimplente: true, Elegivel: true}) {
			t.Errorf("detalhe inesperado: %s", rec.Body.String())
		}
	})

	t.Run("inadimplente vê o motivo do bloqueio", func(t *testing.T) {
		d := decodificar[votacaoDetalheDTO](t, a.requisitar(http.MethodGet, rota("/api/votacoes/%d", pautaAberta.ID), a.token("sofia"), nil), http.StatusOK)
		if d.MinhaUnidade == nil || d.MinhaUnidade.Elegivel || d.MinhaUnidade.Motivo == "" || d.MinhaUnidade.Percentual != 24 {
			t.Errorf("minha_unidade inesperada: %+v", d.MinhaUnidade)
		}
	})

	t.Run("detalhe da consulta traz projetos com custo e orçamento", func(t *testing.T) {
		d := decodificar[votacaoDetalheDTO](t, a.requisitar(http.MethodGet, rota("/api/votacoes/%d", lista[0].ID), a.token("hugo"), nil), http.StatusOK)
		if len(d.Opcoes) != 4 || d.Opcoes[0].CustoCentavos == nil || *d.Opcoes[0].CustoCentavos != 35000_00 ||
			d.OrcamentoCentavos == nil || *d.OrcamentoCentavos != 60000_00 || d.MinhaUnidade != nil || !d.Apto {
			t.Errorf("detalhe inesperado: %+v", d)
		}
	})

	t.Run("votação inexistente ou id inválido", func(t *testing.T) {
		for _, caminho := range []string{"/api/votacoes/9999", "/api/votacoes/abc", "/api/votacoes/9999/resultado"} {
			conferirErro(t, a.requisitar(http.MethodGet, caminho, a.token("maria"), nil), http.StatusNotFound, "nao_encontrado")
		}
	})
}

func TestVotacaoDaPautaDeCondominio(t *testing.T) {
	a := novoAmbiente(t)
	pauta := a.votacaoPorTitulo(tituloPautaAberta)
	votos := rota("/api/votacoes/%d/votos", pauta)

	t.Run("resultado só depois do encerramento", func(t *testing.T) {
		conferirErro(t, a.requisitar(http.MethodGet, rota("/api/votacoes/%d/resultado", pauta), "", nil), http.StatusConflict, "votacao_nao_encerrada")
	})

	for _, login := range []string{"maria", "joao", "lucia"} {
		rec := a.requisitar(http.MethodPost, votos, a.token(login), `{"escolha":"favor"}`)
		resposta := decodificar[respostaVoto](t, rec, http.StatusCreated)
		if resposta.Mensagem != "Voto registrado com sucesso." || strings.Contains(rec.Body.String(), "favor") {
			t.Errorf("a resposta do voto não pode ecoar a escolha: %s", rec.Body.String())
		}
	}

	erros := []struct {
		nome   string
		login  string
		corpo  string
		status int
		codigo string
	}{
		{"voto repetido", "maria", `{"escolha":"contra"}`, http.StatusConflict, "voto_duplicado"},
		{"inadimplente é bloqueada", "sofia", `{"escolha":"favor"}`, http.StatusForbidden, "inadimplente"},
		{"estudante sem unidade não é apto", "ana", `{"escolha":"favor"}`, http.StatusForbidden, "nao_apto"},
		{"escolha inválida", "maria", `{"escolha":"talvez"}`, http.StatusBadRequest, "requisicao_invalida"},
		{"campo desconhecido", "maria", `{"escolha":"favor","peso":1000}`, http.StatusBadRequest, "requisicao_invalida"},
	}
	for _, e := range erros {
		t.Run(e.nome, func(t *testing.T) {
			conferirErro(t, a.requisitar(http.MethodPost, votos, a.token(e.login), e.corpo), e.status, e.codigo)
		})
	}

	t.Run("encerrar e ver aprovação com 380 de 500", func(t *testing.T) {
		rec := a.requisitar(http.MethodPost, rota("/api/votacoes/%d/encerrar", pauta), a.token("admin"), nil)
		resumo := decodificar[votacaoResumoDTO](t, rec, http.StatusOK)
		conferirChaves(t, rec.Body.Bytes(), camposResumo...)
		if resumo.Status != "encerrada" || resumo.Fim != "2026-11-18T12:00:00Z" {
			t.Errorf("resumo inesperado: %+v", resumo)
		}

		rec = a.requisitar(http.MethodGet, rota("/api/votacoes/%d/resultado", pauta), "", nil)
		r := decodificar[resultadoPautaDTO](t, rec, http.StatusOK)
		conferirChaves(t, rec.Body.Bytes(), "tipo", "votacao_id", "titulo", "descricao", "contagem", "peso_total", "limiar",
			"percentual_favor", "aprovada", "unidades_total", "unidades_votantes")
		esperado := resultadoPautaDTO{Tipo: "pauta", VotacaoID: pauta, Titulo: tituloPautaAberta,
			Descricao: "Reforma da portaria com rateio do custo pela fração ideal de cada unidade.", Contagem: contagemPautaDTO{Favor: 380}, PesoTotal: 500,
			Limiar: r.Limiar, PercentualFavor: 76, Aprovada: true, UnidadesTotal: 6, UnidadesVotantes: 5}
		if r != esperado || r.Limiar < 333.33 || r.Limiar > 333.34 {
			t.Errorf("resultado:\n esperado %+v\n obtido   %+v", esperado, r)
		}
	})

	t.Run("encerrar de novo", func(t *testing.T) {
		conferirErro(t, a.requisitar(http.MethodPost, rota("/api/votacoes/%d/encerrar", pauta), a.token("admin"), nil), http.StatusConflict, "votacao_nao_aberta")
	})

	t.Run("votar depois do encerramento", func(t *testing.T) {
		conferirErro(t, a.requisitar(http.MethodPost, votos, a.token("pedro"), `{"escolha":"favor"}`), http.StatusConflict, "votacao_nao_aberta")
	})
}

func TestEleicaoComSegundoTurno(t *testing.T) {
	a := novoAmbiente(t)
	admin := a.token("admin")
	eleicao := a.votacaoPorTitulo(tituloEleicao)

	rec := a.requisitar(http.MethodGet, rota("/api/votacoes/%d/resultado", eleicao), "", nil)
	boletim := decodificar[resultadoEleicaoDTO](t, rec, http.StatusOK)
	conferirChaves(t, rec.Body.Bytes(), "tipo", "votacao_id", "titulo", "descricao", "turno", "aptos", "participantes", "abstencoes",
		"percentual_comparecimento", "votos", "brancos", "nulos", "validos", "vencedora", "segundo_turno")
	if boletim.Tipo != "eleicao" || boletim.Titulo != tituloEleicao || boletim.Descricao == "" || boletim.Aptos != 8 || boletim.Participantes != 7 || boletim.Abstencoes != 1 ||
		boletim.PercentualComparecimento != 87.5 || boletim.Brancos != 1 || boletim.Validos != 6 || boletim.Vencedora != nil ||
		len(boletim.Votos) != 3 || boletim.Votos[0].Votos != 3 || boletim.Votos[0].PercentualValidos != 50 {
		t.Errorf("boletim inesperado: %+v", boletim)
	}
	if st := boletim.SegundoTurno; !st.Necessario || st.EmpateIndefinido || len(st.Chapas) != 2 || st.VotacaoID != nil ||
		st.Chapas[0].Nome != "Chapa 1 — Movimenta" || st.Chapas[1].Nome != "Chapa 2 — Renovação" {
		t.Errorf("segundo turno inesperado: %+v", st)
	}
	if strings.Contains(rec.Body.String(), "ana") || strings.Contains(rec.Body.String(), "T12:") {
		t.Errorf("o boletim não pode ter eleitores nem horários: %s", rec.Body.String())
	}

	rec = a.requisitar(http.MethodPost, rota("/api/votacoes/%d/segundo-turno", eleicao), admin, `{}`)
	turno2 := decodificar[votacaoDetalheDTO](t, rec, http.StatusCreated)
	if turno2.Turno != 2 || turno2.Status != "aberta" || turno2.VotacaoAnteriorID == nil || *turno2.VotacaoAnteriorID != eleicao ||
		len(turno2.Opcoes) != 2 || turno2.Opcoes[0].Nome != "Chapa 1 — Movimenta" || turno2.Fim != "2026-11-19T12:00:00Z" {
		t.Fatalf("2º turno inesperado: %+v", turno2)
	}

	t.Run("2º turno só abre uma vez", func(t *testing.T) {
		conferirErro(t, a.requisitar(http.MethodPost, rota("/api/votacoes/%d/segundo-turno", eleicao), admin, nil),
			http.StatusConflict, "segundo_turno_indisponivel")
	})

	t.Run("pauta não tem 2º turno", func(t *testing.T) {
		conferirErro(t, a.requisitar(http.MethodPost, rota("/api/votacoes/%d/segundo-turno", a.votacaoPorTitulo(tituloPautaEncerrada)), admin, nil),
			http.StatusConflict, "segundo_turno_indisponivel")
	})

	a.relogio.Avancar(time.Minute)
	votos := rota("/api/votacoes/%d/votos", turno2.ID)
	chapa2 := turno2.Opcoes[1].ID
	for _, login := range []string{"ana", "bruno"} {
		decodificar[respostaVoto](t, a.requisitar(http.MethodPost, votos, a.token(login), fmt.Sprintf(`{"escolha":"chapa","opcao_id":%d}`, chapa2)), http.StatusCreated)
	}

	erros := []struct {
		nome   string
		login  string
		corpo  string
		status int
		codigo string
	}{
		{"chapa do 1º turno não vale no 2º", "carla", fmt.Sprintf(`{"escolha":"chapa","opcao_id":%d}`, boletim.Votos[2].OpcaoID), http.StatusBadRequest, "opcao_desconhecida"},
		{"condômino não é apto na eleição", "maria", `{"escolha":"branco"}`, http.StatusForbidden, "nao_apto"},
		{"voto repetido", "ana", `{"escolha":"nulo"}`, http.StatusConflict, "voto_duplicado"},
	}
	for _, e := range erros {
		t.Run(e.nome, func(t *testing.T) {
			conferirErro(t, a.requisitar(http.MethodPost, votos, a.token(e.login), e.corpo), e.status, e.codigo)
		})
	}

	decodificar[votacaoResumoDTO](t, a.requisitar(http.MethodPost, rota("/api/votacoes/%d/encerrar", turno2.ID), admin, nil), http.StatusOK)

	t.Run("boletim do 2º turno mostra a vencedora", func(t *testing.T) {
		r := decodificar[resultadoEleicaoDTO](t, a.requisitar(http.MethodGet, rota("/api/votacoes/%d/resultado", turno2.ID), "", nil), http.StatusOK)
		if r.Turno != 2 || r.Titulo != tituloEleicao+" — 2º turno" || r.Vencedora == nil || r.Vencedora.Nome != "Chapa 2 — Renovação" || r.Vencedora.OpcaoID != chapa2 ||
			r.Participantes != 2 || r.Abstencoes != 6 || r.SegundoTurno.Necessario {
			t.Errorf("boletim inesperado: %+v", r)
		}
	})

	t.Run("1º turno aponta para o 2º turno", func(t *testing.T) {
		r := decodificar[resultadoEleicaoDTO](t, a.requisitar(http.MethodGet, rota("/api/votacoes/%d/resultado", eleicao), "", nil), http.StatusOK)
		if r.SegundoTurno.VotacaoID == nil || *r.SegundoTurno.VotacaoID != turno2.ID {
			t.Errorf("segundo_turno.votacao_id inesperado: %+v", r.SegundoTurno)
		}
		lista := decodificar[[]votacaoResumoDTO](t, a.requisitar(http.MethodGet, "/api/votacoes", a.token("ana"), nil), http.StatusOK)
		if lista[0].ID != turno2.ID || !lista[0].JaVotou {
			t.Errorf("o 2º turno deveria ser o primeiro da lista, já votado pela ana: %+v", lista[0])
		}
		for _, v := range lista {
			if v.ID == eleicao && (v.SegundoTurnoID == nil || *v.SegundoTurnoID != turno2.ID) {
				t.Errorf("segundo_turno_id do 1º turno inesperado: %+v", v)
			}
		}
	})
}

func TestConsultaComCorteOrcamentario(t *testing.T) {
	a := novoAmbiente(t)
	consulta := a.votacaoPorTitulo(tituloConsulta)
	votos := rota("/api/votacoes/%d/votos", consulta)
	projetos := decodificar[votacaoDetalheDTO](t, a.requisitar(http.MethodGet, rota("/api/votacoes/%d", consulta), a.token("hugo"), nil), http.StatusOK).Opcoes
	impressoras, arduino, gpu, mobiliario := projetos[0].ID, projetos[1].ID, projetos[2].ID, projetos[3].ID

	erros := []struct {
		nome   string
		login  string
		corpo  string
		status int
		codigo string
	}{
		{"cédula incompleta", "hugo", fmt.Sprintf(`{"preferencias":[%d,%d]}`, impressoras, arduino), http.StatusBadRequest, "cedula_invalida"},
		{"cédula com repetição", "hugo", fmt.Sprintf(`{"preferencias":[%d,%d,%d,%d]}`, gpu, gpu, arduino, mobiliario), http.StatusBadRequest, "cedula_invalida"},
		{"preferências com tipo errado", "hugo", `{"preferencias":["gpu"]}`, http.StatusBadRequest, "requisicao_invalida"},
		{"voto repetido do seed", "ana", fmt.Sprintf(`{"preferencias":[%d,%d,%d,%d]}`, gpu, arduino, impressoras, mobiliario), http.StatusConflict, "voto_duplicado"},
	}
	for _, e := range erros {
		t.Run(e.nome, func(t *testing.T) {
			conferirErro(t, a.requisitar(http.MethodPost, votos, a.token(e.login), e.corpo), e.status, e.codigo)
		})
	}

	corpo := fmt.Sprintf(`{"preferencias":[%d,%d,%d,%d]}`, mobiliario, arduino, impressoras, gpu)
	decodificar[respostaVoto](t, a.requisitar(http.MethodPost, votos, a.token("hugo"), corpo), http.StatusCreated)
	decodificar[votacaoResumoDTO](t, a.requisitar(http.MethodPost, rota("/api/votacoes/%d/encerrar", consulta), a.token("admin"), nil), http.StatusOK)

	rec := a.requisitar(http.MethodGet, rota("/api/votacoes/%d/resultado", consulta), "", nil)
	r := decodificar[resultadoConsultaDTO](t, rec, http.StatusOK)
	conferirChaves(t, rec.Body.Bytes(), "tipo", "votacao_id", "titulo", "descricao", "cedulas", "orcamento_centavos",
		"total_selecionado_centavos", "saldo_centavos", "empate_na_linha_de_corte", "ranking")
	// Com a cédula do hugo: Impressoras 18, Arduino 17, GPU 13, Mobiliário 12.
	esperado := []itemRankingDTO{
		{Posicao: 1, OpcaoID: impressoras, Nome: "Impressoras 3D e CNC", Pontos: 18, PrimeirasPreferencias: 2, CustoCentavos: 35000_00, CustoAcumuladoCentavos: 35000_00, DentroDoOrcamento: true},
		{Posicao: 2, OpcaoID: arduino, Nome: "Kit Arduino e Robótica", Pontos: 17, PrimeirasPreferencias: 1, CustoCentavos: 12000_00, CustoAcumuladoCentavos: 47000_00, DentroDoOrcamento: true},
		{Posicao: 3, OpcaoID: gpu, Nome: "Servidor GPU para IA", Pontos: 13, PrimeirasPreferencias: 2, CustoCentavos: 48000_00, CustoAcumuladoCentavos: 95000_00},
		{Posicao: 4, OpcaoID: mobiliario, Nome: "Mobiliário colaborativo", Pontos: 12, PrimeirasPreferencias: 1, CustoCentavos: 9000_00, CustoAcumuladoCentavos: 104000_00},
	}
	if r.Tipo != "consulta" || r.Titulo != tituloConsulta || r.Descricao == "" || r.Cedulas != 6 || r.OrcamentoCentavos != 60000_00 || r.TotalSelecionadoCentavos != 47000_00 ||
		r.SaldoCentavos != 13000_00 || r.EmpateNaLinhaDeCorte || len(r.Ranking) != len(esperado) {
		t.Fatalf("resultado inesperado: %+v", r)
	}
	for i := range esperado {
		if r.Ranking[i] != esperado[i] {
			t.Errorf("linha %d:\n esperado %+v\n obtido   %+v", i, esperado[i], r.Ranking[i])
		}
	}
}

func TestCriarVotacao(t *testing.T) {
	a := novoAmbiente(t)
	admin := a.token("admin")

	t.Run("eleição criada responde 201 com o detalhe", func(t *testing.T) {
		corpo := `{"tipo":"eleicao","titulo":"Representante de turma","descricao":"",
			"inicio":"2026-11-19T12:00:00.000Z","fim":"2026-11-20T12:00:00Z",
			"opcoes":[{"nome":"Chapa A","descricao":"a"},{"nome":"Chapa B","descricao":"b"}],"aptos":[]}`
		rec := a.requisitar(http.MethodPost, "/api/votacoes", admin, corpo)
		d := decodificar[votacaoDetalheDTO](t, rec, http.StatusCreated)
		conferirChaves(t, rec.Body.Bytes(), camposDetalhe...)
		if d.Tipo != "eleicao" || d.Status != "nao_iniciada" || d.Turno != 1 || len(d.Opcoes) != 2 ||
			d.Opcoes[1].Nome != "Chapa B" || d.Opcoes[1].CustoCentavos != nil || d.Inicio != "2026-11-19T12:00:00Z" {
			t.Errorf("detalhe inesperado: %+v", d)
		}
	})

	erros := []struct {
		nome  string
		corpo string
	}{
		{"título vazio", `{"tipo":"pauta","titulo":"","inicio":"2026-11-19T12:00:00Z","fim":"2026-11-20T12:00:00Z","opcoes":[]}`},
		{"fim antes do início", `{"tipo":"pauta","titulo":"X","inicio":"2026-11-20T12:00:00Z","fim":"2026-11-19T12:00:00Z","opcoes":[]}`},
		{"data fora do RFC 3339", `{"tipo":"pauta","titulo":"X","inicio":"19/11/2026","fim":"2026-11-20T12:00:00Z","opcoes":[]}`},
		{"consulta sem orçamento", `{"tipo":"consulta","titulo":"X","inicio":"2026-11-19T12:00:00Z","fim":"2026-11-20T12:00:00Z",
			"opcoes":[{"nome":"A","descricao":"","custo_centavos":1},{"nome":"B","descricao":"","custo_centavos":2}]}`},
		{"campo desconhecido", `{"tipo":"pauta","titulo":"X","inicio":"2026-11-19T12:00:00Z","fim":"2026-11-20T12:00:00Z","turno":2}`},
		{"corpo acima de 64 KiB", `{"tipo":"pauta","titulo":"` + strings.Repeat("a", 70_000) + `"}`},
		{"custo que estouraria a soma do ranking", `{"tipo":"consulta","titulo":"X","inicio":"2026-11-19T12:00:00Z","fim":"2026-11-20T12:00:00Z",
			"opcoes":[{"nome":"A","descricao":"","custo_centavos":5000000000000000000},{"nome":"B","descricao":"","custo_centavos":5000000000000000000}],
			"orcamento_centavos":1}`},
		{"título com caractere NUL", `{"tipo":"pauta","titulo":"a\u0000b","inicio":"2026-11-19T12:00:00Z","fim":"2026-11-20T12:00:00Z","opcoes":[]}`},
		{"janela menor que um microssegundo", `{"tipo":"pauta","titulo":"X","inicio":"2026-11-19T12:00:00.0000001Z","fim":"2026-11-19T12:00:00.0000009Z","opcoes":[]}`},
	}
	for _, e := range erros {
		t.Run(e.nome, func(t *testing.T) {
			conferirErro(t, a.requisitar(http.MethodPost, "/api/votacoes", admin, e.corpo), http.StatusBadRequest, "requisicao_invalida")
		})
	}
}
