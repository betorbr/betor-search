---
title: Betor Search - Camada de Borda BeTor (Fase 1)
status: final
created: 2026-09-19
updated: 2026-09-19
---

# PRD: Betor Search - Camada de Borda BeTor (Fase 1)

## 0. Propósito do Documento
Definir a primeira entrega do projeto Betor Search como camada de borda do BeTor, com foco em produzir uma definição de indexer no formato Cardigann (Prowlarr) para filmes e séries do público brasileiro. Este documento descreve o escopo de produto da fase inicial, critérios de aceite e próximos passos para evolução de endpoints versionados em `/v1`.

## 1. Contexto
O Betor Search atua como uma camada de consumo e organização de dados do projeto open source BeTor, que já raspas fontes como BluDV, Comando e Starck Filmes, estruturando títulos e metadados de torrent.

Nesta primeira entrega, o objetivo principal é disponibilizar um arquivo YML Cardigann válido e testável no Prowlarr, usando como base:
- documentação Cardigann em [docs/prowlarr/cardigann-yml-definition.md](docs/prowlarr/cardigann-yml-definition.md)
- exemplos em [docs/prowlarr/indexers-examples/uindex.yml](docs/prowlarr/indexers-examples/uindex.yml) e [docs/prowlarr/indexers-examples/animeworld-api.yml](docs/prowlarr/indexers-examples/animeworld-api.yml)
- estrutura de campos da API BeTor em [docs/betor/openapi.json](docs/betor/openapi.json)

## 2. Calibração e Escopo
- Nível de rigor: hobby.
- Resultado primário desta fase: definição Cardigann YML.
- Resultado secundário: alinhamento do backlog para endpoints `/v1/*` que dependem da definição aprovada.

## 3. Visão do Produto (Fase 1)
Disponibilizar uma integração inicial entre Prowlarr e Betor Search, permitindo consulta de torrents de filmes e séries com foco em catálogo brasileiro, em um formato Cardigann simples, de baixa complexidade de manutenção e compatível com evolução posterior de API versionada.

## 4. Usuários-Alvo
- Operador self-hosted brasileiro que usa Prowlarr para indexação de mídia.
- Desenvolvedor mantenedor do Betor Search que precisa iterar rapidamente no indexer.

## 5. Jornadas-Chave
- UJ-1. Operador cadastra o indexer no Prowlarr usando o YML e executa busca por filme/série.
	- Entrada: YML disponível e apontando para o host da API.
	- Caminho: adiciona indexer customizado, executa search, recebe resultados com metadados essenciais.
	- Saída: títulos utilizáveis em fluxos de automação (Radarr/Sonarr/Prowlarr).

- UJ-2. Mantenedor ajusta campos e filtros do YML após validação manual.
	- Entrada: feedback de testes no Prowlarr.
	- Caminho: atualiza definição com mudanças mínimas de template/seletores.
	- Saída: nova versão aprovada da definição.

## 6. Features e Requisitos Funcionais

### 6.1 Definição Cardigann Inicial
Descrição: produzir um arquivo YML Cardigann funcional para Prowlarr, focado em filmes e séries, público brasileiro e sem complexidade desnecessária de template.

#### FR-1: Gerar arquivo YML Cardigann válido
Deve existir uma definição YML aderente ao schema Cardigann e ao fluxo de importação de indexer customizado no Prowlarr.

Consequências testáveis:
- O arquivo possui cabeçalho mínimo (`id`, `name`, `description`, `language`, `type`, `encoding`, `links`).
- O arquivo define `caps` com mapeamento de categorias para filmes e séries.
- O arquivo é parseável e não falha por estrutura/indentação YAML.

#### FR-2: Foco funcional em filmes e séries
A definição deve priorizar `search`, `movie-search` e `tv-search`, sem incluir escopos que não agreguem à fase 1.

