#!/bin/sh
#
# DON'T EDIT THIS!
#
# CodeCrafters uses this file to test your code. Don't make any changes here!
#
# DON'T EDIT THIS!
set -e
tmpFile=$(mktemp)
go build -o "$tmpFile" app/*.go
# /usr/local/go/bin/go build -o "$tmpFile" app/*.go #use this instead of line 10 locally running because of chroot requires sudo, and for somre reason sudo requires the full go binary path
exec "$tmpFile" "$@"
