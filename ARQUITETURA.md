# Arquitetura

Este documento explica as três camadas do SDK, por que elas existem separadas e como se conectam. Para o passo a passo de implementação de um banco novo, veja `NOVO-BANCO.md`.

## Visão geral

```
┌─────────────────────────────────────────────────────────────┐
│  cnab (API pública, camada 3)                               │
│  Config, Company, Account, Payee, File, Batch,              │
│  CreditAccount, TED, Pix, PixBankData, BoletoPayment,       │
│  BarcodeTax, DARF, DARFSimple, GPS, CancelPayment           │
└───────────────┬───────────────────────────┬─────────────────┘
                │                           │
                ▼                           ▼
┌───────────────────────────┐    ┌─────────────────────────────────┐
│ internal/engine (camada 1)│    │ cnab/layout (camada 2, contrato)│
│ preenchimento de campos,  │◄───┤ Layout, FieldSpec, RecordSpec,  │
│ montagem de registros,    │    │ Values, Key, Register/Lookup    │
│ sequenciais, trailers,    │    └─────────────────────────────────┘
│ limites do padrão         │                   ▲
└───────────────────────────┘                   │
                                    ┌───────────┴────────────┐
                                    │ layouts/febraban240    │
                                    │ (layout de referência) │
                                    └────────────────────────┘
```

`cnab` é a única camada que o usuário do SDK toca. `cnab/layout` é um pacote separado, pensado para quem vai *escrever* o descritor de um banco, não para quem vai *usar* o SDK no dia a dia. `internal/engine` nunca é importável de fora do módulo.

## Por que `cnab/layout` é um subpacote, e não parte de `cnab`

A tentação natural é colocar a interface `Layout` e o registro de layouts dentro do próprio pacote `cnab`, já que o exemplo de uso do SDK escreve `cnab.RegisterLayout(...)`. O problema é a direção das dependências: o motor (`internal/engine`) precisa conhecer o formato de `FieldSpec`/`RecordSpec` para conseguir renderizar um registro, mas o motor não pode importar `cnab` (isso criaria um ciclo, já que `cnab` importa o motor para gerar o arquivo).

A solução é a mesma usada por `database/sql`/`database/sql/driver` na biblioteca padrão do Go: um pacote pequeno e estável (`cnab/layout`, análogo a `database/sql/driver`) define o contrato (`Layout`, `FieldSpec`, `RecordSpec`, o vocabulário de `Key`, `Register`/`Lookup`), sem depender de `cnab` nem do motor. O motor importa `cnab/layout` para conhecer o formato dos descritores. O pacote `cnab` importa os dois e expõe `cnab.RegisterLayout` como um wrapper fino sobre `layout.Register`, e `cnab.Layout` como um alias de `layout.Layout`. Do ponto de vista de quem só usa o SDK, nada muda: a chamada continua sendo `cnab.RegisterLayout(...)`. Quem vai implementar um banco novo (possivelmente em outro módulo Go) depende só do pacote pequeno `cnab/layout`, não do pacote `cnab` inteiro.

## Camada 1: o motor (`internal/engine`)

O motor não tem nenhuma referência a bancos. Ele só sabe:

