# Addendum - PRD Betor Search (Fase 1)

## Contexto Técnico de Referência
- Fonte funcional principal do formato: [docs/prowlarr/cardigann-yml-definition.md](docs/prowlarr/cardigann-yml-definition.md)
- Exemplos internos de referência:
  - [docs/prowlarr/indexers-examples/uindex.yml](docs/prowlarr/indexers-examples/uindex.yml)
  - [docs/prowlarr/indexers-examples/animeworld-api.yml](docs/prowlarr/indexers-examples/animeworld-api.yml)
- Contrato de dados BeTor a preservar quando aplicável:
  - [docs/betor/openapi.json](docs/betor/openapi.json)

## Notas Técnicas Capturadas para Implementação
- A entrega alvo imediata e validável é um arquivo Cardigann YML.
- A definição deve manter baixa complexidade de template; evitar condicionais em cascata.
- A API de evolução do projeto seguirá prefixo versionado `/v1`.
- O fluxo de aprovação prevê iterações de ajuste no YML antes de avançar para endpoints.

## Pendências de Documentação
- Incluir no README do projeto uma seção: "Como carregar e testar indexer customizado no Prowlarr".
- Documentar exemplos de consultas (movie-search e tv-search) e limitações conhecidas da fase 1.
