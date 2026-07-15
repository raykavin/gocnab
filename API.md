# Referência da API

Referência completa de todos os tipos, funções, métodos e constantes exportados pelos pacotes públicos `cnab` e `cnab/layout`. Para uma visão geral da arquitetura, veja `ARQUITETURA.md`; para exemplos completos e executáveis, veja a pasta `./examples` na raiz do módulo.

Convenções usadas neste documento: valores monetários são sempre `cnab.Cents` (inteiro, centavos); datas são sempre `time.Time`; todo método/função que pode falhar retorna `error` como último valor.

## Sumário

- [Configuração, empresa e conta](#configuração-empresa-e-conta)
- [Documentos (CPF/CNPJ)](#documentos-cpfcnpj)
- [Dinheiro](#dinheiro)
- [Arquivo e lote](#arquivo-e-lote)
- [Produtos e serviços de lote](#produtos-e-serviços-de-lote)
- [Tipos de pagamento](#tipos-de-pagamento)
- [PIX](#pix)
- [Registro de layouts](#registro-de-layouts)
- [Erros](#erros)
- [Pacote `cnab/layout`](#pacote-cnablayout)

---

## Configuração, empresa e conta

### `type Config`

```go
type Config struct {
    Layout  string
    Company Company
    Account Account
    NSA     int
}
```

Agrupa tudo que `NewRemittance` precisa para iniciar um arquivo de remessa.

- `Layout`: nome de um layout já registrado, por exemplo `"febraban240"`.
- `Company`: a empresa pagadora, que envia o arquivo.
- `Account`: a conta bancária de onde saem os pagamentos do arquivo.
- `NSA`: número sequencial do arquivo, positivo, controlado pelo chamador entre os arquivos que ele envia a um mesmo banco.

### `type Company`

```go
type Company struct {
    Name         string
    Registration Document
    Agreement    string
}
```

Identifica a empresa que envia a remessa (a pagadora). `Registration` aceita qualquer `Document` (`CNPJ` ou `CPF`). `Agreement` é o código de convênio da empresa junto ao banco.

Validação (em `NewRemittance`): `Name` e `Agreement` não podem ser vazios; `Registration` não pode ser `nil`. Erro: `*ValidationError{Context: "Company", ...}`.

### `type Account`

```go
type Account struct {
    Branch     string
    Number     string
    CheckDigit string
}
```

Identifica uma conta bancária: agência, número da conta (sem o dígito) e o dígito verificador. Usada tanto para a conta da empresa (em `Config`) quanto para a conta do favorecido (embutida em `CreditAccount`, `TED` e `PixBankData`).

Validação: os três campos são obrigatórios. Erro: `*ValidationError{Context: "Account", ...}`.

### `type Payee`

```go
type Payee struct {
    Name         string
    Registration Document
}
```

Identifica o favorecido de um pagamento: quem recebe. Validação: `Name` não pode ser vazio; `Registration` não pode ser `nil`. Erro: `*ValidationError{Context: "Payee", ...}`.

**Exemplo:**

```go
registration, err := cnab.NewCNPJ("11222333000181")
payee := cnab.Payee{Name: "FORNECEDOR X", Registration: registration}
```

---

## Documentos (CPF/CNPJ)

### `type Document`

```go
type Document interface {
    Digits() string // 11 (CPF) ou 14 (CNPJ) dígitos, sem pontuação
    Kind() string    // "CPF" ou "CNPJ"
}
```

Interface implementada por `CNPJ` e `CPF`. É o tipo aceito por `Company.Registration` e `Payee.Registration`.

### `type CNPJ` / `func NewCNPJ`

```go
type CNPJ string
func NewCNPJ(raw string) (CNPJ, error)
func (c CNPJ) Digits() string
func (c CNPJ) Kind() string // "CNPJ"
```

`NewCNPJ` remove pontuação de `raw` automaticamente e valida: precisa ter 14 dígitos, não pode ser uma sequência de 14 dígitos repetidos, e precisa passar no algoritmo padrão de dígito verificador módulo 11.

**Erros:** `*ValidationError{Context: "CNPJ", ...}` descrevendo exatamente o motivo (tamanho errado ou dígito verificador inválido).

**Exemplo:**

```go
cnpj, err := cnab.NewCNPJ("11.222.333/0001-81")
if err != nil {
    // cnpj inválido
}
```

### `type CPF` / `func NewCPF`

```go
type CPF string
func NewCPF(raw string) (CPF, error)
func (c CPF) Digits() string
func (c CPF) Kind() string // "CPF"
```

Mesma validação de `NewCNPJ`, para CPF (11 dígitos).

---

## Dinheiro

### `type Cents`

```go
type Cents int64
```

Valor monetário como inteiro na unidade mínima da moeda (R$ 252,00 é `Cents(25200)`).

---

## Arquivo e lote

### `type File` / `func NewRemittance`

```go
type File struct{ /* campos privados */ }
func NewRemittance(cfg Config) (*File, error)
```

`NewRemittance` valida `cfg` e inicia um novo arquivo de remessa.

**Validações e erros:**

- `Company`/`Account` inválidos: `*ValidationError`.
- `NSA <= 0`: `*ValidationError{Context: "Config", Reason: "NSA must be greater than zero"}`.
- `Layout` vazio: `*ValidationError`.
- `Layout` não registrado: `*ValidationError` listando os layouts disponíveis.

**Exemplo:** veja a seção "Exemplo mínimo" no `README.md`.

### `func (*File) NewBatch`

```go
func (f *File) NewBatch(product BatchProduct, service BatchService) (*Batch, error)
```

Inicia um novo lote para `product` liquidado via `service`, e o adiciona ao arquivo.

**Erros:**

- `*ValidationError` se `product` ou `service` for o valor zero (use sempre uma das constantes predefinidas, como `cnab.SupplierPayment`/`cnab.PixTransfer`).
- `*LimitExceededError{Limit: "batches_per_file", ...}` quando o arquivo já tem 70 lotes.

### `func (*File) Generate`

```go
func (f *File) Generate() ([]byte, error)
```

Renderiza o conteúdo completo do arquivo CNAB 240: header de arquivo, cada lote (header, pagamentos e trailer) e o trailer de arquivo, cada linha terminada em CRLF.

Antes de gerar, `Generate` revalida que nenhuma data de pagamento se tornou retroativa desde que foi adicionada (o tempo passa entre `AddPayment` e `Generate`). Depois que o motor renderiza o arquivo, `Generate` confere que a contagem total de registros bate com o que os lotes reportam, retornando `*TrailerMismatchError` em caso de divergência; isso não deveria acontecer em uso normal e existe como proteção contra um layout defeituoso. Os demais limites e regras estruturais (sequenciais, totais de trailer por lote) são garantidos por construção dentro do motor.

**Erros:** `*ValidationError` se o arquivo não tiver nenhum lote; `*ValidationError` se alguma data de pagamento estiver no passado; qualquer erro do motor (limite excedido, campo inválido), já traduzido com contexto de lote/registro/campo; `*TrailerMismatchError` (defensivo).

### `func (*File) FileName`

```go
func (f *File) FileName() (string, error)
```

Sugere um nome de arquivo, combinando o nome do layout ativo, o NSA configurado e a data atual (formato `LAYOUT_NNNN_AAAAMMDD.REM`). A maioria dos bancos só exige uma extensão específica (geralmente `.REM`); renomeie o resultado se seu banco exigir outra convenção.

**Erros:** `*ValidationError` se o arquivo não tiver nenhum lote ainda.

### `type Batch` / `func (*Batch) AddPayment`

```go
type Batch struct{ /* campos privados */ }
func (b *Batch) AddPayment(p Payment) error
```

Valida `p` e o adiciona ao lote como um novo movimento. A validação cobre todos os campos obrigatórios do tipo específico de `p` (veja a seção "Tipos de pagamento") e confere que o layout ativo do lote realmente suporta todos os segmentos que `p` precisa; um tipo de pagamento incompatível com o layout escolhido é rejeitado imediatamente, em vez de só falhar depois em `Generate`.

**Erros:**

- `*ValidationError` para um campo obrigatório ausente/inválido, ou para um `p` que seja `nil`.
- `*ValidationError` se o layout ativo não suportar um segmento necessário.
- `*LimitExceededError{Limit: "movements_per_batch", ...}` quando o lote já tem 10.000 movimentos.

---

## Produtos e serviços de lote

### `type BatchProduct`

Identifica o tipo de serviço de um lote (a "modalidade" FEBRABAN, sem campos exportados; use sempre uma constante predefinida):

```go
var SupplierPayment = BatchProduct{...} // Pagamento a Fornecedores
var PayrollPayment  = BatchProduct{...} // Pagamento de Salários
```

`func (p BatchProduct) String() string` retorna um nome descritivo.

### `type BatchService`

Identifica a forma de lançamento de um lote (sem campos exportados; use sempre uma constante predefinida):

```go
var CreditInAccount        = BatchService{...} // crédito em conta, mesmo banco
var TEDTransfer            = BatchService{...} // TED
var PixTransfer            = BatchService{...} // PIX (por chave ou por dados bancários)
var BoletoService          = BatchService{...} // pagamento de boletos
var BarcodeTaxService      = BatchService{...} // conta/tributo com código de barras
var TaxWithoutBarcodeService = BatchService{...} // DARF/GPS sem código de barras
```

`func (s BatchService) String() string` retorna um nome descritivo.

---

## Tipos de pagamento

Todos implementam a interface `Payment` (métodos não exportados, interface fechada: só os tipos deste pacote a implementam). Nenhum é instanciado por `New...`; são structs literais mesmo, com validação acontecendo em `Batch.AddPayment`.

### `type CreditAccount`

```go
type CreditAccount struct {
    Payee      Payee
    Account    Account // conta do favorecido, no mesmo banco do arquivo
    Amount     Cents
    Date       time.Time
    YourNumber string // opcional, "seu número"
}
```

Crédito em conta corrente no mesmo banco (Segmentos A e B). **Validação:** `Payee` e `Account` válidos, `Amount > 0`, `Date` presente e não retroativa.

### `type TED`

```go
type TED struct {
    Payee      Payee
    BankCode   string // código COMPE do banco do favorecido
    Account    Account
    Amount     Cents
    Date       time.Time
    Purpose    PurposeCode
    YourNumber string
}
```

TED (Segmentos A e B, com finalidade). **Validação:** igual a `CreditAccount`, mais `BankCode` e `Purpose` obrigatórios.

#### `type PurposeCode` (finalidade de TED)

```go
const (
    PurposeSupplierPayment PurposeCode = "00005"
    PurposePayroll         PurposeCode = "00006"
    PurposeOwnTransfer     PurposeCode = "00001"
    PurposeOther           PurposeCode = "00009"
)
```

Confirme a tabela exata de códigos com o seu banco antes de depender de um código específico em produção; os valores acima cobrem os casos mais comuns.

### `type BoletoPayment`

```go
type BoletoPayment struct {
    Barcode        string // código de barras de 44 dígitos
    Assignor       Payee  // cedente
    Payer          Payee  // sacado (normalmente a própria empresa)
    DueDate        time.Time
    DocumentAmount Cents // valor de face; se zero, usa Amount
    Discount       Cents // opcional
    Addition       Cents // opcional (juros/multa)
    Amount         Cents // valor efetivamente pago
    Date           time.Time
    YourNumber     string
}
```

Pagamento de boleto (Segmentos J e J-52). **Validação:** `Barcode` com 44 dígitos, `Assignor` e `Payer` válidos, `Amount > 0`, `Date` não retroativa.

### `type BarcodeTax`

```go
type BarcodeTax struct {
    Barcode    string // 44 dígitos
    DueDate    time.Time
    Amount     Cents
    Date       time.Time
    YourNumber string
}
```

Pagamento de conta/tributo com código de barras, como conta de energia (Segmento O). **Validação:** `Barcode` com 44 dígitos, `Amount > 0`, `Date` não retroativa.

### `type DARF`

```go
type DARF struct {
    TaxCode         string // código da receita
    Taxpayer        Payee
    ReferenceNumber string
    Period          time.Time // período de apuração
    DueDate         time.Time
    Principal       Cents
    Fine            Cents // opcional
    Interest        Cents // opcional
    Date            time.Time
    YourNumber      string
}
```

DARF normal, com principal, multa e juros controlados separadamente (Segmento N). **Validação:** `TaxCode` obrigatório, `Taxpayer` válido, `Principal + Fine + Interest > 0`, `Date` não retroativa.

### `type DARFSimple`

```go
type DARFSimple struct {
    TaxCode    string
    Taxpayer   Payee
    DueDate    time.Time
    Amount     Cents // valor total único
    Date       time.Time
    YourNumber string
}
```

DARF simplificado, com um valor total único em vez de principal/multa/juros separados (Segmento N, variante "DARF Simples"). **Validação:** `TaxCode` obrigatório, `Taxpayer` válido, `Amount > 0`, `Date` não retroativa.

### `type GPS`

```go
type GPS struct {
    Taxpayer   Payee
    Period     time.Time // competência
    DueDate    time.Time
    Amount     Cents
    Date       time.Time
    YourNumber string
}
```

Guia da Previdência Social (Segmento N, variante "GPS"). **Validação:** `Taxpayer` válido, `Amount > 0`, `Date` não retroativa.

### `type CancelPayment`

```go
type CancelPayment struct {
    Original Payment
}
```

Cancela um pagamento enviado em uma remessa anterior. Reenvia os mesmos dados do pagamento original, com o tipo de movimento FEBRABAN definido como `"9"` (exclusão) e o código de instrução como `"99"`, automaticamente.

**Validação:** `Original` não pode ser `nil`; os demais campos obrigatórios de `Original` ainda são validados, exceto a regra de data retroativa, que é intencionalmente ignorada aqui (um cancelamento legitimamente referencia uma data de pagamento já passada).

**Exemplo:**

```go
original := cnab.CreditAccount{
    Payee:   payee,
    Account: account,
    Amount:  cnab.Cents(25200),
    Date:    time.Now().AddDate(0, 0, -3), // já enviado anteriormente
}
err := batch.AddPayment(cnab.CancelPayment{Original: original})
```

---

## PIX

### `type Pix`

```go
type Pix struct {
    Key        PixKey
    Payee      Payee
    Amount     Cents
    Date       time.Time
    YourNumber string
}
```

Transferência PIX endereçada por chave (Segmentos A e B). **Validação:** `Key` não pode ser `nil` nem ter valor vazio, `Payee` válido, `Amount > 0`, `Date` não retroativa.

### `type PixKey` e as cinco variantes

```go
type PixKey interface { /* métodos não exportados */ }

type PhoneKey  string // ex.: "+5551998765432"
type EmailKey  string // ex.: "fornecedor@exemplo.com"
type CPFKey    string
type CNPJKey   string
type RandomKey string // "chave aleatória" (UUID)
```

`PixKey` é uma interface fechada: só esses cinco tipos a implementam. Cada um é só o valor da chave, convertido para o tipo correspondente.

**Exemplo:**

```go
cnab.Pix{Key: cnab.EmailKey("fornecedor@exemplo.com"), ...}
cnab.Pix{Key: cnab.PhoneKey("+5551998765432"), ...}
cnab.Pix{Key: cnab.RandomKey("98798987-2398-4732-8743-824732984792"), ...}
```

### `type PixKeyType`

```go
type PixKeyType string
const (
    PixKeyTypePhone  PixKeyType = "phone"
    PixKeyTypeEmail  PixKeyType = "email"
    PixKeyTypeCPF    PixKeyType = "cpf"
    PixKeyTypeCNPJ   PixKeyType = "cnpj"
    PixKeyTypeRandom PixKeyType = "random"
)
```

Identifica qual variante de `PixKey` está em uso. É o tipo de retorno do método interno usado para escolher o código FEBRABAN de 2 dígitos gravado no arquivo (`"01"` telefone, `"02"` e-mail, `"03"` CPF/CNPJ, `"04"` chave aleatória).

### `type PixBankData`

```go
type PixBankData struct {
    Payee      Payee
    BankCode   string // COMPE ou ISPB do banco do favorecido
    Account    Account
    Amount     Cents
    Date       time.Time
    YourNumber string
}
```

Transferência PIX endereçada pelos dados bancários do favorecido em vez de uma chave (Segmentos A e B). **Validação:** `Payee` e `Account` válidos, `BankCode` obrigatório, `Amount > 0`, `Date` não retroativa.

---

## Registro de layouts

### `func RegisterLayout`

```go
func RegisterLayout(name string, l Layout)
```

Torna `l` disponível sob o nome `name` para uso posterior em `Config.Layout`. Pensada para ser chamada uma única vez, a partir da função `init()` de um pacote de banco, no mesmo padrão de registro de drivers do `database/sql`.

**Entra em pânico** (não retorna erro) se `l` for `nil`, se `name` for vazio, se `l.Version()` for vazio, ou se já existir um layout registrado com o mesmo nome: são todos erros de programação detectados na inicialização, não condições de tempo de execução para tratar com `recover`.

### `type Layout`

```go
type Layout = layout.Layout
```

Alias de `layout.Layout` (veja a seção sobre o pacote `cnab/layout`), para que o contrato também possa ser referenciado como `cnab.Layout`.

O layout `"febraban240"` já vem registrado automaticamente: basta importar o pacote `cnab` (o próprio pacote importa `layouts/febraban240` internamente), sem nenhum import adicional.

---

## Erros

Todos os erros abaixo são ponteiros para struct e implementam `error`; use `errors.As` para identificá-los especificamente.

### `type ValidationError`

```go
type ValidationError struct {
    Context string
    Reason  string
}
```

Uma regra de negócio falhou (campo obrigatório ausente, valor inválido, sequência de chamadas da API incorreta). `Context` identifica onde (por exemplo, `"Company"`, `"Pix"`, `"AddPayment"`); `Reason` descreve o motivo.

### `type FieldError`

```go
type FieldError struct {
    Batch  int
    Record string
    Field  string
    Reason string
}
```

Um campo específico, de um registro específico, dentro de um lote específico, não pôde ser renderizado.

### `type LimitExceededError`

```go
type LimitExceededError struct {
    Limit     string // "batches_per_file" ou "movements_per_batch"
    Max       int
    Attempted int
    Batch     int // 0 quando o limite é de arquivo, não de lote
}
```

Um limite estrutural do FEBRABAN (no máximo 70 lotes por arquivo, no máximo 10.000 movimentos por lote) foi excedido.

### `type SequenceError`

```go
type SequenceError struct {
    Context  string
    Expected int
    Got      int
}
```

Um número de sequência de registro ficou fora de ordem. Como só o motor calcula sequenciais, isso só deveria ser observado se um `Layout` estiver malformado.

### `type TrailerMismatchError`

```go
type TrailerMismatchError struct {
    Batch    int // 0 para o trailer de arquivo
    Field    string
    Expected string
    Got      string
}
```

Um total de trailer calculado não bateu com os registros que ele deveria resumir. Como só o motor calcula esses totais, isso só deveria ser observado se um `Layout` estiver malformado; `Generate` roda essa conferência como proteção defensiva, não como validação de uso normal.

---

## Pacote `cnab/layout`

Voltado a quem vai **escrever** o descritor de um banco, não a quem só usa o SDK. Veja `NOVO-BANCO.md` para o guia de uso passo a passo.

### `type FieldKind` e constantes

```go
type FieldKind int
const (
    KindNumeric      FieldKind = iota // "9": alinhado à direita, zero à esquerda
    KindAlphanumeric                   // "X": alinhado à esquerda, espaço à direita
)
func (k FieldKind) String() string
```

### `type FieldSpec`

```go
type FieldSpec struct {
    Name     string
    Start    int // coluna inicial, 1-based, inclusive
    End      int // coluna final, 1-based, inclusive
    Kind     FieldKind
    Decimals int    // casas decimais implícitas, notação 9(n)V(d)
    Key      Key    // chave semântica; vazio se Const estiver definido
    Const    string // valor fixo; vazio se Key estiver definido
}
func (f FieldSpec) Size() int     // End - Start + 1
func (f FieldSpec) IsConst() bool // true se o campo sempre renderiza um literal fixo
```

Descreve uma faixa de colunas de um registro de 240 caracteres. Exatamente um entre `Key` e `Const` deve estar definido.

### `type RecordSpec`

```go
type RecordSpec struct {
    Name   string
    Fields []FieldSpec
}
```

Descreve as 240 colunas completas de um tipo de registro ou segmento. Os campos precisam cobrir as colunas 1 a 240 sem sobreposição nem lacuna; isso é validado uma vez, quando um `Layout` é usado para construir um `internal/engine.Engine`.

```go
func (r RecordSpec) Validate() error
```

Roda antecipadamente a mesma verificação de cobertura (1 a 240, sem lacuna nem sobreposição) que o motor roda ao construir um `Engine`. Útil para validar um `RecordSpec` no momento em que ele é montado, antes de registrar o layout; é o que `NewFromJSON` usa para validar cada registro do arquivo de configuração já no carregamento.

### `type RecordKey` e constantes

```go
type RecordKey string
const (
    FileHeader     RecordKey = "file_header"
    FileTrailer    RecordKey = "file_trailer"
    BatchHeader    RecordKey = "batch_header"
    BatchTrailer   RecordKey = "batch_trailer"
    SegmentA       RecordKey = "segment_a"
    SegmentB       RecordKey = "segment_b"       // crédito em conta / TED / PIX por dados bancários
    SegmentBPix    RecordKey = "segment_b_pix"   // PIX por chave
    SegmentJ       RecordKey = "segment_j"
    SegmentJ52     RecordKey = "segment_j52"
    SegmentO       RecordKey = "segment_o"
    SegmentN       RecordKey = "segment_n"        // DARF normal
    SegmentNSimple RecordKey = "segment_n_simple" // DARF Simples
    SegmentNSocial RecordKey = "segment_n_social" // GPS
)
```

Identifica o papel de uma linha de 240 caracteres dentro de um arquivo CNAB 240.

```go
var AllRecordKeys []RecordKey            // todos os valores acima, nesta ordem
func IsKnownRecordKey(key RecordKey) bool // true se key está em AllRecordKeys
```

### `type Layout`

```go
type Layout interface {
    Name() string
    Version() string
    Record(key RecordKey) (spec RecordSpec, ok bool)
}
```

Contrato que um layout de banco/produto precisa implementar. `Version()` não pode retornar vazio (erro detectado no registro). `Record` retorna `ok=false` para qualquer `RecordKey` que o layout não suporte.

### `func Register` / `func Lookup` / `func Names`

```go
func Register(name string, l Layout) // entra em pânico em nil/nome vazio/versão vazia/duplicidade
func Lookup(name string) (Layout, bool)
func Names() []string // nomes registrados, ordenados
```

`Register` é a função de baixo nível por trás de `cnab.RegisterLayout` (mesmo registro interno). `Names` é útil para mensagens de erro que sugerem alternativas válidas.

### `func NewFromJSON` / `func NewFromJSONFile`

```go
func NewFromJSON(data []byte) (Layout, error)
func NewFromJSONFile(path string) (Layout, error)
```

Constroem um `Layout` a partir de uma descrição em JSON (o mesmo `RecordSpec`/`FieldSpec` dos passos manuais, só que carregados de um arquivo em vez de escritos em Go), sem registrá-lo automaticamente. `NewFromJSONFile` só lê `path` e chama `NewFromJSON`. Nenhuma dependência externa é usada (`encoding/json` é biblioteca padrão).

Validação, toda feita no carregamento (não só quando o layout é usado depois):

- `name` e `version` são obrigatórios no JSON.
- toda chave do objeto `records` precisa ser um dos valores de `AllRecordKeys` (`"file_header"`, `"segment_a"`, etc.).
- todo campo precisa de `kind` igual a `"9"`/`"numeric"` ou `"X"`/`"alphanumeric"`, e de uma faixa de colunas válida (`1 <= start <= end <= 240`).
- um campo nunca pode ter `key` e `const` definidos ao mesmo tempo.
- todo `key` de campo, quando presente, precisa estar em `AllKeys`.
- cada registro precisa passar em `RecordSpec.Validate()` (cobertura de 1 a 240 sem lacuna nem sobreposição).

Toda falha de validação retorna um erro descrevendo o registro e, quando aplicável, o número do campo dentro dele. Veja `NOVO-BANCO.md`, seção "Alternativa: descrevendo o layout em JSON em vez de Go", para o formato completo do arquivo e um exemplo de uso ponta a ponta com `cnab.RegisterLayout`.

### `type Key` e o vocabulário de chaves semânticas/estruturais

```go
type Key string
type Values map[Key]any
```

`Values` é a única estrutura de dados que atravessa a fronteira entre a API de domínio (`cnab`) e um `Layout`. As constantes `Key` estão documentadas em `cnab/layout/value.go` e resumidas em `ARQUITETURA.md`; as mais usadas ao escrever um layout novo são:

**Estruturais** (nunca definidas fora de `internal/engine`): `KeyBatchNumber`, `KeySequence`, `KeyBatchRecordCount`, `KeyBatchAmount`, `KeyBatchCount`, `KeyFileRecordCount`.

**Semânticas** (definidas pela camada `cnab`, lidas pelos `FieldSpec.Key` de um layout): `KeyFileSequenceNumber`, `KeyFileGenerationDate`, `KeyFileGenerationTime`, `KeyCompanyRegistrationKind`, `KeyCompanyRegistration`, `KeyCompanyName`, `KeyAgreement`, `KeyBranch`, `KeyAccountNumber`, `KeyAccountCheckDigit`, `KeyBatchProductCode`, `KeyBatchServiceCode`, `KeyMovementType`, `KeyInstructionCode`, `KeyClearingCode`, `KeyBeneficiaryBankCode`, `KeyBeneficiaryBranch`, `KeyBeneficiaryAccount`, `KeyBeneficiaryCheckDigit`, `KeyPayeeName`, `KeyPayeeDocumentKind`, `KeyPayeeDocument`, `KeyYourNumber`, `KeyAmount`, `KeyPaymentDate`, `KeyPurposeCode`, `KeyPixKeyType`, `KeyPixKeyValue`, `KeyPayeeAddressStreet`/`Number`/`District`/`City`/`State`/`ZipCode`, `KeyBarcode`, `KeyDueDate`, `KeyDocumentAmount`, `KeyDiscountAmount`, `KeyAdditionAmount`, `KeyPayerDocumentKind`, `KeyPayerDocument`, `KeyPayerName`, `KeyAssignorDocumentKind`, `KeyAssignorDocument`, `KeyAssignorName`, `KeyTaxCode`, `KeyTaxpayerDocumentKind`, `KeyTaxpayerIdType`, `KeyTaxpayerDocument`, `KeyTaxpayerName`, `KeyReferenceNumber`, `KeyPeriod`, `KeyPrincipalAmount`, `KeyFineAmount`, `KeyInterestAmount`.

```go
var AllKeys []Key          // toda constante Key acima, estrutural e semântica
func IsKnownKey(k Key) bool // true se k está em AllKeys
```

`IsKnownKey` é o que `NewFromJSON` usa para rejeitar, no carregamento, um `"key"` de campo digitado errado no arquivo de configuração.
