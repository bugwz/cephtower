package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	versions      = []string{"v16.2.15", "v17.2.9", "v18.2.8", "v19.2.5", "v20.2.2"}
	versionLabels = []string{"16.2.15", "17.2.9", "18.2.8", "19.2.5", "20.2.2"}
	argTypes      = map[string]string{
		"str": "CephString", "int": "CephInt", "float": "CephFloat", "bool": "CephBool",
		"Optional[str]": "CephString", "Optional[int]": "CephInt", "Optional[float]": "CephFloat", "Optional[bool]": "CephBool",
		"List[str]": "CephString", "Sequence[str]": "CephString",
	}
	flagValues = map[string]int{
		"NOFORWARD": 1 << 0, "OBSOLETE": 1 << 1, "DEPRECATED": 1 << 2,
		"MGR": 1 << 3, "POLL": 1 << 4, "HIDDEN": 1 << 5,
		"TELL": (1 << 0) | (1 << 5),
	}
	handlerArgs = map[string]bool{"_": true, "self": true, "mgr": true, "inbuf": true, "return": true}
	sourceCache = map[string]string{}
)

type param struct {
	Name       string
	Type       string
	Required   bool
	Repeated   bool
	Positional bool
	Choices    []string
	Range      string
	Goodchars  string
}

type command struct {
	Prefix          string
	Desc            string
	Component       string
	Module          string
	Perm            string
	Source          string
	Versions        map[string]bool
	ParamsByVersion map[string][]param
	FlagsByVersion  map[string][]string
}

