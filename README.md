# BeTor Search

O BeTor Search e uma camada de borda do ecossistema BeTor.
O objetivo principal e integrar os dados de torrents (filmes e series) com metadados externos, expondo uma API versionada para consumo por ferramentas como Prowlarr e automacoes de busca focadas no publico brasileiro.

## Indexer Cardigann

Arquivo da definicao Cardigann atual:

- [betor-search.yml](betor-search.yml)

Esse arquivo pode ser usado como base para o indexer customizado no Prowlarr.

## Executar o serviço

### Binário

```bash
wget -qO betor-search https://github.com/betorbr/betor-search/releases/latest/download/betor-search-linux-amd64
chmod +x betor-search
PORT=8080 ./betor-search
```

Com porta customizada:

```bash
PORT=9090 ./betor-search
```

### Docker

```bash
docker run --rm -p 8080:8080 ghcr.io/betorbr/betor-search:latest
```

Com porta customizada:

```bash
docker run --rm -p 9090:9090 -e PORT=9090 ghcr.io/betorbr/betor-search:latest
```

## Variáveis de ambiente

A aplicação lê as seguintes variáveis em tempo de inicialização:

| Variável | Descrição | Valor padrão / comportamento |
| --- | --- | --- |
| `PORT` | Porta HTTP em que o serviço irá escutar. | `8080` |
| `BETOR_SEARCH_DOWNLOAD_ITEMS_URL` | URL completa do dump público de itens do BeTor. Esse é o único valor de base necessário para sincronizar o catálogo. | `https://v1.betor.top/download-items/` |
| `BETOR_SEARCH_UPDATE_INTERVAL_MINUTES` | Intervalo, em minutos, entre cada sincronização do catálogo em memória. | `30` |

A sincronização do catálogo agora usa o endpoint público do BeTor.

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
  "started_at": "2026-09-19T11:59:00Z",
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

A busca usa filtro por query string. Por padrão, os resultados são ordenados por `inserted_at` em ordem decrescente, exibindo os itens mais recentes primeiro. Os parâmetros atualmente aceitos pelo handler HTTP são:

| Parâmetro | Tipo | Descrição | Valor padrão |
| --- | --- | --- | --- |
| `q` | string | Texto livre buscado no `torrent_name`, `magnet_dn` e `provider_slug`. | vazio |
| `item_type` | string | Filtra por tipo do item: `movie` ou `tv`. | vazio |
| `imdb_id` | string | Filtro exato por IMDb ID. | vazio |
| `tmdb_id` | string | Filtro exato por TMDb ID. | vazio |
| `season` | integer/lista | Filtra por temporada (aceita lista separada por vírgulas). | vazio |
| `episode` | integer/lista | Filtra por episódio (aceita lista separada por vírgulas). | vazio |
| `page` | integer | Página de resultados. Deve ser maior que zero. | `1` |
| `size` | integer | Quantidade de itens por página. Aceita até `100`. | `50` |

Exemplos:

```bash
curl "http://localhost:8080/v1/search/?q=Dune&page=1&size=10"
curl "http://localhost:8080/v1/search/?item_type=movie&imdb_id=tt0000001&page=1&size=20"
curl "http://localhost:8080/v1/search/?q=the+office&item_type=tv&tmdb_id=123&season=2&episode=10&page=1&size=10"
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


## Desenvolvimento

### Pre-requisitos

- Go 1.26+

### Compilar

Compilar binário local do serviço:

```bash
go build -o bin/server ./cmd/server
```

Executar binário compilado do serviço:

```bash
./bin/server
```


### Rodar testes

Executa todos os testes do projeto:

```bash
go test ./...
```

### Executando com Docker

Construa a imagem localmente e rode o container:

```bash
docker build -t betor-search .
docker run --rm -p 8080:8080 betor-search
```

Com porta customizada:

```bash
docker run --rm -p 9090:9090 -e PORT=9090 betor-search
```

### Executando com binário

Compilar o binário localmente:

```bash
go build -trimpath -ldflags "-s -w" -o ./bin/betor-search ./cmd/server
```

Executar o binário:

```bash
PORT=8080 ./bin/betor-search
```

ou:

```bash
PORT=9090 ./bin/betor-search
```

### Executando com `go run`

```bash
go run ./cmd/server
```

Com porta customizada:

```bash
PORT=9090 go run ./cmd/server
```
