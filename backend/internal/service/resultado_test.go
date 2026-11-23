package service

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"time"
)

func TestResultadoExigeVotacaoEncerrada(t *testing.T) {
	c := novoCenario(t)
	agora := c.relogio.Agora()
	aberta := c.votacao(t, NovaVotacao{Tipo: TipoPauta})
	naoIniciada := c.votacao(t, NovaVotacao{Tipo: TipoPauta, Inicio: agora.Add(time.Hour), Fim: agora.Add(2 * time.Hour)})

	testes := []struct {
		nome      string
		votacaoID int64
		esperado  error
	}{
		{"votação aberta", aberta, ErrVotacaoNaoEncerrada},
		{"votação não iniciada", naoIniciada, ErrVotacaoNaoEncerrada},
		{"votação inexistente", 9999, ErrNaoEncontrado},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			_, err := c.svc.Resultado(t.Context(), tc.votacaoID)
			conferirErro(t, err, tc.esperado)
		})
	}
}

func TestResultadoPauta(t *testing.T) {
	unidades := []struct {
		login, identificador string
		fracao               float64
		adimplente           bool
	}{
		{"maria", "Apto 101", 60, true},
		{"joao", "Apto 102", 60, true},
		{"lucia", "Apto 201", 80, true},
		{"pedro", "Apto 202", 80, true},
		{"rafael", "Apto 301", 100, true},
		{"sofia", "Apto 402", 120, false},
	}

	testes := []struct {
		nome     string
		votos    map[string]string // login -> escolha
		esperado ResultadoPauta
	}{
		{
			"maioria simples não basta: 300 de 500 é reprovada",
			map[string]string{"maria": "favor", "joao": "favor", "lucia": "favor", "rafael": "favor", "pedro": "contra"},
			ResultadoPauta{Favor: 300, Contra: 80, PesoTotal: 500, Limiar: 500.0 * 2 / 3, PercentualFavor: 60, UnidadesTotal: 6, UnidadesVotantes: 5},
		},
		{
			"380 de 500 atinge 2/3 e é aprovada",
			map[string]string{"maria": "favor", "joao": "favor", "lucia": "favor", "rafael": "favor", "pedro": "favor"},
			ResultadoPauta{Favor: 380, PesoTotal: 500, Limiar: 500.0 * 2 / 3, PercentualFavor: 76, Aprovada: true, UnidadesTotal: 6, UnidadesVotantes: 5},
		},
		{
			"sem votos é reprovada",
			nil,
			ResultadoPauta{PesoTotal: 500, Limiar: 500.0 * 2 / 3, UnidadesTotal: 6},
		},
		{
			"abstenção conta peso, mas não a favor",
			map[string]string{"rafael": "abstencao", "maria": "contra"},
			ResultadoPauta{Contra: 60, Abstencao: 100, PesoTotal: 500, Limiar: 500.0 * 2 / 3, UnidadesTotal: 6, UnidadesVotantes: 2},
		},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			c := novoCenario(t)
			ids := map[string]int64{}
			for _, u := range unidades {
				ids[u.login] = c.usuario(t, u.login, PapelEleitor)
				c.unidade(t, u.identificador, u.fracao, u.adimplente, ids[u.login])
			}
			// Unidade sem responsável não vota e não entra no peso total.
			if _, err := c.svc.CriarUnidade(t.Context(), "Loja 1", 1000, true, nil); err != nil {
				t.Fatal(err)
			}

			pautaID := c.votacao(t, NovaVotacao{Tipo: TipoPauta, Titulo: "Pauta de teste", Descricao: "Descrição da pauta"})
			for login, escolha := range tc.votos {
				c.votar(t, ids[login], pautaID, Voto{Escolha: escolha})
			}
			c.encerrar(t, pautaID)

			resultado, err := c.svc.Resultado(t.Context(), pautaID)
			if err != nil {
				t.Fatal(err)
			}
			if resultado.Pauta == nil || resultado.Eleicao != nil || resultado.Consulta != nil {
				t.Fatalf("esperava só o resultado de pauta: %+v", resultado)
			}
			tc.esperado.VotacaoID = pautaID
			tc.esperado.Titulo, tc.esperado.Descricao = "Pauta de teste", "Descrição da pauta"
			if *resultado.Pauta != tc.esperado {
				t.Errorf("esperado %+v,\n obtido %+v", tc.esperado, *resultado.Pauta)
			}
		})
	}
}