- **Preencher um campo** (`field.go`): campos numéricos (`9`) são alinhados à direita e completados com zeros à esquerda; um valor que não cabe no tamanho do campo é erro (nunca trunca dinheiro). Campos alfanuméricos (`X`) são maiúsculizados, alinhados à esquerda e completados com espaços à direita; um valor mais longo que o campo é truncado (comportamento real de bancos para nomes longos). Casas decimais implícitas (notação `9(13)V2`) não multiplicam nada em tempo de execução: valores monetários chegam da camada 3 como inteiros no `Cents` (já na unidade mínima da moeda), então o preenchimento é só zero-padding; para outros campos numéricos com decimais, uma string como `"12.5"` é escalonada com aritmética inteira sobre a própria string, nunca com `float64`.
- **Validar o conjunto de caracteres** (`charset.go`): dígitos, letras maiúsculas, espaço e um conjunto de símbolos (incluindo `@` e `_`, necessários para chaves PIX por e-mail). Um caractere fora desse conjunto (por exemplo, uma letra acentuada) é erro, não é descartado silenciosamente. A função `Sanitize` (também exposta como `cnab.Sanitize`) faz a limpeza de acentos e caracteres inválidos de forma explícita e opcional, exatamente como pedido: "validação... com sanitização opcional".
- **Montar um registro de 240 colunas** (`record.go`): ao compilar um `RecordSpec`, o motor ordena os campos por posição inicial e garante que eles cobrem as colunas 1 a 240 sem lacuna nem sobreposição. Isso acontece uma vez, quando o `Engine` é criado, não em cada linha renderizada.
- **Ler um registro de 240 colunas** (`parse.go`): o inverso exato de montar um registro, usado por `cnab.ParseReturn` para decodificar um arquivo de retorno. Um campo com `Key` é extraído da faixa de colunas e devolvido como string (dígitos crus para campo numérico, texto sem os espaços de preenchimento à direita para alfanumérico) quem lê decide se aquilo é um inteiro, uma data ou só um código, exatamente como quem escreve decide o que colocar em `Values` antes de renderizar. Um campo `Const` não é exposto como dado (não tem `Key` para guardar valor), mas seu conteúdo é comparado contra o que a própria renderização produziria para aquele campo (reaproveitando `renderField`, não uma segunda implementação da regra de padding); uma divergência é reportada como `*RecordMismatchError`, nunca ignorada ela normalmente significa um arquivo corrompido ou classificado sob a `RecordKey` errada.
- **Montar o arquivo inteiro** (`build.go`, `trailer.go`, `limits.go`): o motor escreve o header de arquivo, cada lote (header, movimentos, trailer) e o trailer de arquivo, cada linha terminada em `\r\n` (CRLF, hexa 0D0A). Os sequenciais que o padrão exige (número do lote começando em `0001` e incrementando, `0000` no header de arquivo, `9999` no trailer de arquivo, número do registro no lote reiniciando em `00001` a cada lote) são escritos pelo próprio motor em chaves reservadas (`layout.KeyBatchNumber`, `layout.KeySequence`) que a camada de domínio nunca define diretamente. Os totais de trailer (quantidade de registros do lote, soma dos valores, quantidade de lotes, quantidade total de registros) são calculados a partir das linhas efetivamente renderizadas, o que evita por construção uma divergência entre trailer e conteúdo. Os limites do padrão (70 lotes por arquivo, 10.000 movimentos por lote) retornam um erro tipado e explícito quando excedidos.
- **Classificar uma linha** (`classify.go`): `ClassifyLine` lê o tipo de registro (coluna 8) e, para um registro de detalhe, o código de segmento (coluna 14) de uma linha qualquer, sem consultar nenhum `Layout`. Essas duas posições são uma convenção do próprio padrão FEBRABAN, não algo que um layout de banco possa mover pela mesma razão que `recordWidth = 240` é uma constante do pacote em vez de algo que `Layout` declara. É assim que `cnab.ParseReturn` descobre qual `RecordKey` uma linha representa antes de ter qualquer outra forma de saber.

## Camada 2: descritores de layout (`cnab/layout`)

Um layout é, para o motor, só um mapa de `RecordKey` (header de arquivo, header de lote, segmento A, segmento B, etc.) para um `RecordSpec` (lista de `FieldSpec`). A interface é:

```go
type Layout interface {
    Name() string
    Version() string
    Record(key RecordKey) (spec RecordSpec, ok bool)
}
```

Um banco novo se registra chamando `cnab.RegisterLayout("nome-do-banco", MeuLayout{})`, tipicamente dentro de uma função `init()` do seu próprio pacote, do mesmo jeito que um driver de `database/sql` se registra. Isso é o que permite adicionar bancos "sem alterar o core": o motor e o pacote `cnab` nunca são recompilados nem editados para reconhecer um banco novo, eles só consultam o registro em tempo de execução pelo nome informado em `Config.Layout`.

Para quem precisa carregar um layout sem recompilar (por exemplo, definir o layout fora do repositório do SDK), `cnab/layout` também oferece `NewFromJSON`/`NewFromJSONFile`, que fazem o parse do mesmo `RecordSpec`/`FieldSpec` a partir de um arquivo JSON, com validação completa (chave de registro conhecida, `key` semântica conhecida, `key`/`const` mutuamente exclusivos, cobertura de 1 a 240 sem lacuna/sobreposição) já no carregamento. Veja a seção "Alternativa: descrevendo o layout em JSON em vez de Go" em `NOVO-BANCO.md` para o formato completo.

