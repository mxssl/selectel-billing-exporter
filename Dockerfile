FROM golang:1.19.3-alpine3.15 as builder

ENV GO111MODULE=on

WORKDIR /go/src/github.com/mxssl/selectel-billing-exporter
COPY . .

# install deps
RUN apk add --no-cache \
  ca-certificates \
  curl \
  git

RUN CGO_ENABLED=0 \
  GOOS=`go env GOHOSTOS` \
  GOARCH=`go env GOHOSTARCH` \
  go build -v -o app

# copy compiled binary to a clear Alpine Linux image
FROM alpine:3.18.2
WORKDIR /
RUN apk add --no-cache \
  ca-certificates
COPY --from=builder /go/src/github.com/mxssl/selectel-billing-exporter .
RUN chmod +x app
CMD ["./app"]
