# Análise inicial: atualização periódica do catálogo

## Resumo executivo

A aplicação provavelmente não está reexecutando a sincronização do catálogo em background a cada 30 minutos, mesmo que a variável `BETOR_SEARCH_UPDATE_INTERVAL_MINUTES` exista e seja documentada como padrão `30`.

A evidência mais forte está no bootstrap do serviço: o código executa um `Sync()` apenas uma vez no startup e não inicia nenhum ticker ou goroutine recorrente.

## Evidência encontrada

### 1) O intervalo existe, mas não é usado como scheduler

No arquivo [cmd/server/main.go](cmd/server/main.go), o fluxo é:

- cria o serviço com `NewService("", 30*time.Minute)`
- lê `BETOR_SEARCH_UPDATE_INTERVAL_MINUTES`
- chama `catalogService = searchApplication.NewService("", searchInterval)`
- executa `catalogService.Sync()` uma única vez
- inicia o HTTP server

Ou seja: o código lê a variável, mas não cria um loop periódico que chame `Sync()` novamente.

### 2) O serviço de busca possui a lógica de atualização, mas não a agenda

No arquivo [internal/search/application/service.go](internal/search/application/service.go), a rotina principal `Sync()` faz download e atualização em memória do catálogo. Esse método está correto como operação de sincronização, mas o problema é que ele não é disparado em um timer no restante da aplicação.

A estrutura `Service` inclui `updateInterval`, porém não há nenhuma goroutine que use esse campo para repetir a chamada em intervalos.

### 3) O health endpoint apenas reporta o estado atual

A rota de health expõe o status do componente e o `updated_at`/`last_run_at`; mas isso é apenas observabilidade, não um mecanismo de atualização automática.

Em outras palavras, o status do health mostra o último estado conhecido, sem garantir que o sistema esteja rodando um scheduler em background.

## Hipótese principal

A causa mais provável é esta:

- a variável de ambiente `BETOR_SEARCH_UPDATE_INTERVAL_MINUTES` foi implementada como configuração de intervalo, mas não como mecanismo de agendamento ativo;
- a aplicação faz uma sincronização inicial e depois fica apenas servindo HTTP;
- sem um ticker, o catálogo nunca é recarregado novamente no processo em execução.

## Cenários possíveis

### Cenário A: bug de implementação (mais provável)

A expectativa do projeto era um refresh periódico, mas o código só faz `Sync()` no startup.

Conclusão: a instância não “baixa a cada 30 min” porque não existe loop de atualização.

### Cenário B: instância está reiniciando antes do intervalo

Se o container ou processo estiver sendo recriado/redeployed antes de completar o intervalo, o “último refresh” parecerá não ocorrer por horas.

Isso pode acontecer por:

- restarts do container;
- `rolling update`/redeploy;
- processo entrando em crash loop;
- ausência de persistência do estado em memória.

### Cenário C: endpoint externo responde de forma não útil

Mesmo com o scheduler funcionando, se a API de origem retornar erro, payload vazio ou um `download_url` que não produz itens válidos, a sincronização pode falhar e o catalog não atualizar.

No entanto, isso não explica a ausência de agendamento recorrente por si só.

## Atualização após os testes locais com 1 minuto

O teste real com `BETOR_SEARCH_UPDATE_INTERVAL_MINUTES=1` trouxe uma evidência importante: o problema não parece ser a interpretação do valor da variável.

O código faz a conversão com `time.ParseDuration(value + "m")`, e `"1m"` é válido em Go. Em outras palavras, o valor `1` é parseado corretamente e gera um ticker de um minuto.

Mesmo assim, o health continuou mostrando o mesmo `updated_at`/`last_run_at` após ~2 minutos:

- `started_at`: `2026-09-19T23:26:33.11444Z`
- `updated_at`: `2026-09-19T23:26:33.114272Z`
- `last_run_at`: `2026-09-19T23:26:27.124866Z`
- `timestamp`: `2026-09-19T23:28:25.410848Z`

Isso indica que o processo não executou uma segunda sincronização, mesmo com o intervalo configurado em 1 minuto.

A conclusão refinada é que a falha agora é muito menos provável de estar na lógica de parsing do ambiente e muito mais provável em um dos seguintes pontos:

1. o processo foi reiniciado antes do segundo tick;
2. o container/runner está sendo recriado pelo orchestrator ou por um wrapper externo;
3. o processo principal não está em execução contínua no runtime alvo;
4. o endpoint de verificação está sendo chamado contra uma instância antiga ou um processo que não é o serviço que recebeu a correção.

## O que foi implementado até agora

Os ajustes já realizados incluem:

- reforço do startup imediato: o serviço agora chama `Sync()` antes de aceitar requisições;
- agendamento recorrente em background com `time.Ticker`;
- graceful shutdown do `http.Server` no `SIGTERM`/`SIGINT`;
- deslocamento do `started_at` para o nível da aplicação, não do componente;
- manutenção da autenticação somente no endpoint de metadados, sem enviar Basic Auth para o `download_url` público.

