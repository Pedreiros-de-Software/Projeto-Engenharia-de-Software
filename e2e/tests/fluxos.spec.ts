import { expect, test, type Browser, type BrowserContextOptions, type Page } from "@playwright/test";

// Roteiro da demonstração sobre o seed recém-criado; cada teste depende do estado deixado pelos anteriores.
test.describe.configure({ mode: "serial" });

const SENHA_ELEITOR = "consensu";
const paginas = new Map<string, Page>();
const errosNoConsole: string[] = [];
const resultados = new Map<string, string>();

async function abrirPagina(browser: Browser, nome: string, opcoes: BrowserContextOptions = {}): Promise<Page> {
  const { baseURL, locale, timezoneId } = test.info().project.use;
  const contexto = await browser.newContext({
    baseURL,
    locale,
    timezoneId,
    viewport: { width: 1280, height: 900 },
    ...opcoes,
  });
  const pagina = await contexto.newPage();
  pagina.on("console", (mensagem) => {
    // Respostas 4xx esperadas (ex.: senha errada) também viram "Failed to load resource" no console.
    if (mensagem.type() === "error" && !mensagem.text().startsWith("Failed to load resource")) {
      errosNoConsole.push(`[${nome}] ${mensagem.text()}`);
    }
  });
  pagina.on("pageerror", (erro) => errosNoConsole.push(`[${nome}] ${erro.message}`));
  paginas.set(nome, pagina);
  return pagina;
}

async function anonimo(browser: Browser): Promise<Page> {
  return paginas.get("anonimo") ?? abrirPagina(browser, "anonimo");
}

/** Página do usuário já logado; o login acontece uma vez por usuário (a API limita tentativas por minuto). */
async function sessao(browser: Browser, login: string, senha = SENHA_ELEITOR): Promise<Page> {
  const existente = paginas.get(login);
  if (existente) return existente;
  const pagina = await abrirPagina(browser, login);
  await pagina.goto("/login");
  await pagina.getByLabel("Usuário").fill(login);
  await pagina.getByLabel("Senha").fill(senha);
  await pagina.getByRole("button", { name: "Entrar" }).click();
  await expect(pagina.getByRole("button", { name: "Sair" })).toBeVisible();
  await expect(pagina).not.toHaveURL(/\/login$/);
  return pagina;
}

async function irParaPainel(pagina: Page) {
  await pagina.goto("/");
  await expect(pagina.locator("article.cartao-votacao").first()).toBeVisible();
}

function cartao(pagina: Page, titulo: RegExp) {
  return pagina.getByRole("article", { name: titulo });
}

/** Valor para um campo datetime-local, no fuso do navegador, a `deslocamentoMs` de agora. */
async function dataLocal(pagina: Page, deslocamentoMs: number): Promise<string> {
  return pagina.evaluate((ms) => {
    const instante = new Date(Date.now() + ms);
    const doisDigitos = (n: number) => String(n).padStart(2, "0");
    const dia = `${instante.getFullYear()}-${doisDigitos(instante.getMonth() + 1)}-${doisDigitos(instante.getDate())}`;
    return `${dia}T${doisDigitos(instante.getHours())}:${doisDigitos(instante.getMinutes())}`;
  }, deslocamentoMs);
}

/** Abre a página de resultado e confere que o título exibido é o que a API mandou no resultado. */
async function abrirResultado(pagina: Page, caminho: string) {
  const id = caminho.split("/")[2];
  const resposta = pagina.waitForResponse((r) => r.url().endsWith(`/api/votacoes/${id}/resultado`));
  await pagina.goto(caminho);
  const corpo = (await (await resposta).json()) as { titulo?: string };
  await expect(pagina.getByRole("heading", { level: 1 })).toHaveText(corpo.titulo || `Votação nº ${id}`);
}

test.afterEach(async ({}, testInfo) => {
  if (testInfo.status === testInfo.expectedStatus) return;
  for (const [nome, pagina] of paginas) {
    if (!pagina.isClosed()) {
      await testInfo.attach(`tela-${nome}`, { body: await pagina.screenshot({ fullPage: true }), contentType: "image/png" });
    }
  }
});

test.afterAll(async () => {
  for (const pagina of paginas.values()) await pagina.context().close();
});

