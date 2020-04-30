#!/bin/sh

# ----------------------------------------------------------------
#   Copyright (c) ThoughtWorks, Inc.
#   Licensed under the Apache License, Version 2.0
#   See LICENSE.txt in the project root for license information.
# ----------------------------------------------------------------

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
  URL="https://api.bintray.com/content/gauge/$PACKAGE/$BINTRAY_PACKAGE/latest/$i?publish=1&override=1"
  echo "Uploading to : $URL"

  RESPONSE_CODE=$(curl -T $i -u$BINTRAY_USER:$BINTRAY_API_KEY $URL -I -s -w "%{http_code}" -o /dev/null);
  if [[ "${RESPONSE_CODE:0:2}" != "20" ]]; then
    echo "Unable to upload, HTTP response code: $RESPONSE_CODE"
    exit 1
  fi
  echo "HTTP response code: $RESPONSE_CODE"
done;
