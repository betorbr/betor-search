---
title: Betor Search - Endpoint Público de Catalogo
status: draft
created: 2026-09-19
updated: 2026-10-21
---

# PRD: Betor Search - Endpoint Público de Catálogo

## 0. Propósito do Documento
Este PRD atualiza a entrega do Betor Search para refletir o novo comportamento do BeTor: o dump de itens do catálogo agora é público e está disponível em `https://v1.betor.top/download-items/`.

A mudança exige alinhar o comportamento padrão do serviço para esse endpoint público, remover qualquer necessidade de autenticação na sincronização do catálogo e manter a documentação operacional atualizada com exemplos de execução via Docker e binário compilado.

## 1. Visão
A camada Betor Search deve continuar funcionando como uma interface de consulta sobre um catálogo local em memória, porém usando o endpoint público do BeTor como fonte padrão de sincronização. O sistema deve baixar o dump completo de itens no início da execução e em intervalos configurados, sem depender de headers de autenticação, mantendo o serviço simples de self-hosting e compatível com indexadores e integrações locais.

## 2. Contexto e base técnica
- O BeTor expõe o dump do catálogo em `https://v1.betor.top/download-items/` sem exigir autenticação.
- O serviço deve considerar esse endpoint como padrão para bootstrapping e atualização do catálogo em memória.
- O projeto já define endpoints versionados em `/v1` e um arquivo Cardigann para indexação no Prowlarr.
- A aplicação precisa de um componente de sincronização que garanta carregamento inicial, atualização automática em background e status observável.
- A autenticação do cliente do BeTor deixou de ser parte da estratégia operacional da busca local.

## 3. Usuário-alvo
- Desenvolvedor do projeto responsável pela manutenção do catálogo e dos endpoints em `/v1`.
- Operador/self-hosted que executa o serviço localmente com mínimos requisitos de configuração.
- Indexadores e integrações que consultam dados locais em memória sem depender de credenciais internas.

## 4. Jornadas-chave
### UJ-1. Inicialização da aplicação
- Contexto: serviço sobe em ambiente local ou de produção.
- Entrada: aplicação iniciada com configuração padrão ou por variáveis de ambiente.
- Caminho: sistema carrega o catálogo inicial a partir do endpoint público e registra a execução.
- Resolução: o serviço fica disponível e pronto para respostas de busca sem exigir credenciais.

### UJ-2. Atualização periódica do catálogo
- Contexto: aplicação já está rodando.
- Entrada: intervalo configurado de atualização.
- Caminho: scheduler dispara a atualização; a aplicação baixa o dump de `https://v1.betor.top/download-items/`, substitui o catálogo em memória e registra timestamp da última execução.
- Resolução: catálogo atualiza sem reinicialização do processo nem envio de header de autenticação.

### UJ-3. Consulta de busca
- Contexto: catálogo já carregado.
- Entrada: request GET ao endpoint `/v1/search/` com parâmetros básicos de busca.
- Caminho: serviço consulta os dados em memória aplicando filtros e paginação.
- Resolução: resposta retorna itens relevantes em formato compatível com a API do projeto.

## 5. Requisitos Funcionais

### 5.1 Componente de catálogo em memória
#### FR-1: Carregar catálogo no bootstrap usando endpoint público
Ao iniciar a aplicação, o sistema deve buscar o dump de itens a partir de `https://v1.betor.top/download-items/` e carregar os resultados em memória.

Consequências testáveis:
- A aplicação consegue iniciar com a configuração padrão sem qualquer credencial.
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
- O timestamp da última execução pode ser consultado em resposta de endpoint `/health` ou estrutura do componente.

### 5.2 Remoção da autenticação na sincronização
#### FR-4: Não exigir Basic Auth para o dump
O fluxo de sincronização do catálogo não deve depender de `Authorization: Basic ...`, nem de variáveis de ambiente que codifiquem segredos para a API do BeTor.

Consequências testáveis:
- O valor padrão da URL de sincronização não inclui qualquer credencial.
- O código e a documentação não referenciam `BETOR_SEARCH_API_AUTHORIZATION_BASIC_VALUE` como requisito da operação local.
- A execução com ambiente padrão funciona sem segredo, sem header extra e sem falha por ausência de credencial.

### 5.3 Endpoint `/v1/search/`
#### FR-5: Consultar itens em memória
O endpoint `/v1/search/` deve consultar o catálogo em memória, respeitando parâmetros de pesquisa e paginação já definidos no OpenAPI.

Consequências testáveis:
- Requisições com `q` e filtros básicos retornam resultados do catálogo local.
- A resposta usa o formato compatível com o contrato do projeto.

#### FR-6: Suporte a filtros mínimos da primeira entrega
Para a primeira entrega, os filtros devem cobrir os inputs do Cardigann definidos em [betor-search.yml](betor-search.yml), incluindo:
- `q`
- `imdb_id`
- `tmdb_id`
- `item_type`
- `season`
- `episode`
- `page`
- `size`

