---
title: Betor Search - CI/CD com GitHub Actions
status: final
created: 2026-09-19
updated: 2026-09-19
---

# PRD: Betor Search - CI/CD com GitHub Actions

## 0. Propósito do Documento
Este PRD define a estratégia de integração contínua e entrega contínua do projeto Betor Search usando GitHub Actions. O objetivo é automatizar validações de qualidade em pull requests para a branch `main` e publicar artefatos de release quando uma tag iniciada com `v` for criada, incluindo binário Linux, imagem Docker no registro do próprio GitHub e release pública com artefato anexado.

## 1. Contexto
O projeto Betor Search é uma aplicação Go com execução via `go run ./cmd/server` e compilação em binário local. A partir deste PRD, o repositório passará a ter um fluxo automatizado de garantia de qualidade e publicação, reduzindo risco de regressões, padronizando a construção de artefatos e permitindo entrega consistente para ambientes de execução e consumo por terceiros.

O fluxo de entrega deve refletir duas regras operacionais claras:
- ao abrir um PR apontando para `main`, deve disparar a validação de CI;
- ao criar uma tag iniciando com `v`, como `v0.1.0`, deve disparar o processo de CD.

## 2. Calibração e Escopo
- Nível de rigor: internal / launch-ready.
- Resultado primário: automatizar validação e publicação sem introduzir fricção no processo de desenvolvimento.
- Resultado secundário: estabelecer padrão reutilizável para futuras versões do projeto.

## 3. Visão do Produto
Disponibilizar um pipeline GitHub Actions seguro, transparente e de baixa manutenção para garantir que alterações em branches de desenvolvimento sejam validadas antes de merge e que releases versionadas façam build do binário, empacotamento Docker e publicação de artefatos em um fluxo automatizado e rastreável.

## 4. Usuários-Alvo
- Desenvolvedor do projeto que abre PRs e precisa feedback automático de qualidade.
- Maintainer / responsável pela release que publica versões do projeto.
- Operador ou consumidor que utiliza o binário compilado ou a imagem Docker publicada.

## 5. Jornadas-Chave

### UJ-1. Desenvolvedor valida uma mudança antes do merge
- Entrada: pull request aberto contra `main` com código em Go alterado.
- Caminho: GitHub Actions dispara workflow de CI; executa testes, verificação de formatação, build e checagens de qualidade.
- Saída: status claro do PR com aprovação ou falha, permitindo merge somente quando a base estiver saudável.

### UJ-2. Maintainer publica uma release versionada
- Entrada: criação de tag `vX.Y.Z` no repositório.
- Caminho: workflow de CD dispara, gera binário Linux, monta imagem Docker, publica no GitHub Container Registry e cria release com o binário como artefato.
- Saída: release versionada pronta para consumo, com rastreio e artefatos anexados.

### UJ-3. Operador utiliza o artefato publicado
- Entrada: release ou pacote gerado via pipeline.
- Caminho: deve consumir a imagem publicada no GHCR ou o binário linux anexado à release.
- Saída: execução pronta em ambiente de produção ou homologação sem etapas manuais de build.

## 6. Features e Requisitos Funcionais

### 6.1 Pipeline de CI para PRs para `main`
Descrição: automatizar validadores mínimos para garantir que a aplicação em Go ainda está correta antes do merge em `main`.

#### FR-1: Disparo automático em pull request para `main`
A criação de um PR com destino à branch `main` deve acionar o workflow de CI automaticamente.

Consequências testáveis:
- O evento `pull_request` com base em `main` dispara o workflow.
- O status do PR mostra execução do pipeline.

#### FR-2: Execução de validações de Go
O workflow de CI deve executar etapas de qualidade e consistência do código Go para toda a branch alvo do PR.

Consequências testáveis:
- `go test ./...` é executado.
- `go vet ./...` é executado quando aplicável.
- `gofmt -w`/`gofmt -d` ou checagem equivalente de formatação é executada.
- `go build ./...` ou build do alvo principal do serviço confirma compilação do projeto.

#### FR-3: Falha explícita em regressões
Quando qualquer etapa de qualidade falha, o pipeline deve encerrar com erro e sinalizar claramente o ponto de falha.

Consequências testáveis:
- CI retorna status falho quando testes ou build quebram.
- Logs exibem o comando e a falha específica.

### 6.2 Pipeline de CD para tags versionadas
Descrição: disparar a entrega após a criação de uma tag com prefixo `v`.

#### FR-4: Disparo automático em tag no formato `v*`
A criação de qualquer tag que comece com `v`, como `v0.1.0`, deve disparar o workflow de CD.

Consequências testáveis:
- O evento `push` com ref `refs/tags/v*` dispara o workflow.
- Tags sem prefixo `v` não acionam o deploy.

#### FR-5: Geração de binário para Linux
O workflow de CD deve compilar o binário da aplicação para Linux em uma arquitetura compatível com o ambiente de execução do projeto.

Consequências testáveis:
- O binário é gerado no passo de build.
- O artefato final é nomeado de forma determinística, por exemplo `betor-search-linux-amd64`.
- O binário representa a aplicação principal do serviço.

#### FR-6: Geração de imagem Docker e publicação no GHCR
O workflow deve compilar uma imagem Docker do projeto e publicar no GitHub Container Registry (GHCR).

Consequências testáveis:
- `docker/login-action` utiliza o registry do GitHub.
- A imagem é tagueada com a versão da release e com `latest` quando apropriado.
- O repositório tem permissão de escrita em packages.

#### FR-7: Criação de release no GitHub com artefato
A release gerada pelo workflow deve incluir o binário compilado como artefato de distribuição.

