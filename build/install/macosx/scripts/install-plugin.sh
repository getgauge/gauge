#!/bin/sh
plugin_name=$INSTALL_PKG_SESSION_ID

if [ "$plugin_name" == "c#" ]
then
    plugin_name="csharp"
fi

sudo -u $USER /usr/local/bin/gauge install $plugin_name