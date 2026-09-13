# GoCNAB

SDK em Go para geração de arquivos de remessa e leitura de arquivos de retorno no padrão **CNAB 240 FEBRABAN**, com arquitetura multi-banco.

[![Go Reference](https://pkg.go.dev/badge/github.com/raykavin/gocnab.svg)](https://pkg.go.dev/github.com/raykavin/gocnab)
[![Go Version](https://img.shields.io/badge/go-1.26+-00ADD8?logo=go&logoColor=white)](https://golang.org/dl/)
[![Go Report Card](https://goreportcard.com/badge/github.com/raykavin/gocnab)](https://goreportcard.com/report/github.com/raykavin/gocnab)
[![Zero Dependencies](https://img.shields.io/badge/dependencies-none-brightgreen)](go.mod)
[![Release](https://img.shields.io/github/v/release/raykavin/gocnab?logo=github)](https://github.com/raykavin/gocnab/releases)
[![Last Commit](https://img.shields.io/github/last-commit/raykavin/gocnab)](https://github.com/raykavin/gocnab/commits)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

O SDK separa três camadas independentes:

1. **Motor genérico** (`internal/engine`): sabe preencher campos, montar registros de 240 colunas, calcular sequenciais e trailers, e aplicar os limites do padrão FEBRABAN (70 lotes por arquivo, 10.000 movimentos por lote). Não conhece nenhum banco específico.
2. **Descritores de layout** (`cnab/layout` e pacotes `layouts/<nome>`): descrevem, campo a campo, o layout de um banco/produto. O layout de referência `febraban240` (padrão FEBRABAN puro, sem personalização de banco) já vem embutido.
3. **API pública de domínio** (`cnab`): o que o desenvolvedor realmente usa. Nenhuma posição de campo aparece aqui, apenas conceitos como empresa, conta, favorecido e tipos de pagamento (crédito em conta, TED, PIX, boleto, tributos, cancelamento).

Veja [ARQUITETURA.md](ARQUITETURA.md) para o detalhamento das três camadas e [NOVO-BANCO.md](NOVO-BANCO.md) para o passo a passo de como derivar o layout de um banco real a partir de `febraban240`. A referência completa de tipos, funções e constantes exportados está em [API.md](API.md).

## Funcionalidades

- **Remessa CNAB 240** com múltiplos lotes, sequenciais e trailers calculados pelo próprio motor.
- **Tipos de pagamento**: crédito em conta, TED, PIX por chave, PIX por dados bancários, boleto, conta/tributo com código de barras, DARF, DARF Simples, GPS e cancelamento de um pagamento já enviado.
- **Leitura de arquivo de retorno** (`ParseReturn`), com um movimento por Segmento A, J ou O e dados de autenticação do Segmento Z quando o layout do banco o implementa.
- **Conversão de linha digitável** de boleto (47 dígitos) ou de conta/tributo (48 dígitos) em código de barras de 44 dígitos, com conferência dos dígitos verificadores (`ConvertToBarcode`).
- **CNPJ alfanumérico** da Receita Federal aceito em qualquer campo de documento, além do formato histórico todo numérico.
- **Layouts plugáveis**: registrados por nome em `init()`, informados diretamente em tempo de execução (`Config.LayoutSpec`) ou carregados de um arquivo JSON (`layout.NewFromJSON`).
- **Valores monetários sempre inteiros** em centavos (`cnab.Cents`), nunca `float64`.
- **Erros tipados** (`ValidationError`, `FieldError`, `LimitExceededError`, `ReturnParseError`, entre outros), identificáveis com `errors.As`.
- **Zero dependências** além da biblioteca padrão.

## Requisitos

- Go 1.26 ou superior.
- Nenhuma dependência externa.

## Instalação

```bash
go get github.com/raykavin/gocnab
```

## Exemplo mínimo

```go
package main

import (
	"log"
	"time"

	"github.com/raykavin/gocnab/cnab"
)

func main() {
	registration, err := cnab.NewCNPJ("11222333000181")
	if err != nil {
		log.Fatal(err)
	}

	file, err := cnab.NewRemittance(cnab.Config{
		Layout: "febraban240",
		Company: cnab.Company{
			Name:         "ACME LTDA",
			Registration: registration,
			Agreement:    "1234",
		},
		Account: cnab.Account{Branch: "0116", Number: "75890", CheckDigit: "6"},
		NSA:     1,
	})
	if err != nil {
		log.Fatal(err)
	}

	batch, err := file.NewBatch(cnab.SupplierPayment, cnab.PixTransfer)
	if err != nil {
		log.Fatal(err)
	}

	payeeRegistration, _ := cnab.NewCNPJ("11444777000161")
	err = batch.AddPayment(cnab.Pix{
		Key:    cnab.EmailKey("fornecedor@exemplo.com"),
		Payee:  cnab.Payee{Name: "COLABORADOR X", Registration: payeeRegistration},
		Amount: cnab.Cents(25200), // R$ 252,00
		Date:   time.Now().AddDate(0, 0, 1),
	})
	if err != nil {
		log.Fatal(err)
	}

	content, err := file.Generate()
	if err != nil {
		log.Fatal(err)
	}

	name, _ := file.FileName()
	log.Printf("gerado %s com %d bytes", name, len(content))
}
```

Valores monetários são sempre inteiros em centavos (`cnab.Cents`), nunca `float64`. Datas usam `time.Time`. Erros são tipados (`cnab.ValidationError`, `cnab.LimitExceededError`, `cnab.FieldError`, entre outros) e descritivos.

O par produto/serviço do lote (`cnab.SupplierPayment`, `cnab.PixTransfer`, ...) depende do manual do seu banco: alguns exigem produtos próprios para boleto (`cnab.BoletoCollection`) e para tributos (`cnab.TaxPayment`) em vez de aninhá-los em `cnab.SupplierPayment`.

## Código de barras e linha digitável

`BoletoPayment` e `BarcodeTax` recebem o código de barras de 44 dígitos. Quando a entrada é a linha digitável, `ConvertToBarcode` normaliza os dois formatos e ainda diz para qual dos dois tipos de pagamento o resultado serve:

```go
segment, barcode, err := cnab.ConvertToBarcode("34190.10438 51004.791029 01500.080005 1 91820000023000")
if err != nil {
	log.Fatal(err) // dígito verificador não confere, tamanho inválido, caractere inesperado
}

switch segment {
case cnab.SegmentBankSlip:
	err = batch.AddPayment(cnab.BoletoPayment{Barcode: barcode /* ... */})
case cnab.SegmentFeesOrTaxes:
	err = batch.AddPayment(cnab.BarcodeTax{Barcode: barcode /* ... */})
}
```

Todos os dígitos verificadores FEBRABAN são conferidos antes do retorno, então uma linha digitada errada é recusada aqui em vez de virar um código de barras plausível mas errado.

## Processando retorno

```go
content, err := os.ReadFile("retorno_20260105.ret")
if err != nil {
	log.Fatal(err)
}

result, err := cnab.ParseReturn("febraban240", content)
if err != nil {
	log.Fatal(err)
}

for _, m := range result.Movements {
	if m.Accepted() {
		fmt.Printf("%s: liquidado em %s\n", m.YourNumber, m.SettlementDate.Format("2006-01-02"))
	} else {
		fmt.Printf("%s: rejeitado, códigos %v\n", m.YourNumber, m.OccurrenceCodes)
	}
}
```

`ParseReturn` gera um movimento por segmento principal encontrado: Segmento A (crédito em conta, TED, PIX), Segmento J (boleto) e Segmento O (conta/tributo com código de barras). De cada um extrai o "seu número" enviado na remessa, o valor instruído, a data e o valor reais de liquidação e os códigos de ocorrência/rejeição, o suficiente para reconciliar um pagamento. Um Segmento Z posterior preenche os dados de autenticação do movimento, quando o layout do banco o implementa. Os segmentos complementares (B, BPix, J-52) e o Segmento N são ignorados; veja "Processando retorno" em [ARQUITETURA.md](ARQUITETURA.md) para o motivo.

A tabela de códigos de ocorrência é específica de cada banco: confirme com o manual dele antes de interpretar um código.

## Layouts de banco

O layout `febraban240` é o padrão FEBRABAN puro e já vem registrado; ele não representa nenhum banco real (o código de compensação é um `"000"` de referência). Há três formas de usar o layout de um banco:

- **Registro por nome**, o caminho usual para um layout que é constante do binário: o pacote do banco chama `layout.Register` no seu `init()` e o chamador informa o nome em `Config.Layout`.
- **Instância direta**, para um layout que é dado resolvido em tempo de execução: preencha `Config.LayoutSpec` (e use `ParseReturnWithLayout` na leitura de retorno).
- **Descritor JSON**, para trocar o layout sem recompilar: `layout.NewFromJSON`/`layout.NewFromJSONFile`, com validação completa já no carregamento.

O passo a passo para derivar o layout de um banco real a partir do manual dele está em [NOVO-BANCO.md](NOVO-BANCO.md).

## Exemplos completos

A pasta `./examples` tem exemplos para cada cenário coberto pelo SDK:

| Pasta                         | Cenário                                                                                      |
| ----------------------------- | -------------------------------------------------------------------------------------------- |
| `examples/credit_account`     | Crédito em conta corrente                                                                    |
| `examples/ted`                | TED                                                                                          |
| `examples/pix_key`            | PIX por chave                                                                                |
| `examples/pix_bank_data`      | PIX por dados bancários                                                                      |
| `examples/boleto`             | Pagamento de boleto                                                                          |
| `examples/barcode_tax`        | Tributo/conta com código de barras                                                           |
| `examples/darf`               | DARF                                                                                         |
| `examples/gps`                | GPS                                                                                          |
| `examples/cancel_payment`     | Cancelamento de pagamento                                                                    |
| `examples/custom_layout_json` | Layout de banco carregado de um arquivo JSON (`layout.NewFromJSON`), em vez de escrito em Go |
| `examples/parse_return`       | Leitura de um arquivo de retorno (`cnab.ParseReturn`)                                        |

Cada exemplo roda isoladamente, por exemplo:

```bash
go run ./examples/pix_key
```

## Estrutura do projeto

```
cnab/              API pública de domínio: Config, File, Batch, tipos de pagamento, retorno
cnab/layout/       contrato de layout: Layout, FieldSpec, RecordSpec, vocabulário de Key, registro, carregador JSON
internal/engine/   motor genérico: campos, registros de 240 colunas, sequenciais, trailers, limites
layouts/febraban240/  layout de referência do padrão FEBRABAN puro
examples/          programas executáveis, um por cenário
```

## Testes

```bash
go test ./... -cover
```

O pacote `internal/engine` (o motor genérico) mantém cobertura de testes acima de 85%.

## Comandos úteis

```bash
go build ./...         # compila todos os pacotes
go vet ./...           # análise estática
go run ./examples/ted  # executa um exemplo
```

## Documentação

- [ARQUITETURA.md](ARQUITETURA.md): as três camadas do SDK e as decisões de design.
- [API.md](API.md): referência completa da API pública.
- [NOVO-BANCO.md](NOVO-BANCO.md): passo a passo para implementar o descritor de um banco real a partir do manual CNAB dele.

---

## Contribuindo

Contribuições para o gocnab são bem-vindas! Veja algumas formas de ajudar:

- **Reporte bugs e sugira funcionalidades** abrindo issues no GitHub
- **Envie pull requests** com correções de bugs ou novas funcionalidades
- **Melhore a documentação** para ajudar outros usuários e desenvolvedores

---

## Licença

gocnab é distribuído sob a **Licença MIT**.
Para os termos e condições completos da licença, veja o arquivo [LICENSE](LICENSE) no repositório.
