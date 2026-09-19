---
title: Betor Search - Endpoint Search com Cache de Catalogo
status: draft
created: 2026-09-19
updated: 2026-09-19
---

# PRD: Betor Search - Endpoint Search com Cache de Catálogo

## 0. Propósito do Documento
Este PRD define a próxima entrega do Betor Search: o endpoint de busca `/v1/search/` alimentado por um catálogo local em memória, atualizado periodicamente com os itens exportados pela API do BeTor por meio do endpoint `/v1/admin/download-items/`.

O objetivo é permitir busca rápida e estável sobre os itens disponíveis no BeTor, sem depender de consultas em tempo real a cada request, mantendo o serviço resiliente e previsível para uso em indexadores e integrações locais.

## 1. Visão
A busca da camada Betor Search deve funcionar como uma interface de consulta sobre um catálogo local consolidado e sincronizado periodicamente. O sistema deve baixar o dump completo de itens do BeTor, carregar esse payload na memória ao iniciar, manter atualização periódica conforme intervalo configurável, e atender os filtros do endpoint `/v1/search/` principal do projeto.

Esse passo é o fundamento para a próxima geração de uso do projeto em Prowlarr, explorando o contrato já definido em [docs/betor/openapi.json](docs/betor/openapi.json) e o arquivo Cardigann em [betor-search.yml](betor-search.yml).

## 2. Contexto e base técnica
- A API do BeTor expõe `/v1/admin/download-items/` para baixar o catálogo completo de itens em JSON.
- O payload pode vir em diferentes estruturas, incluindo um único objeto quando o exemplo de teste é reduzido, mas em produção deve incluir todos os itens do catálogo.
- O projeto já define endpoints versionados em `/v1` e um arquivo Cardigann para indexação no Prowlarr.
- A aplicação precisa de um componente de sincronização que garanta: carregamento inicial, atualização automática em background e status observável.

## 3. Usuário-alvo
- Desenvolvedor do projeto responsável por manter o catálogo e os endpoints em `/v1`.
- Operador/self-hosted que usa a API para buscas rápidas de filmes e séries.
- Indexadores e integrações que consultam dados locais em memória em vez de depender de scraping direto do BeTor.

## 4. Jornadas-chave
### UJ-1. Inicialização da aplicação
- Contexto: serviço sobe em ambiente local ou de produção.
- Entrada: aplicação iniciada com configuração padrão ou por variáveis de ambiente.
- Caminho: sistema carrega o catálogo inicial, registra a execução e expõe status via health.
- Resolução: o serviço fica disponível e pronto para respostas de busca.

### UJ-2. Atualização periódica do catálogo
- Contexto: aplicação já está rodando.
- Entrada: intervalo configurado de atualização.
- Caminho: scheduler dispara a atualização; a aplicação baixa o dump de `/v1/admin/download-items/`, substitui ou mescla em memória e registra timestamp de última execução.
- Resolução: catálogo atualiza sem necessidade de reinicialização do processo.

### UJ-3. Consulta de busca
- Contexto: catálogo já carregado.
- Entrada: request GET ao endpoint `/v1/search/` com parâmetros básicos de busca.
- Caminho: serviço consulta os dados em memória aplicando filtros e paginação.
- Resolução: resposta retorna itens relevantes em formato compatível com a API do projeto.

## 5. Requisitos Funcionais

### 5.1 Componente de catálogo em memória
#### FR-1: Carregar catálogo no bootstrap
Ao iniciar a aplicação, o sistema deve buscar o dump de itens a partir do endpoint `/v1/admin/download-items/` e carregar os resultados em memória.

Consequências testáveis:
- A aplicação consegue iniciar mesmo sem catálogo prévio.
- No primeiro boot, o componente cria estrutura interna de catálogo e registra o momento da carga inicial.
- Se a carga falhar, o status do componente sinaliza erro e o health reflete esse cenário.

#### FR-2: Atualizar o catálogo em intervalo configurável
O sistema deve reexecutar a sincronização do dump em intervalos configuráveis, com valor padrão de 30 minutos, substituindo completamente o catálogo em memória após cada atualização bem-sucedida.

Consequências testáveis:
- Existe variável de ambiente para configurar a frequência.
- O valor padrão é 30 minutos quando não informado.
- Cada execução atualiza estado interno, substitui o catálogo atual e registra o timestamp da última execução.
- Em caso de falha, o catálogo anterior permanece apenas enquanto houver dados válidos em cache; se não houver carga bem-sucedida anterior, o status do componente fica `DOWN`.

#### FR-3: Manter estado observável do componente
O componente de catálogo deve expor ao menos: status atual, última execução, timestamp da última atualização bem-sucedida e erro associado (quando existir).

Consequências testáveis:
- O health usa esse status para decidir se o serviço está UP ou DEGRADED/ERROR.
- O timestamp da última execução pode ser consultado em resposta de endpoint/health ou estrutura do componente.

### 5.2 Endpoint `/v1/search/`
#### FR-4: Consultar itens em memória
O endpoint `/v1/search/` deve consultar o catálogo em memória, respeitando parâmetros de pesquisa e paginação já definidos no OpenAPI.

