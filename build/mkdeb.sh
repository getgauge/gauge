#!/bin/bash
# Original source: https://github.com/gauge/gauge/blob/master/script/mkdeb

# Usage:
# ./build/mkdeb.sh [--rebuild]

export GOPATH=`pwd`
export GOBIN="$GOPATH/bin"

cd $GOPATH/src/github.com/getgauge/gauge

go get github.com/tools/godep && $GOBIN/godep restore

set -e

function err () {
    echo "ERROR: $1"
    exit 1
}

ROOT=`pwd -P`
DEPLOY_DIR="$ROOT/deploy"
BUILD_DIR="$ROOT/build"
OS=`uname -s | tr '[:upper:]' '[:lower:]'`
ARCH="i386"
NAME="gauge"
FILE_EXT="zip"
FILE_MODE=755
CONTROL_FILE="$BUILD_DIR/packaging/deb/control"

if [ "$OS" != "linux" ]; then
    err "This script can only be run on Linux systems"
fi

if [ "$1" == "--rebuild" ]; then
    REBUILD_NEEDED=1
fi

function rebuild () {
    rm -rf "$DEPLOY_DIR"
    go run build/make.go --all-platforms --target-linux
    go run build/make.go --distro --all-platforms --target-linux
}

function check_and_rebuild() {
    if [ ! -d "$DEPLOY_DIR" ]; then
        echo -e "Building distro packages...\n"
        rebuild
    elif [ ! -z "$REBUILD_NEEDED" ]; then
        echo -e "Rebuild flag set. Rebuilding distro packages...\n"
        rebuild
    else
        echo -e "Reusing existing distro package. Use '--rebuild' to trigger a package rebuild...\n"
    fi
}

function set_arch() {
    if [ -z "$1" ]; then
        ARCHTYPE=$(ls $NAME*.$FILE_EXT | head -1 | rev | cut -d '-' -f 1 | rev | cut -d '.' -f 2)
    else
        ARCHTYPE=$(echo $1 | sed "s/^[a-z]*\///" | rev | cut -d '-' -f 1 | rev | cut -d '.' -f 2)
    fi

    if [ "$ARCHTYPE" == "x86_64" ]; then
        ARCH="amd64"
    else
        ARCH="i386"
    fi
}

function set_version() {
    if [ -z "$1" ]; then
        VERSION=$(ls $NAME*$ARCHTYPE.$FILE_EXT | head -1 | sed "s/\.[^\.]*$//" | sed "s/$NAME-//" | sed "s/-[a-z]*\.[a-z0-9_]*$//")
    else
        VERSION=$(echo `basename $1` | sed "s/^[a-z]*\///" | sed "s/\.[^\.]*$//" | sed "s/$NAME-//" | sed "s/-[a-z]*\.[a-z0-9_]*$//")
    fi
}

function set_pkg_info() {
    PKG="$DEPLOY_DIR/$NAME-$VERSION-$OS.$ARCHTYPE.$FILE_EXT"
    PKG_SRC="$DEPLOY_DIR/$NAME-$VERSION-pkg"
}

function set_info() {
    set_arch "$1"
    set_version "$1"
    set_pkg_info
}

function clean_stage() {
    TARGET_ROOT="$DEPLOY_DIR/deb"
    rm -rf "$TARGET_ROOT"
    mkdir -p "$TARGET_ROOT"
    chmod $FILE_MODE "$TARGET_ROOT"
    TARGET="$TARGET_ROOT/$NAME-$VERSION-$ARCH"
    DEB_PATH="$DEPLOY_DIR/"
}

function prep_deb() {
    echo "Preparing .deb data..."
    mkdir -m $FILE_MODE -p "$TARGET/usr/local/gauge"

    cp -r "$PKG_SRC/bin" "$TARGET/usr/local"

    mkdir -m $FILE_MODE -p "$TARGET/DEBIAN"
    cp "$CONTROL_FILE" "$TARGET/DEBIAN/control"

    chmod +x $TARGET/usr/local/bin/*

    sync

    CONTROL_DATA=$(cat "$TARGET/DEBIAN/control")
    INSTALLED_SIZE=$(du -s $PKG_SRC/bin/ | sed "s/^\([0-9]*\).*$/\1/")
    while [ $INSTALLED_SIZE -lt 1 ]; do
            INSTALLED_SIZE=$(du -s $PKG_SRC/bin/ | sed "s/^\([0-9]*\).*$/\1/")
    done
    echo "$CONTROL_DATA" | sed "s/<version>/$VERSION/" | sed "s/<arch>/$ARCH/" | sed "s/<size>/$INSTALLED_SIZE/" > "$TARGET/DEBIAN/control"

    # Copy generated LICENSE.md to /usr/share/doc/gauge/copyright
    mkdir -m $FILE_MODE -p "$TARGET/usr/share/doc/$NAME"
    cp "$ROOT/LICENSE" "$TARGET/usr/share/doc/$NAME/copyright"
}

function create_deb() {
    echo "Generating .deb..."
    fakeroot dpkg-deb -b "$TARGET"
    mv "$TARGET_ROOT/$NAME-$VERSION-$ARCH.deb" "$DEB_PATH"
}

function cleanup_temp() {
    rm -rf "$TARGET_ROOT"
    rm -rf "$PKG_SRC"
}

function print_status() {
    echo -e "\nCreated .deb package at: $DEB_PATH$NAME-$VERSION-$ARCH.deb"
    echo -e "  Version : $VERSION"
    echo -e "  Arch    : $ARCH\n"
}

function init() {
    check_and_rebuild

    for f in `ls $DEPLOY_DIR/$NAME-*$OS*.$FILE_EXT`; do
        clean_stage

        pushd $DEPLOY_DIR > /dev/null
        set_info "$f"
        unzip -q "$PKG" -d "$PKG_SRC"
        popd > /dev/null

        clean_stage
        prep_deb
        create_deb
        cleanup_temp
        print_status
    done
}

# Let the game begin
init
