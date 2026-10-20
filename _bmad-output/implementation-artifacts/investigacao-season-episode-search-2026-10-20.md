# Investigação: inconsistência de season/episode no endpoint de busca

## Resumo executivo

A inconsistência observada no primeiro item retornado por `https://search.betor.top/v1/search/` está ligada ao contrato de ingestão do item e ao modelo usado pela aplicação de busca.

O payload da fonte (exemplo do item `6ad7aa4533317cb560b16bd1`) traz `episodes` como uma lista de objetos no formato:

```json
"episodes": [
  { "season": 2, "episode": 10 },
  { "season": 2, "episode": 1 }
]
```

Porém, o modelo interno em `internal/search/application/service.go` define `Episode` como:

```go
type Episode struct {
    Number int    `json:"number"`
    Title  string `json:"title,omitempty"`
}
```

Ou seja, o parser tenta ler `number` e `title`, mas a fonte envia `season` e `episode`. Quando o JSON é decodificado, esses campos são ignorados e o valor de `Number` fica zerado. Isso explica a ausência de `season`/`episode` corretos na resposta pública.

## Evidência do payload da fonte

No item que você anexou, a estrutura da fonte é:

```json
"episodes": [
  {
    "season": 2,
    "episode": 10
  },
  {
    "season": 2,
    "episode": 1
  }
],
"seasons": [2]
```

Esse padrão não bate com o campo esperado pelo código da aplicação.

## Evidência do código

### 1) Modelo interno usa campos errados

Arquivo: `internal/search/application/service.go`

Trecho relevante:

```go
type Episode struct {
    Number int    `json:"number"`
    Title  string `json:"title,omitempty"`
}
```

Além disso, o `Item` expõe `Episodes []Episode` e `Seasons []int`:

```go
type Item struct {
    Episodes []Episode `json:"episodes,omitempty"`
    Seasons  []int     `json:"seasons,omitempty"`
}
```

### 2) O parser decodifica sem mapear `season` e `episode`

No `UnmarshalJSON` da struct `Item`, a propriedade `Episodes` é simplesmente atribuída ao valor bruto do JSON:

```go
Episodes        []Episode       `json:"episodes"`
Seasons         []int           `json:"seasons"`
```

Como a estrutura alvo do elemento `Episode` não contém `season` nem `episode`, o unmarshal não preenche `Number`.

### 3) Verificação de reprodução mínima

Conferi com um snippet de Go usando exatamente o mesmo padrão de decodificação:

```go
payload := []byte(`{"episodes":[{"season":2,"episode":10},{"season":2,"episode":1}],"seasons":[2]}`)
var item Item
json.Unmarshal(payload, &item)
```

Resultado observado na execução:

```text
episodes=[]main.Episode{main.Episode{Number:0, Title:""}, main.Episode{Number:0, Title:""}}
seasons=[2]
```

Esse resultado confirma que o código atual não consegue materializar a informação de temporada/episódio a partir do payload real da fonte.

## Hipótese principal

A causa mais provável é um desvio de contrato entre as duas camadas:

- a fonte produz `episodes` com objetos `season`/`episode`;
- a aplicação de busca espera `episodes` com `number`/`title`;
- a API pública expõe o resultado da decodificação sem normalização, então o cliente recebe valores vazios ou incorretos.

Também há um sinal adicional: o parser HTTP de filtros em `ParseHTTPQuery` não trata `season` nem `episode`, ou seja, even que a estrutura interna fosse ajustada, o contrato de busca também não está modelando explicitamente esses campos como query params.

## Impacto

- O endpoint público pode listar um item com `episode` zerado, sem o número correto.
- A listagem do frontend pode mostrar o primeiro item como se não tivesse `season`/`episode` válidos.
- Há risco de bugs em filtros por temporada/episódio, porque o filtro não interpreta esses campos de maneira explícita.

## Conclusão

O problema parece estar na modelagem e no mapeamento de campos, não na ordenação da listagem em si. A lista pode até aparecer como primeiro item por critério de frescor/recência, mas a causa direta da ausência de `season`/`episode` corretos é o mismatch entre o payload da fonte e a estrutura interna do serviço.

## Próxima etapa recomendada

Se você aprovar, eu posso corrigir com um patch focado em:

1. normalizar o payload da fonte para o formato interno esperado;
2. mapear `season` e `episode` corretamente; e
3. adicionar teste de regressão que cubra exatamente esse caso do item `6ad7aa4533317cb560b16bd1`.