// eleicaoEncerrada cria uma eleição com as chapas e os aptos informados, registra os votos
// ("C1", "branco", "nulo"...) e a encerra. Devolve o id e o id de cada chapa pelo nome.
func eleicaoEncerrada(t *testing.T, c *cenario, chapas []string, aptos int, votos []string) (int64, map[string]int64) {
	t.Helper()
	eleitores := make([]int64, aptos)
	for i := range eleitores {
		eleitores[i] = c.usuario(t, fmt.Sprintf("estudante%d", i+1), PapelEleitor)
	}
	id := c.votacao(t, NovaVotacao{Tipo: TipoEleicao, Opcoes: opcoesNomeadas(chapas...), Aptos: eleitores})

	chapaPorNome := map[string]int64{}
	for i, opcaoID := range c.opcoes(t, id) {
		chapaPorNome[chapas[i]] = opcaoID
	}
	for i, escolha := range votos {
		voto := Voto{Escolha: escolha}
		if opcaoID, ehChapa := chapaPorNome[escolha]; ehChapa {
			voto = Voto{Escolha: EscolhaChapa, OpcaoID: &opcaoID}
		}
		c.votar(t, eleitores[i], id, voto)
	}
	c.encerrar(t, id)
	return id, chapaPorNome
}

func TestResultadoEleicao(t *testing.T) {
	testes := []struct {
		nome             string
		chapas           []string
		aptos            int
		votos            []string
		ordem            []string // chapas na ordem do resultado
		percentuais      []float64
		brancos, nulos   int
		vencedora        string
		necessario       bool
		empateIndefinido bool
		segundoTurno     []string
		comparecimento   float64
	}{
		{
			nome: "maioria absoluta vence no 1º turno", chapas: []string{"C1", "C2"}, aptos: 4,
			votos: []string{"C2", "C1", "C2", "branco"}, ordem: []string{"C2", "C1"}, percentuais: []float64{200.0 / 3, 100.0 / 3},
			brancos: 1, vencedora: "C2", comparecimento: 100,
		},
		{
			nome: "metade exata não é maioria: 2º turno entre as duas mais votadas", chapas: []string{"C1", "C2", "C3"}, aptos: 8,
			votos: []string{"C1", "C1", "C1", "C2", "C2", "C3", "branco"}, ordem: []string{"C1", "C2", "C3"},
			percentuais: []float64{50, 100.0 / 3, 100.0 / 6}, brancos: 1, necessario: true, segundoTurno: []string{"C1", "C2"},
			comparecimento: 87.5,
		},
		{
			nome: "empate pelo 2º lugar impede definir o 2º turno", chapas: []string{"C1", "C2", "C3"}, aptos: 4,
			votos: []string{"C3", "C1", "C1", "C2"}, ordem: []string{"C1", "C2", "C3"}, percentuais: []float64{50, 25, 25},
			necessario: true, empateIndefinido: true, comparecimento: 100,
		},
		{
			nome: "só brancos e nulos: sem vencedora e sem 2º turno", chapas: []string{"C1", "C2"}, aptos: 3,
			votos: []string{"branco", "nulo"}, ordem: []string{"C1", "C2"}, percentuais: []float64{0, 0},
			brancos: 1, nulos: 1, comparecimento: 200.0 / 3,
		},
		{
			nome: "ninguém votou", chapas: []string{"C1", "C2"}, aptos: 2,
			ordem: []string{"C1", "C2"}, percentuais: []float64{0, 0},
		},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			c := novoCenario(t)
			id, chapa := eleicaoEncerrada(t, c, tc.chapas, tc.aptos, tc.votos)

			resultado, err := c.svc.Resultado(t.Context(), id)
			if err != nil {
				t.Fatal(err)
			}
			r := resultado.Eleicao
			if r == nil {
				t.Fatalf("esperava resultado de eleição: %+v", resultado)
			}

			validos := len(tc.votos) - tc.brancos - tc.nulos
			if r.VotacaoID != id || r.Titulo != "Votação de teste" || r.Descricao != "" || r.Turno != 1 || r.Aptos != tc.aptos || r.Participantes != len(tc.votos) ||
				r.Abstencoes != tc.aptos-len(tc.votos) || r.Brancos != tc.brancos || r.Nulos != tc.nulos || r.Validos != validos ||
				!quase(r.PercentualComparecimento, tc.comparecimento) {
				t.Errorf("totais inesperados: %+v", r)
			}

			if len(r.Votos) != len(tc.ordem) {
				t.Fatalf("votos por chapa: esperado %d linhas, obtido %+v", len(tc.ordem), r.Votos)
			}
			for i, nome := range tc.ordem {
				linha := r.Votos[i]
				if linha.Nome != nome || linha.OpcaoID != chapa[nome] || !quase(linha.PercentualValidos, tc.percentuais[i]) {
					t.Errorf("linha %d: esperado %s com %.2f%%, obtido %+v", i, nome, tc.percentuais[i], linha)
				}
			}

			switch {
			case tc.vencedora == "" && r.Vencedora != nil:
				t.Errorf("não esperava vencedora, obteve %+v", r.Vencedora)
			case tc.vencedora != "" && (r.Vencedora == nil || *r.Vencedora != ChapaResumo{OpcaoID: chapa[tc.vencedora], Nome: tc.vencedora}):
				t.Errorf("vencedora: esperado %s, obtido %+v", tc.vencedora, r.Vencedora)
			}

			st := r.SegundoTurno
			var nomes []string
			for _, ch := range st.Chapas {
				nomes = append(nomes, ch.Nome)
			}
			if st.Necessario != tc.necessario || st.EmpateIndefinido != tc.empateIndefinido ||
				!slices.Equal(nomes, tc.segundoTurno) || st.Chapas == nil || st.VotacaoID != nil {
				t.Errorf("segundo turno inesperado: %+v", st)
			}
		})
	}
}

