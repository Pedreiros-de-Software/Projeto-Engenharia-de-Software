// Package seed cria os dados de demonstração chamando os mesmos casos de uso usados pela API.
// As senhas daqui servem apenas para demonstração.
package seed

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/service"
	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/voting"
)

const (
	senhaAdmin    = "admin123"
	senhaEleitor  = "consensu"
	duracaoAberta = 7 * 24 * time.Hour
)

type pessoa struct{ login, nome string }

var condominos = []pessoa{
	{"maria", "Maria Souza"}, {"joao", "João Lima"}, {"lucia", "Lúcia Alves"},
	{"pedro", "Pedro Rocha"}, {"rafael", "Rafael Dias"}, {"sofia", "Sofia Prado"},
}

var estudantes = []pessoa{
	{"ana", "Ana Beatriz"}, {"bruno", "Bruno Costa"}, {"carla", "Carla Mendes"}, {"diego", "Diego Nunes"},
	{"elisa", "Elisa Ramos"}, {"fabio", "Fábio Teixeira"}, {"gabi", "Gabriela Pires"}, {"hugo", "Hugo Martins"},
}

// unidades do "Edifício Solar": fração ideal = área em m² (total 500).
var unidades = []struct {
	identificador string
	fracao        float64
	responsavel   string
	adimplente    bool
}{
	{"Apto 101", 60, "maria", true},
	{"Apto 102", 60, "joao", true},
	{"Apto 201", 80, "lucia", true},
	{"Apto 202", 80, "pedro", true},
	{"Apto 301", 100, "rafael", true},
	{"Apto 402", 120, "sofia", false},
}

var chapas = []service.NovaOpcao{
	{Nome: "Chapa 1 — Movimenta", Descricao: "Ana / Victor / Amanda"},
	{Nome: "Chapa 2 — Renovação", Descricao: "João / Ana / Marcela"},
	{Nome: "Chapa 3 — Integração", Descricao: "Paulo / Rita / Caio"},
}

var projetos = []service.NovaOpcao{
	{Nome: "Impressoras 3D e CNC", Descricao: "Equipamentos de prototipagem rápida.", CustoCentavos: centavos(35000_00)},
	{Nome: "Kit Arduino e Robótica", Descricao: "Kits para oficinas de eletrônica e robótica.", CustoCentavos: centavos(12000_00)},
	{Nome: "Servidor GPU para IA", Descricao: "Servidor para treinar modelos de aprendizado de máquina.", CustoCentavos: centavos(48000_00)},
	{Nome: "Mobiliário colaborativo", Descricao: "Mesas e quadros para trabalho em grupo.", CustoCentavos: centavos(9000_00)},
}

func centavos(v int64) *int64 { return &v }

// ErrSeedIncompleto indica um banco com usuários e sem votações: um seed anterior parou no meio.
var ErrSeedIncompleto = errors.New("seed de demonstracao incompleto: ha usuarios mas nenhuma votacao; " +
	"recrie o banco (docker compose down -v) e suba de novo")

// Executar cria o cenário de demonstração se o banco ainda não tiver usuários.
// Devolve false, sem alterar nada, quando já existem usuários, e ErrSeedIncompleto
// quando eles existem sem nenhuma votação (a API não cria usuários; só o seed).
func Executar(ctx context.Context, svc *service.Service) (bool, error) {
	existem, err := svc.ExistemUsuarios(ctx)
	if err != nil {
		return false, err
	}
	if existem {
		return false, conferirSeedCompleto(ctx, svc)
	}

	agora := svc.Agora()
	s := &semeador{
		ctx:      ctx,
		svc:      svc,
		usuarios: map[string]int64{},
		inicio:   agora.Add(-time.Hour),
		fim:      agora.Add(duracaoAberta),
	}
	etapas := []func() error{
		s.criarUsuarios,
		s.criarUnidades,
		s.criarPautaEncerrada,
		s.criarPautaAberta,
		s.criarEleicaoEncerrada,
		s.criarConsultaAberta,
	}
	for _, etapa := range etapas {
		if err := etapa(); err != nil {
			return false, fmt.Errorf("seed de demonstracao: %w", err)
		}
	}
	return true, nil
}

