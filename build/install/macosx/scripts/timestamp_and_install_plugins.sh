#!/bin/sh

config=$HOME/.gauge/config

# create config dir if doesn't exist
if [[ ! -d $config ]]; then
    sudo -u $USER mkdir -p $config
fi

# Copy config file from /usr/local/config to HOME/.gauge/config
sudo -u $USER cp -r /usr/local/config/* $config

# save timestamp of gauge.properties file
gauge_properties_file=$config/gauge.properties
timestamp_file=$config/timestamp.txt

rm $timestamp_file
stat -f "%m" $gauge_properties_file > $timestamp_file

# install default gauge plugins
sudo -u $USER /usr/local/bin/gauge --install html-report
