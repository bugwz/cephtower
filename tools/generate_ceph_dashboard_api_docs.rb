#!/usr/bin/env ruby
# frozen_string_literal: true

require 'fileutils'
require 'set'
require 'yaml'

ROOT = File.expand_path('..', __dir__)
OPENAPI_PATH = File.join(ROOT, 'docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml')
CMAKE_PATH = File.join(ROOT, 'docs/references/ceph/CMakeLists.txt')
OUT_DIR = File.join(ROOT, 'docs/ceph/apis')
API_DIR = File.join(OUT_DIR, 'endpoints')
CEPH_REPO = File.join(ROOT, 'docs/references/ceph')

HTTP_METHODS = %w[get post put patch delete].freeze
CEPH_VERSIONS = %w[v16.2.15 v17.2.9 v18.2.8 v19.2.5 v20.2.2].freeze

CATEGORY_TITLES = {
  'auth' => '认证',
  'block' => '块存储 RBD',
  'cephfs' => 'CephFS',
  'cluster' => '集群',
  'cluster_conf' => '集群配置',
  'crush_rule' => 'CRUSH 规则',
  'daemon' => 'Daemon',
  'erasure_code_profile' => '纠删码配置',
  'feature_toggles' => '功能开关',
  'feedback' => '反馈',
  'grafana' => 'Grafana',
  'hardware' => '硬件',
  'health' => '健康状态',
  'host' => '主机',
  'iscsi' => 'iSCSI',
  'logs' => '日志',
  'mgr' => 'Mgr 模块',
  'monitor' => 'Monitor',
  'motd' => 'MOTD',
  'multi-cluster' => '多集群',
  'nfs-ganesha' => 'NFS Ganesha',
  'nvmeof' => 'NVMe-oF',
  'osd' => 'OSD',
  'perf_counters' => '性能计数器',
  'pool' => '存储池',
  'prometheus' => 'Prometheus',
  'rgw' => 'RGW',
  'role' => '角色',
  'service' => '服务',
  'settings' => '设置',
  'smb' => 'SMB',
  'summary' => '概览',
  'task' => '任务',
  'telemetry' => 'Telemetry',
  'user' => 'Dashboard 用户'
}.freeze

METHOD_ORDER = {
  'get' => 0,
  'post' => 1,
  'put' => 2,
  'patch' => 3,
  'delete' => 4
}.freeze

def ceph_version
  cmake = File.read(CMAKE_PATH)
  match = cmake.match(/project\([^)]*VERSION\s+([0-9.]+)/m)
  match ? match[1] : 'unknown'
end

def git_show(tag, path)
  IO.popen(['git', '-C', CEPH_REPO, 'show', "#{tag}:#{path}"], &:read)
end

def load_openapi_for_tag(tag)
  YAML.safe_load(git_show(tag, 'src/pybind/mgr/dashboard/openapi.yaml'))
end

def operation_key(method, path)
  "#{method.upcase} #{path}"
end

def category_for(path)
  parts = path.split('/').reject(&:empty?)
  return 'root' unless parts.first == 'api'

  parts[1] || 'root'
end

def category_title(category)
  CATEGORY_TITLES.fetch(category, category)
end

def markdown_table(headers, rows)
  out = []
  out << "| #{headers.join(' | ')} |"
  out << "| #{headers.map { '---' }.join(' | ')} |"
  rows.each do |row|
    out << "| #{row.map { |cell| escape_table_cell(cell) }.join(' | ')} |"
  end
  out.join("\n")
end

def escape_table_cell(value)
  text = value.nil? ? '' : value.to_s
  text.gsub("\n", '<br>').gsub('|', '\\|')
end

def schema_type(schema)
  return '' unless schema

  if schema['type']
    if schema['type'] == 'array' && schema['items']
      "array<#{schema_type(schema['items']).empty? ? 'object' : schema_type(schema['items'])}>"
    else
      schema['type'].to_s
    end
  elsif schema['properties']
    'object'
  elsif schema['$ref']
    schema['$ref']
  else
    'object'
  end
end

def schema_summary(schema)
  return '无 schema' unless schema

  type = schema_type(schema)
  required = Array(schema['required'])
  props = schema['properties'] || {}
  if props.empty?
    return type.empty? ? 'object' : type
  end

  "#{type}; fields: #{props.keys.join(', ')}; required: #{required.empty? ? '无' : required.join(', ')}"
end

def yaml_block(value)
  return '无' if value.nil?

  dumped = YAML.dump(value).sub(/\A---\n/, '').rstrip
  dumped.empty? ? '无' : dumped
end

def render_parameters(op)
  params = Array(op['parameters'])
  return "无。\n" if params.empty?

  rows = params.map do |param|
    schema = param['schema'] || {}
    default = if schema.key?('default')
                schema['default']
              elsif param.key?('default')
                param['default']
              end
    [
      param['name'],
      param['in'],
      param['required'] ? '是' : '否',
      schema_type(schema),
      default.nil? ? '' : default.inspect,
      param['description'] || ''
    ]
  end
  "#{markdown_table(%w[名称 位置 必填 类型 默认值 说明], rows)}\n"
