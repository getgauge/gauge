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

set -e

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

if [ -z "$version" ]; then
  version=$(ls $artifactName* | head -1 | sed "s/\.[^\.]*$//" | sed "s/$artifactName-//" | sed "s/-[a-z]*\.[a-z0-9_]*$//");
fi

go get github.com/github/hub
hub --version

dest="/tmp"

function clone_and_change_repo() {
    rm -rf "$dest/$repoName"
    echo "-----------------------------------------"
    echo "Cloning $githubUser/$repoName to $dest/$repoName"
    echo "-----------------------------------------"
    hub clone "$githubUser/$repoName" "$dest/$repoName" --depth 1
    cd "$dest/$repoName"
}

function draft_a_release() {
    artifacts=()
    dir=`pwd`
    for i in `ls`; do
        artifacts+="$dir/$i "
    done

    clone_and_change_repo

    echo "-----------------------------------------"
    echo "Drafting release v$version for $repoName "
    echo "-----------------------------------------"

    echo "$repoName v$version\n\n" > desc.txt

    release_description=$(ruby -e "$(curl -sSfL https://github.com/getgauge/gauge/raw/master/build/create_release_text.rb)" $repoName $githubUser)
    echo "$release_description" >> desc.txt

    hub release create -d -F ./desc.txt $version

    rm -rf desc.txt

    if [ -z "$uploadArtifact" -o "$uploadArtifact" == "yes" ]; then
        echo "Start uploading ..."
        for i in `ls $artifacts`; do
            hub release edit -m "" -a $i $version
            if [ $? -ne 0 ];then
                exit 1
            fi
        done
    else
        echo "Avoiding uploading artifact as uploadArtifact is not set to yes."
    fi
    rm -rf "$dest/$repoName"
}

function publish_a_release() {
    clone_and_change_repo
    echo "-------------------------------------------"
    echo "Publishing release v$version for $repoName "
    echo "-------------------------------------------"
    hub release edit -m "" --draft=false $version
}

if [ -z "$releaseType" -o "$releaseType" == "draft" ]; then
  draft_a_release
else
  publish_a_release
fi