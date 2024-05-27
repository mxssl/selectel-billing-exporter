# syntax=docker/dockerfile:1
FROM golang:1.22.2-alpine3.18 as builder
WORKDIR /go/src/github.com/mxssl/selectel-billing-exporter
COPY . .
RUN <<EOF
  apk add --no-cache \
    ca-certificates \
    curl \
    git
EOF
RUN <<EOF
  CGO_ENABLED=0 \
  go build -v -o app
EOF

FROM alpine:3.20.0
WORKDIR /
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/github.com/mxssl/selectel-billing-exporter .
RUN chmod +x app
CMD ["./app"]
