#!/usr/bin/env python3
"""Generate Ceph command reference docs from vendored Ceph source tags."""

from __future__ import annotations

import ast
import json
import re
import shutil
import subprocess
import warnings
from collections import defaultdict
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any


ROOT = Path(__file__).resolve().parents[1]
CEPH_REPO = ROOT / "docs/references/ceph"
OUT_DIR = ROOT / "docs/ceph/cmds"
VERSIONS = ["v16.2.15", "v17.2.9", "v18.2.8", "v19.2.5", "v20.2.2"]
VERSION_LABELS = [v[1:] for v in VERSIONS]
warnings.filterwarnings("ignore", category=SyntaxWarning)
SOURCE_CACHE: dict[tuple[str, str], str] = {}
SCHEMA_VERIFY_CACHE: dict[tuple[str, str], bool] = {}

ARG_TYPES = {
    "str": "CephString",
    "int": "CephInt",
    "float": "CephFloat",
    "bool": "CephBool",
    "Optional[str]": "CephString",
    "Optional[int]": "CephInt",
    "Optional[float]": "CephFloat",
    "Optional[bool]": "CephBool",
    "List[str]": "CephString",
    "Sequence[str]": "CephString",
}

FLAG_VALUES = {
    "NOFORWARD": 1 << 0,
    "OBSOLETE": 1 << 1,
    "DEPRECATED": 1 << 2,
    "MGR": 1 << 3,
    "POLL": 1 << 4,
    "HIDDEN": 1 << 5,
    "TELL": (1 << 0) | (1 << 5),
}

KNOWN_HANDLER_ARGS = {"_", "self", "mgr", "inbuf", "return"}


@dataclass
class Param:
    name: str
    type: str = "CephString"
    required: bool = True
    repeated: bool = False
    positional: bool = True
    choices: list[str] = field(default_factory=list)
    range: str = ""
    goodchars: str = ""


@dataclass
class Command:
    prefix: str
    desc: str
    component: str
    module: str
    perm: str
    source: str
    versions: set[str] = field(default_factory=set)
    params_by_version: dict[str, list[Param]] = field(default_factory=dict)
    flags_by_version: dict[str, list[str]] = field(default_factory=dict)

    @property
    def group(self) -> str:
        if self.component == "admin socket":
            return f"admin-socket/{self.module}"
        if self.component == "ceph tell":
            return f"tell/{self.module}"
        first = self.prefix.split(" ", 1)[0]
        if first in {"auth", "config", "config-key", "df", "features", "fsid",
                     "health", "log", "node", "quorum_status", "report",
                     "status", "time-sync-status", "versions", "version"}:
            return "cluster"
        if first == "orch":
            return "orchestrator"
        return first


def git_show(tag: str, path: str) -> str:
    cache_key = (tag, path)
    if cache_key in SOURCE_CACHE:
        return SOURCE_CACHE[cache_key]
    value = subprocess.check_output(
        ["git", "-C", str(CEPH_REPO), "show", f"{tag}:{path}"],
        text=True,
        stderr=subprocess.DEVNULL,
    )
    SOURCE_CACHE[cache_key] = value
    return value


def git_files(tag: str, path: str) -> list[str]:
    out = subprocess.check_output(
        ["git", "-C", str(CEPH_REPO), "ls-tree", "-r", "--name-only", tag, path],
        text=True,
    )
    return [line for line in out.splitlines() if line]


def strip_comments(src: str) -> str:
    src = re.sub(r"/\*.*?\*/", "", src, flags=re.S)
    return re.sub(r"//.*", "", src)


def split_macro_args(text: str) -> list[str]:
    args: list[str] = []
    cur: list[str] = []
    depth = 0
    in_str = False
    esc = False
    for ch in text:
        if in_str:
            cur.append(ch)
            if esc:
                esc = False
            elif ch == "\\":
                esc = True
            elif ch == '"':
                in_str = False
            continue
        if ch == '"':
            in_str = True
            cur.append(ch)
        elif ch == "(":
            depth += 1
            cur.append(ch)
        elif ch == ")":
            depth -= 1
            cur.append(ch)
        elif ch == "," and depth == 0:
            args.append("".join(cur).strip())
            cur = []
        else:
            cur.append(ch)
    if cur:
        args.append("".join(cur).strip())
    return args


def c_string(expr: str, constants: dict[str, str]) -> str:
    expr = expr.replace("\\\n", " ")
    for name, value in constants.items():
        expr = re.sub(rf"\b{name}\b", json.dumps(value), expr)
    strings = re.findall(r'"(?:\\.|[^"\\])*"', expr)
    return "".join(ast.literal_eval(s) for s in strings)


def flags(expr: str) -> list[str]:
    names = re.findall(r"\b[A-Z_]+\b", expr)
    found: list[str] = []
    value = 0
    for name in names:
        if name in FLAG_VALUES:
            value |= FLAG_VALUES[name]
    for name, bit in FLAG_VALUES.items():
        if name == "TELL":
            continue
        if value & bit:
            found.append(name.lower())
    return sorted(set(found))


def parse_signature(sig: str) -> tuple[str, list[Param]]:
    positional = True
    prefix: list[str] = []
    params: list[Param] = []
    for token in sig.split():
        if token == "--":
            positional = False
            continue
        if "=" not in token:
            prefix.append(token)
            continue
        pairs = {}
        for part in token.split(","):
            if "=" in part:
                k, v = part.split("=", 1)
                pairs[k] = v
        name = pairs.get("name", "arg")
        typ = pairs.get("type", "CephString")
        if typ == "CephBool":
            positional = False
        params.append(Param(
            name=name,
            type=typ,
            required=pairs.get("req", "true") != "false",
            repeated=pairs.get("n") == "N",
            positional=pairs.get("positional", str(positional).lower()) != "false",
            choices=pairs.get("strings", "").split("|") if pairs.get("strings") else [],
            range=pairs.get("range", "").replace("|", ".."),
            goodchars=pairs.get("goodchars", ""),
        ))
    return " ".join(prefix), params


def parse_cpp_commands(tag: str, path: str, component: str) -> list[tuple[Command, list[Param], list[str]]]:
    try:
        src = git_show(tag, path)
    except subprocess.CalledProcessError:
        return []
    src = strip_comments(src)
    constants = {m.group(1): c_string(m.group(2), {}) for m in re.finditer(r"#define\s+(\w+)\s+(.+)", src)}
    results: list[tuple[Command, list[Param], list[str]]] = []
    for match in re.finditer(r"\b(COMMAND|COMMAND_WITH_FLAG)\s*\(", src):
        name = match.group(1)
        start = match.end()
        depth = 1
        i = start
        in_str = False
        esc = False
        while i < len(src) and depth:
            ch = src[i]
            if in_str:
                if esc:
                    esc = False
                elif ch == "\\":
                    esc = True
                elif ch == '"':
                    in_str = False
            else:
                if ch == '"':
                    in_str = True
                elif ch == "(":
                    depth += 1
                elif ch == ")":
                    depth -= 1
            i += 1
        args = split_macro_args(src[start:i - 1])
        if len(args) < 4:
            continue
        sig = c_string(args[0], constants)
        desc = c_string(args[1], constants)
        module = c_string(args[2], constants)
        perm = c_string(args[3], constants)
        prefix, params = parse_signature(sig)
        fl = flags(args[4]) if name == "COMMAND_WITH_FLAG" and len(args) > 4 else []
        if "hidden" in fl:
            continue
        cmd = Command(prefix, desc, component, module, perm, path)
        results.append((cmd, params, fl))
    return results


