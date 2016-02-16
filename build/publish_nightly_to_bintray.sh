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

if [ "$1" == "--rename" ]; then
    RENAME=1
fi

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
        mv "$f" "`echo $f | tr '[:upper:]' '[:lower:]'`"
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

function snooze () {
    echo "\nSleeping for 30 seconds. Have a coffee..."
    sleep 30s;
    echo "Done sleeping\n"
}

function printMeta () {
    echo "Publishing package : $PACKAGE"
    echo "Version to be uploaded: $VERSION"
}

setVersion
renameToLowerCase
renameNoVersion
renameWithTimestamp
setVersion
printMeta
bintrayUpload
snooze
bintraySetDownloads
