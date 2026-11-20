package service

import (
	"context"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store"
	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/voting"
)

// Resultado tem exatamente um dos campos preenchido, conforme o tipo da votação.
type Resultado struct {
	Pauta    *ResultadoPauta
	Eleicao  *ResultadoEleicao
	Consulta *ResultadoConsulta
}

// ResultadoPauta traz os pesos (frações ideais) somados por escolha.
type ResultadoPauta struct {
	VotacaoID        int64
	Titulo           string
	Descricao        string
	Favor            float64
	Contra           float64
	Abstencao        float64
	PesoTotal        float64
	Limiar           float64 // 2/3 do peso total
	PercentualFavor  float64
	Aprovada         bool
	UnidadesTotal    int
	UnidadesVotantes int
}

type ChapaResumo struct {
	OpcaoID int64
	Nome    string
}

type VotosChapa struct {
	OpcaoID           int64
	Nome              string
	Votos             int
	PercentualValidos float64
}

type SegundoTurno struct {
	Necessario       bool
	EmpateIndefinido bool          // empate pelo 2º lugar impede escolher as duas chapas
	Chapas           []ChapaResumo // as duas chapas, quando necessário e definido
	VotacaoID        *int64        // 2º turno já aberto, se houver
}

// ResultadoEleicao é o boletim público de um turno.
type ResultadoEleicao struct {
	VotacaoID                int64
	Titulo                   string
	Descricao                string
	Turno                    int
	Aptos                    int
	Participantes            int
	Abstencoes               int
	PercentualComparecimento float64
	Votos                    []VotosChapa // ordem do resultado
	Brancos                  int
	Nulos                    int
	Validos                  int
	Vencedora                *ChapaResumo
	SegundoTurno             SegundoTurno
}

type ItemRanking struct {
	Posicao                int
	OpcaoID                int64
	Nome                   string
	Pontos                 int
	PrimeirasPreferencias  int
	CustoCentavos          int64
	CustoAcumuladoCentavos int64
	DentroDoOrcamento      bool
	Empatado               bool
}

// ResultadoConsulta é o ranking Borda com a linha de corte orçamentária.
type ResultadoConsulta struct {
	VotacaoID                int64
	Titulo                   string
	Descricao                string
	Cedulas                  int
	OrcamentoCentavos        int64
	TotalSelecionadoCentavos int64
	SaldoCentavos            int64
	EmpateNaLinhaDeCorte     bool
	Ranking                  []ItemRanking
}

// Resultado apura uma votação encerrada; é público (não depende do usuário). Todas as leituras
// vêm do mesmo retrato do banco, para os totais baterem mesmo com um voto confirmando no meio.
func (s *Service) Resultado(ctx context.Context, votacaoID int64) (Resultado, error) {
	var resultado Resultado
	err := store.EmLeitura(ctx, s.db, func(r store.Repositorios) error {
		v, err := buscarVotacao(ctx, r, votacaoID)
		if err != nil {
			return err
		}
		status, err := statusEm(v, s.relogio())
		if err != nil {
			return err
		}
		if status != StatusEncerrada {
			return ErrVotacaoNaoEncerrada
		}

		switch v.Tipo {
		case TipoPauta:
			resultado.Pauta, err = resultadoPauta(ctx, r, v)
		case TipoEleicao:
			resultado.Eleicao, err = resultadoEleicao(ctx, r, v)
		default:
			resultado.Consulta, err = resultadoConsulta(ctx, r, v)
		}
		return err
	})
	if err != nil {
		return Resultado{}, err
	}
	return resultado, nil
}