func conferirSeedCompleto(ctx context.Context, svc *service.Service) error {
	existem, err := svc.ExistemVotacoes(ctx)
	if err != nil {
		return err
	}
	if !existem {
		return ErrSeedIncompleto
	}
	return nil
}

type semeador struct {
	ctx      context.Context
	svc      *service.Service
	usuarios map[string]int64 // id por login
	inicio   time.Time
	fim      time.Time
}

func (s *semeador) criarUsuarios() error {
	id, err := s.svc.CriarUsuario(s.ctx, "Administração", "admin", senhaAdmin, service.PapelAdmin)
	if err != nil {
		return fmt.Errorf("criando admin: %w", err)
	}
	s.usuarios["admin"] = id

	for _, p := range slices.Concat(condominos, estudantes) {
		id, err := s.svc.CriarUsuario(s.ctx, p.nome, p.login, senhaEleitor, service.PapelEleitor)
		if err != nil {
			return fmt.Errorf("criando usuario %s: %w", p.login, err)
		}
		s.usuarios[p.login] = id
	}
	return nil
}

func (s *semeador) criarUnidades() error {
	for _, u := range unidades {
		responsavel := s.usuarios[u.responsavel]
		if _, err := s.svc.CriarUnidade(s.ctx, u.identificador, u.fracao, u.adimplente, &responsavel); err != nil {
			return fmt.Errorf("criando unidade %s: %w", u.identificador, err)
		}
	}
	return nil
}

// criarPautaEncerrada: 300 de 500 a favor (60%) não atinge 2/3, então a pauta é reprovada.
func (s *semeador) criarPautaEncerrada() error {
	id, err := s.criarVotacao(service.NovaVotacao{
		Tipo:      service.TipoPauta,
		Titulo:    "Pauta #01 — Troca do sistema de câmeras",
		Descricao: "Substituição das câmeras analógicas por um sistema digital com gravação em nuvem.",
	})
	if err != nil {
		return err
	}
	votos := []struct{ login, escolha string }{
		{"maria", voting.OpcaoFavor}, {"joao", voting.OpcaoFavor}, {"lucia", voting.OpcaoFavor},
		{"rafael", voting.OpcaoFavor}, {"pedro", voting.OpcaoContra},
	}
	for _, v := range votos {
		if err := s.votar(v.login, id, service.Voto{Escolha: v.escolha}); err != nil {
			return err
		}
	}
	return s.svc.Encerrar(s.ctx, id)
}

// criarPautaAberta: já votaram Apto 301 e Apto 202; na demonstração, mais três unidades votam.
func (s *semeador) criarPautaAberta() error {
	id, err := s.criarVotacao(service.NovaVotacao{
		Tipo:      service.TipoPauta,
		Titulo:    "Pauta #02 — Aprovação do orçamento para reforma da portaria (R$ 45.000,00)",
		Descricao: "Reforma da portaria com rateio do custo pela fração ideal de cada unidade.",
	})
	if err != nil {
		return err
	}
	for _, login := range []string{"rafael", "pedro"} {
		if err := s.votar(login, id, service.Voto{Escolha: voting.OpcaoFavor}); err != nil {
			return err
		}
	}
	return nil
}

