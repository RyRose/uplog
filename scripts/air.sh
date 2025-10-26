#!/usr/bin/env bash

set -eu

export DEBUG=true

go run github.com/air-verse/air@latest $@
