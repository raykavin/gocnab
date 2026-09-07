package cnab

import (
	"testing"
	"time"

	"github.com/raykavin/gocnab/cnab/layout"
)

func TestPixValidate(t *testing.T) {
	now := time.Now()
	valid := Pix{
		Key:    EmailKey("fornecedor@exemplo.com"),
		Payee:  validPayee(),
		Amount: 1000,
		Date:   now.AddDate(0, 0, 1),
	}
	if err := valid.validate(now, validateOptions{}); err != nil {
		t.Fatalf("validate() error = %v, want nil", err)
	}

	cases := []struct {
		name    string
		payment Pix
	}{
		{"missing key", Pix{Payee: validPayee(), Amount: 1000, Date: now.AddDate(0, 0, 1)}},
		{"empty key value", Pix{Key: EmailKey(""), Payee: validPayee(), Amount: 1000, Date: now.AddDate(0, 0, 1)}},
		{"zero amount", Pix{Key: EmailKey("x@y.com"), Payee: validPayee(), Date: now.AddDate(0, 0, 1)}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.payment.validate(now, validateOptions{}); err == nil {
				t.Fatal("validate() error = nil, want an error")
			}
		})
	}
}

func TestPixKeyTypes(t *testing.T) {
	cases := []struct {
		key      PixKey
		wantType PixKeyType
		wantCode string
	}{
		{PhoneKey("+5551998765432"), PixKeyTypePhone, "01"},
		{EmailKey("a@b.com"), PixKeyTypeEmail, "02"},
		{CPFKey("11144477735"), PixKeyTypeCPF, "03"},
		{CNPJKey("11222333000181"), PixKeyTypeCNPJ, "03"},
		{RandomKey("98798987-2398-4732-8743-824732984792"), PixKeyTypeRandom, "04"},
	}
	for _, c := range cases {
		t.Run(string(c.wantType), func(t *testing.T) {
			if got := c.key.pixKeyType(); got != c.wantType {
				t.Fatalf("pixKeyType() = %q, want %q", got, c.wantType)
			}
			if got := pixKeyFebrabanCode(c.key.pixKeyType()); got != c.wantCode {
				t.Fatalf("pixKeyFebrabanCode() = %q, want %q", got, c.wantCode)
			}
		})
	}
}

func TestPixToSegments(t *testing.T) {
	now := time.Now()
	p := Pix{
		Key:    EmailKey("fornecedor@exemplo.com"),
		Payee:  validPayee(),
		Amount: 25200,
		Date:   now.AddDate(0, 0, 1),
	}

	segments, err := p.toSegments(nil)
	if err != nil {
		t.Fatalf("toSegments() error = %v", err)
	}
	if len(segments) != 2 {
		t.Fatalf("len(segments) = %d, want 2", len(segments))
	}
	if segments[1].Key != layout.SegmentBPix {
		t.Fatalf("segments[1].Key = %q, want SegmentBPix", segments[1].Key)
	}
	if segments[0].Values[layout.KeyClearingCode] != "009" {
		t.Fatalf("KeyClearingCode = %v, want \"009\" (PIX)", segments[0].Values[layout.KeyClearingCode])
	}
	if segments[1].Values[layout.KeyPixKeyValue] != "fornecedor@exemplo.com" {
		t.Fatalf("KeyPixKeyValue = %v", segments[1].Values[layout.KeyPixKeyValue])
	}
	if segments[1].Values[layout.KeyPixKeyType] != "02" {
		t.Fatalf("KeyPixKeyType = %v, want \"02\"", segments[1].Values[layout.KeyPixKeyType])
	}
}

func TestPixBankDataValidateAndSegments(t *testing.T) {
	now := time.Now()
	p := PixBankData{
		Payee:       validPayee(),
		BankCode:    "001",
		ISPB:        "01181521",
		AccountKind: PixBankAccountChecking,
		Account:     validAccount(),
		Amount:      500,
		Date:        now.AddDate(0, 0, 1),
	}
	if err := p.validate(now, validateOptions{}); err != nil {
		t.Fatalf("validate() error = %v, want nil", err)
	}

	segments, err := p.toSegments(nil)
	if err != nil {
		t.Fatalf("toSegments() error = %v", err)
	}
	// Both PIX submodalities (by key and by bank data) use the same
	// Segmento B-Pix, not the plain Segmento B credit/TED uses: Sicredi's
	// manual groups every PIX launch form under "Registro detalhe -
	// segmento B (Obrigatório para PIX)".
	if segments[0].Key != layout.SegmentA || segments[1].Key != layout.SegmentBPix {
		t.Fatalf("unexpected segment keys: %v, %v", segments[0].Key, segments[1].Key)
	}
	if segments[0].Values[layout.KeyClearingCode] != "009" {
		t.Fatalf("KeyClearingCode = %v, want \"009\" (PIX)", segments[0].Values[layout.KeyClearingCode])
	}
	if segments[1].Values[layout.KeyPixKeyType] != "05" {
		t.Fatalf("KeyPixKeyType = %v, want \"05\" (PIX by bank data)", segments[1].Values[layout.KeyPixKeyType])
	}
	wantInfo := validPayee().Registration.Digits() + "01181521" + "01"
	// validPayee()'s Registration is 14 digits (a CNPJ), so no left-padding
	// is exercised here; TestPixBankDataSupplementaryInfoPadding below
	// covers the CPF (11 digit) case.
	if got := segments[0].Values[layout.KeySupplementaryInfo]; got != wantInfo {
		t.Fatalf("KeySupplementaryInfo = %q, want %q", got, wantInfo)
	}

	missingBank := PixBankData{Payee: validPayee(), Account: validAccount(), Amount: 500, Date: now.AddDate(0, 0, 1)}
	if err := missingBank.validate(now, validateOptions{}); err == nil {
		t.Fatal("validate() error = nil, want an error for missing BankCode")
	}
}

// TestPixBankDataSupplementaryInfoPadding confirms a CPF-holding payee (11
// digits) is zero-padded to 14 in the supplementary info blob, an ISPB
// shorter than 8 digits is zero-padded too, and a zero-value AccountKind
// renders as "00" rather than an empty string, so the field always comes
// out exactly 24 characters wide as Sicredi's manual requires.
func TestPixBankDataSupplementaryInfoPadding(t *testing.T) {
	cpf, err := NewCPF("11144477735")
	if err != nil {
		t.Fatalf("NewCPF() error = %v", err)
	}
	p := PixBankData{
		Payee:    Payee{Name: "FAVORECIDO", Registration: cpf},
		BankCode: "001",
		ISPB:     "181521",
		Account:  validAccount(),
		Amount:   500,
		Date:     time.Now().AddDate(0, 0, 1),
	}

	segments, err := p.toSegments(nil)
	if err != nil {
		t.Fatalf("toSegments() error = %v", err)
	}
	got, _ := segments[0].Values[layout.KeySupplementaryInfo].(string)
	if len(got) != 24 {
		t.Fatalf("len(KeySupplementaryInfo) = %d, want 24 (got %q)", len(got), got)
	}
	want := "00011144477735" + "00181521" + "00"
	if got != want {
		t.Fatalf("KeySupplementaryInfo = %q, want %q", got, want)
	}
}
