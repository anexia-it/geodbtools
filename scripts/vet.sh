#!/bin/bash
#
# scripts/vet.sh
# helper script that checks code for go vet warnings
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

EXIT_STATUS=0
for pkg in ${PKG_LIST}
do
    echo "> go vet: ${pkg}..."
    if ! go vet ${pkg}
    then
	EXIT_STATUS=1
    fi
done

exit ${EXIT_STATUS}
