#!/bin/bash
#
# scripts/coverage.sh
# test coverage CI helper script for geodbtools
#
#
# Copyright (C) 2019 Anexia Internetdienstleistungs GmbH

set -eu

ROOT_PACKAGE=${1:-github.com/anexia-it/geodbtools}

mkdir -p cover/
rm -f cover/list_errors.txt
touch cover/list_errors.txt
rm -f cover/*.cov

PKG_LIST=$(go list ${ROOT_PACKAGE}/... 2>cover/list_errors.txt | grep -v '/vendor/' | sort)

if test -f cover/list_errors.txt -a ! -z "$(cat cover/list_errors.txt)"
then
    echo "go list failed: " >&2
    cat cover/list_errors.txt >&2
    exit 1
fi

for pkg in ${PKG_LIST}
do
    pkg_clean=$(echo ${pkg} | sed -e 's|/|_|g')
    echo "> go test: ${pkg}..."
    go test -covermode=count -coverprofile "cover/${pkg_clean}.cov" "${pkg}"
done

echo "mode: count" > cover/coverage.cov
tail -q -n +2 cover/*.cov >> cover/coverage.cov
go tool cover -func=cover/coverage.cov | tee cover/report.txt | egrep '^total:' | awk '{printf "> combined coverage: %s\n", $NF }'
echo "> detailed coverage report: "
cat cover/report.txt
