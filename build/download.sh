#!/bin/sh

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

download() {
    CWD="$(ensure pwd)"
    # Create tmp directory
    tmp="$(ensure mktemp -d)"
    ensure cd "$tmp"

    # Download latest Gauge zip installer
    url="$(ensure curl -Ss https://api.github.com/repos/getgauge/gauge/releases/latest \
    | grep "browser_download_url.*$OS.$ARCH.zip" \
    | cut -d : -f 2,3 \
    | tr -d \")"

    say "Downloading binary from URL:$url"
    ensure curl -L -o gauge.zip $url
    verbose_say "Downloaded the binary to dir:$tmp"

    # unzip and copy the binary to the original directory
    ensure unzip -q gauge.zip

    verbose_say "Copying the binary to $LOCATION"
    if [ -w $LOCATION ]; then
        mkdir -p "$LOCATION"
        ensure cp ./gauge "$LOCATION/"
    else
        echo "You do not have write permision for $LOCATION. Trying with sudo."
        sudo cp ./bin/gauge "$LOCATION/"
    fi

    ensure cd "$CWD"
    verbose_say "Cleaning up..."
    ensure rm -rf "$tmp"

    if $_ansi_escapes_are_valid; then
        printf "\33[1mDownloaded the binary to '$LOCATION'.\33[0m\n" 1>&2
        printf "\33[1mMake sure the above path is available in 'PATH' environment variable.\33[0m\n"
    else
        say "Downloaded the binary to '$LOCATION'."
        say "Make sure the above path is available in 'PATH' environment variable."
    fi
}

set_os_architecture() {
    verbose_say "Detecting architecture"
    local _ostype="$(uname -s)"
    local _cputype="$(uname -m)"

    verbose_say "uname -s reports: $_ostype"
    verbose_say "uname -m reports: $_cputype"

    if [ "$_ostype" = Darwin -a "$_cputype" = i386 ]; then
        # Darwin `uname -s` lies
        if sysctl hw.optional.x86_64 | grep -q ': 1'; then
            local _cputype=x86_64
        fi
    fi

    case "$_ostype" in

        Linux)
            local _ostype=linux
            ;;

        FreeBSD)
            local _ostype=freebsd
            ;;

        DragonFly)
            local _ostype=linux
            ;;

        Darwin)
            local _ostype=darwin
            ;;
        *)
            err "Unknown OS type: $_ostype"
            ;;

    esac

    case "$_cputype" in

        i386 | i486 | i686 | i786 | x86)
            local _cputype=x86
            ;;
        x86_64 | x86-64 | x64 | amd64)
            local _cputype=x86_64
            ;;
        *)
            err "Unknown CPU type: $_cputype"
            ;;
    esac

    verbose_say "OS is $_ostype"
    verbose_say "Architecture is $_cputype"
    ARCH="$_cputype"
    OS="$_ostype"
}

handle_cmd_line_args() {
    LOCATION="/usr/local/bin"
    for _arg in "$@"; do
        case "${_arg%%=*}" in
            --verbose)
                VERBOSE=true
                ;;
            --location)
                if is_value_arg "$_arg" "location"; then
                    LOCATION="$(get_value_arg "$_arg")"
                fi
                ;;
        esac
    done
}

is_value_arg() {
    local _arg="$1"
    local _name="$2"

    echo "$_arg" | grep -q -- "--$_name="
    return $?
}

get_value_arg() {
    local _arg="$1"

    echo "$_arg" | cut -f2 -d=
}

assert_cmds_available() {
    need_cmd echo
    need_cmd curl
    need_cmd mktemp
    need_cmd mkdir
    need_cmd pwd
    need_cmd grep
    need_cmd cut
    need_cmd tr
    need_cmd uname
    need_cmd rm
    need_cmd unzip
}

need_cmd() {
    if ! command -v "$1" > /dev/null 2>&1
    then err "need '$1' (command not found)"
    fi
}

ensure() {
    "$@"
    if [ $? != 0 ]; then err "command failed: $*"; fi
}

say() {
    echo "$1"
}

verbose_say() {
    if [ "$VERBOSE" = true ]; then
        say "[DEBUG] $1"
    fi
}

err() {
    say "$1" >&2
    exit 1
}

main() {
    assert_cmds_available
    local _ansi_escapes_are_valid=false
    if [ -t 2 ]; then
        if [ "${TERM+set}" = 'set' ]; then
            case "$TERM" in
                xterm*|rxvt*|urxvt*|linux*|vt*)
                    _ansi_escapes_are_valid=true
                ;;
            esac
        fi
    fi
    handle_cmd_line_args "$@"
    set_os_architecture
    download
}

main "$@"