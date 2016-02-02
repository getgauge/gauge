#!/bin/sh

# save timestamp of gauge.properties file
prefix=/usr/local
gaugePropertiesFile=$prefix/share/gauge/gauge.properties
timestamp_file=$prefix/share/gauge/timestamp.txt

rm $prefix/share/gauge/timestamp.txt
stat -f "%m" $gaugePropertiesFile > $prefix/share/gauge/timestamp.txt

# install default gauge plugins
sudo -u $USER /usr/local/bin/gauge --install html-report
