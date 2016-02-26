#!/bin/bash
# mkdeb name version arch

# Original source: https://github.com/gauge/gauge/blob/master/script/mkdeb

set -e

SCRIPT=`readlink -f "$0"`
ROOT=`readlink -f $(dirname $SCRIPT)/..`
DEPLOY_DIR="$ROOT/deploy"
BUILD_DIR="$ROOT/build"
ARCH="i386"
NAME="gauge"
FILE_MODE=755
CONTROL_FILE="$BUILD_DIR/deb/control"
POSTINST_FILE="$BUILD_DIR/deb/postinst"

rm -rf "$DEPLOY_DIR"
go run build/make.go
go run build/make.go --distro


pushd $DEPLOY_DIR > /dev/null
  VERSION=$(ls $NAME*zip | head -1 | sed "s/\.[^\.]*$//" | sed "s/$NAME-//" | sed "s/-[a-z]*\.[a-z0-9_]*$//")
  ARCHTYPE=$(ls $NAME*zip | head -1 | rev | cut -d '-' -f 1 | rev | cut -d '.' -f 2)

  if [ ARCHTYPE == "x86_64" ]; then
    ARCH="amd64"
  fi

  PKG="$DEPLOY_DIR/$NAME-$VERSION-linux.$ARCHTYPE.zip"
  PKG_SRC="$DEPLOY_DIR/$NAME-$VERSION-pkg"

  unzip -q "$PKG" -d "$PKG_SRC"
popd > /dev/null

TARGET_ROOT="$DEPLOY_DIR/deb"
rm -rf "$TARGET_ROOT"
mkdir -p "$TARGET_ROOT"
chmod $FILE_MODE "$TARGET_ROOT"
TARGET="$TARGET_ROOT/$NAME-$VERSION-$ARCH"
DEB_PATH="$DEPLOY_DIR/"

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

fakeroot dpkg-deb -b "$TARGET"
mv "$TARGET_ROOT/$NAME-$VERSION-$ARCH.deb" "$DEB_PATH"
rm -rf "$TARGET_ROOT"
rm -rf "$PKG_SRC"
echo -e "\nCreated .deb package at: $DEB_PATH$NAME-$VERSION-$ARCH.deb"
echo -e "  Version : $VERSION"
echo -e "  Arch    : $ARCH\n"
