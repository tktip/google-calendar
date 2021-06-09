#!/bin/sh
# installs tools required for makefile to work

echo "-- Installing revive"
GO111MODULE=off go get github.com/mgechev/revive
GO_GET_EXIT="$?"
echo "-- Installed revive\n"
exit $GO_GET_EXIT
