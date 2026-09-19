---
title: Betor Search - API HTTP Go (Fase 1)
status: final
created: 2026-09-18
updated: 2026-09-18
---

# PRD: Betor Search - API HTTP Go (Fase 1)

## 0. Propósito do Documento
Este PRD define o primeiro incremento da API HTTP em Go do projeto Betor Search, com foco em um endpoint de disponibilidade para estabelecer a base operacional do serviço. O documento é direcionado para planejamento de implementação inicial, validação de escopo e alinhamento de próximos passos de produto técnico. Requisitos funcionais estão organizados por feature com IDs estáveis para referência futura.

## 1. Visão
A primeira entrega do Betor Search API deve disponibilizar um serviço HTTP mínimo e confiável, capaz de responder rapidamente se o processo está ativo. Essa capacidade serve como base para monitoramento, integração com ambientes de execução e validação de vida do serviço em desenvolvimento local e futuras etapas de deploy.

Nesta fase, o objetivo não é refletir saúde de dependências externas, mas sim confirmar disponibilidade do próprio serviço. O endpoint `/health` será o contrato inicial público da API e a referência para evolução posterior de readiness e checks de componentes.

## 2. Usuário-Alvo

### 2.1 Jobs To Be Done
- Como desenvolvedor do projeto, quero validar rapidamente se o serviço subiu corretamente para acelerar ciclos de desenvolvimento.
- Como operador técnico futuro, quero um endpoint estável de disponibilidade para monitorar uptime básico do serviço.

### 2.2 Não-Usuários (v1)
- Usuários finais da plataforma de apostas (não consomem diretamente este endpoint).
- Sistemas que exijam diagnóstico profundo de dependências (fora do escopo desta fase).

### 2.3 Jornadas-Chave
- UJ-1. Desenvolvedor inicia o serviço e verifica disponibilidade.
  - Contexto: ambiente local de desenvolvimento.
  - Estado de entrada: serviço iniciado em uma porta HTTP configurada.
  - Caminho: executa o serviço, faz requisição GET para `/health`, recebe JSON de confirmação.
  - Clímax: confirmação explícita de que o serviço está ativo.
  - Resolução: desenvolvedor prossegue para desenvolvimento de novos endpoints.

## 3. Glossário
- Serviço: processo HTTP em Go responsável por expor endpoints da API.
- Endpoint de Disponibilidade: rota HTTP usada para sinalizar que o Serviço está ativo.
- Componentes: lista de subpartes monitoradas pelo endpoint; nesta fase inicia vazia.
- Timestamp: data/hora em formato ISO 8601 indicando quando a resposta foi gerada.

## 4. Features

### 4.1 Endpoint de Disponibilidade (`/health`)
Descrição: O Serviço expõe uma rota HTTP GET em `/health` que responde sucesso quando o processo está ativo. Esta feature realiza UJ-1. Nesta fase, a verificação é somente de liveness do processo, sem validação de dependências externas.

Requisitos Funcionais:

#### FR-1: Expor rota de health check
O Serviço deve disponibilizar o endpoint `/health` para requisições HTTP GET e HEAD. Realiza UJ-1.

Consequências (testáveis):
- Quando o Serviço estiver ativo e receber GET ou HEAD em `/health`, deve retornar status HTTP 200.
- Para métodos HTTP diferentes de GET e HEAD em `/health`, o Serviço deve retornar HTTP 405 (Method Not Allowed).
- Requisições para caminhos diferentes de `/health` não são cobertas por este requisito.

#### FR-2: Retornar payload JSON padronizado
O Serviço deve retornar o JSON de disponibilidade no formato definido para a fase 1. Realiza UJ-1.

Consequências (testáveis):
- A resposta deve ter `Content-Type` compatível com JSON.
- O corpo da resposta deve conter os campos `status`, `components` e `timestamp`.
- `status` deve ser `UP`.
- `components` deve ser um array vazio nesta fase.
- `timestamp` deve estar em formato ISO 8601.

#### FR-3: Garantir resposta de baixa latência em cenário nominal
O Endpoint de Disponibilidade deve responder rapidamente em execução local para suportar feedback de desenvolvimento.

Consequências (testáveis):
- Em cenário nominal local, o tempo de resposta observado deve ser inferior a 200 ms.
- O endpoint não deve depender de chamada externa para responder nesta fase.

### 4.2 Contrato Inicial da API
Descrição: Define o primeiro contrato público da API com semântica simples e estável para evolução incremental.

Requisitos Funcionais:

#### FR-4: Manter estabilidade do contrato da fase 1
O formato e semântica da resposta de `/health` devem permanecer estáveis durante a fase 1.

Consequências (testáveis):
- Mudanças no payload de `/health` exigem revisão formal deste PRD.
- Consumidores internos de desenvolvimento podem confiar no contrato atual ao longo da fase.

## 5. Não-Objetivos (Explícitos)
- Não implementar readiness check de banco, cache ou APIs externas nesta fase.
- Não implementar autenticação/autorização para o endpoint `/health` nesta fase.
- Não definir neste PRD estratégia completa de observabilidade (tracing/metrics avançadas).

## 6. Escopo MVP

### 6.1 Em Escopo
- Serviço HTTP em Go iniciado e acessível.
- Endpoint GET `/health` funcional.
- Retorno HTTP 200 com JSON contendo `status`, `components` e `timestamp`.
- Porta padrão local definida como `8080`.

### 6.2 Fora de Escopo para MVP
- Endpoints de negócio além de `/health`.
- Validação de saúde de dependências externas.
- Estratégia de versionamento de API pública além da fase inicial.

## 7. Métricas de Sucesso
Primária:
- SM-1: Endpoint `/health` responde com HTTP 200 e payload esperado em 100% dos testes manuais básicos de inicialização. Valida FR-1 e FR-2.

Secundária:
- SM-2: Tempo de resposta do `/health` em cenário local nominal abaixo de 200 ms na maioria das verificações de desenvolvimento. Valida FR-3.

Contra-métrica:
- SM-C1: Complexidade de implementação do health check não deve bloquear a entrega do endpoint base (evitar overengineering prematuro). Contrabalança SM-2.

## 8. Questões em Aberto
- Nenhuma questão em aberto nesta versão.

## 9. Índice de Assunções
- Nenhuma assunção pendente nesta versão.