def daemon_target_for_path(path: str) -> tuple[str, str] | None:
    if path.startswith("src/common/"):
        return "common", "<daemon>.<id>"
    if path.startswith("src/mds/"):
        return "mds", "mds.<id>"
    if path.startswith("src/osd/"):
        return "osd", "osd.<id>"
    if path.startswith("src/mon/"):
        return "mon", "mon.<id>"
    if path.startswith("src/mgr/"):
        return "mgr", "mgr.<id>"
    if path.startswith("src/client/"):
        return "client", "client.<id>"
    if path.startswith("src/osdc/"):
        return "client", "client.<id>"
    if path.startswith("src/rgw/"):
        return "rgw", "client.rgw.<id>"
    if path.startswith("src/os/bluestore/"):
        return "bluestore", "osd.<id>"
    if path.startswith("src/tools/cephfs_mirror/"):
        return "cephfs-mirror", "cephfs-mirror.<id>"
    if path.startswith("src/tools/rbd_mirror/"):
        return "rbd-mirror", "rbd-mirror.<id>"
    return None


def parse_admin_socket_commands(tag: str) -> list[tuple[Command, list[Param], list[str]]]:
    try:
        grep_out = subprocess.check_output(
            ["git", "-C", str(CEPH_REPO), "grep", "-l", "register_command", tag, "--", "src"],
            text=True,
            stderr=subprocess.DEVNULL,
        )
    except subprocess.CalledProcessError:
        return []
    out: list[tuple[Command, list[Param], list[str]]] = []
    files = [line.split(":", 1)[1] for line in grep_out.splitlines() if ":" in line]
    for path in files:
        if not path.endswith((".cc", ".h")) or "/test/" in path or "/tests/" in path:
            continue
        target = daemon_target_for_path(path)
        if not target:
            continue
        try:
            src = git_show(tag, path)
        except subprocess.CalledProcessError:
            continue
        if "register_command" not in src:
            continue
        src = strip_comments(src)
        constants = {m.group(1): c_string(m.group(2), {}) for m in re.finditer(r"#define\s+(\w+)\s+(.+)", src)}
        for match in re.finditer(r"\bregister_command\s*\(", src):
            start = match.end()
            depth = 1
            i = start
            in_str = False
            esc = False
            while i < len(src) and depth:
                ch = src[i]
                if in_str:
                    if esc:
                        esc = False
                    elif ch == "\\":
                        esc = True
                    elif ch == '"':
                        in_str = False
                else:
                    if ch == '"':
                        in_str = True
                    elif ch == "(":
                        depth += 1
                    elif ch == ")":
                        depth -= 1
                i += 1
            args = split_macro_args(src[start:i - 1])
            if len(args) < 2:
                continue
            sig = c_string(args[0], constants).strip()
            if not sig:
                continue
            desc = c_string(args[2], constants).strip() if len(args) > 2 else ""
            module, daemon_target = target
            inner_prefix, params = parse_signature(sig)
            prefix = f"daemon {daemon_target} {inner_prefix}".strip()
            cmd = Command(prefix, desc, "admin socket", module, "admin-socket", path)
            out.append((cmd, params, ["admin-socket"]))
            tell_prefix = f"tell {daemon_target} {inner_prefix}".strip()
            tell_cmd = Command(tell_prefix, desc, "ceph tell", module, "rw", path)
            out.append((tell_cmd, params, ["tell", "admin-socket"]))
    return out


def eval_node(node: ast.AST, names: dict[str, Any]) -> Any:
    if isinstance(node, ast.Constant):
        return node.value
    if isinstance(node, ast.Name):
        return names.get(node.id, "")
    if isinstance(node, ast.JoinedStr):
        return "".join(str(eval_node(v.value, names) if isinstance(v, ast.FormattedValue) else v.value)
                       for v in node.values)
    if isinstance(node, ast.BinOp) and isinstance(node.op, ast.Add):
        return str(eval_node(node.left, names)) + str(eval_node(node.right, names))
    if isinstance(node, ast.List):
        return [eval_node(e, names) for e in node.elts]
    if isinstance(node, ast.Tuple):
        return tuple(eval_node(e, names) for e in node.elts)
    if isinstance(node, ast.Dict):
        return {eval_node(k, names): eval_node(v, names) for k, v in zip(node.keys, node.values)}
    if isinstance(node, ast.Call) and isinstance(node.func, ast.Name) and node.func.id == "Option":
        return "Option"
    return ""


def ann_to_type(node: ast.AST | None) -> str:
    if node is None:
        return "CephString"
    text = ast.unparse(node).replace("typing.", "")
    text = text.replace("Optional[bool]", "Optional[bool]")
    return ARG_TYPES.get(text, "CephString")


def decorator_info(dec: ast.AST) -> tuple[str, str, bool] | None:
    if not isinstance(dec, ast.Call):
        return None
    target = dec.func
    method = ""
    if isinstance(target, ast.Attribute):
        method = target.attr
    elif isinstance(target, ast.Name):
        method = "Write"
    else:
        return None
    name = ast.unparse(target)
    if "CLICommand" not in name and not name.endswith("Command"):
        return None
    prefix = ""
    if dec.args:
        prefix = str(eval_node(dec.args[0], {}))
    for kw in dec.keywords:
        if kw.arg == "prefix":
            prefix = str(eval_node(kw.value, {}))
    if not prefix:
        return None
    perm = "rw"
    if method == "Read":
        perm = "r"
    elif method == "Write":
        perm = "w"
    for kw in dec.keywords:
        if kw.arg == "perm":
            perm = str(eval_node(kw.value, {}))
    poll = any(kw.arg == "poll" and bool(eval_node(kw.value, {})) for kw in dec.keywords)
    return prefix, perm, poll


def parse_function_params(fn: ast.FunctionDef) -> list[Param]:
    params: list[Param] = []
    defaults = len(fn.args.defaults)
    first_default = len(fn.args.args) - defaults
    positional = True
    for idx, arg in enumerate(fn.args.args):
        if arg.arg in KNOWN_HANDLER_ARGS:
            continue
        if arg.arg == "_end_positional_":
            positional = False
            continue
        typ = ann_to_type(arg.annotation)
        if arg.arg == "format" or typ == "CephBool":
            positional = False
        params.append(Param(
            name=arg.arg,
            type=typ,
            required=idx < first_default,
            repeated=False,
            positional=positional,
        ))
    return params


def parse_python_commands(tag: str) -> list[tuple[Command, list[Param], list[str]]]:
    out: list[tuple[Command, list[Param], list[str]]] = []
    try:
        files = git_files(tag, "src/pybind/mgr")
    except subprocess.CalledProcessError:
        return out
    for path in files:
        if not path.endswith(".py") or "/tests/" in path:
            continue
        try:
            src = git_show(tag, path)
            tree = ast.parse(src)
        except Exception:
            continue
        names: dict[str, Any] = {}
        for node in tree.body:
            if isinstance(node, ast.Assign) and len(node.targets) == 1 and isinstance(node.targets[0], ast.Name):
                names[node.targets[0].id] = eval_node(node.value, names)
        module = path.split("/")[3] if len(path.split("/")) > 3 else "mgr"
        for node in ast.walk(tree):
            if isinstance(node, ast.Assign):
                for target in node.targets:
                    if isinstance(target, ast.Name) and target.id in {"COMMANDS", "SSO_COMMANDS"}:
                        commands = eval_node(node.value, names)
                        if isinstance(commands, list):
                            for entry in commands:
                                if not isinstance(entry, dict):
                                    continue
                                sig = str(entry.get("cmd") or entry.get("prefix") or "").strip()
                                if not sig:
                                    continue
                                prefix, params = parse_signature(sig)
                                desc = str(entry.get("desc") or "")
                                perm = str(entry.get("perm") or "rw")
                                cmd = Command(prefix, desc, "mgr module", module, perm, path)
                                out.append((cmd, params, ["mgr"]))
            if isinstance(node, ast.FunctionDef):
                for dec in node.decorator_list:
                    info = decorator_info(dec)
                    if not info:
                        continue
                    prefix, perm, poll = info
                    if module == "smb" and not prefix.startswith("smb "):
                        prefix = f"smb {prefix}"
                    desc = ast.get_docstring(node) or ""
                    params = parse_function_params(node)
                    cmd = Command(prefix, " ".join(desc.split()), "mgr module", module, perm, path)
                    fl = ["mgr"] + (["poll"] if poll else [])
                    out.append((cmd, params, fl))
    return out


