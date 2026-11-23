# Testes de ponta a ponta — Consensu

Dois testes rodam contra o sistema completo do Docker (nginx + API + Postgres), com os usuários do seed de
demonstração:

| Teste | O que cobre | Comando |
|---|---|---|
| `scripts/e2e-api.sh` (raiz) | as 9 histórias pela API HTTP, através do nginx (status, códigos de erro, apuração, sigilo, headers de segurança, rate limit do login) | `BASE_URL=http://localhost:3000 bash scripts/e2e-api.sh` |
| `e2e/` (Playwright) | o roteiro da demonstração pela interface: login, voto ponderado, bloqueio do inadimplente, 2º turno, cédula Borda, encerramento, resultados públicos, formulário de nova votação e tabelas no celular | `cd e2e && npm ci && BASE_URL=http://localhost:3000 npm test` |

Os dois **mudam os dados** (votam, encerram e criam votações) e esperam o seed **recém-criado**. Recrie o banco
antes de cada execução:

```bash
cp .env.example .env   # só na primeira vez, na raiz do repositório
docker compose down -v && docker compose up -d --build --wait
```

Para não mexer no ambiente de desenvolvimento, use outro projeto do Compose e outras portas:

```bash
POSTGRES_PORT=15432 API_PORT=18080 WEB_PORT=13000 docker compose -p consensu-e2e up -d --build --wait
BASE_URL=http://localhost:13000 bash scripts/e2e-api.sh
docker compose -p consensu-e2e down -v
```

## Playwright

- `BASE_URL` define o endereço do site (padrão `http://localhost:3000`).
- Localmente o teste usa o Google Chrome instalado (`channel: "chrome"`). Com a variável `CI` definida, usa o
  Chromium do Playwright: `npx playwright install --with-deps chromium`.
- Os testes rodam em série (`mode: "serial"`), em um único worker e sem nova tentativa: cada um depende do estado
  deixado pelo anterior. Cada usuário faz login uma só vez, porque a API limita as tentativas de login por minuto.
- O relatório HTML fica em `e2e/playwright-report` (`npm run report`); em caso de falha ele traz a captura de tela
  de cada usuário.

## `scripts/e2e-api.sh`

- Precisa de `curl` e `jq`; funciona com o `date`/shell do Linux e do macOS (os horários são calculados pelo `jq`).
- Imprime `PASS`, `FAIL` ou `SKIP` por verificação e o total; sai com código 1 se alguma falhar. `SKIP` aparece
  quando a API ainda não envia um campo conferido (por exemplo, `titulo` no resultado).
- No fim espera 61 s e confere o limite de 10 tentativas de login por minuto. `E2E_PULAR_LIMITE=1` pula essa parte.

No CI, o job `docker` roda `scripts/e2e-api.sh` e o job `e2e` roda o Playwright e publica o relatório como
artefato (`playwright-report`).
