# Guia: implementando o layout de um banco novo

Este guia mostra o passo a passo para transformar o manual CNAB 240 de um banco real em um pacote Go que se registra neste SDK, usando `layouts/febraban240` como ponto de partida. Leia `ARQUITETURA.md` antes, principalmente a seção sobre o vocabulário de `Key` e sobre os segmentos polimórficos (`SegmentB`/`SegmentBPix`, `SegmentN`/`SegmentNSimple`/`SegmentNSocial`).

## Passo 1: organize o pacote

Crie um pacote próprio, fora de `internal/`, por exemplo:

```
layouts/meubanco/
├── meubanco.go        (registro, tipo Layout, tabela de RecordKey -> RecordSpec)
├── file_header.go
├── batch_header.go
├── segment_a.go
├── segment_b.go
├── ...
└── meubanco_test.go
```

Copie a estrutura de `layouts/febraban240` e ajuste. A maioria dos bancos segue o esqueleto genérico do FEBRABAN quase à risca; o trabalho real costuma estar em confirmar posições específicas e códigos de domínio (tipo de serviço, forma de lançamento, código do banco).

## Passo 2: extraia as tabelas de posição do manual

Para cada tipo de registro que seu banco suporta (header de arquivo, header de lote, segmentos A/B/J/J-52/O/N, trailer de lote, trailer de arquivo), monte uma tabela com: posição inicial, posição final, tamanho, tipo (`9` numérico ou `X` alfanumérico), casas decimais implícitas (se houver) e conteúdo (fixo ou variável). A maioria dos manuais de banco já apresenta essa tabela prontamente; o trabalho é conferir se a numeração de colunas é 1-based (geralmente é) e se não há lacunas.

Confira que a soma dos tamanhos de cada registro é exatamente 240. Se não for, há um campo com posição errada ou um trecho "de uso exclusivo FEBRABAN/CNAB" que faltou modelar como preenchimento (filler).

## Passo 3: monte os `FieldSpec`

Cada linha da tabela do passo 2 se torna um `layout.FieldSpec`:

```go
layout.FieldSpec{
    Name:  "PaymentAmount",       // identificador em inglês, para mensagens de erro
    Start: 120,
    End:   134,
    Kind:  layout.KindNumeric,
    Decimals: 2,                   // só relevante para valores fora de Cents
    Key:   layout.KeyAmount,       // OU Const, nunca os dois
}
```

Duas regras:

- Se o conteúdo é **fixo** (marcador de tipo de registro, código de segmento, versão de layout, ou uma zona reservada em branco/zero), use `Const` e deixe `Key` vazio.
- Se o conteúdo **varia** por arquivo/lote/pagamento, use `Key` e deixe `Const` vazio. A chave deve vir do vocabulário já existente em `cnab/layout/value.go` sempre que o conceito já existir ali (por exemplo, `layout.KeyPayeeName`, `layout.KeyAmount`, `layout.KeyBarcode`). Só crie uma chave nova quando o conceito realmente não existir ainda no vocabulário; nesse caso, adicione a constante em `cnab/layout/value.go` com um comentário explicando o significado, e considere se algum tipo de pagamento na camada `cnab` precisa passar a preenchê-la.

Agrupe os `FieldSpec` de um registro em um `layout.RecordSpec`:

```go
var segmentASpec = layout.RecordSpec{
    Name: "segment_a",
    Fields: []layout.FieldSpec{
        numericConst("BankCode", 1, 3, "341"), // código do seu banco aqui
        numeric("ServiceBatchNumber", 4, 7, layout.KeyBatchNumber),
        // ...
    },
}
```

As funções auxiliares `numeric`, `numericConst`, `numericDecimal`, `alpha`, `alphaConst`, `numericFiller` e `alphaFiller` usadas em `layouts/febraban240` são só literais de `FieldSpec`; copie o padrão para o seu próprio pacote (ou importe as equivalentes, se preferir manter um único conjunto compartilhado).

