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

A aplicação lê as seguintes variáveis em tempo de inicialização:

| Variável | Descrição | Valor padrão / comportamento |
| --- | --- | --- |
| `PORT` | Porta HTTP em que o serviço irá escutar. | `8080` |
| `BETOR_SEARCH_API_BASE_URL` | Base da API pública do BeTor usada para montar o endpoint de download de itens. | `https://api.betor.top/` |
| `BETOR_SEARCH_API_AUTHORIZATION_BASIC_VALUE` | Valor já codificado em Base64 para o header `Authorization: Basic ...`. Quando vazio, a aplicação não envia o header. | vazio |
| `BETOR_SEARCH_UPDATE_INTERVAL_MINUTES` | Intervalo, em minutos, entre cada sincronização do catálogo em memória. | `30` |
| `BETOR_SEARCH_DOWNLOAD_ITEMS_URL` | URL completa para sobrescrever o endpoint de dump de itens. Quando vazia, o serviço usa `BETOR_SEARCH_API_BASE_URL + /v1/admin/download-items/`. | opcional |

Observação importante: a variável `BETOR_SEARCH_COMPONENT_NAME` não é lida pelo runtime atual. O nome do componente no health endpoint está hardcoded como `betor-search-catalog` em [cmd/server/main.go](cmd/server/main.go) e [internal/health/application/service.go](internal/health/application/service.go).

Quando `BETOR_SEARCH_API_AUTHORIZATION_BASIC_VALUE` estiver preenchido, a aplicação envia o header `Authorization: Basic <valor>` na sincronização do catálogo. Quando vazio, o header é omitido.

Exemplo de execução com variáveis customizadas:

```bash
PORT=9090 \
BETOR_SEARCH_API_BASE_URL=https://api.betor.top/ \
BETOR_SEARCH_API_AUTHORIZATION_BASIC_VALUE=$(printf '%s' 'betor:senha' | base64) \
BETOR_SEARCH_UPDATE_INTERVAL_MINUTES=15 \
BETOR_SEARCH_DOWNLOAD_ITEMS_URL=https://api.betor.top/v1/admin/download-items/ \
go run ./cmd/server
```

## Endpoints HTTP

### Health

- `GET /health`
- `HEAD /health`

Resposta esperada para `GET` (status `200`):

```json
{
  "status": "UP",
  "components": [
    {
      "name": "betor-search-catalog",
      "status": "UP",
      "updated_at": "2026-09-19T12:00:00Z",
      "last_run_at": "2026-09-19T12:00:00Z",
      "message": "catalog sync successful"
    }
  ],
  "timestamp": "2026-09-19T12:00:00Z"
}
```

Exemplos:

```bash
curl -i http://localhost:8080/health
curl -I http://localhost:8080/health
```

### Busca

- `GET /v1/search/`

A busca usa filtro por query string. Os parâmetros atualmente aceitos pelo handler HTTP são:

| Parâmetro | Tipo | Descrição | Valor padrão |
| --- | --- | --- | --- |
| `q` | string | Texto livre buscado no `torrent_name`, `magnet_dn` e `provider_slug`. | vazio |
| `item_type` | string | Filtra por tipo do item: `movie` ou `tv`. | vazio |
| `imdb_id` | string | Filtro exato por IMDb ID. | vazio |
| `tmdb_id` | string | Filtro exato por TMDb ID. | vazio |
| `page` | integer | Página de resultados. Deve ser maior que zero. | `1` |
| `size` | integer | Quantidade de itens por página. Aceita até `100`. | `50` |

Exemplos:

```bash
curl "http://localhost:8080/v1/search/?q=Dune&page=1&size=10"
curl "http://localhost:8080/v1/search/?item_type=movie&imdb_id=tt0000001&page=1&size=20"
curl "http://localhost:8080/v1/search/?q=the+office&item_type=tv&tmdb_id=123&page=1&size=10"
```

Resposta esperada:

```json
{
  "items": [
    {
      "id": "...",
      "item_type": "movie",
      "torrent_name": "Dune Part Two",
      "provider_slug": "...",
      "provider_url": "https://...",
      "magnet_uri": "magnet:?xt=urn:btih:...",
      "magnet_dn": "...",
      "magnet_xt": "urn:btih:...",
      "languages": ["pt-BR"],
      "imdb_id": "tt...",
      "tmdb_id": "...",
      "inserted_at": "2026-09-19T00:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "size": 10
}
```

Observação: o filtro interno `season` e `episode` existe na lógica de busca, mas o parser HTTP atual não os lê da query string. Portanto, os filtros realmente expostos pela rota `/v1/search/` no código atual são `q`, `item_type`, `imdb_id`, `tmdb_id`, `page` e `size`.

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
