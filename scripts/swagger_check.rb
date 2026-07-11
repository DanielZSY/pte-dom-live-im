#!/usr/bin/env ruby
# frozen_string_literal: true

require 'yaml'

ROOT = File.expand_path('..', __dir__)

def routes_from_file(path, scene_prefixes: false)
  source = File.read(path)
  routes = source.scan(/HandleFunc\("([^"\\]+)"/).flatten
  return routes unless scene_prefixes

  suffixes = %w[
    /room/open /room/close /room/list /room/detail /room/enter /room/leave
    /member/list /seat/action /moderation/action /pk/invite /pk/start /pk/action
    /pk/end /event/send /event/list
  ]
  prefixes = source.scan(/registerSceneRoutes\(mux, "([^"\\]+)"/).flatten
  routes + prefixes.product(suffixes).map { |prefix, suffix| "#{prefix}#{suffix}" }
end

def check(service, router, spec, scene_prefixes: false)
  expected = routes_from_file(router, scene_prefixes: scene_prefixes).uniq.sort
  document = YAML.load_file(spec)
  actual = document.fetch('paths', {}).keys.sort
  missing = expected - actual
  raise "#{service} Swagger 缺少路由: #{missing.join(', ')}" unless missing.empty?

  puts "ok: #{service} (#{expected.length} routes)"
end

check(
  'api-im',
  File.join(ROOT, 'pte-live-im/routers/routers.go'),
  File.join(ROOT, 'pte-live-im/docs/openapi.yaml')
)
check(
  'api-chat',
  File.join(ROOT, 'pte-live-api-chat/api/router/router.go'),
  File.join(ROOT, 'pte-live-api-chat/docs/openapi.yaml'),
  scene_prefixes: true
)
check(
  'api-chat-admin',
  File.join(ROOT, 'pte-live-api-chat-admin/api/router/router.go'),
  File.join(ROOT, 'pte-live-api-chat-admin/docs/openapi.yaml')
)
