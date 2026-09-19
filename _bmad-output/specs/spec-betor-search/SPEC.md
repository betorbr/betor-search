---
id: SPEC-betor-search-sync-search
companions: []
sources:
  - ../planning-artifacts/prds/prd-betor-search-2026-09-19-search/prd.md
  - ../../docs/betor/openapi.json
  - ../../betor-search.yml
---

> Contrato canônico. Este SPEC define o que construir nesta fase de sincronização do catálogo e do endpoint de busca. Ele deve ser lido em conjunto com o PRD aprovado e com o contrato HTTP do BeTor.

# Betor Search API HTTP Go Fase 2

## Why
O projeto precisa ir além do health check inicial e entregar o principal valor do produto: manter um catálogo em memória do BeTor e expor uma busca local por `/v1/search/`. A API do BeTor já fornece o dump de itens em `/v1/admin/download-items/`; a aplicação deve consumi-lo no boot e atualizar esse catálogo periodicamente sem depender de consultas em tempo real a cada request.

## Capabilities

- CAP-1
  - intent: Aplicação consegue inicializar com catálogo válido em memória.
  - success: Ao subir a aplicação, ela chama `/v1/admin/download-items/`, processa a resposta e mantém o catálogo de itens disponível para consultas.

- CAP-2
  - intent: Serviço mantém catálogo atualizado com periodicidade configurável.
  - success: A aplicação executa a sincronização em intervalo configurado pelo ambiente, com padrão de 30 minutos, e substitui o catálogo em memória após sucesso de cada atualização.

- CAP-3
  - intent: Integrador consegue buscar itens por filtros compatíveis com a API do BeTor e com o indexador Cardigann.
  - success: O endpoint `/v1/search/` aceita os parâmetros principais com base em `q`, `imdb_id`, `tmdb_id`, `item_type`, `season`, `episode`, `page` e `size` e retorna somente itens do catálogo local.

- CAP-4
  - intent: Operador consegue verificar saúde do componente crítico de catálogo.
  - success: O endpoint `/health` expõe o componente de sincronização do catálogo e informa seu status (`UP`, `DEGRADED`, `DOWN`) e o timestamp da última execução.

- CAP-5
  - intent: Serviço continua operacional mesmo em falha de sincronização.
  - success: Se a última sincronização falhar, o serviço mantém funcionamento e reporta `DEGRADED` quando ainda houver cache válido em memória; o status geral do health somente fica `DOWN` se o componente estiver em `DOWN`.

## Constraints

- Implementação restrita a Go, com a estrutura atual do projeto em `internal/health` e `cmd/server`.
- O catálogo deve ser mantido em memória e não depender de banco externo nesta fase.
- Atualização do catálogo deve seguir a estratégia de substituição completa após carga bem-sucedida.
- O estado do componente crítico deve respeitar a regra: `UP` = última sincronização ok; `DEGRADED` = última tentativa falhou mas há dados em memória; `DOWN` = nunca houve carregamento com sucesso.
- O status geral do health só deve resultar em `DOWN` quando o componente do catálogo estiver em `DOWN`.
- O contrato de `/health` mantém suporte aos métodos `GET` e `HEAD`, com `405` para outros métodos.
- A frequência da sincronização deve ser parametrizada por variável de ambiente e o padrão oficial é 30 minutos.

## Non-goals

- Não adicionar banco de dados, Redis ou cache distribuído nesta fase.
- Não implementar busca full-text avançada ou ranking sofisticado.
- Não suportar stream incremental de updates; a sincronização é por dump completo.
- Não incluir autenticação ou autorização específica para `/v1/search/` nesta fase.
- Não promover lógica de scraping direto em cada request; a busca local é a fonte da verdade.

## Success signal

A equipe sobe o serviço local, a inicialização carrega o catálogo do endpoint `/v1/admin/download-items/`, a atualização periódica reexecutes em intervalos configurados, o endpoint `/v1/search/` responde com dados do catálogo em memória e o `/health` informa status, timestamp e degradação/queda do componente crítico conforme a última sincronização.

