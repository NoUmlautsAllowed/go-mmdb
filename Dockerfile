# build standalone docker container

# install certs
FROM alpine:latest AS certs
RUN apk --no-cache add ca-certificates

# Start from the latest golang base image
FROM golang:alpine AS golangbuilder

COPY . /go/src/go-mmdb
WORKDIR /go/src/go-mmdb

RUN --mount=type=cache,target=/go-cache,sharing=private \
    export GOCACHE=/go-cache/go-build GOMODCACHE=/go-cache/mod && go mod download && \
    go build -a -tags netgo -v -ldflags '-w -extldflags "-static"' -o /go/bin/server ./cmd/server && \
    go build -a -tags netgo -v -ldflags '-w -extldflags "-static"' -o /go/bin/downloader ./cmd/downloader

FROM scratch

# copy cert files
WORKDIR /etc/ssl/certs
COPY --from=certs /etc/ssl/certs .

# target directory in image
WORKDIR /go-mmdb

# copy go binaries
COPY --from=golangbuilder /go/bin/server .
COPY --from=golangbuilder /go/bin/downloader .

EXPOSE 8080
CMD ["./server"]
