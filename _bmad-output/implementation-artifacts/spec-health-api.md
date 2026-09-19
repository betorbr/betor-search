---
title: Betor Search API - Health Endpoint
created: '2026-09-19'
status: 'done'
route: 'oneshot'
review_loop_iteration: 0
context:
  - '/Users/douglas.paz/dev/betor-search/_bmad-output/specs/spec-betor-search/SPEC.md'
  - '/Users/douglas.paz/dev/betor-search/_bmad-output/planning-artifacts/architecture/architecture-betor-search-2026-09-18/ARCHITECTURE-SPINE.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** O repositório ainda não tem a API HTTP em Go que expõe o endpoint de disponibilidade definido na SPEC, então falta a primeira superfície executável do serviço.

**Approach:** Criar a base mínima do serviço em Go com bootstrap em `cmd/server`, handler para `/health` e resposta JSON compatível com a SPEC e o spine, sem dependências externas.

## Implementation Notes

- Usar a estrutura em camadas prevista no spine: `cmd/server`, `internal/health/application`, `internal/health/contract` e `internal/health/transport`.
- Preservar o contrato do `/health`: GET e HEAD retornam 200; demais métodos retornam 405.
- Resposta deve conter `status`, `components` vazio e `timestamp` em ISO 8601.
- Porta padrão local: `8080`.
- O handler passa a enviar `Allow: GET, HEAD` para respostas válidas e 405.
- Os testes passaram a verificar o wire format exato do JSON do health check.

</frozen-after-approval>

## Review Triage Log

- false — interface de serviço ausente: neste corte, `application.Service` concreto é suficiente porque não há consumidores alternativos nem necessidade de mocking além do teste do handler.
- defer — graceful shutdown e timeouts do servidor: o review identificou a lacuna corretamente, mas ela não bloqueia a entrega do `/health` e fica para um ajuste posterior.
