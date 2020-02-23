#!/usr/bin/env bash
set -o errexit
set -o nounset

# install specific golangci-lint release in travis only
TRAVISBUILD=${TRAVIS:-}
if [ ! -z "${TRAVISBUILD}" ]; then
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin ${GOLANGCILINT_VERSION}
fi

# make reports directory (if it doesn't exist)
mkdir -p ${REPORTS_DIR}

# run golangci-lint, and generate report in travis only
echo "Running golangci-lint"
if [ -z "${TRAVISBUILD}" ]; then
	golangci-lint run ./...
else
	golangci-lint run --out-format=checkstyle ./... | tee /dev/tty > ${REPORTS_DIR}/checkstyle-report.xml
fi
echo "no linting problems found"