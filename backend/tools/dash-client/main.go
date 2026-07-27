package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

var (
	httpMethods = map[string]bool{"get": true, "post": true, "put": true, "patch": true, "delete": true}
	goKeywords  = map[string]bool{
		"break": true, "default": true, "func": true, "interface": true, "select": true,
		"case": true, "defer": true, "go": true, "map": true, "struct": true,
		"chan": true, "else": true, "goto": true, "package": true, "switch": true,
		"const": true, "fallthrough": true, "if": true, "range": true, "type": true,
		"continue": true, "for": true, "import": true, "return": true, "var": true,
	}
)

type operation struct {
	Method         string
	Name           string
	Path           string
	Auth           bool
	Summary        string
	Parameters     []any
	BodySchema     map[string]any
	ResponseSchema map[string]any
}

func main() {
	root, err := repoRoot()
	must(err)
	openapi := mustLoadYAML(filepath.Join(root, "docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml"))
	cephDir := filepath.Join(root, "backend/internal/integration/ceph")
	dashboardDir := filepath.Join(cephDir, "dashboard")
	endpointsDir := filepath.Join(dashboardDir, "endpoints")
	typedDir := filepath.Join(dashboardDir, "typed")
	operationsByCategory := collectOperations(openapi)
	cleanup(cephDir, endpointsDir, typedDir, operationsByCategory)
	seenMethods := 0
	for _, category := range sortedKeys(operationsByCategory) {
		seenMethods += writeCategory(category, operationsByCategory[category], endpointsDir, typedDir)
	}
	fmt.Printf("generated %d raw and typed Ceph dashboard endpoint methods in %d files\n", seenMethods, len(operationsByCategory))
}

func repoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if exists(filepath.Join(wd, "backend/go.mod")) && exists(filepath.Join(wd, "docs/references/ceph")) {
			return wd, nil
		}
		next := filepath.Dir(wd)
		if next == wd {
			return "", fmt.Errorf("could not locate repository root")
		}
		wd = next
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func mustLoadYAML(path string) map[string]any {
	data, err := os.ReadFile(path)
	must(err)
	var value map[string]any
	must(yaml.Unmarshal(data, &value))
	return value
}

func asMap(value any) map[string]any {
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return map[string]any{}
}

func asSlice(value any) []any {
	if typed, ok := value.([]any); ok {
		return typed
	}
	return nil
}

func asString(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func camelize(value string) string {
	var parts []string
	var current []rune
	flush := func() {
		if len(current) == 0 {
			return
		}
		part := string(current)
		runes := []rune(part)
		runes[0] = unicode.ToUpper(runes[0])
		parts = append(parts, string(runes))
		current = nil
	}
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current = append(current, r)
		} else {
			flush()
		}
	}
	flush()
	text := strings.Join(parts, "")
	if text == "" {
		text = "Value"
	}
	first := []rune(text)[0]
	if !unicode.IsLetter(first) && first != '_' {
		text = "X" + text
	}
	if goKeywords[strings.ToLower(text)] {
		text += "Value"
	}
	return text
}

func methodName(httpMethod, path string) string {
	segments := pathParts(path)
	if len(segments) > 0 && segments[0] == "api" {
		segments = segments[1:]
	}
	var suffix []string
	for _, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			suffix = append(suffix, "By"+camelize(strings.TrimSuffix(strings.TrimPrefix(segment, "{"), "}")))
		} else {
			suffix = append(suffix, camelize(segment))
		}
	}
	name := strings.Join(suffix, "")
	if name == "" {
		name = "Root"
	}
	return camelize(httpMethod) + name
}

