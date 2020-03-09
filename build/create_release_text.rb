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

require 'net/http'
require 'uri'
require 'json'

if ARGV.length < 1
  puts 'Usage: ruby create_release_text.rb <owner> <name>
Example: ruby create_release_text.rb getgauge gauge
'
  exit 1
end

repo_name = ARGV[0]
user_name = ARGV[1]

repo = "#{repo_name}/#{user_name}"

api  = "https://api.github.com"

latest_release = URI.join(api, "/#{repo}/releases/latest")
timestamp = JSON.parse(Net::HTTP.get(latest_release))['published_at']

issues_query = "/search/issues?q=is:pr+repo:#{repo}+state:closed"

if not timestamp.nil? || timestamp.empty? 
  issues_query += ":>#{timestamp}"
end

issues_json = Net::HTTP.get(URI.join(api, issues_query))

issues = JSON.parse(issues_json)

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
