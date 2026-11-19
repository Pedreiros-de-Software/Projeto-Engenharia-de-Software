package server

import (
	"time"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/service"
)

// Corpos recebidos.

type pedidoLogin struct {
	Login string `json:"login"`
	Senha string `json:"senha"`
}

type pedidoNovaOpcao struct {
	Nome          string `json:"nome"`
	Descricao     string `json:"descricao"`
	CustoCentavos *int64 `json:"custo_centavos"`
}

type pedidoNovaVotacao struct {
	Tipo              string            `json:"tipo"`
	Titulo            string            `json:"titulo"`
	Descricao         string            `json:"descricao"`
	Inicio            time.Time         `json:"inicio"`
	Fim               time.Time         `json:"fim"`
	Opcoes            []pedidoNovaOpcao `json:"opcoes"`
	OrcamentoCentavos *int64            `json:"orcamento_centavos"`
	Aptos             []int64           `json:"aptos"`
}

type pedidoVoto struct {
	Escolha      string  `json:"escolha"`
	OpcaoID      *int64  `json:"opcao_id"`
	Preferencias []int64 `json:"preferencias"`
}

type pedidoSegundoTurno struct {
	Fim *time.Time `json:"fim"`
}

type pedidoAlterarUnidade struct {
	Adimplente *bool `json:"adimplente"`
}

// Corpos enviados.

type respostaLogin struct {
	Token   string     `json:"token"`
	Usuario usuarioDTO `json:"usuario"`
}

type respostaMe struct {
	Usuario usuarioDTO  `json:"usuario"`
	Unidade *unidadeDTO `json:"unidade"`
}

type respostaVoto struct {
	Mensagem string `json:"mensagem"`
}

type usuarioDTO struct {
	ID    int64  `json:"id"`
	Nome  string `json:"nome"`
	Login string `json:"login"`
	Papel string `json:"papel"`
}

type responsavelDTO struct {
	ID   int64  `json:"id"`
	Nome string `json:"nome"`
}

type unidadeDTO struct {
	ID            int64           `json:"id"`
	Identificador string          `json:"identificador"`
	FracaoIdeal   float64         `json:"fracao_ideal"`
	Percentual    float64         `json:"percentual"`
	Adimplente    bool            `json:"adimplente"`
	Responsavel   *responsavelDTO `json:"responsavel"`
}

type votacaoResumoDTO struct {
	ID                int64  `json:"id"`
	Tipo              string `json:"tipo"`
	Titulo            string `json:"titulo"`
	Descricao         string `json:"descricao"`
	Inicio            string `json:"inicio"`
	Fim               string `json:"fim"`
	Turno             int    `json:"turno"`
	Status            string `json:"status"`
	Apto              bool   `json:"apto"`
	JaVotou           bool   `json:"ja_votou"`
	VotacaoAnteriorID *int64 `json:"votacao_anterior_id"`
	SegundoTurnoID    *int64 `json:"segundo_turno_id"`
}

type opcaoDTO struct {
	ID            int64  `json:"id"`
	Nome          string `json:"nome"`
	Descricao     string `json:"descricao"`
	CustoCentavos *int64 `json:"custo_centavos"`
}

type minhaUnidadeDTO struct {
	Identificador string  `json:"identificador"`
	FracaoIdeal   float64 `json:"fracao_ideal"`
	Percentual    float64 `json:"percentual"`
	Adimplente    bool    `json:"adimplente"`
	Elegivel      bool    `json:"elegivel"`
	Motivo        string  `json:"motivo"`
}

type votacaoDetalheDTO struct {
	votacaoResumoDTO
	Opcoes            []opcaoDTO       `json:"opcoes"`
	OrcamentoCentavos *int64           `json:"orcamento_centavos"`
	MinhaUnidade      *minhaUnidadeDTO `json:"minha_unidade"`
}

type contagemPautaDTO struct {
	Favor     float64 `json:"favor"`
	Contra    float64 `json:"contra"`
	Abstencao float64 `json:"abstencao"`
}

type resultadoPautaDTO struct {
	Tipo             string           `json:"tipo"`
	VotacaoID        int64            `json:"votacao_id"`
	Titulo           string           `json:"titulo"`
	Descricao        string           `json:"descricao"`
	Contagem         contagemPautaDTO `json:"contagem"`
	PesoTotal        float64          `json:"peso_total"`
	Limiar           float64          `json:"limiar"`
	PercentualFavor  float64          `json:"percentual_favor"`
	Aprovada         bool             `json:"aprovada"`
	UnidadesTotal    int              `json:"unidades_total"`
	UnidadesVotantes int              `json:"unidades_votantes"`
}

type chapaDTO struct {
	OpcaoID int64  `json:"opcao_id"`
	Nome    string `json:"nome"`
}

