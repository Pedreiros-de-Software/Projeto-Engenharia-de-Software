package main

import (
	"log"
	"net/http"

	"github.com/onicolasrosa/Projeto-Engenharia-de-Software/backend/internal/config"
	"github.com/onicolasrosa/Projeto-Engenharia-de-Software/backend/internal/database"
	"github.com/onicolasrosa/Projeto-Engenharia-de-Software/backend/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer db.Close()

	router := server.NewRouter(db)

	log.Printf("server listening on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal(err)
	}
}
