# API HTTP do Consensu

A API é servida pelo backend Go (`backend/internal/server`). No Docker, o navegador a acessa pela mesma origem do
site (`http://localhost:3000/api/...`, via nginx); a porta `8080` também responde diretamente, mas só a partir da
própria máquina. Os tipos usados pelo frontend estão em [`frontend/src/api/types.ts`](../frontend/src/api/types.ts).

Os exemplos deste documento são respostas reais do sistema subido com `docker compose` e o seed de demonstração
recém-criado, chamadas através do nginx. Ids, tokens e datas mudam a cada execução.

## Convenções

| Item | Regra |
|---|---|
| Base | todas as rotas JSON ficam sob `/api`; `/health` fica fora |
| Autenticação | header `Authorization: Bearer <token>` (o prefixo `Bearer` não diferencia maiúsculas). O token nunca vai na URL |
| Corpo da requisição | um único objeto JSON, no máximo 64 KiB; campos desconhecidos, tipos errados ou dados depois do objeto dão `400 requisicao_invalida` |
| Datas | RFC 3339. Na saída, sempre UTC com precisão de segundos (`2026-11-19T00:39:17Z`); na entrada, frações de segundo são aceitas (`...T12:00:00.000Z`) |
| Dinheiro | inteiros em centavos (`custo_centavos`, `orcamento_centavos`) |
| `{id}` no caminho | id não numérico ou menor que 1 responde `404 nao_encontrado` |
| Rota ou método desconhecido sob `/api/` | `404 nao_encontrado` em JSON (a API não responde 405) |
| Erro | `{"erro": "<mensagem para o usuário>", "codigo": "<código estável>"}` |
| Cabeçalhos | toda resposta JSON da API tem `Content-Type: application/json` e `Cache-Control: no-store`. Pelo nginx, toda resposta também leva CSP, `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy` e `Permissions-Policy` (ver [README](../README.md#segurança)) |

## Rotas

| Método e caminho | Acesso | Sucesso |
|---|---|---|
| [`GET /health`](#get-health) | público | `200 {"status":"ok"}` |
| [`POST /api/login`](#post-apilogin) | público | `200 {token, usuario}` |
| [`POST /api/logout`](#post-apilogout) | autenticado | `204` sem corpo |
| [`GET /api/me`](#get-apime) | autenticado | `200 {usuario, unidade}` |
| [`GET /api/votacoes`](#get-apivotacoes) | autenticado | `200 VotacaoResumo[]` |
| [`GET /api/votacoes/{id}`](#get-apivotacoesid) | autenticado | `200 VotacaoDetalhe` |
| [`POST /api/votacoes`](#post-apivotacoes) | admin | `201 VotacaoDetalhe` |
| [`POST /api/votacoes/{id}/encerrar`](#post-apivotacoesidencerrar) | admin | `200 VotacaoResumo` |
| [`POST /api/votacoes/{id}/votos`](#post-apivotacoesidvotos) | autenticado | `201 {mensagem}` |
| [`GET /api/votacoes/{id}/resultado`](#get-apivotacoesidresultado) | público | `200 ResultadoPauta \| ResultadoEleicao \| ResultadoConsulta` |
| [`POST /api/votacoes/{id}/segundo-turno`](#post-apivotacoesidsegundo-turno) | admin | `201 VotacaoDetalhe` |
| [`GET /api/unidades`](#get-apiunidades) | admin | `200 Unidade[]` |
| [`PATCH /api/unidades/{id}`](#patch-apiunidadesid) | admin | `200 Unidade` |
| [`GET /api/usuarios`](#get-apiusuarios) | admin | `200 Usuario[]` |

"Autenticado" significa qualquer sessão válida (admin ou eleitor). As rotas de admin respondem `401` sem sessão e
`403 acesso_negado` para eleitor. Não há rota para cadastrar usuários ou unidades: eles vêm do seed de demonstração.

---

### `GET /health`

Verifica se a API alcança o banco. Responde `503 {"status":"database unavailable"}` quando o `Ping` falha. Qualquer
método diferente de GET recebe `405` em texto puro, com `Allow: GET`.

```http
GET /health  →  200
```
```json
{
  "status": "ok"
}
```
```http
POST /health  →  405
method not allowed
```

### `POST /api/login`

Confere login e senha e abre uma sessão de 12 horas. O token tem 43 caracteres (32 bytes aleatórios em base64url).

```json
{"login":"admin","senha":"admin123"}
```
`200`
```json
{
  "token": "s8VERrntk5KmtPUzeibkdY2T3lJRf5EfnZYINCD51KU",
  "usuario": {
    "id": 1,
    "nome": "Administração",
    "login": "admin",
    "papel": "admin"
  }
}
```

Senha errada ou login inexistente (o login diferencia maiúsculas de minúsculas), `401`:
```json
{
  "erro": "Usuário ou senha inválidos.",
  "codigo": "credenciais_invalidas"
}
```

Limite de tentativas: só tentativas com credenciais inválidas contam. São no máximo 10 falhas por minuto para o
mesmo par (IP, login) e 60 falhas por minuto por IP somando todos os logins. Acima disso, `429`:
```json
{
  "erro": "Muitas tentativas de login. Aguarde um minuto e tente de novo.",
  "codigo": "muitas_tentativas"
}
```
Na mesma execução, depois do bloqueio do login `naoexiste`, `admin` continuou entrando com `200`.

### `POST /api/logout`

Apaga a sessão do token enviado. Sem corpo na requisição nem na resposta.

```http
POST /api/logout  →  204
```
Depois disso, o mesmo token recebe `401`:
```json
{
  "erro": "Sessão inválida ou expirada. Faça login novamente.",
  "codigo": "nao_autenticado"
}
```
Sem header `Authorization`, a mensagem é outra (mesmo código):
```json
{
  "erro": "Faça login para continuar.",
  "codigo": "nao_autenticado"
}
```

### `GET /api/me`

Usuário da sessão e a unidade pela qual ele responde (ou `null`). `percentual` é a fração ideal da unidade sobre a
soma das frações das unidades com responsável, de 0 a 100.

`200` (maria, responsável pelo Apto 101):
```json
{
  "usuario": {
    "id": 2,
    "nome": "Maria Souza",
    "login": "maria",
    "papel": "eleitor"
  },
  "unidade": {
    "id": 1,
    "identificador": "Apto 101",
    "fracao_ideal": 60,
    "percentual": 12,
    "adimplente": true,
    "responsavel": {
      "id": 2,
      "nome": "Maria Souza"
    }
  }
}
```
`200` (ana, estudante sem unidade):
```json
{
  "usuario": {
    "id": 8,
    "nome": "Ana Beatriz",
    "login": "ana",
    "papel": "eleitor"
  },
  "unidade": null
}
```

### `GET /api/votacoes`

Todas as votações, das mais novas para as mais antigas, com `status` calculado no instante da requisição
(`nao_iniciada`, `aberta` ou `encerrada`) e dois campos do usuário da sessão:

- `apto`: na pauta, o usuário responde por uma unidade (mesmo inadimplente); na eleição e na consulta, o usuário
  está na lista de aptos. Admin nunca é apto.
- `ja_votou`: a unidade do usuário já votou (pauta) ou o usuário já tem participação registrada (eleição e consulta).

`200` (maria):
```json
[
  {
    "id": 4,
    "tipo": "consulta",
    "titulo": "Orçamento participativo — Laboratório de Inovação",
    "descricao": "Ordene os projetos do mais ao menos prioritário. Os mais bem colocados são financiados até o limite do orçamento.",
    "inicio": "2026-11-18T23:38:17Z",
    "fim": "2026-11-26T00:38:17Z",
    "turno": 1,
    "status": "aberta",
    "apto": true,
    "ja_votou": true,
    "votacao_anterior_id": null,
    "segundo_turno_id": null
  },
  {
    "id": 3,
    "tipo": "eleicao",
    "titulo": "Eleição Diretoria DCE 2026",
    "descricao": "Eleição da diretoria do Diretório Central dos Estudantes. Voto secreto.",
    "inicio": "2026-11-18T23:38:17Z",
    "fim": "2026-11-19T00:38:18Z",
    "turno": 1,
    "status": "encerrada",
    "apto": false,
    "ja_votou": false,
    "votacao_anterior_id": null,
    "segundo_turno_id": null
  },
  {
    "id": 2,
    "tipo": "pauta",
    "titulo": "Pauta #02 — Aprovação do orçamento para reforma da portaria (R$ 45.000,00)",
    "descricao": "Reforma da portaria com rateio do custo pela fração ideal de cada unidade.",
    "inicio": "2026-11-18T23:38:17Z",
    "fim": "2026-11-26T00:38:17Z",
    "turno": 1,
    "status": "aberta",
    "apto": true,
    "ja_votou": false,
    "votacao_anterior_id": null,
    "segundo_turno_id": null
  },
  {
    "id": 1,
    "tipo": "pauta",
    "titulo": "Pauta #01 — Troca do sistema de câmeras",
    "descricao": "Substituição das câmeras analógicas por um sistema digital com gravação em nuvem.",
    "inicio": "2026-11-18T23:38:17Z",
    "fim": "2026-11-19T00:38:18Z",
    "turno": 1,
    "status": "encerrada",
    "apto": true,
    "ja_votou": true,
    "votacao_anterior_id": null,
    "segundo_turno_id": null
  }
]
```

### `GET /api/votacoes/{id}`

`VotacaoResumo` mais `opcoes` (chapas da eleição ou projetos da consulta; `[]` na pauta), `orcamento_centavos`
(só consulta) e `minha_unidade` (só pauta, quando o usuário responde por uma unidade). `motivo` é vazio quando a
unidade é elegível.

`200` (pauta, maria):
```json
{
  "id": 2,
  "tipo": "pauta",
  "titulo": "Pauta #02 — Aprovação do orçamento para reforma da portaria (R$ 45.000,00)",
  "descricao": "Reforma da portaria com rateio do custo pela fração ideal de cada unidade.",
  "inicio": "2026-11-18T23:38:17Z",
  "fim": "2026-11-26T00:38:17Z",
  "turno": 1,
  "status": "aberta",
  "apto": true,
  "ja_votou": false,
  "votacao_anterior_id": null,
  "segundo_turno_id": null,
  "opcoes": [],
  "orcamento_centavos": null,
  "minha_unidade": {
    "identificador": "Apto 101",
    "fracao_ideal": 60,
    "percentual": 12,
    "adimplente": true,
    "elegivel": true,
    "motivo": ""
  }
}
```
`minha_unidade` da mesma pauta vista por sofia (Apto 402, inadimplente):
```json
{
  "identificador": "Apto 402",
  "fracao_ideal": 120,
  "percentual": 24,
  "adimplente": false,
  "elegivel": false,
  "motivo": "Unidade inadimplente: o voto fica bloqueado até a regularização com a administração."
}
```
`200` (consulta, lucia):
```json
{
  "id": 4,
  "tipo": "consulta",
  "titulo": "Orçamento participativo — Laboratório de Inovação",
  "descricao": "Ordene os projetos do mais ao menos prioritário. Os mais bem colocados são financiados até o limite do orçamento.",
  "inicio": "2026-11-18T23:38:17Z",
  "fim": "2026-11-26T00:38:17Z",
  "turno": 1,
  "status": "aberta",
  "apto": true,
  "ja_votou": false,
  "votacao_anterior_id": null,
  "segundo_turno_id": null,
  "opcoes": [
    { "id": 4, "nome": "Impressoras 3D e CNC", "descricao": "Equipamentos de prototipagem rápida.", "custo_centavos": 3500000 },
    { "id": 5, "nome": "Kit Arduino e Robótica", "descricao": "Kits para oficinas de eletrônica e robótica.", "custo_centavos": 1200000 },
    { "id": 6, "nome": "Servidor GPU para IA", "descricao": "Servidor para treinar modelos de aprendizado de máquina.", "custo_centavos": 4800000 },
    { "id": 7, "nome": "Mobiliário colaborativo", "descricao": "Mesas e quadros para trabalho em grupo.", "custo_centavos": 900000 }
  ],
  "orcamento_centavos": 6000000,
  "minha_unidade": null
}
```
Na eleição, cada opção é uma chapa (`custo_centavos: null`); branco e nulo não são opções, são tipos de voto.

`404` para id inexistente (`/api/votacoes/99999`) ou não numérico:
```json
{
  "erro": "Recurso não encontrado.",
  "codigo": "nao_encontrado"
}
```

### `POST /api/votacoes`

Cria a votação, suas opções e seus aptos numa única transação. Corpo:

| Campo | Regra |
|---|---|
| `tipo` | `pauta`, `eleicao` ou `consulta` |
| `titulo` | 1 a 200 caracteres (espaços nas pontas são removidos) |
| `descricao` | até 2000 caracteres |
| `inicio`, `fim` | obrigatórios, `fim` depois de `inicio` (comparados em microssegundos, a precisão do banco) |
| `opcoes` | pauta: `[]`. Eleição: 2 a 20 chapas `{nome, descricao}`, sem custo. Consulta: 2 a 20 projetos `{nome, descricao, custo_centavos}` com `0 ≤ custo_centavos ≤ 100000000000000` (R$ 1 trilhão). Nomes únicos, 1 a 120 caracteres |
| `orcamento_centavos` | só consulta, obrigatório, de 0 a 100000000000000 |
| `aptos` | eleição e consulta: ids de usuários com papel `eleitor` (repetidos são ignorados); omitido ou vazio significa todos os eleitores. Proibido na pauta, em que os aptos são os responsáveis por unidades |

Textos com o caractere U+0000 são recusados.

```json
{"tipo":"consulta","titulo":"Reforma da biblioteca","descricao":"Escolha das prioridades de 2027.","inicio":"2026-11-19T00:40:18Z","fim":"2026-11-20T00:39:18Z","opcoes":[{"nome":"Climatização","descricao":"","custo_centavos":2000000},{"nome":"Acervo novo","descricao":"","custo_centavos":1500050}],"orcamento_centavos":3000000,"aptos":[9,10,11]}
```
`201`
```json
{
  "id": 6,
  "tipo": "consulta",
  "titulo": "Reforma da biblioteca",
  "descricao": "Escolha das prioridades de 2027.",
  "inicio": "2026-11-19T00:40:18Z",
  "fim": "2026-11-20T00:39:18Z",
  "turno": 1,
  "status": "nao_iniciada",
  "apto": false,
  "ja_votou": false,
  "votacao_anterior_id": null,
  "segundo_turno_id": null,
  "opcoes": [
    { "id": 10, "nome": "Climatização", "descricao": "", "custo_centavos": 2000000 },
    { "id": 11, "nome": "Acervo novo", "descricao": "", "custo_centavos": 1500050 }
  ],
  "orcamento_centavos": 3000000,
  "minha_unidade": null
}
```
Fim antes do início, `400`:
```json
{
  "erro": "O fim da votação precisa ser depois do início.",
  "codigo": "requisicao_invalida"
}
```
Campo desconhecido no corpo (`{"tipo":"pauta","titulo":"x","extra":1}`), `400`:
```json
{
  "erro": "Corpo da requisição inválido: envie um único objeto JSON com os campos e tipos esperados.",
  "codigo": "requisicao_invalida"
}
```
Eleitor tentando criar, `403`:
```json
{
  "erro": "Acesso restrito à administração.",
  "codigo": "acesso_negado"
}
```

### `POST /api/votacoes/{id}/encerrar`

Encerra agora uma votação aberta (o `fim` passa a ser o instante atual). Espera os votos que já estão gravando
confirmarem; qualquer voto depois disso recebe `409 votacao_nao_aberta`, de modo que o resultado publicado após o
encerramento não muda. Sem corpo.

`200` (pauta 2, vista pelo admin):
```json
{
  "id": 2,
  "tipo": "pauta",
  "titulo": "Pauta #02 — Aprovação do orçamento para reforma da portaria (R$ 45.000,00)",
  "descricao": "Reforma da portaria com rateio do custo pela fração ideal de cada unidade.",
  "inicio": "2026-11-18T23:38:17Z",
  "fim": "2026-11-19T00:39:17Z",
  "turno": 1,
  "status": "encerrada",
  "apto": false,
  "ja_votou": false,
  "votacao_anterior_id": null,
  "segundo_turno_id": null
}
```
Encerrar de novo (ou encerrar uma votação não iniciada), `409`:
```json
{
  "erro": "Esta votação não está aberta.",
  "codigo": "votacao_nao_aberta"
}
```

### `POST /api/votacoes/{id}/votos`

Registra o voto do usuário da sessão. A resposta nunca repete a escolha. Cada tipo aceita só os seus campos; campos
de outro tipo dão `400 requisicao_invalida`.

| Tipo | Corpo | Regras |
|---|---|---|
| pauta | `{"escolha": "favor" \| "contra" \| "abstencao"}` | o usuário precisa responder por uma unidade (`403 nao_apto`) adimplente (`403 inadimplente`); um voto por unidade (`409 voto_duplicado`) |
| eleição | `{"escolha": "chapa", "opcao_id": n}` ou `{"escolha": "branco" \| "nulo"}` | o usuário precisa estar nos aptos (`403 nao_apto`); `opcao_id` precisa ser chapa desta votação (`400 opcao_desconhecida`); um voto por eleitor e por turno (`409 voto_duplicado`) |
| consulta | `{"preferencias": [ids]}` | aptos como na eleição; a lista precisa conter cada projeto exatamente uma vez, do preferido para o menos preferido (`400 cedula_invalida`) |

Em qualquer tipo, votar fora da janela dá `409 votacao_nao_aberta`.

`{"escolha":"favor"}` na pauta 2 (maria), `201`:
```json
{
  "mensagem": "Voto registrado com sucesso."
}
```
Segundo voto da mesma unidade, `409`:
```json
{
  "erro": "Você já votou nesta votação.",
  "codigo": "voto_duplicado"
}
```
sofia (unidade inadimplente), mesmo enviando direto à API, `403`:
```json
{
  "erro": "Unidade inadimplente: o voto fica bloqueado até a regularização com a administração.",
  "codigo": "inadimplente"
}
```
ana (estudante, sem unidade) na pauta, `403`:
```json
{
  "erro": "Você não está habilitado a votar nesta votação.",
  "codigo": "nao_apto"
}
```
`{"escolha":"talvez"}` na pauta, `400`:
```json
{
  "erro": "Na pauta, envie apenas escolha: favor, contra ou abstencao.",
  "codigo": "requisicao_invalida"
}
```
Eleição (2º turno, id 5): `{"escolha":"chapa","opcao_id":8}` e `{"escolha":"branco"}` dão `201` com a mesma
mensagem. Usar o id de uma chapa do 1º turno (`{"escolha":"chapa","opcao_id":1}`), `400`:
```json
{
  "erro": "Opção desconhecida para esta votação.",
  "codigo": "opcao_desconhecida"
}
```
Consulta: `{"preferencias":[5,6,4,7]}` dá `201`. Lista incompleta (`[5,4]`) ou repetida (`[5,5,6,7]`), `400`:
```json
{
  "erro": "A cédula deve ordenar todos os projetos, cada um uma única vez.",
  "codigo": "cedula_invalida"
}
```

### `GET /api/votacoes/{id}/resultado`

Público (não exige login). Só responde depois do encerramento; antes disso, `409`:
```json
{
  "erro": "O resultado só é divulgado depois do encerramento.",
  "codigo": "votacao_nao_encerrada"
}
```
Todas as leituras vêm de um único retrato do banco (transação `REPEATABLE READ` somente leitura), então os totais
sempre batem entre si. Os três formatos começam com `tipo`, `votacao_id`, `titulo` e `descricao`.

**Pauta.** `contagem` soma frações ideais. `peso_total` inclui as unidades que não votaram;
`limiar = peso_total × 2/3`; `aprovada` quando `contagem.favor ≥ limiar`. Pauta 1 do seed:
```json
{
  "tipo": "pauta",
  "votacao_id": 1,
  "titulo": "Pauta #01 — Troca do sistema de câmeras",
  "descricao": "Substituição das câmeras analógicas por um sistema digital com gravação em nuvem.",
  "contagem": {
    "favor": 300,
    "contra": 80,
    "abstencao": 0
  },
  "peso_total": 500,
  "limiar": 333.3333333333333,
  "percentual_favor": 60,
  "aprovada": false,
  "unidades_total": 6,
  "unidades_votantes": 5
}
```
(Na demonstração da Pauta #02, quando maria, joão e lúcia votam a favor, `favor` chega a 380 e `aprovada` passa a
`true`.)

**Eleição (boletim).** `validos` soma só as chapas; `percentual_validos` usa essa base. `vencedora` exige mais da
metade dos válidos. Sem maioria, no 1º turno, `segundo_turno.necessario` é `true` e `chapas` traz as duas mais
votadas; se o 2º lugar empata com o 3º, `empate_indefinido` é `true` e `chapas` fica vazio. `segundo_turno.votacao_id`
aponta para o 2º turno quando ele já foi aberto. No 2º turno, `segundo_turno` é sempre `necessario: false`. O boletim
não tem horários nem dados de eleitores. Eleição 3 do seed, depois de aberto o 2º turno:
```json
{
  "tipo": "eleicao",
  "votacao_id": 3,
  "titulo": "Eleição Diretoria DCE 2026",
  "descricao": "Eleição da diretoria do Diretório Central dos Estudantes. Voto secreto.",
  "turno": 1,
  "aptos": 8,
  "participantes": 7,
  "abstencoes": 1,
  "percentual_comparecimento": 87.5,
  "votos": [
    { "opcao_id": 1, "nome": "Chapa 1 — Movimenta", "votos": 3, "percentual_validos": 50 },
    { "opcao_id": 2, "nome": "Chapa 2 — Renovação", "votos": 2, "percentual_validos": 33.33333333333333 },
    { "opcao_id": 3, "nome": "Chapa 3 — Integração", "votos": 1, "percentual_validos": 16.666666666666664 }
  ],
  "brancos": 1,
  "nulos": 0,
  "validos": 6,
  "vencedora": null,
  "segundo_turno": {
    "necessario": true,
    "empate_indefinido": false,
    "chapas": [
      { "opcao_id": 1, "nome": "Chapa 1 — Movimenta" },
      { "opcao_id": 2, "nome": "Chapa 2 — Renovação" }
    ],
    "votacao_id": 5
  }
}
```
2º turno (id 5) com 2 votos na Chapa 1 e 1 branco:
```json
{
  "tipo": "eleicao",
  "votacao_id": 5,
  "titulo": "Eleição Diretoria DCE 2026 — 2º turno",
  "descricao": "Eleição da diretoria do Diretório Central dos Estudantes. Voto secreto.",
  "turno": 2,
  "aptos": 8,
  "participantes": 3,
  "abstencoes": 5,
  "percentual_comparecimento": 37.5,
  "votos": [
    { "opcao_id": 8, "nome": "Chapa 1 — Movimenta", "votos": 2, "percentual_validos": 100 },
    { "opcao_id": 9, "nome": "Chapa 2 — Renovação", "votos": 0, "percentual_validos": 0 }
  ],
  "brancos": 1,
  "nulos": 0,
  "validos": 2,
  "vencedora": { "opcao_id": 8, "nome": "Chapa 1 — Movimenta" },
  "segundo_turno": {
    "necessario": false,
    "empate_indefinido": false,
    "chapas": [],
    "votacao_id": null
  }
}
```

**Consulta (ranking Borda com corte orçamentário).** Com N projetos, a 1ª posição de cada cédula vale N pontos e a
última, 1. O ranking ordena por pontos e, no empate, por primeiras preferências; se ainda empatar, mantém a ordem
de cadastro e marca `empatado: true`. Um projeto fica `dentro_do_orcamento` se o custo acumulado (incluindo ele)
cabe no orçamento e todos os anteriores ficaram dentro; o primeiro que não cabe e todos abaixo ficam fora.
`empate_na_linha_de_corte` sinaliza empate entre o último dentro e o primeiro fora. Consulta 4 do seed com 6
cédulas:
```json
{
  "tipo": "consulta",
  "votacao_id": 4,
  "titulo": "Orçamento participativo — Laboratório de Inovação",
  "descricao": "Ordene os projetos do mais ao menos prioritário. Os mais bem colocados são financiados até o limite do orçamento.",
  "cedulas": 6,
  "orcamento_centavos": 6000000,
  "total_selecionado_centavos": 4700000,
  "saldo_centavos": 1300000,
  "empate_na_linha_de_corte": false,
  "ranking": [
    { "posicao": 1, "opcao_id": 4, "nome": "Impressoras 3D e CNC", "pontos": 18, "primeiras_preferencias": 2,
      "custo_centavos": 3500000, "custo_acumulado_centavos": 3500000, "dentro_do_orcamento": true, "empatado": true },
    { "posicao": 2, "opcao_id": 5, "nome": "Kit Arduino e Robótica", "pontos": 18, "primeiras_preferencias": 2,
      "custo_centavos": 1200000, "custo_acumulado_centavos": 4700000, "dentro_do_orcamento": true, "empatado": true },
    { "posicao": 3, "opcao_id": 6, "nome": "Servidor GPU para IA", "pontos": 15, "primeiras_preferencias": 2,
      "custo_centavos": 4800000, "custo_acumulado_centavos": 9500000, "dentro_do_orcamento": false, "empatado": false },
    { "posicao": 4, "opcao_id": 7, "nome": "Mobiliário colaborativo", "pontos": 9, "primeiras_preferencias": 0,
      "custo_centavos": 900000, "custo_acumulado_centavos": 10400000, "dentro_do_orcamento": false, "empatado": false }
  ]
}
```
O Mobiliário (R$ 9.000) cabe no saldo de R$ 13.000, mas fica fora: a política é de linha de corte única.

### `POST /api/votacoes/{id}/segundo-turno`

Cria o 2º turno de uma eleição de 1º turno encerrada sem maioria absoluta: título com o sufixo ` — 2º turno`, as
duas chapas mais votadas (nome e descrição copiados), os mesmos aptos, `inicio` agora e `fim` informado (padrão:
agora + 24 h). Corpo opcional: vazio, `{}` ou `{"fim": "<RFC 3339 no futuro>"}` (fim no passado dá `400`).

`{}`, `201`:
```json
{
  "id": 5,
  "tipo": "eleicao",
  "titulo": "Eleição Diretoria DCE 2026 — 2º turno",
  "descricao": "Eleição da diretoria do Diretório Central dos Estudantes. Voto secreto.",
  "inicio": "2026-11-19T00:39:17Z",
  "fim": "2026-11-20T00:39:17Z",
  "turno": 2,
  "status": "aberta",
  "apto": false,
  "ja_votou": false,
  "votacao_anterior_id": 3,
  "segundo_turno_id": null,
  "opcoes": [
    { "id": 8, "nome": "Chapa 1 — Movimenta", "descricao": "Ana / Victor / Amanda", "custo_centavos": null },
    { "id": 9, "nome": "Chapa 2 — Renovação", "descricao": "João / Ana / Marcela", "custo_centavos": null }
  ],
  "orcamento_centavos": null,
  "minha_unidade": null
}
```
Pedir de novo, `409` (o mesmo código vale para: não é eleição, não é 1º turno, não está encerrada, teve vencedora,
`empate_indefinido` ou nenhum voto válido):
```json
{
  "erro": "O 2º turno só pode ser aberto uma vez, para eleição de 1º turno encerrada sem maioria absoluta e com as duas chapas definidas.",
  "codigo": "segundo_turno_indisponivel"
}
```

### `GET /api/unidades`

Unidades em ordem de id. Eleitor recebe `403 acesso_negado`.

`200` (trecho: 2 de 6 unidades):
```json
[
  {
    "id": 1,
    "identificador": "Apto 101",
    "fracao_ideal": 60,
    "percentual": 12,
    "adimplente": true,
    "responsavel": { "id": 2, "nome": "Maria Souza" }
  },
  {
    "id": 6,
    "identificador": "Apto 402",
    "fracao_ideal": 120,
    "percentual": 24,
    "adimplente": false,
    "responsavel": { "id": 7, "nome": "Sofia Prado" }
  }
]
```

### `PATCH /api/unidades/{id}`

Altera a adimplência. Corpo `{"adimplente": true | false}` (obrigatório).

`{"adimplente":true}` no Apto 402, `200`:
```json
{
  "id": 6,
  "identificador": "Apto 402",
  "fracao_ideal": 120,
  "percentual": 24,
  "adimplente": true,
  "responsavel": {
    "id": 7,
    "nome": "Sofia Prado"
  }
}
```
`{}`, `400`:
```json
{
  "erro": "Informe adimplente: true ou false.",
  "codigo": "requisicao_invalida"
}
```
Outro método no mesmo caminho (por exemplo `POST /api/unidades/6`) responde `404 nao_encontrado`.

### `GET /api/usuarios`

Todos os usuários (admin incluído), em ordem de id, sem hash de senha. Usada pela tela de nova votação para escolher
os aptos (o frontend mostra só os eleitores).

`200` (trecho: 3 de 15 usuários):
```json
[
  { "id": 1, "nome": "Administração", "login": "admin", "papel": "admin" },
  { "id": 2, "nome": "Maria Souza", "login": "maria", "papel": "eleitor" },
  { "id": 8, "nome": "Ana Beatriz", "login": "ana", "papel": "eleitor" }
]
```

---

## Códigos de erro

| Código | Status | Quando |
|---|---|---|
| `requisicao_invalida` | 400 | corpo vazio, malformado, maior que 64 KiB, com campo desconhecido ou tipo errado; dado fora das regras (título, datas, opções, custos, orçamento, aptos, texto com U+0000); campos de outro tipo de votação no voto; `fim` do 2º turno no passado; `adimplente` ausente |
| `cedula_invalida` | 400 | cédula Borda incompleta, com projeto repetido, desconhecido ou a mais |
| `opcao_desconhecida` | 400 | `opcao_id` que não é chapa desta eleição |
| `nao_autenticado` | 401 | sem header `Authorization`, token inválido, expirado ou já encerrado por logout |
| `credenciais_invalidas` | 401 | login inexistente ou senha errada |
| `acesso_negado` | 403 | rota de admin chamada por eleitor |
| `nao_apto` | 403 | usuário fora dos aptos da eleição/consulta, sem unidade na pauta, ou admin tentando votar |
| `inadimplente` | 403 | voto na pauta de unidade inadimplente |
| `nao_encontrado` | 404 | votação ou unidade inexistente, `{id}` inválido, rota ou método desconhecido sob `/api/` |
| `voto_duplicado` | 409 | segundo voto da mesma unidade (pauta) ou do mesmo eleitor (eleição, consulta) |
| `votacao_nao_aberta` | 409 | votar ou encerrar fora da janela (antes do início ou depois do fim) |
| `votacao_nao_encerrada` | 409 | pedir o resultado antes do encerramento |
| `segundo_turno_indisponivel` | 409 | 2º turno não aplicável ou já aberto |
| `muitas_tentativas` | 429 | limite de falhas de login atingido (10/min por IP e login; 60/min por IP) |
| `erro_interno` | 500 | falha inesperada; o cliente recebe mensagem genérica e o detalhe vai só para o log |
