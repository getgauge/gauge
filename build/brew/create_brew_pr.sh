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

gh repo sync $HOMEBREW_GITHUB_USER_NAME/homebrew-core \
  --source Homebrew/homebrew-core \
  --branch master

(gh repo clone $HOMEBREW_GITHUB_USER_NAME/homebrew-core) || (gh repo clone Homebrew/homebrew-core && cd homebrew-core && gh repo fork Homebrew/homebrew-core)

cd homebrew-core
git checkout master

git branch -D $BRANCH || true
git checkout -b $BRANCH

gem install parser
ruby ../brew_update.rb $GAUGE_VERSION ./Formula/g/gauge.rb

git add ./Formula/g/gauge.rb
git commit -m "gauge $GAUGE_VERSION"
git push -f "https://$HOMEBREW_GITHUB_USER_NAME:$GITHUB_TOKEN@github.com/$HOMEBREW_GITHUB_USER_NAME/homebrew-core.git" "gauge-$GAUGE_VERSION"

echo -e "gauge $GAUGE_VERSION \n\n Please cc getgauge/core for any issue." > desc.txt
gh pr create --repo Homebrew/homebrew-core \
  --title "gauge $GAUGE_VERSION" \
  --body-file ./desc.txt \
  --head $HOMEBREW_GITHUB_USER_NAME:$BRANCH
