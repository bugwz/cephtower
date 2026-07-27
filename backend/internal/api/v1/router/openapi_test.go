package router

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"cephtower/backend/internal/api/v1/handler"
	"gopkg.in/yaml.v3"
)

func TestOpenAPIMatchesRoutes(t *testing.T) {
	data, err := os.ReadFile("../../../../api/openapi-v1.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Paths      map[string]map[string]any `yaml:"paths"`
		Components struct {
			Schemas map[string]any `yaml:"schemas"`
		} `yaml:"components"`
	}
	if err := yaml.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	for _, route := range ReadRoutesForContract(handler.New(handler.Dependencies{})) {
		path := document.Paths[route.Path]
		if path == nil || path[strings.ToLower(route.Method)] == nil {
			t.Fatalf("OpenAPI missing %s %s", route.Method, route.Path)
		}
		operation := path[strings.ToLower(route.Method)].(map[string]any)
		responses := operation["responses"].(map[string]any)
		foundSuccess := false
		for status, raw := range responses {
			if !strings.HasPrefix(status, "2") {
				continue
			}
			foundSuccess = true
			response := raw.(map[string]any)
			content := response["content"].(map[string]any)
			for _, media := range content {
				schema := media.(map[string]any)["schema"].(map[string]any)
				ref, _ := schema["$ref"].(string)
				name := strings.TrimPrefix(ref, "#/components/schemas/")
				if name == "" || name == "APIResponse" || document.Components.Schemas[name] == nil {
					t.Fatalf("%s %s has untyped success response %q", route.Method, route.Path, ref)
				}
			}
		}
		if !foundSuccess {
			t.Fatalf("OpenAPI missing success response for %s %s", route.Method, route.Path)
		}
	}
	if strings.Contains(string(data), "additionalProperties: true") {
		t.Fatal("OpenAPI contains an unconstrained additionalProperties declaration")
	}
	refPattern := regexp.MustCompile(`#/components/schemas/([A-Za-z0-9]+)`)
	for _, match := range refPattern.FindAllStringSubmatch(string(data), -1) {
		if document.Components.Schemas[match[1]] == nil {
			t.Fatalf("OpenAPI schema reference %s is unresolved", match[0])
		}
	}
}