Consequências testáveis:
- Consultas por palavra-chave retornam registros relevantes.
- Consultas por IMDb/TMDB retornam resultados consistentes.
- Filtros por `movie` e `tv` funcionam corretamente.
- A lógica de busca do endpoint `/v1/search/` permanece compatível com a definição do indexador.

### 5.4 Componente de health
#### FR-7: Incluir status do catálogo no health
O endpoint `/health` deve incorporar o estado do componente de catálogo como parte de sua resposta.

Consequências testáveis:
- Resposta do health inclui uma estrutura para o componente de catálogo.
- O status do componente deve ser um valor do conjunto `{UP, DOWN, DEGRADED}`.
- O serviço só fica com health geral `DOWN` quando o componente de catálogo estiver em `DOWN`.
- O componente é crítico, mas não deve repassar `DOWN` para o serviço em cenários em que ainda exista cache válido em memória.

#### FR-8: Explicar a última execução
O health deve informar quando a última sincronização foi executada e qual foi o resultado.

Consequências testáveis:
- A resposta expõe timestamp ISO 8601 da última execução.
- O componente reporta `UP` quando a última atualização foi bem-sucedida.
- O componente reporta `DEGRADED` quando a última tentativa falhou, mas o sistema ainda possui dados em memória válidos.
- O componente reporta `DOWN` quando nunca houve carregamento com sucesso.

## 6. Requisitos Não Funcionais
- NFR-1: a sincronização deve ser tolerante a falha; erros de rede ou payload inválido não devem derrubar o serviço.
- NFR-2: o carregamento inicial deve ser eficiente para catálogos médios ou grandes, usando processamento em memória e paginação se necessário.
- NFR-3: as variáveis de ambiente devem seguir convenção simples e legível para self-hosting, sem segredos quando não são necessários.
- NFR-4: o health deve continuar simples para monitoramento básico, sem expor detalhes internos de implementação em excesso.
- NFR-5: a documentação operacional deve incluir instruções de execução com Docker e com binário compilado, com exemplos prontos para uso.

## 7. Variáveis de ambiente

### 7.1 Configuração da sincronização
- `BETOR_SEARCH_DOWNLOAD_ITEMS_URL` — URL completa do dump público de itens do catálogo. Padrão: `https://v1.betor.top/download-items/`.
- `BETOR_SEARCH_UPDATE_INTERVAL_MINUTES` — intervalo em minutos para atualização do catálogo. Padrão: `30`.

### 7.2 Observabilidade
- `BETOR_SEARCH_COMPONENT_NAME` — nome do componente de catálogo para health e logs, opcional.

### 7.3 Remoções
- `BETOR_SEARCH_API_AUTHORIZATION_BASIC_VALUE` deixa de ser necessário e deve ser removido da documentação e das instruções de execução padrão.

## 8. Critérios de Aceite
- CA-1: ao iniciar a aplicação, o catálogo é carregado a partir de `https://v1.betor.top/download-items/` e fica disponível em memória.
- CA-2: a atualização periódica acontece conforme variável de ambiente e usa valor padrão de 30 minutos.
- CA-3: o sistema não exige header `Authorization` nem segredos para sincronizar o catálogo público.
- CA-4: o health inclui status do componente e timestamp da última execução.
- CA-5: o endpoint `/v1/search/` consulta o catálogo local e respeita filtros principais.
- CA-6: a documentação operacional contém um bloco de exemplo com execução via Docker e via binário compilado.

## 9. Fora de Escopo
- Redis, banco de dados ou cache distribuído.
- Reindexação avançada por full-text search.
- Sincronização em streaming de itens incrementais sem dump completo.
- Estrutura de ranking sofisticado ou scoring por qualidade de torrent.

## 10. Decisões Confirmadas
- [DECISION] O catálogo deve usar o endpoint público `https://v1.betor.top/download-items/` como fonte padrão.
- [DECISION] O sincronismo do catálogo não deve depender de autenticação básica nem de credenciais armazenadas em ambiente.
- [DECISION] O endpoint `/v1/search/` deve cobrir os inputs do Cardigann em [betor-search.yml](betor-search.yml), incluindo `q`, `imdb_id`, `tmdb_id`, `item_type`, `season` e `episode`.
- [DECISION] O status do health do componente deve usar `UP`, `DOWN` e `DEGRADED`, seguindo esta regra: `UP` quando a última sincronização foi bem-sucedida; `DEGRADED` quando a última tentativa falhou, mas há dados em memória válidos; `DOWN` quando nunca houve carregamento com sucesso.
- [DECISION] O componente de catálogo é crítico, mas a resposta geral do health só deve ser `DOWN` se o componente estiver em `DOWN`.

## 11. Assumptions
- [ASSUMPTION] A primeira entrega usa catálogo em memória e não banco externo.
- [ASSUMPTION] A sincronização usa o dump completo do endpoint público e não uma stream incremental.
- [ASSUMPTION] O componente de catálogo faz parte do health como um componente crítico de disponibilidade.
- [ASSUMPTION] O status do catálogo será refletido no health seguindo as regras definidas em `UP`, `DEGRADED` e `DOWN`.
