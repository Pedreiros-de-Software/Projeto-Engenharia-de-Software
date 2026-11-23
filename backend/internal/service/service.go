// Package service implementa os casos de uso da aplicação: orquestra o store (persistência)
// e o voting (regras de domínio), controla as transações e recebe o relógio injetado.
package service

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store"
	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/voting"
)

const (
	PapelAdmin   = "admin"
	PapelEleitor = "eleitor"

	TipoPauta    = "pauta"
	TipoEleicao  = "eleicao"
	TipoConsulta = "consulta"

	StatusNaoIniciada = "nao_iniciada"
	StatusAberta      = "aberta"
	StatusEncerrada   = "encerrada"

	// EscolhaChapa é o voto de eleição em uma chapa; branco e nulo usam voting.OpcaoBranco e voting.OpcaoNulo.
	EscolhaChapa = "chapa"
)

const (
	iteracoesSenhaProducao    = 210_000
	validadeSessao            = 12 * time.Hour
	duracaoPadraoSegundoTurno = 24 * time.Hour
	limiteTentativasLogin     = 10 // falhas por IP e login
	limiteTentativasPorIP     = 60 // falhas por IP, somando todos os logins
	janelaTentativasLogin     = time.Minute
)

var (
	ErrNaoAutenticado           = errors.New("sessao ausente, invalida ou expirada")
	ErrCredenciaisInvalidas     = errors.New("login ou senha invalidos")
	ErrMuitasTentativas         = errors.New("muitas tentativas de login")
	ErrNaoEncontrado            = errors.New("recurso nao encontrado")
	ErrNaoApto                  = errors.New("usuario nao apto a votar nesta votacao")
	ErrInadimplente             = errors.New("unidade inadimplente")
	ErrVotoDuplicado            = errors.New("voto ja registrado")
	ErrVotacaoNaoAberta         = errors.New("votacao nao esta aberta")
	ErrVotacaoNaoEncerrada      = errors.New("votacao ainda nao encerrada")
	ErrCedulaInvalida           = errors.New("cedula invalida")
	ErrOpcaoDesconhecida        = errors.New("opcao desconhecida")
	ErrSegundoTurnoIndisponivel = errors.New("segundo turno indisponivel")
)

// ErroValidacao indica um dado de entrada inválido; Mensagem é exibida ao usuário.
type ErroValidacao struct{ Mensagem string }

func (e *ErroValidacao) Error() string { return e.Mensagem }

func invalido(mensagem string) error { return &ErroValidacao{Mensagem: mensagem} }

type Config struct {
	Relogio        func() time.Time // padrão: time.Now
	IteracoesSenha int              // iterações do PBKDF2 para novas senhas; padrão: 210.000
}

type Service struct {
	db                 *sql.DB
	repo               store.Repositorios
	relogio            func() time.Time
	iteracoes          int
	tentativasPorConta *limitador
	tentativasPorIP    *limitador
}

func Novo(db *sql.DB, cfg Config) *Service {
	s := &Service{
		db:        db,
		repo:      store.Novo(db),
		relogio:   cfg.Relogio,
		iteracoes: cfg.IteracoesSenha,

		tentativasPorConta: novoLimitador(limiteTentativasLogin, janelaTentativasLogin),
		tentativasPorIP:    novoLimitador(limiteTentativasPorIP, janelaTentativasLogin),
	}
	if s.relogio == nil {
		s.relogio = time.Now
	}
	if s.iteracoes <= 0 {
		s.iteracoes = iteracoesSenhaProducao
	}
	return s
}

// Agora devolve o instante atual segundo o relógio do serviço.
func (s *Service) Agora() time.Time {
	return s.relogio()
}

// statusEm traduz o estado do domínio para o status exposto pela aplicação.
func statusEm(v store.Votacao, agora time.Time) (string, error) {
	estado, err := voting.StatusVotacao(v.Inicio, v.Fim, agora)
	if err != nil {
		return "", err
	}
	switch estado {
	case voting.VotacaoNaoIniciada:
		return StatusNaoIniciada, nil
	case voting.VotacaoAberta:
		return StatusAberta, nil
	default:
		return StatusEncerrada, nil
	}
}

// buscarVotacao traduz "não encontrado" do store para o erro da aplicação.
func buscarVotacao(ctx context.Context, r store.Repositorios, id int64) (store.Votacao, error) {
	v, err := r.Votacoes.Buscar(ctx, id)
	if errors.Is(err, store.ErrNaoEncontrado) {
		return v, ErrNaoEncontrado
	}
	return v, err
}

// idTexto converte ids do banco nas chaves de texto usadas pelas funções do domínio.
func idTexto(id int64) string {
	return strconv.FormatInt(id, 10)
}

func idsTexto(ids []int64) []string {
	textos := make([]string, len(ids))
	for i, id := range ids {
		textos[i] = idTexto(id)
	}
	return textos
}

// tamanhoEntre conta caracteres (não bytes), para acentos não pesarem no limite, e recusa
// o caractere NUL, que o Postgres não aceita em colunas de texto.
func tamanhoEntre(texto string, minimo, maximo int) bool {
	n := utf8.RuneCountInString(texto)
	return n >= minimo && n <= maximo && !temNUL(texto)
}

func temNUL(texto string) bool {
	return strings.ContainsRune(texto, 0)
}

func percentual(parte, total float64) float64 {
	if total == 0 {
		return 0
	}
	return parte / total * 100
}