## Passo 4: implemente o tipo `Layout`

```go
package meubanco

import "github.com/raykavin/gocnab/cnab/layout"

const bankCode = "341"
const version = "081" // confirme a versão exigida pelo manual do seu banco

type meubanco struct{}

func (meubanco) Name() string    { return "meubanco240" }
func (meubanco) Version() string { return version }

func (meubanco) Record(key layout.RecordKey) (layout.RecordSpec, bool) {
    switch key {
    case layout.FileHeader:
        return fileHeaderSpec, true
    case layout.SegmentA:
        return segmentASpec, true
    // ... cada RecordKey que seu banco suporta
    default:
        return layout.RecordSpec{}, false
    }
}

func init() {
    layout.Register("meubanco240", meubanco{})
}
```

Se o seu banco não suportar algum segmento (por exemplo, não processa GPS), simplesmente não inclua aquele `case`; o valor de retorno `ok=false` já faz o SDK rejeitar, em `Batch.AddPayment`, qualquer tipo de pagamento que precise daquele segmento, com uma mensagem de erro clara.

## Passo 5: decida sobre `SegmentB`/`SegmentBPix` e `SegmentN`/variações

Se o manual do seu banco documenta conteúdo diferente para o Segmento B em pagamentos comuns versus PIX (ou para o Segmento N em DARF/DARF Simples/GPS), implemente cada `RecordKey` correspondente separadamente, com posições fiéis ao manual de cada caso. Se o seu banco não suporta PIX ainda, basta não implementar `SegmentBPix`.

## Passo 6: registre o pacote e use

Do lado do consumidor do SDK, basta importar o pacote do banco (import em branco, já que o registro acontece em `init()`) e referenciar o nome em `Config.Layout`:

```go
import (
    "github.com/raykavin/gocnab/cnab"
    _ "github.com/example/meubanco240" // aciona o init() que registra "meubanco240"
)

file, err := cnab.NewRemittance(cnab.Config{
    Layout: "meubanco240",
    // ...
})
```

## Passo 7: teste o layout

No mínimo, replique os testes de `layouts/febraban240/febraban240_test.go`:

1. O layout se autorregistra (`layout.Lookup("meubanco240")` retorna `ok=true`).
2. Cada `RecordSpec` que o layout expõe soma exatamente 240 colunas.
3. `internal/engine.New(meubanco{})` não retorna erro (isso confere, adicionalmente, que não há sobreposição nem lacuna em nenhum registro, uma verificação mais forte que só somar tamanhos).

Depois disso, gere um arquivo de exemplo completo (um `File` com pelo menos um lote e um pagamento de cada tipo que seu banco suporta) e compare visualmente as posições de campos fixos e conhecidos (código do banco, tipo de registro, código de segmento) com um arquivo de remessa real do seu banco, se tiver um à mão. Esse é o jeito mais rápido de pegar um deslocamento de coluna errado.

## Alternativa: descrevendo o layout em JSON em vez de Go

Os passos 3 e 4 acima descrevem o caminho principal (structs Go, validado em tempo de compilação). Para casos em que o layout precisa ser carregado sem recompilar o binário (por exemplo, times que não programam em Go editando a posição de campos, ou um layout mantido fora do repositório), `cnab/layout` também oferece um carregador de JSON:

```go
func NewFromJSON(data []byte) (layout.Layout, error)
func NewFromJSONFile(path string) (layout.Layout, error)
```

O JSON usa exatamente o mesmo vocabulário dos passos anteriores: as chaves do objeto `records` são os valores de `RecordKey` (`"file_header"`, `"segment_a"`, etc.), e cada campo é um `FieldSpec` com `start`/`end`/`kind` (`"9"` ou `"X"`, também aceita `"numeric"`/`"alphanumeric"`) e `key` ou `const` (nunca os dois). Exemplo mínimo:

```json
{
  "name": "meubanco240",
  "version": "081",
  "records": {
    "file_header": {
      "fields": [
        {"name": "BankCode", "start": 1, "end": 3, "kind": "9", "const": "341"},
        {"name": "ServiceBatchNumber", "start": 4, "end": 7, "kind": "9", "const": "0000"},
        {"name": "RecordType", "start": 8, "end": 8, "kind": "9", "const": "0"},
        {"name": "CompanyName", "start": 73, "end": 102, "kind": "X", "key": "company_name"}
      ]
    }
  }
}
```

Uso:

```go
l, err := layout.NewFromJSONFile("meubanco240.json")
if err != nil {
    log.Fatal(err) // erro de validação já aponta o registro e o campo exatos
}
cnab.RegisterLayout("meubanco240", l)
```

`NewFromJSON`/`NewFromJSONFile` validam a configuração já no carregamento, sem esperar o `Generate()` de um arquivo real para revelar o problema:

- `name` e `version` são obrigatórios.
- toda chave de `records` precisa ser um `RecordKey` conhecido (a mensagem de erro lista os valores válidos).
- todo `key` de campo precisa ser uma chave conhecida do vocabulário (`cnab/layout/value.go`); um nome de chave digitado errado é rejeitado na hora, com o registro e o número do campo que causou o erro.
- um campo nunca pode ter `key` e `const` ao mesmo tempo.
- cada registro precisa cobrir as colunas 1 a 240 sem lacuna nem sobreposição (a mesma verificação que `RecordSpec.Validate()` e o motor fazem).

Só oferecemos JSON, não YAML: a biblioteca padrão do Go tem `encoding/json`, então o carregador de JSON não adiciona nenhuma dependência externa ao projeto; um carregador de YAML exigiria uma dependência externa (não existe parser de YAML na stdlib), o que contraria a decisão de design registrada em `ARQUITETURA.md`. Se seu banco só é descrito em YAML, converta para JSON antes (ou peça para reavaliarmos essa decisão caso o time realmente precise editar YAML diretamente).

O JSON é uma via alternativa para o **mesmo** `RecordSpec`/`FieldSpec` dos passos 3 e 4, não um formato paralelo: qualquer coisa que você conseguiria expressar em Go, consegue expressar em JSON, e vice-versa. Escolha Go quando quiser checagem em tempo de compilação e o layout for versionado junto do código; escolha JSON quando quiser carregar/trocar o layout sem recompilar.

Um exemplo completo e executável está em `examples/custom_layout_json` (`go run ./examples/custom_layout_json`): o descritor completo em `layout.json`, embarcado no binário com `//go:embed`, carregado com `layout.NewFromJSON`, registrado com `cnab.RegisterLayout` e usado para gerar uma remessa de crédito em conta, exatamente como qualquer outro layout.

## Erros comuns

- **Soma de colunas diferente de 240**: normalmente uma zona "de uso exclusivo FEBRABAN/CNAB" que faltou modelar como `numericFiller`/`alphaFiller`.
- **Reaproveitar uma chave semântica com sentido diferente**: por exemplo, usar `KeyTaxpayerDocumentKind` (convenção geral "1=CPF, 2=CNPJ") num campo cujo domínio real é "1=CNPJ, 2=CPF" (é o caso do campo específico de tipo de identificação do Segmento N; veja `layout.KeyTaxpayerIdType` em `cnab/layout/value.go` e `taxpayerIdType` em `cnab/tax_n.go` para o padrão usado nesse caso). Sempre confira o domínio de valores do campo no manual antes de reaproveitar uma chave existente.
- **Esquecer que `KeyBatchNumber`/`KeySequence`/os totais de trailer são escritos pelo motor**: não crie um `FieldSpec` para esses campos usando `Const`; sempre use `Key` apontando para a constante estrutural correspondente, e deixe o motor preencher o valor.
