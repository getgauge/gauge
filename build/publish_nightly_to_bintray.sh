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

echo "Publishing package : $PACKAGE"

VERSION=$(ls $PACKAGE-* | head -1 | sed "s/$PACKAGE-//" | cut -d '-' -f 1);

if [ -z "$VERSION" ]; then
  echo "Could not determine $PACKAGE version"
  exit 1
fi

echo "Version to be uploaded: $VERSION"

CURR_DATE=$(date +"%Y-%m-%d")

for f in $PACKAGE*;
  do mv "$f" "`echo $f | sed s/$VERSION/$VERSION.$PACKAGE_TYPE-$CURR_DATE/`";
done

for i in `ls`; do
  PLATFORM=$(echo $i | sed "s/$PACKAGE-//" | rev | cut -d '-' -f 1 | rev | cut -d '.' -f 1);
  URL="https://api.bintray.com/content/gauge/$PACKAGE/Nightly/$VERSION.$PACKAGE_TYPE-$CURR_DATE/$PLATFORM/$i?publish=1&override=1"
  echo "Uploading to : $URL"
  RESPONSE_CODE=$(curl -T $i -u$BINTRAY_USER:$BINTRAY_API_KEY $URL -I -s -w "%{http_code}" -o /dev/null);
  if [[ "${RESPONSE_CODE:0:2}" != "20" ]]; then
    echo "Unable to upload, HTTP response code: $RESPONSE_CODE"
    exit 1
  fi
  echo "HTTP response code: $RESPONSE_CODE"
done;
