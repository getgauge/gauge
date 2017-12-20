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
  puts 'Usage: ruby create_release_text.rb <repo name>
Example: ruby create_release_text.rb gauge.
'
  exit 1
end

repo_name = ARGV[0]
user_name = ARGV[1]
uri = URI("https://api.github.com/repos/#{user_name}/#{repo_name}/releases/latest")
timestamp = JSON.parse(Net::HTTP.get(uri))['published_at']

if timestamp.nil? || timestamp.empty? 
    uri = URI("https://api.github.com/search/issues?q=repo:#{user_name}/#{repo_name}+state:closed")
else
    uri = URI("https://api.github.com/search/issues?q=repo:#{user_name}/#{repo_name}+closed:>#{timestamp}")
end
response = Net::HTTP.get(uri)

data = JSON.parse(response)

categories = {"new feature" => [], "enhancement" => [], "bug" => [], "misc" => []}
headers = {"new feature" => "New Features", "enhancement" => "Enhancements", "bug" => "Bug Fixes", "misc" => "Miscellaneous"}

data['items'].each do |i|
    issue_text = "- ##{i['number']} - #{i['title']}"
    issue_labels = i['labels'].map {|x| x['name']}
    if (issue_labels & ["moved", "duplicate"]).empty?
        label = issue_labels & categories.keys
        categories[label[0] || "misc"] << issue_text
    end
end

categories.each_key do |k|
    puts "## #{headers[k]}\n\n"
    categories[k].each {|v| puts v}
    puts "\n" 
end