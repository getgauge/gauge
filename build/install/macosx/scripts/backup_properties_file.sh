#!/bin/sh

config=$HOME/.gauge/config
gauge_properties_file=$config/gauge.properties
timestamp_file=$config/timestamp.txt

if [ -f $timestamp_file ] ; then
    current_time_stamp=`stat -f "%m" $gauge_properties_file`
    old_time_stamp=`cat $timestamp_file`
    if [ $current_time_stamp != $old_time_stamp ] ; then
        backup_file=$config/gauge.properties.bak
        echo "If you have Gauge installed already and there are any manual changes in gauge.properties file, a backup of it has been taken at HOME\.gauge\config\gauge.properties.bak. You can restore these configurations later."
        rm -rf $backup_file
        cat $gauge_properties_file > $backup_file
    fi
fi
