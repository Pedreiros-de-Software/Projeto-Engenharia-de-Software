#!/usr/bin/env bash
# Teste de ponta a ponta da API pelo nginx: percorre as 9 histórias com os usuários do seed de demonstração.
#
# Uso:   BASE_URL=http://localhost:3000 bash scripts/e2e-api.sh
# Exige: curl e jq; o sistema completo no ar (na raiz: cp .env.example .env na primeira vez e depois
#        docker compose up --build) com o seed RECÉM-CRIADO.
#        O script vota, encerra votações e cria outras: rode de novo só depois de recriar o banco
#        (docker compose down -v && docker compose up -d --build --wait).
# Saída: uma linha PASS/FAIL/SKIP por verificação e o total; código de saída 1 se alguma falhar.
#        SKIP marca verificação de campo que a API ainda não envia.
# E2E_PULAR_LIMITE=1 pula a verificação do rate limit do login (que espera 61 s pela janela de 1 minuto).
set -u

B=${BASE_URL:-http://localhost:3000}
B=${B%/}
for programa in curl jq; do
  command -v "$programa" >/dev/null || { echo "precisa de $programa no PATH" >&2; exit 2; }
done

PASS=0; FAIL=0; SKIP=0
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

req() { # metodo caminho token corpo -> STATUS, BODY
  local args=(-sS -o "$TMP/body" -D "$TMP/headers" -w '%{http_code}' -X "$1" "$B$2")
  [ -n "${3:-}" ] && args+=(-H "Authorization: Bearer $3")
  [ -n "${4:-}" ] && args+=(-H 'Content-Type: application/json' --data-binary "$4")
  STATUS=$(curl "${args[@]}"); BODY=$(cat "$TMP/body")
}
check() { # descricao esperado obtido
  if [ "$2" == "$3" ]; then PASS=$((PASS+1)); echo "PASS | $1 | $3";
  else FAIL=$((FAIL+1)); echo "FAIL | $1 | esperado=$2 | obtido=$3"; fi
}
check_se_presente() { # descricao campo esperado obtido: confere só quando BODY traz o campo
  if [ "$(echo "$BODY" | jq "has(\"$2\")")" == true ]; then check "$1" "$3" "$4";
  else SKIP=$((SKIP+1)); echo "SKIP | $1 | resposta sem o campo $2"; fi
}
header() { # descricao regex: confere um header da última resposta
  if grep -qiE "$2" "$TMP/headers"; then check "$1" sim sim; else check "$1" sim nao; fi
}
expect() { # descricao metodo caminho token corpo status [codigo]
  req "$2" "$3" "$4" "$5"
  local got="$STATUS"
  if [ -n "${7:-}" ]; then got="$STATUS $(echo "$BODY" | jq -r .codigo 2>/dev/null)"; check "$1" "$6 $7" "$got";
  else check "$1" "$6" "$got"; fi
}
j() { echo "$BODY" | jq -c "$1"; }
login() { # usuario senha -> TOKEN (e confere 200)
  req POST /api/login "" "{\"login\":\"$1\",\"senha\":\"$2\"}"; TOKEN=$(echo "$BODY" | jq -r .token)
  check "login $1" 200 "$STATUS"
}
utc() { jq -nr "(now + $1) | strftime(\"%Y-%m-%dT%H:%M:%SZ\")"; } # instante a $1 segundos de agora (GNU e BSD)

echo "== Aguardando $B/health"
for _ in $(seq 1 30); do curl -fsS "$B/health" >/dev/null 2>&1 && break; sleep 2; done
INI=$(utc -60); FIM=$(utc 86400)

echo "== Site (nginx)"
expect "GET / devolve a SPA" GET / "" "" 200
check "index.html tem o elemento raiz" sim "$(echo "$BODY" | grep -q '<div id="root">' && echo sim || echo nao)"
header "CSP no site" '^content-security-policy: .*default-src .self.'
header "X-Content-Type-Options nosniff" '^x-content-type-options: nosniff'
header "X-Frame-Options DENY" '^x-frame-options: deny'
header "Referrer-Policy no-referrer" '^referrer-policy: no-referrer'
expect "rota da SPA (/votacoes/1) devolve o index.html" GET /votacoes/1 "" "" 200
check "deep link tem o elemento raiz" sim "$(echo "$BODY" | grep -q '<div id="root">' && echo sim || echo nao)"
expect "GET /health pelo nginx" GET /health "" "" 200

echo "== Sessão"
expect "login senha errada" POST /api/login "" '{"login":"maria","senha":"errada"}' 401 credenciais_invalidas
login admin admin123; ADMIN=$TOKEN
check "login admin devolve usuario admin" '{"login":"admin","papel":"admin"}' "$(j '{login:.usuario.login,papel:.usuario.papel}')"
check "token tem 43 caracteres" 43 "${#ADMIN}"
login maria consensu; MARIA=$TOKEN; login joao consensu; JOAO=$TOKEN; login lucia consensu; LUCIA=$TOKEN
login sofia consensu; SOFIA=$TOKEN; login ana consensu; ANA=$TOKEN; login bruno consensu; BRUNO=$TOKEN
expect "sem token -> 401" GET /api/votacoes "" "" 401 nao_autenticado
expect "token inválido -> 401" GET /api/votacoes "tokenfalso" "" 401 nao_autenticado
expect "GET /api/me maria" GET /api/me "$MARIA" "" 200
check "maria: Apto 101, fração 60, 12%" '{"identificador":"Apto 101","fracao_ideal":60,"percentual":12,"adimplente":true}' "$(j '.unidade|{identificador,fracao_ideal,percentual,adimplente}')"
expect "GET /api/me ana (sem unidade)" GET /api/me "$ANA" "" 200
check "ana: unidade null" null "$(j .unidade)"
header "API envia Cache-Control no-store" '^cache-control: no-store'
header "CSP também nas respostas da API" '^content-security-policy:'

echo "== Painel"
expect "GET /api/votacoes admin" GET /api/votacoes "$ADMIN" "" 200
check "4 votações do seed (o seed precisa ser recém-criado)" 4 "$(j length)"
check "mais recentes primeiro (ids desc)" true "$(j '[.[].id] == ([.[].id]|sort|reverse)')"
TITULOS=$(j 'map({key:(.id|tostring),value:{titulo,descricao}})|from_entries')
P1=$(echo "$BODY" | jq '.[]|select(.titulo|startswith("Pauta #01")).id')
P2=$(echo "$BODY" | jq '.[]|select(.titulo|startswith("Pauta #02")).id')
EL=$(echo "$BODY" | jq '.[]|select(.tipo=="eleicao").id')
CO=$(echo "$BODY" | jq '.[]|select(.tipo=="consulta").id')
echo "   ids: pauta1=$P1 pauta2=$P2 eleicao=$EL consulta=$CO"
check "status do seed (P1,P2,EL,CO)" '["encerrada","aberta","encerrada","aberta"]' "$(echo "$BODY" | jq -c "[(.[]|select(.id==$P1).status),(.[]|select(.id==$P2).status),(.[]|select(.id==$EL).status),(.[]|select(.id==$CO).status)]")"
expect "GET /api/votacoes maria" GET /api/votacoes "$MARIA" "" 200
check "maria: pauta2 apto e não votou" '{"apto":true,"ja_votou":false}' "$(echo "$BODY" | jq -c ".[]|select(.id==$P2)|{apto,ja_votou}")"
check "maria: eleição não apta" false "$(echo "$BODY" | jq -c ".[]|select(.id==$EL).apto")"
check "maria: consulta já votou (seed)" true "$(echo "$BODY" | jq -c ".[]|select(.id==$CO).ja_votou")"

echo "== #17/#18 Pauta #02 (voto ponderado, inadimplência)"
expect "detalhe pauta2 maria" GET "/api/votacoes/$P2" "$MARIA" "" 200
check "#17 maria vê unidade, fração e % antes de votar" '{"identificador":"Apto 101","fracao_ideal":60,"percentual":12,"adimplente":true,"elegivel":true,"motivo":""}' "$(j .minha_unidade)"
check "pauta: opcoes [] e orcamento null" '[[],null]' "$(j '[.opcoes,.orcamento_centavos]')"
expect "detalhe pauta2 sofia" GET "/api/votacoes/$P2" "$SOFIA" "" 200
check "#18 sofia: inelegível com motivo" '{"elegivel":false,"motivo":"Unidade inadimplente: o voto fica bloqueado até a regularização com a administração."}' "$(j '.minha_unidade|{elegivel,motivo}')"
expect "#18 sofia vota -> 403 inadimplente" POST "/api/votacoes/$P2/votos" "$SOFIA" '{"escolha":"favor"}' 403 inadimplente
expect "ana (sem unidade) vota na pauta -> 403" POST "/api/votacoes/$P2/votos" "$ANA" '{"escolha":"favor"}' 403 nao_apto
expect "escolha inválida -> 400" POST "/api/votacoes/$P2/votos" "$MARIA" '{"escolha":"talvez"}' 400 requisicao_invalida
expect "campo desconhecido -> 400" POST "/api/votacoes/$P2/votos" "$MARIA" '{"escolha":"favor","x":1}' 400 requisicao_invalida
expect "JSON malformado -> 400" POST "/api/votacoes/$P2/votos" "$MARIA" '{"escolha":' 400 requisicao_invalida
expect "maria vota favor" POST "/api/votacoes/$P2/votos" "$MARIA" '{"escolha":"favor"}' 201
check "resposta não ecoa a escolha" '{"mensagem":"Voto registrado com sucesso."}' "$(j .)"
expect "joao vota favor" POST "/api/votacoes/$P2/votos" "$JOAO" '{"escolha":"favor"}' 201
expect "lucia vota favor" POST "/api/votacoes/$P2/votos" "$LUCIA" '{"escolha":"favor"}' 201
expect "maria de novo -> 409 duplicado" POST "/api/votacoes/$P2/votos" "$MARIA" '{"escolha":"contra"}' 409 voto_duplicado
expect "resultado antes de encerrar (público) -> 409" GET "/api/votacoes/$P2/resultado" "" "" 409 votacao_nao_encerrada
expect "eleitor encerra -> 403" POST "/api/votacoes/$P2/encerrar" "$MARIA" "" 403 acesso_negado
expect "sem token encerra -> 401" POST "/api/votacoes/$P2/encerrar" "" "" 401 nao_autenticado
expect "admin encerra pauta2" POST "/api/votacoes/$P2/encerrar" "$ADMIN" "" 200
check "encerrar devolve status encerrada" encerrada "$(echo "$BODY" | jq -r .status)"
expect "encerrar de novo -> 409" POST "/api/votacoes/$P2/encerrar" "$ADMIN" "" 409 votacao_nao_aberta
expect "votar após encerrar -> 409" POST "/api/votacoes/$P2/votos" "$SOFIA" '{"escolha":"favor"}' 409 votacao_nao_aberta

echo "== #19 Quórum qualificado"
expect "resultado pauta2 sem login" GET "/api/votacoes/$P2/resultado" "" "" 200
check "pauta2: 380/500, limiar 333.33, aprovada, 5 de 6" '{"favor":380,"contra":0,"abstencao":0,"peso_total":500,"limiar":333.33,"percentual_favor":76,"aprovada":true,"votantes":5,"total":6}' "$(j '{favor:.contagem.favor,contra:.contagem.contra,abstencao:.contagem.abstencao,peso_total,limiar:((.limiar*100|round)/100),percentual_favor,aprovada,votantes:.unidades_votantes,total:.unidades_total}')"
check_se_presente "#22 resultado público traz título e descrição da pauta" titulo "$(echo "$TITULOS" | jq -c ".\"$P2\"")" "$(j '{titulo,descricao}')"
expect "resultado pauta1 sem login" GET "/api/votacoes/$P1/resultado" "" "" 200
check "pauta1: 300/500 = 60% < 2/3 -> reprovada" '{"favor":300,"contra":80,"percentual_favor":60,"aprovada":false}' "$(j '{favor:.contagem.favor,contra:.contagem.contra,percentual_favor,aprovada}')"

echo "== #18 c3: regularizar unidade"
expect "eleitor lista unidades -> 403" GET /api/unidades "$MARIA" "" 403 acesso_negado
expect "admin lista unidades" GET /api/unidades "$ADMIN" "" 200
check "6 unidades, total 500, Apto 402 inadimplente" '[6,500,false]' "$(j '[length,(map(.fracao_ideal)|add),(.[]|select(.identificador=="Apto 402").adimplente)]')"
U402=$(echo "$BODY" | jq '.[]|select(.identificador=="Apto 402").id')
expect "eleitor PATCH unidade -> 403" PATCH "/api/unidades/$U402" "$SOFIA" '{"adimplente":true}' 403 acesso_negado
expect "PATCH sem campo -> 400" PATCH "/api/unidades/$U402" "$ADMIN" '{}' 400 requisicao_invalida
expect "PATCH unidade inexistente -> 404" PATCH /api/unidades/99999 "$ADMIN" '{"adimplente":true}' 404 nao_encontrado
expect "admin marca Apto 402 adimplente" PATCH "/api/unidades/$U402" "$ADMIN" '{"adimplente":true}' 200
check "Apto 402 agora adimplente" true "$(j .adimplente)"
expect "admin cria pauta nova" POST /api/votacoes "$ADMIN" "{\"tipo\":\"pauta\",\"titulo\":\"Pauta #03 — Teste\",\"descricao\":\"\",\"inicio\":\"$INI\",\"fim\":\"$FIM\",\"opcoes\":[]}" 201
P3=$(j .id)
expect "sofia regularizada vota na pauta nova" POST "/api/votacoes/$P3/votos" "$SOFIA" '{"escolha":"abstencao"}' 201
expect "admin volta Apto 402 a inadimplente" PATCH "/api/unidades/$U402" "$ADMIN" '{"adimplente":false}' 200

echo "== #20/#21/#22 Eleição, boletim e 2º turno"
expect "boletim turno 1 sem login" GET "/api/votacoes/$EL/resultado" "" "" 200
check "#22 boletim: aptos 8, participantes 7, abstenções 1, 87.5%" '[8,7,1,87.5]' "$(j '[.aptos,.participantes,.abstencoes,.percentual_comparecimento]')"
check "#21 votos por chapa e % dos válidos" '[["Chapa 1 — Movimenta",3,50],["Chapa 2 — Renovação",2,33.33],["Chapa 3 — Integração",1,16.67]]' "$(j '[.votos[]|[.nome,.votos,((.percentual_validos*100|round)/100)]]')"
check "brancos 1, nulos 0, válidos 6" '[1,0,6]' "$(j '[.brancos,.nulos,.validos]')"
check "#21 sem vencedora, 2º turno C1 x C2" '{"vencedora":null,"necessario":true,"empate":false,"chapas":["Chapa 1 — Movimenta","Chapa 2 — Renovação"],"votacao_id":null}' "$(j '{vencedora,necessario:.segundo_turno.necessario,empate:.segundo_turno.empate_indefinido,chapas:[.segundo_turno.chapas[].nome],votacao_id:.segundo_turno.votacao_id}')"
check "#20 boletim sem dados individuais" '[]' "$(j '[paths|map(tostring)|join(".")|select(test("usuario|login|eleitor|horario|criad"))]')"
check_se_presente "#22 boletim público traz título e descrição da eleição" titulo "$(echo "$TITULOS" | jq -c ".\"$EL\"")" "$(j '{titulo,descricao}')"
expect "eleitor abre 2º turno -> 403" POST "/api/votacoes/$EL/segundo-turno" "$ANA" '{}' 403 acesso_negado
expect "2º turno numa pauta -> 409" POST "/api/votacoes/$P1/segundo-turno" "$ADMIN" '{}' 409 segundo_turno_indisponivel
expect "2º turno com fim no passado -> 400" POST "/api/votacoes/$EL/segundo-turno" "$ADMIN" "{\"fim\":\"$INI\"}" 400 requisicao_invalida
expect "admin abre 2º turno (fim informado)" POST "/api/votacoes/$EL/segundo-turno" "$ADMIN" "{\"fim\":\"$FIM\"}" 201
T2=$(j .id)
check "turno 2: aberta, anterior=EL, 2 chapas" "{\"turno\":2,\"status\":\"aberta\",\"anterior\":$EL,\"chapas\":[\"Chapa 1 — Movimenta\",\"Chapa 2 — Renovação\"]}" "$(j '{turno,status,anterior:.votacao_anterior_id,chapas:[.opcoes[].nome]}')"
check "título com sufixo do 2º turno" true "$(j '.titulo|endswith("— 2º turno")')"
C2T2=$(j '.opcoes[1].id')
expect "abrir de novo -> 409" POST "/api/votacoes/$EL/segundo-turno" "$ADMIN" '{}' 409 segundo_turno_indisponivel
expect "detalhe turno 1 aponta o 2º turno" GET "/api/votacoes/$EL" "$ADMIN" "" 200
check "segundo_turno_id do turno 1" "$T2" "$(j .segundo_turno_id)"
C1T1=$(j '.opcoes[0].id')
expect "detalhe turno 2 para ana" GET "/api/votacoes/$T2" "$ANA" "" 200
check "ana apta no turno 2 e não votou" '{"apto":true,"ja_votou":false,"minha_unidade":null}' "$(j '{apto,ja_votou,minha_unidade}')"
expect "chapa do 1º turno no 2º -> 400" POST "/api/votacoes/$T2/votos" "$ANA" "{\"escolha\":\"chapa\",\"opcao_id\":$C1T1}" 400 opcao_desconhecida
expect "chapa sem opcao_id -> 400" POST "/api/votacoes/$T2/votos" "$ANA" '{"escolha":"chapa"}' 400
expect "eleição com preferencias -> 400" POST "/api/votacoes/$T2/votos" "$ANA" '{"preferencias":[1]}' 400 requisicao_invalida
expect "maria (não apta) vota -> 403" POST "/api/votacoes/$T2/votos" "$MARIA" '{"escolha":"branco"}' 403 nao_apto
expect "#20 ana vota Chapa 2" POST "/api/votacoes/$T2/votos" "$ANA" "{\"escolha\":\"chapa\",\"opcao_id\":$C2T2}" 201
check "#20 resposta não ecoa a chapa" '{"mensagem":"Voto registrado com sucesso."}' "$(j .)"
expect "bruno vota Chapa 2" POST "/api/votacoes/$T2/votos" "$BRUNO" "{\"escolha\":\"chapa\",\"opcao_id\":$C2T2}" 201
expect "ana de novo -> 409" POST "/api/votacoes/$T2/votos" "$ANA" '{"escolha":"nulo"}' 409 voto_duplicado
expect "boletim turno 2 antes do fim -> 409" GET "/api/votacoes/$T2/resultado" "" "" 409 votacao_nao_encerrada
expect "admin encerra turno 2" POST "/api/votacoes/$T2/encerrar" "$ADMIN" "" 200
expect "boletim turno 2 sem login" GET "/api/votacoes/$T2/resultado" "" "" 200
check "turno 2: vencedora Chapa 2, 2 votos, 100%" '{"turno":2,"vencedora":"Chapa 2 — Renovação","participantes":2,"aptos":8,"chapa2":[2,100],"necessario":false}' "$(j '{turno,vencedora:.vencedora.nome,participantes,aptos,chapa2:(.votos[]|select(.nome|startswith("Chapa 2"))|[.votos,.percentual_validos]),necessario:.segundo_turno.necessario}')"
check_se_presente "#22 boletim do 2º turno traz o título com sufixo" titulo true "$(j '.titulo|endswith("— 2º turno")')"
expect "boletim turno 1 agora com link" GET "/api/votacoes/$EL/resultado" "" "" 200
check "segundo_turno.votacao_id = T2" "$T2" "$(j .segundo_turno.votacao_id)"

