# Frontend — Sistema de Votação

Interface web do Sistema de Votação, desenvolvida com React e TypeScript.

## Tecnologias

- React
- TypeScript
- Vite

## Requisitos

- Node.js
- npm

## Instalação

Na pasta `frontend`, instale as dependências:

```bash
npm install
```

## Executar localmente

Para iniciar o servidor de desenvolvimento:

```bash
npm run dev
```

Após iniciar, o endereço da aplicação será exibido no terminal.

## Build

Para gerar a versão de produção:

```bash
npm run build
```

Para visualizar localmente a versão gerada:

```bash
npm run preview
```

## Verificações automatizadas

Execute em `frontend/`: `npm ci`, `npm run build`, `npm run lint`, `npm test` e `npm run test:coverage`. Os testes da página inicial validam a identificação do produto e a apresentação dos três contextos de decisão. O relatório cobre os componentes de página em `src/pages/` e fica em `coverage/`; não representa cobertura de funcionalidades de votação ainda não implementadas.
