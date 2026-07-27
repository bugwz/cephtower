package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	httpMethods  = map[string]bool{"get": true, "post": true, "put": true, "patch": true, "delete": true}
	methodOrder  = map[string]int{"get": 0, "post": 1, "put": 2, "patch": 3, "delete": 4}
	cephVersions = []string{"v16.2.15", "v17.2.9", "v18.2.8", "v19.2.5", "v20.2.2"}
	categoryName = map[string]string{
		"auth": "认证", "block": "块存储 RBD", "cephfs": "CephFS", "cluster": "集群",
		"cluster_conf": "集群配置", "crush_rule": "CRUSH 规则", "daemon": "Daemon",
		"erasure_code_profile": "纠删码配置", "feature_toggles": "功能开关", "feedback": "反馈",
		"grafana": "Grafana", "hardware": "硬件", "health": "健康状态", "host": "主机",
		"iscsi": "iSCSI", "logs": "日志", "mgr": "Mgr 模块", "monitor": "Monitor",
		"motd": "MOTD", "multi-cluster": "多集群", "nfs-ganesha": "NFS Ganesha",
		"nvmeof": "NVMe-oF", "osd": "OSD", "perf_counters": "性能计数器",
		"pool": "存储池", "prometheus": "Prometheus", "rgw": "RGW", "role": "角色",
		"service": "服务", "settings": "设置", "smb": "SMB", "summary": "概览",
		"task": "任务", "telemetry": "Telemetry", "user": "Dashboard 用户",
	}
)

type operation struct {
	Path   string
	Method string
	Spec   map[string]any
}

type operationSummary struct {
	Path    string
	Method  string
	Summary string
	Tags    []string
}

