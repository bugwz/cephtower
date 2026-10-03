package s3

import (
	"strings"
	"testing"
)

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
