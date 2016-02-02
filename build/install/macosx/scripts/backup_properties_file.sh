#!/bin/sh

prefix=/usr/local
gaugePropertiesFile=$prefix/share/gauge/gauge.properties
timestamp_file=$prefix/share/gauge/timestamp.txt

if [ -f $timestamp_file ] ; then
    currentTimeStamp=`stat -f "%m" $gaugePropertiesFile`
    oldTimeStamp=`cat $timestamp_file`
    if [ $currentTimeStamp != $oldTimeStamp ] ; then
        backupFile=$prefix/share/gauge/gauge.properties.bak
        echo "There could be some changes in gauge.properties file. Taking a backup of it in $backupFile..."
        rm -rf $backupFile
        cat $gaugePropertiesFile > $backupFile
    fi
fi