type votosChapaDTO struct {
	OpcaoID           int64   `json:"opcao_id"`
	Nome              string  `json:"nome"`
	Votos             int     `json:"votos"`
	PercentualValidos float64 `json:"percentual_validos"`
}

type segundoTurnoDTO struct {
	Necessario       bool       `json:"necessario"`
	EmpateIndefinido bool       `json:"empate_indefinido"`
	Chapas           []chapaDTO `json:"chapas"`
	VotacaoID        *int64     `json:"votacao_id"`
}

type resultadoEleicaoDTO struct {
	Tipo                     string          `json:"tipo"`
	VotacaoID                int64           `json:"votacao_id"`
	Titulo                   string          `json:"titulo"`
	Descricao                string          `json:"descricao"`
	Turno                    int             `json:"turno"`
	Aptos                    int             `json:"aptos"`
	Participantes            int             `json:"participantes"`
	Abstencoes               int             `json:"abstencoes"`
	PercentualComparecimento float64         `json:"percentual_comparecimento"`
	Votos                    []votosChapaDTO `json:"votos"`
	Brancos                  int             `json:"brancos"`
	Nulos                    int             `json:"nulos"`
	Validos                  int             `json:"validos"`
	Vencedora                *chapaDTO       `json:"vencedora"`
	SegundoTurno             segundoTurnoDTO `json:"segundo_turno"`
}

type itemRankingDTO struct {
	Posicao                int    `json:"posicao"`
	OpcaoID                int64  `json:"opcao_id"`
	Nome                   string `json:"nome"`
	Pontos                 int    `json:"pontos"`
	PrimeirasPreferencias  int    `json:"primeiras_preferencias"`
	CustoCentavos          int64  `json:"custo_centavos"`
	CustoAcumuladoCentavos int64  `json:"custo_acumulado_centavos"`
	DentroDoOrcamento      bool   `json:"dentro_do_orcamento"`
	Empatado               bool   `json:"empatado"`
}

type resultadoConsultaDTO struct {
	Tipo                     string           `json:"tipo"`
	VotacaoID                int64            `json:"votacao_id"`
	Titulo                   string           `json:"titulo"`
	Descricao                string           `json:"descricao"`
	Cedulas                  int              `json:"cedulas"`
	OrcamentoCentavos        int64            `json:"orcamento_centavos"`
	TotalSelecionadoCentavos int64            `json:"total_selecionado_centavos"`
	SaldoCentavos            int64            `json:"saldo_centavos"`
	EmpateNaLinhaDeCorte     bool             `json:"empate_na_linha_de_corte"`
	Ranking                  []itemRankingDTO `json:"ranking"`
}

// Conversões dos tipos do service para os DTOs.