Consequências testáveis:
- Requisições com `q` e filtros básicos retornam resultados do catálogo local.
- A resposta usa o formato `SearchPage[ItemSchema]` compatível com o contrato do projeto.

#### FR-5: Suporte a filtros mínimos da primeira entrega
Para a primeira entrega, os filtros devem cobrir os inputs do Cardigann definidos em [betor-search.yml](betor-search.yml), incluindo:
- `q`
- `imdb_id`
- `tmdb_id`
- `item_type`
- `season`
- `episode` (campo `ep` da busca do cardigann)
- `page`
- `size`

Consequências testáveis:
- Consultas por palavra-chave retornam registros relevantes.
- Consultas por IMDb/TMDB retornam resultados consistentes.
- Filtros por `movie` e `tv` funcionam corretamente.
- A lógica de busca do endpoint `/v1/search/` permanece compatível com a definição do indexador.

### 5.3 Componente de health
#### FR-6: Incluir status do catálogo no health
O endpoint `/health` deve incorporar o estado do componente de catálogo como parte de sua resposta.

Consequências testáveis:
- Resposta do health inclui uma estrutura para o componente de catálogo.
- O status do componente deve ser um valor do conjunto `{UP, DOWN, DEGRADED}`.
- O serviço só fica com health geral `DOWN` quando o componente de catálogo estiver em `DOWN`.
- O componente é crítico, mas não deve repassar `DOWN` para o serviço em cenários em que ainda exista cache válido em memória.

#### FR-7: Explicar a última execução
O health deve informar quando a última sincronização foi executada e qual foi o resultado.

Consequências testáveis:
- A resposta expõe timestamp ISO 8601 da última execução.
- O componente reporta `UP` quando a última atualização foi bem-sucedida.
- O componente reporta `DEGRADED` quando a última tentativa falhou, mas o sistema ainda possui dados em memória válidos.
- O componente reporta `DOWN` quando nunca houve carregamento com sucesso.
- Quando a execução falhar, a resposta mantém sinal de erro e o timestamp da última execução ainda é registrado.

## 6. Requisitos Não Funcucionais
- NFR-1: a sincronização deve ser tolerante a falha; erros de rede ou payload inválido não devem derrubar o serviço.
- NFR-2: o carregamento inicial deve ser eficiente para catalogos medianos ou grandes, usando processamento em memória e paginação se necessário.
- NFR-3: as variáveis de ambiente devem seguir convenção simples e legível para self-hosting.
- NFR-4: o health deve continuar simples para monitoramento básico, sem expor detalhes internos de implementação em excesso.

## 7. Variáveis de ambiente

### 7.1 Configuração da sincronização
- `BETOR_SEARCH_UPDATE_INTERVAL_MINUTES` — intervalo em minutos para atualização do catálogo. Padrão: `30`.
- `BETOR_SEARCH_DOWNLOAD_ITEMS_URL` — URL da API de download do catálogo, caso a base seja sobrescrita em ambiente específico. Valor default: `/v1/admin/download-items/` no endpoint configurado da aplicação.

### 7.2 Observabilidade
- `BETOR_SEARCH_COMPONENT_NAME` — nome do componente de catálogo para health e logs, opcional.

## 8. Critérios de Aceite
- CA-1: ao iniciar a aplicação, o catálogo é carregado a partir de `/v1/admin/download-items/` e fica disponível em memória.
- CA-2: a atualização periódica acontece conforme variavel de ambiente e usa valor padrão de 30 minutos.
- CA-3: o health inclui status do componente e timestamp da última execução.
- CA-4: o endpoint `/v1/search/` consulta o catálogo local e respeita filtros principais.
- CA-5: falhas de sincronização não impedem o serviço de responder status básico de saúde.

## 9. Fora de Escopo
- Redis, banco de dados ou cache distribuído.
- Reindexação avançada por full-text search.
- Sincronização em streaming de itens incrementais sem dump completo.
- Estrutura de ranking sofisticado ou scoring por qualidade de torrent.

## 10. Decisões Confirmadas
- [DECISION] O catálogo deve ser substituído integralmente a cada atualização bem-sucedida, sem mesclagem incremental por item.
- [DECISION] O endpoint `/v1/search/` deve cobrir os inputs do Cardigann em [betor-search.yml](betor-search.yml), incluindo `q`, `imdb_id`, `tmdb_id`, `item_type`, `season`, `episode` e `ep`.
- [DECISION] O status do health do componente deve usar `UP`, `DOWN` e `DEGRADED`, seguindo esta regra: `UP` quando a última sincronização foi bem-sucedida; `DEGRADED` quando a última tentativa falhou, mas há dados em memória válidos; `DOWN` quando nunca houve carregamento com sucesso.
- [DECISION] O componente de catálogo é crítico, mas a resposta geral do health só deve ser `DOWN` se o componente estiver em `DOWN`.

## 11. Assumptions
- [ASSUMPTION] A primeira entrega usa catálogo em memória e não banco externo.
- [ASSUMPTION] A sincronização usa o dump completo do endpoint `/v1/admin/download-items/` e não uma stream incremental.
- [ASSUMPTION] O componente de catálogo faz parte do health como um componente crítico de disponibilidade.
- [ASSUMPTION] O status do catálogo será refletido no health seguindo as regras definidas em `UP`, `DEGRADED` e `DOWN`.
