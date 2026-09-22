# Projeto-Engenharia-de-Software

- Nicolas Fernandes dos Santos Rosa - RA186126

## Backend

API em Go, com PostgreSQL. Estrutura:

```
backend/
├── cmd/api/          # ponto de entrada (main.go)
└── internal/
    ├── config/       # leitura de variáveis de ambiente
    ├── database/     # conexão com o Postgres
    └── server/       # rotas HTTP
```

### Rodando localmente

1. Suba o banco:

   ```bash
   docker compose up -d
   ```

2. Configure as variáveis de ambiente (copie o exemplo):

   ```bash
   cd backend
   cp .env.example .env
   export $(cat .env | xargs)
   ```

3. Rode o servidor:

   ```bash
   go run ./cmd/api
   ```

4. Confirme que está no ar:

   ```bash
   curl http://localhost:8080/health
   ```