Consequências testáveis:
- Modos não essenciais (ex.: book/music) ficam ausentes ou minimizados.
- Categorias priorizam equivalentes de Movies/TV conforme Newznab.

#### FR-3: Simplicidade de template
A definição deve manter lógica simples, usando `if/else` apenas quando necessário para compatibilidade.

Consequências testáveis:
- Ausência de cadeias complexas de condicionais quando um valor fixo ou mapeamento direto resolver.
- Presença de comentários curtos apenas quando úteis para manutenção.

#### FR-4: Alinhamento de campos com BeTor
Os campos extraídos/mapeados no Cardigann devem preservar o formato e semântica mais próxima possível dos campos expostos no OpenAPI do BeTor.

Consequências testáveis:
- `title`, `download/magnet`, `size`, `seeders`, `leechers`, `date`, `imdbid`, `tmdbid` e metadados relevantes seguem formato consistente com [docs/betor/openapi.json](docs/betor/openapi.json).
- Campos ausentes no provider recebem fallback explícito quando aplicável.

#### FR-5: Escopo de idioma e público
A definição e metadados associados devem refletir foco em conteúdo e uso no contexto brasileiro.

Consequências testáveis:
- `language` e descrição do indexer coerentes com público BR.
- Filtros e mapeamentos priorizam comportamento esperado para filmes/séries BR.

### 6.2 Trilha de Evolução para API `/v1`
Descrição: preparar continuidade para endpoints versionados que suportarão e estabilizarão a definição aprovada.

#### FR-6: Backlog inicial de endpoints versionados
O PRD deve explicitar que a próxima etapa será implementação incremental sob prefixo `/v1`.

Consequências testáveis:
- Cada novo endpoint planejado após aprovação do YML começa com `/v1/`.
- Decisões de contrato evitam breaking change prematuro na fase inicial.

## 7. Requisitos Não Funcionais
- NFR-1 Manutenibilidade: YML legível por humanos, com estrutura simples.
- NFR-2 Compatibilidade: definição importável no Prowlarr sem ajustes manuais fora da configuração esperada.
- NFR-3 Evolução: base preparada para iterar rapidamente por pull requests pequenos.

## 8. Fora de Escopo (Fase 1)
- Implementar todos os endpoints finais do Betor Search.
- Cobrir categorias além de filmes/séries.
- Otimizações avançadas de ranking/relevância de busca.
- Estratégia completa de observabilidade/telemetria.

## 9. Critérios de Aceite da Fase
- CA-1 Existe um arquivo YML Cardigann aprovado pelo solicitante.
- CA-2 O YML consegue ser carregado/testado no Prowlarr para busca de filmes/séries.
- CA-3 Campos principais retornados são compatíveis com o formato da API BeTor.
- CA-4 Foi definida trilha de continuidade para endpoints `/v1`.

## 10. Entregáveis
- E-1 Definição Cardigann YML inicial (artefato principal desta fase).
- E-2 Ajustes iterativos na definição até aprovação.
- E-3 Lista de próximos endpoints `/v1` dependentes da definição.
- E-4 Tarefa de documentação: atualizar README com instruções de uso/teste do indexer customizado no Prowlarr.

## 11. Open Questions
- O indexer inicial será tratado como `public` ou haverá necessidade de `settings` para autenticação/token?
- Quais providers do BeTor entram ativos na primeira versão do YML e quais ficam para iteração seguinte?
- Qual critério mínimo de qualidade para considerar o YML “aprovado” (ex.: 3 buscas reais com resultados úteis)?

## 12. Assumptions (Fast Path)
- [ASSUMPTION] A fase 1 usará um único arquivo de definição Cardigann e não múltiplos perfis por provider.
- [ASSUMPTION] O fluxo inicial prioriza magnet/download e metadados básicos; campos avançados podem entrar depois.
- [ASSUMPTION] O nível hobby permite validação manual inicial no Prowlarr sem suíte automatizada completa neste momento.
