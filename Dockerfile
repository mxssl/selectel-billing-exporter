FROM golang:1.21.6-alpine3.18 as builder

ENV GO111MODULE=on

WORKDIR /go/src/github.com/mxssl/selectel-billing-exporter
COPY . .

RUN apk add --no-cache \
  ca-certificates \
  curl \
  git

RUN CGO_ENABLED=0 \
  go build -v -o app

FROM alpine:3.19.1
WORKDIR /
RUN apk add --no-cache \
  ca-certificates
COPY --from=builder /go/src/github.com/mxssl/selectel-billing-exporter .
RUN chmod +x app
CMD ["./app"]
