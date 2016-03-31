# Copyright 2015 ThoughtWorks, Inc.

# This file is part of Gauge.

# Gauge is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.

# Gauge is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.

# You should have received a copy of the GNU General Public License
# along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

require 'parser/current'
require 'net/http'
require 'uri'
require 'json'


if ARGV.length < 2
  puts 'Minimum two arguments required.
Usage: ruby brew_update.rb <version> <path to file>.
Example: ruby brew_update.rb 0.3.2 Library/Formula/gauge.rb.
'
  exit 1
end

Parser::Builders::Default.emit_lambda = true # opt-in to most recent AST format
filter_dep = %w(gopkg.in/check.v1 golang.org/x/sys/unix golang.org/x/tools/go/ast/astutil golang.org/x/tools/go/exact golang.org/x/tools/go/types github.com/golang/protobuf/proto)
dependency_map = {}
code = File.read(ARGV[1])
uri = URI('https://raw.githubusercontent.com/getgauge/gauge/master/Godeps/Godeps.json')
response = Net::HTTP.get(uri)
data = JSON.parse(response)
data['Deps'].each { |dep| dependency_map[dep['ImportPath']] = dep['Rev'] }
dependency_map['golang.org/x/tools'] = dependency_map['golang.org/x/tools/go/exact']
dependency_map['github.com/golang/protobuf'] = dependency_map['github.com/golang/protobuf/proto']
filter_dep.each { |dep| dependency_map.delete(dep) }
template = '
go_resource "%s" do
  url "https://%s.git",
      :revision => "%s"
end
'

class Processor < AST::Processor
  attr_accessor :dependency_map, :old_sha256

  def initialize()
    @dependency_map = {}
    @last_value = ''
  end

  def on_begin(node)
    node.children.each { |c| process(c) }
  end

  def on_class(node)
    node.children.each { |c| process(c) }
  end

  def on_block(node)
    node.children.each { |c| process(c) }
  end

  def on_send(node)
    if node.children[1].to_s == 'sha256' and node.children[2].children[0].instance_of? String
      @old_sha256 = node.children[2].children[0]
    end
    if node.children[1].to_s == 'go_resource'
      @dependency_map[node.children[2].children[0].to_s] = ''
      @last_value = node.children[2].children[0].to_s
    elsif node.children[1].to_s == 'url' and node.children[3] != nil
      @dependency_map[@last_value] = node.children[3].children[0].children[1].children[0].to_s
    end
  end
end

ast = Processor.new
ast.process(Parser::CurrentRuby.parse(code))

`curl -O -L https://github.com/getgauge/gauge/archive/v#{ARGV[0]}.tar.gz`
sha256 = `shasum -a 256 v#{ARGV[0]}.tar.gz`.split[0]

code = code.sub! ast.old_sha256, sha256
code = code.gsub(%r{(https://github.com/getgauge/gauge/archive/)v\d?.\d?.\d?.tar.gz}, "https://github.com/getgauge/gauge/archive/v#{ARGV[0]}.tar.gz")
ast.dependency_map.keys.each { |key| code = code.sub! ast.dependency_map[key], dependency_map[key] }

File.write(ARGV[1], code)

diff = dependency_map.keys - ast.dependency_map.keys
if diff.length > 0
  puts "There are new dependencies: #{diff.join}"
  diff.each { |d| puts template % [d, d, dependency_map[d]] }
end
puts 'Update done.'