func main() {
	root, err := repoRoot()
	must(err)
	cephRepo := filepath.Join(root, "docs/references/ceph")
	outDir := filepath.Join(root, "docs/ceph/cmds")
	commands := mergeCommands(cephRepo)
	writeDocs(outDir, commands)
	fmt.Printf("generated %d commands into %s\n", len(commands), relative(root, outDir))
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

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func relative(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return rel
}

func gitShow(cephRepo, tag, path string) (string, error) {
	key := tag + ":" + path
	if value, ok := sourceCache[key]; ok {
		return value, nil
	}
	cmd := exec.Command("git", "-C", cephRepo, "show", tag+":"+path)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	value := string(out)
	sourceCache[key] = value
	return value, nil
}

func gitFiles(cephRepo, tag, path string) []string {
	cmd := exec.Command("git", "-C", cephRepo, "ls-tree", "-r", "--name-only", tag, path)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var files []string
	for _, line := range strings.Split(string(out), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}

func stripComments(src string) string {
	block := regexp.MustCompile(`(?s)/\*.*?\*/`)
	src = block.ReplaceAllString(src, "")
	line := regexp.MustCompile(`//.*`)
	return line.ReplaceAllString(src, "")
}

func splitMacroArgs(text string) []string {
	var args []string
	var cur strings.Builder
	depth := 0
	inString := false
	escaped := false
	for _, r := range text {
		if inString {
			cur.WriteRune(r)
			if escaped {
				escaped = false
			} else if r == '\\' {
				escaped = true
			} else if r == '"' {
				inString = false
			}
			continue
		}
		switch r {
		case '"':
			inString = true
			cur.WriteRune(r)
		case '(':
			depth++
			cur.WriteRune(r)
		case ')':
			depth--
			cur.WriteRune(r)
		case ',':
			if depth == 0 {
				args = append(args, strings.TrimSpace(cur.String()))
				cur.Reset()
			} else {
				cur.WriteRune(r)
			}
		default:
			cur.WriteRune(r)
		}
	}
	if strings.TrimSpace(cur.String()) != "" {
		args = append(args, strings.TrimSpace(cur.String()))
	}
	return args
}

func cString(expr string, constants map[string]string) string {
	expr = strings.ReplaceAll(expr, "\\\n", " ")
	for name, value := range constants {
		expr = regexp.MustCompile(`\b`+regexp.QuoteMeta(name)+`\b`).ReplaceAllString(expr, strconv.Quote(value))
	}
	re := regexp.MustCompile(`"(?:\\.|[^"\\])*"`)
	var out strings.Builder
	for _, token := range re.FindAllString(expr, -1) {
		value, err := strconv.Unquote(token)
		if err == nil {
			out.WriteString(value)
		}
	}
	return out.String()
}

func flags(expr string) []string {
	names := regexp.MustCompile(`\b[A-Z_]+\b`).FindAllString(expr, -1)
	value := 0
	for _, name := range names {
		value |= flagValues[name]
	}
	var found []string
	for name, bit := range flagValues {
		if name == "TELL" {
			continue
		}
		if value&bit != 0 {
			found = append(found, strings.ToLower(name))
		}
	}
	sort.Strings(found)
	return unique(found)
}

func parseSignature(sig string) (string, []param) {
	positional := true
	var prefix []string
	var params []param
	for _, token := range strings.Fields(sig) {
		if token == "--" {
			positional = false
			continue
		}
		if !strings.Contains(token, "=") {
			prefix = append(prefix, token)
			continue
		}
		pairs := map[string]string{}
		for _, part := range strings.Split(token, ",") {
			if key, value, ok := strings.Cut(part, "="); ok {
				pairs[key] = value
			}
		}
		name := pairs["name"]
		if name == "" {
			name = "arg"
		}
		typ := pairs["type"]
		if typ == "" {
			typ = "CephString"
		}
		if typ == "CephBool" {
			positional = false
		}
		paramPositional := positional
		if value, ok := pairs["positional"]; ok {
			paramPositional = value != "false"
		}
		var choices []string
		if value := pairs["strings"]; value != "" {
			choices = strings.Split(value, "|")
		}
		params = append(params, param{
			Name: name, Type: typ, Required: pairs["req"] != "false", Repeated: pairs["n"] == "N",
			Positional: paramPositional, Choices: choices, Range: strings.ReplaceAll(pairs["range"], "|", ".."),
			Goodchars: pairs["goodchars"],
		})
	}
	return strings.Join(prefix, " "), params
}

func parseCPPCommands(cephRepo, tag, path, component string) []extractedCommand {
	src, err := gitShow(cephRepo, tag, path)
	if err != nil {
		return nil
	}
	src = stripComments(src)
	constants := parseDefines(src)
	var results []extractedCommand
	for _, match := range regexp.MustCompile(`\b(COMMAND|COMMAND_WITH_FLAG)\s*\(`).FindAllStringSubmatchIndex(src, -1) {
		name := src[match[2]:match[3]]
		start := match[1]
		end := matchingParenEnd(src, start)
		if end < 0 {
			continue
		}
		args := splitMacroArgs(src[start:end])
		if len(args) < 4 {
			continue
		}
		sig := cString(args[0], constants)
		desc := cString(args[1], constants)
		module := cString(args[2], constants)
		perm := cString(args[3], constants)
		prefix, params := parseSignature(sig)
		fl := []string{}
		if name == "COMMAND_WITH_FLAG" && len(args) > 4 {
			fl = flags(args[4])
		}
		if contains(fl, "hidden") {
			continue
		}
		results = append(results, extractedCommand{
			Command: &command{Prefix: prefix, Desc: desc, Component: component, Module: module, Perm: perm, Source: path},
			Params:  params, Flags: fl,
		})
	}
	return results
}

type extractedCommand struct {
	Command *command
	Params  []param
	Flags   []string
}

func parseDefines(src string) map[string]string {
	re := regexp.MustCompile(`(?m)#define\s+(\w+)\s+(.+)$`)
	constants := map[string]string{}
	for _, match := range re.FindAllStringSubmatch(src, -1) {
		constants[match[1]] = cString(match[2], constants)
	}
	return constants
}

func matchingParenEnd(src string, start int) int {
	depth := 1
	inString := false
	escaped := false
	for i := start; i < len(src); i++ {
		ch := src[i]
		if inString {
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == '"' {
				inString = false
			}
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func daemonTargetForPath(path string) (module, target string, ok bool) {
	prefixes := []struct{ prefix, module, target string }{
		{"src/common/", "common", "<daemon>.<id>"},
		{"src/mds/", "mds", "mds.<id>"},
		{"src/osd/", "osd", "osd.<id>"},
		{"src/mon/", "mon", "mon.<id>"},
		{"src/mgr/", "mgr", "mgr.<id>"},
		{"src/client/", "client", "client.<id>"},
		{"src/osdc/", "client", "client.<id>"},
		{"src/rgw/", "rgw", "client.rgw.<id>"},
		{"src/os/bluestore/", "bluestore", "osd.<id>"},
		{"src/tools/cephfs_mirror/", "cephfs-mirror", "cephfs-mirror.<id>"},
		{"src/tools/rbd_mirror/", "rbd-mirror", "rbd-mirror.<id>"},
	}
	for _, item := range prefixes {
		if strings.HasPrefix(path, item.prefix) {
			return item.module, item.target, true
		}
	}
	return "", "", false
}

func parseAdminSocketCommands(cephRepo, tag string) []extractedCommand {
	cmd := exec.Command("git", "-C", cephRepo, "grep", "-l", "register_command", tag, "--", "src")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var results []extractedCommand
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, ":") {
			continue
		}
		path := strings.SplitN(line, ":", 2)[1]
		if !(strings.HasSuffix(path, ".cc") || strings.HasSuffix(path, ".h")) || strings.Contains(path, "/test/") || strings.Contains(path, "/tests/") {
			continue
		}
		module, daemonTarget, ok := daemonTargetForPath(path)
		if !ok {
			continue
		}
		src, err := gitShow(cephRepo, tag, path)
		if err != nil || !strings.Contains(src, "register_command") {
			continue
		}
		src = stripComments(src)
		constants := parseDefines(src)
		for _, match := range regexp.MustCompile(`\bregister_command\s*\(`).FindAllStringIndex(src, -1) {
			start := match[1]
			end := matchingParenEnd(src, start)
			if end < 0 {
				continue
			}
			args := splitMacroArgs(src[start:end])
			if len(args) < 2 {
				continue
			}
			sig := strings.TrimSpace(cString(args[0], constants))
			if sig == "" {
				continue
			}
			desc := ""
			if len(args) > 2 {
				desc = strings.TrimSpace(cString(args[2], constants))
			}
			innerPrefix, params := parseSignature(sig)
			prefix := strings.TrimSpace("daemon " + daemonTarget + " " + innerPrefix)
			results = append(results, extractedCommand{
				Command: &command{Prefix: prefix, Desc: desc, Component: "admin socket", Module: module, Perm: "admin-socket", Source: path},
				Params:  params, Flags: []string{"admin-socket"},
			})
			tellPrefix := strings.TrimSpace("tell " + daemonTarget + " " + innerPrefix)
			results = append(results, extractedCommand{
				Command: &command{Prefix: tellPrefix, Desc: desc, Component: "ceph tell", Module: module, Perm: "rw", Source: path},
				Params:  params, Flags: []string{"admin-socket", "tell"},
			})
		}
	}
	return results
}

func parsePythonCommands(cephRepo, tag string) []extractedCommand {
	var results []extractedCommand
	for _, path := range gitFiles(cephRepo, tag, "src/pybind/mgr") {
		if !strings.HasSuffix(path, ".py") || strings.Contains(path, "/tests/") {
			continue
		}
		src, err := gitShow(cephRepo, tag, path)
		if err != nil {
			continue
		}
		module := "mgr"
		parts := strings.Split(path, "/")
		if len(parts) > 3 {
			module = parts[3]
		}
		results = append(results, parsePythonCommandDicts(src, path, module)...)
		results = append(results, parsePythonDecorators(src, path, module)...)
	}
	return results
}

func parsePythonCommandDicts(src, path, module string) []extractedCommand {
	var results []extractedCommand
	re := regexp.MustCompile(`(?s)\{[^{}]*(?:['"]cmd['"]|['"]prefix['"])[^{}]*\}`)
	for _, block := range re.FindAllString(src, -1) {
		sig := pyStringValue(block, "cmd")
		if sig == "" {
			sig = pyStringValue(block, "prefix")
		}
		if strings.TrimSpace(sig) == "" {
			continue
		}
		prefix, params := parseSignature(sig)
		desc := pyStringValue(block, "desc")
		perm := pyStringValue(block, "perm")
		if perm == "" {
			perm = "rw"
		}
		results = append(results, extractedCommand{
			Command: &command{Prefix: prefix, Desc: desc, Component: "mgr module", Module: module, Perm: perm, Source: path},
			Params:  params, Flags: []string{"mgr"},
		})
	}
	return results
}

func parsePythonDecorators(src, path, module string) []extractedCommand {
	lines := strings.Split(src, "\n")
	var results []extractedCommand
	var decorators []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "@") {
			if strings.Contains(trimmed, "CLICommand") || strings.Contains(trimmed, "Command") {
				decorators = append(decorators, trimmed)
			}
			continue
		}
		if strings.HasPrefix(trimmed, "def ") {
			fnName, args := parsePythonDef(trimmed)
			_ = fnName
			for _, dec := range decorators {
				prefix, perm, poll, ok := decoratorInfo(dec)
				if !ok {
					continue
				}
				if module == "smb" && !strings.HasPrefix(prefix, "smb ") {
					prefix = "smb " + prefix
				}
				results = append(results, extractedCommand{
					Command: &command{Prefix: prefix, Desc: "", Component: "mgr module", Module: module, Perm: perm, Source: path},
					Params:  parseFunctionParams(args), Flags: append([]string{"mgr"}, pollFlag(poll)...),
				})
			}
			decorators = nil
			continue
		}
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			decorators = nil
		}
	}
	return results
}

