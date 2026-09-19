---
name: Betor Search API HTTP Go Fase 1
type: architecture-spine
purpose: build-substrate
altitude: feature
paradigm: Layered HTTP Service
scope: Endpoint de disponibilidade /health e contrato inicial da API
status: final
created: 2026-09-18
updated: 2026-09-18
binds:
  - CAP-1
  - CAP-2
  - CAP-3
  - CAP-4
sources:
  - ../../../specs/spec-betor-search/SPEC.md
companions: []
---

# Architecture Spine - Betor Search API HTTP Go Fase 1

## Design Paradigm
Arquitetura em camadas enxuta para serviço HTTP:
- Transport layer: recebe requisições HTTP e escreve respostas.
- Application layer: aplica regras do caso de uso health check.
- Domain contract: define formato de saída e invariantes de comportamento.

## Invariants & Rules

### AD-1 - Superfície HTTP estável para health check

- Binds: CAP-1, CAP-4
- Prevents: Implementações divergentes que aceitem métodos inconsistentes em /health.
- Rule: O endpoint /health deve aceitar apenas GET e HEAD com HTTP 200 em serviço ativo; qualquer outro método em /health retorna HTTP 405.

### AD-2 - Envelope de resposta canônico

- Binds: CAP-2, CAP-4
- Prevents: Variações de payload e semântica que quebrem consumidores técnicos.
- Rule: A resposta de /health deve ser JSON com campos status, components e timestamp, onde status = UP, components = [] e timestamp em ISO 8601.

### AD-3 - Isolamento operacional do health check

- Binds: CAP-3
- Prevents: Falso negativo por indisponibilidade de dependência externa e aumento de latência do check.
- Rule: O caso de uso de /health não consulta banco, cache ou APIs externas; o resultado depende apenas da disponibilidade do processo HTTP.

### AD-4 - Ownership único de bootstrap HTTP

- Binds: CAP-1, CAP-3
- Prevents: Inicialização duplicada do servidor e deriva de configuração de porta.
- Rule: O bootstrap do servidor HTTP pertence somente ao módulo de entrada em cmd/server, com porta padrão local 8080.

```mermaid
flowchart LR
  client[HTTP Client] --> handler[/health handler]
  handler --> app[Health Application Service]
  app --> contract[Health Response Contract]
  contract --> handler
```

## Consistency Conventions

| Concern | Convention |
| --- | --- |
| Naming (entities, files, interfaces, events) | Handler com sufixo Handler, serviço de caso de uso com sufixo Service, contrato de saída em arquivos response.go. |
| Data & formats (ids, dates, error shapes, envelopes) | timestamp em UTC ISO 8601 RFC3339; envelope de sucesso fixo para /health; erro de método inválido com status 405. |
| State & cross-cutting (mutation, errors, logging, config, auth) | Endpoint /health é stateless; config de porta via env com fallback 8080; auth fora de escopo nesta fase. |

## Stack

| Name | Version |
| --- | --- |
| Go | 1.27.1 |
| net/http (stdlib) | Go 1.27.1 |

## Structural Seed

```text
betor-search/
  cmd/
    server/        # bootstrap HTTP e wiring principal
  internal/
    health/
      application/ # regra de caso de uso health
      contract/    # estrutura de resposta canônica
      transport/   # handler e adaptação HTTP
```

## Capability -> Architecture Map

| Capability / Area | Lives in | Governed by |
| --- | --- | --- |
| CAP-1 | internal/health/transport + cmd/server | AD-1, AD-4, Layered HTTP Service |
| CAP-2 | internal/health/contract + internal/health/transport | AD-2, convenção de formatos |
| CAP-3 | internal/health/application | AD-3, AD-4 |
| CAP-4 | internal/health/contract + internal/health/transport | AD-1, AD-2 |

## Deferred

- Estratégia de readiness com dependências externas (banco, cache, APIs) fica para fase posterior quando esses componentes existirem.
- Estratégia completa de observabilidade (metrics, tracing e log estruturado) fica para próximo incremento técnico.
- Definição de autenticação/autorização global da API fica para fase com endpoints de negócio.
