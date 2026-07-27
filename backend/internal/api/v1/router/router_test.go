package router

import (
	"cephtower/backend/internal/api/v1/handler"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRouteContract(t *testing.T) {
	routes := ReadRoutesForContract(handler.New(handler.Dependencies{}))
	seen := map[string]bool{}
	for _, route := range routes {
		key := route.Method + " " + route.Path
		if seen[key] {
			t.Fatalf("duplicate route %s", key)
		}
		seen[key] = true
		if route.Method == "" || route.Path == "" || route.Handler == nil {
			t.Fatalf("incomplete route %#v", route)
		}
		if strings.Contains(route.Path, "{") {
			t.Fatalf("route must take identifiers from body, not path: %s %s", route.Method, route.Path)
		}
		if strings.Contains(route.Path, "-") {
			t.Fatalf("route must split resource path segments with /, not -: %s %s", route.Method, route.Path)
		}
		if strings.HasPrefix(route.Path, "/cluster/") && !allowedClusterManagementPath(route.Path) {
			t.Fatalf("cluster-scoped resource route must not use /cluster prefix: %s %s", route.Method, route.Path)
		}
	}
}

func allowedClusterManagementPath(path string) bool {
	switch path {
	case "/cluster/probe", "/cluster/refresh", "/cluster/capabilities":
		return true
	default:
		return false
	}
}

func TestRouteDefinitionsUseLiteralThreeValueRoutes(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || strings.HasSuffix(name, "_test.go") || name == "routers.go" {
			continue
		}
		path := filepath.Join(".", name)
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if strings.Contains(string(source), "return []Route{{") || strings.Contains(string(source), "}, {\"") {
			t.Fatalf("%s has multiple routes on one line; write one Route per line", name)
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if ok {
				if ident, ok := call.Fun.(*ast.Ident); ok {
					switch ident.Name {
					case "read", "readExternal", "mutate":
						t.Fatalf("%s uses route helper %s; use Route{\"METHOD\", \"/path\", handler}", name, ident.Name)
					}
				}
			}
			literal, ok := node.(*ast.CompositeLit)
			if !ok || len(literal.Elts) != 3 {
				return true
			}
			method, ok := literal.Elts[0].(*ast.BasicLit)
			if !ok || method.Kind != token.STRING {
				return true
			}
			if _, ok := literal.Elts[1].(*ast.BasicLit); !ok {
				t.Fatalf("%s builds route path dynamically; use a complete string literal", name)
			}
			handler, ok := literal.Elts[2].(*ast.SelectorExpr)
			if !ok {
				t.Fatalf("%s builds route handler dynamically; use h.HandlerName", name)
			}
			if receiver, ok := handler.X.(*ast.Ident); !ok || receiver.Name != "h" {
				t.Fatalf("%s route handler must use h.HandlerName", name)
			}
			return true
		})
	}
}
