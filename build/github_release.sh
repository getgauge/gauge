go get -v -u github.com/aktau/github-release

if [ -z "$GITHUB_TOKEN" ]; then
  echo "GITHUB_TOKEN is not set"
  exit 1
fi

version=$(ls $repoName-*-darwin.x86.zip | sed "s/^$repoName-\([^;]*\)-darwin.x86.zip/\1/")

$GOPATH/bin/github-release release -u getgauge -r $repoName --draft -t "v$version" -d "## New Features: ## Enhancements: ## Bug Fixes:" -n "$repoName $version"

for i in `ls *.zip`; do
    $GOPATH/bin/github-release -v upload -u getgauge -r $repoName -t "v$version" -n $i -f $i
done
