# Consensu

Plataforma web de votação para decisões coletivas com regras diferentes. A mesma aplicação atende três contextos:
a **assembleia de condomínio**, em que a pauta é votada a favor, contra ou abstenção, cada voto vale a fração ideal
da unidade, unidade inadimplente não vota e aprovar exige 2/3 do peso total; a **eleição de Centro Acadêmico**, com
voto secreto entre chapas, branco e nulo, maioria absoluta dos válidos, 2º turno entre as duas mais votadas e
boletim público; e a **priorização de projetos**, em que cada participante ordena todos os projetos, a apuração é
pelo método Borda e o ranking mostra quais cabem no orçamento. Projeto final de Engenharia de Software da equipe
**Pedreiros de Software**.

## Equipe

| Integrante | RA | GitHub |
|---|---|---|
| Nicolas Fernandes dos Santos Rosa | 186126 | `onicolasrosa` |
| Felipe Pires Araujo | 186697 | `felipepires-nu` |
| Pedro Romasanta Rosa | 206599 | `romapedro` |
| André Lucas Loubet Souza | 253333 | `andluca` |
| Matheus Henrique Aliberto de Francisco | 239930 | `MatheusHAF` |

## Funcionalidades

Rastreabilidade completa, critério a critério, em [docs/requisitos.md](docs/requisitos.md).

