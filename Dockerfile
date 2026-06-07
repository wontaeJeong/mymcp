FROM golang:1.23-alpine AS build

ARG HTTP_PROXY
ARG HTTPS_PROXY
ARG NO_PROXY
ARG http_proxy
ARG https_proxy
ARG no_proxy

RUN HTTP_PROXY="${HTTP_PROXY}" \
    HTTPS_PROXY="${HTTPS_PROXY}" \
    NO_PROXY="${NO_PROXY}" \
    http_proxy="${http_proxy}" \
    https_proxy="${https_proxy}" \
    no_proxy="${no_proxy}" \
    apk add --no-cache ca-certificates

COPY certs/ /usr/local/share/ca-certificates/internal/
RUN update-ca-certificates

WORKDIR /src

COPY go.mod ./
RUN HTTP_PROXY="${HTTP_PROXY}" \
    HTTPS_PROXY="${HTTPS_PROXY}" \
    NO_PROXY="${NO_PROXY}" \
    http_proxy="${http_proxy}" \
    https_proxy="${https_proxy}" \
    no_proxy="${no_proxy}" \
    go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH
RUN HTTP_PROXY="${HTTP_PROXY}" \
    HTTPS_PROXY="${HTTPS_PROXY}" \
    NO_PROXY="${NO_PROXY}" \
    http_proxy="${http_proxy}" \
    https_proxy="${https_proxy}" \
    no_proxy="${no_proxy}" \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=${TARGETARCH:-$(go env GOARCH)} go build -trimpath -ldflags="-s -w" -o /out/mymcp ./cmd/mymcp

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/mymcp /mymcp

USER 65532:65532
EXPOSE 3000
ENTRYPOINT ["/mymcp"]
