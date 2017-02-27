#!/bin/sh

config=$HOME/config
gaugePropertiesFile=$config/gauge.properties
timestamp_file=$config/timestamp.txt

if [ -f $timestamp_file ] ; then
    currentTimeStamp=`stat -f "%m" $gaugePropertiesFile`
    oldTimeStamp=`cat $timestamp_file`
    if [ $currentTimeStamp != $oldTimeStamp ] ; then
        backupFile=$config/gauge.properties.bak
        echo "If you have Gauge installed already and there are any manual changes in gauge.properties file, a backup of it has been taken at GAUGE_INSTALL_LOCATION\config\gauge.properties.bak or HOME\.gauge\config\gauge.properties.bak. You can restore these configurations later."
        rm -rf $backupFile
        cat $gaugePropertiesFile > $backupFile
    fi
fi
