#!/usr/bin/env bash

set -eu

if [ -f secrets/air-credentials.json ]; then
	export OAUTH_CREDENTIALS=$(<secrets/air-credentials.json)
fi

export DEBUG=true

go run github.com/air-verse/air@latest $@
