# Frontend — Consensu

SPA em React 19 + TypeScript + Vite 8 que consome a API HTTP/JSON do backend Go (`/api`). Atende os três
contextos do Consensu: assembleia de condomínio (pauta com voto ponderado), eleição de centro acadêmico (voto
secreto, maioria absoluta e 2º turno) e priorização de projetos (cédula Borda com corte orçamentário).

## Requisitos

- Node.js `^20.19.0 || >=22.12.0` (campo `engines`); a versão de referência está em `.nvmrc` (22).
- npm (o `package-lock.json` aponta para o registro público `registry.npmjs.org`).

## Executar em modo de desenvolvimento

1. Suba o Postgres e a API (porta 8080), a partir da raiz do repositório. A API lê só variáveis de ambiente, então o
   `.env` é exportado antes:

```bash
cp .env.example .env    # só na primeira vez: credenciais do Postgres local
docker compose up -d postgres
cd backend && cp .env.example .env && set -a && . ./.env && set +a && go run ./cmd/api
```

Para subir o sistema inteiro no Docker (site em `http://localhost:3000`), veja o [README da raiz](../README.md).

2. Em outro terminal, na pasta `frontend`:

```bash
npm ci
npm run dev
```

O Vite abre em `http://localhost:5173` e faz proxy de `/api` e `/health` para `http://localhost:8080`
(mesma origem, sem CORS). Para apontar para outra API, defina `VITE_PROXY_TARGET`:

```bash
VITE_PROXY_TARGET=http://localhost:9090 npm run dev
```

O mesmo proxy vale para `npm run preview` (serve o build de produção em `dist/`).

## Scripts

| Script | O que faz |
|---|---|
| `npm run dev` | servidor de desenvolvimento com proxy para a API |
| `npm run build` | checagem de tipos (`tsc -b`, modo `strict`) e build de produção em `dist/` |
| `npm run preview` | serve o build de produção localmente (com o mesmo proxy) |
| `npm run lint` | ESLint (regras de TypeScript e de hooks do React) |
| `npm test` | testes com Vitest + Testing Library |
| `npm run test:coverage` | testes com cobertura (relatório em `coverage/`; falha abaixo de 90% de linhas/funções/instruções ou 85% de ramos) |

## Telas e rotas

| Rota | Tela | Acesso |
|---|---|---|
| `/login` | Login | público |
| `/` | Painel: votações agrupadas por contexto, status, "Votar", "Ver resultado" e, para admin, "Encerrar agora" | autenticado |
| `/votacoes/:id` | Votar: pauta (peso da unidade e bloqueio de inadimplente), eleição (voto secreto, branco e nulo) ou consulta (ordenação Borda) | autenticado |
| `/votacoes/:id/resultado` | Resultado: pauta (quórum de 2/3), boletim da eleição (2º turno) ou ranking com linha de corte do orçamento | público |
| `/admin/nova-votacao` | Cadastro de votação (tipo, janela, opções, orçamento e eleitores aptos) | admin |
| `/admin/unidades` | Unidades do condomínio e alteração de adimplência | admin |

Rotas autenticadas redirecionam para `/login`; as de admin mostram "Acesso restrito" para eleitores e não
aparecem no menu deles. A autorização de verdade é feita pelo servidor em toda rota.

## Estrutura

```
src/
  api/          types.ts (tipos do contrato HTTP), client.ts (todas as chamadas à API), erros.ts (ApiErro)
  auth/         AuthContext.tsx (AuthProvider: sessão via /api/me, entrar, sair) e useAuth.ts
  components/   Layout (cabeçalho), rotas protegidas, cartão de votação, badges, diálogo de confirmação, alertas
  pages/        Home (painel), Login, Votar, Resultado, NaoEncontrada
    votar/      cédulas de pauta, eleição e consulta
    resultado/  resultado da pauta, boletim da eleição e ranking da consulta
    admin/      NovaVotacao (+ validação em montarNovaVotacao.ts) e Unidades
  utils/        formatação pt-BR (moeda, percentual, datas), rótulos e o hook useCarregar
  styles/       reset.css, theme.css (variáveis de cor), app.css (classes comuns) e styles-pages/*
  test/         setup do Vitest, renderizador com sessão falsa e dados de exemplo
```

`src/api/client.ts` é o único ponto que chama `fetch`: usa a base relativa `/api`, envia
`Authorization: Bearer <token>`, transforma respostas de erro `{erro, codigo}` em `ApiErro` e, em 401, apaga o
token e avisa o `AuthProvider` (o usuário volta para o login).

Formatação: moeda em BRL a partir de centavos, percentuais com 2 casas e datas no fuso do navegador
(`dd/mm/aaaa às hh:mm`). Os campos `datetime-local` são convertidos para RFC 3339 (UTC) antes do envio.
Os campos em reais (custo e orçamento da consulta) aceitam `35000`, `35.000`, `35.000,50`, `12000,50` e
`12.5`: ponto seguido de três dígitos é separador de milhar. Qualquer outro formato é recusado, e o
formulário mostra ao lado do campo o valor entendido (por exemplo `Custo: R$ 35.000,00`) antes do envio.

## Testes

Os testes ficam ao lado do código (`*.test.ts(x)`) e simulam a API com `vi.mock("…/api/client")`; o próprio
`client.ts` é testado com um `fetch` falso. Cobrem: login; painel agrupando por contexto e mostrando os botões por
status; voto de pauta mostrando peso e bloqueando o inadimplente; eleição que não exibe a escolha depois do voto;
consulta que só habilita o envio com todos os projetos ordenados e envia a ordem certa; resultados (aprovada e
reprovada, 2º turno, empate, linha de corte do orçamento); telas de administração; formatação. Os testes rodam no
fuso `America/Sao_Paulo` (definido em `vitest.config.ts`) para que as datas sejam determinísticas.

A cobertura é medida em `src/**`, exceto `main.tsx`, `api/types.ts`, arquivos `.d.ts`, `src/test/` e os próprios
testes.

Os testes de ponta a ponta da interface, contra o sistema completo no Docker, ficam em `../e2e` (Playwright).

## Notas de segurança

- O token de sessão fica no `localStorage` (`consensu.token`) e só trafega no header `Authorization`, nunca na URL.
  `localStorage` é legível por qualquer script da página; por isso a aplicação não usa `dangerouslySetInnerHTML`
  nem carrega scripts nem fontes de terceiros, e todo texto vindo da API é renderizado pelo React com escape
  automático. No Docker, o nginx que serve o build (`nginx.conf`) aplica CSP e os demais headers de segurança.
- Uma resposta 401 apaga o token local; o logout apaga o token mesmo se o servidor não responder.
- Esconder links de admin é só usabilidade: a API valida papel e elegibilidade em toda requisição.
- A tela de voto da eleição e da consulta não mostra a escolha depois de registrada, e a API nunca a devolve.
- Mensagens de erro exibidas são as do campo `erro` da API; falhas sem corpo viram uma mensagem genérica.
