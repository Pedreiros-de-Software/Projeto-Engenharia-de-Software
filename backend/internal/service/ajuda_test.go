package service

import (
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store/storetest"
)

// instanteInicial é o "agora" de todos os cenários, até o teste avançar o relógio.
var instanteInicial = time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)

const senhaDeTeste = "senha-de-teste"

// iteracoesDeTeste mantém o PBKDF2 rápido nos testes; produção usa 210.000.
const iteracoesDeTeste = 1000

type relogioFalso struct {
	mu    sync.Mutex
	agora time.Time
}

func (r *relogioFalso) Agora() time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.agora
}

func (r *relogioFalso) Avancar(d time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agora = r.agora.Add(d)
}

type cenario struct {
	svc     *Service
	db      *sql.DB
	relogio *relogioFalso
}

func novoCenario(t *testing.T) *cenario {
	t.Helper()
	db := storetest.NovoBanco(t)
	relogio := &relogioFalso{agora: instanteInicial}
	return &cenario{
		svc:     Novo(db, Config{Relogio: relogio.Agora, IteracoesSenha: iteracoesDeTeste}),
		db:      db,
		relogio: relogio,
	}
}

func (c *cenario) usuario(t *testing.T, login, papel string) int64 {
	t.Helper()
	id, err := c.svc.CriarUsuario(t.Context(), "Nome "+login, login, senhaDeTeste, papel)
	if err != nil {
		t.Fatalf("criando usuário %s: %v", login, err)
	}
	return id
}

func (c *cenario) unidade(t *testing.T, identificador string, fracao float64, adimplente bool, responsavelID int64) int64 {
	t.Helper()
	id, err := c.svc.CriarUnidade(t.Context(), identificador, fracao, adimplente, &responsavelID)
	if err != nil {
		t.Fatalf("criando unidade %s: %v", identificador, err)
	}
	return id
}

// votacao cria uma votação aberta de uma hora antes até uma hora depois de agora,
// a menos que a janela seja informada.
func (c *cenario) votacao(t *testing.T, nova NovaVotacao) int64 {
	t.Helper()
	if nova.Inicio.IsZero() {
		agora := c.relogio.Agora()
		nova.Inicio, nova.Fim = agora.Add(-time.Hour), agora.Add(time.Hour)
	}
	if nova.Titulo == "" {
		nova.Titulo = "Votação de teste"
	}
	id, err := c.svc.CriarVotacao(t.Context(), nova)
	if err != nil {
		t.Fatalf("criando votação: %v", err)
	}
	return id
}

// opcoes devolve os ids das opções da votação, na ordem de cadastro.
func (c *cenario) opcoes(t *testing.T, votacaoID int64) []int64 {
	t.Helper()
	detalhe, err := c.svc.Detalhar(t.Context(), 0, votacaoID)
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]int64, len(detalhe.Opcoes))
	for i, o := range detalhe.Opcoes {
		ids[i] = o.ID
	}
	return ids
}

func (c *cenario) votar(t *testing.T, usuarioID, votacaoID int64, voto Voto) {
	t.Helper()
	if err := c.svc.Votar(t.Context(), usuarioID, votacaoID, voto); err != nil {
		t.Fatalf("votando: %v", err)
	}
}

func (c *cenario) encerrar(t *testing.T, votacaoID int64) {
	t.Helper()
	if err := c.svc.Encerrar(t.Context(), votacaoID); err != nil {
		t.Fatalf("encerrando: %v", err)
	}
}

// contar executa um SELECT count(*) direto no banco, para conferir o que foi gravado.
func (c *cenario) contar(t *testing.T, consulta string, args ...any) int {
	t.Helper()
	var total int
	if err := c.db.QueryRow(consulta, args...).Scan(&total); err != nil {
		t.Fatal(err)
	}
	return total
}

func opcoesNomeadas(nomes ...string) []NovaOpcao {
	opcoes := make([]NovaOpcao, len(nomes))
	for i, nome := range nomes {
		opcoes[i] = NovaOpcao{Nome: nome, Descricao: "Descrição de " + nome}
	}
	return opcoes
}

func projetos(custos ...int64) []NovaOpcao {
	opcoes := make([]NovaOpcao, len(custos))
	for i, custo := range custos {
		opcoes[i] = NovaOpcao{Nome: string(rune('A' + i)), CustoCentavos: ptr(custo)}
	}
	return opcoes
}

func ptr[T any](v T) *T { return &v }

// conferirErro compara erros sentinela com errors.Is; para ErroValidacao basta o tipo.
func conferirErro(t *testing.T, obtido, esperado error) {
	t.Helper()
	var validacao *ErroValidacao
	if errors.As(esperado, &validacao) {
		if !errors.As(obtido, &validacao) {
			t.Errorf("esperava ErroValidacao, obteve %v", obtido)
		}
		return
	}
	if !errors.Is(obtido, esperado) {
		t.Errorf("erro: esperado %v, obtido %v", esperado, obtido)
	}
}

// errValidacao é o valor usado nas tabelas de teste para "qualquer ErroValidacao".
var errValidacao error = &ErroValidacao{}
