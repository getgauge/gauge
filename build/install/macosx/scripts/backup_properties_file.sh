#!/bin/sh

prefix=/usr/local
gaugePropertiesFile=$prefix/share/gauge/gauge.properties
timestamp_file=$prefix/share/gauge/timestamp.txt

if [ -f $timestamp_file ] ; then
    currentTimeStamp=`stat -f "%m" $gaugePropertiesFile`
    oldTimeStamp=`cat $timestamp_file`
    if [ $currentTimeStamp != $oldTimeStamp ] ; then
        backupFile=$prefix/share/gauge/gauge.properties.bak
        echo "If you have Gauge installed already and there are any manual changes in gauge.properties file, a backup of it has been taken at GAUGE_INSTALL_LOCATION\share\gauge\gauge.properties.bak. You can restore these configurations later."
        rm -rf $backupFile
        cat $gaugePropertiesFile > $backupFile
    fi
fi
