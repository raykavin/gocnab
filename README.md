# GoCNAB

SDK em Go para geração de arquivos de remessa no padrão **CNAB 240 FEBRABAN**, com arquitetura multi-banco.

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

## Instalação

```bash
go get github.com/raykavin/gocnab
```

Requer Go 1.26 ou superior. Nenhuma dependência externa além da biblioteca padrão.

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
		Payee:  cnab.Payee{Name: "FORNECEDOR X", Registration: payeeRegistration},
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

## Exemplos completos

A pasta `./examples` tem um programa `go run`-ável para cada cenário coberto pelo SDK:

| Pasta | Cenário |
|---|---|
| `examples/credit_account` | Crédito em conta corrente |
| `examples/ted` | TED |
| `examples/pix_key` | PIX por chave |
| `examples/pix_bank_data` | PIX por dados bancários |
| `examples/boleto` | Pagamento de boleto |
| `examples/barcode_tax` | Tributo/conta com código de barras |
| `examples/darf` | DARF |
| `examples/gps` | GPS |
| `examples/cancel_payment` | Cancelamento de pagamento |
| `examples/custom_layout_json` | Layout de banco carregado de um arquivo JSON (`layout.NewFromJSON`), em vez de escrito em Go |

Cada exemplo roda isoladamente, por exemplo:

```bash
go run ./examples/pix_key
```

## Documentação

- [ARQUITETURA.md](ARQUITETURA.md): as três camadas do SDK e as decisões de design.
- [API.md](API.md): referência completa da API pública.
- [NOVO-BANCO.md](NOVO-BANCO.md): passo a passo para implementar o descritor de um banco real a partir do manual CNAB dele.

## Testes

```bash
go test ./... -cover
```

O pacote `internal/engine` (o motor genérico) mantém cobertura de testes acima de 85%.

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