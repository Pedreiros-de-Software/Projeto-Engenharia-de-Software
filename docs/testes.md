# Estratégia de testes

Os testes vão do domínio puro até o sistema completo no Docker. Cada nível roda no CI
(`.github/workflows/ci.yml`); os quatro jobs `build`, `lint`, `testes` e `cobertura` são verificações obrigatórias
para integrar em `develop` e `main`.

## Níveis

| Nível | Onde | Ferramenta | O que verifica | Job do CI |
|---|---|---|---|---|
| Unitário do domínio | `backend/internal/voting/*_test.go` | `go test` (tabelas de casos) | regras dos três contextos, com classes de equivalência e valores limite | `testes`, `cobertura` |
| Mutação | `backend/internal/voting` | Gremlins v0.6.0 | se os testes do domínio detectam alterações no código | `mutacao` |
| Integração com Postgres | `backend/internal/store`, `service`, `seed` | `go test` + Postgres 16 | SQL, migrações, transações, travas, regras dos casos de uso | `testes`, `cobertura` (serviço `postgres`) |
| HTTP | `backend/internal/server` | `net/http/httptest` | rotas, códigos de erro, autorização, contrato JSON, middlewares | `testes`, `cobertura` |
| Componentes do frontend | `frontend/src/**/*.test.ts(x)` | Vitest 4 + Testing Library + user-event (jsdom) | telas, formulários, formatação, cliente HTTP | `testes`, `cobertura` |
| Ponta a ponta da API | `scripts/e2e-api.sh` | `curl` + `jq` | as 9 histórias pela API, através do nginx, com o seed | `docker` |
| Ponta a ponta da interface | `e2e/tests/fluxos.spec.ts` | Playwright 1.65 (Chromium) | o roteiro da demonstração no navegador | `e2e` |

### Domínio: classes de equivalência e valores limite

As funções de `voting` são puras (decisão D1 da [arquitetura](arquitetura.md)), então cada regra é testada com uma
tabela de casos, sem banco nem relógio. Os três escopos da A5 seguem com os nomes de caso prefixados pela classe
(`CE3: total zerado`, `VL1: exatamente 2/3`):

| Função | Classes de equivalência | Valores limite |
|---|---|---|
| `AtingiuQuorum` (`quorum_test.go`) | `TestAtingiuQuorumClassesEquivalencia`: CE1 aprovada (razão ≥ 2/3), CE2 reprovada, CE3 total ≤ 0, CE4 votos < 0, CE5 votos > total, CE6 NaN, CE7 infinito | `TestAtingiuQuorumValorLimite`: VL1 em torno de 2/3 (com `math.Nextafter`), VL2 unanimidade e logo acima do total, VL3 total em torno de zero |
| `VerificarMaioriaAbsoluta` (`ca_eleicao_test.go`) | `TestVerificarMaioriaAbsolutaClassesEquivalencia`: vencedor, 2º turno, sem votos, entradas negativas | `TestVerificarMaioriaAbsolutaValorLimite`: 49×50, 50×50, 51×50, 0 e −1 |
| `StatusVotacao` (`periodo_test.go`) | `TestStatusVotacaoClassesEquivalencia`: não iniciada, aberta, encerrada, janela invertida, data zerada | `TestStatusVotacaoValorLimite`: 1 ns antes e exatamente no início e no fim, janela de 1 ns |

As regras novas seguem o mesmo estilo: `TestApurarEleicao` e `TestMontarBoletim` (maioria absoluta, branco e nulo,
empate pelo 2º lugar, totais inconsistentes), `TestCalcularApuracaoBorda` e `TestValidarCedulaBorda` (pontos,
primeiras preferências, empate persistente, cédula incompleta ou repetida), `TestAplicarCorteOrcamentario` (linha
de corte, empate na linha, estouro de `int64`), `TestApurarPauta` e `TestContarVotos` (Strategy e Facade). Os casos
`TestBoletimDaEleicaoDoSeed` e `TestConsultaDoSeed` conferem os números do cenário de demonstração.

### Mutação

O job `mutacao` roda o Gremlins sobre `internal/voting` e **falha abaixo de 95% de eficácia**. O limiar vem de um
arquivo de configuração passado com `--config`, porque a v0.6.0 ignora a flag `--threshold-efficacy`. O job usa
`setup-go` com `cache: false`: com o cache de build quente, a execução de referência cai para cerca de 50 ms, o
tempo-limite de cada mutante fica curto demais e todos estouram (eficácia 0%).

Resultado atual: 97 mutantes mortos, 2 sobreviventes, 0 por tempo esgotado, eficácia 97,98%, cobertura de mutantes
100%. Os dois sobreviventes são `CONDITIONALS_BOUNDARY` em `quorum.go:9` e são equivalentes, ou seja, nenhum teste
consegue matá-los porque o resultado não muda:

