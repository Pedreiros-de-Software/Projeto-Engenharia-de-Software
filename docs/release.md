# Processo de liberação

## Versionamento

- **SemVer** (`MAJOR.MINOR.PATCH`), com tags anotadas prefixadas por `v`. A versão final da disciplina é **`v1.0.0`**.
- A tag aponta para o commit de merge em `main`. Uma tag publicada não é movida: correções saem como `v1.0.1`,
  `v1.0.2` etc.
- As tags das atividades anteriores (por exemplo `a5-v1.0.0`) identificam entregas parciais; a versão do produto é
  a `v1.0.0`.
- O que entra em cada versão fica registrado no [`CHANGELOG.md`](../CHANGELOG.md).

## Modelo de branches (GitFlow)

| Branch | Papel |
|---|---|
| `main` | versões liberadas; cada merge aqui recebe uma tag |
| `develop` | integração contínua do trabalho |
| `feature/*`, `ci/*` | uma tarefa por branch, criada a partir de `develop`, integrada por PR em `develop` |
| `release/vX.Y.Z` | preparação de uma versão, criada a partir de `develop`, integrada por PR em `main` |
| `hotfix/vX.Y.Z` | correção urgente, criada a partir de `main`, integrada por PR em `main` |

As branches `main` e `develop` são protegidas pelo ruleset **`protect-main-develop`** do GitHub (ativo, sem
exceções):

- só entram mudanças por pull request, com **1 aprovação**;
- a aprovação tem de ser de **alguém diferente de quem fez o último push** (`require_last_push_approval`), e um push
  novo descarta aprovações anteriores (`dismiss_stale_reviews_on_push`);
- todas as conversas do review precisam estar resolvidas;
- as verificações **`build`, `lint`, `testes` e `cobertura`** do CI precisam passar;
- não é permitido apagar a branch nem forçar push.

O CI (`.github/workflows/ci.yml`) roda em todo pull request para `main` ou `develop` e em todo push nessas branches.

## Passo a passo da v1.0.0

Pré-requisito: todo o trabalho da versão já está em `develop`, integrado por PRs com o CI verde.

### 1. Criar a branch de release a partir de `develop`

```bash
git switch develop
git pull --ff-only origin develop
git switch -c release/v1.0.0
```

### 2. Ajustes finais da versão

Na branch de release entram só ajustes da liberação: a data da entrada `[1.0.0]` do `CHANGELOG.md` (troque
`AAAA-MM-DD` pela data da liberação) e, se necessário, correções de documentação. (`frontend/package.json` e `e2e/package.json` já estão na versão `1.0.0`.)

```bash
git add CHANGELOG.md
git commit -m "chore(release): prepara v1.0.0"
git push -u origin release/v1.0.0
```

### 3. Pull request para `main`

```bash
gh pr create --base main --head release/v1.0.0 \
  --title "Release v1.0.0" \
  --body "Versão final do Consensu. Mudanças em CHANGELOG.md."
gh pr checks --watch          # build, lint, testes e cobertura precisam passar
```

Outro integrante revisa e aprova (não pode ser quem fez o último push na branch). Depois da aprovação, integre com
commit de merge, que preserva o histórico de `develop` em `main`:

```bash
gh pr merge release/v1.0.0 --merge
```

### 4. Tag anotada em `main`

```bash
git switch main
git pull --ff-only origin main
git log -1 --oneline          # confira que é o commit de merge da release
git tag -a v1.0.0 -m "Consensu v1.0.0"
git push origin v1.0.0
```

### 5. Workflow de release

O push da tag dispara `.github/workflows/release.yml` (gatilho: tags `v*`, permissão `contents: write`):

1. sobe um Postgres 16 e roda os testes do backend (`go test -race -count=1 ./...`) e do frontend (`npm test`);
2. gera `consensu-api-linux-amd64` (binário estático: `CGO_ENABLED=0`, `-trimpath -ldflags="-s -w"`);
3. gera `frontend-dist.tar.gz` (conteúdo de `frontend/dist` após `npm run build`);
4. calcula `SHA256SUMS.txt`;
5. publica a GitHub Release com `gh release create "$GITHUB_REF_NAME" dist/* --repo "$GITHUB_REPOSITORY"
   --title "Consensu $GITHUB_REF_NAME" --generate-notes --verify-tag` (na tag `v1.0.0`, o título fica
   "Consensu v1.0.0").

O checkout usa `persist-credentials: false`: o token com permissão de escrita não fica no `git config` e só o
passo que cria a release o recebe (`GH_TOKEN`).

Acompanhe e confira:

```bash
gh run list --workflow release.yml --limit 1
gh run watch                  # escolha a execução da tag
gh release view v1.0.0
```

Se o workflow falhar por um motivo externo (instabilidade do runner, por exemplo), rode de novo com
`gh run rerun <id>`. Se a falha exigir mudar código, faça uma correção e libere `v1.0.1`, sem mover a tag `v1.0.0`.

### 6. Back-merge de `main` em `develop`

`develop` também é protegida, então o back-merge é um pull request de `main` para `develop`:

```bash
gh pr create --base develop --head main \
  --title "chore: back-merge da v1.0.0 em develop" \
  --body "Traz para develop o commit de merge da release v1.0.0."
gh pr checks --watch
gh pr merge main --merge      # depois da aprovação de outro integrante
```

Depois disso, `release/v1.0.0` pode ser apagada.

## Usar os artefatos da release

```bash
gh release download v1.0.0 -D consensu-v1.0.0
cd consensu-v1.0.0
sha256sum -c SHA256SUMS.txt          # no macOS: shasum -a 256 -c SHA256SUMS.txt
```

- `consensu-api-linux-amd64` precisa de `DATABASE_URL` (e aceita `PORT` e `SEED_DEMO`); aplica as migrações
  sozinho ao iniciar. Ele não serve o frontend.
- `frontend-dist.tar.gz` é o build estático da SPA. Precisa de um servidor que entregue os arquivos, devolva
  `index.html` nas rotas da SPA e encaminhe `/api/` e `/health` para a API, como o
  [`frontend/nginx.conf`](../frontend/nginx.conf).

O jeito mais simples de executar exatamente a versão liberada é pelo código da tag:

```bash
git clone https://github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software.git
cd Projeto-Engenharia-de-Software
git checkout v1.0.0
cp .env.example .env
docker compose up --build
```

## Correções depois da liberação (hotfix)

```bash
git switch main && git pull --ff-only origin main
git switch -c hotfix/v1.0.1
# correção + entrada [1.0.1] no CHANGELOG.md
git push -u origin hotfix/v1.0.1
gh pr create --base main --head hotfix/v1.0.1 --title "Hotfix v1.0.1"
# após aprovação e CI verde:
gh pr merge hotfix/v1.0.1 --merge
git switch main && git pull --ff-only origin main
git tag -a v1.0.1 -m "Consensu v1.0.1" && git push origin v1.0.1
# e o back-merge de main em develop, como no passo 6
```
