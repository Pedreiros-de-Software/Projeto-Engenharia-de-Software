package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

//go:embed migrations/*.sql
var arquivosMigracao embed.FS

// chaveMigracao identifica o pg_advisory_lock que impede duas instâncias de migrarem ao mesmo tempo.
const chaveMigracao = 2026_0001

// Migrar aplica, em ordem lexicográfica e cada um em sua transação, os arquivos de
// migrations/ ainda não registrados na tabela schema_migrations.
func Migrar(ctx context.Context, db *sql.DB) error {
	// O advisory lock pertence à sessão, então todo o processo usa uma única conexão.
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("obtendo conexao para migrar: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", chaveMigracao); err != nil {
		return fmt.Errorf("travando migracoes: %w", err)
	}
	defer func() {
		_, _ = conn.ExecContext(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock($1)", chaveMigracao)
	}()

	if _, err := conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		versao      TEXT PRIMARY KEY,
		aplicada_em TIMESTAMPTZ NOT NULL
	)`); err != nil {
		return fmt.Errorf("criando schema_migrations: %w", err)
	}

	versoes, err := listarMigracoes()
	if err != nil {
		return err
	}
	for _, versao := range versoes {
		if err := aplicarSeNecessario(ctx, conn, versao); err != nil {
			return fmt.Errorf("migracao %s: %w", versao, err)
		}
	}
	return nil
}

// listarMigracoes devolve os nomes dos arquivos .sql; fs.ReadDir já os entrega ordenados.
func listarMigracoes() ([]string, error) {
	entradas, err := fs.ReadDir(arquivosMigracao, "migrations")
	if err != nil {
		return nil, fmt.Errorf("lendo migracoes: %w", err)
	}
	var versoes []string
	for _, entrada := range entradas {
		if strings.HasSuffix(entrada.Name(), ".sql") {
			versoes = append(versoes, entrada.Name())
		}
	}
	return versoes, nil
}

func aplicarSeNecessario(ctx context.Context, conn *sql.Conn, versao string) error {
	var aplicada bool
	err := conn.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE versao = $1)", versao).Scan(&aplicada)
	if err != nil || aplicada {
		return err
	}

	conteudo, err := fs.ReadFile(arquivosMigracao, path.Join("migrations", versao))
	if err != nil {
		return err
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, string(conteudo)); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO schema_migrations (versao, aplicada_em) VALUES ($1, now())", versao); err != nil {
		return err
	}
	return tx.Commit()
}