echo "== #23/#24/#25 Consulta (Borda e corte orçamentário)"
expect "detalhe consulta lucia" GET "/api/votacoes/$CO" "$LUCIA" "" 200
check "4 projetos com custo, orçamento R\$ 60.000,00" '[4,6000000,[3500000,1200000,4800000,900000]]' "$(j '[(.opcoes|length),.orcamento_centavos,[.opcoes[].custo_centavos]]')"
I=$(j '.opcoes[]|select(.nome|startswith("Impressoras")).id'); A=$(j '.opcoes[]|select(.nome|startswith("Kit Arduino")).id')
G=$(j '.opcoes[]|select(.nome|startswith("Servidor GPU")).id'); M=$(j '.opcoes[]|select(.nome|startswith("Mobili")).id')
expect "#23 cédula incompleta -> 400" POST "/api/votacoes/$CO/votos" "$LUCIA" "{\"preferencias\":[$A,$G,$I]}" 400 cedula_invalida
expect "#23 cédula com repetido -> 400" POST "/api/votacoes/$CO/votos" "$LUCIA" "{\"preferencias\":[$A,$A,$G,$I]}" 400 cedula_invalida
expect "#23 cédula com id desconhecido -> 400" POST "/api/votacoes/$CO/votos" "$LUCIA" "{\"preferencias\":[$A,$G,$I,999999]}" 400 cedula_invalida
expect "consulta com escolha -> 400" POST "/api/votacoes/$CO/votos" "$LUCIA" '{"escolha":"favor"}' 400 requisicao_invalida
expect "maria (já votou no seed) -> 409" POST "/api/votacoes/$CO/votos" "$MARIA" "{\"preferencias\":[$A,$G,$I,$M]}" 409 voto_duplicado
expect "#23 lucia envia [Arduino, GPU, Impressoras, Mobiliário]" POST "/api/votacoes/$CO/votos" "$LUCIA" "{\"preferencias\":[$A,$G,$I,$M]}" 201
expect "ranking antes de encerrar -> 409" GET "/api/votacoes/$CO/resultado" "" "" 409 votacao_nao_encerrada
expect "admin encerra consulta" POST "/api/votacoes/$CO/encerrar" "$ADMIN" "" 200
expect "ranking sem login" GET "/api/votacoes/$CO/resultado" "" "" 200
check "#24 pontos Borda 18/18/15/9, 1ªs pref 2/2/2/0" '[["Impressoras 3D e CNC",18,2],["Kit Arduino e Robótica",18,2],["Servidor GPU para IA",15,2],["Mobiliário colaborativo",9,0]]' "$(j '[.ranking[]|[.nome,.pontos,.primeiras_preferencias]]')"
check "#24 empate persistente marcado e ordem de cadastro" '[true,true,false,false]' "$(j '[.ranking[].empatado]')"
check "#25 corte: dentro/fora e custo acumulado" '[[true,3500000],[true,4700000],[false,9500000],[false,10400000]]' "$(j '[.ranking[]|[.dentro_do_orcamento,.custo_acumulado_centavos]]')"
check "#25 cédulas 6, total 47000_00, saldo 13000_00, sem empate no corte" '[6,6000000,4700000,1300000,false]' "$(j '[.cedulas,.orcamento_centavos,.total_selecionado_centavos,.saldo_centavos,.empate_na_linha_de_corte]')"
check_se_presente "#22 ranking público traz título e descrição da consulta" titulo "$(echo "$TITULOS" | jq -c ".\"$CO\"")" "$(j '{titulo,descricao}')"

