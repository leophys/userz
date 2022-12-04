#!/bin/sh

mkdir /plugin \
    && go build -buildmode=plugin -o /plugin/ ./internal/pollednotifier/... \
    && go test -v -p 1 -timeout 600s ./tests/... -tags=integration
