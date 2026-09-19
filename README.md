# Betor Search

## Objetivo do projeto

O Betor Search e uma camada de borda do ecossistema BeTor.
O objetivo principal e integrar os dados de torrents (filmes e series) com metadados externos, expondo uma API versionada para consumo por ferramentas como Prowlarr e automacoes de busca focadas no publico brasileiro.

## Indexer Cardigann

Arquivo da definicao Cardigann atual:

- [betor-search.yml](betor-search.yml)

Esse arquivo pode ser usado como base para o indexer customizado no Prowlarr.

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

## Variáveis de ambiente

| Variável | Descrição | Valor padrão |
| --- | --- | --- |
| `PORT` | Porta HTTP em que o serviço irá escutar. | `8080` |
| `BETOR_SEARCH_API_BASE_URL` | Base URL da API pública do BeTor usada para montar o endpoint de itens. | `https://api.betor.top/` |
| `BETOR_SEARCH_API_AUTHORIZATION_BASIC_VALUE` | Valor do header `Authorization: Basic ...` já codificado em Base64 e pronto para uso. Quando vazio, a aplicação não envia o header. | vazio |
| `BETOR_SEARCH_UPDATE_INTERVAL_MINUTES` | Intervalo, em minutos, entre cada sincronização do catálogo em memória com a API do BeTor. | `30` |
| `BETOR_SEARCH_DOWNLOAD_ITEMS_URL` | URL completa opcional para sobrescrever o endpoint de dump de itens. Quando não definida, o serviço combina `BETOR_SEARCH_API_BASE_URL` com `/v1/admin/download-items/`. | `https://api.betor.top/v1/admin/download-items/` |
| `BETOR_SEARCH_COMPONENT_NAME` | Nome do componente de catálogo usado em logs e observabilidade do health. | `betor-search-catalog` |

Quando `BETOR_SEARCH_API_AUTHORIZATION_BASIC_VALUE` estiver preenchido, a aplicação envia o header `Authorization: Basic <valor>` no request de sincronização do catálogo. Quando estiver vazio, o header é omitido.

Exemplo de execução com variáveis customizadas:

```bash
PORT=9090 \
BETOR_SEARCH_API_BASE_URL=https://api.betor.top/ \
BETOR_SEARCH_API_AUTHORIZATION_BASIC_VALUE=$(printf '%s' 'betor:senha' | base64) \
BETOR_SEARCH_UPDATE_INTERVAL_MINUTES=15 \
BETOR_SEARCH_DOWNLOAD_ITEMS_URL=https://api.betor.top/v1/admin/download-items/ \
BETOR_SEARCH_COMPONENT_NAME=betor-catalog \
go run ./cmd/server
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