echo "== Admin: criar votações e validações"
expect "eleitor lista usuários -> 403" GET /api/usuarios "$MARIA" "" 403 acesso_negado
expect "admin lista usuários" GET /api/usuarios "$ADMIN" "" 200
check "15 usuários, nenhum campo de senha" '[15,["id","login","nome","papel"]]' "$(j '[length,(.[0]|keys)]')"
ESTUD=$(echo "$BODY" | jq -c '[.[]|select(.login|IN("ana","bruno","carla")).id]')
expect "eleitor cria votação -> 403" POST /api/votacoes "$MARIA" "{\"tipo\":\"pauta\",\"titulo\":\"x\",\"descricao\":\"\",\"inicio\":\"$INI\",\"fim\":\"$FIM\",\"opcoes\":[]}" 403 acesso_negado
expect "fim antes do início -> 400" POST /api/votacoes "$ADMIN" "{\"tipo\":\"pauta\",\"titulo\":\"x\",\"descricao\":\"\",\"inicio\":\"$FIM\",\"fim\":\"$INI\",\"opcoes\":[]}" 400 requisicao_invalida
expect "eleição com 1 chapa -> 400" POST /api/votacoes "$ADMIN" "{\"tipo\":\"eleicao\",\"titulo\":\"x\",\"descricao\":\"\",\"inicio\":\"$INI\",\"fim\":\"$FIM\",\"opcoes\":[{\"nome\":\"A\",\"descricao\":\"\"}]}" 400 requisicao_invalida
expect "consulta sem orçamento -> 400" POST /api/votacoes "$ADMIN" "{\"tipo\":\"consulta\",\"titulo\":\"x\",\"descricao\":\"\",\"inicio\":\"$INI\",\"fim\":\"$FIM\",\"opcoes\":[{\"nome\":\"A\",\"descricao\":\"\",\"custo_centavos\":1},{\"nome\":\"B\",\"descricao\":\"\",\"custo_centavos\":2}]}" 400 requisicao_invalida
expect "admin cria eleição com 3 aptos (formato do frontend, ms)" POST /api/votacoes "$ADMIN" "{\"tipo\":\"eleicao\",\"titulo\":\"Eleição teste\",\"descricao\":\"d\",\"inicio\":\"${INI%Z}.000Z\",\"fim\":\"${FIM%Z}.000Z\",\"opcoes\":[{\"nome\":\"A\",\"descricao\":\"\"},{\"nome\":\"B\",\"descricao\":\"\"}],\"aptos\":$ESTUD}" 201
NE=$(j .id)
expect "bruno (apto) vota branco na eleição nova" POST "/api/votacoes/$NE/votos" "$BRUNO" '{"escolha":"branco"}' 201
expect "maria (fora dos aptos) -> 403" POST "/api/votacoes/$NE/votos" "$MARIA" '{"escolha":"branco"}' 403 nao_apto
expect "admin cria consulta" POST /api/votacoes "$ADMIN" "{\"tipo\":\"consulta\",\"titulo\":\"Consulta teste\",\"descricao\":\"\",\"inicio\":\"$INI\",\"fim\":\"$FIM\",\"opcoes\":[{\"nome\":\"A\",\"descricao\":\"\",\"custo_centavos\":100},{\"nome\":\"B\",\"descricao\":\"\",\"custo_centavos\":200}],\"orcamento_centavos\":250}" 201
check "consulta criada com custos e orçamento" '[[100,200],250,"consulta"]' "$(j '[[.opcoes[].custo_centavos],.orcamento_centavos,.tipo]')"
expect "admin não vota em eleição (não apto)" POST "/api/votacoes/$NE/votos" "$ADMIN" '{"escolha":"branco"}' 403 nao_apto

