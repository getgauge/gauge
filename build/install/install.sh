#!/bin/bash

# Copyright 2015 ThoughtWorks, Inc.

# This file is part of Gauge.

# Gauge is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.

# Gauge is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.

# You should have received a copy of the GNU General Public License
# along with Gauge.  If not, see <http://www.gnu.org/licenses/>.
set -e


# converts a ',' separated string into list.
convert_to_list() {
    old_iFS="$IFS"
    IFS=","
    IFS=${IFS:0:1} # this is useful to format your code with tabs
    list=( $1 )
    IFS="$old_iFS"
}

# Install all the plugins in interactive mode.
install_plugins_interactively() {
    plugins_list=( html-report )
    if [[ -z "$GAUGE_PLUGINS" ]]; then
        echo "Enter comma(',') separated list of plugins which you would like to install :- "
        read -e plugins
        if [[ ! -z $plugins ]]; then
            convert_to_list $plugins
            plugins_list=( ${list[@]} ${plugins_list[@]} )
            plugins_list=($(echo "${plugins_list[@]}" | tr ' ' '\n' | sort -u | tr '\n' ' '))
        fi
    else
        convert_to_list $GAUGE_PLUGINS
        plugins_list=( ${list[@]} ${plugins_list[@]} )
        plugins_list=($(echo "${plugins_list[@]}" | tr ' ' '\n' | sort -u | tr '\n' ' '))
    fi
    for plugin in "${plugins_list[@]}"
    do
        echo "Installing plugin $plugin ..."
        gauge --install $plugin
    done
}

# Print usage of this script
display_usage() {
	echo -e "On Linux, this script installs gauge and it's plugins.\n\nUsage:\n./install.sh\n\nSet GAUGE_PREFIX env to install gauge at custom location.
Set GAUGE_PLUGINS env to install plugins along with gauge.
Exp:-
    GAUGE_PREFIX=my/custom/path ./install.sh
    GAUGE_PLUGINS=java,ruby,spectacle ./install.sh
    GAUGE_PREFIX=my/custom/path GAUGE_PLUGINS=xml-report,java ./install.sh"
}

# Find absolute path
get_absolute_path (){
    [[ -d $1 ]] && { cd "$1"; echo "$(pwd -P)"; } ||
    { cd "$(dirname "$1")" || exit 1; echo "$(pwd -P)/$(basename "$1")"; }
}

# Set GAUGE_ROOT environment variable
set_gaugeroot() {
    # ensure gauge is on PATH
    if [[ -z "$(which gauge)" ]]; then
        echo "Adding gauge to system path..."
        echo "PATH=$PATH:$prefix/bin" >> ~/.profile
        updated_profile=1
    fi
    # ensure GAUGE_ROOT is set
    if [[ -z "$GAUGE_ROOT" ]]; then
        echo "Adding GAUGE_ROOT to environment..."
        echo "export GAUGE_ROOT=$config"  >> ~/.profile
        updated_profile=1
    fi
    if [[ $updated_profile ]] ; then
        source ~/.profile
    fi
    echo -e "GAUGE_ROOT has been set. If you face errors, run '$ source ~/.profile'\n"
}


# Creates installation prefix and configuration dirs if doesn't exist
create_prefix_if_does_not_exist() {
    [[ -d $prefix ]] || echo "Creating $prefix ..." && mkdir -p $prefix
}


change_permission_if_needed() {
    if [[ -d $HOME/.gauge && ! -w $HOME/.gauge ]]; then
        echo "The dir .gauge already exist but was created with eleveted permision."
        echo "Enter [1] to change permissions or Enter [2] to delete the dir (By default it will change the permissions)"
        read -e choice
        group=`id -ng`
        case $choice in
            1)
                sudo chown -R $USER:$group $HOME/.gauge ;;
            2)
                sudo rm -rf ~/.gauge ;;
            *)
                sudo chown -R $USER:$group $HOME/.gauge ;;
        esac
    fi
}

# Creates configuration dirs in interactive mode if doesn't exist
create_config_interactively_if_does_not_exist() {
    if [[ ! -d $config ]]; then
        echo "Creating $config ..."
        change_permission_if_needed
        mkdir -p $config
    fi
}

check_permissions() {
    if [[ -d $HOME/.gauge && ! -w $HOME/.gauge ]]; then
            echo "The dir .gauge already exist but was created with eleveted permission. Please change paermissions for $HOME/.gauge dir or delete it."
            exit 1
    fi
}

