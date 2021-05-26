# ----------------------------------------------------------------
#   Copyright (c) ThoughtWorks, Inc.
#   Licensed under the Apache License, Version 2.0
#   See LICENSE.txt in the project root for license information.
# ----------------------------------------------------------------*/

#!/bin/bash

set -e

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
git commit -m "gauge $GAUGE_VERSION"
git push "https://$HOMEBREW_GITHUB_USER_NAME:$GITHUB_TOKEN@github.com/$HOMEBREW_GITHUB_USER_NAME/homebrew-core.git" "gauge-$GAUGE_VERSION"
echo -e "gauge $GAUGE_VERSION \n\n Please cc getgauge/core for any issue." > desc.txt
hub pull-request --base Homebrew:master --head $HOMEBREW_GITHUB_USER_NAME:$BRANCH -F ./desc.txt
