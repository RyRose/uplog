# syntax=docker/dockerfile:1

# Directory storing built binary.
ARG BINDIR=/usr/local/bin

# Full path to built binary.
ARG BINPATH=${BINDIR}/uplog

# Directory storing source files.
ARG SRCDIR=/usr/src/uplog

##################
# Generate Templ #
##################

FROM ghcr.io/a-h/templ:v0.3.833 AS generate-templ-stage
ARG SRCDIR

# Copy all non-generated source files.
COPY --chown=65532:65532 go.mod go.sum sqlc.yaml tailwind.config.js package.json package-lock.json ${SRCDIR}/
COPY --chown=65532:65532 cmd ${SRCDIR}/cmd
COPY --chown=65532:65532 internal ${SRCDIR}/internal
COPY --chown=65532:65532 web ${SRCDIR}/web

# Generate *_templ.go files.
WORKDIR ${SRCDIR}
RUN ["templ", "generate"]

#################
# Generate sqlc #
#################

FROM sqlc/sqlc:1.27.0 AS generate-sqlc-stage
ARG SRCDIR

# Copy all source files (templ-generated and non-generated).
COPY --from=generate-templ-stage ${SRCDIR} ${SRCDIR}

# Generate sqlc files.
WORKDIR ${SRCDIR}
RUN ["/workspace/sqlc", "generate"]

#####################
# Generate tailwind #
#####################

FROM node:22.9.0 AS generate-tailwind-stage
ARG SRCDIR

# Copy all source files (templ-generated, sqlc-generated, and non-generated).
COPY --from=generate-sqlc-stage ${SRCDIR} ${SRCDIR}

WORKDIR ${SRCDIR}
RUN npm ci
RUN npx tailwindcss -i web/app/input.css -o web/static/css/output.css --minify

#########
# Build #
#########

FROM golang:1.23 AS build-stage
ARG BINDIR
ARG SRCDIR

# Cache go build to be faster.
ENV GOCACHE=/cache/go
ENV GOMODCACHE=/cache/gomod

# Enable cgo for sqlite.
ENV CGO_ENABLED=1

# Copy all source files (generated and non-generated).
COPY --from=generate-tailwind-stage ${SRCDIR} ${SRCDIR}

# Download+verify dependencies and build server statically-linked.
WORKDIR ${SRCDIR}
RUN --mount=type=cache,target=/cache <<EOF
set -eux
go mod download
go mod verify
go build -v \
	-ldflags="-extldflags=-static" \
	-tags sqlite_omit_load_extension \
	-o ${BINDIR} ./...
EOF

##########
# Deploy #
##########

FROM alpine:3.20
ARG SRCDIR
ARG BINPATH

RUN <<EOF
set -eux

# Add a new user with minimal privileges.
addgroup --gid 1000 appgroup
adduser --uid 1000 --ingroup appgroup --disabled-password appuser

# Add tzdata to support TZ env variable. This gives proper timezone in logs.
apk add tzdata

# Make /data and assign it perms to appuser:appgroup.
mkdir /data
chown appuser /data
chgrp appgroup /data
EOF

# Copy web assets and mark them as owned by new user.
COPY --chown=appuser:appgroup --from=build-stage ${SRCDIR}/web ${SRCDIR}/web
COPY --chown=appuser:appgroup docs/swagger.json ${SRCDIR}/docs/swagger.json
COPY --chown=appuser:appgroup docs/swagger.yaml ${SRCDIR}/docs/swagger.yaml

# Copy statically-linked go binary and mark as owned by new user.
COPY --chown=appuser:appgroup --from=build-stage ${BINPATH} ${BINPATH}

ENV PORT=8080
ENV DATABASE_PATH=/data/workout.db

# Application version (set by build system).
ARG VERSION=unknown
ENV VERSION=${VERSION}

USER appuser:appgroup
WORKDIR ${SRCDIR}
CMD ["uplog"]