# Creates installation prefix and configuration dirs if doesn't exist. Exits if not able to create.
create_config_if_does_not_exist() {
    if [[ ! -d $config ]]; then
        echo "Creating $config ..."
        check_permissions
        mkdir -p $config
    fi
}

# Copy gauge binaries in $prefix dir
copy_gauge_binaries() {
    # check for write permissions and Install gauge, asks for sudo access if not permitted
    if [ ! -w $prefix -a $prefix = "/usr/local" ]; then
        echo
        echo "You do not have write permissions for $prefix"
        echo "Running script as sudo "
        sudo cp -rf bin $prefix
        echo "Installed gauge binaries at $prefix"
    else
        cp -rf bin $prefix
        echo "Installed gauge binaries at $prefix"
    fi
}

get_time_stamp() {
    if [[ `uname` != "Linux" ]]; then
        time_stamp=$(stat -f "%m" $1)
    else
        time_stamp=$(date +%s -r $1)
    fi
}

# copy gauge configuration at $config
copy_gauge_configuration_files() {
    gauge_properties_file=$config/gauge.properties
    if [[ -f $config/timestamp.txt ]]; then
        current_time_stamp=`get_time_stamp $gauge_properties_file`
        old_time_stamp=`cat $config/timestamp.txt`
        if [[ $current_time_stamp != $old_time_stamp ]]; then
            backup_file=$config/gauge.properties.bak
            echo "If you have Gauge installed already and there are any manual changes in gauge.properties file, a backup of it has been taken at HOME/.gauge/config/gauge.properties.bak. You can restore these configurations later."
            rm -rf $backup_file
            cat $gauge_properties_file > $backup_file
        fi
    fi
    cp -rf config/* $config
    get_time_stamp $gauge_properties_file > $config/timestamp.txt
}

# Do the installation
install_gauge_interactively() {
    if [[ -z "$GAUGE_PREFIX" ]]; then
        prefix=/usr/local
        echo "Installing gauge at $prefix/bin"
        echo -e "Enter custom install location :-"
        read -e install_location
        if [[ ! -z $install_location ]]; then
            prefix=$(get_absolute_path ${install_location/\~/$HOME})
        fi
    else
        prefix=$GAUGE_PREFIX
    fi

    config=$HOME/.gauge/config

    create_prefix_if_does_not_exist
    copy_gauge_binaries
    create_config_interactively_if_does_not_exist
    copy_gauge_configuration_files
    set_gaugeroot
    source ~/.profile
    echo -e "Gauge core successfully installed.\n"
}

install_gauge_noninteractively() {
    if [[ -z "$GAUGE_PREFIX" ]]; then
        prefix=/usr/local
        echo "Installing gauge at $prefix/bin"
    else
        prefix=$GAUGE_PREFIX
    fi

    config=$HOME/.gauge/config
    create_prefix_if_does_not_exist
    copy_gauge_binaries
    create_config_if_does_not_exist
    copy_gauge_configuration_files
    set_gaugeroot
    source ~/.profile
    echo -e "Gauge core successfully installed.\n"
}

install_plugins_noninteractively() {
    plugins_list=( html-report )
    if [[ ! -z "$GAUGE_PLUGINS" ]]; then
        convert_to_list $GAUGE_PLUGINS
        plugins_list=( ${list[@]} ${plugins_list[@]} )
        plugins_list=($(echo "${plugins_list[@]}" | tr ' ' '\n' | sort -u | tr '\n' ' '))
    fi
    for plugin in "${plugins_list[@]}"
    do
        echo "Installing plugin $plugin ..."
        gauge --install $plugin
    done
}


# perform gauge installation in interactives mode
do_interactive_installation() {
    install_gauge_interactively
    install_plugins_interactively
}


# perform gauge installation in non tty mode
do_noninteractive_installation() {
    install_gauge_noninteractively
    install_plugins_noninteractively
}


# check whether user has supplied -h or --help . If yes display usage if no diplay usage with an error
if [[ $# != 0 ]]; then
    if [[ ( $@ == "--help") || $@ == "-h" ]]
    then
        display_usage
        exit 0
    else
        echo -e "unknown option $@. \n"
        display_usage
        exit 1
    fi
fi


# If tty then perform installation in interactive mode.
if tty -s; then
    do_interactive_installation
else
    do_noninteractive_installation
fi
