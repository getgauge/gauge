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

if [ -z "$githubUser" ]; then
  echo "userName is not set. using getgauge as default."
  githubUser="getgauge"
fi

if [ -z "$repoName" ]; then
  echo "repoName is not set"
  exit 1
fi

if [ -z "$artifactName" ]; then
  echo "artifactName is not set"
  artifactName=$repoName
fi

if [ -z "$GITHUB_TOKEN" ]; then
  echo "GITHUB_TOKEN is not set"
  exit 1
fi

go get -v -u github.com/aktau/github-release

version=$(ls $artifactName* | head -1 | sed "s/\.[^\.]*$//" | sed "s/$artifactName-//" | sed "s/-[a-z]*\.[a-z0-9_]*$//");
echo "------------------------------"
echo "Releasing $repoName v$version"
echo "------------------------------"

release_description=$(ruby -e "$(curl -sSfL https://github.com/getgauge/gauge/raw/master/build/create_release_text.rb)" $repoName $githubUser)

$GOPATH/bin/github-release release -u $githubUser -r $repoName --draft -t "v$version" -d "$release_description" -n "$repoName $version"

for i in `ls`; do
    $GOPATH/bin/github-release -v upload -u $githubUser -r $repoName -t "v$version" -n $i -f $i
    if [ $? -ne 0 ];then
        exit 1
    fi
done