test("rota protegida leva ao login e senha errada mostra a mensagem da API", async ({ browser }) => {
  const pagina = await anonimo(browser);
  await pagina.goto("/");
  await expect(pagina).toHaveURL(/\/login$/);

  await pagina.getByLabel("Usuário").fill("maria");
  await pagina.getByLabel("Senha").fill("errada");
  await pagina.getByRole("button", { name: "Entrar" }).click();

  await expect(pagina.getByText("Usuário ou senha inválidos.")).toBeVisible();
});

test("painel do admin agrupa por contexto e oferece Encerrar agora nas votações abertas", async ({ browser }) => {
  const admin = await sessao(browser, "admin", "admin123");
  await irParaPainel(admin);

  await expect(
    admin.getByRole("heading", { level: 3, name: /Eleições de CA.*Condomínio.*Priorização de projetos/ }),
  ).toBeVisible();
  for (const grupo of ["Assembleia de condomínio", "Eleição do CA", "Priorização de projetos"]) {
    await expect(admin.getByRole("heading", { level: 2, name: grupo })).toBeVisible();
  }
  await expect(admin.getByRole("button", { name: "Encerrar agora" })).toHaveCount(2);
  await expect(admin.getByRole("banner")).toContainText("Administração");
  await expect(admin.getByRole("link", { name: "Unidades" })).toBeVisible();
  await expect(cartao(admin, /^Eleição Diretoria DCE 2026$/).getByText("Não apto")).toHaveCount(0);
});

