# BUILD ENVIRONMENT

FROM golang:alpine AS builder

# All these steps will be cached
RUN apk update && apk add --no-cache ca-certificates
WORKDIR $GOPATH/src/github.com/ropenttd/cdn_version_scraper
COPY go.mod .
COPY go.sum .

# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download

# Then copy the rest of this source code
COPY . .

# And build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/cdn_version_scraper

# END BUILD ENVIRONMENT
# DEPLOY ENVIRONMENT

FROM scratch
MAINTAINER duck. <me@duck.me.uk>

# Copy the executable
COPY --from=builder /go/bin/cdn_version_scraper /cdn_version_scraper
# and the ca-certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# And set it as the entrypoint
ENTRYPOINT ["/cdn_version_scraper"]