func pathParts(path string) []string {
	var parts []string
	for _, part := range strings.Split(path, "/") {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func categoryFor(path string) string {
	segments := pathParts(path)
	category := "root"
	if len(segments) > 0 {
		if segments[0] == "api" && len(segments) > 1 {
			category = segments[1]
		} else {
			category = segments[0]
		}
	}
	return strings.ReplaceAll(category, "-", "_")
}

func jwtRequired(operation map[string]any) bool {
	for _, value := range asSlice(operation["security"]) {
		if _, ok := asMap(value)["jwt"]; ok {
			return true
		}
	}
	return false
}

func schemaContent(container map[string]any) map[string]any {
	content := asMap(container["content"])
	if len(content) == 0 {
		return nil
	}
	for _, mime := range []string{"application/json", "application/vnd.ceph.api.v1.0+json"} {
		if spec, ok := content[mime]; ok {
			item := asMap(spec)
			if schema := asMap(item["schema"]); len(schema) > 0 {
				return schema
			}
			return item
		}
	}
	for _, key := range sortedKeys(content) {
		item := asMap(content[key])
		if schema := asMap(item["schema"]); len(schema) > 0 {
			return schema
		}
		return item
	}
	return nil
}

func successSchema(operation map[string]any) map[string]any {
	responses := asMap(operation["responses"])
	for _, code := range []string{"200", "201", "202", "204"} {
		if response, ok := responses[code]; ok {
			return schemaContent(asMap(response))
		}
	}
	for _, code := range sortedKeys(responses) {
		if strings.HasPrefix(code, "2") {
			return schemaContent(asMap(responses[code]))
		}
	}
	return nil
}

func requestBodySchema(operation map[string]any) map[string]any {
	return schemaContent(asMap(operation["requestBody"]))
}

func collectOperations(openapi map[string]any) map[string][]operation {
	operationsByCategory := map[string][]operation{}
	seen := map[string]bool{}
	for path, methods := range asMap(openapi["paths"]) {
		for httpMethod, value := range asMap(methods) {
			if !httpMethods[httpMethod] {
				continue
			}
			op := asMap(value)
			name := methodName(httpMethod, path)
			if seen[name] {
				panic("duplicate generated method " + name)
			}
			seen[name] = true
			category := categoryFor(path)
			operationsByCategory[category] = append(operationsByCategory[category], operation{
				Method: strings.ToUpper(httpMethod), Name: name, Path: path, Auth: jwtRequired(op),
				Summary: strings.TrimSpace(asString(op["summary"])), Parameters: asSlice(op["parameters"]),
				BodySchema: requestBodySchema(op), ResponseSchema: successSchema(op),
			})
		}
	}
	for category := range operationsByCategory {
		sort.Slice(operationsByCategory[category], func(i, j int) bool {
			left, right := operationsByCategory[category][i], operationsByCategory[category][j]
			if left.Path != right.Path {
				return left.Path < right.Path
			}
			return left.Method < right.Method
		})
	}
	return operationsByCategory
}

func goFieldName(name string, used map[string]bool) string {
	base := camelize(name)
	candidate := base
	index := 2
	for used[candidate] {
		candidate = fmt.Sprintf("%s%d", base, index)
		index++
	}
	used[candidate] = true
	return candidate
}

func schemaGoType(schema map[string]any, typeName string, definitions *[]string) string {
	if len(schema) == 0 {
		return "json.RawMessage"
	}
	if nullable, ok := schema["nullable"].(bool); ok && nullable {
		copySchema := map[string]any{}
		for key, value := range schema {
			if key != "nullable" {
				copySchema[key] = value
			}
		}
		nested := schemaGoType(copySchema, typeName, definitions)
		if !strings.HasPrefix(nested, "[]") && !strings.HasPrefix(nested, "map[") && nested != "json.RawMessage" {
			return "*" + nested
		}
		return nested
	}
	switch asString(schema["type"]) {
	case "object", "":
		properties := asMap(schema["properties"])
		if len(properties) == 0 {
			if additional := asMap(schema["additionalProperties"]); len(additional) > 0 {
				return "map[string]" + schemaGoType(additional, typeName+"Value", definitions)
			}
			return "map[string]json.RawMessage"
		}
		required := map[string]bool{}
		for _, value := range asSlice(schema["required"]) {
			required[asString(value)] = true
		}
		used := map[string]bool{}
		var lines []string
		lines = append(lines, "type "+typeName+" struct {")
		for _, propName := range sortedKeys(properties) {
			fieldName := goFieldName(propName, used)
			fieldType := schemaGoType(asMap(properties[propName]), typeName+fieldName, definitions)
			tag := propName
			if !required[propName] {
				tag += ",omitempty"
			}
			lines = append(lines, "\t"+fieldName+" "+fieldType+" `json:\""+tag+"\"`")
		}
		lines = append(lines, "}")
		*definitions = append(*definitions, strings.Join(lines, "\n"))
		return typeName
	case "array":
		return "[]" + schemaGoType(asMap(schema["items"]), typeName+"Item", definitions)
	case "integer":
		return "int"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	case "string":
		return "string"
	default:
		return "json.RawMessage"
	}
}

func defineNamedSchema(typeName string, schema map[string]any, definitions *[]string) string {
	if len(schema) == 0 {
		*definitions = append(*definitions, "type "+typeName+" = EmptyResponse")
		return "EmptyResponse"
	}
	goType := schemaGoType(schema, typeName, definitions)
	if goType != typeName {
		*definitions = append(*definitions, "type "+typeName+" "+goType)
	}
	return typeName
}

func queryFieldType(param map[string]any) string {
	switch asString(asMap(param["schema"])["type"]) {
	case "integer":
		return "int"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	case "array":
		return "[]string"
	default:
		return "string"
	}
}

func querySetter(fieldName string, param map[string]any) string {
	name := asString(param["name"])
	required, _ := param["required"].(bool)
	schemaType := asString(asMap(param["schema"])["type"])
	value := "request." + fieldName
	target := value
	condition, close := "", ""
	if !required {
		target = "*" + value
		condition = "\tif " + value + " != nil {\n"
		close = "\t}\n"
	}
	var line string
	switch schemaType {
	case "integer":
		line = "\tquery.Set(" + quote(name) + ", strconv.Itoa(" + target + "))\n"
	case "number":
		line = "\tquery.Set(" + quote(name) + ", strconv.FormatFloat(" + target + ", 'f', -1, 64))\n"
	case "boolean":
		line = "\tquery.Set(" + quote(name) + ", strconv.FormatBool(" + target + "))\n"
	case "array":
		line = "\tfor _, value := range " + target + " {\n\t\tquery.Add(" + quote(name) + ", value)\n\t}\n"
	default:
		line = "\tquery.Set(" + quote(name) + ", " + target + ")\n"
	}
	return condition + line + close
}

func quote(value string) string {
	return fmt.Sprintf("%q", value)
}

func rawHeader(category string) string {
	return `// Code generated by the Ceph Dashboard client generator; DO NOT EDIT.

package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
)

// ` + category + ` endpoints from docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml.
`
}

func typedHeader(category string, needsJSON, needsStrconv bool) string {
	imports := []string{"\t\"context\""}
	if needsJSON {
		imports = append(imports, "\t\"encoding/json\"")
	}
	imports = append(imports, "\t\"net/http\"", "\t\"net/url\"")
	if needsStrconv {
		imports = append(imports, "\t\"strconv\"")
	}
	return `// Code generated by the Ceph Dashboard client generator; DO NOT EDIT.

package typed

import (
` + strings.Join(imports, "\n") + `
)

// ` + category + ` typed endpoints from docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml.
`
}

func writeCategory(category string, operations []operation, endpointsDir, typedDir string) int {
	rawBody := rawHeader(category)
	var typedDefs []string
	var typedMethods []string
	needsStrconv := false
	for _, operation := range operations {
		doc := operation.Summary
		if doc == "" {
			doc = operation.Method + " " + operation.Path
		}
		rawBody += "\n// " + operation.Name + " calls " + operation.Method + " " + operation.Path + ".\n"
		rawBody += "// " + strings.Join(strings.Fields(doc), " ") + "\n"
		rawBody += "func (c *Client) " + operation.Name + "(ctx context.Context, request OperationRequest) (json.RawMessage, error) {\n"
		rawBody += "\treturn c.do(ctx, http.Method" + camelize(strings.ToLower(operation.Method)) + ", " + quote(operation.Path) + ", request, " + fmt.Sprint(operation.Auth) + ")\n"
		rawBody += "}\n"

		requestType := operation.Name + "Request"
		responseType := operation.Name + "Response"
		bodyType := operation.Name + "Body"
		var localDefs []string
		var fieldLines []string
		var pathLines []string
		queryLines := []string{"\tquery := url.Values{}\n"}
		usedFields := map[string]bool{}
		for _, value := range operation.Parameters {
			param := asMap(value)
			location := asString(param["in"])
			if location != "path" && location != "query" {
				continue
			}
			fieldName := goFieldName(asString(param["name"]), usedFields)
			if location == "path" {
				fieldLines = append(fieldLines, "\t"+fieldName+" string `path:\""+asString(param["name"])+"\"`")
				pathLines = append(pathLines, "\t\t"+quote(asString(param["name"]))+": request."+fieldName+",\n")
			} else {
				baseType := queryFieldType(param)
				if baseType == "int" || baseType == "float64" || baseType == "bool" {
					needsStrconv = true
				}
				fieldType := baseType
				required, _ := param["required"].(bool)
				if !required {
					fieldType = "*" + baseType
				}
				fieldLines = append(fieldLines, "\t"+fieldName+" "+fieldType+" `query:\""+asString(param["name"])+"\"`")
				queryLines = append(queryLines, querySetter(fieldName, param))
			}
		}
		bodyGoType := ""
		if len(operation.BodySchema) > 0 {
			bodyGoType = defineNamedSchema(bodyType, operation.BodySchema, &localDefs)
			fieldLines = append(fieldLines, "\tBody "+bodyGoType+" `json:\"-\"`")
		}
		defineNamedSchema(responseType, operation.ResponseSchema, &localDefs)
		typedDefs = append(typedDefs, localDefs...)
		if len(fieldLines) == 0 {
			typedDefs = append(typedDefs, "type "+requestType+" struct{}")
		} else {
			typedDefs = append(typedDefs, "type "+requestType+" struct {\n"+strings.Join(fieldLines, "\n")+"\n}")
		}
		method := "// " + operation.Name + " calls " + operation.Method + " " + operation.Path + " with typed request and response values.\n"
		method += "func (c *Client) " + operation.Name + "(ctx context.Context, request " + requestType + ") (" + responseType + ", error) {\n"
		method += "\tpath := map[string]string{\n" + strings.Join(pathLines, "") + "\t}\n"
		method += strings.Join(queryLines, "")
		method += "\tvar body any\n\tbody = nil\n"
		if bodyGoType != "" {
			method += "\tbody = request.Body\n"
		}
		method += "\tvar response " + responseType + "\n"
		method += "\terr := c.doJSON(ctx, http.Method" + camelize(strings.ToLower(operation.Method)) + ", " + quote(operation.Path) + ", path, query, body, " + fmt.Sprint(operation.Auth) + ", &response)\n"
		method += "\treturn response, err\n"
		method += "}\n"
		typedMethods = append(typedMethods, method)
	}
	must(os.WriteFile(filepath.Join(endpointsDir, category+".go"), []byte(rawBody), 0o644))
	typedBody := typedHeader(category, strings.Contains(strings.Join(typedDefs, "\n"), "json.RawMessage"), needsStrconv)
	typedBody += "\n" + strings.Join(typedDefs, "\n\n") + "\n\n" + strings.Join(typedMethods, "\n")
	must(os.WriteFile(filepath.Join(typedDir, category+".go"), []byte(typedBody), 0o644))
	return len(operations)
}

func cleanup(cephDir, endpointsDir, typedDir string, operationsByCategory map[string][]operation) {
	must(os.MkdirAll(endpointsDir, 0o755))
	must(os.MkdirAll(typedDir, 0o755))
	removeGlob(filepath.Join(cephDir, "zz_generated_*.go"))
	removeGlob(filepath.Join(endpointsDir, "generated_*.go"))
	typedFiles, err := filepath.Glob(filepath.Join(typedDir, "*.go"))
	must(err)
	for _, path := range typedFiles {
		if filepath.Base(path) != "client.go" {
			must(os.Remove(path))
		}
	}
	for category := range operationsByCategory {
		must(os.RemoveAll(filepath.Join(endpointsDir, category+".go")))
		must(os.RemoveAll(filepath.Join(typedDir, category+".go")))
	}
}

func removeGlob(pattern string) {
	paths, err := filepath.Glob(pattern)
	must(err)
	for _, path := range paths {
		must(os.Remove(path))
	}
}
