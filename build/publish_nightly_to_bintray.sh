#!/bin/sh
set -e

if [ -z "$PACKAGE" ]; then
  echo "PACKAGE is not set"
  exit 1
fi

if [ -z "$PACKAGE_TYPE" ]; then
  echo "PACKAGE_TYPE is not set"
  exit 1
fi

if [ -z "$BINTRAY_USER" ]; then
  echo "BINTRAY_USER is not set"
  exit 1
fi

if [ -z "$BINTRAY_API_KEY" ]; then
  echo "BINTRAY_API_KEY is not set"
  exit 1
fi

if [ -z "$BINTRAY_PACKAGE" ]; then
    BINTRAY_PACKAGE="Nightly"
fi

if [ -z "$RENAME" ]; then
    RENAME=0
fi

if [ -z "$GITHUB_SSH_PRIVATE_KEY" ]; then
  echo "GITHUB_SSH_PRIVATE_KEY is not set"
  exit 1
fi

command -v jq >/dev/null 2>&1 || { echo >&2 "jq is not installed, aborting."; exit 1; }

PACKAGE_FILE_PREFIX=$(echo $PACKAGE | tr '[:upper:]' '[:lower:]')

function setVersion () {
    VERSION=$(ls $PACKAGE_FILE_PREFIX* | head -1 | sed "s/\.[^\.]*$//" | sed "s/$PACKAGE_FILE_PREFIX-//" | sed "s/-[a-z]*\.[a-z0-9_]*$//");
}

function renameNoVersion () {
    if [ "$NOVERSION" == "1" ]; then
        VERSION="latest"
        echo "Not checking for package version"
        for f in $PACKAGE_FILE_PREFIX*;
        do mv "$f" "`echo $f | sed s/$PACKAGE_FILE_PREFIX/$PACKAGE_FILE_PREFIX-$VERSION/`";
        done
    else
        if [ -z "$VERSION" ]; then
            echo "Could not determine $PACKAGE version"
            exit 1
        fi
    fi
}

function renameWithTimestamp () {
    if [ "$RENAME" == "1" ]; then
        CURR_DATE=$(date +"%Y-%m-%d")

        for f in $PACKAGE_FILE_PREFIX*;
        do mv "$f" "`echo $f | sed s/$VERSION/$VERSION.$PACKAGE_TYPE-$CURR_DATE/`";
        done
    fi
}

function renameToLowerCase () {
    for f in `ls`; do
        NEW_FILENAME=$(echo $f | tr '[:upper:]' '[:lower:]')
        if [ "$f" != "$NEW_FILENAME" ]; then
            mv "$f" "$NEW_FILENAME"
        fi
    done
}

function getPlatformFromFileName () {
    if [ "$NOPLATFORM" == "1" ]; then
        echo ""
    else
        PLATFORM=$(echo $1 | sed "s/$PACKAGE_FILE_PREFIX-//" | rev | cut -d '-' -f 1 | rev | cut -d '.' -f 1);
        echo "$PLATFORM/"
    fi
}

function bintrayUpload () {
    for i in `ls`; do
        PLATFORM=$( getPlatformFromFileName $i )
        URL="https://api.bintray.com/content/gauge/$PACKAGE/$BINTRAY_PACKAGE/$VERSION/$PLATFORM$i?publish=1&override=1"

        echo "Uploading to : $URL"

        RESPONSE_CODE=$(curl -T $i -u$BINTRAY_USER:$BINTRAY_API_KEY $URL -I -s -w "%{http_code}" -o /dev/null);
        if [[ "${RESPONSE_CODE:0:2}" != "20" ]]; then
            echo "Unable to upload, HTTP response code: $RESPONSE_CODE"
            exit 1
        fi
        echo "HTTP response code: $RESPONSE_CODE"
        if [[ $PLATFORM == "" ]]; then
            eval "PLATFORM_INDEPENDENT_FILE=https://dl.bintray.com/gauge/$PACKAGE/$i"
        elif [[ $i == *"x86_64"* ]]; then
            eval "${PLATFORM:0:${#PLATFORM}-1}_x86_64=https://dl.bintray.com/gauge/$PACKAGE/$PLATFORM$i"
        else
            eval "${PLATFORM:0:${#PLATFORM}-1}_x86=https://dl.bintray.com/gauge/$PACKAGE/$PLATFORM$i"
        fi
    done;
}

