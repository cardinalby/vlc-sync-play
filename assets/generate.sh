#!/bin/bash

SRC_TRAY_ICON="icons/tray_icon.psd"
DST_PNG="generated/tray_icon.png"
DST_ICO="generated/tray_icon.ico"

IMAGEMAGICK_IMG="dpokidov/imagemagick"

docker run --rm -v .:/imgs dpokidov/imagemagick /imgs/${SRC_TRAY_ICON}[0] PNG:/imgs/${DST_PNG} & \
docker run --rm -v .:/imgs dpokidov/imagemagick /imgs/${SRC_TRAY_ICON}[0] \
-define icon:auto-resize="256,128,96,64,48,32,16" ICO:/imgs/${DST_ICO}