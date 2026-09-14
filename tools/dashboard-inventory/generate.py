"""Generate a source navigation index; this is not a feature parity claim."""
from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[2]
SOURCE = ROOT / 'docs/references/ceph/src/pybind/mgr/dashboard'
OUTPUT = ROOT / 'docs/ceph/dashboard-source-index.md'


def link(path):
    return str(Path('..') / '..' / path.relative_to(ROOT))


def generate():
    lines = [
        '# Ceph Dashboard 前端与 CephTower API 源码索引', '',
        '由 `python3 tools/dashboard-inventory/generate.py` 生成。', '',
        '此索引列出参考前端服务的声明方法及本项目已注册的 API，供逐项追踪页面 → 服务 →',
        'Dashboard 控制器 → Ceph 命令 → CephTower API 使用。方法名来自静态扫描，可能包括',
        '界面辅助方法；存在方法、路由或命令包装不代表已经完成端到端实现或实机验证。', '',
        '实现进展和差异见 [dashboard-implementation.md](dashboard-implementation.md)。', '',
        '## 参考前端服务', '', '| 服务源码 | 声明方法（包含辅助方法） |', '| --- | --- |',
    ]
    for path in sorted((SOURCE / 'frontend/src/app/shared/api').glob('*.service.ts')):
        methods = re.findall(r'^  (?:(?:public|private|protected|async|static) +)*([A-Za-z_$][\w$]*)\s*(?:<[^\n]+>)?\(', path.read_text(), re.M)
        methods = list(dict.fromkeys(m for m in methods if m != 'constructor'))
        lines.append(f'| [{path.stem}]({link(path)}) | ' + ', '.join(f'`{m}`' for m in methods) + ' |')
    lines += ['', '## 参考控制器', '', '| 源码 | API 路由声明 |', '| --- | --- |']
    for path in sorted((SOURCE / 'controllers').glob('*.py')):
        if path.name.startswith('_'):
            continue
        routes = re.findall(r'(?:APIRouter|UIRouter)\(\s*[\'\"]([^\'\"]+)', path.read_text())
        lines.append(f'| [{path.stem}]({link(path)}) | ' + ', '.join(f'`{r}`' for r in dict.fromkeys(routes)) + ' |')
    lines += ['', '## CephTower 已注册 API', '', '| 路由源码 | 方法和路径 | Handler |', '| --- | --- | --- |']
    for path in sorted((ROOT / 'backend/internal/api/v1/router').glob('*.go')):
        if path.name.endswith('_test.go'):
            continue
        for method, route, handler in re.findall(r'\{"(GET|POST|PUT|PATCH|DELETE)", "([^"]+)", h\.(\w+)\}', path.read_text()):
            lines.append(f'| [{path.name}]({link(path)}) | `{method} /api/v1{route}` | `{handler}` |')
    return '\n'.join(lines) + '\n'


if __name__ == '__main__':
    OUTPUT.write_text(generate())
    print(OUTPUT.relative_to(ROOT))
