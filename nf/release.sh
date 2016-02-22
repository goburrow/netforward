#!/bin/sh
set -e

build() {
	echo "Compiling for $1-$2"
	GOOS=$1 GOARCH=$2 go build
	if [ "$1" == "windows" ]; then
		file nf.exe
		zip "nf.$1-$2.zip" nf.exe
	else
		file nf
		tar cvzf "nf.$1-$2.tar.gz" nf
	fi
}

build linux amd64
build darwin amd64
build windows amd64
build linux arm

rm nf nf.exe