func pyStringValue(block, key string) string {
	re := regexp.MustCompile(`['"]` + regexp.QuoteMeta(key) + `['"]\s*:\s*(['"])(.*?)\1`)
	if match := re.FindStringSubmatch(block); len(match) == 3 {
		return match[2]
	}
	return ""
}

func parsePythonDef(line string) (string, []string) {
	start := strings.Index(line, "def ")
	if start < 0 {
		return "", nil
	}
	rest := line[start+4:]
	nameEnd := strings.Index(rest, "(")
	if nameEnd < 0 {
		return strings.TrimSpace(rest), nil
	}
	name := strings.TrimSpace(rest[:nameEnd])
	close := strings.LastIndex(rest, ")")
	if close < nameEnd {
		return name, nil
	}
	argsText := rest[nameEnd+1 : close]
	var args []string
	for _, arg := range splitMacroArgs(argsText) {
		args = append(args, strings.TrimSpace(arg))
	}
	return name, args
}

func decoratorInfo(dec string) (prefix, perm string, poll bool, ok bool) {
	if !strings.Contains(dec, "CLICommand") && !strings.Contains(dec, "Command") {
		return "", "", false, false
	}
	method := "Write"
	if strings.Contains(dec, ".Read") {
		method = "Read"
	}
	perm = "rw"
	if method == "Read" {
		perm = "r"
	} else if method == "Write" {
		perm = "w"
	}
	if open := strings.Index(dec, "("); open >= 0 {
		close := strings.LastIndex(dec, ")")
		if close > open {
			args := splitMacroArgs(dec[open+1 : close])
			if len(args) > 0 && !strings.Contains(args[0], "=") {
				prefix = unquotePythonString(args[0])
			}
			for _, arg := range args {
				key, value, has := strings.Cut(arg, "=")
				if !has {
					continue
				}
				key = strings.TrimSpace(key)
				value = strings.TrimSpace(value)
				switch key {
				case "prefix":
					prefix = unquotePythonString(value)
				case "perm":
					perm = unquotePythonString(value)
				case "poll":
					poll = value == "True" || value == "true"
				}
			}
		}
	}
	return prefix, perm, poll, prefix != ""
}

