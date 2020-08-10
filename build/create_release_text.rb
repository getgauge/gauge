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

    categories = {"feature" => [], 
                  "bug" => []}

    headers = {"feature" => "Features", 
               "bug" => "Bug Fixes"}

    issues['items'].each do |issue|
      issue_text = "- ##{issue['number']} - #{issue['title']}"
      label_for_display = issue['labels'].map {|x| x['name']} & categories.keys
      if not label_for_display.empty?
        categories[label_for_display[0]] << issue_text
      end
    end

    categories.each_key do |category|
      puts "## #{headers[category]}\n\n"
      puts 'None' if categories[category].empty? 
      categories[category].each {|v| puts v}
      puts "\n" 
    end
  else
    raise "
    Could not fetch release information for github repo: 

      https://github.com/#{repo}

    Please check if this is a valid repo."
end