end

def render_request_body(op)
  body = op['requestBody']
  return "无请求体。\n" unless body

  required = body['required'] ? '是' : '否'
  out = ["请求体必填：#{required}"]
  content = body['content'] || {}
  content.each do |mime, spec|
    schema = spec['schema'] || spec
    out << ""
    out << "- Content-Type: `#{mime}`"
    out << "- Schema: #{schema_summary(schema)}"
    out << ""
    out << "```yaml"
    out << yaml_block(schema)
    out << "```"
  end
  out << "" if content.empty?
  out << "无 content schema。" if content.empty?
  "#{out.join("\n")}\n"
end

def render_responses(op)
  responses = op['responses'] || {}
  out = []
  responses.sort_by { |code, _| code.to_s }.each do |code, response|
    out << "#### `#{code}`"
    out << ""
    out << response['description'].to_s unless response['description'].to_s.empty?
    content = response['content'] || {}
    if content.empty?
      out << ""
      out << "无响应体 schema。"
      out << ""
      next
    end
    content.each do |mime, spec|
      schema = spec['schema'] || spec
      out << ""
      out << "- Content-Type: `#{mime}`"
      out << "- Schema: #{schema_summary(schema)}"
      out << ""
      out << "```yaml"
      out << yaml_block(schema)
      out << "```"
    end
    out << ""
  end
  out.join("\n")
end

def operation_anchor(method, path)
  "#{method}-#{path}".downcase.gsub(%r{[^a-z0-9]+}, '-').gsub(/\A-|-+\z/, '')
end

def render_version_support(method, path, version_indexes)
  info = support_info(method, path, version_indexes)
  version_text = info[:versions].empty? ? '无' : info[:versions].join(', ')
  out = []
  out << "#### 版本支持"
  out << ""
  out << "- 支持版本：#{version_text}"
  out << "- 首次出现在扫描范围：#{info[:since] || '无'}"
  out << "- v20.2.2 当前文档支持：#{info[:current] ? '是' : '否'}"
  out << ""
  out.join("\n")
end

def render_operation(method, path, op, version_indexes)
  title = op['summary'] || "#{method.upcase} #{path}"
  tags = Array(op['tags'])
  security = if op.key?('security')
               yaml_block(op['security']).gsub("\n", ' ')
             else
               'OpenAPI 未声明 JWT security，通常为公开或由控制器单独处理'
             end
  out = []
  out << "### `#{method.upcase} #{path}`"
  out << ""
  out << "- 摘要：#{title}"
  out << "- Tags：#{tags.empty? ? '无' : tags.map { |tag| "`#{tag}`" }.join(', ')}"
  out << "- 安全：#{security}"
  out << ""
  out << render_version_support(method, path, version_indexes)
  out << "#### 请求参数"
  out << ""
  out << render_parameters(op)
  out << ""
  out << "#### 请求体"
  out << ""
  out << render_request_body(op)
  out << ""
  out << "#### 返回消息"
  out << ""
  out << render_responses(op)
  out << ""
  out.join("\n")
end

def collect_operations(openapi)
  grouped = Hash.new { |h, k| h[k] = [] }
  openapi['paths'].each do |path, item|
    item.each do |method, op|
      next unless HTTP_METHODS.include?(method)

      grouped[category_for(path)] << [path, method, op]
    end
  end
  grouped.each_value do |ops|
    ops.sort_by! { |path, method, _op| [path, METHOD_ORDER.fetch(method, 99)] }
  end
  grouped
end

def collect_operation_index(openapi)
  index = {}
  openapi['paths'].each do |path, item|
    item.each do |method, op|
      next unless HTTP_METHODS.include?(method)

      index[operation_key(method, path)] = {
        'path' => path,
        'method' => method,
        'summary' => op['summary'],
        'tags' => Array(op['tags'])
      }
    end
  end
  index
end

def build_version_indexes
  CEPH_VERSIONS.to_h do |tag|
    [tag, collect_operation_index(load_openapi_for_tag(tag))]
  end
end

def support_info(method, path, version_indexes)
  key = operation_key(method, path)
  versions = CEPH_VERSIONS.select { |tag| version_indexes.fetch(tag).key?(key) }
  {
    key: key,
    versions: versions,
    since: versions.first,
    current: versions.include?('v20.2.2')
  }
end

def write_category(category, operations, version, version_indexes)
  filename = "#{category}.md"
  path = File.join(API_DIR, filename)
  title = category_title(category)
  out = []
  out << "# Ceph #{version} Dashboard API - #{title}"
  out << ""
  out << "> 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`。"
  out << "> 本文档由 `tools/generate_ceph_dashboard_api_docs.rb` 生成，按 `/api/#{category}` 路径域归类。"
  out << "> 版本支持扫描范围：#{CEPH_VERSIONS.join(', ')}。"
  out << ""
  out << "## 接口目录"
  out << ""
  operations.each do |path_name, method, op|
    summary = op['summary'] || "#{method.upcase} #{path_name}"
    out << "- [`#{method.upcase} #{path_name}`](##{operation_anchor(method, path_name)}) - #{summary}"
  end
  out << ""
  out << "## 接口详情"
  out << ""
  operations.each do |path_name, method, op|
    out << render_operation(method, path_name, op, version_indexes)
  end
  File.write(path, out.join("\n"))
  filename