func unquotePythonString(value string) string {
	value = strings.TrimSpace(value)
	if len(value) < 2 {
		return value
	}
	if (value[0] == '\'' && value[len(value)-1] == '\'') || (value[0] == '"' && value[len(value)-1] == '"') {
		return value[1 : len(value)-1]
	}
	return value
}

func parseFunctionParams(args []string) []param {
	var params []param
	positional := true
	firstDefault := len(args)
	for i, arg := range args {
		if strings.Contains(arg, "=") {
			firstDefault = i
			break
		}
	}
	for idx, arg := range args {
		nameType := strings.TrimSpace(strings.Split(arg, "=")[0])
		name := nameType
		annotation := ""
		if n, a, ok := strings.Cut(nameType, ":"); ok {
			name = strings.TrimSpace(n)
			annotation = strings.TrimSpace(a)
		}
		if handlerArgs[name] {
			continue
		}
		if name == "_end_positional_" {
			positional = false
			continue
		}
		typ := argTypes[annotation]
		if typ == "" {
			typ = "CephString"
		}
		if name == "format" || typ == "CephBool" {
			positional = false
		}
		params = append(params, param{Name: name, Type: typ, Required: idx < firstDefault, Positional: positional})
	}
	return params
}

func pollFlag(poll bool) []string {
	if poll {
		return []string{"poll"}
	}
	return nil
}