function bintraySetDownloads () {
    for i in `ls`; do
        PLATFORM=$( getPlatformFromFileName $i )
        URL="https://api.bintray.com/file_metadata/gauge/$PACKAGE/$PLATFORM$i"

        echo "Putting $i in $PACKAGE's download list"
        RESPONSE_CODE=$(curl -X PUT -d "{ \"list_in_downloads\": true }" -H "Content-Type: application/json" -u$BINTRAY_USER:$BINTRAY_API_KEY $URL -s -w "%{http_code}" -o /dev/null);
        if [[ "${RESPONSE_CODE:0:2}" != "20" ]]; then
            echo "Unable to put in download list, HTTP response code: $RESPONSE_CODE"
            exit 1
        fi
        echo "HTTP response code: $RESPONSE_CODE"
    done
}

function cleanOldNightlyVersions() {
    URL="https://api.bintray.com/packages/gauge/$PACKAGE/$BINTRAY_PACKAGE"
    versions=($(curl -X GET -H "Content-Type: application/json" -u$BINTRAY_USER:$BINTRAY_API_KEY $URL | jq -r '.versions'))
    for v in ${versions[@]:11}; do
        version=$(echo $v | sed -e 's/,//' -e 's/"//g')
        if [ $version !=  "]" ]; then
            echo "Deleting version: $version"
            DELETE_URL="$URL/versions/$version"
            RESPONSE_CODE=$(curl -X DELETE -H "Content-Type: application/json" -u$BINTRAY_USER:$BINTRAY_API_KEY $DELETE_URL -s -w "%{http_code}" -o /dev/null);
            if [[ "${RESPONSE_CODE:0:2}" != "20" ]]; then
                echo "Unable to delete version : $v, HTTP response code: $RESPONSE_CODE"
                exit 1
            fi
            echo "HTTP response code: $RESPONSE_CODE"
        fi
    done;
}

function snooze () {
    echo "\nSleeping for 30 seconds. Have a coffee..."
    sleep 30s;
    echo "Done sleeping\n"
}

function printMeta () {
    echo "Publishing package : $PACKAGE"
    echo "Version to be uploaded: $VERSION"
}

function updateRepo () {
    if [ "$UPDATE_INSTALL_JSON" != "1" ]; then
        return 0
    fi
    eval $(ssh-agent) && echo -e \"$GITHUB_SSH_PRIVATE_KEY\" | ssh-add -
    git clone git@github.com:getgauge/gauge-nightly-repository.git
    cd gauge-nightly-repository
    if [[ $PLATFORM_INDEPENDENT_FILE != "" ]]; then
            windows_x86=$PLATFORM_INDEPENDENT_FILE
            windows_x86_64=$PLATFORM_INDEPENDENT_FILE
            linux_x86=$PLATFORM_INDEPENDENT_FILE
            linux_x86_64=$PLATFORM_INDEPENDENT_FILE
            darwin_x86=$PLATFORM_INDEPENDENT_FILE
            darwin_x86_64=$PLATFORM_INDEPENDENT_FILE
    fi
    versionInfo="[{\"version\": \"$VERSION\",\"gaugeVersionSupport\": {\"minimum\": \"0.6.1\",\"maximum\": \"\"},\"install\": {\"windows\": [],\"linux\": [],\"darwin\": []},\"DownloadUrls\": {\"x86\":
{\"windows\": "\"$windows_x86\"", \"linux\": "\"$linux_x86\"",\"darwin\": "\"$darwin_x86\""},\"x64\": {\"windows\": "\"$windows_x86_64\"",\"linux\": "\"$linux_x86_64\"",\"darwin\": "\"$darwin_x86_64\""}}}]"
    if [ -z "$INSTALL_PLUGIN_JSON" ]; then
        echo "INSTALL_PLUGIN_JSON is not set"
        exit 1
    fi

    json=`cat $INSTALL_PLUGIN_JSON | jq ".versions=$versionInfo" $INSTALL_PLUGIN_JSON`
    echo $json | jq . > "$INSTALL_PLUGIN_JSON"
    git add "$INSTALL_PLUGIN_JSON"
    git commit -m "Updating nightly version for $INSTALL_PLUGIN_JSON"
    git push origin master
    cd ../
    rm -rf gauge-nightly-repository
}

renameToLowerCase
setVersion
renameNoVersion
renameWithTimestamp
setVersion
printMeta
bintrayUpload
updateRepo
snooze
bintraySetDownloads
cleanOldNightlyVersions
