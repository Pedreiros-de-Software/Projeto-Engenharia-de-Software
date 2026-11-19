package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/config"
	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/database"
	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/seed"
	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/server"
	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/service"
	"github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend/internal/store"
)

const tempoParaDesligar = 10 * time.Second

func main() {
	if err := executar(); err != nil {
		log.Fatal(err)
	}
}

// executar monta a aplicação: config → banco → migrações → (seed) → service → rotas → servidor HTTP.
func executar() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("carregando configuracao: %w", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("conectando ao banco: %w", err)
	}
	defer fecharBanco(db)

	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()

	if err := store.Migrar(ctx, db); err != nil {
		return fmt.Errorf("aplicando migracoes: %w", err)
	}

	svc := service.Novo(db, service.Config{})
	if cfg.SeedDemo {
		// Um sinal no meio do seed não o interrompe: ele termina e depois servir encerra.
		criado, err := seed.Executar(context.WithoutCancel(ctx), svc)
		if err != nil {
			return err
		}
		if criado {
			log.Print("seed de demonstracao criado")
		} else {
			log.Print("banco ja tem usuarios: seed de demonstracao ignorado")
		}
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.NewRouter(db, svc),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return servir(ctx, srv)
}

// servir atende até receber SIGINT ou SIGTERM e então espera as requisições em andamento terminarem.
func servir(ctx context.Context, srv *http.Server) error {
	erros := make(chan error, 1)
	go func() {
		log.Printf("servidor ouvindo em %s", srv.Addr)
		erros <- srv.ListenAndServe()
	}()

	select {
	case err := <-erros:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("servidor HTTP: %w", err)
		}
		return nil
	case <-ctx.Done():
	}

	log.Print("sinal recebido: desligando o servidor")
	ctxDesligar, cancelar := context.WithTimeout(context.Background(), tempoParaDesligar)
	defer cancelar()
	return srv.Shutdown(ctxDesligar)
}

func fecharBanco(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("fechando conexao com o banco: %v", err)
	}
}
