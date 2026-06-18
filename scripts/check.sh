#!/usr/bin/env bash
# Gate local: roda tudo que o CI roda, antes de commitar.
set -euo pipefail
cd "$(dirname "$0")/.."

echo "==> gofmt"
unformatted="$(gofmt -l .)"
if [ -n "$unformatted" ]; then
	echo "Arquivos não formatados (rode 'gofmt -w .'):"
	echo "$unformatted"
	exit 1
fi

echo "==> go vet"
go vet ./...

echo "==> golangci-lint"
if command -v golangci-lint >/dev/null 2>&1; then
	golangci-lint run ./...
else
	echo "  (golangci-lint não instalado — pulando; CI ainda roda)"
fi

echo "==> go test -race"
go test -race ./...

echo "==> go build"
go build -o cts.exe .

echo "OK — gate verde."
