# Requisitos e rastreabilidade (v1.0.0)

Os requisitos estão nas issues do GitHub (Atividade 3): épicos
[#14](https://github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/issues/14),
[#15](https://github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/issues/15) e
[#16](https://github.com/Pedreiros-de-Software/Projeto-Engenharia-de-Software/issues/16) e histórias #17 a #25. Este
documento liga cada critério de aceitação ao código, ao endpoint, à tela e ao teste automatizado que o verificam.

## Problema

Decisões coletivas seguem regras diferentes conforme o contexto, e fazer a apuração à mão é trabalhoso e sujeito a
erro. Alinhado ao ODS 16 da ONU (metas 16.6 e 16.7: instituições transparentes e decisões participativas), o Consensu
atende três contextos:

1. **Assembleia de condomínio** (#14): a pauta é votada a favor, contra ou abstenção; o voto vale a fração ideal da
   unidade; unidade inadimplente não vota; aprovar exige 2/3 do peso total cadastrado.
2. **Eleição de Centro Acadêmico** (#15): voto secreto, um estudante um voto por turno, chapas, branco e nulo;
   vence quem tem maioria absoluta dos válidos; senão, 2º turno entre as duas mais votadas; boletim público após o
   encerramento.
3. **Priorização de projetos** (#16): cada participante ordena todos os projetos; a apuração é pelo método Borda;
   o ranking mostra quais projetos cabem no orçamento.

As necessidades são hipóteses da equipe (não houve entrevistas), e as regras representam o escopo acadêmico, sem
promessa de conformidade jurídica.

## Atores

| Ator | No sistema | O que faz |
|---|---|---|
| Condômino | usuário `eleitor` responsável por uma unidade | vota nas pautas pela sua unidade |
| Estudante / participante | usuário `eleitor` incluído nos aptos | vota na eleição e na consulta |
| Administração (síndico, comissão eleitoral, organizador) | usuário `admin` | cria votações, encerra, abre o 2º turno, atualiza a adimplência das unidades; não vota |
| Visitante | sem login | consulta resultados de votações encerradas |

O papel "Auditor" previsto na A1 saiu do escopo: o boletim é público.

## Épicos e histórias

| Épico | Histórias |
|---|---|
| #14 Assembleia de condomínio | #17 voto ponderado, #18 bloqueio por inadimplência, #19 quórum qualificado |
| #15 Eleição de Centro Acadêmico | #20 voto secreto, #21 maioria absoluta e 2º turno, #22 boletim público |
| #16 Priorização de projetos | #23 ordenação de preferências, #24 apuração Borda, #25 ranking orçamentário |

## Rastreabilidade por critério de aceitação

Convenções das colunas:

- **Backend**: arquivos em `backend/internal/` (a migração é `store/migrations/0001_esquema_inicial.sql`).
- **Tela**: arquivos em `frontend/src/pages/`.
- **Testes**: funções Go (pacote entre parênteses; o texto entre aspas é o subteste), testes do frontend
  (`arquivo: descrição`) e os testes de ponta a ponta: `e2e-api` = seção de `scripts/e2e-api.sh`, `PW` = teste do
  Playwright em `e2e/tests/fluxos.spec.ts`.
- **Situação**: atendido, parcial ou limitação. As notas L1 a L3 estão em [Limitações](#fora-do-escopo-e-limitações).

### #17 Voto ponderado por peso cadastrado da unidade

| Critério | Backend | Endpoint | Tela | Testes | Situação |
|---|---|---|---|---|---|
| Recuperar e exibir o peso da unidade antes da confirmação do voto | `service/votacoes.go` (`Detalhar` → `minha_unidade`), `service/unidades.go` | `GET /api/votacoes/{id}` | `votar/VotarPauta.tsx`: unidade, fração ideal e % do total na tela e no diálogo de confirmação | `TestDetalhar` (service); `TestListarEDetalharVotacoes` "detalhe da pauta mostra a unidade e o peso antes do voto" (server); `Votar.test.tsx`: "mostra unidade, fração ideal e peso antes de confirmar e envia a escolha"; PW "#17 maria vê unidade, fração e peso antes de confirmar e aprova a Pauta #02" | atendido |
| Somar o peso da unidade à opção escolhida na apuração | `voting/pauta.go` (`ApurarPauta`), `voting/ponderacao.go` (`PesoPorUnidade`), `service/resultado.go` | `GET /api/votacoes/{id}/resultado` | `resultado/ResultadoDaPauta.tsx` | `TestApurarPauta`, `TestContarVotos` (voting); `TestResultadoPauta` (service); `TestVotacaoDaPautaDeCondominio` (server); e2e-api "#19 Quórum qualificado"; PW "joão e lúcia também aprovam a Pauta #02 (380 de 500 ≥ 2/3)" | atendido |
| Impedir mais de um voto do mesmo condômino na mesma pauta | migração: `votos_pauta` com chave `(votacao_id, unidade_id)` e `unidades.responsavel_id UNIQUE`; `service/votar.go` | `POST /api/votacoes/{id}/votos` → `409 voto_duplicado` | `Votar.tsx` ("Você já votou nesta votação") e etiqueta "Você já votou" no painel (`components/CartaoVotacao.tsx`) | `TestVotarPauta` "segundo voto da mesma unidade" (service); `TestVotosPauta` (store); `TestApurarPauta` (voting); `Votar.test.tsx`: "quem já votou vê o aviso em vez da cédula", "mostra o erro do servidor (ex.: voto duplicado) e mantém a cédula" | atendido |

### #18 Restrição de voto por inadimplência

| Critério | Backend | Endpoint | Tela | Testes | Situação |
|---|---|---|---|---|---|
| Consultar a adimplência antes de permitir a votação | `service/votar.go` (chama `voting.ValidarElegibilidade` antes de gravar), `service/votacoes.go` (`elegivel`, `motivo`) | `GET /api/votacoes/{id}` | `votar/VotarPauta.tsx` | `TestValidarElegibilidade` (voting); `TestDetalhar` (service) | atendido |
| Bloquear o voto de unidade inadimplente e avisar, inclusive quando enviado direto à API | `voting/elegibilidade.go` (`ErrInadimplente`), `service/votar.go`, `server/respostas.go` | `POST /api/votacoes/{id}/votos` → `403 inadimplente` com o motivo | `votar/VotarPauta.tsx`: "Voto bloqueado", sem botões de voto | `TestVotarPauta` "unidade inadimplente é bloqueada" (service); `TestListarEDetalharVotacoes` "inadimplente vê o motivo do bloqueio" (server); `Votar.test.tsx`: "bloqueia o inadimplente: mostra o peso, o motivo e nenhum botão de voto"; e2e-api "#17/#18 Pauta #02"; PW "#18 sofia (Apto 402 inadimplente) vê o motivo do bloqueio e não tem botões de voto" | atendido |
| Permitir ao administrador atualizar a situação das unidades | `service/unidades.go` (`AlterarAdimplencia`), `server/unidades.go`, `server/middleware.go` (`somenteAdmin`) | `GET /api/unidades`, `PATCH /api/unidades/{id}` | `admin/Unidades.tsx` | `TestUnidades` "marcar adimplente libera o voto" (service); `TestUsuariosEUnidades`, `TestRotasDeAdminRecusamEleitor` (server); `Unidades.test.tsx`: "marca a unidade inadimplente como adimplente e atualiza a linha"; e2e-api "#18 c3: regularizar unidade"; PW "#18 admin regulariza o Apto 402 e volta a marcá-lo como inadimplente" | atendido |

### #19 Verificação de quórum qualificado

| Critério | Backend | Endpoint | Tela | Testes | Situação |
|---|---|---|---|---|---|
| No encerramento, comparar o peso favorável com o peso total cadastrado | `service/resultado.go` (só depois de encerrada; peso total = unidades com responsável), `voting/pauta.go` | `GET /api/votacoes/{id}/resultado` (`409 votacao_nao_encerrada` antes) | `resultado/ResultadoDaPauta.tsx` (linha "Não votaram" incluída no total) | `TestResultadoPauta`, `TestResultadoExigeVotacaoEncerrada` (service); `TestApurarPauta` "Unidade ausente conta no peso total" (voting) | atendido (L3) |
| Aprovar somente com pelo menos 2/3 exatos do peso total | `voting/quorum.go` (`AtingiuQuorum`) | idem | idem | `TestAtingiuQuorumClassesEquivalencia`, `TestAtingiuQuorumValorLimite` "VL1: exatamente 2/3" (voting); `TestApurarPauta` "Aprovada com exatamente dois tercos"; `TestResultadoPauta` "maioria simples não basta: 300 de 500 é reprovada" (service) | atendido (L1) |
| Apresentar no relatório o estado de aprovação e os valores usados | `service/resultado.go` (`contagem`, `peso_total`, `limiar`, `percentual_favor`, `aprovada`, `unidades_total`, `unidades_votantes`) | idem | `resultado/ResultadoDaPauta.tsx`: veredito, limiar, pesos e % do total, nota sobre 2/3 do peso total incluindo quem não votou | `Resultado.test.tsx`: "reprovada com 60% a favor: mostra pesos, % do total, limiar e unidades", "aprovada quando o peso a favor atinge o limiar"; PW "#19 resultados públicos das pautas: 2/3 do peso total decide (sem login)" | atendido |

### #20 Votação secreta entre chapas concorrentes

| Critério | Backend | Endpoint | Tela | Testes | Situação |
|---|---|---|---|---|---|
| No máximo um voto por estudante autenticado e por turno, com peso unitário | `service/votar.go` (aptos; `participacoes` com chave `(votacao_id, usuario_id)`), `service/encerramento.go` (cada turno é uma votação), `voting/eleicao.go` (`PesoUnitario`) | `POST /api/votacoes/{id}/votos` (`401` sem login, `403 nao_apto`, `409 voto_duplicado`) | `votar/VotarEleicao.tsx` | `TestVotarEleicao` "segundo voto do mesmo eleitor", "eleitor fora da lista de aptos" (service); `TestEleicaoComSegundoTurno`, `TestRotasExigemLogin` (server); e2e-api "#20/#21/#22" | atendido |
| Armazenar a participação separada da cédula com a escolha | migração: `participacoes` × `cedulas` (sem eleitor, sem horário, UUID aleatório); `service/votar.go` (mesma transação); `store/votos.go` | idem | — | `TestColunasDasTabelasDeVotoSecreto`, `TestParticipacoesECedulas` (store); `TestVotarEleicao` "cédulas anônimas e participações gravadas em separado" (service) | atendido (L2) |
| Apresentar chapas homologadas e as opções branco e nulo | `service/votacoes.go` (`opcoes` = chapas), `service/votar.go` (branco e nulo como tipo de cédula) | `GET /api/votacoes/{id}` | `votar/VotarEleicao.tsx`: selo "Voto secreto", cartões das chapas com descrição, "Votar em branco", "Votar nulo" | `Votar.test.tsx`: "mostra o selo de voto secreto, as chapas com descrição, branco e nulo"; `TestDetalhar` (service) | atendido (chapas cadastradas pela administração são as homologadas; não há fluxo de homologação) |
| Não expor a associação estudante-escolha nas respostas, boletins ou logs | `server/votacoes.go` (resposta fixa), `server/middleware.go` (log só com método, caminho, status e duração), `server/dto.go` (boletim só com totais) | `POST .../votos` → `201 {"mensagem": "Voto registrado com sucesso."}` | `Votar.tsx`: tela "Voto registrado" sem a escolha | `TestVotacaoDaPautaDeCondominio` (resposta não repete a escolha), `TestEleicaoComSegundoTurno` (boletim sem nomes de eleitores nem horários), `TestRegistrarAcessoNaoLogaCorpoNemHeaders` (server); `Votar.test.tsx`: "confirma a chapa, envia o voto e a tela final não mostra a escolha"; PW "#20 ana e bruno votam no 2º turno e a tela final não mostra a chapa escolhida" | atendido (L2) |

### #21 Maioria absoluta e indicação de segundo turno

| Critério | Backend | Endpoint | Tela | Testes | Situação |
|---|---|---|---|---|---|
| Excluir brancos e nulos do total de válidos | `voting/eleicao.go` (`ApurarEleicao`: `Validos` soma só as chapas) | `GET /api/votacoes/{id}/resultado` (`validos`) | `resultado/BoletimDaEleicao.tsx` | `TestApurarEleicao` "Brancos e nulos não entram na base da maioria absoluta" (voting) | atendido |
| Declarar vencedora a chapa com mais da metade dos válidos | `voting/ca_eleicao.go` (`VerificarMaioriaAbsoluta`), `voting/eleicao.go` | idem (`vencedora`) | `resultado/BoletimDaEleicao.tsx` | `TestVerificarMaioriaAbsolutaClassesEquivalencia`, `TestVerificarMaioriaAbsolutaValorLimite`, `TestApurarEleicao` "Vencedora com um voto a mais que as demais somadas" (voting); `TestResultadoEleicao` (service); `Resultado.test.tsx`: "anuncia a chapa vencedora com maioria absoluta" | atendido |
| Sem maioria absoluta, preparar novo pleito com as duas chapas mais votadas | `voting/eleicao.go` (`SegundoTurno`), `service/encerramento.go` (`AbrirSegundoTurno`: as duas chapas, os mesmos aptos, uma única vez) | `GET .../resultado` (`segundo_turno`), `POST /api/votacoes/{id}/segundo-turno` | `resultado/BoletimDaEleicao.tsx`: "2º turno entre X e Y", formulário "Abrir 2º turno" (admin) e link para o 2º turno | `TestAbrirSegundoTurno`, `TestAbrirSegundoTurnoIndisponivel` (service); `TestEleicaoComSegundoTurno` (server); `Resultado.test.tsx`: "sem maioria absoluta anuncia o 2º turno entre as duas mais votadas", "admin abre o 2º turno com a data de término escolhida"; PW "#21/#22 boletim do 1º turno sem maioria absoluta e abertura do 2º turno pelo admin" | atendido (o sistema indica as chapas; a abertura é uma ação da administração) |
| Sem votos válidos, não declarar vencedora nem gerar 2º turno | `voting/ca_eleicao.go`, `voting/eleicao.go`, `service/encerramento.go` | `GET .../resultado`; `POST .../segundo-turno` → `409` | `resultado/BoletimDaEleicao.tsx`: "Sem votos válidos: sem vencedora" | `TestApurarEleicao` "Somente brancos e nulos…" (voting); `TestResultadoEleicao`, `TestAbrirSegundoTurnoIndisponivel` "eleição sem votos válidos" (service); `Resultado.test.tsx`: "sem votos válidos não há vencedora" | atendido |

### #22 Consulta pública ao boletim de apuração

| Critério | Backend | Endpoint | Tela | Testes | Situação |
|---|---|---|---|---|---|
| Disponibilizar o boletim somente após o encerramento | `service/resultado.go`; rota pública em `server/server.go` | `GET /api/votacoes/{id}/resultado` (sem login; `409 votacao_nao_encerrada` antes) | rota pública `/votacoes/:id/resultado` (`App.tsx`, `Resultado.tsx`) | `TestResultadoExigeVotacaoEncerrada` (service); `TestVotacaoDaPautaDeCondominio` "resultado só depois do encerramento" (server); `Resultado.test.tsx`: "votação ainda aberta mostra a mensagem do servidor"; `App.test.tsx`: "o resultado é público: funciona sem login e oferece o botão Entrar" | atendido |
| Informar aptos, participantes, abstenções e distribuição por opção | `voting/eleicao.go` (`MontarBoletim`), `service/resultado.go` | idem | `resultado/BoletimDaEleicao.tsx` | `TestMontarBoletim`, `TestBoletimDaEleicaoDoSeed` (voting); `Resultado.test.tsx`: "mostra comparecimento, votos por chapa, brancos, nulos e a base dos percentuais"; PW "#21/#22 boletins públicos dos dois turnos (sem login)" | atendido |
| Identificar os totais usados nos percentuais | `service/resultado.go` (`validos`, `aptos`) | idem | `resultado/BoletimDaEleicao.tsx`: "Percentuais calculados sobre N votos válidos (brancos e nulos excluídos)" e comparecimento sobre os aptos | mesmo teste do `Resultado.test.tsx` acima | atendido |
| Não expor escolhas individuais nem horários que associem eleitor e voto | `server/dto.go` (boletim sem horários e sem dados de eleitores), migração (cédula sem horário) | idem | idem | `TestEleicaoComSegundoTurno` (server); `TestColunasDasTabelasDeVotoSecreto` (store) | atendido (L2) |

### #23 Ordenação de preferências de projetos

| Critério | Backend | Endpoint | Tela | Testes | Situação |
|---|---|---|---|---|---|
| Ordenar todos os projetos sem repetir posição | — (interface) | — | `votar/VotarConsulta.tsx`: "Projetos a ordenar" → "Sua ordem", com subir, descer e remover | `Votar.test.tsx`: "subir, descer e remover reorganizam a ordem antes do envio"; PW "#23 lúcia ordena todos os projetos antes de enviar a cédula Borda" | atendido |
| Habilitar o envio só com todos ordenados | — (interface) | — | `votar/VotarConsulta.tsx`: "Enviar cédula" desabilitado até completar | `Votar.test.tsx`: "habilita o envio só quando todos os projetos estão ordenados" | atendido |
| Rejeitar no backend listas incompletas, duplicadas ou com projetos desconhecidos | `voting/borda.go` (`ValidarCedulaBorda`), `service/votar.go` | `POST /api/votacoes/{id}/votos` → `400 cedula_invalida` | `votar/VotarConsulta.tsx` mostra o erro | `TestValidarCedulaBorda` (voting); `TestVotarConsulta` (service: incompleta, repetida, desconhecida, a mais); `Votar.test.tsx`: "mostra o erro cedula_invalida vindo do servidor"; e2e-api "#23/#24/#25" | atendido |
| Registrar a lista em cédula anônima | migração: `cedulas` (tipo `ranking`) + `cedula_preferencias`; `service/votar.go` (mesma transação da participação) | idem | — | `TestVotarConsulta` "voto repetido não deixa cédula nem preferências extras" (service); `TestParticipacoesECedulas`, `TestColunasDasTabelasDeVotoSecreto` (store) | atendido (L2) |

### #24 Apuração por pontuação decrescente (Borda)

| Critério | Backend | Endpoint | Tela | Testes | Situação |
|---|---|---|---|---|---|
| N pontos para a 1ª posição, N-1 para a 2ª, até 1 para a última | `voting/borda.go` (`CalcularApuracaoBorda`) | `GET /api/votacoes/{id}/resultado` (`pontos`) | regra explicada em `votar/VotarConsulta.tsx` e na legenda de `resultado/RankingDaConsulta.tsx` | `TestCalcularApuracaoBorda` "Opção única recebe 1 ponto por cédula" e demais casos (voting) | atendido |
| Somar os pontos de todas as cédulas válidas | `voting/borda.go`, `service/resultado.go` | idem | `resultado/RankingDaConsulta.tsx` | `TestCalcularApuracaoBorda`, `TestConsultaDoSeed` (voting); `TestResultadoConsulta` (service) | atendido |
| Ordenar pela pontuação acumulada | `voting/borda.go` | idem | idem | `TestCalcularApuracaoBorda`; `Resultado.test.tsx`: "mostra pontos, 1ªs preferências, custo e custo acumulado de cada projeto" | atendido |
| Em empate de pontos, priorizar mais primeiras preferências | `voting/borda.go` (ordenação estável por pontos e primeiras preferências; empate persistente marcado) | idem (`primeiras_preferencias`, `empatado`) | `resultado/RankingDaConsulta.tsx`: marca "Empatado" | `TestCalcularApuracaoBorda` "Empate de pontos desfeito por primeiras preferências", "Empate persistente mantém a ordem de cadastro e marca Empatado"; `Resultado.test.tsx`: "marca projetos empatados e avisa empate na linha de corte" | atendido |

### #25 Visualização do ranking de prioridade orçamentária

| Critério | Backend | Endpoint | Tela | Testes | Situação |
|---|---|---|---|---|---|
| Projetos em ordem decrescente de pontuação, com pontos e custos | `service/resultado.go` (`ranking`: `pontos`, `custo_centavos`, `custo_acumulado_centavos`) | `GET /api/votacoes/{id}/resultado` | `resultado/RankingDaConsulta.tsx` | `TestConsultaComCorteOrcamentario` (server); `Resultado.test.tsx`: "mostra pontos, 1ªs preferências, custo e custo acumulado de cada projeto"; PW "#24/#25 ranking público com pontos, linha de corte, total e saldo (sem login)" | atendido |
| Exibir o corte orçamentário com orçamento disponível e custos acumulados | `voting/orcamento.go` (`AplicarCorteOrcamentario`) | idem (`dentro_do_orcamento`, `total_selecionado_centavos`, `saldo_centavos`, `empate_na_linha_de_corte`) | `resultado/RankingDaConsulta.tsx`: "Linha de corte do orçamento", orçamento, total selecionado, saldo e aviso de empate na linha | `TestAplicarCorteOrcamentario` "Corte estrito deixa fora um projeto mais barato abaixo da linha", "Empate na linha de corte é sinalizado" (voting); `TestResultadoConsulta` (service); `Resultado.test.tsx`: "mostra a linha de corte entre o último dentro e o primeiro fora do orçamento", "mostra orçamento, total selecionado, saldo e a legenda Borda" | atendido |

### Critérios comuns aos épicos

| Critério | Como é atendido | Testes | Situação |
|---|---|---|---|
| "Participantes, opções e períodos precisam ser configurados antes de receber votos" | `service/criar_votacao.go` grava votação, opções e aptos numa só transação; nenhuma rota altera opções ou aptos depois; voto só é aceito com a votação aberta (`service/votar.go`, `voting/periodo.go`). Tela `admin/NovaVotacao.tsx` | `TestCriarVotacaoValidacao`, `TestCriarVotacaoGravaOpcoesEAptos`, `TestVotarExigeVotacaoAberta` (service); `TestStatusVotacaoClassesEquivalencia`, `TestStatusVotacaoValorLimite` (voting); `NovaVotacao.test.tsx` | atendido |
| "Algoritmos isolados não encerram histórias de ponta a ponta" | toda função do domínio é alcançada por store → service → server → interface | `scripts/e2e-api.sh` (144 verificações) e Playwright (17 testes) contra o sistema completo no Docker | atendido |
| #14 "cálculo dos resultados explícito e reproduzível" | os resultados são determinísticos (Borda percorre as opções na ordem de cadastro e ordena de forma estável) e a tela mostra todos os valores usados (pesos, limiar, válidos, pontos, custos acumulados) | testes de resultado acima | atendido |

## Decisões de domínio

As issues deixavam estas decisões em aberto. Na v1.0.0 elas ficaram assim:

| Tema | Decisão | Onde |
|---|---|---|
| Responsável por várias unidades (#17) | cada usuário responde por no máximo uma unidade (`unidades.responsavel_id UNIQUE`); cadastrar uma segunda dá erro de validação | migração; `service/unidades.go`; `TestCriarUnidade` "responsável que já tem unidade" |
| "Peso total cadastrado" (#19) | soma das frações ideais das unidades que têm responsável (no seed, todas). Unidade sem responsável não vota e não entra no total; o `percentual` exibido usa a mesma base | `service/resultado.go`, `service/unidades.go` |
| Maioria simples e presença mínima (#19) | não implementadas; o único critério de aprovação da pauta é 2/3 do peso total | — |
| Voto da pauta (#17) | nominal por unidade, como numa assembleia; não é secreto | `votos_pauta(unidade_id, escolha)` |
| Empate que impede escolher as duas chapas do 2º turno (#21) | se o 2º colocado empata com o 3º, o sistema não escolhe: marca `empate_indefinido` e recusa abrir o 2º turno. A decisão fica com a comissão, fora do sistema | `voting/eleicao.go`; `TestApurarEleicao` "Empate pelo segundo lugar impede definir o segundo turno"; `Resultado.test.tsx`: "empate pelo 2º lugar: o sistema não escolhe e indica decisão da comissão" |
| Empate no 2º turno (#21) | sem vencedora; a tela indica decisão da comissão. Só o 1º turno leva a um 2º turno | `service/resultado.go`; `TestEmpateNoSegundoTurno`; `Resultado.test.tsx`: "empate no 2º turno indica decisão da comissão" |
| Abertura do 2º turno (#21) | ação da administração, uma única vez, com as duas chapas, os mesmos aptos e prazo escolhido (padrão 24 h) | `service/encerramento.go` |
| Empate persistente no Borda (#24) | mesmos pontos e mesmas primeiras preferências: não se inventa desempate; mantém a ordem de cadastro e marca `empatado = true` | `voting/borda.go` |
| Política de orçamento (#25) | linha de corte única: percorre o ranking somando custos; um projeto fica dentro se o acumulado (com ele) cabe e todos os anteriores ficaram dentro; o primeiro que não cabe e todos abaixo ficam fora, mesmo que um mais barato coubesse no saldo. Empate (pontos e primeiras preferências) entre o último dentro e o primeiro fora liga `empate_na_linha_de_corte` | `voting/orcamento.go` |
| Sigilo (#20, #22) | participação e cédula em tabelas separadas, gravadas na mesma transação; cédula sem eleitor, sem horário e com UUID aleatório. Substitui a "criptografia" citada na A1 (ver [arquitetura, D5](arquitetura.md#d5-sigilo-do-voto-pela-separação-entre-participação-e-cédula-nova-muda-a-a1)) | migração; `service/votar.go` |
| Papéis | `admin` (síndico, comissão eleitoral, organizador) e `eleitor`; admin não vota | migração (`CHECK` em `papel`); `server/middleware.go` |
| Aptos da eleição e da consulta | escolhidos pela administração ao criar; lista omitida ou vazia na API significa todos os eleitores (a tela sempre envia a lista marcada e exige ao menos um) | `service/criar_votacao.go`; `admin/NovaVotacao.tsx` |
| Valores | pesos em `float64` (frações do seed são áreas inteiras em m²); dinheiro em centavos `int64`, até R$ 1 trilhão por projeto e por orçamento | `voting`, `service/criar_votacao.go` |

## Fora do escopo e limitações

Fora do escopo da v1.0.0:

- cadastro de usuários e unidades pela API ou pela interface (vêm do seed de demonstração);
- edição ou exclusão de votações depois de criadas;
- papel Auditor, maioria simples, presença mínima, fluxo de homologação de chapas;
- responsável por mais de uma unidade;
- implantação pública com TLS (o projeto roda localmente com Docker Compose).

Limitações conhecidas:

- **L1, pesos em `float64`.** Com frações não inteiras, um voto que deveria dar exatamente 2/3 pode ficar logo
  abaixo (por exemplo, 0,6 de 0,9 vira 0,6666666666666665). Com as áreas inteiras do seed as somas são exatas.
  A solução completa seria guardar pesos inteiros.
- **L2, sigilo contra o administrador do banco.** A aplicação, a API, os boletins e os logs nunca ligam eleitor e
  cédula. Mas participação e cédula são gravadas na mesma transação, então compartilham o valor da coluna de sistema
  `xmin` e ficam próximas no WAL: quem administra o Postgres consegue reassociá-las. O banco é tratado como
  confiável e deve ficar privado (o compose publica a porta 5432 só em `127.0.0.1`).
- **L3, base do peso total.** "Peso total cadastrado" é a soma das unidades com responsável. Uma unidade cadastrada
  sem responsável não entra no denominador.
- No fim natural do prazo (sem encerramento manual), um voto que passou pela checagem milissegundos antes do `fim`
  ainda pode confirmar logo depois; um resultado consultado nesse intervalo pode mudar ao recarregar. O
  encerramento manual não tem esse intervalo.
- Consulta encerrada sem nenhuma cédula marca todos os projetos como empatados (todos com 0 pontos).
- O limite de tentativas de login fica em memória: zera quando a API reinicia e não é compartilhado entre réplicas.
  Atrás do nginx, todos os navegadores chegam à API com o mesmo IP (ver [README, Segurança](../README.md#segurança)).