def key_for(cmd: Command) -> tuple[str, str, str]:
    return (cmd.component, cmd.module, cmd.prefix)


def merge_commands() -> dict[tuple[str, str, str], Command]:
    merged: dict[tuple[str, str, str], Command] = {}
    for tag in VERSIONS:
        version = tag[1:]
        extracted: list[tuple[Command, list[Param], list[str]]] = []
        extracted += parse_cpp_commands(tag, "src/mon/MonCommands.h", "ceph mon")
        extracted += parse_cpp_commands(tag, "src/mgr/MgrCommands.h", "ceph mgr")
        extracted += parse_admin_socket_commands(tag)
        extracted += parse_python_commands(tag)
        for cmd, params, fl in extracted:
            k = key_for(cmd)
            if k not in merged:
                merged[k] = cmd
            existing = merged[k]
            existing.versions.add(version)
            existing.params_by_version[version] = params
            existing.flags_by_version[version] = fl
            if not existing.desc and cmd.desc:
                existing.desc = cmd.desc
    return merged


def version_range(versions: set[str]) -> str:
    present = [v for v in VERSION_LABELS if v in versions]
    if present == VERSION_LABELS:
        return "16.2.15 - 20.2.2"
    return ", ".join(present)


def render_params(params: list[Param]) -> str:
    if not params:
        return "无。"
    rows = ["| 参数 | 类型 | 必填 | 位置参数 | 说明 |", "| --- | --- | --- | --- | --- |"]
    for p in params:
        detail = []
        if p.repeated:
            detail.append("可重复")
        if p.choices:
            detail.append("可选值: " + ", ".join(p.choices))
        if p.range:
            detail.append("范围: " + p.range)
        if p.goodchars:
            detail.append("字符集: `" + p.goodchars + "`")
        rows.append(
            f"| `{p.name}` | `{p.type}` | {'是' if p.required else '否'} | "
            f"{'是' if p.positional else '否'} | {'; '.join(detail) or '-'} |"
        )
    return "\n".join(rows)


def logical_prefix(prefix: str) -> str:
    parts = prefix.split()
    if len(parts) >= 3 and parts[0] in {"daemon", "tell"}:
        return " ".join(parts[2:])
    return prefix


def return_shape(prefix: str) -> str:
    daemon_like = prefix.split(maxsplit=1)[0] in {"daemon", "tell"}
    prefix = logical_prefix(prefix)
    daemon_patterns = [
        (r"^status$", "object; 常见字段: name/rank 状态、state、版本、地址、健康/性能摘要；具体字段随 daemon 类型变化"),
        (r"^session ls$", "array/object; MDS client sessions，常见字段: id/num, entity, addr, state, caps, leases, requests, completed_requests, client_metadata"),
        (r"^client ls$", "array/object; MDS client 列表，通常与 session ls 结构接近"),
        (r"^session config$", "object/string; MDS session 配置或操作结果"),
        (r"^session (evict|kill)$", "string/object; 操作确认、被驱逐 client/session 信息或错误说明"),
        (r"^(ops|dump_ops_in_flight|dump_blocked_ops|dump_historic_ops|dump_historic_slow_ops|dump_historic_ops_by_duration)$", "object; daemon op dump，常见字段: ops[]/historic_ops[]，每项包含 description, duration, initiated_at, type_data/events"),
        (r"^dump_blocked_ops_count$", "object/int; blocked op 计数"),
        (r"^perf dump", "object; perf counter 值，按 logger/counter 分组"),
        (r"^perf schema", "object; perf counter schema，包含 type、description、metric_type 等元数据"),
        (r"^config show$", "object; daemon 当前配置键值"),
        (r"^config get$", "string/object; 单个配置值"),
        (r"^config help$", "object; 配置项 schema 和说明"),
        (r"^heap", "object/string; heap profiler/allocator 状态或操作结果"),
        (r"^dump cache", "object; MDS cache dump，常见字段包含 inode/dentry/dirfrag/cache 结构，体量可能很大"),
        (r"^cache status$", "object/string; MDS cache 使用量、limit、trim 状态等摘要"),
        (r"^dump tree", "object; MDS subtree/cache tree 信息"),
        (r"^dump loads", "object; MDS load/subtree load 信息"),
        (r"^dump snaps", "object; snap server 或 snap 信息"),
        (r"^damage ls$", "array/object; MDS damage table 条目"),
        (r"^openfiles ls$", "array/object; MDS 打开文件列表"),
        (r"^get subtrees$", "array/object; MDS subtree 列表和 authority 信息"),
        (r"^smart", "object; device SMART/health metrics"),
        (r"^list_devices$", "array; daemon 管理的 device 列表"),
        (r"^dump_watchers$", "array/object; OSD watcher 列表"),
        (r"^dump_blocklist$", "array/object; OSD blocklist 条目"),
        (r"^dump_scrubs$", "object/array; OSD scrub 状态"),
        (r"^dump_pgstate_history$", "object/array; PG state history"),
    ]
    common_patterns = [
        (r"^status$", "object; 常见字段: fsid, health, election_epoch, quorum, quorum_names, monmap, osdmap, pgmap, mgrmap, fsmap, servicemap"),
        (r"^health( detail)?$", "object; 常见字段: status, checks, mutes; detail 模式会包含每个 health check 的 severity, summary, detail"),
        (r"^df$", "object; 常见字段: stats, stats_by_class, pools[]; pool 项通常包含 name, id, stats"),
        (r"^report$", "object; 集群完整报告，通常包含 maps、service 状态、配置摘要、crush/osd/pg/fs/mon/mgr 信息"),
        (r"^versions$", "object; 按 daemon 类型聚合版本，常见键: overall, mon, mgr, osd, mds, rgw"),
        (r"^features$", "object; 连接特性摘要，包含 client/mon/osd 等 feature 集合或计数"),
        (r"^quorum_status$", "object; 常见字段: election_epoch, quorum, quorum_names, quorum_leader_name, monmap"),
        (r"^time-sync-status$", "object; monitor 时间同步状态，包含 skew/latency/health 等信息"),
        (r"^mon dump", "object; 常见字段: epoch, fsid, created, modified, min_mon_release, mons[]"),
        (r"^mon stat$", "object/string; monitor quorum 摘要；JSON 模式通常包含 monmap/quorum 摘要"),
        (r"^mon metadata", "object/array; daemon metadata，单 daemon 为 object，多 daemon 为 array/object 集合"),
        (r"^mon versions$", "object; monitor 版本聚合"),
        (r"^mgr dump", "object; 常见字段: epoch, active_name, active_addr, available, standbys, modules, services"),
        (r"^mgr services$", "object; service name 到 URL/endpoint 的映射"),
        (r"^mgr module ls$", "object; enabled/disabled/always_on 等模块列表"),
        (r"^mgr metadata", "object/array; manager daemon metadata"),
        (r"^mgr versions$", "object; mgr 版本聚合"),
        (r"^osd stat$", "object; 常见字段: epoch, num_osds, num_up_osds, num_in_osds, osdmap_first_committed, osdmap_last_committed"),
        (r"^osd dump", "object; OSDMap dump，常见字段: epoch, fsid, flags, pools[], osds[], pg_upmap, pg_temp"),
        (r"^osd tree", "object; 常见字段: nodes[], stray[], summary; nodes 含 id, name, type, status, reweight, children"),
        (r"^osd df", "object; 常见字段: nodes[], summary; nodes 含 id/name/class/weight/reweight/kb/kb_used/utilization/pgs"),
        (r"^osd perf$", "object; OSD 性能摘要，常见字段含 osd id、commit/apply latency"),
        (r"^osd pool stats", "array/object; pool 统计，常见字段: pool_name, pool_id, recovery, client_io_rate"),
        (r"^osd pool ls", "array; pool 名称或 pool 详情，detail 模式包含 id/name/pg_num/size/min_size 等"),
        (r"^osd pool get", "object; 指定 pool 属性名到值的映射"),
        (r"^osd pool get-quota", "object; 常见字段: pool_name, quota_max_objects, quota_max_bytes"),
        (r"^osd crush dump$", "object; CRUSH map，常见字段: devices, types, buckets, rules, tunables, choose_args"),
        (r"^osd crush tree", "object; CRUSH 树，常见字段: nodes[], shadow_trees[]"),
        (r"^osd crush rule (ls|dump)", "array/object; CRUSH rule 列表或单条 rule 详情"),
        (r"^osd metadata", "object/array; OSD metadata，包含 hostname、device、os、kernel、ceph_version 等"),
        (r"^osd versions$", "object; OSD 版本聚合"),
        (r"^pg stat$", "object/string; PG 状态摘要，JSON 通常包含 pgmap 或 pg 统计字段"),
        (r"^pg dump", "object; PG map dump，按 dumpcontents 可能包含 summary, pools, osds, pgs, pgs_brief"),
        (r"^pg ls", "array; PG 条目列表，字段随版本和筛选项变化"),
        (r"^pg map", "object/string; 指定 PG 的 acting/up/primary 映射信息"),
        (r"^fs dump", "object; FSMap dump，常见字段: epoch, filesystems[], standbys"),
        (r"^fs ls$", "array; CephFS 文件系统列表，常见字段: name, metadata_pool, metadata_pool_id, data_pool_ids"),
        (r"^fs get", "object; 单个 CephFS 的 mdsmap/filesystem 详情"),
        (r"^fs status$", "object; CephFS 状态，常见字段: mdsmap, pools, clients, ranks"),
        (r"^fs volume ls", "array; volume 列表，常见字段: name"),
        (r"^fs volume info", "object; volume 信息，常见字段: mon_addrs, pools, metadata_pool, data_pool_ids"),
        (r"^mds metadata", "object/array; MDS metadata，包含 hostname、addr、ceph_version 等"),
        (r"^mds versions$", "object; MDS 版本聚合"),
        (r"^auth (ls|list)$", "object/array; auth entity 列表，常见字段: entity, key, caps"),
        (r"^auth (get|export)$", "keyring text/object; 默认输出 keyring 文本；JSON 支持取决于对应版本 formatter"),
        (r"^auth (get-key|print-key|print_key)$", "string; secret key 文本"),
        (r"^config dump", "array; 配置项列表，常见字段: name, value, section, level, source"),
        (r"^config get", "string/object; 单个配置值"),
        (r"^config show", "object; daemon 当前配置键值"),
        (r"^config-key (dump|ls)$", "object/array; config-key 键或键值列表"),
        (r"^device ls$", "array; device 列表，常见字段: devid, daemons, location, life_expectancy"),
        (r"^device info", "object; 单个 device 详情"),
        (r"^device get-health-metrics", "object; device health metrics 时间序列/指标集合"),
        (r"^orch (host|device|ls|ps|status|upgrade|tuned-profile|certmgr)", "object/array; orchestrator 返回结构由后端模块决定，常见为服务/daemon/host/device 列表或 completion 状态"),
        (r"^rbd .*", "object/array/string; RBD mgr 模块返回结构由 rbd_support handler 决定，列表类命令通常为 array/object"),
        (r"^rgw .*", "object/array/string; RGW mgr 模块返回结构由 rgw handler 或 radosgw-admin 输出决定"),
        (r"^nfs .*", "object/array/string; NFS 模块返回 export/cluster 配置或操作结果"),
        (r"^smb .*", "object/array/string; SMB 模块通过 object_format 输出资源对象、列表或操作结果"),
        (r"^telemetry (show|preview|show-all|preview-all)", "object; telemetry report payload"),
        (r"^telemetry (status|channel ls|collection ls)", "object/array; telemetry 状态、channel 或 collection 列表"),
    ]
    patterns = daemon_patterns if daemon_like else common_patterns
    for pattern, shape in patterns:
        if re.match(pattern, prefix):
            return shape
    return "未在命令声明中定义固定 schema；需以对应版本 handler/formatter 实际输出为准"


