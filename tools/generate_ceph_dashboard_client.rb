#!/usr/bin/env ruby
# frozen_string_literal: true

require 'fileutils'
require 'set'
require 'yaml'

ROOT = File.expand_path('..', __dir__)
OPENAPI_PATH = File.join(ROOT, 'docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml')
CEPH_DIR = File.join(ROOT, 'backend/internal/ceph')
ENDPOINTS_DIR = File.join(CEPH_DIR, 'endpoints')
TYPED_DIR = File.join(CEPH_DIR, 'typed')
HTTP_METHODS = %w[get post put patch delete].freeze
GO_KEYWORDS = Set.new(%w[
  break default func interface select case defer go map struct chan else goto package
  switch const fallthrough if range type continue for import return var
]).freeze

def camelize(value)
  text = value.to_s
    .split(/[^a-zA-Z0-9]+/)
    .reject(&:empty?)
    .map { |part| part[0].upcase + part[1..] }
    .join
  text = 'Value' if text.empty?
  text = "X#{text}" unless text.match?(/\A[A-Za-z_]/)
  text += 'Value' if GO_KEYWORDS.include?(text.downcase)
  text
end

def method_name(http_method, path)
  segments = path.split('/').reject(&:empty?)
  segments.shift if segments.first == 'api'

  suffix = segments.map do |segment|
    if segment.start_with?('{') && segment.end_with?('}')
      "By#{camelize(segment[1...-1])}"
    else
      camelize(segment)
    end
  end.join

  "#{camelize(http_method)}#{suffix.empty? ? 'Root' : suffix}"
end

def category_for(path)
  segments = path.split('/').reject(&:empty?)
  category = segments.first == 'api' ? segments[1] : segments[0]
  (category || 'root').tr('-', '_')
end

def jwt_required?(operation)
  Array(operation['security']).any? { |entry| entry.key?('jwt') }
end

def schema_content(container)
  content = container&.fetch('content', nil)
  return nil unless content

  preferred = content['application/json'] ||
              content['application/vnd.ceph.api.v1.0+json'] ||
              content.values.first
  return nil unless preferred

  preferred['schema'] || preferred
end

def success_schema(operation)
  responses = operation['responses'] || {}
  code = %w[200 201 202 204].find { |status| responses[status] } ||
         responses.keys.select { |status| status.to_s.start_with?('2') }.sort.first
  return nil unless code

  schema_content(responses[code])
end

def request_body_schema(operation)
  schema_content(operation['requestBody'])
end

def go_field_name(name, used)
  base = camelize(name)
  candidate = base
  index = 2
  while used.include?(candidate)
    candidate = "#{base}#{index}"
    index += 1
  end
  used << candidate
  candidate
end

def schema_go_type(schema, type_name, definitions)
  return 'json.RawMessage' unless schema.is_a?(Hash)

  if schema['nullable']
    nested = schema_go_type(schema.merge('nullable' => false), type_name, definitions)
    return "*#{nested}" unless nested.start_with?('[]') || nested.start_with?('map[') || nested == 'json.RawMessage'
    return nested
  end

  case schema['type']
  when 'object', nil
    properties = schema['properties'] || {}
    if properties.empty?
      additional = schema['additionalProperties']
      return "map[string]#{schema_go_type(additional, "#{type_name}Value", definitions)}" if additional.is_a?(Hash)

      return 'map[string]json.RawMessage'
    end

    used = Set.new
    required = Set.new(Array(schema['required']))
    lines = ["type #{type_name} struct {"]
    properties.each do |prop_name, prop_schema|
      field_name = go_field_name(prop_name, used)
      field_type = schema_go_type(prop_schema, "#{type_name}#{field_name}", definitions)
      tag = required.include?(prop_name) ? prop_name : "#{prop_name},omitempty"
      lines << "\t#{field_name} #{field_type} `json:\"#{tag}\"`"
    end
    lines << "}"
    definitions << lines.join("\n")
    type_name
  when 'array'
    "[]#{schema_go_type(schema['items'] || {}, "#{type_name}Item", definitions)}"
  when 'integer'
    'int'
  when 'number'
    'float64'
  when 'boolean'
    'bool'
  when 'string'
    'string'
  else
    'json.RawMessage'
  end
