#!/bin/bash
if [ -z "$prefix" ];
then
	prefix=/usr/local
fi

echo "creating $prefix if it doesn't exist"
[ -d $prefix ] || mkdir $prefix
echo "Copying files to $prefix"
cp -rf bin share $prefix
echo "Done installing Gauge to $prefix"
echo "Installing plugin - htm-report"
gauge --install html-report
