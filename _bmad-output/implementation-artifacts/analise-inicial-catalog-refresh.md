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

## Verificação que eu recomendo antes de implementar

1. Confirmar se o processo do serviço está vivo há várias horas e se houve algum restart.
2. Verificar logs do container: buscar por mensagens como `initial catalog sync failed`, `catalog sync successful`, ou qualquer erro de rede.
3. Validar se a variável `BETOR_SEARCH_UPDATE_INTERVAL_MINUTES` está realmente definida no ambiente da instância.
4. Checar se a aplicação está em execução como processo único sem reinicialização do container.

## Recomendação de próxima etapa

Se a sua avaliação confirmar que a instância fica estável e o catálogo continua sem recarregar, a correção mais provável é implementar um ticker em background no bootstrap do serviço.

A implementação esperada seria:

- iniciar um `time.Ticker` com o intervalo configurado;
- chamar `Sync()` em loop;
- tratar falhas sem derrubar o processo;
- registrar logs de sucesso/erro; e
- manter o comportamento do health atualizado.

## Decisão de ação

- Se você confirmar que o processo não tem reinício e ainda assim não atualiza, eu sigo para a implementação da correção.
- Se você quiser, eu também posso abrir a implementação com estratégia adicional de tolerância a falhas e logs mais claros.
