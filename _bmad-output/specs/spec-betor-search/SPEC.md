---
id: SPEC-betor-search
companions: []
sources:
  - ../planning-artifacts/prds/prd-betor-search-2026-09-18/prd.md
---

> Contrato canônico. Este SPEC define o que construir nesta fase. Não há companions necessários para o escopo atual.

# Betor Search API HTTP Go Fase 1

## Why
O projeto precisa de uma base mínima e confiável para expor a API HTTP em Go e validar disponibilidade operacional desde o início. O endpoint de saúde cria o primeiro contrato público, acelera o ciclo de desenvolvimento e reduz ambiguidade para as próximas entregas.

## Capabilities

- CAP-1
  - intent: Desenvolvedor ou operador consegue verificar se o serviço HTTP está ativo por meio do endpoint de disponibilidade.
  - success: Uma requisição GET ou HEAD para /health em serviço ativo retorna HTTP 200.

- CAP-2
  - intent: Consumidor técnico consegue obter um payload de disponibilidade consistente em JSON.
  - success: A resposta de /health contém os campos status, components e timestamp, com status igual a UP, components como array vazio e timestamp em ISO 8601.

- CAP-3
  - intent: Equipe consegue validar disponibilidade sem depender de sistemas externos nesta fase inicial.
  - success: Em cenário local nominal, /health responde abaixo de 200 ms e sem chamadas externas.

- CAP-4
  - intent: Time mantém previsibilidade de integração ao preservar o contrato inicial de /health durante a fase 1.
  - success: Mudanças de semântica ou formato na resposta de /health só ocorrem após revisão formal do artefato.

## Constraints

- Implementação restrita a API HTTP em Go, com porta padrão local 8080.
- No endpoint /health, métodos diferentes de GET e HEAD devem retornar HTTP 405.
- O health check desta fase não pode consultar dependências externas.

## Non-goals

- Não implementar readiness de banco, cache ou APIs externas nesta fase.
- Não implementar autenticação ou autorização para /health nesta fase.
- Não definir estratégia completa de observabilidade nesta fase.

## Success signal

A equipe sobe o serviço local, chama /health e recebe HTTP 200 com payload válido em todas as verificações básicas de inicialização, mantendo latência nominal abaixo de 200 ms.

