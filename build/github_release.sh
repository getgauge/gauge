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

if [ -z "$repoName" ]; then
  echo "repoName is not set"
  exit 1
fi

if [ -z "$GITHUB_TOKEN" ]; then
  echo "GITHUB_TOKEN is not set"
  exit 1
fi

if [ -z "$GITHUB_SSH_PRIVATE_KEY" ]; then
  echo "GITHUB_SSH_PRIVATE_KEY is not set"
  exit 1
fi

eval $(ssh-agent) && echo -e \"$GITHUB_SSH_PRIVATE_KEY\" | ssh-add -

go get -v -u github.com/aktau/github-release

version=$(ls $repoName* | head -1 | sed "s/\.[^\.]*$//" | sed "s/$repoName-//" | sed "s/-[a-z]*\.[a-z0-9_]*$//");
echo "------------------------------"
echo "Releasing $repoName v$version"
echo "------------------------------"

release_description=$(ruby -e "$(curl -sSfL https://github.com/getgauge/gauge/raw/master/build/create_release_text.rb)" $repoName)

$GOPATH/bin/github-release release -u getgauge -r $repoName --draft -t "v$version" -d "$release_description" -n "$repoName $version"

for i in `ls`; do
    $GOPATH/bin/github-release -v upload -u getgauge -r $repoName -t "v$version" -n $i -f $i
done
