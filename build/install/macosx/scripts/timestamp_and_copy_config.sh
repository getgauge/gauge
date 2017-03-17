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
time_stamp=`stat -f "%m" $gauge_properties_file`

# remove older timestamp create new timestamp of gauge.properties file.
rm $timestamp_file
sudo -u $USER sh -c "echo $time_stamp > $config/timestamp.txt"