| Contexto | História | O que o sistema faz |
|---|---|---|
| Assembleia de condomínio (#14) | #17 Voto ponderado | o condômino vê a unidade, a fração ideal e o percentual do total antes de confirmar; vota aprovar, rejeitar ou abster-se; um voto por unidade |
| | #18 Inadimplência | unidade inadimplente vê o motivo e não tem botões de voto; a API também recusa (`403 inadimplente`); a administração altera a adimplência na tela Unidades |
| | #19 Quórum qualificado | aprovada só com ≥ 2/3 do peso total cadastrado, incluindo quem não votou; o relatório mostra pesos, limiar, percentual e unidades votantes |
| Eleição de Centro Acadêmico (#15) | #20 Voto secreto | só eleitores aptos, um voto por turno, chapas com descrição, branco e nulo; quem votou e o que foi votado ficam em tabelas separadas; a resposta não repete a escolha |
| | #21 Maioria absoluta e 2º turno | válidos excluem branco e nulo; vence quem tem mais da metade; sem maioria, o boletim indica as duas mais votadas e a administração abre o 2º turno; empate pelo 2º lugar é sinalizado para a comissão |
| | #22 Boletim público | depois do encerramento, sem login: aptos, participantes, abstenções, votos e percentuais por chapa, brancos, nulos e a base dos percentuais |
| Priorização de projetos (#16) | #23 Ordenação | o participante ordena todos os projetos; o envio só é liberado com a lista completa; o backend recusa cédula incompleta, repetida ou com projeto desconhecido |
| | #24 Borda | N pontos para o 1º lugar até 1 para o último; empate desfeito por primeiras preferências; empate persistente é marcado |
| | #25 Ranking orçamentário | pontos, custo, custo acumulado, linha de corte do orçamento, total selecionado e saldo |

A administração cria votações dos três tipos (janela, opções, orçamento e eleitores aptos), encerra votações
abertas com confirmação e abre o 2º turno. Valores em reais aceitam o formato brasileiro (`35.000,50`).

## Executar com Docker

Requisito: Docker com Compose v2.17 ou mais novo.

```bash
cp .env.example .env       # só na primeira vez
docker compose up --build
```

As credenciais do Postgres local ficam no `.env` da raiz, que o git ignora, copiado de `.env.example`. O
`docker-compose.yml` não tem senha embutida e para com uma mensagem clara se o `.env` estiver faltando.

Ou, em segundo plano, esperando os três serviços ficarem saudáveis:

```bash
docker compose up -d --build --wait
docker compose ps          # postgres, api e web como (healthy)
```

| Endereço | O que é |
|---|---|
| http://localhost:3000 | o site (nginx servindo a SPA) |
| http://localhost:3000/api/... | a API pela mesma origem do site |
| http://localhost:3000/health | saúde da API e do banco (`{"status":"ok"}`) |
| http://localhost:8080 | a API direto (só a partir da própria máquina) |
| `localhost:5432` | Postgres, usuário/senha/banco do `.env` (`consensu` no `.env.example`; só a partir da própria máquina) |

Portas ocupadas? Troque-as: `WEB_PORT=3001 API_PORT=8081 POSTGRES_PORT=5433 docker compose up --build` (as mesmas
variáveis podem ir no `.env` da raiz).

Na primeira subida, com o banco vazio, a API aplica as migrações e cria os dados de demonstração.

### Usuários de demonstração

As senhas abaixo existem só para demonstração e só são criadas com `SEED_DEMO=true` (o `docker-compose.yml` liga
essa opção).

| Login | Senha | Papel | Perfil no cenário |
|---|---|---|---|
| `admin` | `admin123` | admin | Administração: cria, encerra, abre 2º turno, altera unidades (não vota) |
| `maria` | `consensu` | eleitor | Apto 101, fração 60 (12%) |
| `joao` | `consensu` | eleitor | Apto 102, fração 60 (12%) |
| `lucia` | `consensu` | eleitor | Apto 201, fração 80 (16%) |
| `pedro` | `consensu` | eleitor | Apto 202, fração 80 (16%); já votou na Pauta #02 |
| `rafael` | `consensu` | eleitor | Apto 301, fração 100 (20%); já votou na Pauta #02 |
| `sofia` | `consensu` | eleitor | Apto 402, fração 120 (24%), **inadimplente** |
| `ana`, `bruno`, `carla`, `diego`, `elisa`, `fabio`, `gabi`, `hugo` | `consensu` | eleitor | estudantes aptos na eleição |

Todos os eleitores são aptos na consulta; `ana`, `bruno`, `carla`, `maria` e `joao` já enviaram cédula.

| Id | Votação | Estado inicial | O que mostra |
|---|---|---|---|
| 1 | Pauta #01 — Troca do sistema de câmeras | encerrada | reprovada com 300 de 500 (60%): maioria simples não basta |
| 2 | Pauta #02 — Aprovação do orçamento para reforma da portaria (R$ 45.000,00) | aberta por 7 dias | 180 de 500 a favor; com maria, joão e lúcia chega a 380 ≥ 333,33 e é aprovada; sofia é bloqueada |
| 3 | Eleição Diretoria DCE 2026 | encerrada (1º turno) | Chapa 1 com 3 de 6 válidos (50%): 2º turno entre Chapa 1 e Chapa 2 |
| 4 | Orçamento participativo — Laboratório de Inovação | aberta por 7 dias | 4 projetos, orçamento de R$ 60.000,00, 5 cédulas registradas |

Roteiro curto de demonstração: maria, joão e lúcia aprovam a Pauta #02; sofia tenta e é bloqueada; o admin encerra
a pauta e o resultado mostra APROVADA. No boletim da eleição (público), o admin abre o 2º turno; ana e bruno votam;
o admin encerra e o boletim mostra a vencedora. lúcia ordena os projetos da consulta; o admin encerra e o ranking
mostra a linha de corte.

### Parar e recomeçar

```bash
docker compose down        # para, mantendo os dados
docker compose down -v     # para e apaga o banco; a próxima subida recria os dados de demonstração
```

O seed só roda com o banco sem usuários, então reiniciar não restaura o cenário: use `docker compose down -v`. As
votações abertas do seed duram 7 dias a partir da primeira subida. Se a API recusar subir com a mensagem
`seed de demonstracao incompleto`, o banco ficou pela metade: rode `docker compose down -v` e suba de novo.

## Modo de desenvolvimento

| Ferramenta | Versão |
|---|---|
| Go | 1.26.1 (`backend/go.mod`) |
| Node.js | `^20.19.0 \|\| >=22.12.0` (`frontend/package.json`); recomendado 22, como em `frontend/.nvmrc` |
| Docker | com Compose v2.17+ (para o Postgres e o sistema completo) |
| `curl` e `jq` | só para `scripts/e2e-api.sh` |

1. Banco (se o sistema completo estiver no ar, pare antes a API e o site: `docker compose stop api web`):

   ```bash
   cp .env.example .env       # só na primeira vez, na raiz do repositório
   docker compose up -d postgres
   ```

2. API, em `backend/`. A API lê só variáveis de ambiente (não carrega `.env` sozinha), por isso o `.env` é
   exportado antes:

   ```bash
   cd backend
   cp .env.example .env
   set -a; . ./.env; set +a
   go run ./cmd/api           # ouve em :8080
   curl http://localhost:8080/health
   ```

3. Frontend, em outro terminal, em `frontend/`:

   ```bash
   cd frontend
   npm ci
   npm run dev                # http://localhost:5173
   ```

   O Vite encaminha `/api` e `/health` para `http://localhost:8080` (mesma origem, sem CORS). Para outra API:
   `VITE_PROXY_TARGET=http://localhost:9090 npm run dev`.

### Variáveis de ambiente

| Variável | Onde | Padrão | Uso |
|---|---|---|---|
| `DATABASE_URL` | API | obrigatória | URL do Postgres, por exemplo `postgres://consensu:consensu@localhost:5432/consensu?sslmode=disable`. No `docker-compose.yml` a URL vai sem usuário e senha, que chegam por `PGUSER`, `PGPASSWORD` e `PGDATABASE` (variáveis padrão do Postgres, lidas pelo driver pgx) |
| `PORT` | API | `8080` | porta HTTP |
| `SEED_DEMO` | API | `false` | `true` ou `false` (lido com `strconv.ParseBool`, que também aceita `1`/`0`, `t`/`f` e `TRUE`/`FALSE`; outro valor impede a API de subir). Com `true`, cria os dados de demonstração se não houver usuários |
| `TEST_DATABASE_URL` | testes Go | não definida | Postgres dos testes de integração; sem ela, esses testes são ignorados |
| `VITE_PROXY_TARGET` | Vite (`dev` e `preview`) | `http://localhost:8080` | destino do proxy de `/api` e `/health` |
| `BASE_URL` | testes de ponta a ponta | `http://localhost:3000` | endereço do site testado |
| `POSTGRES_PORT`, `API_PORT`, `WEB_PORT` | `docker-compose.yml` | `5432`, `8080`, `3000` | portas publicadas no host |

## Estrutura do projeto

```
.
├── backend/                    API em Go (módulo github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend)
│   ├── cmd/api/                composition root: config, banco, migrações, seed, rotas e servidor HTTP
│   ├── internal/config/        leitura de PORT, DATABASE_URL e SEED_DEMO
│   ├── internal/database/      conexão com o Postgres (database/sql + pgx)
│   ├── internal/server/        rotas HTTP, middlewares, DTOs JSON e tradução de erros para status
│   ├── internal/service/       casos de uso: autenticação, votar, encerrar, 2º turno, apurar
│   ├── internal/store/         repositórios SQL e executor de migrações
│   │   ├── migrations/         esquema do banco (embutido no binário)
│   │   └── storetest/          banco isolado por teste
│   ├── internal/seed/          cenário de demonstração
│   ├── internal/voting/        regras de domínio em funções puras
│   ├── Dockerfile              imagem da API (alpine, usuário sem privilégios)
│   ├── .env.example            variáveis do modo de desenvolvimento
│   └── .golangci.yml           configuração do lint
├── frontend/                   SPA em React 19 + TypeScript + Vite 8
│   ├── src/api/                cliente HTTP e tipos do contrato da API
│   ├── src/auth/               sessão do usuário
│   ├── src/components/         layout, rotas protegidas, cartões, diálogo de confirmação
│   ├── src/pages/              telas: painel, login, votar, resultado, admin
│   ├── src/utils/              formatação pt-BR (moeda, percentual, datas)
│   ├── Dockerfile              build com Node 22 servido por nginx
│   ├── nginx.conf              SPA, proxy de /api e /health, headers de segurança
│   └── README.md               detalhes do frontend
├── e2e/                        testes de ponta a ponta da interface (Playwright)
├── scripts/e2e-api.sh          teste de ponta a ponta da API
├── docs/                       arquitetura, requisitos, API, testes e liberação
├── .github/workflows/          ci.yml (integração) e release.yml (liberação)
├── docker-compose.yml          postgres + api + web
└── CHANGELOG.md                histórico de versões
```

## Testes e qualidade

Detalhes, estratégia e números em [docs/testes.md](docs/testes.md).

```bash
# Backend (em backend/)
go test -race ./...                                                    # sem banco
TEST_DATABASE_URL='postgres://consensu:consensu@localhost:5432/consensu?sslmode=disable' \
  go test -race -count=1 ./...                                         # com Postgres (docker compose up -d postgres)
go test -race -count=1 -covermode=atomic -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
go mod tidy -diff
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run ./...

# Frontend (em frontend/)
npm ci && npm run lint && npm test && npm run test:coverage && npm run build

# Ponta a ponta (na raiz, com o seed recém-criado)
docker compose down -v && docker compose up -d --build --wait
BASE_URL=http://localhost:3000 bash scripts/e2e-api.sh                # E2E_PULAR_LIMITE=1 pula a espera de 61 s
docker compose down -v && docker compose up -d --build --wait
cd e2e && npm ci && BASE_URL=http://localhost:3000 npm test
```

Teste de mutação (Gremlins v0.6.0 em `internal/voting`, eficácia mínima de 95%): veja o comando com container
Linux em [docs/testes.md](docs/testes.md#como-executar).

Resultado atual: backend com 614 casos de teste (cobertura total 85,2%, domínio 100%), frontend com 195 testes
(99,8% das instruções), mutação com eficácia de 97,98%, 144 verificações da API e 17 testes do Playwright, todos
passando.

## Integração contínua

`.github/workflows/ci.yml` roda em pull requests e pushes para `main` e `develop`.

| Job | O que faz | Verificação obrigatória |
|---|---|---|
| `build` | compila o backend e o frontend | sim |
| `lint` | `go mod tidy -diff`, golangci-lint v2.15.0, ESLint | sim |
| `testes` | testes do backend com Postgres 16 e `-race`; testes do frontend | sim |
| `cobertura` | cobertura do backend e do frontend; artefatos `coverage-backend` e `coverage-frontend` | sim |
| `mutacao` | Gremlins em `internal/voting`; falha abaixo de 95% de eficácia | não |
| `docker` | build das imagens, sistema completo no ar e `scripts/e2e-api.sh` pelo nginx | não |
| `e2e` | sistema completo no ar e Playwright; artefato `playwright-report` | não |

As verificações obrigatórias são exigidas pelo ruleset `protect-main-develop`, que protege `main` e `develop`: só
entram mudanças por pull request, com uma aprovação de alguém que não fez o último push e com todas as conversas
resolvidas.

## Processo de desenvolvimento e liberação

A Atividade 1 definiu um processo iterativo inspirado no Scrum, com ciclos de duas semanas, papéis por foco
(produto, facilitação, referência técnica, qualidade, integração e entrega), Definition of Ready e Definition of
Done (critérios de aceite validados e revisão aprovada por outro integrante). O backlog está nas issues do GitHub
(épicos #14 a #16 e histórias #17 a #25). Cada tarefa é feita numa branch `feature/*` criada a partir de `develop`
e integrada por pull request com revisão de outro integrante e CI, exigidos pelo ruleset `protect-main-develop`. A
versão final sai pelo fluxo GitFlow: `release/v1.0.0` a partir de `develop`, pull request para `main`, tag anotada
`v1.0.0` em `main`, workflow de release publicando a GitHub Release (binário da API, build da SPA e checksums) e
back-merge de `main` em `develop`. Comandos em [docs/release.md](docs/release.md).

## Segurança

Ameaças consideradas e como o sistema responde:

| Ameaça | Proteção |
|---|---|
| Acesso indevido a funções | autorização no servidor em toda rota: públicas só `/health`, `POST /api/login` e o resultado de votação encerrada; as demais exigem sessão; criar, encerrar, abrir 2º turno, unidades e usuários exigem `admin`. Quem pode votar (aptos, responsável pela unidade, adimplência) é decidido pelo `service`. Esconder links na interface é só usabilidade |
| Vazamento de senhas | PBKDF2-SHA256 da biblioteca padrão, 210.000 iterações, sal aleatório de 16 bytes, comparação em tempo constante; login inexistente gasta o mesmo custo de hash (não revela quais logins existem) |
| Roubo de sessão | token aleatório de 32 bytes, enviado só no header `Authorization` (nunca na URL); o banco guarda só o SHA-256; validade de 12 horas; logout apaga a sessão; respostas JSON da API com `Cache-Control: no-store` |
| Força bruta no login | só falhas contam: no máximo 10 por minuto por par (IP, login) e 60 por minuto por IP, depois `429 muitas_tentativas` |
| Injeção de SQL | SQL só com parâmetros (`$1`) |
| XSS e clickjacking | React escapa todo texto; sem `dangerouslySetInnerHTML` e sem scripts ou fontes de terceiros; o nginx envia `Content-Security-Policy` restritiva (`default-src 'self'`, `frame-ancestors 'none'`), `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer` e `Permissions-Policy` em todas as respostas, inclusive da API |
| Abuso de recursos | corpo limitado a 64 KiB, JSON estrito (campos desconhecidos recusados), timeouts no servidor HTTP |
| Vazamento de informação | erro 500 com mensagem genérica (detalhe só no log); o log de acesso registra só método, caminho, status e duração, nunca corpo, query ou headers |
| Quebra do sigilo do voto | participação e cédula em tabelas separadas; a cédula não tem eleitor nem horário e tem id UUID aleatório; o voto não é ecoado; o boletim só tem totais |
| Exposição do banco | no compose, as portas do Postgres e da API só são publicadas em `127.0.0.1` do host (o nginx fala com a API pela rede interna do compose); só o site fica acessível pela rede |

Limites que precisam ser conhecidos:

- **Sigilo e o administrador do banco.** O sigilo vale para a aplicação, a API, os boletins e os logs. Quem tem
  acesso direto ao Postgres é confiável: participação e cédula são gravadas na mesma transação (o que a regra de
  atomicidade exige), compartilham o valor da coluna de sistema `xmin` e ficam vizinhas no WAL, e um operador
  consegue reassociá-las. Mantenha o banco privado.
- **Limite de login atrás do proxy.** Através do nginx, todos os navegadores chegam à API com o IP do container
  do nginx. Por isso qualquer pessoa pode bloquear uma conta específica por até um minuto com 10 senhas erradas,
  ou todos os logins pelo site por até um minuto com 60 falhas. A API usa o endereço TCP da conexão e ignora
  `X-Forwarded-For`/`X-Real-IP` de propósito, porque esses headers podem ser forjados. O limite fica em memória:
  zera quando a API reinicia e não é compartilhado entre réplicas.
- **Token no `localStorage`.** É legível por qualquer script da página; a mitigação é a CSP e a ausência de código
  de terceiros.
- **Ambiente de demonstração.** O compose usa HTTP sem TLS, o `.env.example` traz a senha padrão do banco
  (`consensu`) e o seed tem senhas públicas. Não é uma configuração de produção.

## Limitações conhecidas

- Pesos em `float64`: com frações não inteiras, um resultado de exatamente 2/3 pode ficar logo abaixo do limiar (o
  seed usa áreas inteiras, cujas somas são exatas).
- Não há cadastro de usuários nem de unidades pela interface ou pela API; eles vêm do seed. Cada usuário responde
  por no máximo uma unidade.
- Votações não podem ser editadas nem apagadas depois de criadas.
- No fim natural do prazo (sem encerramento manual), um voto que passou pela checagem milissegundos antes do `fim`
  ainda pode confirmar logo depois; o encerramento manual não tem esse intervalo.
- Empate pelo 2º lugar no 1º turno, ou empate no 2º turno, fica para decisão da comissão, fora do sistema.
- Consulta encerrada sem nenhuma cédula marca todos os projetos como empatados.
- Os testes de ponta a ponta precisam do seed recém-criado (eles votam e encerram votações).
- Não há implantação pública: o sistema roda localmente com Docker Compose; a release publica o binário da API e o
  build da SPA, mas não imagens Docker, e o binário não serve o frontend.

## Documentação

| Documento | Conteúdo |
|---|---|
| [docs/arquitetura.md](docs/arquitetura.md) | containers, componentes, decisões D1 a D6, modelo de dados, fluxo do voto secreto, padrões e o que mudou desde a A4 |
| [docs/requisitos.md](docs/requisitos.md) | atores, épicos, rastreabilidade por critério de aceitação, decisões de domínio e limitações |
| [docs/api.md](docs/api.md) | todas as rotas com exemplos reais e a tabela de códigos de erro |
| [docs/testes.md](docs/testes.md) | estratégia por nível, como o CI roda cada um e números atuais |
| [docs/release.md](docs/release.md) | versionamento e procedimento de liberação da v1.0.0 |
| [CHANGELOG.md](CHANGELOG.md) | mudanças da v1.0.0 |
| [frontend/README.md](frontend/README.md) | telas, rotas e estrutura do frontend |
| [e2e/README.md](e2e/README.md) | como rodar os testes de ponta a ponta |
