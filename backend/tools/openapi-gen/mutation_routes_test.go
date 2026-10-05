package main

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/api/v1/router"
)

func TestMutationRouteActionsMatchHandlers(t *testing.T) {
	// Read the actual wrapper bodies rather than deriving actions from route names.
	wrappers := regexp.MustCompile(`func \(h \*Handler\) (\w+)\([^\n]*\) \{\s*h\.MutateResource\("[^"]+", "([^"]+)"`)
	routes := regexp.MustCompile(`\{"(\w+)", "([^"]+)", h\.(\w+)\}`)
	read := func(pattern string) []string {
		files, err := filepath.Glob(pattern)
		if err != nil || len(files) == 0 {
			t.Fatalf("source files: %v", err)
		}
		var sources []string
		for _, file := range files {
			if strings.HasSuffix(file, "_test.go") {
				continue
			}
			body, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			sources = append(sources, string(body))
		}
		return sources
	}
	actions := map[string]string{}
	for _, source := range read("../../internal/api/v1/handler/*.go") {
		for _, match := range wrappers.FindAllStringSubmatch(source, -1) {
			actions[match[1]] = match[2]
		}
	}
	// This handler validates native semantics and removes routing fields before enqueueing.
	encryptionSource, err := os.ReadFile("../../internal/api/v1/handler/rgw_encryption.go")
	if err != nil {
		t.Fatal(err)
	}
	encryptionAction := regexp.MustCompile(`Action:\s*"([^"]+)"`).FindStringSubmatch(string(encryptionSource))
	if len(encryptionAction) != 2 {
		t.Fatal("missing encryption enqueue action")
	}
	actions["UpdateRGWEncryptionConfiguration"] = encryptionAction[1]
	expected := map[string]string{}
	for _, source := range read("../../internal/api/v1/router/*.go") {
		for _, match := range routes.FindAllStringSubmatch(source, -1) {
			if action, ok := actions[match[3]]; ok {
				expected[match[1]+" "+match[2]] = action
			}
		}
	}
	if !reflect.DeepEqual(expected, mutationRouteActions) {
		t.Fatal("mutation route mapping differs from runtime handlers")
	}
	for key, action := range expected {
		parts := strings.SplitN(key, " ", 2)
		got, ok := requestSchema(router.Route{Method: parts[0], Path: parts[1]})
		want, exists := handler.MutationRequestContract(action)
		if !ok || !exists || !reflect.DeepEqual(got, want) {
			t.Errorf("%s does not use %s contract", key, action)
		}
	}
	// Every registered route must have an explicit schema or be a supported read.
	for _, route := range router.ReadRoutesForContract(handler.New(handler.Dependencies{})) {
		requestSchema(route)
	}
}

func TestMutationSchemasDoNotLeakBucketConfigurationFields(t *testing.T) {
	for _, key := range []string{"POST /host", "POST /pool", "PATCH /rgw/bucket"} {
		parts := strings.SplitN(key, " ", 2)
		schema, _ := requestSchema(router.Route{Method: parts[0], Path: parts[1]})
		for _, name := range []string{"kind", "document"} {
			if _, exists := schema.Fields[name]; exists {
				t.Errorf("%s leaks %s", key, name)
			}
		}
	}
	schema, _ := requestSchema(router.Route{Method: "PATCH", Path: "/rgw/bucket/policy"})
	if !schema.Fields["kind"].Required || !schema.Fields["document"].Required {
		t.Fatal("required configuration fields lost")
	}
}

func TestHostSSHPasswordSchema(t *testing.T) {
	schema, _ := requestSchema(router.Route{Method: "PATCH", Path: "/host/ssh"})
	var output strings.Builder
	writeRequestSchema(&output, schema, 0)
	if !strings.Contains(output.String(), "writeOnly: true") || !strings.Contains(output.String(), "type: 'null'") {
		t.Fatal("nullable secret metadata lost")
	}
	if _, ok := schema.Fields["ssh_private_key"]; ok {
		t.Fatal("unsupported SSH field")
	}
}
