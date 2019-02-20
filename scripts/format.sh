#!/bin/bash
#
# scripts/format.sh
# helper script that fixes formatting
#
#
# Copyright (C) 2019 Anexia Internetdienstleistungs GmbH

set -eu

ROOT_PACKAGE=${1:-github.com/anexia-it/geodbtools}

mkdir -p cover/
PKG_LIST=$(go list ${ROOT_PACKAGE}/... 2>cover/list_errors.txt | grep -v '/vendor/' | sort)

if test -f cover/list_errors.txt -a ! -z "$(cat cover/list_errors.txt)"
then
    echo "go list failed: " >&2
    cat cover/list_errors.txt >&2
    exit 1
fi

for pkg in ${PKG_LIST}
do
    echo "> go fmt: ${pkg}..."
    go fmt ${pkg}
done
