FROM golang:1.13.5-alpine3.10 as builder

ENV GO111MODULE=on

WORKDIR /go/src/github.com/mxssl/selectel_billing_exporter
COPY . .

# install external depends
RUN apk add --no-cache \
  ca-certificates \
  curl \
  git
RUN CGO_ENABLED=0 \
  GOOS=`go env GOHOSTOS` \
  GOARCH=`go env GOHOSTARCH` \
  go build -v -mod=vendor -o app

# Copy compiled binary to clear Alpine Linux image
FROM alpine:3.11.0
WORKDIR /
RUN apk add --no-cache \
  ca-certificates
COPY --from=builder /go/src/github.com/mxssl/selectel_billing_exporter .
RUN chmod +x app
CMD ["./app"]