Consequências testáveis:
- Existe release vinculada à tag criada.
- O binário compactado ou não compactado é anexado à release.
- O pacote e a imagem publicada mantêm a mesma referência de versão.

### 6.3 Requisitos de segurança e operabilidade
Descrição: o pipeline deve seguir princípios de segurança e rastreabilidade compatíveis com GitHub Actions.

#### FR-8: Permissões mínimas e explícitas
O workflow deve declarar permissões necessárias e não mais do que isso.

Consequências testáveis:
- `contents: write` para criar release.
- `packages: write` para publicar imagem no GHCR.
- permissões explícitas em cada workflow ou job.

#### FR-9: Rastreabilidade da versão
Cada artefato publicado deve refletir a tag/version ou o SHA usado no trigger.

Consequências testáveis:
- Tags no formato `vX.Y.Z` são usadas em release e imagem.
- O usuário consegue identificar exatamente qual versão do código gerou o artefato.

#### FR-10: Reuso e manutenção simples
A solução deve ser simples de evoluir e ser mantida por pessoas do projeto sem necessidade de grande especialização em GitHub Actions.

Consequências testáveis:
- Workflows são separados por responsabilidade (`ci.yml` e `cd.yml` ou equivalente).
- O processo usa etapas claras e bem nomeadas.

## 7. Requisitos Não Funcionais
- NFR-1 Segurança: uso de `GITHUB_TOKEN` e permissões mínimas; evitar segredos manuscritos sempre que possível.
- NFR-2 Confiabilidade: CI deve bloquear merges com regressão e o CD deve falhar de forma observável quando qualquer etapa de build/publicação falhar.
- NFR-3 Reprodutibilidade: build do binário e imagem devem ser rastreáveis à tag e ao commit correto.
- NFR-4 Simplicidade de manutenção: workflow bem dividido, sem lógica excessivamente complexa ou duplicada.
- NFR-5 Observabilidade: logs e status de job devem permitir diagnóstico rápido de falhas.
- NFR-6 Compatibilidade: solução deve funcionar com um projeto Go típico e com a arquitetura atual do serviço Betor Search.

## 8. Fora de Escopo
- Publicar em registries externos além do GHCR.
- Criar deploy automático para servidores, Kubernetes, Docker Compose ou qualquer plataforma de runtime fora do repositório.
- Configurar multiarch além da arquitetura alvo inicialmente definida.
- Gerenciar releases por ambiente diferenciado (dev/staging/prod) na mesma etapa inicial.

## 9. Critérios de Aceite
- CA-1 Um PR aberto para `main` dispara automaticamente o workflow de CI.
- CA-2 O workflow de CI executa testes, checagem de formato e build do projeto Go.
- CA-3 O workflow falha corretamente quando qualquer validação do projeto é quebrada.
- CA-4 Uma tag iniciando com `v` dispara automaticamente o workflow de CD.
- CA-5 O workflow de CD gera um binário Linux e o publica como artefato de release.
- CA-6 O workflow de CD constrói uma imagem Docker e a publica no GHCR.
- CA-7 O projeto cria uma release no GitHub associada à tag e inclui o binário como artefato.
- CA-8 O processo é documentado de forma suficiente para manutenção por desenvolvedores do projeto.

## 10. Entregáveis
- E-1 Workflow de CI de PR para `main`.
- E-2 Workflow de CD para tags `v*`.
- E-3 Binário Linux da aplicação compilado no processo de release.
- E-4 Imagem Docker publicada no GHCR.
- E-5 Release pública no GitHub com o artefato binário anexo.
- E-6 Documentação mínima no README com o comportamento do pipeline e critérios de release.

## 11. Open Questions
- O projeto precisa suportar apenas arquitetura `linux/amd64` na primeira versão ou também outras arquiteturas em uma segunda etapa?
- A release pública deve incluir o binário compactado em `.tar.gz` ou o artefato bruto gerado pela compilação?
- A imagem Docker precisará de um `Dockerfile` específico para execução do serviço ou será reutilizado um padrão já existente?
- O projeto exige branch de release específica além de tags `v*` para fluxo de publicação?

## 12. Assumptions (Fast Path)
- [ASSUMPTION] O primeiro ciclo de CI/CD usará GitHub Actions nativos do repositório, sem dependência de terceiros ou plataformas externas.
- [ASSUMPTION] O fluxo inicial cobre o caso principal: validação em PR e publicação em tag `v*`.
- [ASSUMPTION] O processo de release prioriza simplicidade, rastreabilidade e confiabilidade em vez de suporte a múltiplos ambientes ou arquiteturas na fase inicial.
- [ASSUMPTION] O binário da aplicação será do tipo Linux para execução do serviço em ambientes containerizados ou servidores compatíveis.
- [ASSUMPTION] O repositório aceita permissões de package e release no GitHub via token padrão do próprio GitHub Actions.

## 13. Recomendação de Implementação
A implementação inicial deve seguir uma divisão clara:

1. `ci.yml` executa em `pull_request` para `main`.
2. `cd.yml` executa em `push` para tags `v*`.
3. O CI valida o projeto Go com testes e build mínimo.
4. O CD compila o artefato Linux, publica a imagem Docker no GHCR e cria a release com artefato anexado.
5. O README do projeto documenta a convenção de versionamento e o comportamento esperado em cada workflow.

Esse formato garante fluxo simples, previsível e consistente com a maturidade atual do projeto, ao mesmo tempo em que deixa espaço para evoluções futuras como multi-arch, deploy automático e ambientes adicionais.