// formatarData usa RFC 3339 em UTC, o formato de datas do contrato.
func formatarData(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func paraUsuarioDTO(u service.Usuario) usuarioDTO {
	return usuarioDTO{ID: u.ID, Nome: u.Nome, Login: u.Login, Papel: u.Papel}
}

func paraUnidadeDTO(u service.Unidade) unidadeDTO {
	dto := unidadeDTO{
		ID:            u.ID,
		Identificador: u.Identificador,
		FracaoIdeal:   u.FracaoIdeal,
		Percentual:    u.Percentual,
		Adimplente:    u.Adimplente,
	}
	if u.Responsavel != nil {
		dto.Responsavel = &responsavelDTO{ID: u.Responsavel.ID, Nome: u.Responsavel.Nome}
	}
	return dto
}

func paraResumoDTO(v service.VotacaoResumo) votacaoResumoDTO {
	return votacaoResumoDTO{
		ID:                v.ID,
		Tipo:              v.Tipo,
		Titulo:            v.Titulo,
		Descricao:         v.Descricao,
		Inicio:            formatarData(v.Inicio),
		Fim:               formatarData(v.Fim),
		Turno:             v.Turno,
		Status:            v.Status,
		Apto:              v.Apto,
		JaVotou:           v.JaVotou,
		VotacaoAnteriorID: v.VotacaoAnteriorID,
		SegundoTurnoID:    v.SegundoTurnoID,
	}
}

func paraDetalheDTO(d service.VotacaoDetalhe) votacaoDetalheDTO {
	dto := votacaoDetalheDTO{
		votacaoResumoDTO:  paraResumoDTO(d.VotacaoResumo),
		Opcoes:            make([]opcaoDTO, len(d.Opcoes)),
		OrcamentoCentavos: d.OrcamentoCentavos,
	}
	for i, o := range d.Opcoes {
		dto.Opcoes[i] = opcaoDTO{ID: o.ID, Nome: o.Nome, Descricao: o.Descricao, CustoCentavos: o.CustoCentavos}
	}
	if u := d.MinhaUnidade; u != nil {
		dto.MinhaUnidade = &minhaUnidadeDTO{
			Identificador: u.Identificador,
			FracaoIdeal:   u.FracaoIdeal,
			Percentual:    u.Percentual,
			Adimplente:    u.Adimplente,
			Elegivel:      u.Elegivel,
			Motivo:        u.Motivo,
		}
	}
	return dto
}

// paraResultadoDTO devolve o DTO do tipo de resultado preenchido pelo service.
func paraResultadoDTO(r service.Resultado) any {
	switch {
	case r.Pauta != nil:
		return paraResultadoPautaDTO(*r.Pauta)
	case r.Eleicao != nil:
		return paraResultadoEleicaoDTO(*r.Eleicao)
	default:
		return paraResultadoConsultaDTO(*r.Consulta)
	}
}

func paraResultadoPautaDTO(r service.ResultadoPauta) resultadoPautaDTO {
	return resultadoPautaDTO{
		Tipo:             service.TipoPauta,
		VotacaoID:        r.VotacaoID,
		Titulo:           r.Titulo,
		Descricao:        r.Descricao,
		Contagem:         contagemPautaDTO{Favor: r.Favor, Contra: r.Contra, Abstencao: r.Abstencao},
		PesoTotal:        r.PesoTotal,
		Limiar:           r.Limiar,
		PercentualFavor:  r.PercentualFavor,
		Aprovada:         r.Aprovada,
		UnidadesTotal:    r.UnidadesTotal,
		UnidadesVotantes: r.UnidadesVotantes,
	}
}

func paraChapaDTO(c service.ChapaResumo) chapaDTO {
	return chapaDTO{OpcaoID: c.OpcaoID, Nome: c.Nome}
}

func paraResultadoEleicaoDTO(r service.ResultadoEleicao) resultadoEleicaoDTO {
	dto := resultadoEleicaoDTO{
		Tipo:                     service.TipoEleicao,
		VotacaoID:                r.VotacaoID,
		Titulo:                   r.Titulo,
		Descricao:                r.Descricao,
		Turno:                    r.Turno,
		Aptos:                    r.Aptos,
		Participantes:            r.Participantes,
		Abstencoes:               r.Abstencoes,
		PercentualComparecimento: r.PercentualComparecimento,
		Votos:                    make([]votosChapaDTO, len(r.Votos)),
		Brancos:                  r.Brancos,
		Nulos:                    r.Nulos,
		Validos:                  r.Validos,
		SegundoTurno: segundoTurnoDTO{
			Necessario:       r.SegundoTurno.Necessario,
			EmpateIndefinido: r.SegundoTurno.EmpateIndefinido,
			Chapas:           make([]chapaDTO, len(r.SegundoTurno.Chapas)),
			VotacaoID:        r.SegundoTurno.VotacaoID,
		},
	}
	for i, v := range r.Votos {
		dto.Votos[i] = votosChapaDTO{OpcaoID: v.OpcaoID, Nome: v.Nome, Votos: v.Votos, PercentualValidos: v.PercentualValidos}
	}
	for i, c := range r.SegundoTurno.Chapas {
		dto.SegundoTurno.Chapas[i] = paraChapaDTO(c)
	}
	if r.Vencedora != nil {
		vencedora := paraChapaDTO(*r.Vencedora)
		dto.Vencedora = &vencedora
	}
	return dto
}

func paraResultadoConsultaDTO(r service.ResultadoConsulta) resultadoConsultaDTO {
	dto := resultadoConsultaDTO{
		Tipo:                     service.TipoConsulta,
		VotacaoID:                r.VotacaoID,
		Titulo:                   r.Titulo,
		Descricao:                r.Descricao,
		Cedulas:                  r.Cedulas,
		OrcamentoCentavos:        r.OrcamentoCentavos,
		TotalSelecionadoCentavos: r.TotalSelecionadoCentavos,
		SaldoCentavos:            r.SaldoCentavos,
		EmpateNaLinhaDeCorte:     r.EmpateNaLinhaDeCorte,
		Ranking:                  make([]itemRankingDTO, len(r.Ranking)),
	}
	for i, p := range r.Ranking {
		dto.Ranking[i] = itemRankingDTO{
			Posicao:                p.Posicao,
			OpcaoID:                p.OpcaoID,
			Nome:                   p.Nome,
			Pontos:                 p.Pontos,
			PrimeirasPreferencias:  p.PrimeirasPreferencias,
			CustoCentavos:          p.CustoCentavos,
			CustoAcumuladoCentavos: p.CustoAcumuladoCentavos,
			DentroDoOrcamento:      p.DentroDoOrcamento,
			Empatado:               p.Empatado,
		}
	}
	return dto
}
