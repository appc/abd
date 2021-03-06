#!/usr/bin/env bash
set -e

# Gets the directory that this script is stored in.
# https://stackoverflow.com/questions/59895/can-a-bash-script-tell-what-directory-its-stored-in
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

ORG_PATH="github.com/appc"
REPO_PATH="${ORG_PATH}/abd"

if [ ! -h ${DIR}/gopath/src/${REPO_PATH} ]; then
  mkdir -p ${DIR}/gopath/src/${ORG_PATH}
  cd ${DIR} && ln -s ../../../.. gopath/src/${REPO_PATH} || exit 255
fi

export GO15VENDOREXPERIMENT=1
export GOBIN=${DIR}/bin
export GOPATH=${DIR}/gopath
export GOOS GOARCH

eval $(go env)

if [ "${GOOS}" = "freebsd" ]; then
    # /usr/bin/cc is clang on freebsd, but we need to tell it to go to
    # make it generate proper flavour of code that doesn't emit
    # warnings.
    export CC=clang
fi

echo "Building abd..."
go build -o ${GOBIN}/abd ${REPO_PATH}

for STRATEGY in io.abd.https-dns io.abd.local io.abd.nfs io.abd.noop; do
  go build -o ${GOBIN}/${STRATEGY} ${REPO_PATH}/strategies/${STRATEGY}
done
