#!/bin/sh

if [ "$1" == "" ]; then
    VERSION=$(git describe --tags $(git rev-list --tags --max-count=1))
    echo "Building with latest git version $VERSION"
else
    VERSION="$1"
    echo "Building with explicit version $VERSION"
fi

REVISION=$(git rev-parse --short HEAD)
BUILD_TIME=$(date +"%Y%m%d.%H%M%S")

LDFLAGS="-X github.com/scribblerockerz/parachute/cmd/version.Version=$VERSION"
LDFLAGS="$LDFLAGS -X github.com/scribblerockerz/parachute/cmd/version.Revision=$REVISION"
LDFLAGS="$LDFLAGS -X github.com/scribblerockerz/parachute/cmd/version.BuildDate=$BUILD_TIME"

# build executables
CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags="$LDFLAGS" -o bin/parachute *.go && \
echo "Successfully build executable for LINUX to ./bin/parachute"

CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -ldflags="$LDFLAGS" -o bin/parachute-macos *.go && \
echo "Successfully build executable for MACOS to ./bin/parachute-macos"