// resultadoPauta usa a Facade voting.ApurarPauta; o peso total inclui as unidades que não votaram.
func resultadoPauta(ctx context.Context, r store.Repositorios, v store.Votacao) (*ResultadoPauta, error) {
	unidades, err := r.Unidades.Listar(ctx)
	if err != nil {
		return nil, err
	}
	votos, err := r.Votos.ListarVotosPauta(ctx, v.ID)
	if err != nil {
		return nil, err
	}

	condominos := []voting.Condomino{}
	for _, u := range unidades {
		if u.ResponsavelID != nil {
			condominos = append(condominos, condominoDe(u))
		}
	}
	votosDominio := make([]voting.Voto, len(votos))
	for i, voto := range votos {
		votosDominio[i] = voting.Voto{EleitorID: idTexto(voto.UnidadeID), Opcao: voto.Escolha}
	}

	apuracao, err := voting.ApurarPauta(condominos, votosDominio)
	if err != nil {
		return nil, err
	}
	favor := apuracao.Contagem[voting.OpcaoFavor]
	return &ResultadoPauta{
		VotacaoID:        v.ID,
		Titulo:           v.Titulo,
		Descricao:        v.Descricao,
		Favor:            favor,
		Contra:           apuracao.Contagem[voting.OpcaoContra],
		Abstencao:        apuracao.Contagem[voting.OpcaoAbstencao],
		PesoTotal:        apuracao.PesoTotal,
		Limiar:           apuracao.PesoTotal * 2 / 3,
		PercentualFavor:  percentual(favor, apuracao.PesoTotal),
		Aprovada:         apuracao.Aprovada,
		UnidadesTotal:    len(condominos),
		UnidadesVotantes: len(votos),
	}, nil
}

// apuracaoEleicao é o resultado do domínio mais as chapas, indexadas pelo id em texto.
type apuracaoEleicao struct {
	resultado voting.ResultadoEleicao
	chapas    map[string]store.Opcao
}

func (a apuracaoEleicao) resumoDaChapa(chave string) ChapaResumo {
	chapa := a.chapas[chave]
	return ChapaResumo{OpcaoID: chapa.ID, Nome: chapa.Nome}
}

// apurarEleicao converte as cédulas anônimas em votos do domínio e aplica voting.ApurarEleicao.
func apurarEleicao(ctx context.Context, r store.Repositorios, v store.Votacao) (apuracaoEleicao, error) {
	opcoes, err := r.Votacoes.ListarOpcoes(ctx, v.ID)
	if err != nil {
		return apuracaoEleicao{}, err
	}
	cedulas, err := r.Votos.ListarCedulasEleicao(ctx, v.ID)
	if err != nil {
		return apuracaoEleicao{}, err
	}

	chaves := make([]string, len(opcoes))
	porChave := make(map[string]store.Opcao, len(opcoes))
	for i, o := range opcoes {
		chaves[i] = idTexto(o.ID)
		porChave[chaves[i]] = o
	}
	votos := make([]voting.Voto, len(cedulas))
	for i, c := range cedulas {
		votos[i] = voting.Voto{Opcao: c.Tipo} // branco ou nulo
		if c.Tipo == EscolhaChapa {
			votos[i].Opcao = idTexto(*c.OpcaoID)
		}
	}

	resultado, err := voting.ApurarEleicao(chaves, votos)
	return apuracaoEleicao{resultado: resultado, chapas: porChave}, err
}