- `totalFracoes <= 0` → `< 0`: com total 0, ou os votos são maiores que o total (a guarda seguinte devolve
  `false`), ou também são 0 e a divisão `0/0` dá NaN, que não é ≥ 2/3. Resultado: `false` nos dois códigos.
- `votosFavoraveis < 0` → `<= 0`: com 0 votos favoráveis e total positivo, a razão 0 também não atinge 2/3.

Na A5, o mutante analisado foi o `>` → `>=` em `votosFavoraveis > totalFracoes`, que rejeitava a unanimidade; o caso
VL2 (100 de 100) passou a matá-lo.

### Integração com Postgres

`storetest.NovoBanco(t)` cria um schema exclusivo por teste (`teste_<hex aleatório>`), conecta com `search_path`
apontando para ele, aplica as migrações e apaga o schema no fim. Assim os testes rodam em paralelo no mesmo banco
sem interferir entre si nem nos dados do modo de desenvolvimento (schema `public`). `storetest.NovoBancoVazio(t)`
faz o mesmo sem migrar, e é usado para testar o próprio executor de migrações. Sem a variável
`TEST_DATABASE_URL`, esses testes chamam `t.Skip`: `go test ./...` continua passando sem banco, com 75 casos
ignorados.

Os testes do `service` e do `server` injetam um relógio falso (`relogioFalso`) e usam 1.000 iterações do PBKDF2
em vez de 210.000, para rodar rápido. Os testes do `server` montam a API real (rotas + service + banco de teste)
com o seed de demonstração e fazem requisições com `httptest`.

Alguns testes cobrem concorrência de verdade, com goroutines e `-race`:

| Teste | O que prova |
|---|---|
| `TestMigrarEIdempotenteEConcorrente` (store) | 6 migrações simultâneas num schema vazio terminam com cada arquivo aplicado uma vez (`pg_advisory_lock`) |
| `TestTravaDoVotoContraEncerramento` (store) | o encerramento espera um voto em andamento; votos não se bloqueiam entre si; um voto que esperou o encerramento lê o novo fim |
| `TestEmLeituraVeUmRetratoFixoESoLe` (store) | a leitura do resultado vê um retrato fixo e não aceita escrita |
| `TestVotosConcorrentesComEncerramento` (service) | 20 eleitores × 20 rodadas votando enquanto a votação é encerrada: o boletim publicado logo após o encerramento é o final |
| `TestResultadoConsistenteComVotosNoFimDoPrazo` (service) | resultado consultado enquanto votos confirmam perto do fim do prazo nunca dá erro nem totais inconsistentes |
| `TestLoginLimitaTentativasSimultaneas` (service) | 30 logins errados simultâneos: exatamente 10 `credenciais_invalidas` e 20 `muitas_tentativas` |

### Frontend

Os testes ficam ao lado dos componentes e simulam a API com `vi.mock("…/api/client")`; o próprio `client.ts` é
testado com um `fetch` falso. Rodam no fuso `America/Sao_Paulo` (definido em `vitest.config.ts`) para as datas
serem determinísticas. A cobertura é medida em `src/**` (exceto `main.tsx`, `api/types.ts`, `src/test/` e os
testes) e **falha abaixo de 90% de instruções, funções e linhas ou 85% de ramos**.

### Ponta a ponta

Os dois testes de ponta a ponta usam o sistema completo do Docker (nginx + API + Postgres) com o seed
**recém-criado**, porque votam, encerram e criam votações.

- `scripts/e2e-api.sh` percorre as 9 histórias pela API através do nginx, confere códigos de erro, os headers de
  segurança, o sigilo (o voto não é ecoado, o boletim não tem dados individuais) e o limite de tentativas de login
  (espera 61 s para começar numa janela nova; `E2E_PULAR_LIMITE=1` pula essa parte). Imprime `PASS`, `FAIL` ou
  `SKIP` por verificação e sai com código 1 se algo falhar.
- `e2e/tests/fluxos.spec.ts` (Playwright) executa o roteiro da demonstração com 7 usuários do seed (admin, maria,
  joão, lúcia, sofia, ana e bruno) e um visitante sem login: login, voto ponderado,
  bloqueio do inadimplente, 2º turno, cédula Borda, encerramento, resultados públicos, criação de consulta com
  valores em reais, foco preso no diálogo de confirmação, tabelas roláveis no celular (390 px) e ausência de erros
  no console (inclui violações de CSP). Em caso de falha, anexa capturas de tela de cada usuário ao relatório.

## Números da v1.0.0

