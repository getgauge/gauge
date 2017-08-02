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

YELLOW='\033[1;33m'
NC='\033[0m'

# converts a ',' separated string into list.
convert_to_list() {
    old_iFS="$IFS"
    IFS=","
    IFS=${IFS:0:1} # this is useful to format your code with tabs
    list=( $1 )
    IFS="$old_iFS"
}

# Execute gauge --iplugins=$@nstall <plugin> for a provided list
install_plugins() {
    for plugin in $@
    do
        echo "Installing plugin $plugin ..."
        $prefix/bin/gauge install $plugin
    done
    echo -e "${YELLOW}GAUGE_ROOT has been set in ~/.profile. If you face errors, run '$ source ~/.profile'\n${NC}"
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
    install_plugins "${plugins_list[@]}"
}

# Install plugins mentioned in $GAUGE_PLUGINS
install_plugins_noninteractively() {
    plugins_list=( html-report )
    if [[ ! -z "$GAUGE_PLUGINS" ]]; then
        convert_to_list $GAUGE_PLUGINS
        plugins_list=( ${list[@]} ${plugins_list[@]} )
        plugins_list=($(echo "${plugins_list[@]}" | tr ' ' '\n' | sort -u | tr '\n' ' '))
    fi
    install_plugins "${plugins_list[@]}"
}

# Find absolute path
get_absolute_path (){
    [[ -d $1 ]] && { cd "$1"; echo "$(pwd -P)"; } ||
    { cd "$(dirname "$1")" || exit 1; echo "$(pwd -P)/$(basename "$1")"; }
}

# Set GAUGE_ROOT and GAUGE binaries to environment variable
add_gauge_in_path() {
    # ensure gauge is on PATH
    if [[ "$(which gauge)" != $prefix/bin ]]; then
        echo "Adding gauge to system path..."
        echo "PATH=$PATH:$prefix/bin" >> ~/.profile
    fi

    source ~/.profile
}

# check permission for nested dir in reverse order and create non existing dir
create_nested_repo() {
    parent=$(dirname "$1")
    if [[ -d $parent ]]; then
        echo "Creating $1"
        if [[ -w $parent ]]; then
            mkdir -p $1
        else
            echo "You do not have write paermisions for '$parent ."
            sudo mkdir -p $1
        fi
    else
        create_nested_repo $parent
    fi
}



# Creates installation prefix and configuration dirs if doesn't exist
create_prefix_interactively() {
    if  [[ ! -d $prefix ]]; then
        create_nested_repo $prefix
    fi
}

# Creates installation prefix and configuration dirs if doesn't exist in non tty mode
create_prefix_noninteractively(){
     [[ -d $prefix ]] || echo "Creating $prefix ..." && mkdir -p $prefix
}

# Give option to change the permission or delete the dir if needed
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


# Copy gauge binaries in $prefix dir
copy_gauge_binaries_interactively() {
    # check for write permissions and Install gauge, asks for sudo access if not permitted
    if [[ ! -w $prefix || $prefix == "/usr/local" ]]; then
        echo "You do not have write permissions for $prefix"
        echo "Running script as sudo "
        sudo cp -rf bin $prefix
        echo "Installed gauge binaries at $prefix"
    else
        cp -rf bin $prefix
        echo "Installed gauge binaries at $prefix"
    fi
}

# Copy gauge binaries in $prefix dir
copy_gauge_binaries_noninteractively() {
    cp -rf bin $prefix
    echo "Installed gauge binaries at $prefix"
}

# Get last modified timestamp of the file
get_time_stamp() {
    if [[ `uname` != "Linux" ]]; then
        time_stamp=$(stat -f "%m" $1)
    else
        time_stamp=$(date +%s -r $1)
    fi
}

# Set prefix for installion in interactive mode
set_prefix_interavctively() {
    if [[ -z "$GAUGE_PREFIX" ]]; then
        prefix=/usr/local
        echo "Installing gauge at $prefix/bin"
        echo -e "Enter custom install location :-"
        read -e install_location
        if [[ ! -z $install_location ]]; then
            prefix=$(get_absolute_path ${install_location/\~/$HOME})
        fi
    else
        if [[ "$GAUGE_ROOT" != "" && "$GAUGE_ROOT" != "$GAUGE_PREFIX" && "$GAUGE_ROOT" != "$config" ]]; then
            echo "Previous installation was at $GAUGE_ROOT/bin. Enter [1] to use the same location or [2] to use $GAUGE_PREFIX for installation (By default it will be $GAUGE_PREFIX):-"
            read -e choice
            case $choice in
                1)
                    prefix=$GAUGE_ROOT ;;
                2)
                    prefix=$GAUGE_PREFIX ;;
                *)
                    prefix=$GAUGE_PREFIX ;;
            esac
        else
            prefix=$GAUGE_PREFIX
        fi
        echo "Installing gauge at $prefix/bin"
    fi
}

# Set prefix for installion in noninteractive mode
set_prefix_noninteractively() {
    if [[ -z $GAUGE_PREFIX ]]; then
        prefix=/usr/local
    else
        if [[ $ForceInstall ]]; then
            prefix=$GAUGE_PREFIX
        else
            if [[ "$GAUGE_ROOT" != "" && "$GAUGE_ROOT" != "/usr/local" && "$GAUGE_ROOT" != "$config" && "$GAUGE_ROOT" != "$GAUGE_PREFIX" ]]; then
                echo "Previous installation was at  $GAUGE_ROOT. Cannot proceed with installation use --force to install."
                exit 1
            else
                prefix=$GAUGE_PREFIX
            fi
        fi
    fi
}

show_post_install_messages(){
    echo -e "Gauge core successfully installed.\n"
    echo -e "We are constantly looking to make Gauge better, and report usage statistics anonymously over time. If you do not want to participate please read instructions https://manpage.getgauge.io/gauge_telemetry_off.html on how to turn it off.\n"
}

# Install Gauge interactively
install_gauge_interactively() {
    config=$HOME/.gauge/config
    set_prefix_interavctively
    create_prefix_interactively
    copy_gauge_binaries_interactively
    add_gauge_in_path
    show_post_install_messages
}

# Install gauge noninteractively
install_gauge_noninteractively() {
    config=$HOME/.gauge/config
    set_prefix_noninteractively
    create_prefix_noninteractively
    copy_gauge_binaries_noninteractively
    add_gauge_in_path
    show_post_install_messages
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


# Print usage of this script
display_usage() {
	echo -e "On Linux, this script installs gauge and it's plugins.\n\nUsage:\n./install.sh\n\nSet GAUGE_PREFIX env to install gauge at custom location.
Set GAUGE_PLUGINS env to install plugins along with gauge.
Exp:-
    GAUGE_PREFIX=my/custom/path ./install.sh
    GAUGE_PLUGINS=java,ruby,spectacle ./install.sh
    GAUGE_PREFIX=my/custom/path GAUGE_PLUGINS=xml-report,java ./install.sh"
}

# check whether user has supplied -h or --help . If yes display usage if no diplay usage with an error
if [[ $# != 0 ]]; then
    case $@ in
        "-h")
            display_usage
            exit 0 ;;
        "--help")
            display_usage
            exit 0 ;;
        "--force")
            ForceInstall=true ;;
        *)
            echo -e "unknown option $@. \n"
            display_usage
            exit 1 ;;
    esac
fi

# If tty then perform installation in interactive mode.
if [[ "$CONTINUOUS_INTEGRATION" != "true" ]] && [[ "$CI" != "true" ]] && tty -s; then
    do_interactive_installation
else
    do_noninteractive_installation
fi