def output_channel(cmd: Command) -> str:
    p = cmd.prefix
    if any(token in p for token in [" getmap", " getcrushmap"]) or p.endswith(" getmap") or p.endswith(" getcrushmap"):
        return "stdout 或 `-o <file>`; map 类命令通常输出二进制 map 数据，JSON 不适用"
    if p.startswith("auth ") and any(x in p for x in ["get-key", "print-key", "print_key"]):
        return "stdout; secret key 文本"
    if p.startswith("auth ") and any(x in p for x in ["get", "export"]):
        return "stdout 或 `-o <file>`; 默认 keyring 文本"
    if cmd.component == "admin socket":
        return "stdout; 直接连接本机 daemon admin socket"
    if cmd.component == "ceph tell":
        return "stdout; monitor 将请求转发给目标 daemon admin socket"
    return "stdout; 错误信息通常在 stderr"


def json_support(cmd: Command) -> str:
    p = cmd.prefix
    if any(token in p for token in [" getmap", " getcrushmap"]) or p.endswith(" getmap") or p.endswith(" getcrushmap"):
        return "否，通常为二进制 map 输出"
    if p.startswith("auth ") and any(x in p for x in ["get-key", "print-key", "print_key"]):
        return "否，返回 key 文本"
    if cmd.component in {"admin socket", "ceph tell"}:
        return "部分支持；命令 handler 使用 formatter 时可追加 `--format json`/`--format json-pretty`"
    if cmd.perm == "r" or p.split()[0] in {"df", "health", "status", "report", "versions", "osd", "pg", "fs", "mon", "mds", "mgr", "device", "orch", "rbd", "nfs", "rgw", "smb", "telemetry"}:
        return "通常支持 `--format json` 或 `--format json-pretty`；少数文本/二进制命令除外"
    return "通常不需要；写操作多返回确认文本或空 stdout"


def field_table(rows: list[tuple[str, str, str]]) -> str:
    lines = ["| 字段 | 类型 | 说明 |", "| --- | --- | --- |"]
    for name, typ, desc in rows:
        lines.append(f"| `{name}` | `{typ}` | {desc} |")
    return "\n".join(lines)


