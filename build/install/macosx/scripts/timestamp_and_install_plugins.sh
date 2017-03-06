#!/bin/sh

config=$HOME/.gauge/config

# create config dir if doesn't exist
if [[ ! -d $config ]]; then
    sudo -u $USER mkdir -p $config
fi

# Copy config file from /usr/local/gauge/config to HOME/.gauge/config
sudo -u $USER cp -r /usr/local/gauge/config/* $config
sudo rm -rf /usr/local/gauge

# save timestamp of gauge.properties file
gauge_properties_file=$config/gauge.properties
timestamp_file=$config/timestamp.txt

rm $timestamp_file
stat -f "%m" $gauge_properties_file > $timestamp_file