func mergeCommands(cephRepo string) map[string]*command {
	merged := map[string]*command{}
	for _, tag := range versions {
		version := strings.TrimPrefix(tag, "v")
		extracted := []extractedCommand{}
		extracted = append(extracted, parseCPPCommands(cephRepo, tag, "src/mon/MonCommands.h", "ceph mon")...)
		extracted = append(extracted, parseCPPCommands(cephRepo, tag, "src/mgr/MgrCommands.h", "ceph mgr")...)
		extracted = append(extracted, parseAdminSocketCommands(cephRepo, tag)...)
		extracted = append(extracted, parsePythonCommands(cephRepo, tag)...)
		for _, item := range extracted {
			key := commandKey(item.Command)
			existing := merged[key]
			if existing == nil {
				existing = item.Command
				existing.Versions = map[string]bool{}
				existing.ParamsByVersion = map[string][]param{}
				existing.FlagsByVersion = map[string][]string{}
				merged[key] = existing
			}
			existing.Versions[version] = true
			existing.ParamsByVersion[version] = item.Params
			existing.FlagsByVersion[version] = item.Flags
			if existing.Desc == "" && item.Command.Desc != "" {
				existing.Desc = item.Command.Desc
			}
		}
	}
	return merged
}

func commandKey(cmd *command) string {
	return cmd.Component + "\x00" + cmd.Module + "\x00" + cmd.Prefix
}

func commandGroup(cmd *command) string {
	if cmd.Component == "admin socket" {
		return "admin-socket/" + cmd.Module
	}
	if cmd.Component == "ceph tell" {
		return "tell/" + cmd.Module
	}
	first := strings.SplitN(cmd.Prefix, " ", 2)[0]
	cluster := map[string]bool{
		"auth": true, "config": true, "config-key": true, "df": true, "features": true,
		"fsid": true, "health": true, "log": true, "node": true, "quorum_status": true,
		"report": true, "status": true, "time-sync-status": true, "versions": true, "version": true,
	}
	if cluster[first] {
		return "cluster"
	}
	if first == "orch" {
		return "orchestrator"
	}
	return first
}

func versionRange(present map[string]bool) string {
	var values []string
	for _, version := range versionLabels {
		if present[version] {
			values = append(values, version)
		}
	}
	if len(values) == len(versionLabels) {
		return "16.2.15 - 20.2.2"
	}
	return strings.Join(values, ", ")
}

func renderParams(params []param) string {
	if len(params) == 0 {
		return "无。"
	}
	lines := []string{"| 参数 | 类型 | 必填 | 位置参数 | 说明 |", "| --- | --- | --- | --- | --- |"}
	for _, p := range params {
		var detail []string
		if p.Repeated {
			detail = append(detail, "可重复")
		}
		if len(p.Choices) > 0 {
			detail = append(detail, "可选值: "+strings.Join(p.Choices, ", "))
		}
		if p.Range != "" {
			detail = append(detail, "范围: "+p.Range)
		}
		if p.Goodchars != "" {
			detail = append(detail, "字符集: `"+p.Goodchars+"`")
		}
		if len(detail) == 0 {
			detail = append(detail, "-")
		}
		lines = append(lines, fmt.Sprintf("| `%s` | `%s` | %s | %s | %s |", p.Name, p.Type, yesNo(p.Required), yesNo(p.Positional), strings.Join(detail, "; ")))
	}
	return strings.Join(lines, "\n")
}

func yesNo(value bool) string {
	if value {
		return "是"
	}
	return "否"
}