def return_fields(prefix: str) -> str:
    daemon_like = prefix.split(maxsplit=1)[0] in {"daemon", "tell"}
    p = logical_prefix(prefix)
    daemon_specs: list[tuple[str, list[tuple[str, str, str]]]] = [
        (r"^session ls$|^client ls$", [
            ("[]", "array/object", "MDS client session 列表；不同版本可能直接返回数组或包装对象"),
            ("[].id / [].num", "integer", "client/session 编号"),
            ("[].entity", "string", "客户端 entity，例如 `client.1234`"),
            ("[].addr", "string", "客户端连接地址"),
            ("[].state", "string", "session 状态"),
            ("[].caps", "integer/object", "客户端持有的 capability 数量或 capability dump"),
            ("[].leases", "integer/object", "lease 数量或 lease dump"),
            ("[].requests", "integer/array", "正在处理的请求数量或请求列表"),
            ("[].completed_requests", "integer", "已完成请求数量"),
            ("[].client_metadata", "object", "客户端 metadata，例如 hostname、kernel_version、ceph_version 等"),
        ]),
        (r"^(ops|dump_ops_in_flight|dump_blocked_ops|dump_historic_ops|dump_historic_slow_ops|dump_historic_ops_by_duration)$", [
            ("ops[] / historic_ops[]", "array", "op 条目列表，具体数组名随命令变化"),
            ("[].description", "string", "op 描述"),
            ("[].initiated_at", "string/timestamp", "op 发起时间"),
            ("[].age / duration", "number/string", "op 已运行时长"),
            ("[].type_data", "object", "op 类型相关字段"),
            ("[].events[]", "array", "op 事件时间线"),
        ]),
        (r"^dump_blocked_ops_count$", [("count", "integer", "blocked op 数量")]),
        (r"^status$", [
            ("name", "string", "daemon 名称或 rank 名称"),
            ("state", "string", "daemon 当前状态"),
            ("rank", "integer", "MDS rank；非 MDS daemon 可能不存在"),
            ("version", "string", "Ceph 版本"),
            ("addrs", "object/string", "daemon 地址"),
        ]),
        (r"^perf dump", [
            ("<logger>", "object", "perf logger 名称"),
            ("<logger>.<counter>", "integer/number/object", "counter 当前值；直方图 counter 可能为对象"),
        ]),
        (r"^perf schema", [
            ("<logger>.<counter>.type", "integer/string", "counter 类型"),
            ("<logger>.<counter>.description", "string", "counter 说明"),
            ("<logger>.<counter>.metric_type", "string", "指标类型"),
            ("<logger>.<counter>.value_type", "string", "值类型"),
        ]),
        (r"^config show$", [("<option>", "string/number/bool", "daemon 当前配置键值")]),
        (r"^config get$", [("value", "string", "请求的单个配置值，部分版本直接返回裸字符串")]),
        (r"^config help$", [
            ("name", "string", "配置项名称"),
            ("type", "string", "配置类型"),
            ("level", "string", "配置级别"),
            ("desc", "string", "配置说明"),
            ("default", "any", "默认值"),
        ]),
        (r"^cache status$", [
            ("num_inodes", "integer", "MDS cache inode 数"),
            ("num_caps", "integer", "capability 数"),
            ("max_size", "integer", "cache size 限制"),
            ("trim_queue_len", "integer", "trim 队列长度"),
        ]),
        (r"^smart$", [
            ("devid", "string", "设备 ID"),
            ("ata_smart_attributes / nvme_smart_health_information_log", "object", "SMART 指标，字段取决于设备类型"),
        ]),
        (r"^list_devices$", [
            ("[]", "array", "daemon 关联设备列表"),
            ("[].devid", "string", "设备 ID"),
            ("[].dev", "string", "设备名"),
            ("[].path", "string", "设备路径"),
        ]),
    ]
    common_specs: list[tuple[str, list[tuple[str, str, str]]]] = [
        (r"^status$", [
            ("fsid", "string", "集群 FSID"),
            ("health", "object", "健康状态对象"),
            ("health.status", "string", "整体健康状态，例如 `HEALTH_OK/WARN/ERR`"),
            ("health.checks", "object", "按 health check 名称索引的检查结果"),
            ("election_epoch", "integer", "monitor election epoch"),
            ("quorum", "array[integer]", "quorum monitor rank 列表"),
            ("quorum_names", "array[string]", "quorum monitor 名称列表"),
            ("monmap", "object", "monitor map 摘要"),
            ("osdmap", "object", "OSD map 摘要"),
            ("pgmap", "object", "PG map 统计摘要"),
            ("mgrmap", "object", "manager map 摘要"),
            ("fsmap", "object", "CephFS map 摘要"),
            ("servicemap", "object", "服务 map 摘要"),
        ]),
        (r"^health( detail)?$", [
            ("status", "string", "整体健康状态"),
            ("checks", "object", "按检查项名称索引的详细检查结果"),
            ("checks.<name>.severity", "string", "检查项严重级别"),
            ("checks.<name>.summary.message", "string", "摘要信息"),
            ("checks.<name>.summary.count", "integer", "摘要计数"),
            ("checks.<name>.detail[]", "array", "detail 模式下的详细条目"),
            ("mutes", "array/object", "已 mute 的健康检查"),
        ]),
        (r"^df$", [
            ("stats.total_bytes", "integer", "集群总容量"),
            ("stats.total_used_bytes", "integer", "已用容量"),
            ("stats.total_avail_bytes", "integer", "可用容量"),
            ("stats.num_objects", "integer", "对象数量"),
            ("stats_by_class", "object", "按 device class 聚合的容量统计"),
            ("pools[]", "array", "pool 容量统计列表"),
            ("pools[].name", "string", "pool 名称"),
            ("pools[].id", "integer", "pool ID"),
            ("pools[].stats.bytes_used", "integer", "pool 已用字节"),
            ("pools[].stats.max_avail", "integer", "pool 可用容量估算"),
            ("pools[].stats.objects", "integer", "pool 对象数"),
        ]),
        (r"^report$", [
            ("health", "object", "健康状态"),
            ("monmap / osdmap / pgmap / mgrmap / fsmap", "object", "各核心 map"),
            ("crushmap", "object", "CRUSH map"),
            ("config", "object/array", "配置摘要"),
            ("metadata", "object", "daemon metadata 聚合"),
        ]),
        (r"^versions$", [
            ("overall", "object", "全 daemon 版本分布"),
            ("mon / mgr / osd / mds / rgw", "object", "按 daemon 类型统计版本分布"),
        ]),
        (r"^quorum_status$", [
            ("election_epoch", "integer", "election epoch"),
            ("quorum", "array[integer]", "quorum rank 列表"),
            ("quorum_names", "array[string]", "quorum 名称列表"),
            ("quorum_leader_name", "string", "quorum leader 名称"),
            ("monmap", "object", "monitor map"),
        ]),
        (r"^mon dump", [
            ("epoch", "integer", "monmap epoch"),
            ("fsid", "string", "集群 FSID"),
            ("created / modified", "string", "创建/修改时间"),
            ("min_mon_release", "integer/string", "最小 monitor release"),
            ("mons[]", "array", "monitor 列表"),
            ("mons[].rank", "integer", "monitor rank"),
            ("mons[].name", "string", "monitor 名称"),
            ("mons[].addr / public_addrs", "string/object", "monitor 地址"),
        ]),
        (r"^mgr dump", [
            ("epoch", "integer", "mgrmap epoch"),
            ("active_name", "string", "active mgr 名称"),
            ("active_addr", "string", "active mgr 地址"),
            ("available", "boolean", "是否有可用 mgr"),
            ("standbys[]", "array", "standby mgr 列表"),
            ("modules", "array/object", "mgr module 状态"),
            ("services", "object", "mgr 暴露服务地址"),
        ]),
        (r"^osd stat$", [
            ("epoch", "integer", "OSDMap epoch"),
            ("num_osds", "integer", "OSD 总数"),
            ("num_up_osds", "integer", "up OSD 数"),
            ("num_in_osds", "integer", "in OSD 数"),
            ("osdmap_first_committed", "integer", "最早 committed epoch"),
            ("osdmap_last_committed", "integer", "最新 committed epoch"),
        ]),
        (r"^osd dump", [
            ("epoch", "integer", "OSDMap epoch"),
            ("fsid", "string", "集群 FSID"),
            ("created / modified", "string", "OSDMap 时间戳"),
            ("flags", "string/array", "OSDMap flags"),
            ("pools[]", "array", "pool 定义列表"),
            ("pools[].pool", "integer", "pool ID"),
            ("pools[].pool_name", "string", "pool 名称"),
            ("pools[].type", "integer/string", "pool 类型"),
            ("pools[].size / min_size", "integer", "副本/最小副本数"),
            ("pools[].pg_num / pgp_num", "integer", "PG/PGP 数"),
            ("osds[]", "array", "OSD 条目列表"),
            ("osds[].osd", "integer", "OSD ID"),
            ("osds[].up / in", "integer", "up/in 状态"),
            ("osds[].weight", "number", "CRUSH weight"),
            ("osds[].public_addr / cluster_addr", "string", "OSD 地址"),
            ("pg_upmap / pg_temp / primary_temp", "array/object", "PG 临时映射覆盖"),
        ]),
        (r"^osd tree", [
            ("nodes[]", "array", "CRUSH/OSD 树节点"),
            ("nodes[].id", "integer", "节点 ID"),
            ("nodes[].name", "string", "节点名称"),
            ("nodes[].type", "string", "节点类型，如 root/host/osd"),
            ("nodes[].type_id", "integer", "节点类型 ID"),
            ("nodes[].children", "array[integer]", "子节点 ID"),
            ("nodes[].status", "string", "OSD 状态，非 OSD 节点可能不存在"),
            ("nodes[].reweight", "number", "OSD reweight"),
            ("stray[]", "array", "不在 CRUSH 树中的 OSD"),
        ]),
        (r"^osd df", [
            ("nodes[]", "array", "OSD/CRUSH 节点利用率"),
            ("nodes[].id / name / type", "integer/string", "节点标识"),
            ("nodes[].device_class", "string", "设备类别"),
            ("nodes[].kb / kb_used / kb_avail", "integer", "容量 KiB"),
            ("nodes[].utilization", "number", "利用率百分比"),
            ("nodes[].var", "number", "相对平均利用率"),
            ("nodes[].pgs", "integer", "PG 数"),
            ("summary", "object", "汇总统计"),
        ]),
        (r"^osd perf$", [
            ("osd_perf_infos[]", "array", "OSD 性能条目"),
            ("osd_perf_infos[].id", "integer", "OSD ID"),
            ("osd_perf_infos[].perf_stats.commit_latency_ms", "number", "commit 延迟"),
            ("osd_perf_infos[].perf_stats.apply_latency_ms", "number", "apply 延迟"),
        ]),
        (r"^osd pool stats", [
            ("[]", "array", "pool 统计列表"),
            ("[].pool_name", "string", "pool 名称"),
            ("[].pool_id", "integer", "pool ID"),
            ("[].recovery", "object", "恢复状态/速率"),
            ("[].client_io_rate", "object", "客户端 IO 速率"),
        ]),
        (r"^osd pool ls", [
            ("[]", "array", "pool 名称列表或 detail 对象列表"),
            ("[].poolnum / [].pool", "integer", "pool ID，detail 模式"),
            ("[].poolname / [].pool_name", "string", "pool 名称，detail 模式"),
            ("[].pg_num / [].pgp_num", "integer", "PG/PGP 数，detail 模式"),
            ("[].size / [].min_size", "integer", "副本/最小副本数，detail 模式"),
        ]),
        (r"^osd crush dump$", [
            ("devices[]", "array", "CRUSH device 列表"),
            ("types[]", "array", "bucket 类型"),
            ("buckets[]", "array", "bucket 列表"),
            ("rules[]", "array", "CRUSH rule 列表"),
            ("tunables", "object", "CRUSH tunables"),
            ("choose_args", "object/array", "choose 参数"),
        ]),
        (r"^pg dump", [
            ("pg_stats[] / pgs[]", "array", "PG 统计列表"),
            ("pool_stats[]", "array", "pool 级 PG 统计"),
            ("osd_stats[]", "array", "OSD 级 PG 统计"),
            ("summary", "object", "PGMap 摘要"),
            ("stamp", "string", "统计时间戳"),
            ("version", "integer", "PGMap 版本"),
        ]),
        (r"^fs dump", [
            ("epoch", "integer", "FSMap epoch"),
            ("filesystems[]", "array", "CephFS 文件系统列表"),
            ("filesystems[].id", "integer", "文件系统 ID"),
            ("filesystems[].mdsmap", "object", "MDSMap"),
            ("filesystems[].mdsmap.fs_name", "string", "文件系统名称"),
            ("filesystems[].mdsmap.metadata_pool", "integer", "metadata pool ID"),
            ("filesystems[].mdsmap.data_pools", "array[integer]", "data pool IDs"),
            ("standbys[]", "array", "standby MDS 列表"),
        ]),
        (r"^fs ls$", [
            ("[]", "array", "文件系统列表"),
            ("[].name", "string", "文件系统名称"),
            ("[].metadata_pool", "string", "metadata pool 名称"),
            ("[].metadata_pool_id", "integer", "metadata pool ID"),
            ("[].data_pools[]", "array[string]", "data pool 名称列表"),
            ("[].data_pool_ids[]", "array[integer]", "data pool ID 列表"),
        ]),
        (r"^fs status$", [
            ("mdsmap", "object", "MDS 状态"),
            ("pools[]", "array", "CephFS 使用的 pool"),
            ("clients", "object/integer", "客户端数量/详情"),
            ("ranks[]", "array", "MDS rank 状态"),
        ]),
        (r"^auth (ls|list)$", [
            ("auth_dump[]", "array", "auth entity 列表"),
            ("auth_dump[].entity", "string", "entity 名称"),
            ("auth_dump[].key", "string", "secret key"),
            ("auth_dump[].caps", "object", "按 mon/osd/mds/mgr 等子系统分组的 caps"),
        ]),
        (r"^config dump", [
            ("[]", "array", "配置项列表"),
            ("[].name", "string", "配置项名称"),
            ("[].value", "string", "配置值"),
            ("[].section", "string", "作用域/daemon section"),
            ("[].level", "string", "配置级别"),
            ("[].source", "string", "配置来源"),
        ]),
        (r"^device ls$", [
            ("devices[] / []", "array", "device 列表"),
            ("[].devid", "string", "设备 ID"),
            ("[].daemons[]", "array[string]", "关联 daemon"),
            ("[].location[]", "array", "设备位置信息"),
            ("[].life_expectancy_min/max", "string/timestamp", "寿命预测范围"),
            ("[].wear_level", "number", "磨损程度"),
        ]),
    ]
    specs = daemon_specs if daemon_like else common_specs
    for pattern, rows in specs:
        if re.match(pattern, p):
            return field_table(rows)
    return "未能从命令声明静态确定详细字段；需要结合对应版本 handler/formatter 源码或运行 `--format json` 样本继续补全。"


