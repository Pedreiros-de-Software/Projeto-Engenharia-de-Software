package service

import (
	"context"
	"strings"
	"time"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store"
)

const (
	maxTitulo    = 200
	maxDescricao = 2000
	maxNomeOpcao = 120
	minOpcoes    = 2
	maxOpcoes    = 20

	// maxCentavos (R$ 1 trilhão) mantém exata a soma de até maxOpcoes custos, tanto em int64
	// quanto nos números do JavaScript (abaixo de 2^53).
	maxCentavos = 100_000_000_000_000
)

type NovaOpcao struct {
	Nome          string
	Descricao     string
	CustoCentavos *int64 // obrigatório na consulta; não se aplica aos demais tipos
}

type NovaVotacao struct {
	Tipo              string
	Titulo            string
	Descricao         string
	Inicio            time.Time
	Fim               time.Time
	Opcoes            []NovaOpcao
	OrcamentoCentavos *int64  // obrigatório na consulta; não se aplica aos demais tipos
	Aptos             []int64 // eleição e consulta: ids de eleitores; vazio = todos os eleitores
}

// CriarVotacao valida e grava a votação, suas opções e seus aptos numa única transação.
func (s *Service) CriarVotacao(ctx context.Context, nova NovaVotacao) (int64, error) {
	nova = semEspacosNasPontas(nova)
	// O banco guarda microssegundos; a validação usa a mesma precisão.
	nova.Inicio, nova.Fim = nova.Inicio.Truncate(time.Microsecond), nova.Fim.Truncate(time.Microsecond)
	if err := validarNovaVotacao(nova); err != nil {
		return 0, err
	}

	var aptos []int64
	if nova.Tipo != TipoPauta {
		var err error
		if aptos, err = s.resolverAptos(ctx, nova.Aptos); err != nil {
			return 0, err
		}
	}

	var id int64
	err := store.EmTransacao(ctx, s.db, func(r store.Repositorios) error {
		var err error
		id, err = r.Votacoes.Inserir(ctx, store.Votacao{
			Tipo:              nova.Tipo,
			Titulo:            nova.Titulo,
			Descricao:         nova.Descricao,
			Inicio:            nova.Inicio,
			Fim:               nova.Fim,
			Turno:             1,
			OrcamentoCentavos: nova.OrcamentoCentavos,
		})
		if err != nil {
			return err
		}
		for i, o := range nova.Opcoes {
			opcao := store.Opcao{VotacaoID: id, Nome: o.Nome, Descricao: o.Descricao, CustoCentavos: o.CustoCentavos, Posicao: i + 1}
			if _, err := r.Votacoes.InserirOpcao(ctx, opcao); err != nil {
				return err
			}
		}
		for _, usuarioID := range aptos {
			if err := r.Votacoes.InserirApto(ctx, id, usuarioID); err != nil {
				return err
			}
		}
		return nil
	})
	return id, err
}

func semEspacosNasPontas(n NovaVotacao) NovaVotacao {
	n.Titulo = strings.TrimSpace(n.Titulo)
	n.Descricao = strings.TrimSpace(n.Descricao)
	opcoes := make([]NovaOpcao, len(n.Opcoes))
	for i, o := range n.Opcoes {
		opcoes[i] = NovaOpcao{Nome: strings.TrimSpace(o.Nome), Descricao: strings.TrimSpace(o.Descricao), CustoCentavos: o.CustoCentavos}
	}
	n.Opcoes = opcoes
	return n
}

func validarNovaVotacao(n NovaVotacao) error {
	switch {
	case n.Tipo != TipoPauta && n.Tipo != TipoEleicao && n.Tipo != TipoConsulta:
		return invalido("Tipo de votação inválido: use pauta, eleicao ou consulta.")
	case !tamanhoEntre(n.Titulo, 1, maxTitulo):
		return invalido("O título deve ter entre 1 e 200 caracteres.")
	case !tamanhoEntre(n.Descricao, 0, maxDescricao):
		return invalido("A descrição pode ter no máximo 2000 caracteres.")
	case n.Inicio.IsZero() || n.Fim.IsZero():
		return invalido("Informe o início e o fim da votação.")
	case !n.Fim.After(n.Inicio):
		return invalido("O fim da votação precisa ser depois do início.")
	case n.Tipo == TipoPauta:
		return validarPauta(n)
	default:
		return validarOpcoes(n)
	}
}

func validarPauta(n NovaVotacao) error {
	switch {
	case len(n.Opcoes) > 0:
		return invalido("A pauta não tem opções: o voto é a favor, contra ou abstenção.")
	case n.OrcamentoCentavos != nil:
		return invalido("Orçamento só se aplica à priorização de projetos.")
	case len(n.Aptos) > 0:
		return invalido("Na pauta votam as unidades com responsável: não informe aptos.")
	}
	return nil
}

// validarOpcoes cobre eleição (chapas, sem custo) e consulta (projetos com custo e orçamento).
func validarOpcoes(n NovaVotacao) error {
	if len(n.Opcoes) < minOpcoes || len(n.Opcoes) > maxOpcoes {
		return invalido("Cadastre de 2 a 20 opções.")
	}

	consulta := n.Tipo == TipoConsulta
	nomes := make(map[string]bool, len(n.Opcoes))
	for _, o := range n.Opcoes {
		switch {
		case !tamanhoEntre(o.Nome, 1, maxNomeOpcao):
			return invalido("Cada opção precisa de um nome com até 120 caracteres.")
		case !tamanhoEntre(o.Descricao, 0, maxDescricao):
			return invalido("A descrição de cada opção pode ter no máximo 2000 caracteres.")
		case nomes[o.Nome]:
			return invalido("Os nomes das opções não podem se repetir.")
		case consulta && (o.CustoCentavos == nil || *o.CustoCentavos < 0):
			return invalido("Informe o custo (zero ou mais) de cada projeto.")
		case consulta && *o.CustoCentavos > maxCentavos:
			return invalido("O custo de cada projeto pode ser de no máximo R$ 1 trilhão.")
		case !consulta && o.CustoCentavos != nil:
			return invalido("Custo só se aplica aos projetos da priorização.")
		}
		nomes[o.Nome] = true
	}

	switch {
	case consulta && (n.OrcamentoCentavos == nil || *n.OrcamentoCentavos < 0):
		return invalido("Informe o orçamento (zero ou mais) da priorização.")
	case consulta && *n.OrcamentoCentavos > maxCentavos:
		return invalido("O orçamento pode ser de no máximo R$ 1 trilhão.")
	case !consulta && n.OrcamentoCentavos != nil:
		return invalido("Orçamento só se aplica à priorização de projetos.")
	}
	return nil
}

// resolverAptos confere que cada id é de um eleitor (sem repetir); lista vazia significa todos os eleitores.
func (s *Service) resolverAptos(ctx context.Context, ids []int64) ([]int64, error) {
	usuarios, err := s.repo.Usuarios.Listar(ctx)
	if err != nil {
		return nil, err
	}
	eleitores := map[int64]bool{}
	todos := []int64{}
	for _, u := range usuarios {
		if u.Papel == PapelEleitor {
			eleitores[u.ID] = true
			todos = append(todos, u.ID)
		}
	}
	if len(ids) == 0 {
		return todos, nil
	}

	aptos := []int64{}
	incluidos := map[int64]bool{}
	for _, id := range ids {
		if !eleitores[id] {
			return nil, invalido("A lista de aptos contém usuário inexistente ou que não é eleitor.")
		}
		if !incluidos[id] {
			incluidos[id] = true
			aptos = append(aptos, id)
		}
	}
	return aptos, nil
}
