package service

import (
	"testing"
	"time"
)

func TestListarVotacoes(t *testing.T) {
	c := novoCenario(t)
	adminID := c.usuario(t, "admin", PapelAdmin)
	mariaID := c.usuario(t, "maria", PapelEleitor)
	anaID := c.usuario(t, "ana", PapelEleitor)
	c.unidade(t, "Apto 101", 60, true, mariaID)

	agora := c.relogio.Agora()
	pauta := c.votacao(t, NovaVotacao{Tipo: TipoPauta})
	eleicao := c.votacao(t, NovaVotacao{Tipo: TipoEleicao, Opcoes: opcoesNomeadas("C1", "C2"), Aptos: []int64{anaID}})
	consulta := c.votacao(t, NovaVotacao{Tipo: TipoConsulta, Opcoes: projetos(1, 2), OrcamentoCentavos: ptr[int64](3)})
	futura := c.votacao(t, NovaVotacao{Tipo: TipoPauta, Inicio: agora.Add(time.Hour), Fim: agora.Add(2 * time.Hour)})

	c.votar(t, mariaID, pauta, Voto{Escolha: "favor"})
	c.votar(t, anaID, eleicao, Voto{Escolha: "branco"})

	type situacao struct {
		status        string
		apto, jaVotou bool
	}
	testes := []struct {
		nome      string
		usuarioID int64
		esperado  map[int64]situacao
	}{
		{"responsável por unidade", mariaID, map[int64]situacao{
			futura:   {StatusNaoIniciada, true, false},
			consulta: {StatusAberta, true, false},
			eleicao:  {StatusAberta, false, false},
			pauta:    {StatusAberta, true, true},
		}},
		{"estudante apto na eleição", anaID, map[int64]situacao{
			futura:   {StatusNaoIniciada, false, false},
			consulta: {StatusAberta, true, false},
			eleicao:  {StatusAberta, true, true},
			pauta:    {StatusAberta, false, false},
		}},
		{"administrador não vota", adminID, map[int64]situacao{
			futura:   {StatusNaoIniciada, false, false},
			consulta: {StatusAberta, false, false},
			eleicao:  {StatusAberta, false, false},
			pauta:    {StatusAberta, false, false},
		}},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			resumos, err := c.svc.ListarVotacoes(t.Context(), tc.usuarioID)
			if err != nil {
				t.Fatal(err)
			}
			ordem := []int64{futura, consulta, eleicao, pauta}
			if len(resumos) != len(ordem) {
				t.Fatalf("esperava %d votações, obteve %d", len(ordem), len(resumos))
			}
			for i, r := range resumos {
				if r.ID != ordem[i] {
					t.Errorf("posição %d: esperado id %d (mais recentes primeiro), obtido %d", i, ordem[i], r.ID)
				}
				obtido := situacao{r.Status, r.Apto, r.JaVotou}
				if obtido != tc.esperado[r.ID] {
					t.Errorf("votação %d (%s): esperado %+v, obtido %+v", r.ID, r.Tipo, tc.esperado[r.ID], obtido)
				}
			}
		})
	}
}

func TestDetalhar(t *testing.T) {
	c := novoCenario(t)
	mariaID := c.usuario(t, "maria", PapelEleitor)
	sofiaID := c.usuario(t, "sofia", PapelEleitor)
	anaID := c.usuario(t, "ana", PapelEleitor)
	c.unidade(t, "Apto 101", 60, true, mariaID)
	c.unidade(t, "Apto 402", 120, false, sofiaID)
	pauta := c.votacao(t, NovaVotacao{Tipo: TipoPauta, Descricao: "Reforma"})
	eleicao := c.votacao(t, NovaVotacao{Tipo: TipoEleicao, Opcoes: opcoesNomeadas("C1", "C2")})

	testes := []struct {
		nome         string
		usuarioID    int64
		votacaoID    int64
		minhaUnidade *MinhaUnidade
	}{
		{"pauta para responsável adimplente", mariaID, pauta,
			&MinhaUnidade{Identificador: "Apto 101", FracaoIdeal: 60, Percentual: percentual(60, 180), Adimplente: true, Elegivel: true}},
		{"pauta para responsável inadimplente", sofiaID, pauta,
			&MinhaUnidade{Identificador: "Apto 402", FracaoIdeal: 120, Percentual: percentual(120, 180), Motivo: MotivoInadimplente}},
		{"pauta para usuário sem unidade", anaID, pauta, nil},
		{"eleição não mostra unidade", mariaID, eleicao, nil},
	}
	for _, tc := range testes {
		t.Run(tc.nome, func(t *testing.T) {
			detalhe, err := c.svc.Detalhar(t.Context(), tc.usuarioID, tc.votacaoID)
			if err != nil {
				t.Fatal(err)
			}
			switch {
			case tc.minhaUnidade == nil && detalhe.MinhaUnidade != nil:
				t.Errorf("não esperava minha_unidade, obteve %+v", detalhe.MinhaUnidade)
			case tc.minhaUnidade != nil && (detalhe.MinhaUnidade == nil || *detalhe.MinhaUnidade != *tc.minhaUnidade):
				t.Errorf("minha_unidade: esperado %+v, obtido %+v", tc.minhaUnidade, detalhe.MinhaUnidade)
			}
		})
	}

	t.Run("pauta não tem opções nem orçamento", func(t *testing.T) {
		detalhe, _ := c.svc.Detalhar(t.Context(), mariaID, pauta)
		if detalhe.Opcoes == nil || len(detalhe.Opcoes) != 0 || detalhe.OrcamentoCentavos != nil || detalhe.Descricao != "Reforma" {
			t.Errorf("detalhe inesperado: %+v", detalhe)
		}
	})

	t.Run("eleição lista as chapas", func(t *testing.T) {
		detalhe, _ := c.svc.Detalhar(t.Context(), anaID, eleicao)
		if len(detalhe.Opcoes) != 2 || detalhe.Opcoes[0].Nome != "C1" || detalhe.Opcoes[0].CustoCentavos != nil || !detalhe.Apto {
			t.Errorf("detalhe inesperado: %+v", detalhe)
		}
	})

	t.Run("votação inexistente", func(t *testing.T) {
		_, err := c.svc.Detalhar(t.Context(), mariaID, 9999)
		conferirErro(t, err, ErrNaoEncontrado)
	})
}
