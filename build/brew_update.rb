# ----------------------------------------------------------------
#   Copyright (c) ThoughtWorks, Inc.
#   Licensed under the Apache License, Version 2.0
#   See LICENSE.txt in the project root for license information.
# ----------------------------------------------------------------

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
code = File.read(ARGV[1])

class Processor < AST::Processor
  attr_accessor :old_sha256

  def initialize()
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
  end
end

ast = Processor.new
ast.process(Parser::CurrentRuby.parse(code))

`curl -O -L https://github.com/getgauge/gauge/archive/v#{ARGV[0]}.tar.gz`
sha256 = `shasum -a 256 v#{ARGV[0]}.tar.gz`.split[0]

code = code.sub! ast.old_sha256, sha256
code = code.gsub(%r{(https://github.com/getgauge/gauge/archive/)v\d?.\d?.\d?.tar.gz}, "https://github.com/getgauge/gauge/archive/v#{ARGV[0]}.tar.gz")

File.write(ARGV[1], code)

puts 'Update done.'