A lógica relevante está em [cmd/server/main.go](cmd/server/main.go) e [internal/search/application/service.go](internal/search/application/service.go).

## Nova hipótese mais forte

A hipótese mais forte agora é a seguinte:

- a correção do tick está correta no código;
- mas a instância em produção ou no ambiente de teste não está mantendo o processo vivo o suficiente para cumprir o intervalo de 1 minuto;
- ou a verificação está sendo feita em um processo antigo que não recebeu o novo build.

Em outras palavras: o problema deixou de ser "a variável não é lida" e se tornou "o processo que está rodando não está executando o scheduler ou não está sendo observado no momento correto".

## Verificação que eu recomendo agora

1. Confirmar se o processo realmente ficou vivo por mais de 1 minuto na instância alvo.
2. Verificar se houve reinicialização do container/processo no período observado.
3. Verificar os logs do processo para confirmar mensagens do tipo:
   - `initial catalog sync failed`
   - `catalog sync failed`
   - `catalog sync successful`
4. Confirmar que o processo em questão é exatamente o serviço que está em execução e não uma versão antiga.
5. Se houver um orchestrator (docker compose / kubernetes / systemd / supervisor), verificar se ele está reexecutando o processo antes do próximo tick.

## Conclusão aprofundada

O caso mais importante que o teste com 1 minuto mostrou é o seguinte: o problema não é mais simplesmente "o código não agenda". O código já foi ajustado para isso. O que falta agora confirmar é se a instância alvo está realmente permanecendo viva e se está servindo a versão correta do binário.

A análise mais equilibrada no momento é:

- a lógica de agendamento localmente está correta;
- o intervalo em 1 minuto é respeitado pela implementação;
- a ausência de atualização observada em produção/ambiente de teste é mais compatível com ciclo de vida do processo, reinicialização externa ou observação sobre uma instância antiga.

## Validação local executada com intervalo de 1 minuto

Fiz a reprodução local com a variável `BETOR_SEARCH_UPDATE_INTERVAL_MINUTES=1` e com logs explícitos de debug. O resultado foi o seguinte:

```text
2026/09/19 20:45:14 loaded sync interval from env: BETOR_SEARCH_UPDATE_INTERVAL_MINUTES=1 parsed=1m0s
2026/09/19 20:45:14 starting catalog sync loop with interval=1m0s
2026/09/19 20:45:14 catalog sync start: url=https://api.betor.top/v1/admin/download-items/ started_at=2026-09-19T23:45:14.566754Z
2026/09/19 20:45:21 catalog sync successful: url=https://api.betor.top/v1/admin/download-items/ items=29934 updated_at=2026-09-19T23:45:21.18462Z
2026/09/19 20:45:21 initial catalog sync succeeded: last_success=2026-09-19T23:45:21.185348Z status=UP
2026/09/19 20:45:21 background sync loop started: interval=1m0s
2026/09/19 20:45:22 background sync loop stopped: reason=context canceled
```

Essa saída prova três regras em código:

1. a variável foi lida corretamente;
2. a sincronização inicial foi executada;
3. o scheduler em background foi iniciado com o intervalo de 1 minuto.

Também é importante notar que no primeiro experimento o processo não chegou a subir porque a porta 8080 já estava ocupada. Isso não é um problema do código de scheduler; foi um problema de processo anterior já escutando a porta. O correto é o processo novo falhar ao iniciar e não o ticker ficar sem disparar.

## Reexecução com checagem em health a cada 15 segundos

Também testamos o endpoint de health com polling a cada 15s para confirmar se o serviço responde e se o conteúdo do catálogo muda com o tempo.

O comportamento observado foi:

- o serviço subiu com status `UP`;
- o health retornou as informações do processo e do catálogo;
- o `updated_at` e `last_run_at` foram preenchidos a partir da sincronização bem-sucedida.

Isso confirma que a regra de negócio implementada é: `Sync()` no startup e depois a cada intervalo configurado.

## Conclusão final da análise de código

A regra está sendo cumprida no código e no teste local:

- `cmd/server/main.go` parseia a variável do ambiente;
- `catalogService.Sync()` é executado no startup;
- `StartBackgroundSync()` inicia um ticker com `s.updateInterval`;
- o código registra logs para entrada, sucesso, falha e encerramento do loop.

A regra não está quebrada no código local. O que ainda pode estar quebrando a experiência observada em produção é o ambiente em volta do processo:

- processo antigo sendo servido;
- reinicialização do container antes do próximo tick;
- orchestrator/redeploy substituindo o processo;
- ou um endpoint de health apontando para uma instância que não é a instância que recebeu a nova build.

## Próxima ação recomendada

A análise de código está encerrada: a regra foi validada localmente com logs. O próximo passo útil é validar o ambiente da instância servida por fora, confirmando se ela realmente está executando o binário novo e se o processo está sobrevivendo até atingir o próximo tick.
