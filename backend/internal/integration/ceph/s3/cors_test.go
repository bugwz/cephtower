package s3

import (
	"strings"
	"testing"
)

func TestCORSMaxAgeNativeSemantics(t *testing.T) {
	for text, want := range map[string]uint32{"": 0, "0": 0, "-0": 0, "+001": 1, " \t42": 42, "4294967294": 4294967294, "-18446744073709551615": 1} {
		got, err := corsMaxAge(text)
		if err != nil || got == nil || *got != want {
			t.Fatalf("%q: %v %v want %d", text, got, err, want)
		}
	}
	for _, text := range []string{"4294967295", "4294967296", "18446744073709551615", "18446744073709551616", "-1", "-18446744073709551616", strings.Repeat("9", 100)} {
		got, err := corsMaxAge(text)
		if err != nil || got != nil {
			t.Fatalf("%q: expected native omission, got %v %v", text, got, err)
		}
	}
	for _, text := range []string{" ", "+", "-", "1 ", "1\n", "1.5", "0x10", "1_000", "１２", "\u00a01"} {
		if _, err := corsMaxAge(text); err == nil {
			t.Fatalf("accepted %q", text)
		}
		body := []byte("<CORSConfiguration><CORSRule><AllowedOrigin>*</AllowedOrigin><MaxAgeSeconds>" + text + "</MaxAgeSeconds></CORSRule></CORSConfiguration>")
		if err := ValidateBucketConfiguration("cors", body); err == nil {
			t.Fatalf("accepted invalid XML age %q", text)
		}
	}
}

func TestBucketCORSNativeRuleValidation(t *testing.T) {
	wrap := func(rule string) []byte {
		return []byte("<CORSConfiguration><CORSRule>" + rule + "</CORSRule></CORSConfiguration>")
	}
	for _, extra := range []string{"", "<AllowedMethod>get</AllowedMethod><AllowedMethod>COPY</AllowedMethod>", "<ID>" + strings.Repeat("a", 255) + "</ID>", "<AllowedHeader>x-*</AllowedHeader><ExposeHeader/>"} {
		if err := ValidateBucketConfiguration("cors", wrap("<AllowedOrigin>*</AllowedOrigin>"+extra)); err != nil {
			t.Fatal(err)
		}
	}
	for _, rule := range []string{"", "<AllowedMethod>GET</AllowedMethod>", "<AllowedOrigin/>", "<AllowedOrigin>**</AllowedOrigin>", "<AllowedOrigin>*</AllowedOrigin><AllowedMethod>PATCH</AllowedMethod>", "<AllowedOrigin>*</AllowedOrigin><AllowedHeader/>", "<AllowedOrigin>*</AllowedOrigin><ID>" + strings.Repeat("é", 128) + "</ID>", "<AllowedOrigin>*</AllowedOrigin><ID/><ID/>", "<AllowedOrigin>*</AllowedOrigin><Unknown/>", "<AllowedOrigin><Nested/></AllowedOrigin>", "<AllowedOrigin>*</AllowedOrigin><MaxAgeSeconds>1</MaxAgeSeconds><MaxAgeSeconds>2</MaxAgeSeconds>"} {
		if err := ValidateBucketConfiguration("cors", wrap(rule)); err == nil {
			t.Fatalf("accepted %s", rule)
		}
	}
}
