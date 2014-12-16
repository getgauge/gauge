#!/bin/bash
if [ -z "$prefix" ];
then
	prefix=/usr/local
  echo "creating $prefix if it doesn't exist"
  [ -d $prefix ] || mkdir $prefix
  echo "Copying files to $prefix"
  sudo /bin/sh -c "cp -rf bin share $prefix"
else
  echo "creating $prefix if it doesn't exist"
  [ -d $prefix ] || mkdir $prefix
  echo "Copying files to $prefix"
  cp -rf bin share $prefix
fi

echo "Done installing Gauge to $prefix"
echo "Installing plugin - html-report"
gauge --install html-report
