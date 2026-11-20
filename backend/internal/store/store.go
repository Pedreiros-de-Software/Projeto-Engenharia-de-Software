package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNaoEncontrado = errors.New("registro nao encontrado")
	ErrDuplicado     = errors.New("registro duplicado")
)

// codigoViolacaoUnica é o SQLSTATE do Postgres para violação de UNIQUE ou PRIMARY KEY.
const codigoViolacaoUnica = "23505"

// executor é o que os repositórios usam do banco; *sql.DB e *sql.Tx o satisfazem.
type executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Repositorios agrupa os repositórios de cada agregado sobre o mesmo banco ou transação.
type Repositorios struct {
	Usuarios Usuarios
	Sessoes  Sessoes
	Unidades Unidades
	Votacoes Votacoes
	Votos    Votos
}

// Novo devolve os repositórios operando diretamente sobre o banco.
func Novo(db *sql.DB) Repositorios {
	return sobre(db)
}

func sobre(db executor) Repositorios {
	return Repositorios{
		Usuarios: Usuarios{db: db},
		Sessoes:  Sessoes{db: db},
		Unidades: Unidades{db: db},
		Votacoes: Votacoes{db: db},
		Votos:    Votos{db: db},
	}
}

// EmTransacao executa fn com repositórios ligados a uma transação: confirma se fn
// devolver nil e desfaz tudo se fn devolver erro.
func EmTransacao(ctx context.Context, db *sql.DB, fn func(Repositorios) error) error {
	return emTransacao(ctx, db, nil, fn)
}

// EmLeitura executa fn numa transação somente leitura REPEATABLE READ: todas as consultas
// de fn enxergam o mesmo retrato do banco, mesmo que outras transações confirmem no meio.
func EmLeitura(ctx context.Context, db *sql.DB, fn func(Repositorios) error) error {
	return emTransacao(ctx, db, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true}, fn)
}

func emTransacao(ctx context.Context, db *sql.DB, opcoes *sql.TxOptions, fn func(Repositorios) error) error {
	tx, err := db.BeginTx(ctx, opcoes)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if err := fn(sobre(tx)); err != nil {
		return err
	}
	return tx.Commit()
}

// linha é a parte comum de *sql.Row e *sql.Rows usada para ler uma linha.
type linha interface {
	Scan(dest ...any) error
}

// coletar percorre todas as linhas de uma consulta convertendo cada uma com ler.
func coletar[T any](rows *sql.Rows, ler func(linha) (T, error)) ([]T, error) {
	defer func() { _ = rows.Close() }()

	itens := []T{}
	for rows.Next() {
		item, err := ler(rows)
		if err != nil {
			return nil, err
		}
		itens = append(itens, item)
	}
	return itens, rows.Err()
}

// traduzirErro converte erros do driver nos erros sentinela do pacote.
func traduzirErro(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNaoEncontrado
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == codigoViolacaoUnica {
		return ErrDuplicado
	}
	return err
}

// exigirUmaLinha devolve ErrNaoEncontrado quando um UPDATE não alterou nenhuma linha.
func exigirUmaLinha(resultado sql.Result) error {
	alteradas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if alteradas == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

func lerID(l linha) (int64, error) {
	var id int64
	err := l.Scan(&id)
	return id, err
}
