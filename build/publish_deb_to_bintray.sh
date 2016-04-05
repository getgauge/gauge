#!/bin/sh
set -e

if [ -z "$REPO" ]; then
  REPO="gauge-deb"
fi

if [ -z "$PACKAGE" ]; then
    PACKAGE="gauge"
fi

if [ -z "$DISTRIBUTIONS" ]; then
  DISTRIBUTIONS="stable"
fi

if [ -z "$COMPONENTS" ]; then
  COMPONENTS="main"
fi

if [ -z "$USER" ]; then
  echo "USER is not set"
  exit 1
fi

if [ -z "$API_KEY" ]; then
  echo "API_KEY is not set"
  exit 1
fi


PACKAGE_FILE_PREFIX=$(echo "$PACKAGE" | tr '[:upper:]' '[:lower:]');

setVersion () {
    #VERSION=$(ls $PACKAGE_FILE_PREFIX*.deb | head -1 | sed "s/\.[^\.]*$//" | sed "s/$PACKAGE_FILE_PREFIX-//" | sed "s/-[a-z]*\.[a-z0-9_]*$//");
    VERSION=$(ls $PACKAGE_FILE_PREFIX*.deb | head -1 | cut -d '-' -f 2)
}

getArchFromFileName () {
    ARCH=$(echo $1 | sed "s/$PACKAGE_FILE_PREFIX-//" | rev | cut -d '-' -f 1 | rev | cut -d '.' -f 1);
    echo $ARCH
}

setUploadDirPath () {
    UPLOADDIRPATH="pool/$COMPONENTS/$(echo $PACKAGE | head -c1)"
}

bintrayUpload () {
    for i in `ls $PACKAGE_FILE_PREFIX*.deb`; do
		ARCH=$( getArchFromFileName $i )
        URL="https://api.bintray.com/content/gauge/$REPO/$PACKAGE/$VERSION/$UPLOADDIRPATH/$i;deb_distribution=$DISTRIBUTIONS;deb_component=$COMPONENTS;deb_architecture=$ARCH?publish=1&override=1"

        echo "Uploading:"
        echo "\tversion: $VERSION"
        echo "\tarch: $ARCH"
        echo "\tURL: $URL"

        RESPONSE_CODE=$(curl -T $i -u$USER:$API_KEY $URL -I -s -w "%{http_code}" -o /dev/null);
        if [ "$(echo $RESPONSE_CODE | head -c2)" != "20" ]; then
            echo "Unable to upload, HTTP response code: $RESPONSE_CODE"
            exit 1
        fi
        echo "HTTP response code: $RESPONSE_CODE"
    done;
}

bintraySetDownloads () {
    for i in `ls *.deb`; do
        ARCH=$( getArchFromFileName $i )
        URL="https://api.bintray.com/file_metadata/gauge/$REPO/$UPLOADDIRPATH/$i"

        echo "Putting $i in $PACKAGE's download list"
        RESPONSE_CODE=$(curl -X PUT -d "{ \"list_in_downloads\": true }" -H "Content-Type: application/json" -u$USER:$API_KEY $URL -s -w "%{http_code}" -o /dev/null);

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
    echo "Publishing deb: $PACKAGE"
    echo "Version to be uploaded: $VERSION"
}

setVersion
printMeta
setUploadDirPath
bintrayUpload
snooze
bintraySetDownloads
