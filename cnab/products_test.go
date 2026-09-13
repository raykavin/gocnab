package cnab

import (
	"testing"
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
)

// The codes below are FEBRABAN's, as published in the bank manuals that use
// them; these tests exist because the two families are easy to confuse. A
// "tipo de serviço" and a "forma de lançamento" sit side by side in columns
// 10-13 of a lote header, are both two digits, and in one case share the
// same number for different meanings: 22 is the tax *product*, while the
// barcoded-tax *service* is 11.
//
// Repeating a product code into a service code produced a pair no bank
// manual lists, and Sicredi refused the whole remittance with "tipo de
// arquivo inválido" — a file-level critique with no line and no position,
// because the bank stops before parsing anything. That is expensive to
// diagnose from the outside, so the codes are pinned here at the source.

func TestBatchProductCodes(t *testing.T) {
	cases := []struct {
		name    string
		product BatchProduct
		want    string
	}{
		{"boleto eletrônico", BoletoCollection, "03"},
		{"pagamento fornecedor", SupplierPayment, "20"},
		{"pagamento de contas, tributos e impostos", TaxPayment, "22"},
		{"pagamento de salários", PayrollPayment, "30"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.product.code != c.want {
				t.Errorf("%s code = %q, want %q", c.product, c.product.code, c.want)
			}
		})
	}
}

func TestBatchServiceCodes(t *testing.T) {
	cases := []struct {
		name    string
		service BatchService
		want    string
	}{
		{"crédito em conta corrente", CreditInAccount, "01"},
		// 11, not 22: 22 is TaxPayment's product code, and a lote that
		// repeats it here announces a pair that addresses no segment.
		{"pagamento de contas e tributos com código de barras", BarcodeTaxService, "11"},
		{"tributo - DARF normal", DARFService, "16"},
		{"tributo - GPS", GPSService, "17"},
		{"tributo - DARF simples", DARFSimpleService, "18"},
		{"liquidação de títulos do próprio banco", BoletoService, "30"},
		{"pagamento de títulos de outros bancos", OtherBankBoletoService, "31"},
		{"TED - transferência entre clientes", TEDTransfer, "41"},
		{"PIX transferência", PixTransfer, "45"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.service.code != c.want {
				t.Errorf("%s code = %q, want %q", c.service, c.service.code, c.want)
			}
		})
	}
}

// TestBatchServiceCodesAreDistinct guards the mistake directly: no two
// services may share a code, since the code is the only thing the bank
// reads to tell one lote's settlement form from another's.
func TestBatchServiceCodesAreDistinct(t *testing.T) {
	all := []BatchService{
		CreditInAccount, BarcodeTaxService,
		DARFService, GPSService, DARFSimpleService,
		BoletoService, OtherBankBoletoService, TEDTransfer, PixTransfer,
	}
	seen := make(map[string]BatchService, len(all))
	for _, s := range all {
		if other, taken := seen[s.code]; taken {
			t.Errorf("services %s and %s share the code %q", other, s, s.code)
			continue
		}
		seen[s.code] = s
	}
}

// TestTaxServicesMatchTheirSegment pairs each barcodeless tax service with
// the payment type it settles. The two travel together: the service code in
// the lote header is how the bank knows which Segmento N variant to expect,
// and the payment type is what actually writes it. A pair that disagrees
// produces a lote the bank reads with the wrong structure.
//
// This is the check that a single "tax without barcode" service could not
// support — one code cannot match three segments.
func TestTaxServicesMatchTheirSegment(t *testing.T) {
	now := time.Now()

	cases := []struct {
		name        string
		service     BatchService
		wantCode    string
		payment     Payment
		wantSegment layout.RecordKey
	}{
		{
			name:    "DARF normal",
			service: DARFService, wantCode: "16",
			payment: DARF{
				TaxCode: "0220", Taxpayer: validPayee(), ReferenceNumber: "12345",
				Period: now, DueDate: now.AddDate(0, 0, 5),
				Principal: 1000, Fine: 100, Interest: 50, Date: now.AddDate(0, 0, 1),
			},
			wantSegment: layout.SegmentN,
		},
		{
			name:    "GPS",
			service: GPSService, wantCode: "17",
			payment: GPS{
				Taxpayer: validPayee(), Period: now, DueDate: now.AddDate(0, 0, 5),
				Amount: 1150, Date: now.AddDate(0, 0, 1),
			},
			wantSegment: layout.SegmentNSocial,
		},
		{
			name:    "DARF simples",
			service: DARFSimpleService, wantCode: "18",
			payment: DARFSimple{
				TaxCode: "6106", Taxpayer: validPayee(), DueDate: now.AddDate(0, 0, 5),
				Amount: 1150, Date: now.AddDate(0, 0, 1),
			},
			wantSegment: layout.SegmentNSimple,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.service.code != c.wantCode {
				t.Errorf("service code = %q, want %q", c.service.code, c.wantCode)
			}
			segments, err := c.payment.toSegments(nil)
			if err != nil {
				t.Fatalf("toSegments() error = %v", err)
			}
			if len(segments) != 1 {
				t.Fatalf("toSegments() returned %d segments, want 1", len(segments))
			}
			if segments[0].Key != c.wantSegment {
				t.Errorf("segment = %q, want %q", segments[0].Key, c.wantSegment)
			}
		})
	}
}
