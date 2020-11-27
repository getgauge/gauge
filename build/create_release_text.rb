# ----------------------------------------------------------------
#   Copyright (c) ThoughtWorks, Inc.
#   Licensed under the Apache License, Version 2.0
#   See LICENSE.txt in the project root for license information.
# ----------------------------------------------------------------

require 'net/http'
require 'uri'
require 'json'

if ARGV.length < 1
  puts 'Usage: ruby create_release_text.rb <owner> <name>
Example: ruby create_release_text.rb getgauge gauge
'
  exit 1
end

repo = "#{ARGV[0]}/#{ARGV[1]}"

api  = "https://api.github.com"

latest_release = URI.join(api, "/repos/#{repo}/releases/latest")
timestamp = JSON.parse(Net::HTTP.get(latest_release))['published_at']

issues_query = "/search/issues?q=is:pr+repo:#{repo}+closed"

if not timestamp.nil? || timestamp.empty? 
  issues_query += ":>#{timestamp}"
end

uri = URI.join(api, issues_query)
req = Net::HTTP::Get.new(uri)
req['Accept'] = "application/vnd.github.v3.full+json"
http = Net::HTTP.new(uri.hostname, uri.port)
http.use_ssl = true
response = http.request(req)

case response
  when Net::HTTPSuccess
    issues = JSON.parse(response.body)

    issues['items'].each do |issue|
      puts "- ##{issue['number']} - #{issue['title']}"
    end

  else
    raise "
    Could not fetch release information for github repo: 

      https://github.com/#{repo}

    Please check if this is a valid repo."
end