func fileForGroup(outDir, group string) string {
	mapping := map[string]string{
		"cluster": "cluster/core.md", "auth": "cluster/auth.md", "config": "cluster/config.md",
		"config-key": "cluster/config-key.md", "mon": "components/mon.md", "mgr": "components/mgr.md",
		"mds": "components/mds.md", "fs": "components/cephfs.md", "osd": "components/osd.md",
		"pg": "components/pg.md", "orch": "mgr-modules/orchestrator.md", "orchestrator": "mgr-modules/orchestrator.md",
		"device": "mgr-modules/device.md", "dashboard": "mgr-modules/dashboard.md", "rbd": "mgr-modules/rbd.md",
		"rgw": "mgr-modules/rgw.md", "nfs": "mgr-modules/nfs.md", "smb": "mgr-modules/smb.md",
		"telemetry": "mgr-modules/telemetry.md", "prometheus": "mgr-modules/prometheus.md",
		"progress": "mgr-modules/progress.md", "iostat": "mgr-modules/iostat.md", "influx": "mgr-modules/influx.md",
		"telegraf": "mgr-modules/telegraf.md", "feedback": "mgr-modules/feedback.md", "alerts": "mgr-modules/alerts.md",
		"hello": "mgr-modules/hello.md", "count": "mgr-modules/hello.md", "healthcheck": "mgr-modules/prometheus.md",
	}
	if strings.HasPrefix(group, "admin-socket/") || strings.HasPrefix(group, "tell/") {
		return filepath.Join(outDir, group+".md")
	}
	if value, ok := mapping[group]; ok {
		return filepath.Join(outDir, value)
	}
	return filepath.Join(outDir, "mgr-modules", group+".md")
}

func writeDocs(outDir string, commands map[string]*command) {
	must(os.RemoveAll(outDir))
	must(os.MkdirAll(outDir, 0o755))
	grouped := map[string][]*command{}
	for _, cmd := range commands {
		path := fileForGroup(outDir, commandGroup(cmd))
		grouped[path] = append(grouped[path], cmd)
	}
	var paths []string
	for path := range grouped {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	indexLines := []string{
		"# Ceph 命令参考", "",
		"> 来源：`docs/references/ceph` 的 Git tags：`v16.2.15`、`v17.2.9`、`v18.2.8`、`v19.2.5`、`v20.2.2`。",
		"> 本文档由 Ceph 命令文档生成器自动生成。", "",
		"本文档整理 Ceph monitor/mgr command table、admin socket command 与 mgr Python 模块声明的 `ceph ...` 命令，",
		"用于后续在 `backend/internal/integration/ceph` 中新增直接执行 Ceph CLI 的能力。", "",
		"## 返回约定", "",
		"- 命令可通过 `ceph <prefix> --format json` 或 `--format json-pretty` 请求 JSON 输出时，仅记录 JSON 输出形态。",
		"- Ceph 源码的命令表声明参数、权限、模块与帮助文本；没有集中声明完整返回 schema。",
		"- 写操作通常返回确认文本或空输出；失败时返回非 0 退出码，并在 stdout/stderr 中携带错误说明。",
		"- `admin-socket/` 目录记录 `ceph daemon <daemon>.<id> <cmd>` 这类 per-daemon 命令。",
		"- `tell/` 目录把同一批 per-daemon 命令展开为 `ceph tell <daemon>.<id> <cmd>` 的调用形式。", "",
		"## 目录", "",
	}
	for _, path := range paths {
		rel, _ := filepath.Rel(outDir, path)
		title := strings.TrimSuffix(filepath.ToSlash(rel), ".md")
		indexLines = append(indexLines, "- ["+title+"]("+filepath.ToSlash(rel)+")")
	}
	must(os.WriteFile(filepath.Join(outDir, "index.md"), []byte(strings.Join(indexLines, "\n")+"\n"), 0o644))

	for _, path := range paths {
		cmds := grouped[path]
		sort.Slice(cmds, func(i, j int) bool {
			left, right := cmds[i], cmds[j]
			if left.Prefix != right.Prefix {
				return left.Prefix < right.Prefix
			}
			if left.Module != right.Module {
				return left.Module < right.Module
			}
			return left.Component < right.Component
		})
		must(os.MkdirAll(filepath.Dir(path), 0o755))
		title := strings.Title(strings.ReplaceAll(strings.TrimSuffix(filepath.Base(path), ".md"), "-", " "))
		lines := []string{"# " + title, "", "> 本文档由 Ceph 命令文档生成器自动生成，请勿手动修改。", ""}
		for _, cmd := range cmds {
			latest := ""
			for i := len(versionLabels) - 1; i >= 0; i-- {
				if _, ok := cmd.ParamsByVersion[versionLabels[i]]; ok {
					latest = versionLabels[i]
					break
				}
			}
			params := cmd.ParamsByVersion[latest]
			fl := cmd.FlagsByVersion[latest]
			lines = append(lines,
				"## `ceph "+cmd.Prefix+"`", "",
				"- 组件/模块：`"+cmd.Component+"` / `"+cmd.Module+"`",
				"- 支持版本："+versionRange(cmd.Versions),
				"- 权限：`"+cmd.Perm+"`",
				"- 来源：`"+cmd.Source+"`",
			)
			if len(fl) > 0 {
				lines = append(lines, "- 标志：`"+strings.Join(fl, ", ")+"`")
			}
			desc := strings.TrimSpace(cmd.Desc)
			if desc == "" {
				desc = "-"
			}
			lines = append(lines,
				"- 含义："+desc, "",
				"### 参数", "",
				renderParams(params), "",
				"### 返回信息", "",
				returnNote(cmd), "",
			)
		}
		must(os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644))
	}
}

