#!/bin/bash
# Original source: https://github.com/gauge/gauge/blob/master/script/mkdeb

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
FILE_MODE=755
CONTROL_FILE="$BUILD_DIR/deb/control"
POSTINST_FILE="$BUILD_DIR/deb/postinst"

if [ "$OS" != "linux" ]; then
    err "This script can only be run on Linux systems"
fi

if [ "$1" == "--rebuild" ]; then
    REBUILD_NEEDED=1
fi

function rebuild () {
    rm -rf "$DEPLOY_DIR"
    go run build/make.go
    go run build/make.go --distro
}

function check_and_rebuild() {
    if [ ! -d "$DEPLOY_DIR" ]; then
        echo "Building distro packages..."
        rebuild
    elif [ ! -z "$REBUILD_NEEDED" ]; then
        echo "Rebuild flag set. Rebuilding distro packages..."
        rebuild
    else
        echo "Reusing existing distro package. Use '--rebuild' to trigger a package rebuild..."
    fi
}

function set_version() {
    VERSION=$(ls $NAME*zip | head -1 | sed "s/\.[^\.]*$//" | sed "s/$NAME-//" | sed "s/-[a-z]*\.[a-z0-9_]*$//")
}

function set_arch() {
    ARCHTYPE=$(ls $NAME*zip | head -1 | rev | cut -d '-' -f 1 | rev | cut -d '.' -f 2)

    if [ ARCHTYPE == "x86_64" ]; then
        ARCH="amd64"
    fi
}

function set_pkg_info() {
    PKG="$DEPLOY_DIR/$NAME-$VERSION-$OS.$ARCHTYPE.zip"
    PKG_SRC="$DEPLOY_DIR/$NAME-$VERSION-pkg"
}

function set_info() {
    set_version
    set_arch
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
    mkdir -m $FILE_MODE -p "$TARGET/usr"

    cp -r "$PKG_SRC/bin" "$TARGET/usr/"
    cp -r "$PKG_SRC/share" "$TARGET/usr/"

    mkdir -m $FILE_MODE -p "$TARGET/DEBIAN"
    cp "$CONTROL_FILE" "$TARGET/DEBIAN/control"
    cp "$POSTINST_FILE" "$TARGET/DEBIAN/postinst"

    CONTROL_DATA=`cat "$TARGET/DEBIAN/control"`
    echo "$CONTROL_DATA" | sed "s/<version>/$VERSION/" | sed "s/<arch>/$ARCH/" > "$TARGET/DEBIAN/control"

    # Copy generated LICENSE.md to /usr/share/doc/gauge/copyright
    mkdir -m $FILE_MODE -p "$TARGET/usr/share/doc/$NAME"
    cp "$ROOT/LICENSE" "$TARGET/usr/share/doc/$NAME/copyright"
}

function create_deb() {
    echo "Generating .deb"
    fakeroot dpkg-deb -b "$TARGET"
    mv "$TARGET_ROOT/$NAME-$VERSION-$ARCH.deb" "$DEB_PATH"
}

function init() {
    check_and_rebuild

    pushd $DEPLOY_DIR > /dev/null
    set_info
    unzip -q "$PKG" -d "$PKG_SRC"
    popd > /dev/null

    clean_stage
    prep_deb
    create_deb
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

init
cleanup_temp
print_status
