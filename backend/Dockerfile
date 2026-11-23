# Etapa 1: compila a API como binário estático (sem cgo).
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/consensu-api ./cmd/api

# Etapa 2: imagem mínima que só executa o binário, com usuário sem privilégios.
FROM alpine:3.23
RUN adduser -D -H -u 10001 consensu
COPY --from=build /out/consensu-api /usr/local/bin/consensu-api
USER consensu
ENV PORT=8080
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=30s --retries=3 \
  CMD wget -q -O /dev/null "http://127.0.0.1:${PORT}/health" || exit 1
ENTRYPOINT ["consensu-api"]
