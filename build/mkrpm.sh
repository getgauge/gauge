#!/bin/bash
# Original source: https://github.com/gauge/gauge/blob/master/script/mkrpm

# Usage:
# ./build/mkrpm.sh [--rebuild]

set -e
if [[ -z $GOPATH ]]; then
    export GOPATH=`pwd`
fi
if [[ -z $GOBIN ]]; then
    export GOBIN="$GOPATH/bin"
fi

cd $GOPATH/src/github.com/getgauge/gauge

go get github.com/tools/godep && $GOBIN/godep restore

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
RELEASE=all
SPEC_FILE="$BUILD_DIR/packaging/rpm/gauge.spec"

if [ "$OS" != "linux" ]; then
    err "This script can only be run on Linux systems"
fi

if [ "$1" == "--rebuild" ]; then
    REBUILD_NEEDED=1
fi

if [ "$2" == "--nightly" ]; then
    NIGHTLY="--nightly"
fi

function rebuild () {
    rm -rf "$DEPLOY_DIR"
    go run build/make.go --all-platforms --target-linux $NIGHTLY
    go run build/make.go --distro --all-platforms --target-linux $NIGHTLY
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
        ARCH="x86_64"
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
    RPM_VERSION=$(echo $VERSION | sed "s/-//g")
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
    TARGET="$HOME/rpmbuild"
    rm -rf "$TARGET"
    RPM_PATH="$DEPLOY_DIR/"
}

function prep_rpm() {
    echo "Preparing .rpm data..."
    rpmdev-setuptree

    mkdir -m $FILE_MODE -p "$TARGET/BUILD/bin/"
    cp -r "$PKG_SRC/" "$TARGET/BUILD/bin/"

    SPEC_DATA=`cat "$SPEC_FILE"`
    echo "$SPEC_DATA" | sed "s/<version>/$RPM_VERSION/g" | sed "s/<release>/$RELEASE/g" > "$TARGET/SPECS/gauge.spec"
    cat $TARGET/SPECS/gauge.spec
    # Copy generated LICENSE.md to /usr/share/doc/gauge/copyright
    mkdir -m $FILE_MODE -p "$TARGET/BUILD/usr/local/share/doc/$NAME"
    cp "$ROOT/LICENSE" "$TARGET/BUILD/usr/local/share/doc/$NAME/copyright"
}

function create_rpm() {
    echo "Generating .rpm..."
    rpmbuild --target $ARCH-redhat-linux -ba "$TARGET/SPECS/gauge.spec"
    mv $TARGET/RPMS/$ARCH/$NAME-$RPM_VERSION-$RELEASE.$ARCH.rpm "$RPM_PATH/"
}

function cleanup_temp() {
    rm -rf "$TARGET"
    rm -rf "$PKG_SRC"
}

function print_status() {
    echo -e "\nCreated .rpm package in: $RPM_PATH$NAME-$RPM_VERSION-$RELEASE.$ARCH.rpm"
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
        prep_rpm
        create_rpm
        cleanup_temp
        print_status
    done
}

# Let the game begin
init
