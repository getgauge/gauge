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

#!/bin/bash

set -e

export GITHUB_TOKEN="$HOMEBREW_GITHUB_API_TOKEN"
export BRANCH="gauge-$GAUGE_VERSION"

git config --global user.name "$HOMEBREW_GITHUB_USER_NAME"
git config --global user.email "$HOMEBREW_GITHUB_USER_EMAIL"


(hub clone homebrew-core) || (hub clone Homebrew/homebrew-core && cd homebrew-core && hub fork)

cd homebrew-core
git remote add upstream https://github.com/Homebrew/homebrew-core.git
git fetch upstream
git checkout master
git merge upstream/master

git branch -D $BRANCH || true
git checkout -b $BRANCH

gem install parser
ruby ../brew_update.rb $GAUGE_VERSION ./Formula/gauge.rb

git add ./Formula/gauge.rb
git commit -m "gauge v$GAUGE_VERSION"
git push "https://$HOMEBREW_GITHUB_USER_NAME:$GITHUB_TOKEN@github.com/$HOMEBREW_GITHUB_USER_NAME/homebrew-core.git" "gauge-$GAUGE_VERSION"
echo -e "gauge v$GAUGE_VERSION \n\n Please cc getgauge/core for any issue." > desc.txt
hub pull-request --base Homebrew:master --head $HOMEBREW_GITHUB_USER_NAME:$BRANCH -F ./desc.txt