### A ponte genérica: o vocabulário de `Key`

O motor só sabe fazer `values[fieldSpec.Key]`. Ele não sabe o que é um "nome de favorecido". O que faz a ponte entre a camada de domínio (que sabe o que é um favorecido) e um layout de banco (que só sabe posições de coluna) é um vocabulário estável de chaves, exportado em `cnab/layout` como constantes `Key`, dividido em duas famílias:

- **Chaves estruturais** (prefixo `sys.`, escritas só pelo motor): `KeyBatchNumber`, `KeySequence`, `KeyBatchRecordCount`, `KeyBatchAmount`, `KeyBatchCount`, `KeyFileRecordCount`. Código fora de `internal/engine` nunca deve definir essas chaves.
- **Chaves semânticas** (escritas pela camada `cnab`, lidas pelos `FieldSpec.Key` de cada layout): `KeyAmount`, `KeyPaymentDate`, `KeyPayeeName`, `KeyPixKeyValue`, `KeyBarcode`, `KeyTaxCode`, e assim por diante (lista completa em `cnab/layout/value.go` e em `docs/API.md`).

Qualquer banco, presente (`febraban240`) ou futuro, amarra suas colunas físicas a essas mesmas chaves semânticas. Qualquer tipo de pagamento da camada 3 popula essas mesmas chaves, independente de qual layout está ativo. Se um banco realmente precisar de um campo sem equivalente semântico ainda, o vocabulário é estendido de forma aditiva em `cnab/layout`, nunca dentro do motor.

### Segmentos polimórficos: `SegmentB`/`SegmentBPix` e `SegmentN`/`SegmentNSimple`/`SegmentNSocial`

O padrão FEBRABAN real reaproveita a mesma letra de segmento para conteúdos fisicamente diferentes dependendo do tipo de lançamento: o Segmento B de um crédito em conta comum carrega endereço, enquanto o Segmento B de uma transferência PIX carrega a chave PIX nas mesmas colunas. O mesmo acontece no Segmento N entre DARF Normal, DARF Simples e GPS. Como cada `RecordSpec` neste SDK é uma lista fixa e simples de campos, esse polimorfismo é modelado como `RecordKey`s distintos (`SegmentB` e `SegmentBPix`; `SegmentN`, `SegmentNSimple` e `SegmentNSocial`) em vez de um único `RecordSpec` com significado variável. Cada tipo de pagamento da camada 3 já sabe qual `RecordKey` usar (`Pix` usa `SegmentBPix`, `CreditAccount`/`TED`/`PixBankData` usam `SegmentB`).

## Camada 3: API pública (`cnab`)

O usuário nunca vê posição de campo. Ele descreve:

- **Quem envia**: `Company` (nome, CNPJ/CPF, convênio) e `Account` (agência, conta, dígito).
- **Quem recebe**: `Payee` (nome, CNPJ/CPF).
- **O que está sendo pago**: um valor concreto de `Payment` (`CreditAccount`, `TED`, `Pix`, `PixBankData`, `BoletoPayment`, `BarcodeTax`, `DARF`, `DARFSimple`, `GPS`, `CancelPayment`).

`Payment` é uma interface fechada (seus métodos não são exportados): só os tipos definidos neste pacote a implementam. Cada tipo sabe se validar (`validate`, campos obrigatórios, valores positivos, data não retroativa) e se converter em um ou mais `DetailSegment` (`toSegments`), usando só o vocabulário de `Key` da camada 2.

`File` e `Batch` orquestram a montagem: `NewRemittance` valida a configuração e resolve o layout pelo nome; `NewBatch` cria um lote associado a um `BatchProduct`/`BatchService`; `AddPayment` valida o pagamento, converte em segmentos e confere que o layout ativo realmente suporta os segmentos necessários (rejeitando na hora um tipo de pagamento incompatível com o banco escolhido, em vez de só falhar depois em `Generate`); `Generate` monta o `engine.FileInput` completo e delega ao motor.

### Onde cada validação acontece