end

def define_named_schema(type_name, schema, definitions)
  if schema.nil?
    definitions << "type #{type_name} = EmptyResponse"
    return 'EmptyResponse'
  end

  go_type = schema_go_type(schema, type_name, definitions)
  definitions << "type #{type_name} #{go_type}" unless go_type == type_name
  type_name
end

def query_field_type(param)
  schema = param['schema'] || {}
  case schema['type']
  when 'integer'
    'int'
  when 'number'
    'float64'
  when 'boolean'
    'bool'
  when 'array'
    '[]string'
  else
    'string'
  end
end

def query_setter(field_name, param)
  name = param.fetch('name')
  required = param['required']
  schema = param['schema'] || {}
  value = "request.#{field_name}"
  target = required ? value : "*#{value}"
  condition = required ? '' : "\tif #{value} != nil {\n"
  close = required ? '' : "\t}\n"

  line = case schema['type']
         when 'integer'
           "\tquery.Set(#{name.inspect}, strconv.Itoa(#{target}))\n"
         when 'number'
           "\tquery.Set(#{name.inspect}, strconv.FormatFloat(#{target}, 'f', -1, 64))\n"
         when 'boolean'
           "\tquery.Set(#{name.inspect}, strconv.FormatBool(#{target}))\n"
         when 'array'
           required ? "\tfor _, value := range #{target} {\n\t\tquery.Add(#{name.inspect}, value)\n\t}\n" : "\tfor _, value := range #{target} {\n\t\tquery.Add(#{name.inspect}, value)\n\t}\n"
         else
           "\tquery.Set(#{name.inspect}, #{target})\n"
         end

  "#{condition}#{line}#{close}"
end

def raw_header(category)
  <<~GO
    // Code generated by tools/generate_ceph_dashboard_client.rb; DO NOT EDIT.

    package endpoints

    import (
    	"context"
    	"encoding/json"
    	"net/http"
    )

    // #{category} endpoints from docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml.
  GO
end

def typed_header(category, needs_json, needs_strconv)
  imports = ["\t\"context\""]
  imports << "\t\"encoding/json\"" if needs_json
  imports.concat(["\t\"net/http\"", "\t\"net/url\""])
  imports << "\t\"strconv\"" if needs_strconv
  <<~GO
    // Code generated by tools/generate_ceph_dashboard_client.rb; DO NOT EDIT.

    package typed

    import (
    #{imports.join("\n")}
    )

    // #{category} typed endpoints from docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml.
  GO
end

openapi = YAML.safe_load(File.read(OPENAPI_PATH))
operations_by_category = Hash.new { |hash, key| hash[key] = [] }
seen_methods = {}

openapi.fetch('paths').each do |path, methods|
  methods.each do |http_method, operation|
    next unless HTTP_METHODS.include?(http_method)

    name = method_name(http_method, path)
    raise "duplicate generated method #{name}" if seen_methods[name]

    seen_methods[name] = true
    operations_by_category[category_for(path)] << {
      method: http_method.upcase,
      name: name,
      path: path,
      auth: jwt_required?(operation),
      summary: operation['summary'].to_s.strip,
      parameters: Array(operation['parameters']),
      body_schema: request_body_schema(operation),
      response_schema: success_schema(operation)
    }
  end
end

[ENDPOINTS_DIR, TYPED_DIR].each { |dir| FileUtils.mkdir_p(dir) }
Dir.glob(File.join(CEPH_DIR, 'zz_generated_*.go')).each { |path| File.delete(path) }
Dir.glob(File.join(ENDPOINTS_DIR, 'generated_*.go')).each { |path| File.delete(path) }
Dir.glob(File.join(TYPED_DIR, '*.go')).each { |path| File.delete(path) unless File.basename(path) == 'client.go' }
operations_by_category.keys.each do |category|
  [ENDPOINTS_DIR, TYPED_DIR].each do |dir|
    path = File.join(dir, "#{category}.go")
    File.delete(path) if File.exist?(path)
  end
