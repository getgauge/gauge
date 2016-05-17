#!/bin/sh
set -e

if [ -z "$REPO" ]; then
  REPO="gauge-rpm"
fi

if [ -z "$PACKAGE" ]; then
    PACKAGE="gauge"
fi

if [ -z "$BINTRAY_PACKAGE" ]; then
    PACKAGE="gauge-nightly"
fi

if [ -z "$USER" ]; then
  echo "USER is not set"
  exit 1
fi

if [ -z "$API_KEY" ]; then
  echo "API_KEY is not set"
  exit 1
fi

if [ -z "$PASSPHRASE" ]; then
  echo "PASSPHRASE is not set"
  exit 1
fi


PACKAGE_FILE_PREFIX=$(echo "$PACKAGE" | tr '[:upper:]' '[:lower:]');

setVersion () {
    VERSION=$(ls $PACKAGE_FILE_PREFIX*.rpm | head -1 | sed "s/\.[^\.]*$//" | sed "s/$PACKAGE_FILE_PREFIX-//" | sed "s/-[a-z]*\.[a-z0-9_]*$//" | rev | sed "s/^[a-z0-9]*-//" | rev);
}

getArchFromFileName () {
    ARCH=$(echo $1 | sed "s/$PACKAGE_FILE_PREFIX-//" | rev | cut -d '-' -f 1 | rev | cut -d '.' -f 1);
    echo $ARCH
}

setUploadDirPath () {
    UPLOADDIRPATH="$BINTRAY_PACKAGE/$PACKAGE/$VERSION"
}

bintrayUpload () {
    for i in `ls $PACKAGE_FILE_PREFIX*.rpm`; do
		ARCH=$( getArchFromFileName $i )
        URL="https://api.bintray.com/content/gauge/$REPO/$BINTRAY_PACKAGE/$VERSION/$UPLOADDIRPATH/$i?publish=1&override=1"

        echo "Uploading:"
        echo "\tversion: $VERSION"
        echo "\tarch: $ARCH"
        echo "\tURL: $URL"

        RESPONSE_CODE=$(curl -H "X-GPG-PASSPHRASE: $PASSPHRASE" -T $i -u$USER:$API_KEY "$URL" -I -s -w "%{http_code}" -o /dev/null);
        if [ "$(echo $RESPONSE_CODE | head -c2)" != "20" ]; then
            echo "Unable to upload, HTTP response code: $RESPONSE_CODE"
            exit 1
        fi
        echo "HTTP response code: $RESPONSE_CODE"
    done;
}

bintraySetDownloads () {
    for i in `ls *.rpm`; do
        ARCH=$( getArchFromFileName $i )
        URL="https://api.bintray.com/file_metadata/gauge/$REPO/$UPLOADDIRPATH/$i"

        echo "Putting $i in $PACKAGE's download list"
        RESPONSE_CODE=$(curl -X PUT -d "{ \"list_in_downloads\": true }" -H "Content-Type: application/json" -u$USER:$API_KEY "$URL" -s -w "%{http_code}" -o /dev/null);

        if [ "$(echo $RESPONSE_CODE | head -c2)" != "20" ]; then
            echo "Unable to put in download list, HTTP response code: $RESPONSE_CODE"
            exit 1
        fi
        echo "HTTP response code: $RESPONSE_CODE"
    done
}

snooze () {
    echo "\nSleeping for 30 seconds. Have a coffee..."
    sleep 30s;
    echo "Done sleeping\n"
}

printMeta () {
    echo "Publishing rpm: $PACKAGE"
    echo "Version to be uploaded: $VERSION"
}

setVersion
printMeta
setUploadDirPath
bintrayUpload
snooze
bintraySetDownloads
