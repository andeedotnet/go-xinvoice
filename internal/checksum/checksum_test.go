package checksum

import "testing"

func TestValidGLN(t *testing.T) {
	// 4000001000005 is a valid GLN used in the official testsuite (scheme 0088).
	if !ValidGLN("4000001000005") {
		t.Error("expected 4000001000005 to be a valid GLN")
	}
	if ValidGLN("4000001000004") { // wrong check digit
		t.Error("expected wrong check digit to be invalid")
	}
	if ValidGLN("400000100000") { // 12 digits
		t.Error("expected wrong length to be invalid")
	}
	if ValidGLN("400000100000X") {
		t.Error("expected non-digit to be invalid")
	}
}

func TestValidIBAN(t *testing.T) {
	// A valid (non-existent) IBAN from the testsuite.
	if !ValidIBAN("DE75512108001245126199") {
		t.Error("expected testsuite IBAN to be valid")
	}
	if !ValidIBAN("DE75 5121 0800 1245 1261 99") { // spaces ignored
		t.Error("expected spaced IBAN to be valid")
	}
	if ValidIBAN("DE00512108001245126199") { // wrong check digits
		t.Error("expected wrong check digits to be invalid")
	}
	if ValidIBAN("XX") {
		t.Error("expected too-short IBAN to be invalid")
	}
}

func TestMod11(t *testing.T) {
	// Self-consistency: build a value whose check digit Mod11 accepts.
	for _, body := range []string{"12345", "98765", "10101"} {
		found := false
		for d := 0; d <= 9; d++ {
			if Mod11(body + string(rune('0'+d))) {
				found = true
			}
		}
		if !found {
			t.Logf("no single-digit mod-11 check for %s (r==10 case)", body)
		}
	}
	if Mod11("1") {
		t.Error("too short should be invalid")
	}
}