echo "== Diversos"
expect "votação inexistente -> 404" GET /api/votacoes/99999 "$ADMIN" "" 404 nao_encontrado
expect "id não numérico -> 404" GET /api/votacoes/abc "$ADMIN" "" 404 nao_encontrado
expect "resultado inexistente -> 404" GET /api/votacoes/99999/resultado "" "" 404 nao_encontrado
BIG=$(head -c 70000 /dev/zero | tr '\0' 'a')
expect "corpo > 64 KiB -> 400" POST /api/votacoes "$ADMIN" "{\"tipo\":\"pauta\",\"titulo\":\"$BIG\"}" 400 requisicao_invalida
expect "método errado em /health -> 405" POST /health "" "" 405
expect "logout" POST /api/logout "$JOAO" "" 204
expect "token após logout -> 401" GET /api/me "$JOAO" "" 401 nao_autenticado

if [ "${E2E_PULAR_LIMITE:-}" == 1 ]; then
  SKIP=$((SKIP+1)); echo "SKIP | rate limit do login | E2E_PULAR_LIMITE=1"
else
  echo "== Rate limit do login (espera 61 s para começar numa janela nova)"
  sleep 61
  CODES=""
  for _ in $(seq 1 11); do req POST /api/login "" '{"login":"maria","senha":"errada"}'; CODES="$CODES$STATUS "; done
  check "10 tentativas 401, a 11ª 429" "401 401 401 401 401 401 401 401 401 401 429 " "$CODES"
  check "código muitas_tentativas" muitas_tentativas "$(echo "$BODY" | jq -r .codigo)"
fi

echo
echo "TOTAL: $PASS PASS, $FAIL FAIL, $SKIP SKIP"
[ "$FAIL" -eq 0 ]