func quase(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestResultadoConsulta(t *testing.T) {
	type projeto struct {
		nome  string
		custo int64
	}
	seed := []projeto{{"Impressoras", 35000_00}, {"Arduino", 12000_00}, {"GPU", 48000_00}, {"Mobiliário", 9000_00}}

	c := novoCenario(t)
	eleitores := map[string]int64{}
	for _, login := range []string{"ana", "bruno", "carla", "maria", "joao"} {
		eleitores[login] = c.usuario(t, login, PapelEleitor)
	}
	opcoes := make([]NovaOpcao, len(seed))
	for i, p := range seed {
		opcoes[i] = NovaOpcao{Nome: p.nome, CustoCentavos: ptr(p.custo)}
	}
	consultaID := c.votacao(t, NovaVotacao{Tipo: TipoConsulta, Titulo: "Orçamento participativo", Descricao: "Ordene os projetos.",
		Opcoes: opcoes, OrcamentoCentavos: ptr[int64](60000_00)})
	ids := c.opcoes(t, consultaID)
	impressoras, arduino, gpu, mobiliario := ids[0], ids[1], ids[2], ids[3]

	cedulas := map[string][]int64{
		"ana":   {impressoras, arduino, gpu, mobiliario},
		"bruno": {gpu, impressoras, arduino, mobiliario},
		"carla": {arduino, impressoras, mobiliario, gpu},
		"maria": {impressoras, mobiliario, arduino, gpu},
		"joao":  {gpu, arduino, impressoras, mobiliario},
	}
	for login, preferencias := range cedulas {
		c.votar(t, eleitores[login], consultaID, Voto{Preferencias: preferencias})
	}
	c.encerrar(t, consultaID)

	resultado, err := c.svc.Resultado(t.Context(), consultaID)
	if err != nil {
		t.Fatal(err)
	}
	r := resultado.Consulta
	if r == nil {
		t.Fatalf("esperava resultado de consulta: %+v", resultado)
	}
	if r.VotacaoID != consultaID || r.Titulo != "Orçamento participativo" || r.Descricao != "Ordene os projetos." || r.Cedulas != 5 || r.OrcamentoCentavos != 60000_00 ||
		r.TotalSelecionadoCentavos != 47000_00 || r.SaldoCentavos != 13000_00 || r.EmpateNaLinhaDeCorte {
		t.Errorf("totais inesperados: %+v", r)
	}

	esperado := []ItemRanking{
		{Posicao: 1, OpcaoID: impressoras, Nome: "Impressoras", Pontos: 16, PrimeirasPreferencias: 2, CustoCentavos: 35000_00, CustoAcumuladoCentavos: 35000_00, DentroDoOrcamento: true},
		{Posicao: 2, OpcaoID: arduino, Nome: "Arduino", Pontos: 14, PrimeirasPreferencias: 1, CustoCentavos: 12000_00, CustoAcumuladoCentavos: 47000_00, DentroDoOrcamento: true},
		{Posicao: 3, OpcaoID: gpu, Nome: "GPU", Pontos: 12, PrimeirasPreferencias: 2, CustoCentavos: 48000_00, CustoAcumuladoCentavos: 95000_00},
		{Posicao: 4, OpcaoID: mobiliario, Nome: "Mobiliário", Pontos: 8, CustoCentavos: 9000_00, CustoAcumuladoCentavos: 104000_00},
	}
	if !slices.Equal(r.Ranking, esperado) {
		t.Errorf("ranking:\n esperado %+v\n obtido   %+v", esperado, r.Ranking)
	}
}

func TestResultadoConsultaSemCedulas(t *testing.T) {
	c := novoCenario(t)
	consultaID := c.votacao(t, NovaVotacao{Tipo: TipoConsulta, Opcoes: projetos(10, 20), OrcamentoCentavos: ptr[int64](15)})
	c.encerrar(t, consultaID)

	resultado, err := c.svc.Resultado(t.Context(), consultaID)
	if err != nil {
		t.Fatal(err)
	}
	r := resultado.Consulta
	// Sem cédulas, todos ficam com 0 pontos: mantém a ordem de cadastro e marca o empate.
	if r.Cedulas != 0 || len(r.Ranking) != 2 || r.Ranking[0].Nome != "A" || !r.Ranking[0].Empatado ||
		!r.Ranking[0].DentroDoOrcamento || r.Ranking[1].DentroDoOrcamento || !r.EmpateNaLinhaDeCorte {
		t.Errorf("resultado inesperado: %+v", r)
	}
}

func TestResultadoConsultaComCustosNoLimite(t *testing.T) {
	c := novoCenario(t)
	custos := make([]int64, maxOpcoes)
	for i := range custos {
		custos[i] = maxCentavos
	}
	consultaID := c.votacao(t, NovaVotacao{Tipo: TipoConsulta, Opcoes: projetos(custos...), OrcamentoCentavos: ptr[int64](maxCentavos)})
	c.encerrar(t, consultaID)

	resultado, err := c.svc.Resultado(t.Context(), consultaID)
	if err != nil {
		t.Fatalf("a soma dos custos máximos deveria caber na apuração: %v", err)
	}
	r := resultado.Consulta
	ultimo := r.Ranking[len(r.Ranking)-1]
	if r.TotalSelecionadoCentavos != maxCentavos || r.SaldoCentavos != 0 || ultimo.CustoAcumuladoCentavos != maxOpcoes*maxCentavos ||
		ultimo.CustoAcumuladoCentavos > 1<<53 {
		t.Errorf("resultado inesperado: total %d, saldo %d, acumulado %d", r.TotalSelecionadoCentavos, r.SaldoCentavos, ultimo.CustoAcumuladoCentavos)
	}
}
