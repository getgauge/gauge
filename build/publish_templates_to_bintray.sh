#!/bin/sh

if [ -z "$BINTRAY_USER" ]; then
  echo "BINTRAY_USER is not set"
  exit 1
fi

if [ -z "$BINTRAY_API_KEY" ]; then
  echo "BINTRAY_API_KEY is not set"
  exit 1
fi

if [ -z "$PACKAGE" ]; then
    PACKAGE="Templates"
fi

if [ -z "$BINTRAY_PACKAGE" ]; then
    BINTRAY_PACKAGE="gauge-templates"
fi

for i in `ls *.zip`; do
  LANG=$(echo $i | cut -d '.' -f 1 | cut -d '_' -f 1);
  URL="https://api.bintray.com/content/gauge/$PACKAGE/$BINTRAY_PACKAGE/latest/$LANG/$i?publish=1&override=1"
  echo "Uploading to : $URL"

  RESPONSE_CODE=$(curl -T $i -u$BINTRAY_USER:$BINTRAY_API_KEY $URL -I -s -w "%{http_code}" -o /dev/null);
  if [[ "${RESPONSE_CODE:0:2}" != "20" ]]; then
    echo "Unable to upload, HTTP response code: $RESPONSE_CODE"
    exit 1
  fi
  echo "HTTP response code: $RESPONSE_CODE"
done;