| Regra | Momento |
|---|---|
| Campos obrigatórios de `Company`/`Account`, dígito verificador de CNPJ/CPF | `NewRemittance` |
| Campos obrigatórios do favorecido/pagamento, valores não positivos | `Batch.AddPayment` |
| Data de pagamento no passado | `AddPayment` e novamente em `Generate` (o tempo passa entre as duas chamadas) |
| Layout ausente ou não registrado, versão de layout ausente | `NewRemittance` e validação eager do motor |
| Limites de 70 lotes / 10.000 movimentos | `NewBatch`/`AddPayment` (falha imediata) |
| Sequencial fora de ordem, trailer divergente | Evitado por construção (só o motor calcula esses valores); `Generate` ainda roda uma conferência defensiva do total de registros, que só deveria disparar em caso de bug no próprio layout |
| Caracteres inválidos | Validado na renderização de cada campo alfanumérico; use `cnab.Sanitize` antes de montar o pagamento se os dados de entrada podem conter acentos |

## Processando retorno

Um arquivo de retorno tem exatamente a mesma estrutura física de 240 colunas que uma remessa é o mesmo layout, na direção contrária. `cnab.ParseReturn(layoutName, content)` (`return.go`) faz o caminho inverso de `NewRemittance`+`Generate`:

1. Divide `content` em linhas de 240 colunas (aceita CRLF ou apenas LF, tolera uma linha vazia final).
2. Classifica cada linha com `engine.ClassifyLine` (colunas 8 e 14, fixas pelo padrão).
3. Ignora header/trailer de arquivo e de lote, e qualquer segmento de detalhe que não seja o Segmento A.
4. Decodifica cada Segmento A com `engine.Engine.ParseRecord`, extraindo `KeyYourNumber`, `KeyAmount`, e três chaves novas que só um arquivo de retorno preenche: `KeySettlementDate`, `KeySettlementAmount` e `KeyOccurrenceCodes` ("data/valor real da efetivação do pagamento" e "códigos/motivos de ocorrências para retorno").

Por que só o Segmento A: ele já carrega tudo que uma reconciliação de pagamento precisa (liquidou? quando? por quanto? se não liquidou, por quê? qual é o "seu número" para casar de volta com o pagamento original). Os Segmentos B/BPix carregam documento/endereço/chave PIX do favorecido dados que quem chama já tem nos seus próprios registros, e que, no caso de B vs. BPix, exigiriam olhar o Segmento A anterior (ou o conteúdo físico do próprio B) para desambiguar qual dos dois layouts uma linha "B" representa, já que o padrão FEBRABAN reaproveita a mesma letra para os dois. Nenhum desses dois problemas vale a complexidade adicional para o que `ParseReturn` precisa entregar hoje: são pulados, não um erro.

As três chaves novas foram acrescentadas a `layouts/febraban240/segment_a.go` reaproveitando colunas que já existiam como `Const "0"`/preenchimento em branco (o padrão sempre reservou essas colunas ali; só não havia leitor para elas). Isso é compatível com qualquer código de geração existente: uma chave de `Values` não definida renderiza exatamente igual ao `Const` que ela substituiu (ver `resolveValue` em `field.go`), então nenhum tipo de pagamento precisou mudar.

O motivo de banco (`ReturnMovement.OccurrenceCodes`) sai como os códigos de 2 dígitos crus, não traduzidos a tabela que mapeia um código para o que ele significa é específica de cada banco, o mesmo motivo pelo qual `TED.Purpose` já saía como código sem tradução embutida no SDK.

## Cobertura de testes

Cada camada tem sua própria suíte de testes: `internal/engine` cobre preenchimento/alinhamento/truncamento, leitura simétrica de campo e registro (incluindo o caso de divergência em campo `Const`), classificação de linha, montagem completa de arquivo com verificação byte a byte (incluindo o CRLF), cálculo de trailers e sequenciais com múltiplos lotes, e os casos de limite excedido, mantendo cobertura acima de 85%. `cnab/layout` cobre o registro (duplicidade, concorrência). `layouts/febraban240` confere que cada registro cobre exatamente as 240 colunas e que o motor aceita o layout inteiro. `cnab` cobre a validação de cada tipo de pagamento, o fluxo ponta a ponta contra `febraban240`, `ParseReturn` contra uma remessa real gerada pelo próprio SDK com as colunas de retorno preenchidas manualmente (não há amostra de um banco real para um layout de referência que não representa banco nenhum), e a tradução de erro na fronteira pública tanto para `ParseReturn` quanto para `Generate` (este via um número de conta de favorecido maior do que a coluna do layout suporta, que `AddPayment` não tem como recusar antes de o motor tentar renderizá-lo).