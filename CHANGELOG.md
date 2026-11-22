# Changelog

Todas as mudanças relevantes do Consensu ficam registradas aqui. O formato segue o
[Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/) e o projeto usa [versionamento semântico](https://semver.org/lang/pt-BR/).

## [1.0.0] - AAAA-MM-DD

Primeira versão completa: as nove histórias (#17 a #25) dos três contextos funcionam de ponta a ponta, da interface
ao banco.

### Adicionado

**Assembleia de condomínio**
- Voto em pauta (aprovar, rejeitar, abster-se) ponderado pela fração ideal da unidade, com unidade, fração e
  percentual exibidos antes da confirmação; um voto por unidade (#17).
- Bloqueio do voto de unidade inadimplente na tela e na API (`403 inadimplente`), com o motivo; tela de unidades
  para a administração alterar a adimplência (#18).
- Apuração por quórum qualificado de 2/3 do peso total cadastrado, incluindo unidades que não votaram, com
  relatório de pesos, limiar, percentual e unidades votantes (#19).

**Eleição de Centro Acadêmico**
- Voto secreto em chapa, branco ou nulo, um por eleitor apto e por turno; participação e cédula em tabelas
  separadas, cédula sem eleitor, sem horário e com UUID aleatório; a resposta do voto não repete a escolha (#20).
- Apuração por maioria absoluta dos votos válidos; sem maioria, indicação das duas chapas mais votadas e abertura
  do 2º turno pela administração (mesmos aptos, uma única vez); empate pelo 2º lugar sinalizado sem escolha
  automática (#21).
- Boletim público após o encerramento, sem login: aptos, participantes, abstenções, votos e percentuais por chapa,
  brancos, nulos e a base dos percentuais (#22).

**Priorização de projetos**
- Cédula de ordenação completa dos projetos (subir, descer, remover), enviada só quando completa; o backend recusa
  listas incompletas, repetidas ou com projetos desconhecidos (`400 cedula_invalida`) (#23).
- Apuração Borda (N pontos para a 1ª posição até 1 para a última), desempate por primeiras preferências e empate
  persistente marcado, mantendo a ordem de cadastro (#24).
- Ranking com custos, custo acumulado e linha de corte única do orçamento, total selecionado, saldo e aviso de
  empate na linha de corte (#25).

**Aplicação**
- API HTTP/JSON com 14 rotas (`docs/api.md`): login e logout, sessão, listagem e detalhe de votações, criação,
  voto, encerramento, resultado público, 2º turno, unidades e usuários.
- Autenticação por token de sessão (32 bytes aleatórios; o banco guarda só o SHA-256; validade de 12 horas), senhas
  com PBKDF2-SHA256 (210.000 iterações) e papéis `admin` e `eleitor`.
- SPA em React 19 + TypeScript + Vite 8 com painel por contexto, telas de voto e resultado dos três tipos, nova
  votação (com valores em reais no formato brasileiro e prévia do valor entendido) e unidades.
- Diálogos de confirmação com o foco preso dentro deles e devolvido ao botão de origem; tabelas largas roláveis
  dentro do próprio quadro no celular.
- Persistência em PostgreSQL 16 com migração SQL embutida no binário, aplicada ao iniciar (`pg_advisory_lock`).
- Dados de demonstração (`SEED_DEMO=true`): 15 usuários, 6 unidades do "Edifício Solar", 2 pautas, 1 eleição e
  1 consulta.

**Infraestrutura, CI e liberação**
- `docker compose up --build` sobe Postgres, API e web (nginx) com healthchecks em cadeia; Postgres e API só
  aceitam conexões da própria máquina. As credenciais do Postgres vêm do `.env` da raiz (modelo em `.env.example`),
  sem senha embutida no compose.
- Imagens: API em `alpine:3.23` com usuário sem privilégios; web em `nginx:1.29-alpine` com CSP e headers de
  segurança.
- CI com os jobs `build`, `lint` (inclui `go mod tidy -diff` e golangci-lint v2.15.0), `testes` e `cobertura` (com
  Postgres), `mutacao` (falha abaixo de 95% de eficácia), `docker` (sistema completo + `scripts/e2e-api.sh`) e
  `e2e` (Playwright).
- Workflow de release em tags `v*`: testes, binário `consensu-api-linux-amd64`, `frontend-dist.tar.gz`,
  `SHA256SUMS.txt` e GitHub Release com notas geradas.

**Testes**
- Testes de integração com Postgres em schema isolado por teste (`store`, `service`, `seed`), testes HTTP com a API
  real (`server`) e testes de concorrência (voto × encerramento, leitura consistente do resultado, migração,
  limite de login).
- Testes de componentes do frontend com Vitest e Testing Library, com cobertura mínima obrigatória.
- Teste de ponta a ponta da API pelo nginx (`scripts/e2e-api.sh`) e da interface (Playwright, `e2e/`).

**Documentação**
- `README.md`, `docs/arquitetura.md`, `docs/requisitos.md`, `docs/api.md`, `docs/testes.md` e `docs/release.md`.

### Alterado

- Módulo Go renomeado para `github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/backend`.
- `voting.ValidarElegibilidade` devolve `error` com erros sentinela (`ErrInadimplente`, `ErrIdentificadorInvalido`,
  `ErrFracaoIdealInvalida`) em vez de `(bool, string)`.
- `voting.ContarVotos` devolve `ErrPonderacaoAusente` quando a estratégia é nula, em vez de entrar em pânico.
- `voting.CalcularApuracaoBorda` é determinística (ordem de cadastro, ordenação estável), recusa opções repetidas e
  marca empates; a apuração Borda fica no pacote `voting`.
- A checagem "não é NaN nem infinito" de `quorum.go` e `elegibilidade.go` virou o helper `finito`.
- `server.NewRouter` recebe também o `service` e devolve um `http.Handler` com middlewares; `/health` mantém o
  comportamento anterior.
- CI: actions atualizadas (`checkout`, `setup-go`, `setup-node`, `upload-artifact` v7; `golangci-lint-action` v9),
  Go `1.26.x` com `check-latest`, Node lido de `frontend/.nvmrc`.
- Frontend: TypeScript em modo `strict`; sobras do template do Vite removidas; ícone do Consensu.

### Segurança

- Autorização verificada no servidor em toda rota; SQL só com parâmetros; corpo limitado a 64 KiB e JSON estrito;
  timeouts no servidor HTTP; erro 500 genérico para o cliente.
- Log de acesso sem corpo, query ou headers (senha, token e voto nunca são registrados).
- Limite de tentativas de login: só falhas contam, 10 por minuto por IP e login e 60 por minuto por IP.

### Limitações conhecidas

- Pesos em `float64`: com frações não inteiras, um resultado de exatamente 2/3 pode ficar logo abaixo do limiar.
- O sigilo vale para a aplicação, a API, os boletins e os logs; quem administra o banco consegue reassociar
  participação e cédula (mesma transação: mesmo `xmin`, posições vizinhas no WAL).
- O limite de login fica em memória e, atrás do nginx, todos os navegadores compartilham o IP do proxy.
- No fim natural do prazo, um voto iniciado milissegundos antes pode confirmar logo depois do `fim`.
- Usuários e unidades vêm só do seed; não há cadastro pela interface nem pela API.
- Execução local (Docker Compose, HTTP sem TLS); a release não publica imagens Docker e o binário não serve o
  frontend.

[1.0.0]: https://github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/releases/tag/v1.0.0
