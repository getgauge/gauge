#!/bin/sh

config=$HOME/.gauge/config

# create config dir if doesn't exist
if [[ ! -d $config ]]; then
    mkdir -p $config
fi

# Copy config file from /usr/local/share to HOME/.gauge/config
cp -r /usr/local/share/gauge/* $config

# save timestamp of gauge.properties file
config=$HOME/.gauge/config
gaugePropertiesFile=$config/gauge.properties
timestamp_file=$config/timestamp.txt

rm $prefix/share/gauge/timestamp.txt
stat -f "%m" $gaugePropertiesFile > $config/timestamp.txt

# install default gauge plugins
sudo -u $USER /usr/local/bin/gauge --install html-report