func resultadoEleicao(ctx context.Context, r store.Repositorios, v store.Votacao) (*ResultadoEleicao, error) {
	apuracao, err := apurarEleicao(ctx, r, v)
	if err != nil {
		return nil, err
	}
	aptos, err := r.Votacoes.ContarAptos(ctx, v.ID)
	if err != nil {
		return nil, err
	}
	participantes, err := r.Votos.ContarParticipacoes(ctx, v.ID)
	if err != nil {
		return nil, err
	}
	boletim, err := voting.MontarBoletim(aptos, participantes, apuracao.resultado)
	if err != nil {
		return nil, err
	}

	apurado := apuracao.resultado
	resultado := &ResultadoEleicao{
		VotacaoID:                v.ID,
		Titulo:                   v.Titulo,
		Descricao:                v.Descricao,
		Turno:                    v.Turno,
		Aptos:                    boletim.Aptos,
		Participantes:            boletim.Participantes,
		Abstencoes:               boletim.Abstencoes,
		PercentualComparecimento: boletim.PercentualComparecimento,
		Votos:                    make([]VotosChapa, len(apurado.Chapas)),
		Brancos:                  apurado.Brancos,
		Nulos:                    apurado.Nulos,
		Validos:                  apurado.Validos,
		SegundoTurno:             SegundoTurno{Chapas: []ChapaResumo{}, VotacaoID: v.SegundoTurnoID},
	}
	for i, chapa := range apurado.Chapas {
		resumo := apuracao.resumoDaChapa(chapa.Opcao)
		resultado.Votos[i] = VotosChapa{
			OpcaoID:           resumo.OpcaoID,
			Nome:              resumo.Nome,
			Votos:             chapa.Votos,
			PercentualValidos: boletim.PercentuaisValidos[chapa.Opcao],
		}
	}
	if apurado.Vencedora != "" {
		vencedora := apuracao.resumoDaChapa(apurado.Vencedora)
		resultado.Vencedora = &vencedora
	}
	// Só o 1º turno leva a um 2º turno; empate no 2º turno fica para a comissão.
	if v.Turno == 1 && apurado.Maioria.PrecisaSegundoTurno {
		resultado.SegundoTurno.Necessario = true
		resultado.SegundoTurno.EmpateIndefinido = apurado.EmpateIndefinido
		for _, chave := range apurado.SegundoTurno {
			resultado.SegundoTurno.Chapas = append(resultado.SegundoTurno.Chapas, apuracao.resumoDaChapa(chave))
		}
	}
	return resultado, nil
}

// resultadoConsulta aplica voting.CalcularApuracaoBorda e depois voting.AplicarCorteOrcamentario.
func resultadoConsulta(ctx context.Context, r store.Repositorios, v store.Votacao) (*ResultadoConsulta, error) {
	opcoes, err := r.Votacoes.ListarOpcoes(ctx, v.ID)
	if err != nil {
		return nil, err
	}
	cedulas, err := r.Votos.ListarCedulasRanking(ctx, v.ID)
	if err != nil {
		return nil, err
	}

	chaves := make([]string, len(opcoes))
	porChave := make(map[string]store.Opcao, len(opcoes))
	custos := make(map[string]int64, len(opcoes))
	for i, o := range opcoes {
		chaves[i] = idTexto(o.ID)
		porChave[chaves[i]] = o
		if o.CustoCentavos != nil {
			custos[chaves[i]] = *o.CustoCentavos
		}
	}
	votos := make([]voting.VotoBorda, len(cedulas))
	for i, preferencias := range cedulas {
		votos[i] = voting.VotoBorda{Preferencias: idsTexto(preferencias)}
	}

	ranking, err := voting.CalcularApuracaoBorda(chaves, votos)
	if err != nil {
		return nil, err
	}
	var orcamento int64
	if v.OrcamentoCentavos != nil {
		orcamento = *v.OrcamentoCentavos
	}
	corte, err := voting.AplicarCorteOrcamentario(ranking, custos, orcamento)
	if err != nil {
		return nil, err
	}

	resultado := &ResultadoConsulta{
		VotacaoID:                v.ID,
		Titulo:                   v.Titulo,
		Descricao:                v.Descricao,
		Cedulas:                  len(cedulas),
		OrcamentoCentavos:        corte.Orcamento,
		TotalSelecionadoCentavos: corte.TotalSelecionado,
		SaldoCentavos:            corte.Saldo,
		EmpateNaLinhaDeCorte:     corte.EmpateNaLinhaDeCorte,
		Ranking:                  make([]ItemRanking, len(corte.Projetos)),
	}
	for i, p := range corte.Projetos {
		opcao := porChave[p.Opcao]
		resultado.Ranking[i] = ItemRanking{
			Posicao:                i + 1,
			OpcaoID:                opcao.ID,
			Nome:                   opcao.Nome,
			Pontos:                 p.Pontos,
			PrimeirasPreferencias:  p.PrimeirasPreferencias,
			CustoCentavos:          p.Custo,
			CustoAcumuladoCentavos: p.CustoAcumulado,
			DentroDoOrcamento:      p.DentroDoOrcamento,
			Empatado:               p.Empatado,
		}
	}
	return resultado, nil
}
