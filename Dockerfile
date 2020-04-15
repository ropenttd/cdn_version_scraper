# BUILD ENVIRONMENT

FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
# All these steps will be cached
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/github.com/ropenttd/cdn_version_scraper/pkg/cdn_version_scraper
COPY go.mod .
COPY go.sum .

# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download

# Then copy the rest of this source code
COPY . .

# And build the binary
RUN go build -o /go/bin/cdn_version_scraper

# END BUILD ENVIRONMENT
# DEPLOY ENVIRONMENT

FROM scratch
MAINTAINER duck. <me@duck.me.uk>

# Copy the executable
COPY --from=builder /go/bin/cdn_version_scraper /usr/local/bin/cdn_version_scraper

# And set it as the entrypoint
ENTRYPOINT ["/usr/local/bin/cdn_version_scraper"]
