#!/bin/bash
#
# scripts/ci_checks.sh
# helper script that runs all CI checks
#
#
# Copyright (C) 2019 Anexia Internetdienstleistungs GmbH

set -eu

ROOT_PACKAGE=${1:-github.com/anexia-it/geodbtools}

THIS_PATH=$(dirname $0)

echo "> format"
${THIS_PATH}/format.sh

echo "> format diff check"
diff -u <(echo -n) <(git diff)

echo "> lint"
${THIS_PATH}/lint.sh

echo "> vet"
${THIS_PATH}/vet.sh

echo "> staticcheck"
${THIS_PATH}/staticcheck.sh

echo "> coverage"
${THIS_PATH}/coverage.sh

echo "> build"
${THIS_PATH}/build.sh