end

operations_by_category.sort.each do |category, operations|
  raw_body = raw_header(category).dup
  typed_defs = []
  typed_methods = []
  needs_strconv = false

  operations.sort_by { |operation| [operation[:path], operation[:method]] }.each do |operation|
    doc = operation[:summary].empty? ? "#{operation[:method]} #{operation[:path]}" : operation[:summary]

    raw_body << "\n// #{operation[:name]} calls #{operation[:method]} #{operation[:path]}.\n"
    raw_body << "// #{doc.gsub(%r{\s+}, ' ')}\n"
    raw_body << "func (c *Client) #{operation[:name]}(ctx context.Context, request OperationRequest) (json.RawMessage, error) {\n"
    raw_body << "\treturn c.do(ctx, http.Method#{camelize(operation[:method].downcase)}, #{operation[:path].inspect}, request, #{operation[:auth]})\n"
    raw_body << "}\n"

    request_type = "#{operation[:name]}Request"
    response_type = "#{operation[:name]}Response"
    body_type = "#{operation[:name]}Body"
    local_defs = []
    field_lines = []
    path_lines = []
    query_lines = ["\tquery := url.Values{}\n"]
    used_fields = Set.new

    operation[:parameters].each do |param|
      next unless %w[path query].include?(param['in'])

      field_name = go_field_name(param.fetch('name'), used_fields)
      if param['in'] == 'path'
        field_lines << "\t#{field_name} string `path:\"#{param.fetch('name')}\"`"
        path_lines << "\t\t#{param.fetch('name').inspect}: request.#{field_name},\n"
      else
        base_type = query_field_type(param)
        needs_strconv = true if %w[int float64 bool].include?(base_type)
        field_type = param['required'] ? base_type : "*#{base_type}"
        field_lines << "\t#{field_name} #{field_type} `query:\"#{param.fetch('name')}\"`"
        query_lines << query_setter(field_name, param)
      end
    end

    body_go_type = nil
    if operation[:body_schema]
      body_go_type = define_named_schema(body_type, operation[:body_schema], local_defs)
      field_lines << "\tBody #{body_go_type} `json:\"-\"`"
    end
    define_named_schema(response_type, operation[:response_schema], local_defs)

    typed_defs.concat(local_defs)
    typed_defs << if field_lines.empty?
                    "type #{request_type} struct{}"
                  else
                    "type #{request_type} struct {\n#{field_lines.join("\n")}\n}"
                  end

    method = +"// #{operation[:name]} calls #{operation[:method]} #{operation[:path]} with typed request and response values.\n"
    method << "func (c *Client) #{operation[:name]}(ctx context.Context, request #{request_type}) (#{response_type}, error) {\n"
    method << "\tpath := map[string]string{\n#{path_lines.join}\t}\n"
    method << query_lines.join
    method << "\tvar body any\n"
    method << "\tbody = nil\n"
    method << "\tbody = request.Body\n" if body_go_type
    method << "\tvar response #{response_type}\n"
    method << "\terr := c.doJSON(ctx, http.Method#{camelize(operation[:method].downcase)}, #{operation[:path].inspect}, path, query, body, #{operation[:auth]}, &response)\n"
    method << "\treturn response, err\n"
    method << "}\n"
    typed_methods << method
  end

  File.write(File.join(ENDPOINTS_DIR, "#{category}.go"), raw_body)
  File.write(
    File.join(TYPED_DIR, "#{category}.go"),
    typed_header(category, typed_defs.any? { |definition| definition.include?('json.RawMessage') }, needs_strconv) + "\n" + typed_defs.join("\n\n") + "\n\n" + typed_methods.join("\n")
  )
end

puts "generated #{seen_methods.size} raw and typed Ceph dashboard endpoint methods in #{operations_by_category.size} files"