func main() {
	root, err := repoRoot()
	must(err)
	cephRepo := filepath.Join(root, "docs/references/ceph")
	openapi := mustLoadYAML(filepath.Join(cephRepo, "src/pybind/mgr/dashboard/openapi.yaml"))
	version := cephVersion(filepath.Join(cephRepo, "CMakeLists.txt"))
	grouped := collectOperations(openapi)
	versionIndexes := buildVersionIndexes(cephRepo)
	outDir := filepath.Join(root, "docs/ceph/apis")
	apiDir := filepath.Join(outDir, "endpoints")
	cleanupGeneratedFiles(outDir, apiDir)
	for _, category := range sortedKeys(grouped) {
		writeCategory(apiDir, category, grouped[category], version, versionIndexes)
	}
	writeIndex(outDir, grouped, version, versionIndexes)
	writeCompatibility(outDir, versionIndexes)
	total := 0
	for _, ops := range grouped {
		total += len(ops)
	}
	fmt.Printf("Generated %d markdown files for %d operations.\n", len(grouped)+2, total)
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

func gitShow(cephRepo, tag, path string) ([]byte, error) {
	cmd := exec.Command("git", "-C", cephRepo, "show", tag+":"+path)
	return cmd.Output()
}

func loadOpenAPIForTag(cephRepo, tag string) map[string]any {
	data, err := gitShow(cephRepo, tag, "src/pybind/mgr/dashboard/openapi.yaml")
	must(err)
	var value map[string]any
	must(yaml.Unmarshal(data, &value))
	return value
}

func cephVersion(cmakePath string) string {
	data, err := os.ReadFile(cmakePath)
	must(err)
	re := regexp.MustCompile(`(?s)project\([^)]*VERSION\s+([0-9.]+)`)
	if match := re.FindSubmatch(data); len(match) == 2 {
		return string(match[1])
	}
	return "unknown"
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

func collectOperations(openapi map[string]any) map[string][]operation {
	grouped := map[string][]operation{}
	for path, item := range asMap(openapi["paths"]) {
		for method, spec := range asMap(item) {
			if !httpMethods[method] {
				continue
			}
			category := categoryFor(path)
			grouped[category] = append(grouped[category], operation{Path: path, Method: method, Spec: asMap(spec)})
		}
	}
	for category := range grouped {
		sort.Slice(grouped[category], func(i, j int) bool {
			left, right := grouped[category][i], grouped[category][j]
			if left.Path != right.Path {
				return left.Path < right.Path
			}
			return methodOrder[left.Method] < methodOrder[right.Method]
		})
	}
	return grouped
}

func collectOperationIndex(openapi map[string]any) map[string]operationSummary {
	index := map[string]operationSummary{}
	for path, item := range asMap(openapi["paths"]) {
		for method, spec := range asMap(item) {
			if !httpMethods[method] {
				continue
			}
			op := asMap(spec)
			index[operationKey(method, path)] = operationSummary{
				Path: path, Method: method, Summary: asString(op["summary"]), Tags: stringSlice(op["tags"]),
			}
		}
	}
	return index
}

func buildVersionIndexes(cephRepo string) map[string]map[string]operationSummary {
	indexes := map[string]map[string]operationSummary{}
	for _, tag := range cephVersions {
		indexes[tag] = collectOperationIndex(loadOpenAPIForTag(cephRepo, tag))
	}
	return indexes
}

func operationKey(method, path string) string {
	return strings.ToUpper(method) + " " + path
}

func categoryFor(path string) string {
	parts := pathParts(path)
	if len(parts) == 0 || parts[0] != "api" {
		return "root"
	}
	if len(parts) == 1 {
		return "root"
	}
	return parts[1]
}

func categoryTitle(category string) string {
	if title, ok := categoryName[category]; ok {
		return title
	}
	return category
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

func stringSlice(value any) []string {
	var values []string
	for _, item := range asSlice(value) {
		values = append(values, asString(item))
	}
	return values
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func markdownTable(headers []string, rows [][]string) string {
	var b strings.Builder
	b.WriteString("| " + strings.Join(headers, " | ") + " |\n")
	separators := make([]string, len(headers))
	for i := range separators {
		separators[i] = "---"
	}
	b.WriteString("| " + strings.Join(separators, " | ") + " |")
	for _, row := range rows {
		escaped := make([]string, len(row))
		for i, cell := range row {
			escaped[i] = escapeTableCell(cell)
		}
		b.WriteString("\n| " + strings.Join(escaped, " | ") + " |")
	}
	return b.String()
}

func escapeTableCell(value string) string {
	value = strings.ReplaceAll(value, "\n", "<br>")
	return strings.ReplaceAll(value, "|", `\|`)
}

func schemaType(schema map[string]any) string {
	if len(schema) == 0 {
		return ""
	}
	if typ := asString(schema["type"]); typ != "" {
		if typ == "array" {
			items := schemaType(asMap(schema["items"]))
			if items == "" {
				items = "object"
			}
			return "array<" + items + ">"
		}
		return typ
	}
	if _, ok := schema["properties"]; ok {
		return "object"
	}
	if ref := asString(schema["$ref"]); ref != "" {
		return ref
	}
	return "object"
}

func schemaSummary(schema map[string]any) string {
	if len(schema) == 0 {
		return "无 schema"
	}
	typ := schemaType(schema)
	props := asMap(schema["properties"])
	if len(props) == 0 {
		if typ == "" {
			return "object"
		}
		return typ
	}
	required := stringSlice(schema["required"])
	requiredText := "无"
	if len(required) > 0 {
		requiredText = strings.Join(required, ", ")
	}
	return fmt.Sprintf("%s; fields: %s; required: %s", typ, strings.Join(sortedKeys(props), ", "), requiredText)
}

func yamlBlock(value any) string {
	if value == nil {
		return "无"
	}
	data, err := yaml.Marshal(value)
	must(err)
	text := strings.TrimSpace(strings.TrimPrefix(string(data), "---\n"))
	if text == "" {
		return "无"
	}
	return text
}

func renderParameters(op map[string]any) string {
	params := asSlice(op["parameters"])
	if len(params) == 0 {
		return "无。\n"
	}
	var rows [][]string
	for _, value := range params {
		param := asMap(value)
		schema := asMap(param["schema"])
		defaultValue := ""
		if v, ok := schema["default"]; ok {
			defaultValue = fmt.Sprintf("%#v", v)
		} else if v, ok := param["default"]; ok {
			defaultValue = fmt.Sprintf("%#v", v)
		}
		rows = append(rows, []string{
			asString(param["name"]), asString(param["in"]), boolText(param["required"]),
			schemaType(schema), defaultValue, asString(param["description"]),
		})
	}
	return markdownTable([]string{"名称", "位置", "必填", "类型", "默认值", "说明"}, rows) + "\n"
}

func boolText(value any) string {
	if b, ok := value.(bool); ok && b {
		return "是"
	}
	return "否"
}

func renderRequestBody(op map[string]any) string {
	body := asMap(op["requestBody"])
	if len(body) == 0 {
		return "无请求体。\n"
	}
	lines := []string{"请求体必填：" + boolText(body["required"])}
	content := asMap(body["content"])
	for _, mime := range sortedKeys(content) {
		spec := asMap(content[mime])
		schema := asMap(spec["schema"])
		if len(schema) == 0 {
			schema = spec
		}
		lines = append(lines, "", "- Content-Type: `"+mime+"`", "- Schema: "+schemaSummary(schema), "", "```yaml", yamlBlock(schema), "```")
	}
	if len(content) == 0 {
		lines = append(lines, "", "无 content schema。")
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderResponses(op map[string]any) string {
	responses := asMap(op["responses"])
	var lines []string
	for _, code := range sortedKeys(responses) {
		response := asMap(responses[code])
		lines = append(lines, "#### `"+code+"`", "")
		if desc := asString(response["description"]); desc != "" {
			lines = append(lines, desc)
		}
		content := asMap(response["content"])
		if len(content) == 0 {
			lines = append(lines, "", "无响应体 schema。", "")
			continue
		}
		for _, mime := range sortedKeys(content) {
			spec := asMap(content[mime])
			schema := asMap(spec["schema"])
			if len(schema) == 0 {
				schema = spec
			}
			lines = append(lines, "", "- Content-Type: `"+mime+"`", "- Schema: "+schemaSummary(schema), "", "```yaml", yamlBlock(schema), "```")
		}
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func operationAnchor(method, path string) string {
	re := regexp.MustCompile(`[^a-z0-9]+`)
	anchor := re.ReplaceAllString(strings.ToLower(method+"-"+path), "-")
	return strings.Trim(anchor, "-")
}

func supportInfo(method, path string, indexes map[string]map[string]operationSummary) (versions []string, since string, current bool) {
	key := operationKey(method, path)
	for _, tag := range cephVersions {
		if _, ok := indexes[tag][key]; ok {
			versions = append(versions, tag)
		}
	}
	if len(versions) > 0 {
		since = versions[0]
	}
	current = contains(versions, "v20.2.2")
	return versions, since, current
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func renderVersionSupport(method, path string, indexes map[string]map[string]operationSummary) string {
	versions, since, current := supportInfo(method, path, indexes)
	versionText := "无"
	if len(versions) > 0 {
		versionText = strings.Join(versions, ", ")
	}
	if since == "" {
		since = "无"
	}
	currentText := "否"
	if current {
		currentText = "是"
	}
	return strings.Join([]string{
		"#### 版本支持", "",
		"- 支持版本：" + versionText,
		"- 首次出现在扫描范围：" + since,
		"- v20.2.2 当前文档支持：" + currentText,
		"",
	}, "\n")
}

func renderOperation(op operation, indexes map[string]map[string]operationSummary) string {
	title := asString(op.Spec["summary"])
	if title == "" {
		title = strings.ToUpper(op.Method) + " " + op.Path
	}
	tags := stringSlice(op.Spec["tags"])
	tagText := "无"
	if len(tags) > 0 {
		quoted := make([]string, len(tags))
		for i, tag := range tags {
			quoted[i] = "`" + tag + "`"
		}
		tagText = strings.Join(quoted, ", ")
	}
	security := "OpenAPI 未声明 JWT security，通常为公开或由控制器单独处理"
	if _, ok := op.Spec["security"]; ok {
		security = strings.ReplaceAll(yamlBlock(op.Spec["security"]), "\n", " ")
	}
	return strings.Join([]string{
		"### `" + strings.ToUpper(op.Method) + " " + op.Path + "`", "",
		"- 摘要：" + title,
		"- Tags：" + tagText,
		"- 安全：" + security,
		"",
		renderVersionSupport(op.Method, op.Path, indexes),
		"#### 请求参数", "",
		renderParameters(op.Spec), "",
		"#### 请求体", "",
		renderRequestBody(op.Spec), "",
		"#### 返回消息", "",
		renderResponses(op.Spec), "",
	}, "\n")
}

func writeCategory(apiDir, category string, ops []operation, version string, indexes map[string]map[string]operationSummary) {
	must(os.MkdirAll(apiDir, 0o755))
	var lines []string
	lines = append(lines,
		"# Ceph "+version+" Dashboard API - "+categoryTitle(category), "",
		"> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。",
		"> 本文档由 Ceph Dashboard API 文档生成器自动生成，按 `/api/"+category+"` 路径域归类。",
		"> 版本支持扫描范围："+strings.Join(cephVersions, ", ")+"。", "",
		"## 接口目录", "",
	)
	for _, op := range ops {
		summary := asString(op.Spec["summary"])
		if summary == "" {
			summary = strings.ToUpper(op.Method) + " " + op.Path
		}
		lines = append(lines, "- [`"+strings.ToUpper(op.Method)+" "+op.Path+"`](#"+operationAnchor(op.Method, op.Path)+") - "+summary)
	}
	lines = append(lines, "", "## 接口详情", "")
	for _, op := range ops {
		lines = append(lines, renderOperation(op, indexes))
	}
	must(os.WriteFile(filepath.Join(apiDir, category+".md"), []byte(strings.Join(lines, "\n")), 0o644))
}

func versionStatRows(indexes map[string]map[string]operationSummary) [][]string {
	var rows [][]string
	for _, tag := range cephVersions {
		paths := map[string]bool{}
		for _, op := range indexes[tag] {
			paths[op.Path] = true
		}
		rows = append(rows, []string{tag, fmt.Sprint(len(paths)), fmt.Sprint(len(indexes[tag]))})
	}
	return rows
}

func writeIndex(outDir string, grouped map[string][]operation, version string, indexes map[string]map[string]operationSummary) {
	totalOps := 0
	paths := map[string]bool{}
	for _, ops := range grouped {
		totalOps += len(ops)
		for _, op := range ops {
			paths[op.Path] = true
		}
	}
	var rows [][]string
	for _, category := range sortedKeys(grouped) {
		filename := category + ".md"
		rows = append(rows, []string{categoryTitle(category), "`" + category + "`", "[endpoints/" + filename + "](endpoints/" + filename + ")", fmt.Sprint(len(grouped[category]))})
	}
	lines := []string{
		"# Ceph " + version + " Mgr Dashboard API", "",
		"本文档集整理自 Ceph v" + version + " 源码内置 Dashboard OpenAPI 描述，用于本项目后续通过 mgr dashboard API 操作 Ceph 集群。", "",
		"## 来源与调用约定", "",
		"- OpenAPI 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`",
		"- 版本来源：`docs/references/ceph/CMakeLists.txt` 中的 `VERSION " + version + "`",
		"- 版本支持扫描范围：" + strings.Join(cephVersions, ", "),
		"- API 基础路径：OpenAPI `basePath` 为 `/`，接口路径以 `/api/...` 为主。",
		"- 认证方式：带 `security: [{jwt: []}]` 的接口使用 Bearer JWT。通常先调用 `POST /api/auth` 获取 `token`，后续请求使用 `Authorization: Bearer <token>`。",
		"- 公开接口：未声明 `security` 的接口通常为公开入口或由控制器单独处理，调用前仍应结合部署侧 Dashboard 配置确认。",
		"- 内容类型：请求体通常为 `application/json`；响应内容类型通常为 `application/vnd.ceph.api.v1.0+json`，部分接口使用其他 API 版本 MIME。",
		"- 异步任务：许多写操作可能返回 `202`，表示操作仍在执行，需要查询任务队列。",
		"- 通用错误：多数接口包含 `400`、`401`、`403`、`500`。具体响应体以运行时 Dashboard 返回为准。",
		"- 版本兼容性总览：[compatibility.md](compatibility.md)", "",
		"## 分类索引", "",
		markdownTable([]string{"分类", "路径域", "文档", "接口数"}, rows), "",
		"## 统计", "",
		"- 路径数：" + fmt.Sprint(len(paths)),
		"- 接口操作数：" + fmt.Sprint(totalOps),
		"- 分类数：" + fmt.Sprint(len(grouped)), "",
		"## 各版本接口数量", "",
		markdownTable([]string{"版本", "路径数", "接口操作数"}, versionStatRows(indexes)), "",
	}
	must(os.WriteFile(filepath.Join(outDir, "index.md"), []byte(strings.Join(lines, "\n")), 0o644))
}

func writeCompatibility(outDir string, indexes map[string]map[string]operationSummary) {
	current := indexes["v20.2.2"]
	allKeys := map[string]bool{}
	for _, index := range indexes {
		for key := range index {
			allKeys[key] = true
		}
	}
	var removed []string
	for key := range allKeys {
		if _, ok := current[key]; !ok {
			removed = append(removed, key)
		}
	}
	sort.Strings(removed)
	lines := []string{
		"# Ceph Dashboard API 版本兼容性", "",
		"本文档汇总 " + strings.Join(cephVersions, ", ") + " 的 Dashboard API operation 支持情况。", "",
		"## 版本统计", "",
		markdownTable([]string{"版本", "路径数", "接口操作数"}, versionStatRows(indexes)), "",
		"## v20.2.2 未包含的历史接口", "",
	}
	if len(removed) == 0 {
		lines = append(lines, "扫描范围内没有发现旧版本存在但 v20.2.2 已不在 OpenAPI 中的接口。")
	} else {
		var rows [][]string
		for _, key := range removed {
			var versions []string
			for _, tag := range cephVersions {
				if _, ok := indexes[tag][key]; ok {
					versions = append(versions, tag)
				}
			}
			rows = append(rows, []string{key, strings.Join(versions, ", ")})
		}
		lines = append(lines, markdownTable([]string{"接口", "支持版本"}, rows))
	}
	lines = append(lines, "")
	must(os.WriteFile(filepath.Join(outDir, "compatibility.md"), []byte(strings.Join(lines, "\n")), 0o644))
}

func cleanupGeneratedFiles(outDir, apiDir string) {
	must(os.MkdirAll(outDir, 0o755))
	must(os.RemoveAll(apiDir))
	must(os.MkdirAll(apiDir, 0o755))
}
