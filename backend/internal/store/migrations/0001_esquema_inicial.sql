CREATE TABLE usuarios (
  id          BIGSERIAL PRIMARY KEY,
  nome        TEXT NOT NULL,
  login       TEXT NOT NULL UNIQUE,
  senha_hash  TEXT NOT NULL,
  papel       TEXT NOT NULL CHECK (papel IN ('admin','eleitor'))
);

-- token_hash é o sha256 (hex) do token; o token em si não é guardado.
CREATE TABLE sessoes (
  token_hash  TEXT PRIMARY KEY,
  usuario_id  BIGINT NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
  expira_em   TIMESTAMPTZ NOT NULL
);

CREATE TABLE unidades (
  id             BIGSERIAL PRIMARY KEY,
  identificador  TEXT NOT NULL UNIQUE,
  fracao_ideal   DOUBLE PRECISION NOT NULL CHECK (fracao_ideal > 0),
  adimplente     BOOLEAN NOT NULL DEFAULT TRUE,
  responsavel_id BIGINT UNIQUE REFERENCES usuarios(id)
);

CREATE TABLE votacoes (
  id                  BIGSERIAL PRIMARY KEY,
  tipo                TEXT NOT NULL CHECK (tipo IN ('pauta','eleicao','consulta')),
  titulo              TEXT NOT NULL,
  descricao           TEXT NOT NULL DEFAULT '',
  inicio              TIMESTAMPTZ NOT NULL,
  fim                 TIMESTAMPTZ NOT NULL,
  turno               INT NOT NULL DEFAULT 1 CHECK (turno IN (1,2)),
  votacao_anterior_id BIGINT UNIQUE REFERENCES votacoes(id),
  orcamento_centavos  BIGINT CHECK (orcamento_centavos >= 0),
  CHECK (fim > inicio)
);

CREATE TABLE opcoes (
  id             BIGSERIAL PRIMARY KEY,
  votacao_id     BIGINT NOT NULL REFERENCES votacoes(id) ON DELETE CASCADE,
  nome           TEXT NOT NULL,
  descricao      TEXT NOT NULL DEFAULT '',
  custo_centavos BIGINT CHECK (custo_centavos >= 0),
  posicao        INT NOT NULL,
  UNIQUE (votacao_id, nome),
  UNIQUE (votacao_id, posicao)
);

-- Eleitores habilitados (eleição e consulta).
CREATE TABLE aptos (
  votacao_id BIGINT NOT NULL REFERENCES votacoes(id) ON DELETE CASCADE,
  usuario_id BIGINT NOT NULL REFERENCES usuarios(id),
  PRIMARY KEY (votacao_id, usuario_id)
);

-- Quem já votou; sem escolha e sem horário.
CREATE TABLE participacoes (
  votacao_id BIGINT NOT NULL REFERENCES votacoes(id) ON DELETE CASCADE,
  usuario_id BIGINT NOT NULL REFERENCES usuarios(id),
  PRIMARY KEY (votacao_id, usuario_id)
);

-- Cédula anônima: sem eleitor, sem horário.
CREATE TABLE cedulas (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  votacao_id BIGINT NOT NULL REFERENCES votacoes(id) ON DELETE CASCADE,
  tipo       TEXT NOT NULL CHECK (tipo IN ('chapa','branco','nulo','ranking')),
  opcao_id   BIGINT REFERENCES opcoes(id),
  CHECK ((tipo = 'chapa') = (opcao_id IS NOT NULL))
);

-- Lista ordenada da cédula Borda.
CREATE TABLE cedula_preferencias (
  cedula_id UUID NOT NULL REFERENCES cedulas(id) ON DELETE CASCADE,
  posicao   INT NOT NULL CHECK (posicao >= 1),
  opcao_id  BIGINT NOT NULL REFERENCES opcoes(id),
  PRIMARY KEY (cedula_id, posicao),
  UNIQUE (cedula_id, opcao_id)
);

-- Voto nominal ponderado da unidade.
CREATE TABLE votos_pauta (
  votacao_id BIGINT NOT NULL REFERENCES votacoes(id) ON DELETE CASCADE,
  unidade_id BIGINT NOT NULL REFERENCES unidades(id),
  escolha    TEXT NOT NULL CHECK (escolha IN ('favor','contra','abstencao')),
  PRIMARY KEY (votacao_id, unidade_id)
);
