# Betor Search

## Pre-requisitos

- Go 1.26+

## Executar o serviço

Por padrão o serviço sobe na porta `8080`.

```bash
go run ./cmd/server
```

Com porta customizada:

```bash
PORT=9090 go run ./cmd/server
```

## Validar endpoint de health

Requisição GET:

```bash
curl -i http://localhost:8080/health
```

Requisição HEAD:

```bash
curl -I http://localhost:8080/health
```

Resposta esperada para GET (status 200):

```json
{"status":"UP","components":[],"timestamp":"<ISO-8601>"}
```

## Rodar testes

Executa todos os testes do projeto:

```bash
go test ./...
```

## Compilar

Compilar binário local do serviço:

```bash
go build -o bin/server ./cmd/server
```

Executar binário compilado do serviço:

```bash
./bin/server
```