VERIFIED_RETURN_SCHEMAS = [
    {
        "name": "ceph status",
        "component": "ceph mon",
        "module": "mon",
        "prefix": "status",
        "shape": "object; 字段由 `Monitor::handle_command(status)` 的 formatter 输出确认",
        "checks": [
            ("src/mon/Monitor.cc", [
                r'dump_stream\("fsid"\)',
                r'get_health_status\(false, f, nullptr\)',
                r'dump_unsigned\("election_epoch"',
                r'open_array_section\("quorum"\)',
                r'open_array_section\("quorum_names"\)',
                r'dump_int\(\s*"quorum_age"',
                r'open_object_section\("monmap"\)',
                r'open_object_section\("osdmap"\)',
                r'open_object_section\("pgmap"\)',
                r'open_object_section\("fsmap"\)',
                r'open_object_section\("mgrmap"\)',
                r'dump_object\("servicemap"',
                r'open_object_section\("progress_events"\)',
            ]),
        ],
        "fields": [
            ("fsid", "string", "集群 FSID", "`src/mon/Monitor.cc`"),
            ("health", "object", "`healthmon()->get_health_status(false, f, nullptr)` 输出的健康对象", "`src/mon/Monitor.cc`, `src/mon/HealthMonitor.cc`"),
            ("election_epoch", "integer", "monitor election epoch", "`src/mon/Monitor.cc`"),
            ("quorum[]", "array[integer]", "quorum monitor rank 列表；元素字段名为 `rank`", "`src/mon/Monitor.cc`"),
            ("quorum_names[]", "array[string]", "quorum monitor 名称列表；元素字段名为 `id`", "`src/mon/Monitor.cc`"),
            ("quorum_age", "integer", "当前 quorum 持续时间", "`src/mon/Monitor.cc`"),
            ("monmap", "object", "`MonMap::dump_summary` 输出", "`src/mon/Monitor.cc`"),
            ("osdmap", "object", "`OSDMap::print_summary` 输出", "`src/mon/Monitor.cc`"),
            ("pgmap", "object", "`MgrStatMonitor::print_summary` 输出", "`src/mon/Monitor.cc`"),
            ("fsmap", "object", "`FSMap::print_summary` 输出", "`src/mon/Monitor.cc`"),
            ("mgrmap", "object", "`MgrMap::print_summary` 输出", "`src/mon/Monitor.cc`"),
            ("servicemap", "object", "`mgrstatmon()->get_service_map()` 输出", "`src/mon/Monitor.cc`"),
            ("progress_events", "object", "mgr progress event map", "`src/mon/Monitor.cc`"),
        ],
    },
    {
        "name": "ceph health",
        "component": "ceph mon",
        "module": "mon",
        "prefix": "health",
        "shape": "object; 字段由 `HealthMonitor::get_health_status` 和 `health_check_t::dump` 输出确认",
        "checks": [
            ("src/mon/HealthMonitor.cc", [
                r'open_object_section\("health"\)',
                r'dump_stream\("status"\)',
                r'open_object_section\("checks"\)',
                r'dump_bool\("muted"',
                r'open_array_section\("mutes"\)',
            ]),
            ("src/mon/health_check.h", [
                r'dump_stream\("severity"\)',
                r'open_object_section\("summary"\)',
                r'dump_string\("message", summary\)',
                r'dump_int\("count", count\)',
                r'open_array_section\("detail"\)',
                r'dump_string\("message", p\)',
                r'dump_string\("code", code\)',
                r'dump_bool\("sticky", sticky\)',
            ]),
        ],
        "fields": [
            ("status", "string", "整体健康状态", "`src/mon/HealthMonitor.cc`"),
            ("checks", "object", "按 health check code 索引的检查结果对象", "`src/mon/HealthMonitor.cc`"),
            ("checks.<code>.severity", "string", "检查项严重级别", "`src/mon/health_check.h`"),
            ("checks.<code>.summary.message", "string", "检查项摘要文本", "`src/mon/health_check.h`"),
            ("checks.<code>.summary.count", "integer", "检查项计数", "`src/mon/health_check.h`"),
            ("checks.<code>.detail[].message", "string", "`detail` 模式下的详细条目文本", "`src/mon/health_check.h`"),
            ("checks.<code>.muted", "boolean", "该检查项是否被 mute", "`src/mon/HealthMonitor.cc`"),
            ("mutes[]", "array[object]", "mute 记录列表", "`src/mon/HealthMonitor.cc`"),
            ("mutes[].code", "string", "被 mute 的 health check code", "`src/mon/health_check.h`"),
            ("mutes[].ttl", "string", "mute 过期时间；仅 ttl 非空时输出", "`src/mon/health_check.h`"),
            ("mutes[].sticky", "boolean", "是否 sticky mute", "`src/mon/health_check.h`"),
            ("mutes[].summary", "string", "mute 摘要", "`src/mon/health_check.h`"),
            ("mutes[].count", "integer", "mute 计数", "`src/mon/health_check.h`"),
        ],
    },
    {
        "name": "mds session ls",
        "component": "*",
        "module": "mds",
        "prefix": "session ls",
        "shape": "object; `sessions` 数组中的 `session` 对象字段由 `MDSRankDispatcher::dump_sessions` 与 `Session::dump` 输出确认",
        "checks": [
            ("src/mds/MDSRank.cc", [
                r'open_array_section\("sessions"\)',
                r'open_object_section\("session"\)',
                r's->dump\(f, cap_dump\)',
            ]),
            ("src/mds/SessionMap.cc", [
                r'dump_int\("id"',
                r'dump_object\("entity"',
                r'dump_string\("state"',
                r'dump_int\("num_leases"',
                r'dump_int\("num_caps"',
                r'open_array_section\("caps"\)',
                r'dump_unsigned\("request_load_avg"',
                r'dump_float\("uptime"',
                r'dump_unsigned\("requests_in_flight"',
                r'dump_unsigned\("num_completed_requests"',
                r'dump_unsigned\("num_completed_flushes"',
                r'dump_bool\("reconnecting"',
                r'dump_object\("recall_caps"',
                r'dump_object\("release_caps"',
                r'dump_object\("recall_caps_throttle"',
                r'dump_object\("recall_caps_throttle2o"',
                r'dump_object\("session_cache_liveness"',
                r'dump_object\("cap_acquisition"',
                r'dump_unsigned\("last_trim_completed_requests_tid"',
                r'dump_unsigned\("last_trim_completed_flushes_tid"',
                r'open_array_section\("delegated_inos"\)',
            ]),
        ],
        "fields": [
            ("sessions[]", "array", "session 列表", "`src/mds/MDSRank.cc`"),
            ("sessions[].session.id", "integer", "client/session 编号", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.entity", "object", "客户端 entity 实例", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.state", "string", "session 状态", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.num_leases", "integer", "lease 数量", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.num_caps", "integer", "capability 数量", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.caps[]", "array[object]", "`cap_dump=true` 时输出的 capability 列表", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.request_load_avg", "integer", "open/stale session 的请求负载均值", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.uptime", "float", "session 存活时间", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.requests_in_flight", "integer", "正在处理的请求数", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.num_completed_requests", "integer", "已完成请求数", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.num_completed_flushes", "integer", "已完成 flush 数", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.reconnecting", "boolean", "是否处于 reconnecting", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.recall_caps", "object", "recall caps 状态对象", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.release_caps", "object", "release caps 状态对象", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.recall_caps_throttle", "object", "recall caps throttle 状态", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.recall_caps_throttle2o", "object", "second-order recall caps throttle 状态", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.session_cache_liveness", "object", "session cache liveness 状态", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.cap_acquisition", "object", "cap acquisition 状态", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.last_trim_completed_requests_tid", "integer", "最近 trim completed requests tid", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.last_trim_completed_flushes_tid", "integer", "最近 trim completed flushes tid", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.delegated_inos[]", "array[object]", "delegated inode range 列表", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.delegated_inos[].start", "string", "inode range 起点", "`src/mds/SessionMap.cc`"),
            ("sessions[].session.delegated_inos[].length", "integer", "inode range 长度", "`src/mds/SessionMap.cc`"),
        ],
    },
]


def schema_matches(schema: dict[str, Any], cmd: Command) -> bool:
    prefix = logical_prefix(cmd.prefix)
    if schema["prefix"] != prefix:
        return False
    if schema["module"] != cmd.module:
        return False
    return schema["component"] == "*" or schema["component"] == cmd.component


def source_checks_pass(schema: dict[str, Any], versions: set[str]) -> bool:
    cache_key = (schema["name"], ",".join(sorted(versions)))
    if cache_key in SCHEMA_VERIFY_CACHE:
        return SCHEMA_VERIFY_CACHE[cache_key]
    tags = [f"v{version}" for version in versions]
    for tag in tags:
        for path, patterns in schema["checks"]:
            try:
                src = git_show(tag, path)
            except subprocess.CalledProcessError:
                SCHEMA_VERIFY_CACHE[cache_key] = False
                return False
            for pattern in patterns:
                if not re.search(pattern, src):
                    SCHEMA_VERIFY_CACHE[cache_key] = False
                    return False
    SCHEMA_VERIFY_CACHE[cache_key] = True
    return True


def verified_schema(cmd: Command) -> dict[str, Any] | None:
    for schema in VERIFIED_RETURN_SCHEMAS:
        if schema_matches(schema, cmd) and source_checks_pass(schema, cmd.versions):
            return schema
    return None


def return_shape_verified(cmd: Command) -> str:
    schema = verified_schema(cmd)
    if schema:
        return schema["shape"]
    return "未记录字段；未能在该命令支持的所有版本源码中完成字段校验，避免写入猜测 schema"


def field_table_verified(rows: list[tuple[str, str, str, str]]) -> str:
    lines = ["| 字段 | 类型 | 说明 | 源码确认 |", "| --- | --- | --- | --- |"]
    for name, typ, desc, source in rows:
        lines.append(f"| `{name}` | `{typ}` | {desc} | {source} |")
    return "\n".join(lines)


def return_fields_verified(cmd: Command) -> str:
    schema = verified_schema(cmd)
    if not schema:
        return "未记录字段；该命令的返回字段尚未在其支持版本源码中逐项校验，避免写入猜测结果。"
    return field_table_verified(schema["fields"])


def return_note(cmd: Command) -> str:
    lines = [
        f"- 成功：退出码 `0`；输出通道：{output_channel(cmd)}。",
        "- 失败：退出码非 `0`；stderr/stdout 包含 Ceph 返回的错误说明，常见为 `EINVAL`、`ENOENT`、`EACCES`、`EBUSY` 等 errno 语义。",
        f"- JSON：{json_support(cmd)}。",
        f"- JSON 返回形态：{return_shape_verified(cmd)}。",
        "",
        "#### 返回字段",
        "",
        return_fields_verified(cmd),
    ]
    if cmd.component == "admin socket":
        lines.insert(3, "- 调用语义：`ceph daemon` 直接访问本机 daemon 的 admin socket，要求调用端能访问对应 socket 文件。")
    if cmd.component == "ceph tell":
        lines.insert(3, "- 调用语义：`ceph tell` 通过 monitor 将请求转发给目标 daemon；返回内容与对应 admin socket 命令一致。")
    return "\n".join(lines)


def file_for_group(group: str) -> Path:
    mapping = {
        "cluster": "cluster/core.md",
        "auth": "cluster/auth.md",
        "config": "cluster/config.md",
        "config-key": "cluster/config-key.md",
        "mon": "components/mon.md",
        "mgr": "components/mgr.md",
        "mds": "components/mds.md",
        "fs": "components/cephfs.md",
        "osd": "components/osd.md",
        "pg": "components/pg.md",
        "orch": "mgr-modules/orchestrator.md",
        "orchestrator": "mgr-modules/orchestrator.md",
        "device": "mgr-modules/device.md",
        "dashboard": "mgr-modules/dashboard.md",
        "rbd": "mgr-modules/rbd.md",
        "rgw": "mgr-modules/rgw.md",
        "nfs": "mgr-modules/nfs.md",
        "smb": "mgr-modules/smb.md",
        "telemetry": "mgr-modules/telemetry.md",
        "prometheus": "mgr-modules/prometheus.md",
        "progress": "mgr-modules/progress.md",
        "iostat": "mgr-modules/iostat.md",
        "influx": "mgr-modules/influx.md",
        "telegraf": "mgr-modules/telegraf.md",
        "feedback": "mgr-modules/feedback.md",
        "alerts": "mgr-modules/alerts.md",
        "hello": "mgr-modules/hello.md",
        "count": "mgr-modules/hello.md",
        "healthcheck": "mgr-modules/prometheus.md",
    }
    if group.startswith("admin-socket/"):
        return OUT_DIR / f"{group}.md"
    if group.startswith("tell/"):
        return OUT_DIR / f"{group}.md"
    return OUT_DIR / mapping.get(group, f"mgr-modules/{group}.md")


def write_docs(commands: dict[tuple[str, str, str], Command]) -> None:
    if OUT_DIR.exists():
        shutil.rmtree(OUT_DIR)
    OUT_DIR.mkdir(parents=True)
    grouped: dict[Path, list[Command]] = defaultdict(list)
    for cmd in commands.values():
        grouped[file_for_group(cmd.group)].append(cmd)

    index_lines = [
        "# Ceph 命令参考",
        "",
        "> 来源：`docs/references/ceph` 的 Git tags：`v16.2.15`、`v17.2.9`、`v18.2.8`、`v19.2.5`、`v20.2.2`。",
        "> 本文档由 `tools/generate_ceph_command_docs.py` 生成。",
        "",
        "本文档整理 Ceph monitor/mgr command table 与 mgr Python 模块声明的 `ceph ...` 命令，",
        "用于后续在 `backend/internal/integrations/ceph` 中新增直接执行 Ceph CLI 的能力。",
        "",
        "## 返回约定",
        "",
        "- 命令可通过 `ceph <prefix> --format json` 或 `--format json-pretty` 请求 JSON 输出时，仅记录 JSON 输出形态。",
        "- 每条命令的返回信息包含成功/失败通道和 JSON 支持情况。",
        "- 返回字段只在该命令支持的所有版本源码中逐项校验通过后记录，并在字段表中标注源码确认位置。",
        "- Ceph 源码的命令表声明参数、权限、模块与帮助文本；没有集中声明完整返回 schema，未校验字段不会写入文档。",
        "- 写操作通常返回确认文本或空输出；失败时返回非 0 退出码，并在 stdout/stderr 中携带错误说明。",
        "- `admin-socket/` 目录记录 `ceph daemon <daemon>.<id> <cmd>` 这类 per-daemon 命令；远程场景通常也可通过 `ceph tell <daemon>.<id> <cmd>` 调用。",
        "- `tell/` 目录把同一批 per-daemon 命令展开为 `ceph tell <daemon>.<id> <cmd>` 的调用形式。",
        "",
        "## 目录",
        "",
    ]
    for path in sorted(grouped):
        rel = path.relative_to(OUT_DIR)
        title = rel.with_suffix("").as_posix()
        index_lines.append(f"- [{title}]({rel.as_posix()})")
    (OUT_DIR / "index.md").write_text("\n".join(index_lines) + "\n", encoding="utf-8")

    for path, cmds in grouped.items():
        path.parent.mkdir(parents=True, exist_ok=True)
        title = path.stem.replace("-", " ").title()
        lines = [
            f"# {title}",
            "",
            "> 自动生成；请修改 `tools/generate_ceph_command_docs.py` 后重新生成。",
            "",
        ]
        for cmd in sorted(cmds, key=lambda c: (c.prefix, c.module, c.component)):
            latest = next((v for v in reversed(VERSION_LABELS) if v in cmd.params_by_version), "")
            params = cmd.params_by_version.get(latest, [])
            fl = cmd.flags_by_version.get(latest, [])
            lines.extend([
                f"## `ceph {cmd.prefix}`",
                "",
                f"- 组件/模块：`{cmd.component}` / `{cmd.module}`",
                f"- 支持版本：{version_range(cmd.versions)}",
                f"- 权限：`{cmd.perm}`",
                f"- 来源：`{cmd.source}`",
            ])
            if fl:
                lines.append(f"- 标志：`{', '.join(fl)}`")
            lines.extend([
                f"- 含义：{cmd.desc.strip() or '-'}",
                "",
                "### 参数",
                "",
                render_params(params),
                "",
                "### 返回信息",
                "",
                return_note(cmd),
                "",
            ])
            versions_with_different_params = [
                v for v in VERSION_LABELS
                if v in cmd.params_by_version and cmd.params_by_version[v] != params
            ]
            if versions_with_different_params:
                lines.extend(["### 版本差异", ""])
                for v in versions_with_different_params:
                    lines.extend([f"- `{v}` 参数：", "", render_params(cmd.params_by_version[v]), ""])
        path.write_text("\n".join(lines), encoding="utf-8")


def main() -> None:
    commands = merge_commands()
    write_docs(commands)
    print(f"generated {len(commands)} commands into {OUT_DIR.relative_to(ROOT)}")


if __name__ == "__main__":
    main()