func returnNote(cmd *command) string {
	lines := []string{
		"- 成功：退出码 `0`；输出通道：" + outputChannel(cmd) + "。",
		"- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。",
		"- JSON：" + jsonSupport(cmd) + "。",
		"- JSON 返回形态：未在命令声明中定义固定 schema；需以对应版本 handler/formatter 实际输出为准。",
	}
	if cmd.Component == "admin socket" {
		lines = append(lines, "- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。")
	}
	if cmd.Component == "ceph tell" {
		lines = append(lines, "- 调用语义：`ceph tell` 通过 monitor 将请求转发给目标 daemon；返回内容与对应 admin socket 命令一致。")
	}
	return strings.Join(lines, "\n")
}

func outputChannel(cmd *command) string {
	p := cmd.Prefix
	if strings.Contains(p, " getmap") || strings.Contains(p, " getcrushmap") || strings.HasSuffix(p, " getmap") || strings.HasSuffix(p, " getcrushmap") {
		return "stdout 或 `-o <file>`; map 类命令通常输出二进制 map 数据，JSON 不适用"
	}
	if strings.HasPrefix(p, "auth ") && (strings.Contains(p, "get-key") || strings.Contains(p, "print-key") || strings.Contains(p, "print_key")) {
		return "stdout; secret key 文本"
	}
	if strings.HasPrefix(p, "auth ") && (strings.Contains(p, "get") || strings.Contains(p, "export")) {
		return "stdout 或 `-o <file>`; 默认 keyring 文本"
	}
	if cmd.Component == "admin socket" {
		return "stdout; 直接连接本机 daemon admin socket"
	}
	if cmd.Component == "ceph tell" {
		return "stdout; monitor 将请求转发给目标 daemon admin socket"
	}
	return "stdout; 错误信息通常在 stderr"
}

func jsonSupport(cmd *command) string {
	p := cmd.Prefix
	if strings.Contains(p, " getmap") || strings.Contains(p, " getcrushmap") || strings.HasSuffix(p, " getmap") || strings.HasSuffix(p, " getcrushmap") {
		return "否，通常为二进制 map 输出"
	}
	if strings.HasPrefix(p, "auth ") && (strings.Contains(p, "get-key") || strings.Contains(p, "print-key") || strings.Contains(p, "print_key")) {
		return "否，返回 key 文本"
	}
	if cmd.Component == "admin socket" || cmd.Component == "ceph tell" {
		return "部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`"
	}
	first := strings.SplitN(p, " ", 2)[0]
	readish := map[string]bool{"df": true, "health": true, "status": true, "report": true, "versions": true, "osd": true, "pg": true, "fs": true, "mon": true, "mds": true, "mgr": true, "device": true, "orch": true, "rbd": true, "nfs": true, "rgw": true, "smb": true, "telemetry": true}
	if cmd.Perm == "r" || readish[first] {
		return "通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外"
	}
	return "通常不需要；写操作多返回确认文本或空 stdout"
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func unique(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := []string{values[0]}
	for _, value := range values[1:] {
		if value != out[len(out)-1] {
			out = append(out, value)
		}
	}
	return out
}
