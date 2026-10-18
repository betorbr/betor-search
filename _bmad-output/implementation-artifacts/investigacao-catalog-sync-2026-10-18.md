# Investigação inicial: catálogo zerado / desatualizado

## Contexto

A instância em `search.betor.top` apresenta comportamento inconsistentes: em alguns momentos o catálogo aparece zerado e, em outros, continua com dados antigos mesmo depois das sincronizações programadas.

O objetivo deste artefato é registrar as hipóteses e evidências que já podem ser observadas no código, para orientar a correção em seguida.

## Evidências do código

### 1) A sincronização atual destrói o catálogo quando a resposta upstream vem vazia

No método `Sync()` em `internal/search/application/service.go`:

- o serviço marca `status = "DEGRADED"`;
- chama `fetchItemsFromURL()`;
- se `len(items) == 0`, ele retorna erro `empty catalog returned by BeTor`;
- e ainda assim, no caminho de sucesso, faz `s.items = items`.

Ou seja, quando o provedor responde com payload vazio, a app entra em `DOWN` e o catálogo em memória deixa de ser útil. Não existe fallback para o último catálogo válido.

Ponto crítico:

```go
if len(items) == 0 {
    err := errors.New("empty catalog returned by BeTor")
    s.mu.Lock()
    s.status = "DOWN"
    s.lastError = err.Error()
    s.mu.Unlock()
    return err
}
```

A consequência prática é que um upstream momentaneamente vazio pode causar "items zerados" na interface, mesmo quando o catálogo anterior ainda era válido.

### 2) O serviço não preserva um snapshot bom para usar em caso de falha parcial

O tipo `Service` mantém `items []Item` em memória, mas não guarda:

- último catálogo válido;
- histórico de último payload saudável;
- regra de segurança para reusar dados antigos quando o upstream falha ou entrega resposta vazia.

Isso torna a instância vulnerável a:

- resposta vazia do BeTor;
- falha de rede temporal;
- payload incompleto ou parcialmente atualizado;
- atualização de catálogo sem validação de freshness.

### 3) O intervalo de sincronização default pode deixar os dados antigos por tempo demais

O serviço lê `BETOR_SEARCH_UPDATE_INTERVAL_MINUTES` em `ParseUpdateIntervalFromEnv()` e usa `30 * time.Minute` como default quando a variável não é definida.

O background loop em `StartBackgroundSync()` roda por ticker com esse intervalo:

```go
func (s *Service) StartBackgroundSync(ctx context.Context) {
    ticker := time.NewTicker(s.updateInterval)
    ...
    case <-ticker.C:
        if err := s.Sync(); err != nil {
            log.Printf("scheduled catalog sync failed: %v", err)
        }
    }
}
```

Isso significa que, se o upstream ficar inconsistente por vários intervalos, o catálogo pode continuar desatualizado por bastante tempo antes de ser corrigido.

### 4) A recuperação de dados depende muito do formato vindo do upstream

A função `decodeItems()` tenta extrair itens de várias chaves (`items`, `data`, `results`, `records`) e aceita vários formatos, mas ainda assim trata payloads vazios como erro fatal.

Isso é um bom sinal de que o pipeline de ingestão está sensível a variações de payload da API externa, sem nenhuma estratégia de tolerância a falha.

## Hipóteses prováveis

1. Upstream responde com payload vazio ou incompleto em alguns ciclos.
   Isso explicaria os "items zerados" porque a app trata isso como catalog failure e não mantém o último bom estado.

2. O intervalo de atualização e/ou a ausência de fallback deixam o serviço rodando com dados obsoletos por muito tempo.
   Isso explicaria a sensação de catálogo desatualizado.

3. A aplicação só aceita um formato funcional do payload e não faz nenhuma validação extra de freshness/consistência do catálogo recebido.
   Isso pode traduzir em desatualização quando o upstream entrega dados antigos ou incompletos.

## Risco específico no código atual

O ponto mais crítico é a combinação de:

- resposta vazia => erro fatal;
- sem snapshot do último catálogo válido;
- atualização in-memory direta sobre o estado atual;
- sem política de rejeição de payloads inválidos ou antigos.

Em outras palavras, um problema transitório na fonte pode causar perda real de catálogo na instância em execução.

## Plano de investigação e correção

### Investigação

1. Capturar logs do serviço em produção para confirmar se houve ciclos com `empty catalog returned by BeTor` ou `download_url ... no usable item list was found`.
2. Validar o payload real do upstream em momentos em que os itens ficam zerados.
3. Comparar o comportamento em produção com a resposta de `BETOR_SEARCH_DOWNLOAD_ITEMS_URL` ou `BETOR_SEARCH_API_BASE_URL` configurado.
4. Checar se a instância está recebendo payloads vazios, incompletos ou antigos em vez de uma atualização real.

### Correção a ser implementada

1. Manter um snapshot do último catálogo válido e reutilizá-lo quando um refresh falhar ou vier vazio.
2. Tratar payload vazio/incompleto como "degraded" e não como destruição do catálogo atual.
3. Adicionar validações de consistência antes de sobrescrever `s.items`.
4. Expor métricas/estado mais claros: `last_success`, `last_error`, `status`, `items_count`.
5. Cobrir com testes de regressão os cenários:
   - upstream retorna vazio após catálogo válido;
   - upstream retorna payload antigo/incompleto;
   - background sync falha sem zerar o catálogo atual.

## Arquivos relevantes

- `cmd/server/main.go`
- `internal/search/application/service.go`
- `internal/search/application/service_test.go`

## Status

Artefato de investigação inicial concluído. A próxima etapa será implementar a correção defensiva com testes cobrindo a regressão de catálogo zerado/desatualizado.