test("#17 maria vê unidade, fração e peso antes de confirmar e aprova a Pauta #02", async ({ browser }) => {
  const maria = await sessao(browser, "maria");
  await irParaPainel(maria);
  await expect(cartao(maria, /^Eleição Diretoria DCE 2026$/).getByText("Não apto")).toBeVisible();

  await cartao(maria, /^Pauta #02/).getByRole("link", { name: "Votar" }).click();
  const unidade = maria.getByRole("region", { name: "Sua unidade" });
  await expect(unidade).toContainText("Apto 101");
  await expect(unidade).toContainText("60");
  await expect(unidade).toContainText(/12,00\s?% do total/);

  await maria.getByRole("button", { name: "Aprovar" }).click();
  const dialogo = maria.getByRole("dialog", { name: "Confirmar voto" });
  await expect(dialogo).toContainText(/Seu voto: APROVAR pela unidade Apto 101, com peso de 12,00\s?% do total/);
  const confirmar = dialogo.getByRole("button", { name: "Confirmar: Aprovar" });
  await expect(confirmar).toBeFocused();
  await maria.keyboard.press("Tab");
  await expect(dialogo.getByRole("button", { name: "Cancelar" })).toBeFocused();
  await maria.keyboard.press("Tab");
  await expect(confirmar).toBeFocused();

  await confirmar.click();
  await expect(maria.getByRole("heading", { name: "Voto registrado" })).toBeVisible();
});

test("joão e lúcia também aprovam a Pauta #02 (380 de 500 ≥ 2/3)", async ({ browser }) => {
  for (const login of ["joao", "lucia"]) {
    const pagina = await sessao(browser, login);
    await irParaPainel(pagina);
    await cartao(pagina, /^Pauta #02/).getByRole("link", { name: "Votar" }).click();
    await pagina.getByRole("button", { name: "Aprovar" }).click();
    await pagina.getByRole("dialog").getByRole("button", { name: "Confirmar: Aprovar" }).click();
    await expect(pagina.getByRole("heading", { name: "Voto registrado" })).toBeVisible();
  }
});

test("#18 sofia (Apto 402 inadimplente) vê o motivo do bloqueio e não tem botões de voto", async ({ browser }) => {
  const sofia = await sessao(browser, "sofia");
  await irParaPainel(sofia);
  await cartao(sofia, /^Pauta #02/).getByRole("link", { name: "Votar" }).click();

  await expect(sofia.getByText("Voto bloqueado")).toBeVisible();
  await expect(sofia.locator("main")).toContainText("Unidade inadimplente");
  await expect(sofia.getByRole("button", { name: /^(Aprovar|Rejeitar|Abster-se)$/ })).toHaveCount(0);
});

test("#21/#22 boletim do 1º turno sem maioria absoluta e abertura do 2º turno pelo admin", async ({ browser }) => {
  const admin = await sessao(browser, "admin", "admin123");
  await irParaPainel(admin);
  await cartao(admin, /^Eleição Diretoria DCE 2026$/).getByRole("link", { name: "Ver resultado" }).click();

  await expect(admin.getByText("2º turno entre Chapa 1 — Movimenta e Chapa 2 — Renovação")).toBeVisible();
  await expect(admin.locator("main")).toContainText(/Comparecimento de 87,50\s?%/);
  await expect(admin.locator("main")).toContainText("Percentuais calculados sobre 6 votos válidos");

  await admin.getByLabel("Término do 2º turno").fill(await dataLocal(admin, 24 * 3600 * 1000));
  await admin.getByRole("button", { name: "Abrir 2º turno" }).click();
  await expect(admin.getByText("2º turno aberto")).toBeVisible();

  await admin.getByRole("link", { name: "Ir para o 2º turno" }).click();
  await expect(admin.getByRole("heading", { level: 1, name: /— 2º turno$/ })).toBeVisible();
  await expect(admin.getByText("Acompanhamento")).toBeVisible();
  await expect(admin.getByText("Você não está apto")).toHaveCount(0);
});

test("#20 ana e bruno votam no 2º turno e a tela final não mostra a chapa escolhida", async ({ browser }) => {
  for (const login of ["ana", "bruno"]) {
    const eleitor = await sessao(browser, login);
    await irParaPainel(eleitor);
    await cartao(eleitor, /— 2º turno$/).getByRole("link", { name: "Votar" }).click();

    await expect(eleitor.getByText("Voto secreto", { exact: true })).toBeVisible();
    await expect(eleitor.locator(".cartao-chapa")).toHaveCount(2);
    await expect(eleitor.getByRole("button", { name: "Votar em branco" })).toBeVisible();
    await expect(eleitor.getByRole("button", { name: "Votar nulo" })).toBeVisible();

    await eleitor.getByRole("button", { name: /Chapa 2 — Renovação/ }).click();
    await eleitor.getByRole("dialog").getByRole("button", { name: "Confirmar voto" }).click();
    await expect(eleitor.getByRole("heading", { name: "Voto registrado" })).toBeVisible();
    await expect(eleitor.locator("main")).not.toContainText(/Renovação|Movimenta|Chapa/);
  }
});

test("#23 lúcia ordena todos os projetos antes de enviar a cédula Borda", async ({ browser }) => {
  const lucia = await sessao(browser, "lucia");
  await irParaPainel(lucia);
  await cartao(lucia, /^Orçamento participativo/).getByRole("link", { name: "Votar" }).click();
  await expect(lucia.getByRole("heading", { name: "Cédula de priorização" })).toBeVisible();

  const enviar = lucia.getByRole("button", { name: "Enviar cédula" });
  const aOrdenar = lucia.getByRole("region", { name: "Projetos a ordenar" });
  await expect(enviar).toBeDisabled();
  await aOrdenar.getByRole("button", { name: /Servidor GPU/ }).click();
  await aOrdenar.getByRole("button", { name: /Kit Arduino/ }).click();
  await aOrdenar.getByRole("button", { name: /Impressoras/ }).click();
  await expect(enviar).toBeDisabled();
  await aOrdenar.getByRole("button", { name: /Mobiliário/ }).click();
  await lucia.getByRole("button", { name: "Subir Kit Arduino e Robótica" }).click();
  await expect(enviar).toBeEnabled();
  await expect(lucia.getByRole("region", { name: "Sua ordem" }).getByRole("listitem").first()).toContainText(
    "Kit Arduino",
  );

  await enviar.click();
  const dialogo = lucia.getByRole("dialog", { name: "Confirmar cédula" });
  await expect(dialogo.getByRole("listitem")).toHaveText([
    "Kit Arduino e Robótica",
    "Servidor GPU para IA",
    "Impressoras 3D e CNC",
    "Mobiliário colaborativo",
  ]);
  await dialogo.getByRole("button", { name: "Enviar cédula" }).click();
  await expect(lucia.getByRole("heading", { name: "Voto registrado" })).toBeVisible();
});

test("admin encerra a Pauta #02, o 2º turno e a consulta, sempre com confirmação", async ({ browser }) => {
  const admin = await sessao(browser, "admin", "admin123");
  await irParaPainel(admin);

  for (const titulo of [/^Pauta #02/, /— 2º turno$/, /^Orçamento participativo/]) {
    await cartao(admin, titulo).getByRole("button", { name: "Encerrar agora" }).click();
    await admin.getByRole("dialog", { name: "Encerrar votação agora?" }).getByRole("button", { name: "Encerrar votação" }).click();
    await expect(cartao(admin, titulo).getByRole("link", { name: "Ver resultado" })).toBeVisible();
  }
  await expect(admin.getByRole("button", { name: "Encerrar agora" })).toHaveCount(0);

  for (const [chave, titulo] of [
    ["pauta1", /^Pauta #01/],
    ["pauta2", /^Pauta #02/],
    ["eleicao", /^Eleição Diretoria DCE 2026$/],
    ["consulta", /^Orçamento participativo/],
  ] as const) {
    const href = await cartao(admin, titulo).getByRole("link", { name: "Ver resultado" }).getAttribute("href");
    resultados.set(chave, href ?? "");
  }
});

test("#19 resultados públicos das pautas: 2/3 do peso total decide (sem login)", async ({ browser }) => {
  const pagina = await anonimo(browser);

  await abrirResultado(pagina, resultados.get("pauta2") ?? "/votacoes/2/resultado");
  await expect(pagina.getByText("APROVADA", { exact: true })).toBeVisible();
  await expect(pagina.locator("main")).toContainText(/unidades que não votaram/i);

  await abrirResultado(pagina, resultados.get("pauta1") ?? "/votacoes/1/resultado");
  await expect(pagina.getByText("REPROVADA", { exact: true })).toBeVisible();
  await expect(pagina.locator("main")).toContainText(/60,00\s?% do peso total a favor/);
});

test("#21/#22 boletins públicos dos dois turnos (sem login)", async ({ browser }) => {
  const pagina = await anonimo(browser);

  await abrirResultado(pagina, resultados.get("eleicao") ?? "/votacoes/3/resultado");
  await expect(pagina.getByText("2º turno entre Chapa 1 — Movimenta e Chapa 2 — Renovação")).toBeVisible();
  await expect(pagina.getByRole("button", { name: "Abrir 2º turno" })).toHaveCount(0);

  await pagina.getByRole("link", { name: "Ir para o 2º turno" }).click();
  await expect(pagina).toHaveURL(/\/votacoes\/\d+\/resultado$/);
  await expect(pagina.getByText("Chapa vencedora")).toBeVisible();
  await expect(pagina.locator(".veredito")).toContainText("Chapa 2 — Renovação");
  await expect(pagina.getByRole("heading", { level: 1 })).not.toHaveText("Eleição Diretoria DCE 2026");
});

test("#24/#25 ranking público com pontos, linha de corte, total e saldo (sem login)", async ({ browser }) => {
  const pagina = await anonimo(browser);

  await abrirResultado(pagina, resultados.get("consulta") ?? "/votacoes/4/resultado");
  const ranking = pagina.getByRole("table", { name: "Ranking pela pontuação Borda" });
  await expect(ranking.getByText("Linha de corte do orçamento")).toBeVisible();
  await expect(ranking.getByText("Empatado")).toHaveCount(2);
  await expect(ranking.getByText("Dentro do orçamento")).toHaveCount(2);
  await expect(ranking.getByText("Fora do orçamento")).toHaveCount(2);
  await expect(pagina.locator("main")).toContainText(/R\$\s47\.000,00/);
  await expect(pagina.locator("main")).toContainText(/R\$\s13\.000,00/);
});

test("tabelas largas rolam dentro do próprio quadro num celular, sem alargar a página", async ({ browser }) => {
  const celular = await abrirPagina(browser, "celular", {
    viewport: { width: 390, height: 844 },
    isMobile: true,
    hasTouch: true,
  });
  await celular.goto(resultados.get("consulta") ?? "/votacoes/4/resultado");
  const rolagem = celular.getByRole("region", { name: "Ranking pela pontuação Borda" });
  await expect(rolagem).toBeVisible();

  const larguras = await celular.evaluate(() => document.documentElement.scrollWidth);
  expect(larguras).toBeLessThanOrEqual(390);
  const quadro = await rolagem.evaluate((e) => ({ conteudo: e.scrollWidth, visivel: e.clientWidth }));
  expect(quadro.conteudo).toBeGreaterThan(quadro.visivel);

  const admin = await sessao(browser, "admin", "admin123");
  await admin.setViewportSize({ width: 390, height: 844 });
  await admin.goto("/admin/unidades");
  await expect(admin.getByRole("region", { name: "Unidades cadastradas" })).toBeVisible();
  expect(await admin.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(390);
  await admin.setViewportSize({ width: 1280, height: 900 });
});

test("#18 admin regulariza o Apto 402 e volta a marcá-lo como inadimplente", async ({ browser }) => {
  const admin = await sessao(browser, "admin", "admin123");
  await admin.goto("/admin/unidades");

  await admin.getByRole("button", { name: "Marcar Apto 402 como adimplente" }).click();
  await expect(admin.getByText("Unidade Apto 402 marcada como adimplente.")).toBeVisible();
  await admin.getByRole("button", { name: "Marcar Apto 402 como inadimplente" }).click();
  await expect(admin.getByText("Unidade Apto 402 marcada como inadimplente.")).toBeVisible();
});

test("admin cria uma consulta digitando os valores no formato brasileiro", async ({ browser }) => {
  const admin = await sessao(browser, "admin", "admin123");
  await admin.goto("/admin/nova-votacao");

  await admin.getByLabel("Tipo de votação").selectOption("consulta");
  await admin.getByLabel("Título").fill("Priorização — Biblioteca 2027");
  await admin.locator("#descricao").fill("Projetos para a biblioteca.");
  await admin.getByLabel("Início").fill(await dataLocal(admin, -60 * 1000));
  await admin.getByLabel("Fim").fill(await dataLocal(admin, 3 * 24 * 3600 * 1000));
  const nomes = admin.getByLabel("Nome", { exact: true });
  const custos = admin.getByLabel("Custo (R$)");
  await nomes.nth(0).fill("Novos computadores");
  await custos.nth(0).fill("20.000");
  await nomes.nth(1).fill("Climatização");
  await custos.nth(1).fill("15.000,50");
  await admin.getByLabel("Orçamento disponível (R$)").fill("30.000");

  await expect(admin.getByText(/^Custo: R\$\s20\.000,00$/)).toBeVisible();
  await expect(admin.getByText(/^Custo: R\$\s15\.000,50$/)).toBeVisible();
  await expect(admin.getByText(/^Orçamento: R\$\s30\.000,00$/)).toBeVisible();

  const pedido = admin.waitForRequest((r) => r.method() === "POST" && new URL(r.url()).pathname === "/api/votacoes");
  await admin.getByRole("button", { name: "Criar votação" }).click();
  const corpo = (await pedido).postDataJSON() as { opcoes: { custo_centavos: number }[]; orcamento_centavos: number };
  expect(corpo.opcoes.map((o) => o.custo_centavos)).toEqual([2_000_000, 1_500_050]);
  expect(corpo.orcamento_centavos).toBe(3_000_000);

  await expect(admin.getByText('Votação "Priorização — Biblioteca 2027" criada com sucesso.')).toBeVisible();
  await expect(cartao(admin, /^Priorização — Biblioteca 2027$/)).toBeVisible();
});

test("eleitor não acessa telas de admin, rota inexistente mostra 404 e Sair volta ao login", async ({ browser }) => {
  const maria = await sessao(browser, "maria");

  await maria.goto("/admin/unidades");
  await expect(maria.getByText("Acesso restrito")).toBeVisible();
  await expect(maria.getByRole("link", { name: "Unidades" })).toHaveCount(0);

  await maria.goto("/rota/que/nao/existe");
  await expect(maria.getByText("Página não encontrada")).toBeVisible();

  await maria.goto("/");
  await maria.getByRole("button", { name: "Sair" }).click();
  await expect(maria).toHaveURL(/\/login$/);
});

test("nenhum erro no console do navegador durante o roteiro (inclui violações de CSP)", () => {
  expect(errosNoConsole).toEqual([]);
});