// criarEleicaoEncerrada: Chapa 1 tem 3 de 6 válidos (50%), sem maioria absoluta: pede 2º turno.
func (s *semeador) criarEleicaoEncerrada() error {
	aptos := make([]int64, len(estudantes))
	for i, e := range estudantes {
		aptos[i] = s.usuarios[e.login]
	}
	id, err := s.criarVotacao(service.NovaVotacao{
		Tipo:      service.TipoEleicao,
		Titulo:    "Eleição Diretoria DCE 2026",
		Descricao: "Eleição da diretoria do Diretório Central dos Estudantes. Voto secreto.",
		Opcoes:    chapas,
		Aptos:     aptos,
	})
	if err != nil {
		return err
	}
	chapaPorNome, err := s.opcoesPorNome(id)
	if err != nil {
		return err
	}

	votos := []struct{ login, escolha string }{
		{"ana", chapas[0].Nome}, {"bruno", chapas[0].Nome}, {"carla", chapas[0].Nome},
		{"diego", chapas[1].Nome}, {"elisa", chapas[1].Nome}, {"fabio", chapas[2].Nome},
		{"gabi", voting.OpcaoBranco},
	}
	for _, v := range votos {
		voto := service.Voto{Escolha: v.escolha}
		if opcaoID, ehChapa := chapaPorNome[v.escolha]; ehChapa {
			voto = service.Voto{Escolha: service.EscolhaChapa, OpcaoID: &opcaoID}
		}
		if err := s.votar(v.login, id, voto); err != nil {
			return err
		}
	}
	return s.svc.Encerrar(s.ctx, id)
}

// criarConsultaAberta: cinco cédulas já registradas; todos os eleitores são aptos.
func (s *semeador) criarConsultaAberta() error {
	id, err := s.criarVotacao(service.NovaVotacao{
		Tipo:              service.TipoConsulta,
		Titulo:            "Orçamento participativo — Laboratório de Inovação",
		Descricao:         "Ordene os projetos do mais ao menos prioritário. Os mais bem colocados são financiados até o limite do orçamento.",
		Opcoes:            projetos,
		OrcamentoCentavos: centavos(60000_00),
	})
	if err != nil {
		return err
	}
	projeto, err := s.opcoesPorNome(id)
	if err != nil {
		return err
	}
	impressoras, arduino := projeto[projetos[0].Nome], projeto[projetos[1].Nome]
	gpu, mobiliario := projeto[projetos[2].Nome], projeto[projetos[3].Nome]

	cedulas := []struct {
		login        string
		preferencias []int64
	}{
		{"ana", []int64{impressoras, arduino, gpu, mobiliario}},
		{"bruno", []int64{gpu, impressoras, arduino, mobiliario}},
		{"carla", []int64{arduino, impressoras, mobiliario, gpu}},
		{"maria", []int64{impressoras, mobiliario, arduino, gpu}},
		{"joao", []int64{gpu, arduino, impressoras, mobiliario}},
	}
	for _, c := range cedulas {
		if err := s.votar(c.login, id, service.Voto{Preferencias: c.preferencias}); err != nil {
			return err
		}
	}
	return nil
}

// criarVotacao cria a votação aberta desde uma hora atrás e por sete dias.
func (s *semeador) criarVotacao(nova service.NovaVotacao) (int64, error) {
	nova.Inicio, nova.Fim = s.inicio, s.fim
	id, err := s.svc.CriarVotacao(s.ctx, nova)
	if err != nil {
		return 0, fmt.Errorf("criando %q: %w", nova.Titulo, err)
	}
	return id, nil
}

func (s *semeador) votar(login string, votacaoID int64, voto service.Voto) error {
	if err := s.svc.Votar(s.ctx, s.usuarios[login], votacaoID, voto); err != nil {
		return fmt.Errorf("registrando voto de %s na votacao %d: %w", login, votacaoID, err)
	}
	return nil
}

func (s *semeador) opcoesPorNome(votacaoID int64) (map[string]int64, error) {
	detalhe, err := s.svc.Detalhar(s.ctx, s.usuarios["admin"], votacaoID)
	if err != nil {
		return nil, err
	}
	ids := make(map[string]int64, len(detalhe.Opcoes))
	for _, o := range detalhe.Opcoes {
		ids[o.Nome] = o.ID
	}
	return ids, nil
}