end

def version_stat_rows(version_indexes)
  CEPH_VERSIONS.map do |tag|
    index = version_indexes.fetch(tag)
    [tag, index.values.map { |op| op['path'] }.uniq.size, index.size]
  end
end

def write_index(grouped, version, version_indexes)
  total_ops = grouped.values.flatten(1).size
  out = []
  out << "# Ceph #{version} Mgr Dashboard API"
  out << ""
  out << "本文档集整理自 Ceph v#{version} 源码内置 Dashboard OpenAPI 描述，用于本项目后续通过 mgr dashboard API 操作 Ceph 集群。"
  out << ""
  out << "## 来源与调用约定"
  out << ""
  out << "- OpenAPI 来源：`docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml`"
  out << "- 版本来源：`docs/references/ceph/CMakeLists.txt` 中的 `VERSION #{version}`"
  out << "- 版本支持扫描范围：#{CEPH_VERSIONS.join(', ')}"
  out << "- API 基础路径：OpenAPI `basePath` 为 `/`，接口路径以 `/api/...` 为主。"
  out << "- 认证方式：带 `security: [{jwt: []}]` 的接口使用 Bearer JWT。通常先调用 `POST /api/auth` 获取 `token`，后续请求使用 `Authorization: Bearer <token>`。"
  out << "- 公开接口：未声明 `security` 的接口通常为公开入口或由控制器单独处理，调用前仍应结合部署侧 Dashboard 配置确认。"
  out << "- 内容类型：请求体通常为 `application/json`；响应内容类型通常为 `application/vnd.ceph.api.v1.0+json`，部分接口使用其他 API 版本 MIME。"
  out << "- 异步任务：许多写操作可能返回 `202`，表示操作仍在执行，需要查询任务队列。"
  out << "- 通用错误：多数接口包含 `400`、`401`、`403`、`500`。具体响应体以运行时 Dashboard 返回为准。"
  out << "- 版本兼容性总览：[compatibility.md](compatibility.md)"
  out << ""
  out << "## 分类索引"
  out << ""
  rows = grouped.keys.sort.map do |category|
    filename = "#{category}.md"
    [category_title(category), "`#{category}`", "[endpoints/#{filename}](endpoints/#{filename})", grouped[category].size]
  end
  out << markdown_table(%w[分类 路径域 文档 接口数], rows)
  out << ""
  out << "## 统计"
  out << ""
  out << "- 路径数：#{grouped.values.flatten(1).map(&:first).uniq.size}"
  out << "- 接口操作数：#{total_ops}"
  out << "- 分类数：#{grouped.keys.size}"
  out << ""
  out << "## 各版本接口数量"
  out << ""
  out << markdown_table(%w[版本 路径数 接口操作数], version_stat_rows(version_indexes))
  out << ""
  File.write(File.join(OUT_DIR, 'index.md'), out.join("\n"))
end

def write_compatibility(version_indexes)
  current_keys = version_indexes.fetch('v20.2.2').keys.to_set
  all_keys = version_indexes.values.flat_map(&:keys).uniq.sort
  removed_keys = all_keys.reject { |key| current_keys.include?(key) }

  out = []
  out << "# Ceph Dashboard API 版本兼容性"
  out << ""
  out << "本文档汇总 #{CEPH_VERSIONS.join(', ')} 的 Dashboard API operation 支持情况。"
  out << ""
  out << "## 版本统计"
  out << ""
  out << markdown_table(%w[版本 路径数 接口操作数], version_stat_rows(version_indexes))
  out << ""
  out << "## v20.2.2 未包含的历史接口"
  out << ""
  if removed_keys.empty?
    out << "扫描范围内没有发现旧版本存在但 v20.2.2 已不在 OpenAPI 中的接口。"
  else
    rows = removed_keys.map do |key|
      versions = CEPH_VERSIONS.select { |tag| version_indexes.fetch(tag).key?(key) }
      [key, versions.join(', ')]
    end
    out << markdown_table(%w[接口 支持版本], rows)
  end
  out << ""
  File.write(File.join(OUT_DIR, 'compatibility.md'), out.join("\n"))
end

def cleanup_generated_files(grouped)
  FileUtils.mkdir_p(OUT_DIR)
  grouped.keys.each do |category|
    FileUtils.rm_f(File.join(OUT_DIR, "#{category}.md"))
  end
  FileUtils.rm_rf(API_DIR)
  FileUtils.mkdir_p(API_DIR)
end

openapi = YAML.safe_load(File.read(OPENAPI_PATH))
version = ceph_version
grouped = collect_operations(openapi)
version_indexes = build_version_indexes
cleanup_generated_files(grouped)

grouped.keys.sort.each do |category|
  write_category(category, grouped[category], version, version_indexes)
end
write_index(grouped, version, version_indexes)
write_compatibility(version_indexes)

puts "Generated #{grouped.keys.size + 2} markdown files for #{grouped.values.flatten(1).size} operations."