| Suíte | Resultado |
|---|---|
| Backend (`go test -race` com Postgres) | 102 funções de teste, 614 casos contando subtestes, 0 falhas |
| Frontend (Vitest) | 14 arquivos, 195 testes, 0 falhas |
| Mutação (`internal/voting`) | 97 mortos, 2 sobreviventes equivalentes, 0 por tempo esgotado: eficácia 97,98% |
| `scripts/e2e-api.sh` | 144 PASS, 0 FAIL, 0 SKIP |
| Playwright | 17 testes, 17 passaram |

Cobertura do backend (instruções, `go test -cover` com Postgres):

| Pacote | Cobertura | Observação |
|---|---|---|
| `internal/voting` | 100,0% | domínio |
| `internal/config` | 100,0% | |
| `internal/server` | 89,9% | |
| `internal/service` | 88,3% | |
| `internal/store` | 88,1% | |
| `internal/seed` | 79,8% | |
| `internal/database` | 75,0% | faltam o erro de `sql.Open` (o pgx só valida ao conectar) e a falha do `Close` depois de um `Ping` com erro |
| `cmd/api` | 0,0% | composition root, sem testes unitários; exercitado pelos jobs `docker` e `e2e` |
| `internal/store/storetest` | 0,0% | utilitário de teste |
| **total** | **85,2%** | |

As linhas não cobertas em `server`, `service` e `store` são, na maior parte, retornos de erro do banco
(`return err` depois de uma consulta).

Cobertura do frontend (`npm run test:coverage`): instruções 99,8% (500/501), ramos 97,75% (348/356), funções 100%
(174/174), linhas 100% (470/470).

## Como executar

Backend (a partir de `backend/`):

```bash
go test -race ./...                                        # sem banco: testes de integração são ignorados

docker compose up -d postgres                              # na raiz, com o .env copiado de .env.example
export TEST_DATABASE_URL='postgres://consensu:consensu@localhost:5432/consensu?sslmode=disable'
go test -race -count=1 ./...                               # todos os testes

go test -race -count=1 -covermode=atomic -coverprofile=coverage.out ./...
go tool cover -func=coverage.out                           # cobertura por função e total
go tool cover -html=coverage.out -o coverage.html          # relatório navegável
```

Mutação: no macOS o Gremlins marca todos os mutantes como tempo esgotado (problema do ambiente local), então rode
num container Linux, com os mesmos comandos do CI (a partir da raiz do repositório):

```bash
docker run --rm -v "$PWD/backend":/src:ro golang:1.26 bash -ec '
  cp -r /src /work && cd /work
  go install github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0
  printf "unleash:\n  threshold:\n    efficacy: 95\n" > /tmp/gremlins.yaml
  gremlins unleash --config /tmp/gremlins.yaml ./internal/voting'
```

Frontend (a partir de `frontend/`):

```bash
npm ci
npm test                 # Vitest
npm run test:coverage    # com cobertura e limites mínimos; relatório em coverage/
```

Ponta a ponta (a partir da raiz; recrie o banco antes de cada execução):

```bash
cp .env.example .env     # só na primeira vez
docker compose down -v && docker compose up -d --build --wait
BASE_URL=http://localhost:3000 bash scripts/e2e-api.sh

docker compose down -v && docker compose up -d --build --wait
cd e2e && npm ci && BASE_URL=http://localhost:3000 npm test
npm run report           # abre o relatório HTML (e2e/playwright-report)
```

Para não mexer no ambiente de desenvolvimento, use outro projeto do Compose e outras portas, como em
[`e2e/README.md`](../e2e/README.md). Localmente o Playwright usa o Google Chrome instalado; com a variável `CI`
definida, usa o Chromium baixado por `npx playwright install --with-deps chromium`.

## No CI

| Job | Testes | Obrigatório |
|---|---|---|
| `build` | compila backend e frontend (`tsc -b` em modo `strict` + `vite build`) | sim |
| `lint` | `go mod tidy -diff`, golangci-lint v2.15.0 (errcheck, govet, staticcheck, ineffassign, unused, gofmt), ESLint | sim |
| `testes` | `go test -race -count=1 ./...` com serviço Postgres 16 e `TEST_DATABASE_URL`; `npm test` | sim |
| `cobertura` | mesmos testes com cobertura; tabela da cobertura do backend no resumo do job; artefatos `coverage-backend` e `coverage-frontend` | sim |
| `mutacao` | Gremlins em `internal/voting`, falha abaixo de 95% | não |
| `docker` | `docker compose build`, sobe o sistema, confere `/health`, a SPA, um link direto e o login pelo nginx, roda `scripts/e2e-api.sh` | não |
| `e2e` | sobe o sistema e roda o Playwright; publica o artefato `playwright-report` | não |
