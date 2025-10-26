#!/usr/bin/env bash

set -eu

# Clean go files.
go clean

# Delete all generated files.
find . -type f -name "*_templ.go" -exec rm {} \;

# Clear tmp/ files.
rm -r ./tmp/

# Clear generated tailwind output file.
rm ./web/static/css/output.css

# Clear generated sqlc files.
rm -r ./internal/sqlc/workoutdb/
